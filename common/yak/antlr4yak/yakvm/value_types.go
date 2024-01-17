package yakvm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yaklang/yaklang/common/utils"
)

var (
	literalReflectType_Byte              = reflect.TypeOf(byte(0))
	literalReflectType_Bytes             = reflect.TypeOf([]byte{})
	literalReflectType_String            = reflect.TypeOf("")
	literalReflectType_Int               = reflect.TypeOf(0)
	literalReflectType_Int8              = reflect.TypeOf(int8(0))
	literalReflectType_Int16             = reflect.TypeOf(int16(0))
	literalReflectType_Int32             = reflect.TypeOf(int32(0))
	literalReflectType_Int64             = reflect.TypeOf(int64(0))
	literalReflectType_Uint              = reflect.TypeOf(uint(0))
	literalReflectType_Uint8             = reflect.TypeOf(uint8(0))
	literalReflectType_Uint16            = reflect.TypeOf(uint16(0))
	literalReflectType_Uint32            = reflect.TypeOf(uint32(0))
	literalReflectType_Uint64            = reflect.TypeOf(uint64(0))
	literalReflectType_Float32           = reflect.TypeOf(float32(0.1))
	literalReflectType_Float64           = reflect.TypeOf(float64(0.1))
	literalReflectType_Bool              = reflect.TypeOf(false)
	literalReflectType_Interface         = reflect.TypeOf((*interface{})(nil)).Elem()
	literalReflectType_YakFunction       = reflect.TypeOf(&Function{})
	literalReflectType_NativeFunction    = reflect.TypeOf(func() {})
	literalReflectType_NativeWarpFuntion = reflect.FuncOf([]reflect.Type{reflect.SliceOf(literalReflectType_Interface)}, []reflect.Type{literalReflectType_Interface}, true)
)

/*
Regarding the description of the type, we can use some simple methods to ensure that when
 1. You can also convert in golang => yak, bad types are cleaned up, such as uint and float32. Other types are acceptable.
    . However, due to limited circumstances, int64 and int are acceptable, but no matter how large they are, there will be a small problem (it should be noted that generally x64 systems , int64 and int are the same)
 2. So in During the calculation process, yak => . In the process of golang, we do not want to call anything of uint type
*/
func IsInt(v interface{}) bool {
	switch v.(type) {
	case int, int64, int8, int16, int32,
		uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func IsFloat(v interface{}) bool {
	switch v.(type) {
	case float64, float32:
		return true
	}
	return false
}

func GuessBasicType(vals ...interface{}) reflect.Type {
	var (
		anyT  = literalReflectType_Interface
		kindI reflect.Kind
	)
	if len(vals) <= 0 {
		return anyT
	}

	last := anyT
	for index, val := range vals {
		kindI = reflect.ValueOf(val).Kind()
		if index == 0 {
			// identifies the first type
			if kindI == reflect.String {
				last = literalReflectType_String
			} else if kindI == reflect.Uint8 {
				last = literalReflectType_Byte
			} else if IsInt(val) {
				last = literalReflectType_Int
			} else if kindI == reflect.Bool {
				last = literalReflectType_Bool
			} else if IsFloat(val) {
				last = literalReflectType_Float64
			}
			continue
		}

		if kindI == reflect.String {
			// is required, there is no compatibility issue with this type
			if last.Kind() != reflect.String {
				return anyT
			}
		} else if IsInt(val) {
			// . Generally speaking, Int and Float should be mutually convertible. Use the most compatible type
			// . Compatibility float64 is the most compatible
			if last.Kind() != reflect.Int {
				if last.Kind() > reflect.Int && last.Kind() <= reflect.Float64 {
					continue
				}
				return anyT
			}
		} else if IsFloat(val) {
			if last.Kind() != reflect.Float64 {
				if last.Kind() >= reflect.Int && last.Kind() < reflect.Float64 {
					last = literalReflectType_Float64
					continue
				}
				return anyT
			}
		} else if kindI == reflect.Bool {
			if last.Kind() != reflect.Bool {
				return anyT
			}
		} else {
			return anyT
		}
	}
	return last
}

func GuessValuesTypeToBasicType(vals ...*Value) reflect.Type {
	var anyT = literalReflectType_Interface
	if len(vals) <= 0 {
		return anyT
	}

	last := anyT
	for index, i := range vals {
		if index == 0 {
			// identifies the first type
			if i.IsByte() {
				last = literalReflectType_Byte
			} else if i.IsString() {
				last = literalReflectType_String
			} else if i.IsBytes() {
				last = literalReflectType_Bytes
			} else if i.IsInt() {
				last = literalReflectType_Int
			} else if i.IsBool() {
				last = literalReflectType_Bool
			} else if i.IsFloat() {
				last = literalReflectType_Float64
			} else if i.Callable() {
				last = literalReflectType_Interface
			}
			continue
		}

		if i.IsStringOrBytes() {
			// is required, there is no compatibility issue with this type
			if last.Kind() != reflect.String && !i.IsBytes() {
				return anyT
			}
		} else if i.IsInt() {
			// . Generally speaking, Int and Float should be mutually convertible. Use the most compatible type
			// . Compatibility float64 is the most compatible
			if last.Kind() != reflect.Int {
				if last.Kind() > reflect.Int && last.Kind() <= reflect.Float64 {
					continue
				}
				return anyT
			}
		} else if i.IsFloat() {
			if last.Kind() != reflect.Float64 {
				if last.Kind() >= reflect.Int && last.Kind() < reflect.Float64 {
					last = literalReflectType_Float64
					continue
				}
				return anyT
			}
		} else if i.IsBool() {
			if last.Kind() != reflect.Bool {
				return anyT
			}
		} else {
			return anyT
		}
	}
	return last
}

//	func ImplicitTypeConversionForPlus(vals ...*Value) reflect.Type {
//		resultType := GuessValuesTypeToBasicType(vals...)
//		if resultType.Kind() == literalReflectType_Interface.Kind() {
//			isString := true
//			for _, val := range vals {
//				if !val.IsString() && !val.IsInt() {
//					isString = false
//					break
//				}
//			}
//			if isString {
//				resultType = literalReflectType_String
//			}
//		}
//		return resultType
//	}
func GuessValuesKindToBasicType(vals ...*Value) reflect.Kind {
	return GuessValuesTypeToBasicType(vals...).Kind()
}
func (v *Frame) AutoConvertYakValueToNativeValue(val *Value) (reflect.Value, error) {
	i := (*interface{})(nil)

	if val.Value == nil {
		return reflect.ValueOf(i), nil
	}
	refV := reflect.ValueOf(val.Value)

	if val.IsYakFunction() {
		err := v.AutoConvertReflectValueByType(&refV, literalReflectType_NativeWarpFuntion)
		if err != nil {
			return reflect.Value{}, err
		}
		return refV, nil
	}
	refType := GuessValuesTypeToBasicType(val)
	err := v.AutoConvertReflectValueByType(&refV, refType)
	if err != nil {
		return reflect.Value{}, err
	}
	return refV, nil
}
func (v *Frame) AutoConvertReflectValueByType(
	reflectValue *reflect.Value,
	reflectType /*, targetReflectType*/ reflect.Type,
) error {

	srcKind := reflectValue.Kind()

	if srcKind == reflect.Invalid {
		*reflectValue = reflect.Zero(reflectType) // work around `reflect: Call using zero Value argument`
		return nil
	}

	// has the same type, no need to convert
	if reflectType == reflectValue.Type() {
		return nil
	}

	targetKind := reflectType.Kind()
	if targetKind == reflect.Interface {
		// proves to be an alias. For example, time.Duration is an alias of the int64 type, but it has its own implementation method, so
		pkgPath := reflectValue.Type().PkgPath()
		if pkgPath != "" {
			return nil
		}
		//if targetReflectType != nil && yaklangspec.DontTyNormalize[targetReflectType] { // don't normalize input type
		//	return nil
		//}
		switch {
		case srcKind > reflect.Int && srcKind <= reflect.Int64:
			*reflectValue = reflect.ValueOf(int(reflectValue.Int()))
		case srcKind >= reflect.Uint && srcKind <= reflect.Uintptr:
			*reflectValue = reflect.ValueOf(int(reflectValue.Uint()))
		case srcKind == reflect.Float32:
			*reflectValue = reflect.ValueOf(reflectValue.Float())
		}

		return nil
	}

	tin := reflectValue.Type()
	if tin == reflectType {
		return nil
	}

	switch targetKind {
	case reflect.Struct:
		if srcKind == reflect.Ptr {
			tin = tin.Elem()
			if tin == reflectType {
				*reflectValue = reflectValue.Elem()
				return nil
			}
		}
	case reflect.Func:
		if tin == literalReflectType_YakFunction && reflectValue.Interface() != nil {
			if v == nil {
				return utils.Errorf("cannot bind Yaklang.Function Calling for VirtualMachine!")
			}
			f := reflectValue.Interface().(*Function)
			*reflectValue = reflect.MakeFunc(reflectType, func(args []reflect.Value) []reflect.Value {
				var vmArgs []*Value
				// fix: unpack variadic args
				if reflectType == literalReflectType_NativeWarpFuntion {
					newArgs, ok := args[0].Interface().([]interface{})
					if ok {
						vmArgs = make([]*Value, len(newArgs))
						for index, value := range newArgs {
							vmArgs[index] = NewAutoValue(value)
						}
					}
				}

				if vmArgs == nil {
					vmArgs = make([]*Value, len(args))
					for index, value := range args {
						vmArgs[index] = NewAutoValue(value.Interface())
					}
				}

				result := v.CallYakFunction(false, f, vmArgs)
				outCount := reflectType.NumOut()
				if outCount <= 0 {
					return nil
				}
				reflectReturn := reflect.ValueOf(result)

				if outCount == 1 {
					expected := reflectType.Out(0)
					err := v.AutoConvertReflectValueByType(&reflectReturn, expected)
					if err != nil {
						panic(fmt.Sprintf("runtime error: cannot convert `%v` to `%v`", reflectReturn.Type().String(), expected.String()))
					}
					return []reflect.Value{reflectReturn}
				}

				var outputResults = make([]reflect.Value, outCount)
				if reflectReturn.Kind() != reflect.Slice || reflectReturn.Len() != outCount {
					panic(fmt.Sprintf("unexpected return value count, we need `%d` values", outCount))
				}
				for i := 0; i < outCount; i++ {
					val := reflectReturn.Index(i)
					expectedType := reflectType.Out(i)
					err := v.AutoConvertReflectValueByType(&val, expectedType)
					if err != nil {
						panic(fmt.Sprintf("runtime error: cannot convert `%v` to `%v`", val.Type().String(), expectedType.String()))
					}
					outputResults[i] = val
				}
				return outputResults
			})
			return nil
		} else {
			return utils.Errorf("cannot convert yaklang.Function to native calling...")
		}
	case reflect.Slice, reflect.Array: // array type conversion
		if srcKind == reflect.Slice || srcKind == reflect.Array {
			resValRef := reflect.MakeSlice(reflectType, reflectValue.Len(), reflectValue.Len())
			reflectValueRef := reflect.ValueOf(reflectValue.Interface())
			for i := 0; i < reflectValueRef.Len(); i++ {
				val := reflectValueRef.Index(i)
				err := v.AutoConvertReflectValueByType(&val, reflectType)
				if err != nil {
					return err
				}
				resValRef.Index(i).Set(val)
			}
			*reflectValue = resValRef
			return nil
		}
	default:
		if targetKind == srcKind || convertible(srcKind, targetKind) {
			*reflectValue = reflectValue.Convert(reflectType)
			return nil
		}
	}
	// 2022.9.12 Added some automatic type conversion devices!
	//    should not be converted. 1. If []byte/[]uint8, the input is string and can be automatically converted to
	//    2. If required is string, input is []byte / []uint8
	if srcKind == reflect.String &&
		targetKind == reflect.Slice && reflectType.Elem().Kind() == reflect.Uint8 {
		strValue, ok := reflectValue.Interface().(string)
		if ok {
			*reflectValue = reflect.ValueOf([]byte(strValue))
			return nil
		}
	}
	if srcKind == reflect.Slice &&
		targetKind == reflect.String && (reflectValue.Type().Elem().Kind() == reflect.Uint8) {
		strValue, ok := reflectValue.Interface().([]byte)
		if ok {
			*reflectValue = reflect.ValueOf(string(strValue))
			return nil
		}
	}

	err := fmt.Errorf("invalid argument type: require `%v`, but we got `%v`", reflectType, tin)
	if strings.HasSuffix(fmt.Sprint(tin), "spec.undefinedType") {
		err = fmt.Errorf("%v\n  Maybe u forgot to define variable?", err)
	}
	return err
}

func convertible(kind, tkind reflect.Kind) bool {
	if tkind >= reflect.Int && tkind <= reflect.Uintptr {
		return kind >= reflect.Int && kind <= reflect.Uintptr
	}
	if tkind == reflect.Float64 || tkind == reflect.Float32 {
		return kind >= reflect.Int && kind <= reflect.Float64
	}
	return false
}
