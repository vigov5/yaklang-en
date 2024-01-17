package yakast

import (
	"fmt"
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

	uuid "github.com/satori/go.uuid"
)

func (y *YakCompiler) VisitGoStmt(raw yak.IGoStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.GoStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("go ")

	// go expr call;
	// First of all, go comes with a pop, which is to flatten the stack
	// . Because the execution stack used for the expression after go should not be the same as others, so it should be given to him Open a new virtual machine and set the starting symbol table
	// . The following content should be
	//  ...
	//  ...
	// 	...
	//  ...
	//  call n and change it to async-call

	id := fmt.Sprintf("go/%v", uuid.NewV4().String())
	_ = id

	if code := i.InstanceCode(); code != nil {
		y.VisitInstanceCode(i.InstanceCode())
	} else {
		y.VisitExpression(i.Expression())
		y.VisitFunctionCall(i.FunctionCall())
	}

	/*
		Create a new Go instruction
	*/
	if lastCode := y.codes[y.GetCodeIndex()]; lastCode.Opcode == yakvm.OpCall {
		// Function instruction
		lastCode.Opcode = yakvm.OpAsyncCall
	}
	return nil
}
