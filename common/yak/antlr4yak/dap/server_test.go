package dap

import (
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-dap"
)

var (
	StopOnEntry = true
	noChildren  = false
	hasChildren = true
)

type onBreakpoint struct {
	execute    func()
	disconnect bool
	terminated bool
}
type helperForSetVariable struct {
	t *testing.T
	c *Client
}

func (h *helperForSetVariable) expectSetVariable(ref int, name, value string) {
	h.t.Helper()
	h.expectSetVariable0(ref, name, value, false)
}

func (h *helperForSetVariable) failSetVariable(ref int, name, value, wantErrInfo string) {
	h.t.Helper()
	h.failSetVariable0(ref, name, value, wantErrInfo, false)
}

func (h *helperForSetVariable) failSetVariableAndStop(ref int, name, value, wantErrInfo string) {
	h.t.Helper()
	h.failSetVariable0(ref, name, value, wantErrInfo, true)
}

func (h *helperForSetVariable) evaluate(frameID int, expr, want string, hasRef bool) {
	h.t.Helper()
	h.c.EvaluateRequest(expr, frameID, "whatever")
	got := h.c.ExpectEvaluateResponse(h.t)
	checkEval(h.t, got, want, hasRef)
}

func (h *helperForSetVariable) evaluateRegex(frameID int, expr, want string, hasRef bool) {
	h.t.Helper()
	h.c.EvaluateRequest(expr, frameID, "whatever")
	got := h.c.ExpectEvaluateResponse(h.t)
	checkEvalRegex(h.t, got, want, hasRef)
}

func (h *helperForSetVariable) expectSetVariable0(ref int, name, value string, wantStop bool) {
	h.t.Helper()

	h.c.SetVariableRequest(ref, name, value)
	if wantStop {
		h.c.ExpectStoppedEvent(h.t)
	}
	if got, want := h.c.ExpectSetVariableResponse(h.t), value; got.Success != true || got.Body.Value != want {
		h.t.Errorf("SetVariableRequest(%v, %v)=%#v, want {Success=true, Body.Value=%q", name, value, got, want)
	}
}

func (h *helperForSetVariable) failSetVariable0(ref int, name, value, wantErrInfo string, wantStop bool) {
	h.t.Helper()

	h.c.SetVariableRequest(ref, name, value)
	if wantStop {
		h.c.ExpectStoppedEvent(h.t)
	}
	resp := h.c.ExpectErrorResponse(h.t)
	if got := resp.Body.Error; !strings.Contains(got.Format, wantErrInfo) {
		h.t.Errorf("got %#v, want error string containing %v", got, wantErrInfo)
	}
}

func (h *helperForSetVariable) variables(ref int) *dap.VariablesResponse {
	h.t.Helper()
	h.c.VariablesRequest(ref)
	return h.c.ExpectVariablesResponse(h.t)
}

func runTest(t *testing.T, name string, filepath string, test func(s *DAPServer, c *Client, program string)) {
	serverStopped := make(chan struct{})

	server, _ := startDAPServer(t, serverStopped)
	client := NewTestClient(server.listener.Addr().String())
	defer client.Close()

	filepath = GetYakTestCasePath(filepath)

	test(server, client, filepath)
	<-serverStopped
}

func runDebugSessionWithBPs(t *testing.T, client *Client, cmdRequest func(), source string, breakpoints []int, onBPs []onBreakpoint) {
	client.InitializeRequest()
	client.ExpectInitializeResponseAndCapabilities(t)

	cmdRequest()
	client.ExpectInitializedEvent(t)
	client.ExpectLaunchResponse(t)
	if source != "" {
		client.SetBreakpointsRequest(source, breakpoints)
		client.ExpectSetBreakpointsResponse(t)
	}

	disconnect := func() {
		client.DisconnectRequest()
		client.ExpectOutputEventDetaching(t)
		client.ExpectDisconnectResponse(t)
		client.ExpectTerminatedEvent(t)
	}

	client.ConfigurationDoneRequest()
	client.ExpectOutputEventHelpInfo(t)
	client.ExpectConfigurationDoneResponse(t)

	for _, onBP := range onBPs {
		client.ExpectStoppedEvent(t)
		onBP.execute()
		if onBP.disconnect {
			disconnect()
			return
		}
		client.ContinueRequest(1)
		client.ExpectContinueResponse(t)
	}

	client.ExpectTerminatedEvent(t) // Let the program run to completion

	disconnect()
}

func startDAPServer(t *testing.T, serverStopped chan struct{}) (server *DAPServer, forceStop chan struct{}) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	stopChan := make(chan struct{})
	forceStop = make(chan struct{})

	server = NewDAPServer(&DAPServerConfig{
		listener: l,
		stopped:  stopChan,
	})

	server.Start()
	// Give server time to start listening for clients
	time.Sleep(100 * time.Millisecond)

	go func() {
		defer func() {
			if serverStopped != nil {
				close(serverStopped)
			}
		}()
		select {
		case <-stopChan:
			t.Log("server stop by call stop function")
		case <-forceStop:
			t.Log("server stop by force")
		}
		server.Stop()
	}()
	return
}

func verifyServerStopped(t *testing.T, server *DAPServer) {
	t.Helper()
	if server.listener != nil {
		if server.listener.Close() == nil {
			t.Error("server should have closed listener after shutdown")
		}
	}
	verifySessionStopped(t, server.session)
}

func verifySessionStopped(t *testing.T, session *DebugSession) {
	t.Helper()
	if session == nil {
		return
	}
	if session.conn == nil {
		t.Error("session must always have a connection")
	}
	verifyConnStopped(t, session.conn)
}

func verifyConnStopped(t *testing.T, conn net.Conn) {
	t.Helper()
	if conn.Close() == nil {
		t.Error("client connection should be closed after shutdown")
	}
}

func validateEvaluateName(t *testing.T, client *Client, frameID int, got *dap.VariablesResponse, i int) {
	t.Helper()
	original := got.Body.Variables[i]
	client.EvaluateRequest(original.EvaluateName, frameID, "this context will be ignored")
	validated := client.ExpectEvaluateResponse(t)
	if original.VariablesReference == 0 && validated.Body.VariablesReference != 0 ||
		original.VariablesReference != 0 && validated.Body.VariablesReference == 0 {
		t.Errorf("\ngot  varref=%d\nwant %d", validated.Body.VariablesReference, original.VariablesReference)
	}

	if original.Value != validated.Body.Result {
		t.Errorf("\ngot  value=%q\nwant %q", validated.Body.Result, original.Value)
	}
}

func checkErrorMessageId(er *dap.ErrorMessage, id int) bool {
	return er != nil && er.Id == id
}

func checkErrorMessageFormat(er *dap.ErrorMessage, fmt string) bool {
	return er != nil && er.Format == fmt
}

func checkScope(t *testing.T, got *dap.ScopesResponse, i int, name string, varRef int) {
	t.Helper()
	if len(got.Body.Scopes) <= i {
		t.Errorf("\ngot  %d\nwant len(Scopes)>%d", len(got.Body.Scopes), i)
	}
	goti := got.Body.Scopes[i]
	if goti.Name != name || (varRef >= 0 && goti.VariablesReference != varRef) || goti.Expensive {
		t.Errorf("\ngot  %#v\nwant Name=%q VariablesReference=%d Expensive=false", goti, name, varRef)
	}
}

func checkChildren(t *testing.T, got *dap.VariablesResponse, parentName string, numChildren int) {
	t.Helper()
	if got.Body.Variables == nil {
		t.Errorf("\ngot  %s children=%#v want []", parentName, got.Body.Variables)
	}
	if len(got.Body.Variables) != numChildren {
		t.Errorf("\ngot  len(%s)=%d (children=%#v)\nwant len=%d", parentName, len(got.Body.Variables), got.Body.Variables, numChildren)
	}
}

func checkStop(t *testing.T, client *Client, thread int, fname string, line int) {
	t.Helper()
	client.ThreadsRequest()
	client.ExpectThreadsResponse(t)

	client.CheckStopLocation(t, thread, fname, line)

	client.ScopesRequest(0)
	client.ExpectScopesResponse(t)

	client.VariablesRequest(1)
	client.ExpectVariablesResponse(t)
}

func checkStackFramesExact(t *testing.T, got *dap.StackTraceResponse,
	wantStartName string, wantStartLine, wantStartID, wantFrames, wantTotalFrames int) {
	t.Helper()
	checkStackFramesNamed("", t, got, wantStartName, wantStartLine, wantStartID, wantFrames, wantTotalFrames, true)
}

func checkStackFramesNamed(testName string, t *testing.T, got *dap.StackTraceResponse,
	wantStartName string, wantStartLine, wantStartFrame, wantFrames, wantTotalFrames int, totalExact bool) {
	t.Helper()
	if totalExact && got.Body.TotalFrames != wantTotalFrames {
		t.Errorf("%s\ngot  %#v\nwant TotalFrames=%d", testName, got.Body.TotalFrames, wantTotalFrames)
	} else if !totalExact && got.Body.TotalFrames < wantTotalFrames {
		t.Errorf("%s\ngot  %#v\nwant TotalFrames>=%d", testName, got.Body.TotalFrames, wantTotalFrames)
	}

	if len(got.Body.StackFrames) != wantFrames {
		t.Errorf("%s\ngot  len(StackFrames)=%d\nwant %d", testName, len(got.Body.StackFrames), wantFrames)
	} else {
		for i := 0; i < wantFrames; i++ {
			frame := got.Body.StackFrames[i]
			// Since the topmost frameID is the largest,
			want := wantTotalFrames - wantStartFrame - (i + 1)
			if frame.Id != want {
				t.Errorf("%s\ngot  %#v\nwant Id=%d", testName, frame, want)
			}
		}
		// Verify the name and line corresponding to the first returned frame (if any).
		// This is useful when the first frame is the frame corresponding to the breakpoint at
		// a predefined line. Line values < 0 are a signal to skip the check (which can be useful
		// for frames in the third-party code, where we do not control the lines).
		if wantFrames > 0 && wantStartLine > 0 && got.Body.StackFrames[0].Line != wantStartLine {
			t.Errorf("%s\ngot  Line=%d\nwant %d", testName, got.Body.StackFrames[0].Line, wantStartLine)
		}
		if wantFrames > 0 && wantStartName != "" && got.Body.StackFrames[0].Name != wantStartName {
			t.Errorf("%s\ngot  Name=%s\nwant %s", testName, got.Body.StackFrames[0].Name, wantStartName)
		}
	}
}

func checkEval(t *testing.T, got *dap.EvaluateResponse, value string, hasRef bool) (ref int) {
	t.Helper()
	if got.Body.Result != value || (got.Body.VariablesReference > 0) != hasRef {
		t.Errorf("\ngot  %#v\nwant Result=%q hasRef=%t", got, value, hasRef)
	}
	return got.Body.VariablesReference
}

func checkEvalIndexed(t *testing.T, got *dap.EvaluateResponse, value string, hasRef bool, indexed, named int) (ref int) {
	t.Helper()
	if got.Body.Result != value || (got.Body.VariablesReference > 0) != hasRef || got.Body.IndexedVariables != indexed || got.Body.NamedVariables != named {
		t.Errorf("\ngot  %#v\nwant Result=%q hasRef=%t IndexedVariables=%d NamedVariables=%d", got, value, hasRef, indexed, named)
	}
	return got.Body.VariablesReference
}

func checkEvalRegex(t *testing.T, got *dap.EvaluateResponse, valueRegex string, hasRef bool) (ref int) {
	t.Helper()
	matched, _ := regexp.MatchString(valueRegex, got.Body.Result)
	if !matched || (got.Body.VariablesReference > 0) != hasRef {
		t.Errorf("\ngot  %#v\nwant Result=%q hasRef=%t", got, valueRegex, hasRef)
	}
	return got.Body.VariablesReference
}

func checkVar(t *testing.T, got *dap.VariablesResponse, i int, name, evalName, value, typ string, useExactMatch, hasRef bool, indexed, named int) (ref int) {
	t.Helper()
	if len(got.Body.Variables) <= i {
		t.Errorf("\ngot  len=%d (children=%#v)\nwant len>%d", len(got.Body.Variables), got.Body.Variables, i)
		return
	}
	if i < 0 {
		for vi, v := range got.Body.Variables {
			if v.Name == name {
				i = vi
				break
			}
		}
	}
	if i < 0 {
		t.Errorf("\ngot  %#v\nwant Variables[i].Name=%q (not found)", got, name)
		return 0
	}

	goti := got.Body.Variables[i]
	matchedName := false
	if useExactMatch {
		matchedName = (goti.Name == name)
	} else {
		matchedName, _ = regexp.MatchString(name, goti.Name)
	}
	if !matchedName || (goti.VariablesReference > 0) != hasRef {
		t.Errorf("\ngot  %#v\nwant Name=%q hasRef=%t", goti, name, hasRef)
	}
	matchedEvalName := false
	if useExactMatch {
		matchedEvalName = (goti.EvaluateName == evalName)
	} else {
		matchedEvalName, _ = regexp.MatchString(evalName, goti.EvaluateName)
	}
	if !matchedEvalName {
		t.Errorf("\ngot  %q\nwant EvaluateName=%q", goti.EvaluateName, evalName)
	}
	matchedValue := false
	if useExactMatch {
		matchedValue = (goti.Value == value)
	} else {
		matchedValue, _ = regexp.MatchString(value, goti.Value)
	}
	if !matchedValue {
		t.Errorf("\ngot  %s=%q\nwant %q", name, goti.Value, value)
	}
	matchedType := false
	if useExactMatch {
		matchedType = (goti.Type == typ)
	} else {
		matchedType, _ = regexp.MatchString(typ, goti.Type)
	}
	if !matchedType {
		t.Errorf("\ngot  %s=%q\nwant %q", name, goti.Type, typ)
	}
	if indexed >= 0 && goti.IndexedVariables != indexed {
		t.Errorf("\ngot  %s=%d indexed\nwant %d indexed", name, goti.IndexedVariables, indexed)
	}
	if named >= 0 && goti.NamedVariables != named {
		t.Errorf("\ngot  %s=%d named\nwant %d named", name, goti.NamedVariables, named)
	}
	return goti.VariablesReference
}

func checkVarExact(t *testing.T, got *dap.VariablesResponse, i int, name, evalName, value, typ string, hasRef bool) (ref int) {
	t.Helper()
	return checkVarExactIndexed(t, got, i, name, evalName, value, typ, hasRef, -1, -1)
}

func checkVarExactIndexed(t *testing.T, got *dap.VariablesResponse, i int, name, evalName, value, typ string, hasRef bool, indexed, named int) (ref int) {
	t.Helper()
	return checkVar(t, got, i, name, evalName, value, typ, true, hasRef, indexed, named)
}

func checkVarRegex(t *testing.T, got *dap.VariablesResponse, i int, name, evalName, value, typ string, hasRef bool) (ref int) {
	t.Helper()
	return checkVarRegexIndexed(t, got, i, name, evalName, value, typ, hasRef, -1, -1)
}

func checkVarRegexIndexed(t *testing.T, got *dap.VariablesResponse, i int, name, evalName, value, typ string, hasRef bool, indexed, named int) (ref int) {
	t.Helper()
	return checkVar(t, got, i, name, evalName, value, typ, false, hasRef, indexed, named)
}

func TestStopNoCilent(t *testing.T) {
	for name, triggerStop := range map[string]func(s *DAPServer, forceStop chan struct{}){
		"force":          func(s *DAPServer, forceStop chan struct{}) { close(forceStop) },
		"listener_close": func(s *DAPServer, forceStop chan struct{}) { s.listener.Close() },
	} {
		t.Run(name, func(t *testing.T) {
			serverStopped := make(chan struct{})
			server, forceStop := startDAPServer(t, serverStopped)

			triggerStop(server, forceStop)
			<-serverStopped
			verifyServerStopped(t, server)
		})
	}
}

func TestStopNoTarget(t *testing.T) {
	for name, triggerStop := range map[string]func(c *Client, forceStop chan struct{}){
		"force":              func(c *Client, forceStop chan struct{}) { close(forceStop) },
		"client_close":       func(c *Client, forceStop chan struct{}) { c.Close() },
		"disconnect_request": func(c *Client, forceStop chan struct{}) { c.DisconnectRequest() },
	} {
		t.Run(name, func(t *testing.T) {
			serverStopped := make(chan struct{})
			server, forceStop := startDAPServer(t, serverStopped)
			client := NewTestClient(server.listener.Addr().String())
			defer client.Close()

			client.InitializeRequest()
			client.ExpectInitializeResponseAndCapabilities(t)

			triggerStop(client, forceStop)
			<-serverStopped
			verifyServerStopped(t, server)
		})
	}
}

func TestStopWithTarget(t *testing.T) {
	for name, triggerStop := range map[string]func(c *Client, forceStop chan struct{}){
		"force":                  func(c *Client, forceStop chan struct{}) { close(forceStop) },
		"client_close":           func(c *Client, forceStop chan struct{}) { c.Close() },
		"disconnect_before_exit": func(c *Client, forceStop chan struct{}) { c.DisconnectRequest() },
		"disconnect_after_exit": func(c *Client, forceStop chan struct{}) {
			c.ContinueRequest(1)
			c.ExpectContinueResponse(t)

			c.ExpectTerminatedEvent(t) // program finished, recv terminated event

			c.DisconnectRequest()
			c.ExpectOutputEventDetaching(t)
			c.ExpectDisconnectResponse(t)
			c.ExpectTerminatedEvent(t)
		},
	} {
		t.Run(name, func(t *testing.T) {
			serverStopped := make(chan struct{})
			server, forceStop := startDAPServer(t, serverStopped)
			client := NewTestClient(server.listener.Addr().String())
			defer client.Close()

			client.InitializeRequest()
			client.ExpectInitializeResponseAndCapabilities(t)
			tc := GetYakTestCasePath(SimpleYakTestCase)

			client.LaunchRequest("exec", tc, StopOnEntry)
			client.ExpectInitializedEvent(t)
			client.ExpectLaunchResponse(t)

			triggerStop(client, forceStop)
			<-serverStopped
			verifyServerStopped(t, server)
		})
	}
}

func TestForceStopWhileStopping(t *testing.T) {
	serverStopped := make(chan struct{})
	server, forceStop := startDAPServer(t, serverStopped)
	client := NewTestClient(server.listener.Addr().String())

	client.InitializeRequest()
	client.ExpectInitializeResponseAndCapabilities(t)
	tc := GetYakTestCasePath(SimpleYakTestCase)

	client.LaunchRequest("exec", tc, StopOnEntry)
	client.ExpectInitializedEvent(t)
	client.Close()
	time.Sleep(time.Microsecond)
	close(forceStop)
	<-serverStopped
	verifyServerStopped(t, server)
}

func TestLaunchStopOnEntry(t *testing.T) {
	runTest(t, "stopOnEntry", SimpleYakTestCase, func(server *DAPServer, client *Client, program string) {
		// 1 >> initialize, << initialize
		client.InitializeRequest()
		initResp := client.ExpectInitializeResponseAndCapabilities(t)
		if initResp.Seq != 0 || initResp.RequestSeq != 1 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=1", initResp)
		}

		// 2 >> launch,  << initialized, << launch
		client.LaunchRequest("exec", program, StopOnEntry)
		initEvent := client.ExpectInitializedEvent(t)
		if initEvent.Seq != 0 {
			t.Errorf("\ngot %#v\nwant Seq=0", initEvent)
		}
		launchResp := client.ExpectLaunchResponse(t)
		if launchResp.Seq != 0 || launchResp.RequestSeq != 2 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=2", launchResp)
		}

		// 3 >> setBreakpoints, << setBreakpoints
		client.SetBreakpointsRequest(program, nil)
		sbpResp := client.ExpectSetBreakpointsResponse(t)
		if sbpResp.Seq != 0 || sbpResp.RequestSeq != 3 || len(sbpResp.Body.Breakpoints) != 0 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=3, len(Breakpoints)=0", sbpResp)
		}

		// 4 >> setExceptionBreakpoints, << setExceptionBreakpoints
		client.SetExceptionBreakpointsRequest()
		sebpResp := client.ExpectSetExceptionBreakpointsResponse(t)
		if sebpResp.Seq != 0 || sebpResp.RequestSeq != 4 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=4", sebpResp)
		}
		// 5 >> configurationDone, << stopped, << configurationDone
		client.ConfigurationDoneRequest()
		stopEvent := client.ExpectStoppedEvent(t)
		if stopEvent.Seq != 0 ||
			stopEvent.Body.Reason != "entry" ||
			stopEvent.Body.ThreadId != 0 ||
			!stopEvent.Body.AllThreadsStopped {
			t.Errorf("\ngot %#v\nwant Seq=0, Body={Reason=\"entry\", ThreadId=0, AllThreadsStopped=true}", stopEvent)
		}
		client.ExpectOutputEventHelpInfo(t)
		cdResp := client.ExpectConfigurationDoneResponse(t)
		if cdResp.Seq != 0 || cdResp.RequestSeq != 5 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=5", cdResp)
		}

		// from varRef=1, that is, Globals acts on it. StopOnEntry is started at the beginning, so continue. Since the execution will end directly after continue, the terminated event
		// 6 >> continue, << continue, << terminated
		client.ContinueRequest(1)
		cResp := client.ExpectContinueResponse(t)
		if cResp.Seq != 0 || cResp.RequestSeq != 6 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=6", cResp)
		}
		termEvent := client.ExpectTerminatedEvent(t)
		if termEvent.Seq != 0 {
			t.Errorf("\ngot %#v\nwant Seq=0", termEvent)
		}

		// 7 >> threads, << threads
		client.ThreadsRequest()
		tResp := client.ExpectThreadsResponse(t)
		if tResp.Seq != 0 || tResp.RequestSeq != 7 || len(tResp.Body.Threads) != 1 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=6 len(Threads)=1", tResp)
		}
		if len(tResp.Body.Threads) < 1 || tResp.Body.Threads[0].Id != 1 || tResp.Body.Threads[0].Name != "[Yak 1] __yak_main__" {
			t.Errorf("\ngot %#v\nwant Id=1, Name=\"[Yak 1] __yak_main__\"", tResp)
		}

		// 8 >> stackTrace, << error
		client.StackTraceRequest(2, 0, 20)
		steResp := client.ExpectInvisibleErrorResponse(t)
		if steResp.Seq != 0 || steResp.RequestSeq != 8 || steResp.Success || !checkErrorMessageFormat(steResp.Body.Error, "Unable to produce stack trace: Can't found Goroutine 2 stack trace") {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=8 Format=\"Unable to produce stack trace: Can't found Goroutine 2 stack trace\"", steResp)
		}

		// 9 >> stackTrace, << stackTrace
		client.StackTraceRequest(1, 0, 20)
		stResp := client.ExpectStackTraceResponse(t)
		if stResp.Seq != 0 || stResp.RequestSeq != 9 || !stResp.Success || stResp.Body.TotalFrames != 1 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=9 len(StackTraces)=1 ", stResp)
		}

		// 10 >> evaluate, << error
		client.EvaluateRequest("{", 0, "repl")
		erResp := client.ExpectInvisibleErrorResponse(t)
		_ = erResp
		if erResp.Seq != 0 || erResp.RequestSeq != 10 || !checkErrorMessageId(erResp.Body.Error, UnableToEvaluateExpression) {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=10 Id=%d", erResp, UnableToEvaluateExpression)
		}

		// 11 >> evaluate, << evaluate
		client.EvaluateRequest("1+1", 0 /*no frame specified*/, "repl")
		evResp := client.ExpectEvaluateResponse(t)
		if evResp.Seq != 0 || evResp.RequestSeq != 11 || evResp.Body.Result != "2" {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=10 Result=2", evResp)
		}

		// 12 >> continue, << continue
		client.ContinueRequest(1)
		contResp := client.ExpectContinueResponse(t)
		if contResp.Seq != 0 || contResp.RequestSeq != 12 || !contResp.Body.AllThreadsContinued {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=12 Body.AllThreadsContinued=true", contResp)
		}

		// 13 >> disconnect, << disconnect
		client.DisconnectRequest()
		oed := client.ExpectOutputEventDetaching(t)
		if oed.Seq != 0 || oed.Body.Category != "console" {
			t.Errorf("\ngot %#v\nwant Seq=0 Category='console'", oed)
		}
		dResp := client.ExpectDisconnectResponse(t)
		if dResp.Seq != 0 || dResp.RequestSeq != 13 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=13", dResp)
		}
		client.ExpectTerminatedEvent(t)
	})
}

func TestLaunchContinueOnEntry(t *testing.T) {
	runTest(t, "continueOnEntry", SimpleYakTestCase, func(server *DAPServer, client *Client, program string) {
		// 1 >> initialize, << initialize
		client.InitializeRequest()
		initResp := client.ExpectInitializeResponseAndCapabilities(t)
		if initResp.Seq != 0 || initResp.RequestSeq != 1 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=1", initResp)
		}

		// 2 >> launch,  << initialized, << launch
		client.LaunchRequest("exec", program, !StopOnEntry)
		initEvent := client.ExpectInitializedEvent(t)
		if initEvent.Seq != 0 {
			t.Errorf("\ngot %#v\nwant Seq=0", initEvent)
		}
		launchResp := client.ExpectLaunchResponse(t)
		if launchResp.Seq != 0 || launchResp.RequestSeq != 2 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=2", launchResp)
		}

		// 3 >> setBreakpoints, << setBreakpoints
		client.SetBreakpointsRequest(program, nil)
		sbpResp := client.ExpectSetBreakpointsResponse(t)
		if sbpResp.Seq != 0 || sbpResp.RequestSeq != 3 || len(sbpResp.Body.Breakpoints) != 0 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=3, len(Breakpoints)=0", sbpResp)
		}

		// 4 >> setExceptionBreakpoints, << setExceptionBreakpoints
		client.SetExceptionBreakpointsRequest()
		sebpResp := client.ExpectSetExceptionBreakpointsResponse(t)
		if sebpResp.Seq != 0 || sebpResp.RequestSeq != 4 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=4", sebpResp)
		}
		// 5 >> configurationDone, << stopped, << configurationDone
		client.ConfigurationDoneRequest()
		client.ExpectOutputEventHelpInfo(t)
		cdResp := client.ExpectConfigurationDoneResponse(t)
		if cdResp.Seq != 0 || cdResp.RequestSeq != 5 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=5", cdResp)
		}

		// "Continue" happens behind the scenes on another goroutine
		client.ExpectTerminatedEvent(t)

		// 6 >> threads, << threads
		client.ThreadsRequest()
		tResp := client.ExpectThreadsResponse(t)
		if tResp.Seq != 0 || tResp.RequestSeq != 6 || len(tResp.Body.Threads) != 1 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=6 len(Threads)=1", tResp)
		}
		if len(tResp.Body.Threads) < 1 || tResp.Body.Threads[0].Id != 1 || tResp.Body.Threads[0].Name != "[Yak 1] __yak_main__" {
			t.Errorf("\ngot %#v\nwant Id=1, Name=\"[Yak 1] __yak_main__\"", tResp)
		}

		// 7 >> disconnect, << disconnect
		client.DisconnectRequest()
		oed := client.ExpectOutputEventDetaching(t)
		if oed.Seq != 0 || oed.Body.Category != "console" {
			t.Errorf("\ngot %#v\nwant Seq=0 Category='console'", oed)
		}
		dResp := client.ExpectDisconnectResponse(t)
		if dResp.Seq != 0 || dResp.RequestSeq != 7 {
			t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=7", dResp)
		}
		client.ExpectTerminatedEvent(t)
	})
}

func TestPreSetBreakPoint(t *testing.T) {
	runTest(t, "PreSetBreakPoint", FuncCallTestcase, func(server *DAPServer, client *Client, program string) {
		client.InitializeRequest()
		client.ExpectInitializeResponseAndCapabilities(t)

		client.LaunchRequest("exec", program, !StopOnEntry)
		client.ExpectInitializedEvent(t)
		client.ExpectLaunchResponse(t)

		client.SetBreakpointsRequest(program, []int{2})
		sResp := client.ExpectSetBreakpointsResponse(t)
		if len(sResp.Body.Breakpoints) != 1 {
			t.Errorf("got %#v, want len(Breakpoints)=1", sResp)
		}
		bkpt0 := sResp.Body.Breakpoints[0]
		if !bkpt0.Verified || bkpt0.Line != 2 || bkpt0.Id != 1 || bkpt0.Source.Name != filepath.Base(program) || bkpt0.Source.Path != program {
			t.Errorf("got breakpoints[0] = %#v, want Verified=true, Line=2, Id=1, Path=%q", bkpt0, program)
		}

		client.SetExceptionBreakpointsRequest()
		client.ExpectSetExceptionBreakpointsResponse(t)

		client.ConfigurationDoneRequest()
		client.ExpectOutputEventHelpInfo(t)
		client.ExpectConfigurationDoneResponse(t)

		// wait program run to breakpoint 2
		time.Sleep(100 * time.Millisecond)

		client.ThreadsRequest()
		// Since we are in async mode while running, we might receive messages in either order.
		for i := 0; i < 2; i++ {
			msg := client.ExpectMessage(t)
			switch m := msg.(type) {
			case *dap.ThreadsResponse:
				// If the thread request arrived while the program was running, we expect to get the dummy response
				// with a single goroutine "Current".
				// If the thread request arrived after the stop, we should get the goroutine stopped at main.Increment.
				if len(m.Body.Threads) != 1 {
					t.Errorf("\ngot %#v\nwant Seq=0, RequestSeq=6 len(Threads)=1", m)
				}
				if len(m.Body.Threads) < 1 || m.Body.Threads[0].Id != 1 || m.Body.Threads[0].Name != "[Yak 1] test" {
					t.Errorf("\ngot  %#v\nwant Id=1, Name=\"[Yak 1] test\"", m.Body.Threads)
				}
			case *dap.StoppedEvent:
				if m.Body.Reason != "breakpoint" || m.Body.ThreadId != 1 || !m.Body.AllThreadsStopped || m.Body.Description != "Trigger normal breakpoint at line 2 in test" {
					t.Errorf("got %#v, want Body={Reason=\"breakpoint\", ThreadId=1, AllThreadsStopped=true}", m)
				}
			default:
				t.Fatalf("got %#v, want ThreadsResponse or StoppedEvent", m)
			}
		}

		client.StackTraceRequest(1, 0, 20)
		stResp := client.ExpectStackTraceResponse(t)
		if stResp.Body.TotalFrames != 2 {
			t.Errorf("\ngot %#v\nwant TotalFrames=2", stResp.Body.TotalFrames)
		}
		checkFrame := func(got dap.StackFrame, id int, name string, sourceName string, line int) {
			t.Helper()
			if got.Id != id || got.Name != name {
				t.Errorf("\ngot  %#v\nwant Id=%d Name=%s", got, id, name)
			}
			if (sourceName != "" && (got.Source == nil || got.Source.Name != sourceName)) || (line > 0 && got.Line != line) {
				t.Errorf("\ngot  %#v\nwant Source.Name=%s Line=%d", got, sourceName, line)
			}
		}
		checkFrame(stResp.Body.StackFrames[0], 1, "test", "", 2)
		checkFrame(stResp.Body.StackFrames[1], 0, "__yak_main__", "", 7)

		client.ScopesRequest(1)
		scopes := client.ExpectScopesResponse(t)
		if len(scopes.Body.Scopes) != 3 {
			t.Errorf("\ngot  %#v\nwant len(Scopes)=3 (Locals)", scopes)
		}
		checkScope(t, scopes, 0, "Globals", 1) // will be received. varRef starts from 1
		checkScope(t, scopes, 1, "Locals1", 2)
		checkScope(t, scopes, 2, "Locals2", 3)

		client.VariablesRequest(1) // needs to be reversed to get the variable
		args := client.ExpectVariablesResponse(t)
		checkChildren(t, args, "Globals", 3)
		checkVarExact(t, args, 0, "a", "a", "1", "int", noChildren)
		checkVarExact(t, args, 1, "b", "b", "2", "int", noChildren)

		client.ContinueRequest(1)
		ctResp := client.ExpectContinueResponse(t)
		if !ctResp.Body.AllThreadsContinued {
			t.Errorf("\ngot  %#v\nwant AllThreadsContinued=true", ctResp.Body)
		}
		client.ExpectTerminatedEvent(t)

		client.PauseRequest(1)
		switch r := client.ExpectMessage(t).(type) {
		case *dap.ErrorResponse:
			if r.Message != "Unable to halt execution" {
				t.Errorf("\ngot  %#v\nwant Message='Unable to halt execution'", r)
			}
		case *dap.PauseResponse:
		default:
			t.Fatalf("Unexpected response type: expect error or pause, got %#v", r)
		}

		client.DisconnectRequest()
		client.ExpectOutputEventDetaching(t)
		client.ExpectDisconnectResponse(t)
		client.ExpectTerminatedEvent(t)
	})
}

func TestStackTraceRequest(t *testing.T) {
	runTest(t, "StackTraceRequest", IncrementTestcase, func(server *DAPServer, client *Client, program string) {
		var stResp *dap.StackTraceResponse
		runDebugSessionWithBPs(t, client, func() {
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{3, 13},
			[]onBreakpoint{{
				execute: func() {
					const NumFrames = 4
					tests := map[string]struct {
						startFrame          int
						levels              int
						wantStartName       string
						wantStartLine       int
						wantStartFrame      int
						wantFramesReturned  int
						wantFramesAvailable int
						exact               bool
					}{
						"all frame levels from 0 to NumFrames":    {0, NumFrames, "Increment", 3, 0, NumFrames, NumFrames, true},
						"subset of frames from 1 to -1":           {1, NumFrames - 1, "Increment", 6, 1, NumFrames - 1, NumFrames, true},
						"load stack in pages: first half":         {0, NumFrames / 2, "Increment", 3, 0, NumFrames / 2, NumFrames, false},
						"load stack in pages: second half":        {NumFrames / 2, NumFrames, "Increment", 6, NumFrames / 2, NumFrames / 2, NumFrames, true},
						"load final stack":                        {NumFrames - 1, NumFrames, "__yak_main__", 11, NumFrames - 1, 1, NumFrames, true},
						"zero levels means all levels":            {0, 0, "Increment", 3, 0, NumFrames, NumFrames, true},
						"zero levels means all remaining levels":  {NumFrames / 2, 0, "Increment", 6, NumFrames / 2, NumFrames / 2, NumFrames, true},
						"negative levels treated as 0 (all)":      {0, -10, "Increment", 3, 0, NumFrames, NumFrames, true},
						"OOB levels is capped at available len":   {0, NumFrames + 1, "Increment", 3, 0, NumFrames, NumFrames, true},
						"OOB levels is capped at available len 1": {1, NumFrames + 1, "Increment", 6, 1, NumFrames - 1, NumFrames, true},
						"negative startFrame treated as 0":        {-10, 0, "Increment", 3, 0, NumFrames, NumFrames, true},
						"OOB startFrame returns empty trace":      {NumFrames, 0, "Increment", -1, -1, 0, NumFrames, true},
					}
					for name, tc := range tests {
						client.StackTraceRequest(1, tc.startFrame, tc.levels)
						stResp = client.ExpectStackTraceResponse(t)
						checkStackFramesNamed(name, t, stResp,
							tc.wantStartName, tc.wantStartLine, tc.wantStartFrame, tc.wantFramesReturned, tc.wantFramesAvailable, tc.exact)
					}

					// check all frames
					client.StackTraceRequest(1, 0, 0)
					stResp = client.ExpectStackTraceResponse(t)
					frames := stResp.Body.StackFrames
					want := []struct {
						wantName string
						wantLine int
						wantID   int
					}{
						{"Increment", 3, 3},
						{"Increment", 6, 2},
						{"Increment", 6, 1},
						{"__yak_main__", 11, 0},
					}
					for i, frame := range frames {
						if frame.Id != want[i].wantID {
							t.Errorf("got  %#v\nwant Id=%d\n", frame, want[i].wantID)
						}
						if want[i].wantLine > 0 && frame.Line != want[i].wantLine {
							t.Errorf("got  Line=%d\nwant %d\n", frame.Line, want[i].wantLine)
						}
						if want[i].wantName != "" && frame.Name != want[i].wantName {
							t.Errorf("got  Name=%s\nwant %s\n", frame.Name, want[i].wantName)
						}
					}
				},
				disconnect: false,
			},
				{
					execute: func() {
						client.StackTraceRequest(1, 0, 0)
						stResp = client.ExpectStackTraceResponse(t)
						checkStackFramesExact(t, stResp, "__yak_main__", 13, 0, 1, 1)
					},
					disconnect: false,
				}})
	})
}

func TestThreadsRequest(t *testing.T) {
	runTest(t, "ThreadsRequest", GoroutineTestcase, func(server *DAPServer, client *Client, program string) {
		runDebugSessionWithBPs(t, client, func() {
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{6},
			[]onBreakpoint{{
				execute: func() {
					checkStop(t, client, 1, "__yak_main__", 6)
					client.SetBreakpointsRequest(program, []int{3})
					client.ExpectSetBreakpointsResponse(t)

					client.ContinueRequest(1)
					client.ExpectContinueResponse(t)

					se := client.ExpectStoppedEvent(t)
					if se.Body.Reason != "breakpoint" || se.Body.ThreadId == 1 {
						t.Errorf("got %#v, want Reason=%q, ThreadId!=1", se, "breakpoint")
					}

					client.ThreadsRequest()
					tr := client.ExpectThreadsResponse(t)

					if len(tr.Body.Threads) != 2 {
						t.Errorf("got %d threads, expected 2\n", len(tr.Body.Threads))
					}

					var selectedFound bool
					for _, thread := range tr.Body.Threads {
						if thread.Id == se.Body.ThreadId {
							selectedFound = true
							break
						}
					}
					if !selectedFound {
						t.Errorf("got %#v, want ThreadId=%d\n", tr.Body.Threads, se.Body.ThreadId)
					}
				},
				disconnect: true,
			}})
	})
}

func TestScopesAndVairablesRequest(t *testing.T) {
	runTest(t, "ScopesAndVairablesRequest", VariablesTestcase, func(server *DAPServer, client *Client, program string) {
		runDebugSessionWithBPs(t, client, func() {
			server.config.extraLibs = TestExtraLibs
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{28},
			[]onBreakpoint{{
				execute: func() {
					var ref int
					frameID := 1

					client.StackTraceRequest(1, 0, 20)
					stack := client.ExpectStackTraceResponse(t)
					checkStackFramesExact(t, stack, "foobar", 28, 0, 2, 2)

					client.ScopesRequest(1)
					scopes := client.ExpectScopesResponse(t)
					checkScope(t, scopes, 0, "Globals", 1)
					checkScope(t, scopes, 1, "Locals1", 2)

					// Globals
					client.VariablesRequest(1)
					globals := client.ExpectVariablesResponse(t)

					checkVarRegex(t, globals, 0, "foobar", "foobar", "function", "yakvm.Function", noChildren)
					checkVarExact(t, globals, 1, "p1", "p1", "10", "int", noChildren)

					// Locals2
					locals2ScopeRef := 3
					client.VariablesRequest(locals2ScopeRef)
					locals := client.ExpectVariablesResponse(t)
					checkChildren(t, locals, "Locals2", 26)

					// checkVarExact(t, locals, -1, "a1", "a1", `foofoofoofoofoofoo`, "string", noChildren)
					// bool
					checkVarExact(t, locals, -1, "b1", "b1", "true", "bool", noChildren)
					checkVarExact(t, locals, -1, "b2", "b2", "false", "bool", noChildren)

					// int
					checkVarExact(t, locals, -1, "a2", "a2", "6", "int", noChildren)
					// neg turn into int64
					checkVarExact(t, locals, -1, "neg", "neg", "-1", "int64", noChildren)
					// int8
					checkVarExact(t, locals, -1, "i8", "i8", "1", "int", noChildren)
					// uint8
					checkVarExact(t, locals, -1, "u8", "u8", "255", "uint8", noChildren)
					// uint16
					checkVarExact(t, locals, -1, "u16", "u16", "65535", "int", noChildren)
					// uint32
					checkVarExact(t, locals, -1, "u32", "u32", "4294967295", "int", noChildren)
					// uint64
					checkVarExact(t, locals, -1, "u64", "u64", "9223372036854775807", "int", noChildren)
					// float64
					checkVarExact(t, locals, -1, "f1", "f1", "1.200000", "float64", noChildren)
					// string
					checkVarExact(t, locals, -1, "a10", "a10", `"ofo"`, "string", noChildren)
					// slice
					ref = checkVarRegex(t, locals, -1, "ba", "ba", `\[\]int\{`, `\[\]int`, hasChildren)
					client.VariablesRequest(ref)
					ba := client.ExpectVariablesResponse(t)
					checkChildren(t, ba, "ba", 201)

					ref = checkVarExact(t, locals, -1, "a4", "a4", "[]int{1, 2}", "[]int", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a4 := client.ExpectVariablesResponse(t)
						checkChildren(t, a4, "a4", 3)
						checkVarExact(t, a4, 1, "[0]", "a4[0]", "1", "int", noChildren)
						checkVarExact(t, a4, 2, "[1]", "a4[1]", "2", "int", noChildren)
					}
					ref = checkVarExact(t, locals, -1, "a5", "a5", "[]int{1, 2, 3, 4, 5}", "[]int", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a5 := client.ExpectVariablesResponse(t)
						checkChildren(t, a5, "a5", 6)
						checkVarExact(t, a5, 1, "[0]", "a5[0]", "1", "int", noChildren)
						checkVarExact(t, a5, 5, "[4]", "a5[4]", "5", "int", noChildren)
						validateEvaluateName(t, client, frameID, a5, 1)
						validateEvaluateName(t, client, frameID, a5, 2)
					}
					ref = checkVarExact(t, locals, -1, "a12", "a12", `[]dap.FooBar{dap.FooBar{Baz: 4, Bur: "d"}, dap.FooBar{Baz: 5, Bur: "e"}}`, "[]dap.FooBar", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a12 := client.ExpectVariablesResponse(t)
						checkChildren(t, a12, "a12", 3)
						checkVarExact(t, a12, 1, "[0]", "a12[0]", `dap.FooBar{Baz: 4, Bur: "d"}`, "dap.FooBar", hasChildren)
						ref = checkVarExact(t, a12, 2, "[1]", "a12[1]", `dap.FooBar{Baz: 5, Bur: "e"}`, "dap.FooBar", hasChildren)
						if ref > 0 {
							client.VariablesRequest(ref)
							a12_1 := client.ExpectVariablesResponse(t)
							checkChildren(t, a12_1, "a12[1]", 2)
							checkVarExact(t, a12_1, 0, "Baz", "a12[1].Baz", "5", "int", noChildren)
							checkVarExact(t, a12_1, 1, "Bur", "a12[1].Bur", `"e"`, "string", noChildren)
							validateEvaluateName(t, client, frameID, a12_1, 0)
							validateEvaluateName(t, client, frameID, a12_1, 1)
						}
					}
					ref = checkVarExact(t, locals, -1, "a13", "a13", `[]*dap.FooBar{&dap.FooBar{Baz: 6, Bur: "f"}, &dap.FooBar{Baz: 7, Bur: "g"}, &dap.FooBar{Baz: 8, Bur: "h"}}`, "[]*dap.FooBar", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a13 := client.ExpectVariablesResponse(t)
						checkChildren(t, a13, "a13", 4)
						checkVarExact(t, a13, 1, "[0]", "a13[0]", `&dap.FooBar{Baz: 6, Bur: "f"}`, "*dap.FooBar", hasChildren)
						checkVarExact(t, a13, 2, "[1]", "a13[1]", `&dap.FooBar{Baz: 7, Bur: "g"}`, "*dap.FooBar", hasChildren)
						ref = checkVarExact(t, a13, 3, "[2]", "a13[2]", `&dap.FooBar{Baz: 8, Bur: "h"}`, "*dap.FooBar", hasChildren)
						if ref > 0 {
							client.VariablesRequest(ref)
							a13_2 := client.ExpectVariablesResponse(t)
							checkChildren(t, a13_2, "a13[2]", 2)
							checkVarExact(t, a13_2, 0, "Baz", "a13[2].Baz", "8", "int", noChildren)
							checkVarExact(t, a13_2, 1, "Bur", "a13[2].Bur", `"h"`, "string", noChildren)
							validateEvaluateName(t, client, frameID, a13_2, 0)
							validateEvaluateName(t, client, frameID, a13_2, 1)
						}
					}

					ref = checkVarExact(t, locals, -1, "ni", "ni", `[]interface {}{[]interface {}{123}}`, `[]interface {}`, hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						ni := client.ExpectVariablesResponse(t)
						checkChildren(t, ni, "ni", 2)
						ref = checkVarExact(t, ni, 1, "[0]", "ni[0]", `[]interface {}{123}`, `interface {}`, hasChildren)
						if ref > 0 {
							client.VariablesRequest(ref)
							ni_0 := client.ExpectVariablesResponse(t)
							checkChildren(t, ni_0, "ni[0]", 2)
							checkVarExact(t, ni_0, 1, "[0]", "ni[0][0]", `123`, `interface {}`, noChildren)
						}
					}

					ref = checkVarExact(t, locals, -1, "ms", "ms", `dap.Nest{Level: 0, Nest: &dap.Nest{Level: 1, Nest: &dap.Nest{Level: 2, Nest: &dap.Nest{Level: 3, Nest: &dap.Nest{Level: 4, Nest: (*dap.Nest)(nil)}}}}}`, "dap.Nest", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						ms := client.ExpectVariablesResponse(t)
						checkChildren(t, ms, "ms", 2)
						checkVarExact(t, ms, 0, "Level", "ms.Level", "0", "int", noChildren)
						ref = checkVarExact(t, ms, 1, "Nest", "ms.Nest", `&dap.Nest{Level: 1, Nest: &dap.Nest{Level: 2, Nest: &dap.Nest{Level: 3, Nest: &dap.Nest{Level: 4, Nest: (*dap.Nest)(nil)}}}}`, "*dap.Nest", hasChildren)
						validateEvaluateName(t, client, frameID, ms, 0)
						validateEvaluateName(t, client, frameID, ms, 1)
						if ref > 0 {
							client.VariablesRequest(ref)
							ms_1 := client.ExpectVariablesResponse(t)
							checkChildren(t, ms_1, "ms.Nest", 2)
							checkVarExact(t, ms_1, 0, "Level", "ms.Nest.Level", "1", "int", noChildren)
							ref = checkVarExact(t, ms_1, 1, "Nest", "ms.Nest.Nest", `&dap.Nest{Level: 2, Nest: &dap.Nest{Level: 3, Nest: &dap.Nest{Level: 4, Nest: (*dap.Nest)(nil)}}}`, "*dap.Nest", hasChildren)
							validateEvaluateName(t, client, frameID, ms_1, 0)
							validateEvaluateName(t, client, frameID, ms_1, 1)
							if ref > 0 {
								client.VariablesRequest(ref)
								ms_1_1 := client.ExpectVariablesResponse(t)
								checkChildren(t, ms_1_1, "ms.Nest.Nest", 2)
								checkVarExact(t, ms_1_1, 0, "Level", "ms.Nest.Nest.Level", "2", "int", noChildren)
								ref = checkVarExact(t, ms_1_1, 1, "Nest", "ms.Nest.Nest.Nest", `&dap.Nest{Level: 3, Nest: &dap.Nest{Level: 4, Nest: (*dap.Nest)(nil)}}`, "*dap.Nest", hasChildren)
								validateEvaluateName(t, client, frameID, ms_1_1, 0)
								validateEvaluateName(t, client, frameID, ms_1_1, 1)
								if ref > 0 {
									client.VariablesRequest(ref)
									ms_1_1_1 := client.ExpectVariablesResponse(t)
									checkChildren(t, ms_1_1_1, "ms.Nest.Nest.Nest", 2)
									checkVarExact(t, ms_1_1_1, 0, "Level", "ms.Nest.Nest.Nest.Level", "3", "int", noChildren)
									ref = checkVarExact(t, ms_1_1_1, 1, "Nest", "ms.Nest.Nest.Nest.Nest", `&dap.Nest{Level: 4, Nest: (*dap.Nest)(nil)}`, "*dap.Nest", hasChildren)
								}
							}
						}
					}

					// array
					ref = checkVarExact(t, locals, -1, "a11", "a11", `[3]dap.FooBar{dap.FooBar{Baz: 1, Bur: "a"}, dap.FooBar{Baz: 2, Bur: "b"}, dap.FooBar{Baz: 3, Bur: "c"}}`, "[3]dap.FooBar", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a11 := client.ExpectVariablesResponse(t)
						checkChildren(t, a11, "a11", 4)
						checkVarExact(t, a11, 1, "[0]", "a11[0]", `dap.FooBar{Baz: 1, Bur: "a"}`, "dap.FooBar", hasChildren)
						ref = checkVarExact(t, a11, 2, "[1]", "a11[1]", `dap.FooBar{Baz: 2, Bur: "b"}`, "dap.FooBar", hasChildren)
						if ref > 0 {
							client.VariablesRequest(ref)
							a11_1 := client.ExpectVariablesResponse(t)
							checkChildren(t, a11_1, "a11[1]", 2)
							checkVarExact(t, a11_1, 0, "Baz", "a11[1].Baz", "2", "int", noChildren)
							checkVarExact(t, a11_1, 1, "Bur", "a11[1].Bur", `"b"`, "string", noChildren)
							validateEvaluateName(t, client, frameID, a11_1, 0)
							validateEvaluateName(t, client, frameID, a11_1, 1)
						}
						checkVarExact(t, a11, 3, "[2]", "a11[2]", `dap.FooBar{Baz: 3, Bur: "c"}`, "dap.FooBar", hasChildren)
					}

					// struct
					ref = checkVarExact(t, locals, -1, "a6", "a6", `dap.FooBar{Baz: 8, Bur: "word"}`, "dap.FooBar", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a6 := client.ExpectVariablesResponse(t)
						checkChildren(t, a6, "a6", 2)
						checkVarExact(t, a6, 0, "Baz", "a6.Baz", "8", "int", noChildren)
						checkVarExact(t, a6, 1, "Bur", "a6.Bur", `"word"`, "string", noChildren)
					}
					ref = checkVarExact(t, locals, -1, "a8", "a8", `dap.FooBar2{Bur: 10, Baz: "feh"}`, "dap.FooBar2", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a8 := client.ExpectVariablesResponse(t)
						checkChildren(t, a8, "a8", 2)
						checkVarExact(t, a8, 0, "Bur", "a8.Bur", "10", "int", noChildren)
						checkVarExact(t, a8, 1, "Baz", "a8.Baz", `"feh"`, "string", noChildren)
					}

					// struct ptr
					ref = checkVarExact(t, locals, -1, "a7", "a7", `&dap.FooBar{Baz: 5, Bur: "strum"}`, "*dap.FooBar", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						a7 := client.ExpectVariablesResponse(t)
						checkChildren(t, a7, "a7", 2)
						checkVarExact(t, a7, 0, "Baz", "a7.Baz", "5", "int", noChildren)
						checkVarExact(t, a7, 1, "Bur", "a7.Bur", `"strum"`, "string", noChildren)
						validateEvaluateName(t, client, frameID, a7, 0)
						validateEvaluateName(t, client, frameID, a7, 1)

					}

					// map
					ref = checkVarExact(t, locals, -1, "mp", "mp", `map[int]interface {}{1: 42, 2: 43}`, "map[int]interface {}", hasChildren)
					if ref > 0 {
						client.VariablesRequest(ref)
						mp := client.ExpectVariablesResponse(t)
						checkChildren(t, mp, "mp", 5)
						checkVarExact(t, mp, 1, "[key 0]", "", "1", "int", noChildren)
						checkVarExact(t, mp, 2, "[value 0]", "mp[1]", "42", "interface {}", noChildren)
						checkVarExact(t, mp, 3, "[key 1]", "", "2", "int", noChildren)
						checkVarExact(t, mp, 4, "[value 1]", "mp[2]", "43", "interface {}", noChildren)
					}

					// nil
					checkVarExact(t, locals, -1, "a9", "a9", "(*dap.FooBar)(nil)", "*dap.FooBar", noChildren)

				},
				disconnect: false,
			}},
		)
	})
}
func TestEvaluateCommandRequest(t *testing.T) {
	runTest(t, "EvaluateCommandRequest", SimpleYakTestCase, func(server *DAPServer, client *Client, program string) {
		runDebugSessionWithBPs(t, client, func() {
			server.config.extraLibs = TestExtraLibs
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{2},
			[]onBreakpoint{{
				execute: func() {
					checkStop(t, client, 1, "__yak_main__", 2)
					client.EvaluateRequest("dbg help", 0, "repl")
					got := client.ExpectEvaluateResponse(t)
					checkEval(t, got, HelpInfo, noChildren)
				},
				disconnect: true,
			}},
		)
	},
	)
}

func TestStepAndNextRequest(t *testing.T) {
	runTest(t, "StepAndNextRequest", StepAndNExtTestcase, func(server *DAPServer, client *Client, program string) {
		runDebugSessionWithBPs(t, client, func() {
			server.config.extraLibs = TestExtraLibs
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{7},
			[]onBreakpoint{{
				execute: func() {
					checkStop(t, client, 1, "initialize", 7)
					expectStop := func(fun string, line int) {
						t.Helper()
						se := client.ExpectStoppedEvent(t)
						if se.Body.Reason != "step" || se.Body.ThreadId != 1 || !se.Body.AllThreadsStopped {
							t.Errorf("got %#v, want Reason=\"step\", ThreadId=1, AllThreadsStopped=true", se)
						}
						checkStop(t, client, 1, fun, line)
					}
					client.StepOutRequest(0)
					client.ExpectStepOutResponse(t)
					expectStop("main", 12)

					client.NextRequest(0)
					client.ExpectNextResponse(t)
					expectStop("main", 13)

					client.StepInRequest(0)
					client.ExpectStepInResponse(t)
					expectStop("square", 1)

					client.NextRequest(0)
					client.ExpectNextResponse(t)
					expectStop("square", 2)
				},
				disconnect: true,
			}},
		)
	},
	)
}

// func TestHardCodedBreakpoints(t *testing.T) {
// 	runTest(t, "HardCodedBreakpoints", HardCodeBreakPointTestcase, func(server *DAPServer, client *Client, program string) {
// 		runDebugSessionWithBPs(t, client, func() {
// 			server.config.extraLibs = TestExtraLibs
// 			client.LaunchRequest("exec", program, !StopOnEntry)
// 		}, program,
// 			[]int{10},
// 			[]onBreakpoint{{
// 				execute: func() {
// 					checkStop(t, client, 0, "main", 10)
// 					client.ContinueRequest(0)
// 					client.ExpectContinueResponse(t)
// 					se := client.ExpectStoppedEvent(t)
// 					_ = se
// 					if se.Body.ThreadId != 0 || se.Body.Reason != "breakpoint" {
// 						t.Errorf("\ngot  %#v\nwant ThreadId=0 Reason=\"breakpoint\"", se)
// 					}
// 					checkStop(t, client, 0, "f", 3)
// 				},
// 				disconnect: true,
// 			}},
// 		)
// 	},
// 	)
// }

func TestPanicBreakpointOnNext(t *testing.T) {
	runTest(t, "panic", PanicTestcase, func(server *DAPServer, client *Client, program string) {
		runDebugSessionWithBPs(t, client, func() {
			server.config.extraLibs = TestExtraLibs
			client.LaunchRequest("exec", program, !StopOnEntry)
		}, program,
			[]int{3},
			[]onBreakpoint{{
				execute: func() {
					checkStop(t, client, 1, "main", 3)

					client.NextRequest(0)
					client.ExpectNextResponse(t)

					text := "\"BOOM!\""
					se := client.ExpectStoppedEvent(t)
					if se.Body.ThreadId != 1 || se.Body.Reason != "exception" || se.Body.Description != "panic" || se.Body.Text != text {
						t.Errorf("\ngot  %#v\nwant ThreadId=1 Reason=\"exception\" Description=\"panic\" Text=%q", se, text)
					}

					client.ExceptionInfoRequest(1)
					eInfo := client.ExpectExceptionInfoResponse(t)
					if eInfo.Body.ExceptionId != "panic" || eInfo.Body.Description != text {
						t.Errorf("\ngot  %#v\nwant ExceptionId=\"panic\" Description=%q", eInfo, text)
					}
				},
				disconnect: true,
			}},
		)
	},
	)
}
