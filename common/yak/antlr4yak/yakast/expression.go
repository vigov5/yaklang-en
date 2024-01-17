package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func (y *YakCompiler) VisitExpressionStmt(raw yak.IExpressionStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ExpressionStmtContext)
	if i == nil {
		return nil
	}

	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	y.VisitExpression(i.Expression())

	// only executes the expression and pushes the stack, which will cause the stack to grow without limit. An expr stmt needs to maintain the balance of the stack, so
	y.pushOpPop()
	return nil
}

func (y *YakCompiler) VisitExpressionList(raw yak.IExpressionListContext) int {
	if y == nil || raw == nil {
		return -1
	}

	i, _ := raw.(*yak.ExpressionListContext)
	if i == nil {
		return -1
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	// If expression is not one, the list is an expression, and executing list 0 directly means skipping the instruction
	exprs := i.AllExpression()
	LenOfExprs := len(exprs)
	defer y.pushListWithLen(LenOfExprs)
	for index, e := range exprs {
		y.VisitExpression(e)
		if index < LenOfExprs-1 {
			y.writeString(", ")
		}
	}

	return LenOfExprs
}

func (y *YakCompiler) VisitExpression(raw yak.IExpressionContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.ExpressionContext)
	if i == nil {
		return nil
	}

	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	if op := i.TypeLiteral(); op != nil {
		if i.LParen() != nil && i.RParen() != nil {
			recoverFormatBufferFunc := y.switchFormatBuffer()
			y.writeString("(")
			expression := i.Expression(0)
			if expression == nil {
				y.pushUndefined()
			} else {
				y.VisitExpression(expression)
			}
			y.writeString(")")
			buf := recoverFormatBufferFunc()
			y.VisitTypeLiteral(op)
			y.writeString(buf)
			y.pushOperator(yakvm.OpTypeCast)
		}
	} else if op := i.Recover(); op != nil {
		y.writeString(op.GetText() + "()")
		y.pushOperator(yakvm.OpRecover)
	} else if op := i.Panic(); op != nil {
		y.writeString(op.GetText() + "(")
		y.VisitExpression(i.Expression(0))
		y.writeString(")")
		y.pushOperator(yakvm.OpPanic)
	} else if s := i.Literal(); s != nil { // needs to be popped. Parse a single literal
		y.VisitLiteral(s)
	} else if s := i.Identifier(); s != nil { // parses the variable
		y.writeString(s.GetText())
		// . When encountering a variable, use the symbol in the expression!
		sym, ok := y.currentSymtbl.GetSymbolByVariableName(s.GetText())
		if !ok {
			y.pushIdentifierName(s.GetText())
			if y.strict {
				id := s.GetText()
				if _, ok := y.extVarsMap[id]; !ok {
					y.currentStartPosition.ColumnNumber += 1
					y.currentEndPosition.ColumnNumber += 2
					err := y.newError(y.GetConstError(notFoundVariable), id)
					y.compilerErrors.Push(err)
					if y.currentSymtbl != y.rootSymtbl {
						info := y.contextInfo.Peek()
						if v, ok := info.(string); !ok || v != "InstanceCode" {
							y.indeterminateUndefinedVar = append(y.indeterminateUndefinedVar, [2]any{id, err})
						}
					}
				}
			}
			return nil
		}
		y.pushRef(sym)
	} else if f := i.FunctionCall(); f != nil { // function call or other atomic operation (.ref / ref() / slice[]）
		y.VisitExpression(i.Expression(0))
		y.VisitFunctionCall(f)
	} else if op := i.AnonymousFunctionDecl(); op != nil { // function declaration ( )=>{}
		y.VisitAnonymousFunctionDecl(op)
	} else if pE := i.ParenExpression(); pE != nil { // '(' expression? ')'
		i := pE.(*yak.ParenExpressionContext)
		if e := i.Expression(); e != nil {
			// . There is an expression
			y.writeString("(")
			y.VisitExpression(e)
			y.writeString(")")
		} else {
			// has only parentheses and no expression.
			y.writeString("()")
			y.pushUndefined()
		}
	} else if op := i.UnaryOperator(); op != nil { // unary op
		y.writeString(op.GetText())
		y.VisitExpression(i.Expression(0))
		opStr := op.GetText()
		switch opStr {
		case "!":
			y.pushOperator(yakvm.OpNot)
		case "+":
			y.pushOperator(yakvm.OpPlus)
		case "-":
			y.pushOperator(yakvm.OpNeg)
		case "<-":
			y.pushOperator(yakvm.OpChan)
		case "^":
			y.pushOperator(yakvm.OpBitwiseNot)
		default:
			y.panicCompilerError(notImplemented, opStr)
		}
	} else if op := i.BitBinaryOperator(); op != nil { // bit binary op
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		y.VisitExpression(i.Expression(1))
		opStr := op.GetText()
		switch opStr {
		case "<<":
			y.pushOperator(yakvm.OpShl)
		case ">>":
			y.pushOperator(yakvm.OpShr)
		case "&":
			y.pushOperator(yakvm.OpAnd)
		case "&^":
			y.pushOperator(yakvm.OpAndNot)
		case "|":
			y.pushOperator(yakvm.OpOr)
		case "^":
			y.pushOperator(yakvm.OpXor)
		default:
			y.panicCompilerError(bitBinaryError, opStr)
		}
	} else if op := i.MultiplicativeBinaryOperator(); op != nil { // op * / %
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		y.VisitExpression(i.Expression(1))
		opStr := op.GetText()
		switch opStr {
		case "*":
			y.pushOperator(yakvm.OpMul)
		case "/":
			y.pushOperator(yakvm.OpDiv)
		case "%":
			y.pushOperator(yakvm.OpMod)
		default:
			y.panicCompilerError(multiplicativeBinaryError, opStr)
		}
	} else if op := i.AdditiveBinaryOperator(); op != nil { // - +
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		y.VisitExpression(i.Expression(1))

		opStr := op.GetText()
		switch opStr {
		case "+":
			y.pushOperator(yakvm.OpAdd)
		case "-":
			y.pushOperator(yakvm.OpSub)
		}
	} else if op := i.ComparisonBinaryOperator(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		y.VisitExpression(i.Expression(1))
		switch op.GetText() {
		case `>`:
			y.pushOperator(yakvm.OpGt)
		case `<`:
			y.pushOperator(yakvm.OpLt)
		case `<=`:
			y.pushOperator(yakvm.OpLtEq)
		case `>=`:
			y.pushOperator(yakvm.OpGtEq)
		case `!=`, `<>`:
			y.pushOperator(yakvm.OpNotEq)
		case `==`:
			y.pushOperator(yakvm.OpEq)
		}
	} else if op := i.ChanIn(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		y.VisitExpression(i.Expression(1))
		y.pushOperator(yakvm.OpSendChan)
	} else if op := i.In(); op != nil {
		y.VisitExpression(i.Expression(0))
		text := op.GetText()
		if op2 := i.NotLiteral(); op2 != nil {
			text = "not " + text
		}
		y.writeStringWithWhitespace(text)
		y.VisitExpression(i.Expression(1))
		y.pushOperator(yakvm.OpIn)
		if op2 := i.NotLiteral(); op2 != nil {
			y.pushOperator(yakvm.OpNot)
		}
	} else if op := i.MakeExpression(); op != nil {
		y.VisitMakeExpression(op)
	} else if op := i.SliceCall(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.VisitSliceCall(op)
	} else if op := i.LogicAnd(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		jmptop := y.pushJmpIfFalseOrPop()
		y.VisitExpression(i.Expression(1))
		jmptop.Unary = y.GetNextCodeIndex()
	} else if op := i.LogicOr(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace(op.GetText())
		jmptop := y.pushJmpIfTrueOrPop()
		y.VisitExpression(i.Expression(1))
		jmptop.Unary = y.GetNextCodeIndex()
	} else if op := i.Question(); op != nil { // can be executed immediately. The ternary condition operator ? :
		// e0 ? e1 : e2
		y.VisitExpression(i.Expression(0))
		y.writeStringWithWhitespace("?")
		jmpf := y.pushJmpIfFalse()
		y.VisitExpression(i.Expression(1))
		y.writeStringWithWhitespace(":")
		jmpEnd := y.pushJmp()
		jmpf.Unary = y.GetNextCodeIndex()
		y.VisitExpression(i.Expression(2))
		jmpEnd.Unary = y.GetNextCodeIndex()
	} else if instanceCode := i.InstanceCode(); instanceCode != nil {
		//to determine whether the current code block can be executed immediately. When it is in the global code block or InstanceCode function,
		inGlobal := false
		if y.currentSymtbl == y.rootSymtbl {
			inGlobal = true
		}
		info := y.contextInfo.Peek()
		if v, ok := info.(string); !ok || v != "InstanceCode" {
			inGlobal = true
		}
		if inGlobal {
			y.contextInfo.Push("InstanceCode")
		}
		// anonymous function, instance code
		y.VisitInstanceCode(instanceCode)
		if inGlobal {
			y.contextInfo.Pop()
		}
	} else if op := i.MemberCall(); op != nil {
		y.VisitExpression(i.Expression(0))
		y.VisitMemberCall(op)
	} else {
		y.panicCompilerError(expressionError, i.GetText())

	}

	return nil
}
