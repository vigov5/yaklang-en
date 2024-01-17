package yakvm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm/vmstack"

	"github.com/pkg/errors"
)

var (
	// injected by the yakast package
	YakDebugCompiler CompilerWrapperInterface
)

type LinesFirstCodeStateMap = map[int]*CodeState
type BreakpointMap = map[int]*Breakpoint
type ObserveBreakPointMap = map[string]*Value

type switchBundle struct {
	linesFirstCodeStateMap LinesFirstCodeStateMap
	breakpointMap          BreakpointMap
	observeBreakPointMap   ObserveBreakPointMap
}

func NewSwitchBundle() *switchBundle {
	return &switchBundle{
		linesFirstCodeStateMap: make(LinesFirstCodeStateMap),
		breakpointMap:          make(BreakpointMap),
		observeBreakPointMap:   make(ObserveBreakPointMap),
	}
}

func GetCodeNumber(code *Code) (int, int, int, int) {
	startLine, startColumn := code.StartLineNumber, code.StartColumnNumber
	endLine, endColumn := code.EndLineNumber, code.EndColumnNumber
	// Special processing for ScopeEnd
	if code.Opcode == OpScopeEnd {
		startLine, startColumn = endLine, endColumn
	}
	return startLine, startColumn, endLine, endColumn
}

type Debugger struct {
	vm           *VirtualMachine
	once         sync.Once
	startWG      sync.WaitGroup  // Used to wait for the program to start
	started      bool            // indicates whether the program has been started.
	finished     bool            // indicates whether the program has ended
	wg           sync.WaitGroup  // When multiple asynchronous functions are executed at the same time, callback breakpoints block the execution of
	initFunc     func(*Debugger) // Initialization function
	callbackFunc func(*Debugger) // breakpoint callback function

	description string     // . Callback information
	frame       *Frame     // Store the currently executed frame
	state       string     // indicates which function
	lock        sync.Mutex // is used for ShouldCallback synchronization

	sourceFilePath                string
	sourceCode                    string
	sourceCodeLines               []string
	codes                         map[string][]*Code // state -> []code
	codePointer                   int
	linePointer                   int
	currentLinesFirstCodeStateMap LinesFirstCodeStateMap // The first opcode index of each line

	// breakpoint
	breakPointCount      int32
	currentBreakPointMap BreakpointMap // . Line -> breakpoint

	// is used to step over, step into, step out
	jmpState *DebuggerState
	// stepOut      bool
	nextState    *DebuggerState
	stepInState  *DebuggerState
	stepoutState *DebuggerState

	// Stop
	halt bool

	// Stop event reason
	stopReason string

	// panic
	vmPanic *VMPanic

	// Stack trace
	StackTraces      map[int]*vmstack.Stack // threadID -> stacktraces
	ThreadStackTrace map[int]*DebuggerState // . The current stackTrace corresponding to each thread.

	// Reference, which is used to store information about frames, scopes, and variable references.
	Reference *Reference

	// Observation breakpoint
	currentObserveBreakPointMap ObserveBreakPointMap

	// Observe expression
	observeExpressions map[string]*Value

	// switch bundle, used to switch between multiple files. LinesFirstCodeStateMap, BreakpointMap, ObserveBreakPointMap
	switchBundleMap map[string]*switchBundle
}

type DebuggerState struct {
	code                 *Code
	lineIndex, codeIndex int
	stackLen             int
	stateName            string
	frame                *Frame
}
type CodeState struct {
	codeIndex int
	state     string
}

type StackTrace struct {
	ID   int
	Name string

	Frame *Frame

	Source             *string
	SourceCode         *string
	Line, Column       int
	EndLine, EndColumn int
}

type StackTraces struct {
	ThreadID    int
	StackTraces []StackTrace
}

func NewDebugger(vm *VirtualMachine, sourceCode string, codes []*Code, init, callback func(*Debugger)) *Debugger {

	debugger := &Debugger{
		started:          false,
		finished:         false,
		StackTraces:      map[int]*vmstack.Stack{int(vm.ThreadIDCount): vmstack.New()},
		ThreadStackTrace: make(map[int]*DebuggerState, 0),
		vm:               vm,
		startWG:          sync.WaitGroup{},
		wg:               sync.WaitGroup{},
		initFunc:         init,
		callbackFunc:     callback,
		sourceCode:       sourceCode,
		sourceCodeLines:  strings.Split(strings.ReplaceAll(sourceCode, "\r", ""), "\n"),
		codes:            make(map[string][]*Code),
		linePointer:      0,
		switchBundleMap:  make(map[string]*switchBundle),

		Reference:          NewReference(),
		observeExpressions: make(map[string]*Value),
	}
	debugger.Init(codes)
	return debugger
}

func NewCodeState(codeIndex int, state string) *CodeState {
	return &CodeState{
		codeIndex: codeIndex,
		state:     state,
	}
}

func (g *Debugger) InitCode(codes []*Code) {
	g.initCode(codes, 1)
}

func (g *Debugger) initCode(codes []*Code, depth int) {
	// Prevent stack explosion
	if depth >= 100000 {
		return
	}
	// Find all functions and their opcodes
	for _, code := range codes {
		if code.Opcode == OpPush {
			v := code.Op1
			if !v.IsYakFunction() {
				continue
			}
			f, _ := v.Value.(*Function)
			funcUUID := f.GetUUID()
			if _, ok := g.codes[funcUUID]; ok {
				continue
			}

			g.codes[funcUUID] = f.codes
			g.initCode(f.codes, depth+1)
		}
	}
}

func (g *Debugger) Init(codes []*Code) {
	g.StartWGAdd()

	g.codes[""] = codes
	g.InitCode(codes)

	hasSet := false
	sourceCodeFilePath := ""
	var (
		currentBundle *switchBundle
	)

	for state, codes := range g.codes {
		for index, code := range codes {
			newSourceCodeFilePath := sourceCodeFilePath
			if code.SourceCodeFilePath != nil {
				newSourceCodeFilePath = *code.SourceCodeFilePath
			}
			if newSourceCodeFilePath != sourceCodeFilePath {
				sourceCodeFilePath = newSourceCodeFilePath
				currentBundle = g.getSwitchBundle(sourceCodeFilePath)
			}

			linesFirstCodeStateMap := currentBundle.linesFirstCodeStateMap

			if _, ok := linesFirstCodeStateMap[code.StartLineNumber]; !ok {
				linesFirstCodeStateMap[code.StartLineNumber] = NewCodeState(index, state)
			}

			// Set currentLinesFirstCodeStateMap and sourceFilePath
			// . Use a relatively stupid method to find the incoming sourceCode bound to the code. The first code with the same sourceCode
			if !hasSet && code.SourceCodePointer != nil && *code.SourceCodePointer == g.sourceCode {
				hasSet = true
				g.SwitchByOtherFileOpcode(code)
			}
		}
	}

	if !hasSet {
		panic(errors.New("debugger init error: can't find source code in opcodes"))
	}
}

func (g *Debugger) InitCallBack() {
	g.once.Do(func() {
		g.initFunc(g)
	})
}

func (g *Debugger) SwitchByOtherFileOpcode(code *Code) {
	if code.SourceCodeFilePath == nil {
		return
	}
	newFilePath := *code.SourceCodeFilePath

	if newFilePath != g.sourceFilePath {
		g.sourceFilePath = newFilePath
		g.sourceCode = *code.SourceCodePointer
		g.sourceCodeLines = strings.Split(strings.ReplaceAll(g.sourceCode, "\r", ""), "\n")

		currentBundle := g.getSwitchBundle(newFilePath)

		// Modify currentLinesFirstCodeStateMap , currentBreakPointMap, currentObserveBreakPointMap
		g.currentLinesFirstCodeStateMap = currentBundle.linesFirstCodeStateMap
		g.currentBreakPointMap = currentBundle.breakpointMap
		g.currentObserveBreakPointMap = currentBundle.observeBreakPointMap
	}
}

func (g *Debugger) getBreakpointMapBySource(path string) map[int]*Breakpoint {
	return g.getSwitchBundle(path).breakpointMap
}

func (g *Debugger) getSwitchBundle(path string) *switchBundle {
	if path == "" {
		return g.switchBundleMap[g.sourceFilePath]
	}
	path = strings.ToLower(path)
	bundle, ok := g.switchBundleMap[path]
	if !ok {
		bundle = NewSwitchBundle()
		g.switchBundleMap[path] = bundle
	}
	return bundle
}

func (g *Debugger) StartWGAdd() {
	g.startWG.Add(1)
}

func (g *Debugger) StartWGDone() {
	g.startWG.Done()
}

func (g *Debugger) StartWGWait() {
	g.startWG.Wait()
}

func (g *Debugger) Wait() {
	g.wg.Wait()
}

func (g *Debugger) Add() {
	g.wg.Add(1)
}

func (g *Debugger) WaitGroupDone() {
	g.wg.Done()
}

func (g *Debugger) Finished() bool {
	return g.finished
}

func (g *Debugger) SetFinished() {
	g.description = "The program is finished"
	g.finished = true
}

func (g *Debugger) CurrentCodeIndex() int {
	return g.codePointer
}

func (g *Debugger) CurrentLine() int {
	return g.linePointer
}

func (g *Debugger) Breakpoints() map[int]*Breakpoint {
	return g.currentBreakPointMap
}

func (g *Debugger) SourceCodeLines() []string {
	return g.sourceCodeLines
}

func (g *Debugger) InRootState() bool {
	return g.State() == ""
}

func (g *Debugger) StateName() string {
	frame := g.frame
	stateName := "__yak_main__"
	if frame == nil {
		return "unknown"
	}
	if f := frame.GetFunction(); f != nil {
		stateName = f.GetActualName()
	}
	return stateName
}

func (g *Debugger) State() string {
	return g.state
}

func (g *Debugger) UpdateByFrame(frame *Frame) {
	if f := frame.GetFunction(); f != nil {
		g.state = f.GetUUID()
	} else {
		g.state = ""
	}
	g.frame = frame
	g.AddFrameRef(frame)
}

func (g *Debugger) CurrentStackTrace() *vmstack.Stack {
	var (
		st *vmstack.Stack
		ok bool
	)
	frame := g.frame
	if frame == nil {
		return nil
	}

	if st, ok = g.StackTraces[g.frame.ThreadID]; !ok {
		st = vmstack.New()
		g.StackTraces[g.frame.ThreadID] = st
	}
	return st
}

func (g *Debugger) CurrentThreadID() int {
	frame := g.frame
	if frame == nil {
		return 0
	}
	return g.frame.ThreadID
}

func (g *Debugger) CurrentFrameID() int {
	frame := g.frame
	if frame == nil {
		return 0
	}
	ref := g.Reference
	i, ok := ref.FrameHM.getReverse(frame)
	if !ok {
		return 0
	}
	return i
}

func (g *Debugger) Codes() []*Code {
	return g.codes[g.State()]
}

func (g *Debugger) CodesInState(state string) []*Code {
	return g.codes[state]
}

func (g *Debugger) VM() *VirtualMachine {
	return g.vm
}
func (g *Debugger) Frame() *Frame {
	return g.frame
}

func (g *Debugger) Description() string {
	return g.description
}

func (g *Debugger) SetDescription(desc string) {
	g.description = desc
}

func (g *Debugger) ResetDescription() {
	g.description = ""
}

func (g *Debugger) StopReason() string {
	return g.stopReason
}

func (g *Debugger) SetStopReason(desc string) {
	g.stopReason = desc
}

func (g *Debugger) ResetStopReason() {
	g.stopReason = ""
}

func (g *Debugger) VMPanic() *VMPanic {
	return g.vmPanic
}

func (g *Debugger) SetVMPanic(p *VMPanic) {
	g.vmPanic = p
}

func (g *Debugger) AddFrameRef(frame *Frame) int {
	ref := g.Reference

	if i, ok := ref.FrameHM.getReverse(frame); !ok {
		return ref.FrameHM.create(frame)
	} else {
		return i
	}
}

func (g *Debugger) AddBreakPointRef(b *Breakpoint) int {
	ref := g.Reference

	if i, ok := ref.BreakPointHM.getReverse(b); !ok {
		return ref.BreakPointHM.create(b)
	} else {
		return i
	}
}
func (g *Debugger) ForceSetVariableRef(id int, v interface{}) {
	ref := g.Reference
	ref.VarHM.forceSet(id, v)
}

func (g *Debugger) AddVariableRef(v interface{}) int {
	ref := g.Reference
	if i, ok := ref.VarHM.getReverse(v); !ok {
		return ref.VarHM.create(v)
	} else {
		return i
	}
}

func (g *Debugger) AddScopeRef(scope *Scope) int {
	ref := g.Reference
	if i, ok := ref.VarHM.getReverse(scope); !ok {
		return ref.VarHM.create(scope)

	} else {
		return i
	}
}

func (g *Debugger) Pause() {
	g.halt = true
}

func (g *Debugger) IsPause() bool {
	return g.halt
}

func (g *Debugger) GetCode(state string, codeIndex int) *Code {
	codes := g.CodesInState(state)
	if codeIndex < 0 || codeIndex >= len(codes) {
		return nil
	}
	return codes[codeIndex]
}

func (g *Debugger) GetLineFirstCode(lineIndex int) (*Code, int, string) {
	if codeState, ok := g.currentLinesFirstCodeStateMap[lineIndex]; ok {
		return g.GetCode(codeState.state, codeState.codeIndex), codeState.codeIndex, codeState.state
	} else {
		return nil, -1, ""
	}
}

func (g *Debugger) debuggerStateToStackTrace(state *DebuggerState) StackTrace {
	frame := state.frame
	fid, ok := g.Reference.FrameHM.getReverse(frame)
	if !ok {
		fid = -1
	}
	code := state.code
	startLine, startColumn, endLine, endColumn := GetCodeNumber(code)

	return StackTrace{
		ID:         fid,
		Name:       state.stateName,
		Frame:      frame,
		SourceCode: code.SourceCodePointer,
		Source:     code.SourceCodeFilePath,
		Line:       startLine,
		Column:     startColumn,
		EndLine:    endLine,
		EndColumn:  endColumn,
	}
}

func (g *Debugger) GetStackTraces() map[int]*StackTraces {
	stackTrace := g.CurrentStackTrace()
	if stackTrace == nil {
		return nil
	}
	if g.linePointer == 0 {
		return nil
	}

	ret := make(map[int]*StackTraces, len(g.StackTraces))

	for threadID, stack := range g.StackTraces {

		ret[threadID] = &StackTraces{
			ThreadID: threadID,
		}

		sts := make([]StackTrace, stack.Len()+1)

		// Add ThreadStackTrace
		if state, ok := g.ThreadStackTrace[threadID]; ok && state.code != nil {
			sts[0] = g.debuggerStateToStackTrace(state)
		}

		index2 := 1
		stack.GetAll(func(i any) {
			if state, ok := i.(*DebuggerState); ok {
				if state.code != nil {
					sts[index2] = g.debuggerStateToStackTrace(state)
					index2++
				}
			}
		})

		ret[threadID].StackTraces = sts
	}

	return ret
}

func (g *Debugger) AddObserveBreakPoint(expr string) error {
	frame := g.frame
	g.currentObserveBreakPointMap[expr] = undefined
	if frame != nil {
		_, _, err := g.Compile(expr)
		if err != nil {
			return errors.Wrap(err, "add observe breakpoint error")
		}
	}
	return nil
}

func (g *Debugger) RemoveObserveBreakPoint(expr string) error {
	if _, ok := g.currentObserveBreakPointMap[expr]; ok {
		delete(g.currentObserveBreakPointMap, expr)
		return nil
	}

	return utils.Errorf("expression [%s] not in observe breakpoint", expr)
}

func (g *Debugger) AddObserveExpression(expr string) error {
	_, _, err := g.Compile(expr)
	if err != nil {
		return errors.Wrap(err, "add observe expression error")
	}
	g.observeExpressions[expr] = undefined
	return nil
}

func (g *Debugger) RemoveObserveExpression(expr string) error {
	if _, ok := g.observeExpressions[expr]; ok {
		delete(g.observeExpressions, expr)
		return nil
	}

	return utils.Errorf("expression [%s] not in observe expression", expr)
}

func (g *Debugger) GetAllObserveExpressions() map[string]*Value {
	return g.observeExpressions
}

func (g *Debugger) GetAllObserveBreakPoint() map[string]*Value {
	return g.currentObserveBreakPointMap
}

func (g *Debugger) addBreakPoint(path string, codeIndex, lineIndex int, condition, hitCondition, state string) (int, error) {
	breakpointMap := g.getBreakpointMapBySource(path)
	if _, ok := breakpointMap[lineIndex]; !ok {
		bp := g.NewBreakPoint(codeIndex, lineIndex, condition, hitCondition, state)
		breakpointMap[lineIndex] = bp
		ref := g.AddBreakPointRef(bp)
		return ref, nil
	} else {
		return -1, errors.Errorf("breakpoint already exists in line %d", lineIndex)
	}
}

func (g *Debugger) ExistBreakPointInLineWithSource(path string, lineIndex int) (*Breakpoint, bool) {
	breakpointMap := g.getBreakpointMapBySource(path)
	if bp, ok := breakpointMap[lineIndex]; ok {
		return bp, true
	} else {
		return nil, false
	}
}

func (g *Debugger) SetBreakPoint(lineIndex int, condition, hitCondition string) (int, error) {
	code, codeIndex, state := g.GetLineFirstCode(lineIndex)
	if code == nil {
		return -1, utils.Errorf("Can't set breakPoint in line %d", lineIndex)
	} else {
		return g.addBreakPoint("", codeIndex, lineIndex, condition, hitCondition, state)
	}
}

func (g *Debugger) SetBreakPointWithSource(path string, lineIndex int, condition, hitCondition string) (int, error) {
	code, codeIndex, state := g.GetLineFirstCode(lineIndex)
	if code == nil {
		return -1, utils.Errorf("Can't set breakPoint in line %d", lineIndex)
	} else {
		return g.addBreakPoint(path, codeIndex, lineIndex, condition, hitCondition, state)
	}
}

func (g *Debugger) SetNormalBreakPoint(lineIndex int) (int, error) {
	return g.SetBreakPoint(lineIndex, "", "")
}

func (g *Debugger) ClearAllBreakPoints() {
	g.currentBreakPointMap = make(map[int]*Breakpoint, 0)
}

func (g *Debugger) ClearBreakpointsInLine(lineIndex int) {
	if _, ok := g.currentBreakPointMap[lineIndex]; ok {
		delete(g.currentBreakPointMap, lineIndex)
	}
}

func (g *Debugger) ClearOtherBreakpointsWithSource(path string, existLines []int) {
	breakpointMap := g.getBreakpointMapBySource(path)
	// Clear all Breakpoints except existLines
	for lineIndex := range breakpointMap {
		if !lo.Contains(existLines, lineIndex) {
			delete(breakpointMap, lineIndex)
		}
	}
}

func (g *Debugger) EnableAllBreakPoints() {
	for _, breakpoint := range g.currentBreakPointMap {
		breakpoint.On = true
	}
}

func (g *Debugger) EnableBreakpointsInLine(lineIndex int) {
	for _, breakpoint := range g.currentBreakPointMap {
		if breakpoint.LineIndex == lineIndex {
			breakpoint.On = true
		}
	}
}

func (g *Debugger) DisableAllBreakPoints() {
	for _, breakpoint := range g.currentBreakPointMap {
		breakpoint.On = false
	}
}

func (g *Debugger) DisableBreakpointsInLine(lineIndex int) {
	for _, breakpoint := range g.currentBreakPointMap {
		if breakpoint.LineIndex == lineIndex {
			breakpoint.On = false
		}
	}
}

func (g *Debugger) StepNext() error {
	stackTrace := g.CurrentStackTrace()
	stackLen := 0
	if stackTrace != nil {
		stackLen = stackTrace.Len()
	}
	g.nextState = &DebuggerState{
		lineIndex: g.linePointer,
		frame:     g.frame,
		stackLen:  stackLen,
	}
	return nil
}

func (g *Debugger) StepIn() error {
	g.GetLineFirstCode(g.linePointer)
	// will not use stackLen, so set it to -1
	g.stepInState = &DebuggerState{
		lineIndex: g.linePointer,
		frame:     g.frame,
	}
	return nil
}

func (g *Debugger) StepOut() error {
	stackTrace := g.CurrentStackTrace()
	stackLen := stackTrace.Len()
	if stackTrace != nil && stackLen > 0 {
		// Frame will not be used, so set it to nil.
		g.stepoutState = &DebuggerState{
			lineIndex: g.linePointer,
			stackLen:  stackLen,
		}
		return nil
	} else {
		return utils.Errorf("Can't not step out")
	}
}

func (g *Debugger) CurrentStackTracePop() {
	stackTrace := g.CurrentStackTrace()
	if stackTrace != nil && stackTrace.Len() > 0 {
		stackTrace.Pop()
	}
}

func (g *Debugger) HitCount(breakpoint *Breakpoint) bool {
	// If the number of hits is greater than 0, then the number of hits is reduced by 1. If it is still greater than 0, keep clicking
	if breakpoint.HitCount > 0 {
		breakpoint.HitCount--
	}
	return breakpoint.HitCount > 0
}

func (g *Debugger) HandleForStepNext() {
	g.nextState = nil
	g.SetStopReason("step")
	g.Callback()
}

func (g *Debugger) HandleForStepIn() {
	g.stepInState = nil
	g.SetStopReason("step")
	g.Callback()
}

func (g *Debugger) HandleForStepOut() {
	g.stepoutState = nil
	g.SetStopReason("step")
	g.Callback()
}

func (g *Debugger) HandleForPause() {
	g.SetStopReason("pause")
	g.Callback()
}

func (g *Debugger) HandleForBreakPoint() {
	g.SetStopReason("breakpoint")
	g.Callback()
}

func (g *Debugger) HandleForNormallyFinished() {
	g.SetStopReason("finished")
	g.Callback()
}

func (g *Debugger) HandleForPanic(vmPanic *VMPanic) {
	if !g.Finished() {
		g.SetFinished()
		g.SetVMPanic(vmPanic)
		g.SetDescription("panic")
		g.SetStopReason("exception")
		g.Callback()
	}
}

func (g *Debugger) ShouldCallback(frame *Frame) {

	g.lock.Lock()
	defer g.lock.Unlock()

	if !g.started {
		g.started = true
		g.StartWGDone()
	}

	codeIndex := frame.codePointer
	g.UpdateByFrame(frame)

	state, stateName := g.State(), g.StateName()
	code := g.GetCode(state, codeIndex)
	lineIndex, _, _, _ := GetCodeNumber(code)
	g.codePointer = codeIndex
	g.linePointer = lineIndex

	g.SwitchByOtherFileOpcode(code)

	stackTrace := g.CurrentStackTrace()

	if code.Opcode == OpCall {
		v := frame.peekN(code.Unary)
		// If the yak function is called synchronously, push stepIn stack
		if v != nil && v.IsYakFunction() {
			defer func() {
				if stackTrace != nil {
					stackTrace.Push(&DebuggerState{
						code:      code,
						codeIndex: codeIndex,
						stateName: stateName,
						frame:     frame,
					})
				}
			}()
		}
	}
	// Back off the stack in defer processing of frame.Exec

	// Update ThreadStackTrace
	g.ThreadStackTrace[g.frame.ThreadID] = &DebuggerState{
		code:      code,
		codeIndex: codeIndex,
		stateName: stateName,
		frame:     frame,
	}

	// Capture error
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				g.description = fmt.Sprintf("Runtime error: %v", rerr)
			} else {
				g.description = fmt.Sprintf("Runtime error: %v", r)
			}
			g.HandleForPanic(NewVMPanic(r))
		}
	}()

	// . If halt, callback
	if g.halt {
		g.halt = false
		g.HandleForPause()
		return
	}

	// . Step
	if g.nextState != nil {
		// If the debugger wants to step through and jmp appears, callback
		if g.jmpState != nil && g.jmpState.codeIndex == codeIndex && g.jmpState.frame == frame {
			g.jmpState = nil
			g.HandleForStepNext()
		} else if g.nextState.frame == frame && g.linePointer > g.nextState.lineIndex {
			// . If the debugger wants to step through and is indeed in the next line, call back
			g.HandleForStepNext()
		} else if stackTrace.Len() < g.nextState.stackLen {
			// When the stack length is less than stepoutState.stackLen, it is possible that a function has returned, and the
			g.stepoutState = nil
			g.HandleForStepNext()
		}
		return
	} else {
		// If not in next state, jmpIndex should be cleared
		defer func() {
			if g.jmpState != nil {
				g.jmpState = nil
			}
		}()
	}

	// Step into
	if g.stepInState != nil {
		// In the same thread, if the debugger wants to step and the frame is different, call back
		if g.stepInState.frame.ThreadID == frame.ThreadID && g.stepInState.frame != frame {
			g.HandleForStepIn()
		} else if g.stepInState.lineIndex < g.linePointer {
			// If this line has been exceeded, callback
			g.HandleForStepIn()
		}
		return
	}

	// Step out of
	if g.stepoutState != nil {
		// If the current stack length is less than stepoutState.stackLen, call back
		if stackTrace.Len() < g.stepoutState.stackLen {
			g.HandleForStepOut()
		}
		return
	}

	if len(g.currentObserveBreakPointMap) > 0 {
		// when the function (normal/should not be triggered when exiting due to
		if code.Opcode != OpReturn && code.Opcode != OpPanic {
			for expr, v := range g.currentObserveBreakPointMap {
				nv, err := g.EvalExpression(expr)
				if nv == nil || err != nil {
					nv = undefined
				}
				if !v.Equal(nv) {
					g.currentObserveBreakPointMap[expr] = nv
					g.description = fmt.Sprintf("Trigger observe breakpoint at line %d in %s", g.linePointer, g.StateName())
					g.HandleForBreakPoint()
					return
				}
			}
		}
	}

	triggered := false
	// If it exists in the breakpoint list, callback
	for _, breakpoint := range g.currentBreakPointMap {

		// If the breakpoint is disabled,
		if !breakpoint.On {
			continue
		}

		// . If it is not in the same state,
		if state != breakpoint.State {
			continue
		}

		// line should be called back Breakpoints include ordinary breakpoints and conditional breakpoints. When the code jumps, the judgment conditions will be relaxed, and only the line numbers need to be the same.
		//
		if breakpoint.CodeIndex == codeIndex || (g.jmpState != nil && breakpoint.LineIndex == lineIndex) {
			// Conditional breakpoints
			condition, hitCondition := breakpoint.Condition, breakpoint.HitCondition
			if condition == "" {
				// If the number of hits is greater than 0, then the number of hits is reduced by 1. If it is still greater than 0, keep clicking
				if g.HitCount(breakpoint) {
					continue
				}
			}

			if condition != "" || hitCondition != "" {
				// If condition is empty, use hitCondition
				cond := condition
				if condition == "" {
					cond = hitCondition
				}
				value, err := g.EvalExpression(cond)

				// If If the condition is not established, click
				if err != nil || value.False() {
					continue
				}

				// If the number of hits is greater than 0, then the number of hits is reduced by 1. If it is still greater than 0, keep clicking
				if g.HitCount(breakpoint) {
					continue
				}

				// If hitCondition is not empty, you also need to judge hitCondition
				if hitCondition != "" {
					value, err := g.EvalExpression(hitCondition)

					// If the condition is not established, continue to point
					if err != nil || value.False() {
						continue
					}

					cond = fmt.Sprintf("%s && %s", condition, hitCondition)
				}

				// Conditions for triggering conditional breakpoints:
				// 1. The condition is established, there is no hitCount and hitCondition
				// 2. hitCount exists and decreases to 0
				// 3. condition is established, hitCondition is established

				g.description = fmt.Sprintf("Trigger conditional breakpoint [%s] at line %d in %s", cond, g.linePointer, g.StateName())
			} else {
				// Ordinary breakpoint
				g.description = fmt.Sprintf("Trigger normal breakpoint at line %d in %s", g.linePointer, g.StateName())
			}

			triggered = true

			break
		}
	}

	if triggered {
		g.HandleForBreakPoint()
	}
}

func (g *Debugger) Callback() {
	g.Add()
	defer g.WaitGroupDone()
	defer g.ResetStopReason()

	// . Handle stopOnEntry situation
	if !g.started {
		g.started = true
		g.StartWGDone()
	}

	// Update observation expression
	if len(g.observeExpressions) > 0 {
		for expr := range g.observeExpressions {
			value, err := g.EvalExpression(expr)
			if err != nil {
				value = undefined
			}
			g.observeExpressions[expr] = value
		}
	}

	g.callbackFunc(g)
}

func (g *Debugger) GetScopesByFrameID(frameID int) map[int]*Scope {
	ref := g.Reference
	frame, ok := ref.FrameHM.get(frameID)
	if !ok {
		return nil
	}
	scopes := make(map[int]*Scope, 0)
	scope := frame.CurrentScope()
	for scope != nil {
		if id, ok := ref.VarHM.getReverse(scope); ok {
			scopes[id] = scope
		}
		scope = scope.parent
	}
	return scopes
}

func (g *Debugger) GetVariablesByRef(ref int) (interface{}, bool) {
	v, ok := g.Reference.VarHM.get(ref)
	if !ok {
		return nil, false
	}
	return v, true
}

func (g *Debugger) GetVariablesRef(v interface{}) (int, bool) {
	i, ok := g.Reference.VarHM.getReverse(v)
	if !ok {
		return 0, false
	}
	return i, true
}

func (g *Debugger) CompileWithFrame(code string, frame *Frame) (*Frame, CompilerWrapperInterface, error) {
	var err error
	frame = NewSubFrame(frame)
	frame.EnableDebuggerEval()
	sym := frame.CurrentScope().GetSymTable()

	c := YakDebugCompiler.NewWithSymbolTable(sym)
	c.Compiler(code)
	exist, err := YakDebugCompiler.GetNormalErrors()
	if exist {
		return nil, nil, errors.Wrap(err, "compile code error")
	}
	return frame, c, nil
}

func (g *Debugger) CompileWithFrameID(code string, frameID int) (*Frame, CompilerWrapperInterface, error) {
	targetFrame, ok := g.Reference.FrameHM.get(frameID)
	if !ok {
		return nil, nil, errors.New("frame not found")
	}

	return g.CompileWithFrame(code, targetFrame)
}

func (g *Debugger) Compile(code string) (*Frame, CompilerWrapperInterface, error) {
	return g.CompileWithFrame(code, g.frame)
}

func (g *Debugger) evalExpressionWithOpCodes(opcode []*Code, frame *Frame) (*Value, error) {
	var err error

	if len(opcode) == 0 {
		return nil, errors.New("eval code error: no opcode")
	}

	// does special processing for opcode, change pop to return
	if opcode[len(opcode)-1].Opcode == OpPop {
		opcode[len(opcode)-1].Opcode = OpReturn
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				if rerr, ok := r.(error); ok {
					err = errors.Wrap(rerr, "eval code error")
				} else if rstr, ok := r.(string); ok {
					err = errors.Wrap(errors.New(rstr), "eval code error")
				}
			}
		}()
		frame.Exec(opcode)
	}()

	return frame.GetLastStackValue(), err
}

func (g *Debugger) EvalExpressionWithFrameID(expr string, frameID int) (*Value, error) {
	var err error

	frame, compiler, err := g.CompileWithFrameID(expr, frameID)
	if err != nil {
		return nil, errors.Wrap(err, "eval code error")
	}

	opcode := compiler.GetOpcodes()
	return g.evalExpressionWithOpCodes(opcode, frame)
}

func (g *Debugger) EvalExpression(expr string) (*Value, error) {
	var err error

	frame, compiler, err := g.Compile(expr)
	if err != nil {
		return nil, errors.Wrap(err, "eval code error")
	}

	opcode := compiler.GetOpcodes()

	return g.evalExpressionWithOpCodes(opcode, frame)
}
