package yakvm

import (
	"bytes"
	"fmt"
	"math/rand"
	"unicode"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
)

func NewBytesMethodFactory(f func([]byte) interface{}) MethodFactory {
	return func(vm *Frame, i interface{}) interface{} {
		raw, ok := i.([]byte)
		if !ok {
			raw = []byte(fmt.Sprint(i))
		}
		return f(raw)
	}
}

var bytesBuildinMethod = map[string]*buildinMethod{
	"First": {
		Name:       "First",
		ParamTable: nil,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() rune {
				return rune(s[0])
			}
		}),
		Description: "Get the first character of the byte array",
	},
	"Reverse": {
		Name:       "Reverse",
		ParamTable: nil,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() []byte {
				// runes is to deal with Chinese character problems, this is reasonable
				runes := []rune(string(s))
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return []byte(string(runes))
			}
		}),
		Description: "Reverse word Section array",
	},
	"Shuffle": {
		Name:       "Shuffle",
		ParamTable: nil,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() []byte {
				runes := []rune(string(s))
				rand.Shuffle(len(runes), func(i, j int) {
					runes[i] = runes[j]
				})
				return []byte(string(runes))
			}
		}),
		Description: "Randomly scramble the byte array",
	},
	"Fuzz": {
		Name:       "Fuzz",
		ParamTable: nil,
		Snippet:    `Fuzz(${1:{"params": "value"\}})$0`,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(i ...interface{}) []string {
				var opts []mutate.FuzzConfigOpt
				if len(i) > 0 {
					opts = append(opts, mutate.Fuzz_WithParams(i[0]))
				}

				if len(i) > 1 {
					log.Warn("string.Fuzz only need one param as {{params(...)}} source")
				}

				res, err := mutate.FuzzTagExec(string(s), opts...)
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
		ParamTable: []string{"subslice"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(subslice []byte) bool {
				if len(subslice) == 0 {
					return true
				}
				return bytes.Contains(s, subslice)
			}
		}),
		Description: "Determine whether byte array contains sub-byte array",
	},
	"IContains": {
		Name:       "IContains",
		ParamTable: []string{"subslice"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(subslice []byte) bool {
				if len(subslice) == 0 {
					return true
				}
				return bytes.Contains(bytes.ToLower(s), bytes.ToLower(subslice))
			}
		}),
		Description: "Determine whether the byte array contains a sub-byte array, ignore the case",
	},
	"ReplaceN": {
		Name:       "ReplaceN",
		ParamTable: []string{"old", "new", "n"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(old, new []byte, n int) []byte {
				return bytes.Replace(s, old, new, n)
			}
		},
		),
		Description: "Replace the sub-byte array in the byte array",
	},
	"ReplaceAll": {
		Name:       "ReplaceAll",
		ParamTable: []string{"old", "new"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(old, new []byte) []byte {
				return bytes.ReplaceAll(s, old, new)
			}
		},
		),
		Description: "Replace all sub-byte arrays in the byte array",
	},
	"Split": {
		Name:       "Split",
		ParamTable: []string{"separator"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(sep []byte) [][]byte {
				return bytes.Split(s, sep)
			}
		},
		),
		Description: "Split byte array",
	},
	"SplitN": {
		Name:       "SplitN",
		ParamTable: []string{"separator", "n"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(sep []byte, n int) [][]byte {
				return bytes.SplitN(s, sep, n)
			}
		},
		),
		Description: "Split byte array into at most N parts",
	},
	"Join": {
		Name:       "Join",
		ParamTable: []string{"separator"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(i interface{}) []byte {
				return bytes.Join(utils.InterfaceToBytesSlice(i), s)
			}
		},
		),
		Description: "Connect the byte array",
	},
	"Trim": {
		Name:            "Trim",
		ParamTable:      []string{"cutset"},
		IsVariadicParam: true,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(cutslice ...[]byte) []byte {
				if cutslice != nil {
					return bytes.Trim(s, string(bytes.Join(cutslice, []byte{})))
				}

				return bytes.TrimSpace(s)
			}
		},
		),
		Description: "Remove The cutsets at both ends of the byte array",
	},
	"TrimLeft": {
		Name:            "TrimLeft",
		ParamTable:      []string{"cutstr"},
		IsVariadicParam: true,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(cutslice ...[]byte) []byte {
				if cutslice != nil {
					return bytes.TrimLeft(s, string(bytes.Join(cutslice, []byte{})))
				}

				return bytes.TrimLeftFunc(s, unicode.IsSpace)
			}
		}),
		Description: "Remove the cutset at the left end of the byte array",
	},
	"TrimRight": {
		Name:       "TrimRight",
		ParamTable: []string{"cutstr"}, IsVariadicParam: true,
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(cutslice ...[]byte) []byte {
				if cutslice != nil {
					return bytes.TrimRight(s, string(bytes.Join(cutslice, []byte{})))
				}

				return bytes.TrimRightFunc(s, unicode.IsSpace)
			}
		}),
		Description: "Remove the cutset at the right end of the byte array",
	},
	"HasPrefix": {
		Name:       "HasPrefix",
		ParamTable: []string{"prefix"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(prefix []byte) bool {
				return bytes.HasPrefix(s, prefix)
			}
		},
		),
		Description: ". Determine whether the byte array starts with prefix",
	},
	"RemovePrefix": {
		Name:       "RemovePrefix",
		ParamTable: []string{"prefix"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(prefix []byte) []byte {
				return bytes.TrimPrefix(s, prefix)
			}
		},
		),
		Description: "Remove the prefix",
	},
	"HasSuffix": {
		Name:       "HasSuffix",
		ParamTable: []string{"suffix"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(suffix []byte) bool {
				return bytes.HasSuffix(s, suffix)
			}
		},
		),
		Description: "Determine whether the byte array ends with suffix",
	},
	"RemoveSuffix": {
		Name:       "RemoveSuffix",
		ParamTable: []string{"suffix"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(suffix []byte) []byte {
				return bytes.TrimSuffix(s, suffix)
			}
		},
		),
		Description: "Remove the suffix",
	},
	"Zfill": {
		Name:       "Zfill",
		ParamTable: []string{"width"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			zeroBytes := []byte{'0'}
			return func(width int) []byte {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					return append(bytes.Repeat(zeroBytes, width-lenOfS), s...)
				}
			}
		},
		),
		Description: "byte array with 0",
	},
	"Rzfill": {
		Name:       "Rzfill",
		ParamTable: []string{"width"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			zeroBytes := []byte{'0'}
			return func(width int) []byte {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					return append(s, bytes.Repeat(zeroBytes, width-lenOfS)...)
				}
			}
		},
		),
		Description: ". Fill the right side of the byte array with 0",
	},
	"Ljust": {
		Name:       "Ljust",
		ParamTable: []string{"width"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			zeroBytes := []byte{' '}
			return func(width int, fill ...[]byte) []byte {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					fillBytes := zeroBytes
					if len(fill) > 0 {
						fillBytes = fill[0]
					}
					return append(s, bytes.Repeat(fillBytes, width-lenOfS)...)
				}
			}
		},
		),
		Description: "Fill the left side of the byte array with spaces",
	},
	"Rjust": {
		Name:       "Rjust",
		ParamTable: []string{"width"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			zeroBytes := []byte{' '}
			return func(width int, fill ...[]byte) []byte {
				lenOfS := len(s)
				if width <= lenOfS {
					return s
				} else {
					fillBytes := zeroBytes
					if len(fill) > 0 {
						fillBytes = fill[0]
					}
					return append(bytes.Repeat(fillBytes, width-lenOfS), s...)
				}
			}
		},
		),
		Description: ". Fill the right side of the byte array with spaces",
	},
	"Count": {
		Name:       "Count",
		ParamTable: []string{"subslice"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(subslice []byte) int {
				return bytes.Count(s, subslice)
			}
		},
		),
		Description: "Count the number of occurrences of subslice in the byte array",
	},
	"Find": {
		Name:       "Find",
		ParamTable: []string{"subslice"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(subslice []byte) int {
				return bytes.Index(s, subslice)
			}
		},
		),
		Description: "Find the position where subslice first appears in the byte array, if not found, return -1",
	},
	"Rfind": {
		Name:       "Rfind",
		ParamTable: []string{"subslice"},
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func(subslice []byte) int {
				return bytes.LastIndex(s, subslice)
			}
		},
		),
		Description: "Find byte array The position where subslice last appeared. If not found, -1 is returned.",
	},
	"Lower": {
		Name: "Lower",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() []byte {
				return bytes.ToLower(s)
			}
		},
		),
		Description: "Convert byte array to lowercase",
	},
	"Upper": {
		Name: "Upper",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() []byte {
				return bytes.ToUpper(s)
			}
		},
		),
		Description: "Convert the byte array to uppercase",
	},
	"Title": {
		Name: "Title",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() []byte {
				return bytes.Title(s)
			}
		},
		),
		Description: ". Convert the byte array to Title format (i.e. The first letter of all words is capitalized, the rest are lowercase)",
	},
	"IsLower": {
		Name: "IsLower",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return bytes.Equal(bytes.ToLower(s), s)
			}
		},
		),
		Description: "Determine whether byte array is lowercase",
	},
	"IsUpper": {
		Name: "IsUpper",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return bytes.Equal(bytes.ToUpper(s), s)
			}
		},
		),
		Description: "Determine whether the byte array is uppercase",
	},
	"IsTitle": {
		Name: "IsTitle",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return bytes.Equal(bytes.Title(s), s)
			}
		},
		),
		Description: "Determine whether the byte array is in Title format",
	},
	"IsAlpha": {
		Name: "IsAlpha",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[a-zA-Z]+$`)
			}
		},
		),
		Description: ". Determine whether the byte array is the letter",
	},
	"IsDigit": {
		Name: "IsDigit",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[0-9]+$`)
			}
		},
		),
		Description: "Determine whether the byte array Fill the left side of the digital",
	},
	"IsAlnum": {
		Name: "IsAlnum",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[a-zA-Z0-9]+$`)
			}
		},
		),
		Description: "Determine the word Whether the section array is letters or numbers",
	},
	"IsPrintable": {
		Name: "IsPrintable",
		HandlerFactory: NewBytesMethodFactory(func(s []byte) interface{} {
			return func() bool {
				return utils.MatchAllOfRegexp(s, `^[\x20-\x7E]+$`)
			}
		},
		),
		Description: "Determine whether the byte array is a printable character",
	},
}

func init() {
	aliasBytesBuildinMethod("ReplaceAll", "Replace")
	aliasBytesBuildinMethod("Find", "IndexOf")
	aliasBytesBuildinMethod("Rfind", "LastIndexOf")
	aliasBytesBuildinMethod("HasPrefix", "StartsWith")
	aliasBytesBuildinMethod("HasSuffix", "EndsWith")
}

func aliasBytesBuildinMethod(origin string, target string) {
	aliasBuildinMethod(bytesBuildinMethod, origin, target)
}
