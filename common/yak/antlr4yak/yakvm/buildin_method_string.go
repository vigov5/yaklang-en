package yakvm

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
)

func NewStringMethodFactory(f func(string) interface{}) MethodFactory {
	return func(vm *Frame, i interface{}) interface{} {
		raw, ok := i.(string)
		if !ok {
			raw = fmt.Sprint(i)
		}
		return f(raw)
	}
}

var stringBuildinMethod = map[string]*buildinMethod{
	"First": {
		Name:       "First",
		ParamTable: nil,

		HandlerFactory: NewStringMethodFactory(func(c string) interface{} {
			return func() rune {
				return rune(c[0])
			}
		}),
		Description: "Get the string The first character",
	},
	"Reverse": {
		Name:       "Reverse",
		ParamTable: nil,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() string {
				// runes is to deal with Chinese character problems, this is reasonable
				runes := []rune(s)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return string(runes)
			}
		}),
		Description: "Reverse characters String",
	},
	"Shuffle": {
		Name:       "Shuffle",
		ParamTable: nil,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() string {
				raw := []rune(s)
				rand.Shuffle(len(raw), func(i, j int) {
					raw[i] = raw[j]
				})
				return string(raw)
			}
		}),
		Description: "Randomly scramble the string",
	},
	"Fuzz": {
		Name:       "Fuzz",
		ParamTable: nil,
		Snippet:    `Fuzz(${1:{"params": "value"\}})$0`,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(i ...interface{}) []string {
				var opts []mutate.FuzzConfigOpt
				if len(i) > 0 {
					opts = append(opts, mutate.Fuzz_WithParams(i[0]))
				}

				if len(i) > 1 {
					log.Warn("string.Fuzz only need one param as {{params(...)}} source")
				}

				res, err := mutate.FuzzTagExec(s, opts...)
				if err != nil {
					log.Errorf("fuzz tag error: %s", err)
					return nil
				}
				return res
			}
		}),
	},
	"Contains": {
		Name:       "Contains",
		ParamTable: []string{"substr"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(substr string) bool {
				if len(substr) == 0 {
					return true
				}
				return strings.Contains(s, substr)
			}
		}),
		Description: "Determine whether the string contains a substring",
	},
	"IContains": {
		Name:       "IContains",
		ParamTable: []string{"substr"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(substr string) bool {
				if len(substr) == 0 {
					return true
				}
				return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
			}
		}),
		Description: "Determine whether the string contains a substring",
	},
	"ReplaceN": {
		Name:       "ReplaceN",
		ParamTable: []string{"old", "new", "n"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(old, new string, n int) string {
				return strings.Replace(s, old, new, n)
			}
		},
		),
		Description: "Replace the substring in the string",
	},
	"ReplaceAll": {
		Name:       "ReplaceAll",
		ParamTable: []string{"old", "new"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(old, new string) string {
				return strings.ReplaceAll(s, old, new)
			}
		},
		),
		Description: "replaces all substrings in the string.",
	},
	"Split": {
		Name:       "Split",
		ParamTable: []string{"separator"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(sep string) []string {
				return strings.Split(s, sep)
			}
		},
		),
		Description: "Split the string",
	},
	"SplitN": {
		Name:       "SplitN",
		ParamTable: []string{"separator", "n"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(sep string, n int) []string {
				return strings.SplitN(s, sep, n)
			}
		},
		),
		Description: "Split the string into N parts at most",
	},
	"Join": {
		Name:       "Join",
		ParamTable: []string{"slice"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(i interface{}) string {
				return strings.Join(utils.InterfaceToStringSlice(i), s)
			}
		},
		),
		Description: "Connect the string",
	},
	"Trim": {
		Name:            "Trim",
		ParamTable:      []string{"cutstr"},
		IsVariadicParam: true,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(cutset ...string) string {
				if cutset != nil {
					return strings.Trim(s, strings.Join(cutset, ""))
				}

				return strings.TrimSpace(s)
			}
		},
		),
		Description: "removes the cutsets at both ends of the string.",
	},
	"TrimLeft": {
		Name:            "TrimLeft",
		ParamTable:      []string{"cutstr"},
		IsVariadicParam: true,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(cutset ...string) string {
				if cutset != nil {
					return strings.TrimLeft(s, strings.Join(cutset, ""))
				}

				return strings.TrimLeftFunc(s, unicode.IsSpace)
			}
		}),
		Description: "Remove the cutset at the left end of the string",
	},
	"TrimRight": {
		Name:       "TrimRight",
		ParamTable: []string{"cutstr"}, IsVariadicParam: true,
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(cutset ...string) string {
				if cutset != nil {
					return strings.TrimRight(s, strings.Join(cutset, ""))
				}

				return strings.TrimRightFunc(s, unicode.IsSpace)
			}
		}),
		Description: "Removes the cutset at the right end of the string",
	},
	"HasPrefix": {
		Name:       "HasPrefix",
		ParamTable: []string{"prefix"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(prefix string) bool {
				return strings.HasPrefix(s, prefix)
			}
		},
		),
		Description: "Determine whether the string starts with prefix",
	},
	"RemovePrefix": {
		Name:       "RemovePrefix",
		ParamTable: []string{"prefix"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(prefix string) string {
				return strings.TrimPrefix(s, prefix)
			}
		},
		),
		Description: "Remove the prefix",
	},
	"HasSuffix": {
		Name:       "HasSuffix",
		ParamTable: []string{"suffix"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(suffix string) bool {
				return strings.HasSuffix(s, suffix)
			}
		},
		),
		Description: "Determine whether the string ends with suffix",
	},
	"RemoveSuffix": {
		Name:       "RemoveSuffix",
		ParamTable: []string{"suffix"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(suffix string) string {
				return strings.TrimSuffix(s, suffix)
			}
		},
		),
		Description: "Remove the suffix",
	},
	"Zfill": {
		Name:       "Zfill",
		ParamTable: []string{"width"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(width int) string {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					return strings.Repeat("0", width-lenOfS) + s
				}
			}
		},
		),
		Description: "Fill the left side of the string with 0",
	},
	"Rzfill": {
		Name:       "Rzfill",
		ParamTable: []string{"width"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(width int) string {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					return s + strings.Repeat("0", width-lenOfS)
				}
			}
		},
		),
		Description: "Fill the right side of the string with 0",
	},
	"Ljust": {
		Name:       "Ljust",
		ParamTable: []string{"width"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(width int, fill ...string) string {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					fillStr := " "
					if len(fill) > 0 {
						fillStr = fill[0]
					}
					return s + strings.Repeat(fillStr, width-lenOfS)
				}
			}
		},
		),
		Description: "Left side of the string Fill the side with spaces",
	},
	"Rjust": {
		Name:       "Rjust",
		ParamTable: []string{"width"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(width int, fill ...string) string {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					fillStr := " "
					if len(fill) > 0 {
						fillStr = fill[0]
					}
					return strings.Repeat(fillStr, width-lenOfS) + s
				}
			}
		},
		),
		Description: "Fill the right side of the string with spaces",
	},
	"Count": {
		Name:       "Count",
		ParamTable: []string{"substr"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(substr string) int {
				return strings.Count(s, substr)
			}
		},
		),
		Description: "Counts the number of times substr appears in the string",
	},
	"Find": {
		Name:       "Find",
		ParamTable: []string{"substr"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(substr string) int {
				return strings.Index(s, substr)
			}
		},
		),
		Description: "Find the first occurrence of substr in the string, if not found, return -1",
	},
	"Rfind": {
		Name:       "Rfind",
		ParamTable: []string{"substr"},
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func(substr string) int {
				return strings.LastIndex(s, substr)
			}
		},
		),
		Description: "Find the last occurrence of substr in the string, if not found Return -1",
	},
	"Lower": {
		Name: "Lower",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() string {
				return strings.ToLower(s)
			}
		},
		),
		Description: "Convert the string to lowercase",
	},
	"Upper": {
		Name: "Upper",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() string {
				return strings.ToUpper(s)
			}
		},
		),
		Description: "Convert the string to uppercase",
	},
	"Title": {
		Name: "Title",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() string {
				return strings.Title(s)
			}
		},
		),
		Description: "converts the string into Title format (i.e. the first character of all words). Letters are uppercase, the rest are lowercase)",
	},
	"IsLower": {
		Name: "IsLower",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return strings.ToLower(s) == s
			}
		},
		),
		Description: "Determine whether the string is lowercase",
	},
	"IsUpper": {
		Name: "IsUpper",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return strings.ToUpper(s) == s
			}
		},
		),
		Description: "determines whether the string is uppercase.",
	},
	"IsTitle": {
		Name: "IsTitle",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return strings.Title(s) == s
			}
		},
		),
		Description: "determines whether the string is in Title format.",
	},
	"IsAlpha": {
		Name: "IsAlpha",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[a-zA-Z]+$`)
			}
		},
		),
		Description: "Determines whether the string is a letter",
	},
	"IsDigit": {
		Name: "IsDigit",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[0-9]+$`)
			}
		},
		),
		Description: "Determine whether the string is a number",
	},
	"IsAlnum": {
		Name: "IsAlnum",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[a-zA-Z0-9]+$`)
			}
		},
		),
		Description: "Determine whether the string is a letter or number",
	},
	"IsPrintable": {
		Name: "IsPrintable",
		HandlerFactory: NewStringMethodFactory(func(s string) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[\x20-\x7E]+$`)
			}
		},
		),
		Description: "Determines whether the string is a printable character",
	},
}

func init() {
	aliasStringBuildinMethod("ReplaceAll", "Replace")
	aliasStringBuildinMethod("Find", "IndexOf")
	aliasStringBuildinMethod("Rfind", "LastIndexOf")
	aliasStringBuildinMethod("HasPrefix", "StartsWith")
	aliasStringBuildinMethod("HasSuffix", "EndsWith")
}

func aliasStringBuildinMethod(origin string, target string) {
	aliasBuildinMethod(stringBuildinMethod, origin, target)
}
