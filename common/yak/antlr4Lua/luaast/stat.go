package luaast

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/yaklang/yaklang/common/log"
	lua "github.com/yaklang/yaklang/common/yak/antlr4Lua/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
	"strings"
)

func (l *LuaTranslator) VisitStat(raw lua.IStatContext) interface{} {
	if l == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*lua.StatContext)

	if i == nil {
		return nil
	}

	if s := i.SemiColon(); s != nil {
		return nil
	}

	if s := i.Varlist(); s != nil {
		if t := i.Explist(); t != nil {
			l.VisitExpList(t)
		}
		l.VisitVarList(true, s)
		l.pushGlobalAssign()
		return nil
	}

	if s := i.Functioncall(); s != nil {
		l.VisitFunctionCall(s)
		l.pushOpPop()
		return nil
	}

	if s := i.Label(); s != nil {
		l.VisitLabel(s)
		return nil
	}

	// The break statement terminates the execution of a `while`, `repeat`, or `for` loop, skipping to the next statement after the loop
	if s := i.Break(); s != nil {
		if !(l.NowInWhile() || l.NowInRepeat() || l.NowInFor()) {
			log.Warnf("Syntax Error: Break should be in `while`, `repeat`, or `for` loop, it will crash in future")
		}

		l.pushBreak()
		return nil
	}

	// The goto statement transfers the program control to a label. For syntactical reasons, labels in Lua are considered statements too
	if s := i.Goto(); s != nil {
		l.VisitGoto(i.NAME().GetText())
		return nil
	}

	if s, w, f := i.Do(), i.While(), i.For(); s != nil && w == nil && f == nil {
		l.VisitBlock(i.Block(0))
		return nil
	}

	// no `continue` in lua
	if s := i.While(); s != nil {
		/*
			while-scope
			exp
			jmp-to-end-if-false
			block-scope
			block-scope-end
			jmp-to-exp
			while-scope-end

			When there is a break,
			while-scope
			exp
			jmp-to-end-if-false
			block-scope
			break
			block-scope-end
			jmp-to-exp
			while-scope-end
		*/
		var toEnds []*yakvm.Code

		startIndex := l.GetNextCodeIndex()
		l.enterWhileContext(startIndex)

		f := l.SwitchSymbolTableInNewScope("while-loop", uuid.NewV4().String())

		exp := i.Exp(0)
		l.VisitExp(exp)
		toEnds = append(toEnds, l.pushJmpIfFalse())

		if i.Do() != nil {
			l.VisitBlock(i.Block(0))
		}

		l.pushJmp().Unary = startIndex + 1 // avoid jmp to new-scope
		var whileEnd = l.GetNextCodeIndex()

		f()
		// of a slice that has not been set in the parsed block. break
		l.exitWhileContext(whileEnd + 1)

		// sets the toEnd position
		for _, toEnd := range toEnds {
			if toEnd != nil {
				toEnd.Unary = whileEnd
			}
		}

		return nil
	}
	//. In the repeat–until loop, the inner block does not end at the until keyword,
	//but only after the condition. So, the condition can refer to local variables
	//declared inside the loop block
	// . Consider using syntactic sugar to implement
	if s := i.Repeat(); s != nil {
		var toEnds []*yakvm.Code

		startIndex := l.GetNextCodeIndex()
		l.enterWhileContext(startIndex)

		f := l.SwitchSymbolTableInNewScope("repeat-loop", uuid.NewV4().String())

		l.VisitBlock(i.Block(0))
		// . Put the exp condition into the block-scope. In this way, the local life in the block can still be used when making conditional judgment until until.
		var endScope *yakvm.Code
		endScope, l.codes = l.codes[len(l.codes)-1], l.codes[:len(l.codes)-1]
		exp := i.Exp(0)
		l.VisitExp(exp)
		l.codes = append(l.codes, endScope)

		toEnds = append(toEnds, l.pushJmpIfTrue()) // . If the condition of repeat is not true, it will always loop like for-while
		l.pushJmp().Unary = startIndex + 1         // to avoid opening the scope repeatedly.

		var untilEnd = l.GetNextCodeIndex()

		f()
		// of a slice that has not been set in the parsed block. break
		l.exitWhileContext(untilEnd + 1)

		// sets the toEnd position
		for _, toEnd := range toEnds {
			if toEnd != nil {
				toEnd.Unary = untilEnd
			}
		}

		return nil
	}

	if s := i.If(); s != nil {
		conditionExprCnt, blockCnt := 0, 0
		l.VisitExp(i.Exp(conditionExprCnt))
		var jmpfCode = l.pushJmpIfFalse() // . If the condition is not true, jump to the else branch
		l.VisitBlock(i.Block(blockCnt))
		var jmp = l.pushJmp() // If the condition is true, jump to the next statement of the entire if-else.
		elseIndex := l.GetNextCodeIndex()
		jmpfCode.Unary = elseIndex
		for range i.AllElseIf() {
			conditionExprCnt++
			blockCnt++
			l.VisitExp(i.Exp(conditionExprCnt))
			jmpfCode := l.pushJmpIfFalse()
			l.VisitBlock(i.Block(blockCnt))
			l.codes = append(l.codes, jmp)
			elseIndex := l.GetNextCodeIndex()
			jmpfCode.Unary = elseIndex

		}
		if i.Else() != nil {
			blockCnt++
			l.VisitBlock(i.Block(blockCnt))
		}
		jmp.Unary = l.GetNextCodeIndex()
		return nil
	}

	if s := i.For(); s != nil {
		if i.Namelist() == nil && i.NAME() != nil {
			f := l.SwitchSymbolTableInNewScope("for-numerical", uuid.NewV4().String())
			defer f()
			iterateVarName := i.NAME().GetText()
			// first. Assign var to
			iterateVarID := l.currentSymtbl.NewSymbolWithoutName()
			l.pushLeftRef(iterateVarID)
			l.VisitExp(i.Exp(0))
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()
			// only calculates the condition once.
			conditionId := l.currentSymtbl.NewSymbolWithoutName()
			l.pushLeftRef(conditionId)
			l.VisitExp(i.Exp(1))
			// for the following We can judge whether to execute the third statement based on the condition. We need to cache the result into the intermediate symbol
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()

			var stepExp lua.IExpContext
			var stepId int
			if i.Exp(2) != nil {
				stepExp = i.Exp(2)
				stepId = l.currentSymtbl.NewSymbolWithoutName()
				l.pushLeftRef(stepId)
				l.VisitExp(stepExp)
				l.pushOperator(yakvm.OpFastAssign)
				l.pushOpPop()
			}
			// after the execution body ends, it should jump back unconditionally At the beginning, re-judge
			// on the left. However, the third statement for ;; . It should be a block. After executing the explanation, execute the third statement
			l.pushLeftRef(iterateVarID)
			if stepExp != nil { // step
				l.pushRef(iterateVarID)
				l.pushRef(stepId)
				l.pushOperator(yakvm.OpSub)
			} else {
				l.pushRef(iterateVarID)
				l.pushInteger(1, "1")
				l.pushOperator(yakvm.OpSub)
			}
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()

			innerWhile := l.SwitchSymbolTableInNewScope("for-numerical-while-inner", uuid.NewV4().String())
			innerStartIndex := l.GetNextCodeIndex()
			l.enterWhileContext(innerStartIndex)

			l.pushLeftRef(iterateVarID)
			if stepExp != nil { // step
				l.pushRef(iterateVarID)
				l.pushRef(stepId)
				l.pushOperator(yakvm.OpAdd)
			} else {
				l.pushRef(iterateVarID)
				l.pushInteger(1, "1")
				l.pushOperator(yakvm.OpAdd)
			}
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()

			var lastAnd, lastAnd1 *yakvm.Code
			if stepExp != nil {
				l.pushRef(stepId)
				l.pushInteger(0, "0")
				l.pushOperator(yakvm.OpGtEq)

				jmptop1 := l.pushJmpIfFalse()

				l.pushRef(iterateVarID)
				l.pushRef(conditionId)
				l.pushOperator(yakvm.OpGt)

				jmpOr := l.pushJmpIfTrue()
				jmptop1.Unary = l.GetNextCodeIndex()

				l.pushRef(stepId)
				l.pushInteger(0, "0")
				l.pushOperator(yakvm.OpLt)

				lastAnd = l.pushJmpIfFalse()

				l.pushRef(iterateVarID)
				l.pushRef(conditionId)
				l.pushOperator(yakvm.OpLt)

				lastAnd1 = l.pushJmpIfFalse()

				jmpOr.Unary = l.GetNextCodeIndex()
				l.pushBreak()
			} else { // default step is 1
				l.pushInteger(1, "1")
				l.pushInteger(0, "0")
				l.pushOperator(yakvm.OpGtEq)

				jmptop1 := l.pushJmpIfFalse()

				l.pushRef(iterateVarID)
				l.pushRef(conditionId)
				l.pushOperator(yakvm.OpGt)

				jmpOr := l.pushJmpIfTrue()
				jmptop1.Unary = l.GetNextCodeIndex()

				l.pushInteger(1, "1")
				l.pushInteger(0, "0")
				l.pushOperator(yakvm.OpLt)

				lastAnd = l.pushJmpIfFalse()

				l.pushRef(iterateVarID)
				l.pushRef(conditionId)
				l.pushOperator(yakvm.OpLt)

				lastAnd1 = l.pushJmpIfFalse()

				jmpOr.Unary = l.GetNextCodeIndex()
				l.pushBreak()

			}

			lastAnd.Unary = l.GetNextCodeIndex()
			lastAnd1.Unary = l.GetNextCodeIndex()

			fakeIterateVarID, err := l.currentSymtbl.NewSymbolWithReturn(iterateVarName)
			if err != nil {
				l.panicCompilerError(autoCreateSymbolFailed, iterateVarName)
			}
			l.pushLeftRef(fakeIterateVarID) // . Inject the fake variable
			l.pushRef(iterateVarID)
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()

			l.VisitBlock(i.Block(0))
			l.pushJmp().Unary = innerStartIndex

			var innerWhileEnd = l.GetNextCodeIndex()

			innerWhile()
			l.exitWhileContext(innerWhileEnd)

			return nil
		}
		if i.Namelist() != nil && i.Explist() != nil {
			nameList := strings.Split(i.Namelist().GetText(), ",")
			var nameRef []int
			recoverSymtbl := l.SwitchSymbolTableInNewScope("for-iterate", uuid.NewV4().String())
			defer recoverSymtbl()

			l.VisitExpList(i.Explist())
			iterFuncID := l.currentSymtbl.NewSymbolWithoutName()
			internalStateID := l.currentSymtbl.NewSymbolWithoutName()
			initialValueID := l.currentSymtbl.NewSymbolWithoutName()

			l.pushLeftRef(iterFuncID)
			l.pushLeftRef(internalStateID)
			l.pushLeftRef(initialValueID)
			l.pushListWithLen(3)
			l.pushLocalAssign()

			for _, name := range nameList {
				sym, err := l.currentSymtbl.NewSymbolWithReturn(name)
				if err != nil {
					l.panicCompilerError(constError(fmt.Sprintf("cannot create `%v` variable in generic for reason: ", name)), err.Error())
				}
				nameRef = append(nameRef, sym)
			}

			innerWhile := l.SwitchSymbolTableInNewScope("for-iterate-inner", uuid.NewV4().String())

			innerStartIndex := l.GetNextCodeIndex()
			l.enterWhileContext(innerStartIndex)

			l.pushRef(iterFuncID)
			l.pushRef(internalStateID)
			l.pushRef(initialValueID)
			l.pushCall(2)

			l.pushListWithLen(1)

			for _, ref := range nameRef {
				l.pushLeftRef(ref)
			}

			l.pushListWithLen(len(nameRef))
			l.pushLocalAssign()

			l.pushLeftRef(initialValueID)
			l.pushRef(nameRef[0])
			l.pushOperator(yakvm.OpFastAssign)
			l.pushOpPop()

			l.pushRef(initialValueID)
			l.pushUndefined()
			l.pushOperator(yakvm.OpEq)

			jmp := l.pushJmpIfFalse()
			l.pushBreak()
			jmpIndex := l.GetNextCodeIndex()
			jmp.Unary = jmpIndex

			l.VisitBlock(i.Block(0))
			l.pushJmp().Unary = innerStartIndex

			var innerWhileEnd = l.GetNextCodeIndex()
			innerWhile()
			l.exitWhileContext(innerWhileEnd)

			return nil
		}
	}

	if s, w := i.Function(), i.Local(); s != nil && w == nil {
		l.VisitFuncNameAndBody(i.Funcname(), i.Funcbody())
		return nil
	}

	if s := i.Local(); s != nil {
		if i.Function() != nil {
			// NAME here is necessary no anonymous function allowed
			l.VisitLocalFuncNameAndBody(i.NAME().GetText(), i.Funcbody())
			return nil
		} else {
			list := i.Attnamelist().(*lua.AttnamelistContext)
			nameList := list.AllNAME()
			// fixed: Ignore this first. attrib attributeList := list.AllAttrib()
			if expList := i.Explist(); expList != nil {
				l.VisitExpList(expList)
			} else { // only declares no value.
				for range nameList {
					l.pushUndefined()
				}
				l.pushListWithLen(len(nameList))
			}

			for index, varName := range nameList {
				l.VisitLocalVarWithName(true, varName.GetText(), list.Attrib(index))
			}
			l.pushListWithLen(len(nameList))
			l.pushLocalAssign()
			return nil
		}
	}

	return nil
}

func (l *LuaTranslator) VisitLastStat(raw lua.ILaststatContext) interface{} {
	if l == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*lua.LaststatContext)
	if i == nil {
		return nil
	}
	if i.Return() != nil {
		if expList := i.Explist(); expList != nil {
			l.VisitExpList(expList)
		}
		l.pushOperator(yakvm.OpReturn)
	}
	if i.Continue() != nil {
		// TODO: This continues As a last stat situation, I have never encountered Lua. Logically speaking, there is no continue keyword. Put
		panic("TODO")
	}
	if i.Break() != nil {
		if !(l.NowInWhile() || l.NowInRepeat() || l.NowInFor()) {
			log.Warnf("Syntax Error: Break should be in `while`, `repeat`, or `for` loop, it will crash in future")
		}

		l.pushBreak()
		return nil
	}
	return nil
}
