package dap

import (
	"fmt"
	"github.com/yaklang/yaklang/common/utils/cli"
	"os"
	"path/filepath"

	"github.com/google/go-dap"
	"github.com/yaklang/yaklang/common/yak"
	"github.com/yaklang/yaklang/common/yak/antlr4yak"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func (ds *DebugSession) RunProgramInDebugMode(request *dap.LaunchRequest, debug bool, program string, args []string) {
	raw, err := os.ReadFile(program)
	if err != nil {
		ds.sendErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
			fmt.Sprintf("read file[%s] error: %v", program, err))
		raw = []byte{}
	}

	var absPath = program
	if !filepath.IsAbs(absPath) {
		absPath, err = filepath.Abs(absPath)
		if err != nil {
			ds.sendErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
				fmt.Sprintf("get abs file path[%s] error: %v", program, err))
		}
	}

	engine := yak.NewScriptEngine(100)
	if debug {
		engine.SetDebug(true)
		d := NewDAPDebugger()

		// Wait for initialization
		d.InitWGAdd()

		// Set callback
		engine.SetDebugInit(d.Init())
		engine.SetDebugCallback(d.CallBack())

		d.source = &Source{AbsPath: absPath, Name: filepath.Base(absPath)}

		ds.debugger = d
		d.session = ds
	}
	// launch is completed
	ds.LaunchWg.Done()

	// inject args in cli
	cli.InjectCliArgs(args)

	// inject extra libs
	if len(ds.config.extraLibs) > 0 {
		engine.RegisterEngineHooks(func(engine *antlr4yak.Engine) error {
			engine.ImportLibs(ds.config.extraLibs)
			return nil
		})
	}

	err = engine.ExecuteMain(string(raw), absPath)
	// If it is vmpanic, it has been processed in the debugger
	if err != nil && !yakvm.IsVMPanic(err) {
		ds.sendErrorResponse(request.Request, FailedToLaunch, "Failed to launch",
			fmt.Sprintf("run file[%s] error: %v", absPath, err))
	}
}
