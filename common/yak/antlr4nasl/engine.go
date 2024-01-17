package antlr4nasl

import (
	"context"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/yaklang/yaklang/common/fp"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4nasl/lib"
	"github.com/yaklang/yaklang/common/yak/antlr4nasl/visitors"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
	"github.com/yaklang/yaklang/embed"
	"os"
	"path"
	"path/filepath"
)

type Engine struct {
	debug                          bool
	naslLibsPath, dependenciesPath string
	naslLibPatch                   map[string]func(string) string
	compiler                       *visitors.Compiler
	vm                             *yakvm.VirtualMachine
	description                    bool
	sourceCode                     string
	scriptObj                      *NaslScriptInfo
	host                           string
	proxies                        []string
	Kbs                            *NaslKBs
	//scriptExecMutexs               map[string]*sync.Mutex
	//scriptExecMutexsLock           *sync.Mutex
	buildInMethodHook map[string]func(origin NaslBuildInMethod, engine *Engine, params *NaslBuildInMethodParam) (interface{}, error)
	preferences       map[string]any
	logger            *log.Logger
}

func NewWithKbs(kbs *NaslKBs) *Engine {
	table := yakvm.NewSymbolTable()
	vm := yakvm.NewWithSymbolTable(table)
	vm.GetConfig().SetClosureSupport(false)
	vm.GetConfig().SetFunctionNumberCheck(false)
	vm.GetConfig().SetYVMMode(yakvm.NASL)
	engine := &Engine{
		compiler:     visitors.NewCompilerWithSymbolTable(table),
		vm:           vm,
		naslLibPatch: make(map[string]func(string) string),
		Kbs:          kbs,
		logger:       log.GetLogger("NASL Logger"),
		//loadedScriptsLock: &sync.Mutex{},
		buildInMethodHook: make(map[string]func(origin NaslBuildInMethod, engine *Engine, params *NaslBuildInMethodParam) (interface{}, error)),
		//scriptExecMutexs:  make(map[string]*sync.Mutex),
	}
	engine.compiler.SetNaslLib(GetNaslLibKeys())
	engine.compiler.RegisterVisitHook("a", func(c *visitors.Compiler, ctx antlr.ParserRuleContext) {
		if start := ctx.GetStart(); start != nil {
			c.SetStartPosition(start.GetLine(), start.GetColumn())
		}
		if end := ctx.GetStop(); end != nil {
			c.SetStopPosition(end.GetLine(), end.GetColumn())
		}
	})
	vm.SetVar("__method_proxy__", func(params [][]interface{}) interface{} {
		var funName string
		if params != nil && len(params) > 0 && len(params[0]) == 1 {
			if v, ok := params[0][0].(int); ok {
				name, ok := engine.compiler.GetSymbolTable().GetNameByVariableId(v)
				if ok {
					funName = name
				}
			}
		}
		fn := NaslLib[funName]
		naslParams := &NaslBuildInMethodParam{
			mapParams: make(map[string]*yakvm.Value),
		}
		for _, p := range params[1:] {
			name, ok := engine.compiler.GetSymbolTable().GetNameByVariableId(p[0].(int))
			if ok {
				naslParams.mapParams[name] = yakvm.NewAutoValue(p[1])
			}
			naslParams.listParams = append(naslParams.listParams, yakvm.NewAutoValue(p[1]))
		}
		return fn(engine, naslParams)
		//panic("call build in method error: not found symbol id")
	})
	vm.SetVar("__OpCallCallBack__", func(name string) {
		// from the built-in file. Do some function debugging work.
		if name == "http_recv_headers2" {
			print()
		}
	})
	vm.SetVar("__nasl_global_var_table", make(map[int]*yakvm.Value))
	vm.ImportLibs(lib.NaslBuildInNativeMethod)
	engine.scriptObj = NewNaslScriptObject()
	return engine
}
func New() *Engine {
	return NewWithKbs(NewNaslKBs())
}

//	func (engine *Engine) GetScriptMuxByName(name string) *sync.Mutex {
//		engine.scriptExecMutexsLock.Lock()
//		defer engine.scriptExecMutexsLock.Unlock()
//		if v, ok := engine.scriptExecMutexs[name]; ok {
//			return v
//		}
//		engine.scriptExecMutexs[name] = &sync.Mutex{}
//		return engine.scriptExecMutexs[name]
//	}
func (engine *Engine) RegisterBuildInMethodHook(name string, hook func(origin NaslBuildInMethod, engine *Engine, params *NaslBuildInMethodParam) (interface{}, error)) {
	engine.buildInMethodHook[name] = hook
}
func (engine *Engine) UnRegisterBuildInMethodHook(name string) {
	delete(engine.buildInMethodHook, name)
}

//	func (e *Engine) MarkScriIsLoadptIsLoaded(scriptName string) {
//		e.loadedScriptsLock.Lock()
//		defer e.loadedScriptsLock.Unlock()
//		e.loadedScripts[scriptName] = struct{}{}
//	}
//
//	func (e *Engine) IsScriptLoaded(scriptName string) bool {
//		e.loadedScriptsLock.Lock()
//		defer e.loadedScriptsLock.Unlock()
//		_, ok := e.loadedScripts[scriptName]
//		return ok
//	}
func (engine *Engine) SetProxies(proxies ...string) {
	engine.proxies = proxies
}
func (engine *Engine) GetScriptObject() *NaslScriptInfo {
	return engine.scriptObj
}
func (engine *Engine) GetKBData() map[string]interface{} {
	return engine.Kbs.GetData()
}
func (engine *Engine) SetIncludePath(path string) {
	engine.naslLibsPath = path
}
func (engine *Engine) SetDependenciesPath(path string) {
	engine.dependenciesPath = path
}
func (engine *Engine) Debug(bool2 ...bool) {
	if len(bool2) == 0 {
		engine.debug = true
	} else {
		engine.debug = bool2[0]
	}
}

func (engin *Engine) CallNativeFunction(name string, mapParam map[string]interface{}, sliceParam []interface{}) (interface{}, error) {
	params := NewNaslBuildInMethodParam()
	for _, i1 := range sliceParam {
		params.listParams = append(params.listParams, yakvm.NewAutoValue(i1))
	}
	for k, v := range mapParam {
		params.mapParams[k] = yakvm.NewAutoValue(v)
	}
	if fn, ok := NaslLib[name]; ok {
		return fn(engin, params), nil
	}
	return nil, utils.Errorf("not found build in method: %s", name)

}

func (engine *Engine) SetKBs(kbs *NaslKBs) {
	engine.Kbs = kbs
}

func (engine *Engine) ServiceScan(target string, ports string) ([]*fp.MatchResult, error) {
	return ServiceScan(target, ports)
}
func (engine *Engine) InitBuildInLib() {
	engine.vm.ImportLibs(lib.NaslBuildInNativeMethod)
	engine.vm.ImportLibs(lib.BuildInVars)
}
func (e *Engine) Compile(code string) error {
	e.compiler.SetExternalVariableNames(e.vm.GetExternalVariableNames())
	e.compiler.Debug(e.debug)
	ok := e.compiler.Compile(code)
	if !ok {
		return e.compiler.GetMergeError()
	}
	return nil
}
func (e *Engine) IsDebug() bool {
	return e.debug
}
func (e *Engine) SafeRunFile(path string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = utils.Error(e)
		}
	}()
	return e.RunFile(path)
}
func (e *Engine) RunScript(script *NaslScriptInfo) error {
	e.logger.Debugf("Running script %s", script.OriginFileName)
	e.scriptObj = script
	return e.safeEvalWithFileName(script.Script, script.OriginFileName)
}

func (e *Engine) EvalInclude(name string) error {
	// from the built-in file. First search from the local file. Otherwise, search for
	var sourceBytes []byte
	libPath := path.Join(e.naslLibsPath, name)
	codes, err := os.ReadFile(libPath)
	if err == nil {
		sourceBytes = codes
	}
	//Local file loading failed. Load
	if sourceBytes == nil {
		data, err := embed.Asset("data/nasl-incs/" + name)
		if err != nil {
			err = utils.Errorf("not found include file: %s", name)
			e.logger.Error(err)
			return err
		}
		sourceBytes = data
	}
	return e.safeEvalWithFileName(string(sourceBytes), name)
}
func (e *Engine) LoadScript(path string) (*NaslScriptInfo, error) {
	e.SetDescription(true)
	oldIns := e.GetScriptObject()
	defer func() {
		e.SetDescription(false)
		e.scriptObj = oldIns
	}()
	e.scriptObj = NewNaslScriptObject()
	e.scriptObj.OriginFileName = filepath.Base(path)
	code, err := os.ReadFile(path)
	if err != nil {
		script, err := NewNaslScriptObjectFromDb(e.scriptObj.OriginFileName)
		if err != nil {
			return nil, utils.Errorf("not found script file: %s", path)
		}
		return script, err
	} else {
		err = e.safeEvalWithFileName(string(code), e.scriptObj.OriginFileName)
		return e.scriptObj, err
	}
}
func (e *Engine) RunFile(path string) error {
	e.scriptObj.OriginFileName = filepath.Base(path)
	code, err := os.ReadFile(path)
	if err != nil {
		script, err := NewNaslScriptObjectFromDb(e.scriptObj.OriginFileName)
		if err != nil {
			return utils.Errorf("not found script file: %s", path)
		}
		return e.RunScript(script)
	} else {
		recoverSource := e.compiler.SetSourceCodeFilePath(path)
		defer recoverSource()
		return e.safeEvalWithFileName(string(code), e.scriptObj.OriginFileName)
	}

}

func (e *Engine) Eval(code string) error {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(*yakvm.VMPanicSignal); ok {
				e.logger.Infof("script [%s] exit with value: %v", e.scriptObj.OriginFileName, v.Info)
				if e.debug {
					e.logger.Infof("script additional info: %v", v.AdditionalInfo)
				}
			} else {
				panic(err)
			}
		}
	}()
	recoverCode := e.compiler.SetSourceCode(code)
	defer func() { recoverCode() }()
	e.sourceCode = code
	e.scriptObj.Script = code
	e.vm.SetVar("__this__", e.scriptObj)
	e.vm.SetVar("description", e.description)
	err := e.Compile(code)
	if err != nil {
		return err
	}
	err = e.vm.ExecYakCode(context.Background(), code, e.compiler.GetCodes(), yakvm.None)
	if err != nil {
		return err
	}
	if e.description {
		if e.scriptObj != nil && e.scriptObj.OID != "" {
			return e.scriptObj.Save()
		}
	}
	return nil
}
func (e *Engine) safeEvalWithFileName(code string, fileName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = utils.Error(e)
			}
		}
	}()
	if fileName != "" {
		if v, ok := e.naslLibPatch[fileName]; ok {
			code = v(code)
		}
	}
	recoverFunc := e.compiler.SetSourceCodeFilePath(fileName)
	defer recoverFunc()
	err = e.Eval(code)
	return
}
func (e *Engine) SafeEval(code string) (err error) {
	return e.safeEvalWithFileName(code, "")
}
func (e *Engine) AddNaslLibPatch(lib string, handle func(string2 string) string) {
	e.naslLibPatch[lib] = handle
}

func (e *Engine) GetVirtualMachine() *yakvm.VirtualMachine {
	return e.vm
}
func (e *Engine) GetCompiler() *visitors.Compiler {
	return e.compiler
}

func (e *Engine) SetDescription(b bool) {
	e.description = b
}
