package yakast

import (
	"fmt"

	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"

	uuid "github.com/satori/go.uuid"
)

func (y *YakCompiler) VisitAnonymousFunctionDecl(raw yak.IAnonymousFunctionDeclContext) interface{} {
	if y == nil || raw == nil {
		return nil
	}

	i, _ := raw.(*yak.AnonymousFunctionDeclContext)
	if i == nil {
		return nil
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	var funcName string
	var funcSymbolId int
	if i.FunctionNameDecl() != nil {
		funcName = i.FunctionNameDecl().GetText()
		id, err := y.currentSymtbl.NewSymbolWithReturn(funcName)
		if err != nil {
			y.panicCompilerError(compileError, "cannot create new symbol for function name: "+funcName)
		}
		funcSymbolId = id
	}

	//Functions are divided into closure functions (including arrow functions) and global functions.
	//closure function can be defined anywhere and inherits the parent scope. Global functions can only be defined in the root scope.
	//closure function must use variables to receive or call immediately. The global function scope is global and can be called at any location.
	//closure functions stored in the stack, and global functions stored in global variables.

	// switches the symbol table and code stack.

	recoverCodeStack := y.SwitchCodes()
	recoverSymbolTable := y.SwitchSymbolTable("function", uuid.NewV4().String())
	defer recoverSymbolTable()
	var paramsSymbol []int
	var fun *yakvm.Function
	var isVariable bool
	if i.EqGt() != nil {
		// Processing parameters: Set the symbol table of the domain within the function for the parameter
		if i.LParen() != nil && i.RParen() != nil {
			y.writeString("(")
			paramsSymbol, isVariable = y.VisitFunctionParamDecl(i.FunctionParamDecl())
			y.writeString(")")
			y.writeStringWithWhitespace("=>")
		} else {
			symbolText := i.Identifier().GetText()
			y.writeString(symbolText)
			y.writeStringWithWhitespace("=>")
			symbolId, err := y.currentSymtbl.NewSymbolWithReturn(symbolText)
			if err != nil {
				y.panicCompilerError(compileError, "cannot create identifier["+i.Identifier().GetText()+"] for params (arrow function): "+err.Error())
			}
			paramsSymbol = append(paramsSymbol, symbolId)
		}

		// Arrow function mode
		// Expression and Block need to support
		if i.Block() == nil && i.Expression() == nil {
			y.panicCompilerError(compileError, "BUG: arrow function need expression or block at least")
		}
		if i.Block() != nil {
			y.VisitBlock(i.Block(), true)
			y.pushOperator(yakvm.OpReturn)
		} else {
			// Generally speaking, the stack here is not flat, but because this is inside a function call, the last stack data should be used as the function return value, so there is no need to process it here. In other cases,
			// Implicitly, this is equivalent to () => {return 123;} means () =>123 and ( )=>{return 123}is equivalent. The stack is not flat. At the end of the function,
			// should pop the stack data once and return it. If not, return undefined.
			y.VisitExpression(i.Expression())
			y.pushOperator(yakvm.OpReturn)
		}

		// The compiled FuncCode is matched with the symbol table. Generally speaking, it can be executed and called.
		fun = yakvm.NewFunction(y.codes, y.currentSymtbl)
		if y.sourceCodePointer != nil {
			fun.SetSourceCode(*y.sourceCodePointer)
		}
	} else {
		// to create symbols
		if fn := i.Func(); fn != nil {
			y.writeString(fn.GetText())
		}
		if funcName != "" {
			y.writeString(" ")
			y.writeString(funcName)
		}
		y.writeString("(")
		paramsSymbol, isVariable = y.VisitFunctionParamDecl(i.FunctionParamDecl())
		y.writeString(") ")
		// visit code block.
		y.VisitBlock(i.Block(), true)
		y.pushOperator(yakvm.OpReturn)
		funcCode := y.codes
		// The compiled FuncCode is matched with the symbol table. Generally speaking, it can be executed and called.
		fun = yakvm.NewFunction(funcCode, y.currentSymtbl)
		if y.sourceCodePointer != nil {
			fun.SetSourceCode(*y.sourceCodePointer)
		}
	}
	if funcName != "" {
		// If the function name exists, set the function name and create a new one. symbol, and tell the function the new symbol for subsequent processing
		fun.SetName(funcName)
		fun.SetSymbol(funcSymbolId)
	}

	//restores the scene.
	recoverCodeStack()

	if fun == nil {
		y.panicCompilerError(compileError, "cannot create yak function from compiler")
	}
	fun.SetParamSymbols(paramsSymbol)
	fun.SetIsVariableParameter(isVariable)
	funcVal := &yakvm.Value{
		TypeVerbose: "anonymous-function",
		Value:       fun,
	}
	if funcName != "" {
		// If there is a function name, perform quick assignment
		funcVal.TypeVerbose = "named-function"
		y.pushLeftRef(fun.GetSymbolId())
		y.pushValue(funcVal)
		y.pushOperator(yakvm.OpFastAssign)
	} else {
		// closure function, push directly to the stack.
		y.pushValueWithCopy(funcVal)
	}

	return nil
}

func (y *YakCompiler) VisitFunctionParamDecl(raw yak.IFunctionParamDeclContext) ([]int, bool) {
	if y == nil || raw == nil {
		return nil, false
	}

	i, _ := raw.(*yak.FunctionParamDeclContext)
	if i == nil {
		return nil, false
	}
	recoverRange := y.SetRange(i.BaseParserRuleContext)
	defer recoverRange()

	ellipsis := i.Ellipsis()
	ids := i.AllIdentifier()
	lenOfIds := len(ids)
	symbols := make([]int, lenOfIds)

	tokenStart := i.BaseParserRuleContext.GetStart().GetColumn()
	lineLength := tokenStart
	eachParamOneLine := false
	identifierTokenLengths := make([]int, lenOfIds)
	for index, id := range ids {
		identifierTokenLengths[index] = len(id.GetText())
		if !eachParamOneLine && identifierTokenLengths[index] > FORMATTER_RECOMMEND_PARAM_LENGTH {
			eachParamOneLine = true
		}
	}

	if lenOfIds == 1 && eachParamOneLine {
		eachParamOneLine = false
	}

	hadIncIndent := false
	comments := getIdentifersSurroundComments(i.GetParser().GetTokenStream(), i.GetStart(), i.GetStop(), lenOfIds)

	for index, id := range ids {
		idText := id.GetText()
		lineLength += identifierTokenLengths[index]

		if lenOfIds > 1 { // . If there is not only one parameter, the maximum length of a single line is exceeded. If the length or any parameter is too long, wrap it in a new line.
			if eachParamOneLine {
				y.writeNewLine()
				if !hadIncIndent {
					y.incIndent()
					hadIncIndent = true
				}
				y.writeIndent()
				lineLength = y.indent*4 + identifierTokenLengths[index]
			} else if lineLength > FORMATTER_MAXWIDTH {
				y.writeNewLine()
				y.writeWhiteSpace(tokenStart)
				lineLength = tokenStart + identifierTokenLengths[index]
			}
		}

		symbolId, err := y.currentSymtbl.NewSymbolWithReturn(idText)
		if err != nil {
			y.panicCompilerError(compileError, "cannot create symbol for function params decl")
		}
		symbols[index] = symbolId

		y.writeString(idText)

		if comments[index] != "" {
			y.writeString(fmt.Sprintf(" /* %s */", comments[index]))
		}

		// If it is the last parameter and there is..., you need to add...
		if index == lenOfIds-1 {
			if ellipsis != nil {
				y.writeString("...")
			}
		}
		// If it is not the last parameter or each parameter is in one line, it must be added.
		if index != lenOfIds-1 || eachParamOneLine {
			y.writeString(", ")
			lineLength += 2
		}
		// If it is the last parameter and each parameter is in one line, it must be changed to
		if index == lenOfIds-1 && eachParamOneLine {
			y.writeNewLine()
			if hadIncIndent {
				y.decIndent()
			}
			y.writeIndent()
		}
	}

	return symbols, ellipsis != nil
}
