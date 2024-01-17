package builtin

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/yaklang"
	yaklangspec "github.com/yaklang/yaklang/common/yak/yaklang/spec"
	"github.com/yaklang/yaklang/common/yak/yaklang/spec/types"

	"github.com/davecgh/go-spew/spew"
)

// -----------------------------------------------------------------------------

// panic Crash and print error message
// Example:
// ```
// panic("something happened")
// ```
func Panic(v interface{}) {
	panic(v)
}

// panicf Crash and print error message according to format
// Example:
// ```
// panicf("something happened: %v", err)
// ```
func Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

// -----------------------------------------------------------------------------

var zeroVal reflect.Value

// mkmap Creates a map of the specified type (map)
// ! Deprecated, use make statement instead
// Example:
// ```
// m = mkmap("string:var") // map[string]any
// ```
func Mkmap(typ interface{}, n ...int) interface{} {
	if len(n) > 0 {
		return reflect.MakeMapWithSize(types.Reflect(typ), n[0]).Interface()
	}
	return reflect.MakeMap(types.Reflect(typ)).Interface()
}

// mapOf returns the mapping type of the specified type, which can be used in mkmap
// ! Deprecated, use make statement instead
// Example:
// ```
// m = mkmap(mapOf("string", "var")) // map[string]any
// ```
func MapOf(key, val interface{}) interface{} {
	return reflect.MapOf(types.Reflect(key), types.Reflect(val))
}

// mapFrom based on the incoming key-value pair Initialize mapping (map)
// ! Deprecated, you can use the map initialization statement instead
// Example:
// ```
// m = mapFrom("a", 1, "b", 2) // {"a": 1, "b": 2}
// ```
func MapFrom(args ...interface{}) interface{} {
	n := len(args)
	if (n & 1) != 0 {
		panic("please use `mapFrom(key1, val1, key2, val2, ...)`")
	}
	if n == 0 {
		return make(map[string]interface{})
	}

	firstKey := kindOf2Args(args, 0)
	switch firstKey {
	case reflect.String:
		switch kindOf2Args(args, 1) {
		case reflect.String:
			ret := make(map[string]string, n>>1)
			for i := 0; i < n; i += 2 {
				ret[args[i].(string)] = args[i+1].(string)
			}
			return ret
		case reflect.Int:
			ret := make(map[string]int, n>>1)
			for i := 0; i < n; i += 2 {
				ret[args[i].(string)] = asInt(args[i+1])
			}
			return ret
		case reflect.Float64:
			ret := make(map[string]float64, n>>1)
			for i := 0; i < n; i += 2 {
				ret[args[i].(string)] = asFloat(args[i+1])
			}
			return ret
		default:
			ret := make(map[string]interface{}, n>>1)
			for i := 0; i < n; i += 2 {
				if t := args[i+1]; t != yaklangspec.Undefined {
					ret[args[i].(string)] = t
				}
			}
			return ret
		}
	case reflect.Int:
		switch kindOf2Args(args, 1) {
		case reflect.String:
			ret := make(map[int]string, n>>1)
			for i := 0; i < n; i += 2 {
				ret[asInt(args[i])] = args[i+1].(string)
			}
			return ret
		case reflect.Int:
			ret := make(map[int]int, n>>1)
			for i := 0; i < n; i += 2 {
				ret[asInt(args[i])] = asInt(args[i+1])
			}
			return ret
		case reflect.Float64:
			ret := make(map[int]float64, n>>1)
			for i := 0; i < n; i += 2 {
				ret[asInt(args[i])] = asFloat(args[i+1])
			}
			return ret
		default:
			ret := make(map[int]interface{}, n>>1)
			for i := 0; i < n; i += 2 {
				if t := args[i+1]; t != yaklangspec.Undefined {
					ret[asInt(args[i])] = t
				}
			}
			return ret
		}
	case reflect.Invalid:
		_, value := valueInterfaceOf2Args(args, 0)
		panic(fmt.Sprintf("use `{key: value, key2: value2}` or `mapFrom` to create map failed:\n"+
			"    mapFrom: type of key should be `string|int` all, unexpected invalid(%v)", strings.TrimSpace(spew.Sdump(value.Interface()))))
	default:
		panic(fmt.Sprintf("use `{key: value, key2: value2}` or `mapFrom` to create map failed:\n"+
			"    mapFrom: type of key should be `string|int`, not %v", firstKey.String()))
	}
}

// delete deletes the key from the map
// Example:
// ```
// m = {"a": 1, "b": 2}
// delete(m, "b")
// println(m) // {"a": 1}
// ```
func Delete(m interface{}, key interface{}) {
	globalMapLock.Lock()
	defer globalMapLock.Unlock()
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("delete map failed: %v", err)
			return
		}
	}()
	reflect.ValueOf(m).SetMapIndex(reflect.ValueOf(key), zeroVal)
}

// set sets the value of an object. The object Can be an array, slice, map or structure (struct) or structure reference (ptr)
// ! Deprecated, you can use an initialization statement or assignment statement instead
// Example:
// ```
// a = make([]int, 3)
// set(a, 0, 100, 1, 200, 2, 300) // a = [100, 200, 300]
// ```
func Set(m interface{}, args ...interface{}) {
	n := len(args)
	if (n & 1) != 0 {
		panic("call with invalid argument count: please use `set(obj, member1, val1, ...)")
	}

	o := reflect.ValueOf(m)
	switch o.Kind() {
	case reflect.Slice, reflect.Array:
		telem := reflect.TypeOf(m).Elem()
		for i := 0; i < n; i += 2 {
			val := autoConvert(telem, args[i+1])
			o.Index(args[i].(int)).Set(val)
		}
	case reflect.Map:
		setMapMember(o, args...)
	default:
		setMember(m, args...)
	}
}

// SetIndex sets a (index, value) pair to an object. object can be a slice, an array, a map or a user-defined class.
func SetIndex(m, key, v interface{}) {
	o := reflect.ValueOf(m)
	switch o.Kind() {
	case reflect.Map:
		var val reflect.Value
		if v == yaklangspec.Undefined {
			val = zeroVal
		} else {
			val = autoConvert(o.Type().Elem(), v)
		}
		globalMapLock.Lock()
		defer globalMapLock.Unlock()
		o.SetMapIndex(reflect.ValueOf(key), val)
	case reflect.Slice, reflect.Array:
		if idx, ok := key.(int); ok {
			o.Index(idx).Set(reflect.ValueOf(v))
		} else {
			panic("slice index isn't an integer value")
		}
	default:
		setMember(m, key, v)
	}
}

type varSetter interface {
	SetVar(name string, val interface{})
}

func setMember(m interface{}, args ...interface{}) {
	if v, ok := m.(varSetter); ok {
		for i := 0; i < len(args); i += 2 {
			v.SetVar(args[i].(string), args[i+1])
		}
		return
	}

	o := reflect.ValueOf(m)
	if o.Kind() == reflect.Ptr {
		o = o.Elem()
		if o.Kind() == reflect.Struct {
			setStructMember(o, args...)
			return
		}
	}
	panic(fmt.Sprintf("type `%v` doesn't support `set` operator", reflect.TypeOf(m)))
}

func setStructMember(o reflect.Value, args ...interface{}) {
	for i := 0; i < len(args); i += 2 {
		key := args[i].(string)
		field := o.FieldByName(strings.Title(key))
		if !field.IsValid() {
			panic(fmt.Sprintf("struct `%v` doesn't has member `%v`", o.Type(), key))
		}
		field.Set(reflect.ValueOf(args[i+1]))
	}
}

func setMapMember(o reflect.Value, args ...interface{}) {
	globalMapLock.Lock()
	defer globalMapLock.Unlock()

	var val reflect.Value
	telem := o.Type().Elem()
	for i := 0; i < len(args); i += 2 {
		key := reflect.ValueOf(args[i])
		t := args[i+1]
		if t == yaklangspec.Undefined {
			val = zeroVal
		} else {
			val = autoConvert(telem, t)
		}
		o.SetMapIndex(key, val)
	}
}

// close is used to close the already opened Channel. Closing a closed channel will cause a runtime crash.
// Example:
// ```
// ch = make(chan int)
// go func() {
// for i = range ch {
// println(i)
// }
// }()
// ch <- 1
// ch <- 2
// close(ch)
// ```
func CloseChan(v interface{}) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Chan {
		panic("close only for channel")
	}
	reflect.ValueOf(v).Close()
}

// get Get the key value from the map, if the key does not exist, return the default value
// Example:
// ```
// m = {"a": 1, "b": 2}
// get(m, "a") // 1
// get(m, "c", "default") // "default"
// ```
func Get(m interface{}, key interface{}, defaultValues ...interface{}) (result interface{}) {
	nilValue := yaklangspec.Undefined
	if yaklang.IsNew() {
		nilValue = nil
	}
	if len(defaultValues) > 0 {
		defer func() {
			if result == nilValue {
				result = defaultValues[0]
			}
		}()
	}

	o := reflect.ValueOf(m)
	switch o.Kind() {
	case reflect.Map:
		v := o.MapIndex(reflect.ValueOf(key))
		if v.IsValid() {
			return v.Interface()
		}
		return nilValue
	case reflect.Slice, reflect.String, reflect.Array:
		indexInt, ok := key.(int)
		if !ok {
			panic(fmt.Sprintf("slice/string/array 's key(%v type: %v) is not `int` type", key, reflect.TypeOf(key)))
		}
		if o.Len() > indexInt {
			return o.Index(key.(int)).Interface()
		} else {
			panic(fmt.Sprintf("length of %v(%v) is %v, but index [%v] is too large", o.Kind().String(), o.Type().String(), o.Len(), indexInt))
		}
	case reflect.Int: // undefined?
		return nilValue
	default:
		return yaklangspec.GetEx(m, key)
	}
}

// GetVar returns a member variable of an object. object can be a slice, an array, a map or a user-defined class.
func GetVar(m interface{}, key interface{}) interface{} {
	return &yaklangspec.DataIndex{Data: m, Index: key}
}

// len returns the length of the collection object. The object can be an array, a slice, a map, a string or a channel.
// Example:
// ```
// a = [1, 2, 3]
// println(len(a)) // 3
// ```
func Len(a interface{}) int {
	if a == nil {
		return 0
	}
	if ch, ok := a.(*yaklangspec.Chan); ok {
		return ch.Data.Len()
	}
	return reflect.ValueOf(a).Len()
}

// cap returns the capacity of the collection object. The object can be an array, a slice or a channel.
// Example:
// ```
// a = make([]int, 0, 3)
// println(cap(a)) // 3
// ```
func Cap(a interface{}) int {
	if a == nil {
		return 0
	}
	if ch, ok := a.(*yaklangspec.Chan); ok {
		return ch.Data.Cap()
	}
	return reflect.ValueOf(a).Cap()
}

// sub returns the array or a subslice of the slice
// ! Deprecated, you can use slicing statements instead
// Example:
// ```
// a = [1, 2, 3, 4, 5]
// b = sub(a, 1, 3) // [2, 3] is equivalent to a[1:3]
// ```
func SubSlice(a, i, j interface{}) interface{} {
	va := reflect.ValueOf(a)
	var i1, j1 int
	if i != nil {
		i1 = asInt(i)
	}
	if j != nil {
		j1 = asInt(j)
	} else {
		j1 = va.Len()
	}
	return va.Slice(i1, j1).Interface()
}

// copy copies the src array/copies the slice to the dst array/slice, and return the number of copied elements
// Example:
// ```
// a = [1, 2, 3]
// b = make([]int, 3)
// copy(b, a)
// println(b) // [1 2 3]
// ```
func Copy(dst, src interface{}) int {
	return reflect.Copy(reflect.ValueOf(dst), reflect.ValueOf(src))
}

// append Appends elements to an array or slice and returns the result
// Example:
// ```
// a = [1, 2, 3]
// a = append(a, 4, 5, 6)
// println(a) // [1 2 3 4 5 6]
// ```
func Append(a interface{}, vals ...interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			// log.Error(err)
			results := funk.Map(a, func(e interface{}) interface{} {
				return e
			}).([]interface{})
			results = append(results, vals...)
			ret = results
		}
	}()

	switch arr := a.(type) {
	case []int:
		return appendInts(arr, vals...)
	case []interface{}:
		return append(arr, vals...)
	case []string:
		return appendStrings(arr, vals...)
	case []byte:
		return appendBytes(arr, vals...)
	case []float64:
		return appendFloats(arr, vals...)
	}

	va := reflect.ValueOf(a)
	telem := va.Type().Elem()
	x := make([]reflect.Value, len(vals))
	for i, v := range vals {
		x[i] = autoConvert(telem, v)
	}
	return reflect.Append(va, x...).Interface()
}

func autoConvert(telem reflect.Type, v interface{}) reflect.Value {
	if v == nil {
		return reflect.Zero(telem)
	}

	val := reflect.ValueOf(v)
	if telem != reflect.TypeOf(v) {
		val = yaklangspec.AutoConvert(val, telem)
	}
	return val
}

func appendFloats(a []float64, vals ...interface{}) interface{} {
	for _, v := range vals {
		switch val := v.(type) {
		case float64:
			a = append(a, val)
		case int:
			a = append(a, float64(val))
		case float32:
			a = append(a, float64(val))
		default:
			panic("unsupported: []float64 append " + reflect.TypeOf(v).String())
		}
	}
	return a
}

func appendInts(a []int, vals ...interface{}) interface{} {
	for _, v := range vals {
		switch val := v.(type) {
		case int:
			a = append(a, val)
		default:
			panic("unsupported: []int append " + reflect.TypeOf(v).String())
		}
	}
	return a
}

func appendBytes(a []byte, vals ...interface{}) interface{} {
	for _, v := range vals {
		switch val := v.(type) {
		case byte:
			a = append(a, val)
		case int:
			a = append(a, byte(val))
		default:
			panic("unsupported: []byte append " + reflect.TypeOf(v).String())
		}
	}
	return a
}

func appendStrings(a []string, vals ...interface{}) interface{} {
	for _, v := range vals {
		switch val := v.(type) {
		case string:
			a = append(a, val)
		default:
			panic("unsupported: []string append " + reflect.TypeOf(v).String())
		}
	}
	return a
}

// mkslice Creates the specified type Slice
// ! Deprecated, use make statement instead
// Example:
// ```
// a = mkslice("var") // []any
// ```
func Mkslice(typ interface{}, args ...interface{}) interface{} {
	n, cap := 0, 0
	if len(args) == 1 {
		if v, ok := args[0].(int); ok {
			n, cap = v, v
		} else {
			panic("second param type of func `slice` must be `int`")
		}
	} else if len(args) > 1 {
		if v, ok := args[0].(int); ok {
			n = v
		} else {
			panic("2nd param type of func `slice` must be `int`")
		}
		if v, ok := args[1].(int); ok {
			cap = v
		} else {
			panic("3rd param type of func `slice` must be `int`")
		}
	}
	typSlice := reflect.SliceOf(types.Reflect(typ))
	return reflect.MakeSlice(typSlice, n, cap).Interface()
}

// sliceFrom initializes the slice (slice) according to the incoming key-value pair
// ! Deprecated, you can use the slice initialization statement instead of
// Example:
// ```
// a = sliceFrom(1, 2, 3) // [1, 2, 3]
// ```
func SliceFrom(args ...interface{}) interface{} {
	n := len(args)
	if n == 0 {
		return []interface{}(nil)
	}

	switch kindOfArgs(args) {
	case reflect.Int:
		return appendInts(make([]int, 0, n), args...)
	case reflect.Float64:
		return appendFloats(make([]float64, 0, n), args...)
	case reflect.String:
		return appendStrings(make([]string, 0, n), args...)
	case reflect.Uint8:
		return appendBytes(make([]byte, 0, n), args...)
	default:
		return append(make([]interface{}, 0, n), args...)
	}
}

// SliceFromTy creates a slice from `[]T{a1, a2, ...}`.
func SliceFromTy(args ...interface{}) interface{} {
	got, ok := args[0].(yaklangspec.GoTyper)
	if !ok {
		panic(fmt.Sprintf("`%v` is not a yaklang type", args[0]))
	}
	t := got.GoType()
	n := len(args)
	ret := reflect.MakeSlice(reflect.SliceOf(t), 0, n-1).Interface()
	return Append(ret, args[1:]...)
}

// sliceOf Returns the slice type of the specified type, which can be used in mkslice
// ! Deprecated, use make statement instead
// Example:
// ```
// m = mkslice(sliceOf("var")) // []any
// ```
func SliceOf(typ interface{}) interface{} {
	return reflect.SliceOf(types.Reflect(typ))
}

// StructInit creates a struct object from `structInit(structType, member1, val1, ...)`.
func StructInit(args ...interface{}) interface{} {
	if (len(args) & 1) != 1 {
		panic("call with invalid argument count: please use `structInit(structType, member1, val1, ...)")
	}

	got, ok := args[0].(yaklangspec.GoTyper)
	if !ok {
		panic(fmt.Sprintf("`%v` is not a yaklang type", args[0]))
	}
	t := got.GoType()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("`%v` is not a struct type", args[0]))
	}
	ret := reflect.New(t)
	setStructMember(ret.Elem(), args[1:]...)
	return ret.Interface()
}

// MapInit creates a map object from `mapInit(mapType, member1, val1, ...)`.
func MapInit(args ...interface{}) interface{} {
	if (len(args) & 1) != 1 {
		panic("call with invalid argument count: please use `mapInit(mapType, member1, val1, ...)")
	}

	got, ok := args[0].(yaklangspec.GoTyper)
	if !ok {
		panic(fmt.Sprintf("`%v` is not a yaklang type", args[0]))
	}
	t := got.GoType()
	if t.Kind() != reflect.Map {
		panic(fmt.Sprintf("`%v` is not a map type", args[0]))
	}
	ret := reflect.MakeMap(t)
	setMapMember(ret, args[1:]...)
	return ret.Interface()
}

// print on standard output Format and print the information using the default format in
// Example:
// ```
// print("hello yak")
// print("hello", 1, "2", [1, 2, 3])
// ```
func print(a ...any) (n int, err error) {
	return fmt.Print(a...)
}

// printf Format and print information in standard output according to the format specifier
// Example:
// ```
// printf("hello %s", "yak")
// printf("value = %v", value)
// ```
func printf(format string, a ...any) (n int, err error) {
	return fmt.Printf(format, a...)
}

// println Use default format to format and print message on standard output (including newlines) )
// Example:
// ```
// println("hello world")
// println("hello yak")
// ```
func println(a ...any) (n int, err error) {
	return fmt.Println(a...)
}

// sprint uses the default format to format and returns a string
// Example:
// ```
// sprint({"a": 1, "b": 2}, 1, [1, 2, 3])
// ```
func sprint(a ...any) string {
	return fmt.Sprint(a...)
}

// sprintf formats any number of parameters according to the format specifier and returns the string
// Example:
// ```
// sprintf("%v %d %v", {"a": 1, "b": 2}, 1, [1, 2, 3])
// ```
func sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// sprintln Use the default format for formatting and returns a string (including newlines)
// Example:
// ```
// sprintln({"a": 1, "b": 2}, 1, [1, 2, 3])
// ```
func sprintln(a ...any) string {
	return fmt.Sprintln(a...)
}

// fprint uses the default format to format any number of parameters and writes w
// Example:
// ```
// fprint(os.Stderr, "error")
// ```
func fprint(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprint(w, a...)
}

// fprintf formats any number of parameters according to the format specifier. And write w
// Example:
// ```
// fprintf(os.Stderr, "value = %v", value)
// ```
func fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}

// fprintln Use the default format Format any number of parameters and write w (including newlines)
// Example:
// ```
// fprintln(os.Stderr, "error")
// ```
func fprintln(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

// -----------------------------------------------------------------------------
