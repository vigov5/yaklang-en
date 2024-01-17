package yakast

type constError string
type CompilerLanguage string

const (
	zh CompilerLanguage = "zh"
	en CompilerLanguage = "en"
)

const (
	compileError                              = "compile error: %v"
	breakError                     constError = "break statement can only be used in for or switch"
	continueError                             = "continue statement can only be used in for"
	fallthroughError                          = "fallthrough statement can only be used in switch"
	sliceCallNoParamError                     = "at least one param for slice call"
	sliceCallTooManyParamError                = "too many params for slice call"
	sliceCallStepMustBeNumberError            = "step must be a number"
	CreateSymbolError                         = "SymbolTable cannot create build-in symbol[%s]"
	assertExpressionError                     = "assert statement second argument expect expression"
	bitBinaryError                            = "BUG: unimplemented bit binary operator: %s"
	multiplicativeBinaryError                 = "BUG: unimplemented multiplicative binary operator: %s"
	expressionError                           = "BUG: cannot parse `%s` as expression"
	includeUnquoteError                       = "include path[%s] unquote error: %v"
	includePathNotFoundError                  = "include path[%s] not found"
	includeCycleError                         = "include cycle not allowed: %s"
	readFileError                             = "read file[%s] read error: %v"
	stringLiteralError                        = "invalid string literal: %s"
	notImplemented                            = "[%s] not implemented"
	forceCreateSymbolFailed                   = "BUG: cannot force create symbol for `%s`"
	autoCreateSymbolFailed                    = "BUG: cannot auto create symbol for `%s`"
	integerIsTooLarge                         = "cannot parse `%s` as integer literal... is too large for int64"
	contParseNumber                           = "cannot parse num for literal: %s"
	notFoundDollarVariable                    = "undefined dollor variable: $%v"
	bugMembercall                             = "BUG: no identifier or $identifier to call via member"
	notFoundVariable                          = "(strict mode) undefined variable: %v"
	syntaxUnrecoverableError                  = "grammar parser error: cannot continue to parse syntax (unrecoverable), maybe there are unbalanced brace"
)

var i18n = map[CompilerLanguage]map[constError]string{
	zh: {
		compileError:               "Compilation error: %v",
		breakError:                 "The break statement can only be used in for or switch",
		continueError:              "The continue statement can only be used in for",
		fallthroughError:           "The fallthrough statement can only be used in switch",
		sliceCallNoParamError:      "slice exists The operation requires at least one parameter",
		sliceCallTooManyParamError: "is not implemented. There are too many slicing operation parameters.",
		assertExpressionError:      "called through member The second argument of assert statement must be an expression",
		bitBinaryError:             "BUG: Unimplemented binary bit operator: %s",
		multiplicativeBinaryError:  "BUG : Unimplemented binary operator: %s",
		expressionError:            "BUG: Unable to parse `%s` into an expression",
		includeUnquoteError:        "Contains path [%s] Parsing error: %v",
		includePathNotFoundError:   "contains path [%s] No",
		includeCycleError:          "Loops are not allowed to include files: %s",
		readFileError:              "Reading file [%s] Error: %v",
		stringLiteralError:         "Illegal string literal : %s",
		notImplemented:             "[%s]",
		forceCreateSymbolFailed:    "BUG: Unable to force creation of symbol `%s `",
		autoCreateSymbolFailed:     "BUG: Unable to automatically create symbol `%s`",
		integerIsTooLarge:          "Unable to parse `%s` as an integer because it is too large for int64",
		contParseNumber:            "Unable to parse numeric literal: %s",
		notFoundDollarVariable:     "Undefined $ Variable: $%v",
		bugMembercall:              "BUG: No identifier or$Identifier",
		notFoundVariable:           "(strict mode) Undefined variable: %v",
		syntaxUnrecoverableError:   "Syntax parser error: A syntax error here caused the parser to break (unable to continue), possibly due to unbalanced parentheses",
	},
}

func (y *YakCompiler) GetConstError(e constError) string {
	if y.language == en {
		return string(e)
	}
	if constsInfo, ok := i18n[y.language]; ok {
		if msg, ok := constsInfo[e]; ok {
			return msg
		} else {
			return string(e)
		}
	} else {
		panic("not support language")
	}
	return ""
}
