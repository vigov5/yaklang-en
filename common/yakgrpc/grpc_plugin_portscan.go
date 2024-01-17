package yakgrpc

import (
	"fmt"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"strings"
)

type portScanParam struct {
	Target string
	Port   string
}

const PORTSCAN_PLUGIN_TRIGGER_CODE = `yakit.AutoInitYakit()
log.setLevel("info")

target = cli.String("target", cli.setRequired(true))
port = cli.String("port", cli.setRequired(false))
name = cli.String("script-name", cli.setRequired(true))
cli.check()

log.info("TARGET: %v PORT: %v", target, port)
# input your yak code
res, err := servicescan.Scan(
    target, port, servicescan.active(false), servicescan.maxProbes(3), servicescan.probeTimeout(3),
    servicescan.databaseCache(true),
)
if err != nil {
    yakit.Error("Service scan failed: %v", err)
    die(err)
}

hookManager, err := hook.NewMixPluginCaller()

if err != nil {
    updateStatus("Failed to create plug-in calling module")
    die(err)
}

err = hookManager.LoadPlugin(name)
if err != nil {
    yakit.Error("Failed to load Yak plug-in: %v", err)
    die("no plugin loaded")
}

yakit.Info("Start executing service scan plug-in: %v", name)
for result = range res {
    if result.IsOpen() {
        yakit.Info("OPEN：%v: %s", str.HostPort(result.Target, result.Port), result.GetServiceName())
    } else {
        yakit.Info("CLOSED: %v", str.HostPort(result.Target, result.Port))
    }
    yakit.Info("Scan completed: %v, ready to execute plug-in: %v", str.HostPort(result.Target, result.Port), name)
    hookManager.HandleServiceScanResult(result)
}
`

func (s *Server) generatePortScanParams(scriptName string, params []*ypb.ExecParamItem) ([]*ypb.ExecParamItem, string, error) {
	var param = &portScanParam{}
	param.Port = "80"

	funk.ForEach(params, func(i *ypb.ExecParamItem) {
		switch strings.ToLower(i.GetKey()) {
		case "target":
			param.Target = i.GetValue()
		case "port", "ports":
			param.Port = i.GetValue()
		}
	})

	if param.Target == "" {
		return nil, "", utils.Error("target empty")
	}

	_, port, _ := utils.ParseStringToHostPort(param.Target)
	if port > 0 {
		param.Port += "," + fmt.Sprint(port)
	}

	if param.Port == "" {
		param.Port = "80"
	}
	var newParams []*ypb.ExecParamItem
	newParams = append(newParams, &ypb.ExecParamItem{
		Key:   "target",
		Value: param.Target,
	})
	newParams = append(newParams, &ypb.ExecParamItem{
		Key:   "port",
		Value: param.Port,
	})
	newParams = append(newParams, &ypb.ExecParamItem{
		Key:   "script-name",
		Value: scriptName,
	})
	return newParams, PORTSCAN_PLUGIN_TRIGGER_CODE, nil
}
