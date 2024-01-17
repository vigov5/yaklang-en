package yakgrpc

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

const mitmPluginTestCode = `yakit.AutoInitYakit()
loglevel("info")

target = cli.String("target")
pluginName := cli.String("plugin-name")
yakit.Info("Start executing MITM plug-in independently: %v" % pluginName)

if target == "" {
    die("Unable to perform plug-in scanning, the target corresponding to the current plug-in does not exist: %v" % target)
}
yakit.Info("Scanning target: %v" % target)

manager, err = hook.NewMixPluginCaller()
if err != nil {
    yakit.Error("Failed to create plug-in management module: %v", err)
    die("The plug-in management module cannot create")
}

err = manager.LoadPlugin(pluginName)
if err != nil {
    reason = "Unable to load plug-in: %v Reason: %v" % [pluginName, err]
    yakit.Error(reason)
    die(reason)
}
manager.SetDividedContext(true)
manager.SetConcurrent(20)
defer manager.Wait()

res, err = crawler.Start(target, crawler.maxRequest(10),crawler.disallowSuffix([]))
if err != nil {
    reason = "Unable to perform basic crawling: %v" % err
    yakit.Error(reason)
    die(reason)
}

for req = range res {
    yakit.Info("Check URL: %v", req.Url())
    manager.MirrorHTTPFlow(req.IsHttps(), req.Url(), req.RequestRaw(), req.ResponseRaw(), req.ResponseBody())
}
`

func (s *Server) generateMITMTask(pluginName string, ctx ypb.Yak_ExecServer, params []*ypb.ExecParamItem) error {
	params = append(params, &ypb.ExecParamItem{
		Key:   "plugin-name",
		Value: pluginName,
	})
	return s.Exec(&ypb.ExecRequest{
		Params: params,
		Script: mitmPluginTestCode,
	}, ctx)
}

func execTestCaseMITMHooksCaller(rootCtx context.Context, y *yakit.YakScript, params []*ypb.ExecParamItem, db *gorm.DB, streamFeedback func(r *ypb.ExecResult) error) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	manager := yak.NewYakToCallerManager()
	err := manager.AddForYakit(
		ctx, y.ScriptName, params, y.Content,
		yak.YakitCallerIf(func(result *ypb.ExecResult) error {
			return streamFeedback(result)
		}),
		append(enabledHooks, "__test__")...)
	if err != nil {
		log.Errorf("load mitm hooks code failed: %s", err)
		return utils.Errorf("load mitm failed: %s", err)
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Infof("call %v' clear ", y.ScriptName)
			manager.CallByName("clear")
		}
	}()

	log.Infof("call %v' __test__ ", y.ScriptName)
	manager.CallByName("__test__")
	cancel()

	return nil
}
