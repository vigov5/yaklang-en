package yakvm

import (
	"fmt"
	"math"
	"reflect"
	"runtime"

	"github.com/yaklang/yaklang/common/utils"
)

func (v *Value) NativeAsyncCall(vm *Frame, wavy bool, vs ...*Value) interface{} {
	v.nativeCall(true, wavy, vm, vs...)
	return nil
}

func (v *Value) NativeCall(vm *Frame, wavy bool, vs ...*Value) interface{} {
	return v.nativeCall(false, wavy, vm, vs...)
}

func (v *Value) GetNativeCallFunctionName() string {
	if v == nil {
		return ""
	}

	if v.Value != nil && v.NativeCallable() {
		funcIns := runtime.FuncForPC(reflect.ValueOf(v.Value).Pointer())
		funcName := funcIns.Name()
		return funcName
	}
	return ""
}

func (v *Value) nativeCall(asyncCall, wavy bool, vm *Frame, vs ...*Value) interface{} {
	rets := reflect.ValueOf(v.Value)
	funcType := rets.Type()
	funcName := v.Literal
	if funcName == "" {
		funcName = v.GetNativeCallFunctionName()
	}
	// . This is very imperfect. , a lot of compatibility processing needs to be done.
	numin := funcType.NumIn()
	if funcType.IsVariadic() {
		numin = len(vs)
	}
	args := make([]reflect.Value, numin)
	for i := 0; i < numin; i++ {
		var val interface{}
		if i < len(vs) && vs[i] != nil {
			val = vs[i].Value
		}
		args[i] = reflect.ValueOf(val)
	}
	if funcType.IsVariadic() {
		// Whether it is a variable parameter,
		numInMax := funcType.NumIn()
		if len(args) < numInMax-1 {
			// . Why is it reduced by one here?
			// Because the last parameter of the variable parameter is slice, it can be 0.
			panic("variadic params need at lease params length: " + fmt.Sprint(numInMax-1))
		}

		// takes out the element type
		variadicParamsType := funcType.In(numInMax - 1).Elem()

		for i := 0; i < len(args); i++ {
			argVal := args[i]
			var targetType reflect.Type
			if i >= numInMax-1 {
				targetType = variadicParamsType
			} else {
				targetType = funcType.In(i)
			}
			err := vm.AutoConvertReflectValueByType(&argVal, targetType)
			if err != nil {
				msg := fmt.Sprintf(
					"native func `%s` calling failed: auto convert failed, cannot convert %v(passed) to %v(need)", funcName,
					args[i].Type(), funcType.In(i),
				)
				panic(msg)
			}
			args[i] = argVal
		}
	} else {
		// If the parameter is immutable, the length of the input function parameter list and the required parameter list length should be equal.
		if vm.vm.GetConfig().GetFunctionNumberCheck() && funcType.NumIn() != len(vs) {
			msg := fmt.Sprintf("native func `%s` need [%v] params, actually got [%v] params", funcName, funcType.NumIn(), len(vs))
			panic(msg)
		}
		for i := 0; i < funcType.NumIn(); i++ {
			argVal := args[i]
			err := vm.AutoConvertReflectValueByType(&argVal, funcType.In(i))
			if err != nil {
				msg := fmt.Sprintf(
					"native func `%s` calling failed: auto convert failed, cannot convert %v(passed) to %v(need)", funcName,
					args[i].Type(), funcType.In(i),
				)
				panic(msg)
			}
			args[i] = argVal
			v := argVal.Interface()
			_ = v
		}
	}
	// debug io
	//for _, a := range args {
	//	println(a.Type().String())
	//}
	if asyncCall {
		go func() {
			defer vm.vm.AsyncEnd()
			rets.Call(args)
		}()
		return nil
	}
	returns := rets.Call(args)
	vals := make([]interface{}, len(returns))
	for i, ret := range returns {
		// proves to be an alias. For example, time.Duration is an alias of the int64 type, but it has its own implementation method, so
		pkgPath := ret.Type().PkgPath()
		if pkgPath != "" {
			vals[i] = ret.Interface()
			continue
		}

		switch {
		case ret.Kind() >= reflect.Int && ret.Kind() <= reflect.Int64:
			if ret.Int() > math.MaxInt {
				vals[i] = ret.Int()
			} else {
				vals[i] = int(ret.Int())
			}
		case ret.Kind() >= reflect.Uint && ret.Kind() <= reflect.Uintptr:
			if ret.Uint() > math.MaxInt {
				vals[i] = int64(ret.Convert(literalReflectType_Uint).Uint())
			} else {
				vals[i] = int(ret.Convert(literalReflectType_Int).Int())
			}
		case ret.Kind() == reflect.Float32:
			vals[i] = ret.Float()
		default:
			vals[i] = ret.Interface()
		}
	}

	if wavy && len(vals) > 1 {
		lastValue := vals[len(vals)-1]
		if err, ok := lastValue.(error); ok || lastValue == nil {
			vals = vals[:len(vals)-1]
			if err != nil {
				panic(utils.Errorf("native func `%s` call error: %v", funcName, err))
			}
		}
	}

	if len(vals) == 1 {
		return vals[0]
	}
	return vals
}

func (v *Value) YakFunctionNCall(vm *Frame, vs ...*Value) interface{} {
	return v.yakFunctionNCall(false, vm, vs...)
}

func (v *Value) LuaFunctionNCall(vm *Frame, vs ...*Value) interface{} {
	return v.luaFunctionNCall(false, vm, vs...)
}

func (v *Value) YakFunctionNAsyncCall(vm *Frame, vs ...*Value) {
	v.yakFunctionNCall(true, vm, vs...)
}

func (v *Value) yakFunctionNCall(asyncCall bool, vm *Frame, vs ...*Value) interface{} {
	//args := make([]reflect.Value, len(vs))
	//for i, v := range vs {
	//	args[i] = reflect.ValueOf(v.Value)
	//}

	f, ok := v.Value.(*Function)
	if !ok {
		panic("BUG: yak function type assert failed")
	}
	return vm.CallYakFunction(asyncCall, f, vs)
}

func (v *Value) luaFunctionNCall(asyncCall bool, vm *Frame, vs ...*Value) interface{} {
	f, ok := v.Value.(*Function)
	if !ok {
		panic("BUG: lua function type assert failed")
	}
	return vm.CallLuaFunction(asyncCall, f, vs)
}

func (v *Value) Call(vm *Frame, wavy bool, vs ...*Value) interface{} {
	// natively calls the function
	if v.NativeCallable() {
		return v.NativeCall(vm, wavy, vs...)
	} else {
		return v.YakFunctionNCall(vm, vs...)
	}
}

func (v *Value) CallLua(vm *Frame, vs ...*Value) interface{} {
	// natively calls the function
	if v.NativeCallable() {
		return v.NativeCall(vm, false, vs...)
	} else {
		return v.LuaFunctionNCall(vm, vs...)
	}
}

func (v *Value) AsyncCall(vm *Frame, wavy bool, vs ...*Value) {
	if v.NativeCallable() {
		v.NativeAsyncCall(vm, wavy, vs...)
	} else {
		v.YakFunctionNAsyncCall(vm, vs...)
	}
}
