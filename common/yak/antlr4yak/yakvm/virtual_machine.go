package yakvm

import (
	"context"
	"fmt"
	"sync"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm/vmstack"
)

type ExecFlag int

const (
	None   ExecFlag = 1 << iota // By default, a new stack frame is created to execute the code, and the stack is popped after execution.
	Trace                       // does not exit the site after execution.
	Sub                         // from the frame. Use the data on the top of the stack to continue executing
	Inline                      // uses the last executed Trace to continue executing
	Asnyc                       // . Asynchronous execution of
)

func GetFlag(flags ...ExecFlag) ExecFlag {
	flag := None
	for _, f := range flags {
		flag |= f
	}
	return flag
}

type YakitFeedbacker interface{}
type (
	BreakPointFactoryFun func(v *VirtualMachine) bool
	VirtualMachine       struct {
		// globalVar. It is a global variable of the current engine and belongs to the engine
		globalVar map[string]interface{}
		VMStack   *vmstack.Stack
		rootScope *Scope

		// asyncWaitGroup
		asyncWaitGroup *sync.WaitGroup
		// debug
		debug         bool // internal debug
		debugMode     bool // . External debugger
		debugger      *Debugger
		BreakPoint    []BreakPointFactoryFun
		ThreadIDCount uint64
		//
		yakitFeedbacker YakitFeedbacker
		config          *VirtualMachineConfig
		// map[sha1(caller, callee)]func(any)any
		hijackMapMemberCallHandlers sync.Map
		globalVarFallback           func(string) interface{}
		GetExternalVar              func(name string) (any, bool)
	}
)

func (n *VirtualMachine) RegisterMapMemberCallHandler(caller, callee string, h func(interface{}) interface{}) {
	n.hijackMapMemberCallHandlers.Store(utils.CalcSha1(caller, callee), h)
}

func (n *VirtualMachine) RegisterGlobalVariableFallback(h func(string) interface{}) {
	n.globalVarFallback = h
}

func (n *VirtualMachine) SetYaiktFeedbacker(i YakitFeedbacker) {
	n.yakitFeedbacker = i
}

func (v *VirtualMachine) AddBreakPoint(fun BreakPointFactoryFun) {
	v.BreakPoint = append(v.BreakPoint, fun)
}

func (n *VirtualMachine) GetExternalVariableNames() []string {
	vs := []string{}
	for k := range n.globalVar {
		vs = append(vs, k)
	}
	return vs
}

func (v *VirtualMachine) SetDebug(debug bool) {
	v.debug = debug
}

func (v *VirtualMachine) SetDebugMode(debug bool, sourceCode string, codes []*Code, debugInit, debugCallback func(*Debugger)) {
	v.debugMode = debug
	if !debug {
		return
	}
	if v.debugger == nil {
		v.debugger = NewDebugger(v, sourceCode, codes, debugInit, debugCallback)
	}
}

func (v *VirtualMachine) SetSymboltable(table *SymbolTable) {
	v.rootScope = NewScope(table)
}

func (v *VirtualMachine) AsyncStart() {
	v.asyncWaitGroup.Add(1)
}

func (v *VirtualMachine) AsyncEnd() {
	v.asyncWaitGroup.Done()
}

func (v *VirtualMachine) AsyncWait() {
	v.asyncWaitGroup.Wait()
}

func NewWithSymbolTable(table *SymbolTable) *VirtualMachine {
	v := &VirtualMachine{
		// rootSymbol: table,
		rootScope: NewScope(table),
		VMStack:   vmstack.New(),
		globalVar: make(map[string]interface{}),
		config:    NewVMConfig(),
		// asyncWaitGroup
		asyncWaitGroup: new(sync.WaitGroup),
		// debug
		ThreadIDCount: 1, // is initially 1.
	}
	return v
}

func New() *VirtualMachine {
	return NewWithSymbolTable(NewSymbolTable())
}

// deepCopyLib copies yaklang dependencies to prevent multiple When the engine is running in parallel, hooking the lib causes a concurrent write map error.
func deepCopyLib(libs map[string]interface{}) map[string]interface{} {
	newLib := map[string]interface{}{}
	for k, v := range libs {
		if v1, ok := v.(map[string]interface{}); ok {
			newLib[k] = deepCopyLib(v1)
		} else {
			newLib[k] = v
		}
	}
	return newLib
}

// ImportLibs Import the library into the global variables of the engine. When
func (n *VirtualMachine) ImportLibs(libs map[string]interface{}) {
	for k, v := range deepCopyLib(libs) {
		n.globalVar[k] = v
	}
}

// SetVar imports variables into the global variables of the engine.
func (n *VirtualMachine) SetVar(k string, v interface{}) {
	n.globalVar[k] = v
}

func (n *VirtualMachine) GetVar(name string) (interface{}, bool) {
	ivm := n.VMStack.Peek()
	if ivm == nil {
		val, ok := n.rootScope.GetValueByName(name)
		if ok {
			return val.Value, true
		}
	} else {
		// ivm exists, find the variable
		val, ok := ivm.(*Frame).CurrentScope().GetValueByName(name)
		if ok {
			return val.Value, true
		}
	}
	// . If it does not exist and the root cannot be found, then directly find
	var_, ok := n.globalVar[name]
	if ok {
		return var_, true
	}

	if n.globalVarFallback != nil {
		hijackedGlobal := n.globalVarFallback(name)
		if hijackedGlobal != nil {
			return hijackedGlobal, true
		}
	}

	return undefined, false
}

func (n *VirtualMachine) GetGlobalVar() map[string]interface{} {
	return n.globalVar
}

func (n *VirtualMachine) GetDebugger() *Debugger {
	return n.debugger
}

func (v *VirtualMachine) ExecYakFunction(ctx context.Context, f *Function, args map[int]*Value, flags ...ExecFlag) (interface{}, error) {
	var value interface{}
	finalFlags := []ExecFlag{Sub}
	if len(flags) > 0 {
		finalFlags = flags
	}
	err := v.Exec(ctx, func(frame *Frame) {
		name := f.GetActualName()
		frame.SetVerbose(fmt.Sprintf("function: %s", name))
		frame.SetFunction(f)
		if f.sourceCode != "" {
			frame.SetOriginCode(f.sourceCode)
		}
		// globally. The closure inherits the parent scope
		// if v.config.GetClosureSupport() {
		frame.scope = f.scope
		frame.CreateAndSwitchSubScope(f.symbolTable)
		//} else {
		//	frame.scope = NewScope(f.symbolTable)
		//}
		for id, arg := range args {
			frame.CurrentScope().NewValueByID(id, arg)
		}
		frame.Exec(f.codes)
		if frame.lastStackValue != nil {
			value = frame.lastStackValue.Value
		}
		// if v.config.GetClosureSupport() {
		frame.ExitScope()
		//}
	}, finalFlags...)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (v *VirtualMachine) ExecAsyncYakFunction(ctx context.Context, f *Function, args map[int]*Value) error {
	return v.Exec(ctx, func(frame *Frame) {
		name := f.GetActualName()
		frame.SetVerbose("function: " + name)
		frame.SetFunction(f)
		frame.SetScope(f.scope)
		frame.CreateAndSwitchSubScope(f.symbolTable)
		for id, arg := range args {
			frame.CurrentScope().NewValueByID(id, arg)
		}
		go func() {
			defer func() {
				v.AsyncEnd()
				if err := recover(); err != nil {
					log.Errorf("yakvm async function panic: %v", err)
					utils.PrintCurrentGoroutineRuntimeStack()
				}
			}()

			frame.Exec(f.codes)
			frame.ExitScope()
		}()
	}, Sub, Asnyc)
}

func (v *VirtualMachine) ExecYakCode(ctx context.Context, sourceCode string, codes []*Code, flags ...ExecFlag) error {
	return v.Exec(ctx, func(frame *Frame) {
		frame.SetVerbose("__yak_main__")
		frame.SetOriginCode(sourceCode)
		frame.Exec(codes)
	}, flags...)
}

func (v *VirtualMachine) InlineExecYakCode(ctx context.Context, codes []*Code, flags ...ExecFlag) error {
	return v.Exec(ctx, func(frame *Frame) {
		frame.Exec(codes)
	}, Trace|Sub)
}

var vmstackLock = new(sync.Mutex)

func (v *VirtualMachine) Exec(ctx context.Context, f func(frame *Frame), flags ...ExecFlag) error {
	flag := GetFlag(flags...)

	var frame *Frame
	if flag&Sub == Sub {

		vmstackLock.Lock()
		topFrame := v.VMStack.Peek()
		vmstackLock.Unlock()

		if topFrame == nil {
			panic("BUG: VMStack is empty(Sub)")
		}
		frame = NewSubFrame(topFrame.(*Frame))
	} else if flag&Inline == Inline {
		vmstackLock.Lock()
		topFrame := v.VMStack.Peek()
		vmstackLock.Unlock()

		if topFrame == nil {
			topFrame = NewFrame(v)
			vmstackLock.Lock()
			v.VMStack.Push(topFrame)
			vmstackLock.Unlock()
			log.Debugf("VMStack is empty(Inline), we create new frame")
		}

		frame = topFrame.(*Frame)
		codes := frame.codes
		p := frame.codePointer
		defer func() {
			frame.codes = codes
			frame.codePointer = p
		}()
	} else {
		frame = NewFrame(v)
		for k, val := range v.globalVar {
			frame.GlobalVariables[k] = val
		}
	}

	if flag&Asnyc == Asnyc {
		frame.coroutine = NewCoroutine()
	}

	vmstackLock.Lock()
	v.VMStack.Push(frame)
	vmstackLock.Unlock()

	frame.debug = v.debug
	// . Initialize debugger
	if v.debugMode && v.debugger != nil && v.debugger.initFunc != nil {
		v.debugger.InitCallBack()
	}
	frame.ctx = ctx

	f(frame)

	// . When Trace is not set, it will exit after execution.
	if flag&Trace != Trace {
		vmstackLock.Lock()
		v.VMStack.Pop()
		vmstackLock.Unlock()
	}
	if lastPanic := frame.recover(); lastPanic != nil {
		lastPanic.contextInfos.Peek().(*PanicInfo).SetPositionVerbose(frame.GetVerbose())
		if exitValue, ok := lastPanic.data.(*VMPanicSignal); ok {
			panic(exitValue)
		} else {
			panic(lastPanic)
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func (v *VirtualMachine) CurrentFM() *Frame {
	return v.VMStack.Peek().(*Frame)
}

func (v *VirtualMachine) GetConfig() *VirtualMachineConfig {
	return v.config
}

func (v *VirtualMachine) SetConfig(config *VirtualMachineConfig) {
	v.config = config
}
