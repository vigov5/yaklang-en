package luaast

type constError string
type CompilerLanguage string

const (
	zh CompilerLanguage = "zh"
	en CompilerLanguage = "en"
)

const (
	breakError                     constError = "break statement can only be used in for or switch"
	continueError                             = "continue statement can only be used in for"
	fallthroughError                          = "fallthrough statement can only be used in switch"
	sliceCallNoParamError                     = "at least one param for slice call"
	sliceCallTooManyParamError                = "too many params for slice call"
	sliceCallStepMustBeNumberError            = "step must be a number"
	CreateSymbolError                         = "SymbolTable cannot create build-in symbol[%s]"
	assertExpressionError                     = "assert statement second argument expect expression"
	notImplemented                            = "[%s] not implemented"
	forceCreateSymbolFailed                   = "BUG: cannot force create symbol for `%s`"
	autoCreateSymbolFailed                    = "BUG: cannot auto create symbol for `%s`"
	autoCreateLabelFailed                     = "BUG: cannot auto create label for `%s`"
	labelAlreadyDefined                       = "label '%s' already defined"
	labelNotDefined                           = "no visible label '%s' for <goto>"
	integerIsTooLarge                         = "cannot parse `%s` as integer literal... is too large for int64"
	contParseNumber                           = "cannot parse num for literal: %s"
)

var i18n = map[CompilerLanguage]map[constError]string{
	zh: {
		breakError:                 "break statement can only be used in for or switch",
		continueError:              "continue statement can only be used in for",
		fallthroughError:           "The fallthrough statement can only be used in switch",
		sliceCallNoParamError:      "The slicing operation requires at least one parameter",
		sliceCallTooManyParamError: "slicing operation has too many parameters",
		assertExpressionError:      "The second parameter of the assert statement must be an expression",
		notImplemented:             "[%s]",
		forceCreateSymbolFailed:    "BUG: Unable to force creation of symbol `%s`",
		autoCreateSymbolFailed:     "BUG: Unable to automatically create symbol `%s`",
		integerIsTooLarge:          "Unable to parse `%s` as an integer , because it is too large for int64",
		contParseNumber:            "Unable to parse numeric literal: %s",
	},
}

func (l *LuaTranslator) GetConstError(e constError) string {
	if l.language == en {
		return string(e)
	}
	if constsInfo, ok := i18n[l.language]; ok {
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
