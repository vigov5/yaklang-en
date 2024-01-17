package yakast

import (
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
)

func (y *YakCompiler) VisitFunctionCall(raw yak.IFunctionCallContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.FunctionCallContext)
	if i == nil {
		return nil
	}

	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()
	y.writeString("(")

	// function calls need to push the parameters onto the stack first. When calling
	// , call n indicates how many numbers to take out.
	var argCount = 0
	if i.OrdinaryArguments() != nil {
		argCount, _ = y.VisitOrdinaryArguments(i.OrdinaryArguments())
	}
	y.writeString(")")

	if i.Wavy() != nil {
		y.writeString("~")
		y.pushCallWithWavy(argCount)
	} else {
		y.pushCall(argCount)
	}
	return nil
}

func (y *YakCompiler) VisitOrdinaryArguments(raw yak.IOrdinaryArgumentsContext) (int, bool) {
	if y == nil || raw == nil {
		return 0, false
	}

	i, _ := raw.(*yak.OrdinaryArgumentsContext)
	if i == nil {
		return 0, false
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	ellipsis := i.Ellipsis()
	allExpressions := i.AllExpression()
	lenOfAllExpressions := len(allExpressions)

	expressionTokenLengths := make([]int, lenOfAllExpressions)
	tokenStart := i.BaseParserRuleContext.GetStart().GetColumn()
	lineLength := tokenStart
	eachParamOneLine := false
	// First traverse once and calculate the length of each expression. If it is too long, you need to wrap it.
	for i, e := range allExpressions {
		expressionTokenLengths[i] = len(e.GetText())
		if !eachParamOneLine && expressionTokenLengths[i] > FORMATTER_RECOMMEND_PARAM_LENGTH {
			eachParamOneLine = true
		}
	}
	if lenOfAllExpressions == 1 && eachParamOneLine {
		eachParamOneLine = false
	}

	hadIncIndent := false

	for index, expr := range allExpressions {
		lineLength += expressionTokenLengths[index]

		if lenOfAllExpressions > 1 { // . If there is not only one parameter, the maximum length of a single line is exceeded. If the length or any parameter is too long, wrap it in a new line.
			if eachParamOneLine {
				y.writeNewLine()
				if !hadIncIndent {
					y.incIndent()
					hadIncIndent = true
				}
				y.writeIndent()
				lineLength = y.indent*4 + expressionTokenLengths[index]
			} else if lineLength > FORMATTER_MAXWIDTH {
				y.writeNewLine()
				y.writeWhiteSpace(tokenStart)
				lineLength = tokenStart + expressionTokenLengths[index]
			}
		}

		y.VisitExpression(expr)

		// If it is the last parameter and there is..., you need to add...
		if index == lenOfAllExpressions-1 {
			if ellipsis != nil {
				y.pushEllipsis(lenOfAllExpressions)
				y.writeString("...")
			}
		}
		// If it is not the last parameter or each parameter is in one line, it must be added.
		if index != lenOfAllExpressions-1 || eachParamOneLine {
			y.writeString(", ")
			lineLength += 2
		}
		// If it is the last parameter and each parameter is in one line, it must be changed to
		if index == lenOfAllExpressions-1 && eachParamOneLine {
			y.writeNewLine()
			if hadIncIndent {
				y.decIndent()
			}
			y.writeIndent()
		}
	}

	return len(i.AllExpression()), ellipsis != nil
}
