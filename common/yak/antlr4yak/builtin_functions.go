package antlr4yak

import (
	"context"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

// eval executes any Yak code.
// This function has side effects, that is, it can obtain and change the context in the current engine.
// Example:
// ```
// a = 1
// eval("a++")
// assert a == 2
// ```
func (e *Engine) YakBuiltinEval(code string) {
	vm := e.vm
	topFrame := vm.VMStack.Peek().(*yakvm.Frame)
	ctx := topFrame.GetContext()
	if utils.IsNil(ctx) {
		ctx = context.Background()
	}

	codes, err := e.Compile(code)
	if err != nil {
		panic(err)
	}
	if err = e.vm.ExecYakCode(ctx, code, codes, yakvm.Inline); err != nil {
		panic(err)
	}
}

// yakfmt Format any Yak code and return the formatted code
// Example:
// ```
// yakfmt("for { println(`hello yak`) }")
// ```
func (e *Engine) YakBuiltinfmt(code string) string {
	newCode, err := New().FormattedAndSyntaxChecking(code)
	if err != nil {
		log.Errorf("format and syntax checking met error: %s", err)
		return code
	}
	return newCode
}

// yakfmtWithError Format any Yak code and return the formatted code Code and errors
// Example:
// ```
// yakfmtWithError("for { println(`hello yak`) }")
// ```
func (e *Engine) YakBuiltinfmtWithError(code string) (string, error) {
	return New().FormattedAndSyntaxChecking(code)
}

// getScopeInspects Get all variables in the current scope and return ScopeValue Structure reference slice
// Example:
// ```
// a, b = 1, "yak"
// values, err = getScopeInspects()
// for v in values {
// println(v.Value)
// }
// ```
func (e *Engine) YakBuiltinGetScopeInspects() ([]*ScopeValue, error) {
	return e.GetScopeInspects()
}

// getFromScope gets the variables in the current scope and returns the variable value.
// Example:
// ```
// a, b = 1, "yak"
// { assert getFromScope("a") == 1 }
// { assert getFromScope("b") == "yak" }
// ```
func (e *Engine) YakBuiltinGetFromScope(v string, vals ...any) any {
	val, ok := e.GetVar(v)
	if ok {
		return val
	}
	if len(vals) >= 1 {
		return vals[0]
	}
	return nil
}

// waitAllAsyncCallFinish Wait until all asynchronous calls are completed
// Example:
// ```
// a = 0
// for i in 5 {
// go func(i) {
// time.sleep(i)
// a++
// }(i)
// }
// waitAllAsyncCallFinish()
// assert a == 5
// ```
func (e *Engine) waitAllAsyncCallFinish() {
	e.vm.AsyncWait()
}

func InjectContextBuiltinFunction(engine *Engine) {
	engine.ImportLibs(map[string]interface{}{
		"eval":                   engine.YakBuiltinEval,
		"yakfmt":                 engine.YakBuiltinfmt,
		"yakfmtWithError":        engine.YakBuiltinfmtWithError,
		"getScopeInspects":       engine.YakBuiltinGetScopeInspects,
		"getFromScope":           engine.YakBuiltinGetFromScope,
		"waitAllAsyncCallFinish": engine.waitAllAsyncCallFinish,
	})
}
