package yakvm

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/yaklang/yaklang/common/go-funk"
)

var buildMethodsArray = map[string]interface{}{}

func NewArrayMethodFactory(f func(*Frame, *Value, interface{}) interface{}) MethodFactory {
	return func(vm *Frame, value interface{}) interface{} {
		v := value.(*Value)
		return f(vm, v, v.Value)
	}
}

var arrayBuildinMethod map[string]*buildinMethod

func aliasArrayBuildinMethod(origin string, target string) {
	aliasBuildinMethod(arrayBuildinMethod, origin, target)
}

func init() {
	arrayBuildinMethod = map[string]*buildinMethod{
		"Append": {
			Name:            "Append",
			ParamTable:      []string{"element"},
			IsVariadicParam: true,
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(vi ...interface{}) {
					rv := reflect.ValueOf(i)

					sliceLen := rv.Len() + len(vi)
					vals := make([]interface{}, 0, sliceLen)
					for i := 0; i < rv.Len(); i++ {
						vals = append(vals, rv.Index(i).Interface())
					}
					for _, v := range vi {
						vals = append(vals, v)
					}
					elementType := GuessBasicType(vals...)
					sliceType := reflect.SliceOf(elementType)

					newSlice := reflect.MakeSlice(sliceType, sliceLen, sliceLen)
					for index, e := range vals {
						val := reflect.ValueOf(e)
						err := vm.AutoConvertReflectValueByType(&val, elementType)
						if err != nil {
							panic(fmt.Sprintf("cannot convert %v to %v", val.Type(), elementType))
						}
						newSlice.Index(index).Set(val)
					}
					ref.Assign(vm, NewAutoValue(newSlice.Interface()))
				}
			}),
			Description: "Gets the length from the array/Slice to append the last element",
		},
		"Pop": {
			Name: "Pop",
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(na ...int) interface{} {
					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					n := vLen - 1
					if len(na) > 0 {
						n = na[0]
						if n < 0 {
							n = vLen - 1 + n - 1
						}
						if n > vLen-1 || n < 0 {
							n = vLen - 1
						}
					}
					ret := rv.Index(n).Interface()
					newSlice := reflect.AppendSlice(rv.Slice(0, n), rv.Slice(n+1, vLen))
					ref.Assign(vm, NewAutoValue(newSlice.Interface()))
					return ret
				}
			}),
			Description: "Pop the array/An element of the slice, the default is the last one",
		},
		"Extend": {
			Name:       "Extend",
			ParamTable: []string{"newSlice"},
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(vi interface{}) {
					rv2 := reflect.ValueOf(vi)
					rt2 := rv2.Type().Kind()
					if rt2 != reflect.Array && rt2 != reflect.Slice {
						panic(fmt.Sprintf("extend argument[%v] is not iterable", rv2.Type()))
					}
					rv := reflect.ValueOf(i)

					sliceLen := rv.Len() + rv2.Len()
					vals := make([]interface{}, 0, sliceLen)
					for i := 0; i < rv.Len(); i++ {
						vals = append(vals, rv.Index(i).Interface())
					}
					for i := 0; i < rv2.Len(); i++ {
						vals = append(vals, rv2.Index(i).Interface())
					}
					elementType := GuessBasicType(vals...)
					sliceType := reflect.SliceOf(elementType)

					newSlice := reflect.MakeSlice(sliceType, sliceLen, sliceLen)
					for index, e := range vals {
						val := reflect.ValueOf(e)
						err := vm.AutoConvertReflectValueByType(&val, elementType)
						if err != nil {
							panic(fmt.Sprintf("cannot convert %v to %v", val.Type(), elementType))
						}
						newSlice.Index(index).Set(val)
					}
					ref.Assign(vm, NewAutoValue(newSlice.Interface()))
				}
			}),
			Description: "Use a new array/Slice and extend the original array/Slice",
		},
		"Length": {
			Name: "Length",
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func() int {
					r := reflect.ValueOf(i)
					_ = r.Len()
					switch r.Kind() {
					case reflect.Array:
					case reflect.Slice:
					default:
						panic(fmt.Sprintf("caller type: %v cannot call `length`", r.Type()))
					}
					return r.Len()
				}
			}),
			Description: "Get the length",
		},
		"Capability": {
			Name: "Capability",
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func() int {
					r := reflect.ValueOf(i)
					switch r.Kind() {
					case reflect.Array:
					case reflect.Slice:
					default:
						panic(fmt.Sprintf("caller type: %v cannot call `cap`", r.Type()))
					}
					return r.Cap()
				}
			}),
			Description: "Gets the capacity",
		},
		"StringSlice": {
			Name:        "StringSlice",
			Description: "Convert to []string",
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func() []string {
					rv := reflect.ValueOf(i)
					if rv.Len() <= 0 {
						return nil
					}

					vLen := rv.Len()
					result := make([]string, vLen)
					for i := 0; i < vLen; i++ {
						val := rv.Index(i)
						if a, ok := val.Interface().([]byte); ok {
							result[i] = string(a)
						} else if s, ok := val.Interface().(string); ok {
							result[i] = s
						} else if !val.IsValid() || val.IsZero() {
							result[i] = ""
						} else {
							result[i] = fmt.Sprint(val.Interface())
						}
					}
					return result
				}
			}),
		},
		"GeneralSlice": {
			Name:        "GeneralSlice",
			Description: "is converted into the most generalized Slice type []any ([]interface{})",
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func() []interface{} {
					return funk.Map(i, func(i interface{}) interface{} {
						return i
					}).([]interface{})
				}
			}),
		},
		"Shift": {
			Name:        "Shift",
			Description: "Remove an element from the beginning of the data",
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func() interface{} {
					rv := reflect.ValueOf(i)
					originLen := rv.Len()
					if originLen <= 0 {
						return nil
					}

					target := reflect.MakeSlice(rv.Type(), originLen-1, originLen-1)
					for i := 0; i < originLen-1; i++ {
						target.Index(i).Set(rv.Index(i + 1))
					}
					value.Assign(frame, NewAutoValue(target.Interface()))
					return rv.Index(0).Interface()
				}
			}),
		},
		"Unshift": {
			Name:        "Unshift",
			Description: "from the data Add an element",
			ParamTable:  []string{"element"},
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, caller interface{}) interface{} {
				return func(raw interface{}) {
					rv := reflect.ValueOf(caller)
					vLen := rv.Len()

					vals := make([]interface{}, vLen+1)
					vals[0] = raw
					for i := 0; i < vLen; i++ {
						vals[i+1] = rv.Index(i).Interface()
					}
					target := reflect.MakeSlice(reflect.SliceOf(GuessBasicType(vals...)), vLen+1, vLen+1)
					for i := 0; i < vLen+1; i++ {
						target.Index(i).Set(reflect.ValueOf(vals[i]))
					}
					value.Assign(frame, NewAutoValue(target.Interface()))
				}
			}),
		},
		"Map": {
			Name:        "Map",
			Description: "data at the specified position/The elements in the slice are operated and return the result",
			ParamTable:  []string{"mapFunc"},
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func(handler func(i interface{}) interface{}) interface{} {
					return funk.Map(i, handler)
				}
			}),
		},
		"Filter": {
			Name:        "Filter",
			Description: "data at the specified position/The elements in the slice are operated and return the result",
			ParamTable:  []string{"filterFunc"},
			HandlerFactory: NewArrayMethodFactory(func(frame *Frame, value *Value, i interface{}) interface{} {
				return func(handler func(i interface{}) bool) interface{} {
					return funk.Filter(i, handler)
				}
			}),
		},
		"Insert": {
			Name:       "Insert",
			ParamTable: []string{"index", "element"},
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(n int, vi interface{}) {
					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					if n > vLen {
						n = vLen
					} else if n < 0 {
						n = vLen + n
						if n < 0 {
							n = 0
						}
					}

					sliceLen := rv.Len() + 1
					vals := make([]interface{}, sliceLen)
					for i := 0; i < n; i++ {
						vals[i] = rv.Index(i).Interface()
					}
					vals[n] = vi

					for i := n + 1; i < vLen+1; i++ {
						vals[i] = rv.Index(i - 1).Interface()
					}
					elementType := GuessBasicType(vals...)
					sliceType := reflect.SliceOf(elementType)

					newSlice := reflect.MakeSlice(sliceType, sliceLen, sliceLen)
					for index, e := range vals {
						val := reflect.ValueOf(e)
						err := vm.AutoConvertReflectValueByType(&val, elementType)
						if err != nil {
							panic(fmt.Sprintf("cannot convert %v to %v", val.Type(), elementType))
						}
						newSlice.Index(index).Set(val)
					}
					ref.Assign(vm, NewAutoValue(newSlice.Interface()))
				}
			}),
			Description: "Inserts the element",
		},
		"Remove": {
			Name:       "Remove",
			ParamTable: []string{"element"},
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(vi interface{}) {
					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					n := -1
					for i := 0; i < vLen; i++ {
						if funk.Equal(rv.Index(i).Interface(), vi) {
							n = i
							break
						}
					}
					newSlice := reflect.AppendSlice(rv.Slice(0, n), rv.Slice(n+1, vLen))
					ref.Assign(vm, NewAutoValue(newSlice.Interface()))
				}
			}),
			Description: "Remove the array/The first element of the slice",
		},
		"Reverse": {
			Name: "Reverse",
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func() {
					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					for i := 0; i < vLen/2; i++ {
						temp := reflect.ValueOf(rv.Index(i).Interface())
						temp2 := reflect.ValueOf(rv.Index(vLen - 1 - i).Interface())
						rv.Index(i).Set(temp2)
						rv.Index(vLen - 1 - i).Set(temp)
					}
				}
			}),
			Description: "Reverse the array/Slice",
		},
		"Sort": {
			Name:            "Sort",
			ParamTable:      []string{"reverse"},
			IsVariadicParam: true,
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(reversea ...bool) {
					reverse := false
					if len(reversea) > 0 {
						reverse = reversea[0]
					}
					rv := reflect.ValueOf(i)

					sort.SliceStable(i, func(i, j int) bool {
						if reverse {
							return fmt.Sprint(rv.Index(i).Interface()) > fmt.Sprint(rv.Index(j).Interface())
						}
						return fmt.Sprint(rv.Index(i).Interface()) < fmt.Sprint(rv.Index(j).Interface())
					})
				}
			}),
			Description: "Sort the array/Slice",
		},
		"Clear": {
			Name: "Clear",
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func() {
					nv := reflect.MakeSlice(reflect.SliceOf(literalReflectType_Interface), 0, 0)
					ref.Assign(vm, NewAutoValue(nv.Interface()))
				}
			}),
			Description: "at the beginning Clear the array/Slice",
		},
		"Count": {
			Name:       "Count",
			ParamTable: []string{"element"},
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(vi interface{}) int {
					n := 0

					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					for i := 0; i < vLen; i++ {
						if funk.Equal(rv.Index(i).Interface(), vi) {
							n++
						}
					}
					return n
				}
			}),
			Description: "Calculate the array/The number of elements in the slice",
		},
		"Index": {
			Name:       "Index",
			ParamTable: []string{"indexInt"},
			HandlerFactory: NewArrayMethodFactory(func(vm *Frame, ref *Value, i interface{}) interface{} {
				return func(n int) interface{} {
					rv := reflect.ValueOf(i)
					vLen := rv.Len()
					if n < 0 {
						n = vLen + n
					}
					if n > vLen-1 {
						n = vLen - 1
					} else if n < 0 {
						n = 0
					}

					return rv.Index(n).Interface()
				}
			}),
			Description: "Return the array/The nth element in the slice",
		},
	}

	// alias
	aliasArrayBuildinMethod("Append", "Push")
	aliasArrayBuildinMethod("Extend", "Merge")
	aliasArrayBuildinMethod("Capability", "Cap")
	aliasArrayBuildinMethod("Length", "Len")
}
