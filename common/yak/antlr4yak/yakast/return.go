package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func (y *YakCompiler) VisitReturnStmt(raw yak.IReturnStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ReturnStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("return")

	// This is a stack push operation. The virtual machine needs to record the return value, so return is required as OPCODE to operate the stack.
	if list := i.ExpressionList(); list != nil {
		y.writeString(" ")
		y.VisitExpressionList(list)
	}
	if y.tryDepthStack.Len() > 0 {
		y.pushOperator(yakvm.OpStopCatchError)
	}
	y.pushOperator(yakvm.OpReturn)

	return nil
}
