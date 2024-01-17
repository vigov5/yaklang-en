package yakast

import (
	"github.com/google/uuid"
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func (y *YakCompiler) VisitTryStmt(raw yak.ITryStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}
	i, _ := raw.(*yak.TryStmtContext)
	if i == nil {
		return nil
	}
	y.writeString("try ")
	recoverFormatBufferFunc := y.switchFormatBuffer()
	// Assign error
	recoverSymbolTableAndScope := y.SwitchSymbolTableInNewScope("try-catch-finally", uuid.New().String())

	var id = -1
	var text string
	if identifier := i.Identifier(); identifier != nil {
		text = identifier.GetText()
		id1, err := y.currentSymtbl.NewSymbolWithReturn(text)
		if err != nil {
			y.panicCompilerError(CreateSymbolError, text)
		}
		id = id1
	}

	// Capture the exception that may occur in the try block
	catchErrorOpCode := y.pushOperator(yakvm.OpCatchError) //Start catching the error
	y.tryDepthStack.Push(y.GetNextCodeIndex())
	y.VisitBlock(i.Block(0))
	y.tryDepthStack.Pop()
	y.pushOperator(yakvm.OpStopCatchError) // End catching the error
	jmp1 := y.pushJmp()                    // Jump to the finally block after executing the try block
	y.writeString(recoverFormatBufferFunc())
	y.writeString(" catch ")
	if text != "" {
		y.writeString(text + " ")
	}
	recoverFormatBufferFunc = y.switchFormatBuffer()
	catchErrorOpCode.Op1 = yakvm.NewAutoValue(y.GetCodeIndex()) // Jump to the catch block after catching the exception
	catchErrorOpCode.Op2 = yakvm.NewAutoValue(id)
	y.VisitBlock(i.Block(1)) // catch block
	y.writeString(recoverFormatBufferFunc())
	jmp1.Unary = y.GetNextCodeIndex()
	if finallyBlock := i.Block(2); finallyBlock != nil {
		y.writeString("finally")
		recoverFormatBufferFunc = y.switchFormatBuffer()
		y.VisitBlock(finallyBlock)
		y.writeString(recoverFormatBufferFunc())
	}
	recoverSymbolTableAndScope()
	return nil
}
