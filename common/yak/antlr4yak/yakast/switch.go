package yakast

import (
	"fmt"
	"strings"

	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

	uuid "github.com/satori/go.uuid"
)

type switchContext struct {
	startCodeIndex          int
	switchBreakScopeCounter int
}

func (y *YakCompiler) enterSwitchContext(start int) {
	y.switchDepthStack.Push(&switchContext{
		startCodeIndex: start,
	})
}

func (y *YakCompiler) peekSwitchContext() *switchContext {
	raw, ok := y.switchDepthStack.Peek().(*switchContext)
	if ok {
		return raw
	} else {
		return nil
	}
}

func (y *YakCompiler) exitSwitchContext(end int) {
	start := y.peekSwitchStartIndex()
	if start <= 0 {
		return
	}

	for _, c := range y.codes[start:] {
		if c.Opcode == yakvm.OpBreak && c.Unary <= 0 {
			// Set The jump value of the Break Code of all statements from the beginning to the end of for
			c.Unary = end
		}
	}

	y.switchDepthStack.Pop()
}

func (y *YakCompiler) _VisitSwitchStmt(raw yak.ISwitchStmtContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.SwitchStmtContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("switch ")

	var (
		defaultCodeIndex  int
		switchExprIsEmpty bool
	)

	recoverSymtbl := y.SwitchSymbolTableInNewScope("switch", uuid.NewV4().String())
	defer recoverSymtbl()

	startIndex := y.GetNextCodeIndex()
	y.enterSwitchContext(startIndex)

	symbolName := fmt.Sprintf("$switch:%v$", y.GetNextCodeIndex())
	expressionResultID, err := y.currentSymtbl.NewSymbolWithReturn(symbolName)
	if err != nil {
		y.panicCompilerError(CreateSymbolError, symbolName)
	}

	// Put the value of the expression into the symbol
	if e := i.Expression(); e != nil {
		y.VisitExpression(i.Expression())
		y.pushListWithLen(1)
		// Set the lvalue, this lvalue is a newly created symbol!
		y.pushLeftRef(expressionResultID)
		y.pushListWithLen(1)
		// Create a symbol for the rvalue, this symbol is rightExpressionSymbol
		y.pushOperator(yakvm.OpAssign)
		y.writeString(" {")
	} else {
		//y.pushUndefined()
		switchExprIsEmpty = true
		y.writeString("{")
	}

	y.writeNewLine()

	allcases := i.AllCase()
	lenOfAllCases := len(allcases)
	jmpToEnd := make([]*yakvm.Code, 0, lenOfAllCases)
	jmpToNextCase := make([]*yakvm.Code, 0, lenOfAllCases)
	jmpFallthrough := make([]*yakvm.Code, 0)
	nextCaseIndexs := make([]int, 0, lenOfAllCases)
	nextStmtIndexs := make([]int, 0, lenOfAllCases)

	for index := range allcases {
		recoverSymtbl = y.SwitchSymbolTableInNewScope("case", uuid.NewV4().String())
		y.writeString("case ")

		var jmpToStmt []*yakvm.Code
		// Get the position of the next case
		caseIndex := y.GetNextCodeIndex()
		nextCaseIndexs = append(nextCaseIndexs, caseIndex)

		// if judge
		iExprs := i.ExpressionList(index)
		if iExprs != nil {
			exprs := iExprs.(*yak.ExpressionListContext)
			lenOfExprs := len(exprs.AllExpression())
			// If there is only one expression, directly use eq to judge
			if lenOfExprs == 1 {
				y.VisitExpression(exprs.AllExpression()[0])
				if !switchExprIsEmpty {
					y.pushRef(expressionResultID)
					y.pushOperator(yakvm.OpEq)
				}
			} else { // If there are multiple expressions, short-circuit processing
				for i, e := range exprs.AllExpression() {
					y.VisitExpression(e)
					if !switchExprIsEmpty {
						y.pushRef(expressionResultID)
						y.pushOperator(yakvm.OpEq)
					}
					jmpToStmt = append(jmpToStmt, y.pushJmpIfTrue())
					if i < lenOfExprs-1 {
						y.writeString(", ")
					}
				}
				// Finally add a false, which is used for the following jmpToNextCase condition judgment
				y.pushBool(false)
			}
		}
		y.writeString(":")
		y.writeNewLine()
		y.incIndent()

		// If not equal, jump to the next case
		jmpToNextCase = append(jmpToNextCase, y.pushJmpIfFalse())

		// Get the position of the next statementlist
		stmtIndex := y.GetNextCodeIndex()
		nextStmtIndexs = append(nextStmtIndexs, stmtIndex)

		// Set the condition to short-circuit
		for _, jmp := range jmpToStmt {
			jmp.Unary = stmtIndex
		}

		// Execute the statement in the case. Because there is a fallthrough, you need to obtain the context, so you cannot Use VisitStatementList and VisitStatement directly
		recoverFormatBufferFunc := y.switchFormatBuffer()
		iStmts := i.StatementList(index)
		if iStmts != nil {
			stmts := iStmts.(*yak.StatementListContext)
			allStatement := stmts.AllStatement()
			lenOfAllStatement := len(allStatement)
			for i, istmt := range allStatement {
				if istmt == nil {
					continue
				}
				stmt := istmt.(*yak.StatementContext)
				// Ignore the empty
				if i == 0 && stmt.Empty() != nil {
					continue
				}

				y.writeIndent()

				if s := stmt.FallthroughStmt(); s != nil {
					if y.NowInSwitch() {
						y.writeString("fallthrough")
						y.writeEOS(stmt.Eos())
						jmp := y.pushJmp()
						// Temporarily set to index, it will be set to jump to the next later The position of a statementlist
						jmp.Unary = index
						jmpFallthrough = append(jmpFallthrough, jmp)
						continue
					}
					y.panicCompilerError(fallthroughError)
				} else {
				}

				newline := y.VisitStatement(istmt.(*yak.StatementContext))
				if i < lenOfAllStatement-1 && newline {
					y.writeNewLine()
				}
			}
		}
		buf := recoverFormatBufferFunc()
		buf = strings.Trim(buf, "\n")
		y.writeString(buf)
		y.decIndent()
		y.writeNewLine()

		// Jump to the end of switch
		jmpToEnd = append(jmpToEnd, y.pushJmp())
		recoverSymtbl()
	}

	// Access the default statementlist
	if i.Default() != nil {
		recoverSymtbl = y.SwitchSymbolTableInNewScope("default", uuid.NewV4().String())
		y.writeString("default:")
		y.writeNewLine()
		y.incIndent()

		defaultCodeIndex = y.GetNextCodeIndex()

		recoverFormatBufferFunc := y.switchFormatBuffer()
		stmts := i.StatementList(len(allcases)).(*yak.StatementListContext)
		// y.VisitStatementList(stmts)
		allStatement := stmts.AllStatement()
		lenOfAllStatement := len(allStatement)
		for i, istmt := range allStatement {
			if istmt == nil {
				continue
			}
			stmt := istmt.(*yak.StatementContext)
			// Ignore the empty
			if i == 0 && stmt.Empty() != nil {
				continue
			}

			y.writeIndent()
			newline := y.VisitStatement(istmt.(*yak.StatementContext))
			if i < lenOfAllStatement-1 && newline {
				y.writeNewLine()
			}
		}

		buf := recoverFormatBufferFunc()
		buf = strings.Trim(buf, "\n")
		y.writeString(buf)
		y.decIndent()
		y.writeNewLine()
		recoverSymtbl()
	}

	endCodewithScopeEnd := y.GetNextCodeIndex()
	// at the beginning If there is no default, jump to the end of switch
	if defaultCodeIndex == 0 {
		defaultCodeIndex = endCodewithScopeEnd
	}

	endCode := y.GetNextCodeIndex()

	// Set fallthrough to jump to the position of the next statementlist

	for _, jmp := range jmpFallthrough {
		// The last fallthrough should jump to default
		if jmp.Unary == lenOfAllCases-1 {
			jmp.Unary = defaultCodeIndex
		} else {
			jmp.Unary = nextStmtIndexs[jmp.Unary+1]
		}
	}

	// Sets the jump Go to the position of the next case
	for index, jmp := range jmpToNextCase[:len(jmpToNextCase)-1] {
		jmp.Unary = nextCaseIndexs[index+1]
	}

	// Set the last case to jump to default
	jmpToNextCase[len(jmpToNextCase)-1].Unary = defaultCodeIndex

	// Sets the jump to the position at the end of switch
	for _, jmp := range jmpToEnd {
		jmp.Unary = endCodewithScopeEnd
	}
	// Sets break to jump to the position at the end of switch. There is no need to process the internal scopeEnd, because break comes with scopeEnd
	// Because when the switch is created, the level number of breakScope starts from 0, so break will only exit the scope inside the switch-scope. In the end, a scopeEnd of the switch-scope is still needed.
	// endCode must be the last ScopeEnd
	y.exitSwitchContext(endCode)

	y.writeString("}")

	return nil
}
