package yakgrpc

import (
	"context"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"io/ioutil"
	"os"
)

const execPacketScanCode = `
yakit.AutoInitYakit()

loglevel("info")
productMode = cli.Bool("prod")

// Get flowsID
flows = make([]int64)
flowsFile = cli.String("flows-file")
if flowsFile != "" {
    ports = str.ParseStringToPorts(string(file.ReadFile(flowsFile)[0]))
    for _, port = range ports {
        flows = append(flows, int64(port))
    }
}

if !productMode { yakit.StatusCard("Debug mode", "TRUE") }

if len(flows) <= 0 && !productMode {
    yakit.Info("Debug mode Supplement FlowID")
    flows = append(flows, 1142)
}

caller, err = hook.NewMixPluginCaller()
die(err)

// Load plug-in
plugins = cli.YakitPlugin() // yakit-plugin-file
if len(plugins) <= 0 && !productMode {
    // Set the default plug-in, generally used here for debugging
    plugins = [
        "Spring Cloud Function SPEL expression injection vulnerability detection", "MySQL CVE compliance check: 2016-2022",
    ]
}
loadedPlugins = []
failedPlugins = []
for _, pluginName = range x.If(plugins == undefined, [], plugins) {
    err := caller.LoadPlugin(pluginName)
    if err != nil {
        failedPlugins = append(failedPlugins, pluginName)
        yakit.Error("load plugin[%s] failed: %s", pluginName, err)
    }else{
        yakit.Info("Load plug-in: %v", pluginName)
        loadedPlugins = append(loadedPlugins, pluginName)
    }
}

if len(failedPlugins)>0 {
    yakit.StatusCard("Number of plug-in loading failures", len(failedPlugins))
}

if len(loadedPlugins) <= 0 {
    yakit.StatusCard("Execution failed", "No plug-in is loaded Correctly load", "error")
    die("No plug-in is loaded")
}else{
    yakit.StatusCard("Number of plug-in loads", len(loadedPlugins))
}

packetRawFile = cli.String("packet-file")
packetHttps = cli.Bool("packet-https")

// Set concurrency parameters
pluginConcurrent = 20
packetConcurrent = 20

// Set asynchronous, set concurrency, set Wait
caller.SetDividedContext(true)
caller.SetConcurrent(pluginConcurrent)
defer caller.Wait()

handleHTTPFlow = func(flow) {
    defer func{
        err = recover()
        if err != nil {
            yakit.Error("handle httpflow failed: %s", err)
        }
    }

    reqBytes = []byte(codec.StrconvUnquote(flow.Request)[0])
    rspBytes = []byte(codec.StrconvUnquote(flow.Response)[0])
    _, body = poc.Split(rspBytes)
    // Call the mirror traffic function and control whether to perform port scanning The plug-in calls
    caller.MirrorHTTPFlowEx(true, flow.IsHTTPS, flow.Url, reqBytes, rspBytes, body)
}

handleHTTPRequestRaw = func(https, reqBytes) {
    defer func{
        err = recover()
        if err != nil {
            yakit.Error("handle raw bytes failed: %s", err)
        }
    }

    urlstr = ""
    urlIns, _ = str.ExtractURLFromHTTPRequestRaw(reqBytes, https)
    if urlIns != nil {
        urlstr = urlIns.String()
    }
    rsp, req, _ = poc.HTTP(reqBytes, poc.https(packetHttps))
    header, body = poc.Split(rsp)
    caller.MirrorHTTPFlowEx(true, https, urlstr, reqBytes, rsp, body)
}

// Start getting HTTPFlow ID
if len(flows) > 0 {
    yakit.StatusCard("Scan historical data packets", len(flows))
	for result = range db.QueryHTTPFlowsByID(flows...) {
		// Submit scanning task:
		yakit.Info("Submit scanning task: [%v] %v", result.Method, result.Url)
		handleHTTPFlow(result)
	}
}

if packetRawFile != "" {
    raw, err = file.ReadFile(packetRawFile)
    if err != nil {
        yakit.StatusCard("Reason for failure", "cannot receive parameters", "error")
    }
    handleHTTPRequestRaw(packetHttps, raw)
}
`

func (s *Server) ExecPacketScan(req *ypb.ExecPacketScanRequest, stream ypb.Yak_ExecPacketScanServer) error {
	reqParams := &ypb.ExecRequest{Script: execPacketScanCode}
	var err error

	reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "prod"})

	if len(req.GetHTTPRequest()) > 0 {
		if req.GetHTTPS() {
			reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "packet-https"})
		}
		fp, err := ioutil.TempFile(consts.GetDefaultYakitBaseTempDir(), "scan-packet-file-*.txt")
		if err != nil {
			return utils.Errorf("Failed to create temporary files: %s", err)
		}
		fp.Write(req.GetHTTPRequest())
		fp.Close()
		defer func() {
			os.RemoveAll(fp.Name())
		}()
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "packet-file", Value: fp.Name()})
	}

	// yakit-plugin-file
	var pluginCallback func()
	reqParams.Params, pluginCallback, err = appendPluginNames(reqParams.Params, req.GetPluginList()...)
	if err != nil {
		return utils.Errorf("append plugin names failed: %s", err)
	}
	if pluginCallback != nil {
		defer pluginCallback()
	}

	// httpflow as input
	if len(req.GetHTTPFlow()) > 0 {
		var flows []int = funk.Map(req.GetHTTPFlow(), func(i int64) int {
			return int(i)
		}).([]int)
		flowsStr := utils.ConcatPorts(flows)
		if flowsStr != "" {
			fp, err := consts.TempFile("yakit-packet-scan-httpflow-%v.txt")
			if err != nil {
				return utils.Errorf("cannot create tmp file(for packet-scan): %s", err)
			}
			fp.WriteString(flowsStr)
			log.Infof("start to handle flow(ID): %v", flowsStr)
			fp.Close()
			defer os.RemoveAll(fp.Name())

			reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "flows-file", Value: fp.Name()})
		}
	} else {
		log.Info("no httpflow scanned")
	}

	ctx, cancelCtx := context.WithTimeout(stream.Context(), utils.FloatSecondDuration(float64(req.GetTotalTimeoutSeconds())))
	defer cancelCtx()
	return s.ExecWithContext(ctx, reqParams, stream)
}
