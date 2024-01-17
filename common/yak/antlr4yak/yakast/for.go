package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

	uuid "github.com/satori/go.uuid"
)

type forContext struct {
	startCodeIndex       int
	continueScopeCounter int
	breakScopeCounter    int
	forRangeMode         bool
}

func (y *YakCompiler) enterForContext(start int) {
	y.forDepthStack.Push(&forContext{
		startCodeIndex: start,
	})
}

func (y *YakCompiler) enterForRangeContext(start int) {
	y.forDepthStack.Push(&forContext{
		startCodeIndex: start,
		forRangeMode:   true,
	})
}

func (y *YakCompiler) peekForContext() *forContext {
	raw, ok := y.forDepthStack.Peek().(*forContext)
	if ok {
		return raw
	} else {
		return nil
	}
}

func (y *YakCompiler) exitForContext(end int, continueIndex int) {
	start := y.peekForStartIndex()
	if start < 0 {
		return
	}

	for _, c := range y.codes[start:] {
		if c.Opcode == yakvm.OpBreak && c.Unary <= 0 {
			// Set The jump value of the Break Code of all statements from the beginning to the end of for
			c.Unary = end
			if y.peekForContext().forRangeMode {
				c.Op2 = yakvm.NewIntValue(1) // for range mode
			}
		}

		if c.Opcode == yakvm.OpContinue && c.Unary <= 0 {
			if !y.peekForContext().forRangeMode {
				c.Op1.Value = c.Op1.Value.(int) - 1
			}
			c.Unary = continueIndex
		}
	}

	y.forDepthStack.Pop()
}

func (y *YakCompiler) VisitForStmt(raw yak.IForStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ForStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("for ")

	// _ Record the starting index, usually when continue,
	startIndex := y.GetNextCodeIndex()
	startIndex += 1 // skip new-scope instruction
	y.enterForContext(startIndex)

	var endThirdExpr yak.IForThirdExprContext

	f := y.SwitchSymbolTableInNewScope("for-legacy", uuid.NewV4().String())

	var toEnds []*yakvm.Code
	var conditionSymbol int
	if e := i.Expression(); e != nil {
		y.VisitExpression(e)
		toEnd := y.pushJmpIfFalse()
		toEnds = append(toEnds, toEnd)
	} else if cond := i.ForStmtCond(); cond != nil {
		condIns := cond.(*yak.ForStmtCondContext)
		if entry := condIns.ForFirstExpr(); entry != nil {
			y.VisitForFirstExpr(entry)
		}
		y.writeString("; ")

		startIndex = y.GetNextCodeIndex()
		if condIns.Expression() != nil {
			conditionSymbol = y.currentSymtbl.NewSymbolWithoutName()
			y.pushLeftRef(conditionSymbol)
			y.VisitExpression(condIns.Expression())
			// for the following We can judge whether to execute the third statement based on the condition. We need to cache the result into the intermediate symbol
			y.pushOperator(yakvm.OpFastAssign)
			// The condition should be forEnd not blockEnd
			toEnd := y.pushJmpIfFalse()
			toEnds = append(toEnds, toEnd)
		}
		y.writeString("; ")

		if e := condIns.ForThirdExpr(); e != nil {
			endThirdExpr = e
		}
	}
	// after the execution body ends, it should jump back unconditionally At the beginning, re-judge
	// on the left. However, the third statement for ;; . It should be a block. After executing the explanation, execute the third statement
	recoverFormatBufferFunc := y.switchFormatBuffer()
	y.VisitBlock(i.Block())
	buf := recoverFormatBufferFunc()

	// continue index
	continueIndex := y.GetNextCodeIndex()

	if endThirdExpr != nil {
		if conditionSymbol > 0 {
			y.pushRef(conditionSymbol)
			toEnd := y.pushJmpIfFalse()
			toEnds = append(toEnds, toEnd)
		}
		y.VisitForThirdExpr(endThirdExpr)
		y.writeString(" ")
	}
	y.writeString(buf)
	y.pushJmp().Unary = startIndex
	forEnd := y.GetNextCodeIndex()

	f()
	// of a slice that has not been set in the parsed block. break
	y.exitForContext(forEnd+1, continueIndex)

	// sets the toEnd position
	for _, toEnd := range toEnds {
		if toEnd != nil {
			toEnd.Unary = forEnd
		}
	}

	return nil
}

func (y *YakCompiler) VisitForRangeStmt(raw yak.IForRangeStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ForRangeStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("for ")

	recoverFormatBufferFunc := y.switchFormatBuffer()
	expr := i.Expression()
	if expr == nil {
		y.panicCompilerError(compileError, "for-range/in need expression in right value at least")
	}

	recoverSymtbl := y.SwitchSymbolTableInNewScope("for", uuid.NewV4().String())
	defer recoverSymtbl()

	defaultValueSymbol, err := y.currentSymtbl.NewSymbolWithReturn("_")
	if err != nil {
		y.panicCompilerError(compileError, "cannot create `_` variable, reason: "+err.Error())
	}

	/*
		for range: range After the expression is calculated,
	*/
	// ! A new symbol should not be created for the iterated object. This will cause problems with self-modification, and the lvalue pointed to by the right value is modified after iteration, and all subsequent self-modifications will be invalid.
	// expressionResultID := y.currentSymtbl.NewSymbolWithoutName()
	// y.pushLeftRef(expressionResultID)
	// enter for-range
	y.VisitExpression(expr)
	buf := recoverFormatBufferFunc()
	// y.pushOperator(yakvm.OpFastAssign)
	defer y.pushOpPop()

	// . OpEnterFR will pop the rvalue from the stack to create an iterator. This value does not need to be popped out. Peek each time is enough.
	enterFR := y.pushEnterFR()

	// _ Record the starting index, which is usually the
	startIndex := y.GetNextCodeIndex()
	y.enterForRangeContext(startIndex)

	// Calculate the number of range lvalues 
	n := 0
	if l := i.LeftExpressionList(); l != nil { // access lvalue
		n = len(l.(*yak.LeftExpressionListContext).AllLeftExpression())
		// General For example, there are two lvalues. The assignment of one and two lvalues is different. This depends on the specific implementation of
		// But under for-range it cannot be greater than 2, which is a problem
		if n > 2 && i.Range() != nil {
			y.panicCompilerError(compileError, "`for ... range` accept up to tow left expression value")
		}
	}

	// peek ExpressionResultID Use RangeNext or InNext to iteratively calculate
	// is symbolized. After the iterative calculation, it should be a list, which can be used as the rvalue of the assignment
	rightAtLeast := 1
	if n > 1 {
		rightAtLeast = n
	}

	var nextCode *yakvm.Code
	if i.In() != nil {
		nextCode = y.pushInNext(rightAtLeast)
	} else {
		nextCode = y.pushRangeNext(rightAtLeast) // . Traverse the previous expression
	}

	// of a map, if there is no lvalue, should some values be retained? Of course, _ should be assigned the number of current loops or the first value
	if n <= 0 {
		y.pushLeftRef(defaultValueSymbol)
		y.pushListWithLen(1)
	} else {
		n = y.VisitLeftExpressionList(true, i.LeftExpressionList())
		if n == -1 {
			y.panicCompilerError(compileError, "invalid left expression list")
		}
	}
	y.pushOperator(yakvm.OpAssign) // and assign it to the variable
	if op, op2 := i.In(), i.Range(); op != nil || op2 != nil {
		if op != nil {
			y.writeStringWithWhitespace(op.GetText())
		} else {
			eq, ceq := i.AssignEq(), i.ColonAssignEq()
			if eq != nil {
				y.writeStringWithWhitespace(eq.GetText())
			} else if ceq != nil {
				y.writeStringWithWhitespace(ceq.GetText())
			}
			y.writeString(op2.GetText() + " ")
		}
	}
	y.writeString(buf + " ")
	y.VisitBlock(i.Block())

	exitFR := y.GetNextCodeIndex()
	// exit for-range

	// to set the jump position of the next code. The pipe
	nextCode.Op1 = yakvm.NewIntValue(exitFR)

	forEnd := y.GetNextCodeIndex()

	// of a slice that has not been set in the parsed block. break
	y.exitForContext(forEnd+2, forEnd)

	// sets the jump of enterFR. If it is empty, jump directly to
	enterFR.Unary = forEnd + 1
	y.pushExitFR(startIndex)

	// for range starting position
	// . The number of loops is assigned through _ variables, and the exit condition is range
	// The difference is that for range needs to support three conditions at least
	//   1. Set the range
	//   2. For the range
	//   3. For an integer range, in this case golang does not have
	//  	that comes with the condition. Expected to be for range 4 { println(1) } used to close will print 4 1\n, which is equivalent to for range [0,1, 2,3] {}...

	return nil
}

func (y *YakCompiler) VisitForThirdExpr(raw yak.IForThirdExprContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ForThirdExprContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	if ae := i.AssignExpression(); ae != nil {
		// that jumps unconditionally when continue. The copied expression is a flat stack, and no additional pop is required.
		y.VisitAssignExpression(ae)
		return nil
	}

	if e := i.Expression(); e != nil {
		y.VisitExpression(e)
		y.pushOpPop()
		return nil
	}

	y.panicCompilerError(compileError, "visit first for expr failed")

	return nil
}

func (y *YakCompiler) VisitForFirstExpr(raw yak.IForFirstExprContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ForFirstExprContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	if ae := i.AssignExpression(); ae != nil {
		// that jumps unconditionally when continue. The copied expression is a flat stack, and no additional pop is required.
		y.VisitAssignExpression(ae)
		return nil
	}

	if e := i.Expression(); e != nil {
		y.VisitExpression(e)
		y.pushOpPop()
		return nil
	}

	y.panicCompilerError(compileError, "visit first for expr failed")
	return nil
}
