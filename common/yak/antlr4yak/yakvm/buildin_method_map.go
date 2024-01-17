package yakvm

import (
	"fmt"
	"reflect"

	"github.com/yaklang/yaklang/common/go-funk"
)

func aliasMapBuildinMethod(origin string, target string) {
	aliasBuildinMethod(mapBuildinMethod, origin, target)
}

func NewMapMethodFactory(f func(frame *Frame, v *Value) interface{}) MethodFactory {
	return func(vm *Frame, i interface{}) interface{} {
		mapV := i.(*Value)
		return f(vm, mapV)
	}
}

var mapBuildinMethod map[string]*buildinMethod

func init() {
	mapBuildinMethod = map[string]*buildinMethod{
		"Keys": {
			Name:       "Keys",
			ParamTable: nil,
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func() interface{} {
					return funk.Keys(v.Value)
				}
			}),
			Description: "Get the keys of all elements",
		},
		"Values": {
			Name:       "Values",
			ParamTable: nil,
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func() interface{} {
					return funk.Values(v.Value)
				}
			}),
			Description: "Get the values of all elements",
		},
		"Entries": {
			Name:       "Entries",
			ParamTable: nil,
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func() interface{} {
					ikeys := funk.Keys(v.Value)
					if ikeys == nil {
						return []interface{}{}
					}
					refV := reflect.ValueOf(v.Value)
					var result [][]interface{}
					if funk.IsIteratee(ikeys) {
						funk.ForEach(ikeys, func(key interface{}) {
							v := refV.MapIndex(reflect.ValueOf(key))
							result = append(result, []interface{}{key, v.Interface()})
						})
					}
					return result
				}
			}),
			Description: "Get the entity of all elements",
		},
		"ForEach": {
			Name:       "ForEach",
			ParamTable: []string{"handler"},
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func(f func(k, v interface{})) interface{} {
					funk.ForEach(v.Value, f)
					return nil
				}
			}),
			Description: "Traverse the elements",
		},
		"Set": {
			Name:       "Set",
			ParamTable: []string{"key", "value"},
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, caller *Value) interface{} {
				return func(k, v interface{}) interface{} {
					refV, err := frame.AutoConvertYakValueToNativeValue(NewAutoValue(v))
					if err != nil {
						panic(fmt.Sprintf("runtime error: cannot assign %v to map[index]", v))
					}
					callerRefV := reflect.ValueOf(caller.Value)
					if callerRefV.MapIndex(reflect.ValueOf(k)).IsValid() {
						refV = refV.Convert(callerRefV.MapIndex(reflect.ValueOf(k)).Type())
					}
					callerRefV.SetMapIndex(reflect.ValueOf(k), refV)
					return true
				}
			}),
			Description: "Set the value of the element, add it if the key does not exist",
		},
		"Remove": {
			Name:       "Remove",
			ParamTable: []string{"key"},
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, val *Value) interface{} {
				return func(paramK interface{}) interface{} {
					refMap := reflect.ValueOf(val.Value)
					refMap.SetMapIndex(reflect.ValueOf(paramK), reflect.ValueOf(nil))
					return nil
				}
			}),
			Description: "Remove a value",
		},
		"Has": {
			Name:       "Has",
			ParamTable: []string{"key"},
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func(k interface{}) interface{} {
					var ok bool
					funk.ForEach(v.Value, func(k_, v interface{}) {
						if funk.Equal(k, k_) {
							ok = true
						}
					})
					return ok
				}
			}),
			Description: "Determine whether the map element contains the key",
		},
		"Length": {
			Name:       "Length",
			ParamTable: nil,
			HandlerFactory: NewMapMethodFactory(func(frame *Frame, v *Value) interface{} {
				return func() interface{} {
					return reflect.ValueOf(v.Value).Len()
				}
			}),
			Description: "Get the length of the element",
		},
	}
	aliasMapBuildinMethod("Entries", "Items")
	aliasMapBuildinMethod("Remove", "Delete")
	aliasMapBuildinMethod("Has", "IsExisted")
	aliasMapBuildinMethod("Length", "Len")
}
