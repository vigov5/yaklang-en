package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

	"github.com/google/uuid"
)

func (y *YakCompiler) VisitIfStmt(raw yak.IIfStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.IfStmtContext)
	if i == nil {
		return nil
	}

	ifBlock := i.Block(0)
	if ifBlock == nil {
		y.panicCompilerError(compileError, "no if code block")
	}
	recoverRange := y.SetRange(ifBlock)
	defer recoverRange()
	y.writeString("if ")

	tableRecover := y.SwitchSymbolTableInNewScope("if", uuid.New().String())

	ifCond := i.Expression(0)
	if ifCond == nil {
		y.panicCompilerError(compileError, "no if condition")
	}
	y.VisitExpression(ifCond)
	y.writeString(" ")

	// if condition is true, and the if statement block
	// is implemented using jmpf. If the value after pop stack is considered false, then Jump to
	var jmpfCode = y.pushJmpIfFalse()

	// for else. Compile block

	y.VisitBlock(ifBlock)

	var jmpToEnd []*yakvm.Code
	jmpToEnd = append(jmpToEnd, y.pushJmp())

	// is executed for jmpf.
	jmpfCode.Unary = y.GetNextCodeIndex()
	tableRecover()
	for index := range i.AllElif() {
		tableRecover = y.SwitchSymbolTableInNewScope("elif", uuid.New().String())
		// . The logic of elif and if is exactly the same. Reading an expression
		// and then use jmpf to jump to
		// But elif has a special feature, that is, after the if statement block is executed, the
		// skips the elif statement block, so you need to add a jmp instruction
		y.writeStringWithWhitespace("elif")
		y.VisitExpression(i.Expression(index + 1))
		y.writeString(" ")
		var jmpfCode = y.pushJmpIfFalse()
		y.VisitBlock(i.Block(index + 1))
		jmpToEnd = append(jmpToEnd, y.pushJmp())

		jmpfCode.Unary = y.GetNextCodeIndex()
		tableRecover()
	}

	// at the end of the elif statement block. Set the end character
	if ielseBlock := i.ElseBlock(); ielseBlock != nil {
		elseBlock := ielseBlock.(*yak.ElseBlockContext)
		y.writeStringWithWhitespace("else")
		block := elseBlock.Block()
		elseIf := elseBlock.IfStmt()
		if block != nil {
			tableRecover = y.SwitchSymbolTableInNewScope("else", uuid.New().String())
			y.VisitBlock(block)
			tableRecover()
		} else if elseIf != nil {
			y.VisitIfStmt(elseIf)
		}

	}

	endCode := y.GetCodeIndex()
	for _, jmp := range jmpToEnd {
		jmp.Unary = endCode
	}

	return nil
}
