package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
)

func (y *YakCompiler) VisitSliceCall(raw yak.ISliceCallContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.SliceCallContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	//Check the number of parameters
	exps := i.AllExpression()
	if len(exps) == 0 {
		y.panicCompilerError(sliceCallNoParamError)
	}
	if len(exps) > 3 {
		y.panicCompilerError(sliceCallTooManyParamError)
	}
	y.writeString("[")
	defer y.writeString("]")
	//Solution: When one side is empty,
	childrens := i.GetChildren()
	expect := true // records the status. If the expectation is a number and the result is:, then push a default number and do not switch the status.
	t := 0         // Record the number of parameters
	idEnd := false
	visitChildrens := childrens[1:]
	lenOfVisitChildrens := len(visitChildrens)
	for index, children := range visitChildrens {
		if expect {
			expression, isExpression := children.(*yak.ExpressionContext)

			if isExpression {
				// The expression value type must be int, and the step value cannot be 0, otherwise an error will be reported
				//if t == 2 &&  != nil {
				//	panic(" step cannot be zero")
				//}
				y.VisitExpression(expression)
				expect = !expect
			} else {
				if t == 1 {
					idEnd = true
				}
				if t == 2 {
					y.panicCompilerError(sliceCallStepMustBeNumberError)
				}

				if index != lenOfVisitChildrens-1 {
					y.writeString(":")
				}
				y.pushInteger(0, "0")
			}

			t += 1
		} else { // If the expectation is:, then switch the state directly
			if index != lenOfVisitChildrens-1 {
				y.writeString(":")
			}
			expect = !expect
		}
	}
	y.pushBool(idEnd)
	y.pushIterableCall(t)
	return nil
}
