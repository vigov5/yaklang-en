package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func (y *YakCompiler) VisitDeferStmt(raw yak.IDeferStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.DeferStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("defer ")

	finished := y.SwitchCodes()
	y.VisitExpression(i.Expression())
	// is executed.
	y.pushOpPop()
	funcCode := make([]*yakvm.Code, len(y.codes))
	copy(funcCode, y.codes)
	finished()

	// defer is a statement that guarantees that the stack will be flattened after
	y.pushDefer(funcCode)

	return nil
}
