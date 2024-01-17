package antlr4Lua

import (
	"context"
	"fmt"
	"github.com/yaklang/yaklang/common/yak/antlr4Lua/luaast"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
)

func init() {
	//os.Setenv("LUA_DEBUG", "1")
	// . Using yakvm.Import is the most original import method of yaklang. There are two types of import. One is to import
	// first. The other is to import eval and other built-in functions
	// . Here, some commonly used Lua built-in functions are added to facilitate testing. The key of
	// http://www.lua.org/manual/5.3/manual.html

	// print (·· ·)
	//Receives any number of arguments and prints their values to stdout, using the
	//tostring function to convert each argument to a string. print is not intended
	//for formatted output, but only as a quick way to show a value, for instance for
	//debugging. For complete control over the output, use string.format and io.write.
	// native Lua print Tabs will be added between multiple parameters. There is no big problem here. By default, go uses ws
	Import("print", func(v ...interface{}) {
		toStr := func(x interface{}) (string, bool) {
			switch v := x.(type) {
			case string:
				return v, true
			case int:
				return strconv.Itoa(v), true
			case float64:
				if v == float64(int64(v)) {
					return fmt.Sprintf("%.1f", v), true
				}
				return fmt.Sprintf("%.14g", v), true
			case float32:
				if v == float32(int64(v)) {
					return fmt.Sprintf("%.1f", v), true
				}
				return fmt.Sprintf("%.14g", v), true
			}
			return "", false
		}
		for index, value := range v {
			formattedVal, ok := toStr(value)
			if ok {
				v[index] = formattedVal
			}
		}
		fmt.Println(v)
	})

	Import("raw_print", func(v interface{}) {
		fmt.Println(v)
	})

	//assert (v [, message])
	//Calls error if the value of its argument v is false (i.e., nil or false);
	//otherwise, returns all its arguments. In case of error, message is the
	//error object; when absent, it defaults to "assertion failed!"

	//todo
	//https://www.lua.org/pil/8.3.html
	Import("assert", func(condition ...interface{}) {
		if condition == nil {
			panic("assert nil")
		}
		if len(condition) == 2 {
			Assert := func(condition interface{}, message string) {
				if condition == nil {
					panic(message)
				}
				if boolean, ok := condition.(bool); ok {
					if !boolean {
						panic(message)
					}
				}
			}
			Assert(condition[0], condition[1].(string))
		} else {
			Assert := func(condition interface{}) {
				if boolean, ok := condition.(bool); ok {
					if !boolean {
						panic("assert failed")
					}
				}
			}
			Assert(condition)
		}
	})

	Import("@pow", func(x interface{}, y interface{}) float64 {
		interfaceToFloat64 := func(a interface{}) (float64, bool) {
			switch v := a.(type) {
			case float64:
				return v, true
			case int:
				return float64(v), true
			case int64:
				return float64(v), true
			}
			return 0, false
		}
		index, ok1 := interfaceToFloat64(x)
		base, ok2 := interfaceToFloat64(y)
		if ok1 && ok2 {
			return math.Pow(base, index)
		} else {
			panic("attempt to pow a '" + reflect.TypeOf(base).String() + "' with a '" + reflect.TypeOf(index).String() + "'")
		}

	})

	Import("@floor", func(x interface{}, y interface{}) float64 {
		interfaceToFloat64 := func(a interface{}) (float64, bool) {
			switch v := a.(type) {
			case float64:
				return v, true
			case int:
				return float64(v), true
			case int64:
				return float64(v), true
			}
			return 0, false
		}
		base, ok1 := interfaceToFloat64(x)
		index, ok2 := interfaceToFloat64(y)
		if ok1 && ok2 && (index != 0) {
			res := base / index
			return math.Floor(res)
		} else if (index == 0) && ok1 && ok2 {
			panic("dividend can't be zero!")
		} else {
			panic("attempt to floor a '" + reflect.TypeOf(base).String() + "' with a '" + reflect.TypeOf(index).String() + "'")
		}

	})

	Import("tostring", func(x interface{}) string {
		switch v := x.(type) {
		case string:
			return v
		case int:
			return strconv.Itoa(v)
		case float64:
			if v == float64(int64(v)) {
				return fmt.Sprintf("%.1f", v)
			}
			return fmt.Sprintf("%.14g", v)
		case float32:
			if v == float32(int64(v)) {
				return fmt.Sprintf("%.1f", v)
			}
			return fmt.Sprintf("%.14g", v)
		default:
			panic(fmt.Sprintf("tostring() cannot convert %v", reflect.TypeOf(x).String()))

		}
	})

	Import("@strcat", func(x interface{}, y interface{}) string {
		defer func() {
			if recover() != nil {
				panic(fmt.Sprintf("attempt to concatenate %v with %v", reflect.TypeOf(x).String(), reflect.TypeOf(y).String()))
			}
		}()
		toStr := func(x interface{}) string {
			switch v := x.(type) {
			case string:
				return v
			case int:
				return strconv.Itoa(v)
			case float64:
				return fmt.Sprintf("%.14g", v)
			case float32:
				return fmt.Sprintf("%.14g", v)
			default:
				panic(fmt.Sprintf("tostring() cannot convert %v", x))

			}
		}
		return toStr(x) + toStr(y)
	})

	Import("@getlen", func(x interface{}) int {
		if str, ok := x.(string); ok {
			return len(str)
		}
		rk := reflect.TypeOf(x).Kind()
		if rk == reflect.Map {
			valueOfInputMap := reflect.ValueOf(x)
			if reflect.TypeOf(x).Key().Kind() != reflect.Int && reflect.TypeOf(x).Key().Kind() != reflect.Interface {
				return 0
			}
			mapLen := valueOfInputMap.Len()
			tblLen := 0

			for index := 1; index <= mapLen; index++ {
				value := valueOfInputMap.MapIndex(reflect.ValueOf(index))
				// When using reflect to access map elements, you need to pay attention to check whether the value returned by MapIndex is available,
				// because if the key corresponding to the index does not exist, or the value corresponding to the key is nil, then the MapIndex method will return an invalid Value,
				// through script_engine.go in a form similar to the built-in dependent library. It cannot be used directly. You need to check
				if value.IsValid() && !value.IsZero() {
					tblLen++
				} else {
					return tblLen
				}
			}
			return tblLen
		}
		panic(fmt.Sprintf("attempt to get length of %v", reflect.TypeOf(x).String()))
	})

	Import("next", func(x ...interface{}) (interface{}, interface{}) { // next(table[,index])
		if len(x) == 1 || x[1] == nil {
			keysString := make([]string, 0)
			keysMap := make(map[string]interface{})
			valueOfInputMap := reflect.ValueOf(x[0])
			iter := valueOfInputMap.MapRange()
			for iter.Next() {
				keyStr := yakvm.NewAutoValue(iter.Key().Interface()).String()
				keysString = append(keysString, keyStr)
				keysMap[keyStr] = iter.Key().Interface()
			}
			sort.Strings(keysString)
			return keysMap[keysString[0]], valueOfInputMap.MapIndex(reflect.ValueOf(keysMap[keysString[0]])).Interface()
		} else {
			keysString := make([]string, 0)
			keysMap := make(map[string]interface{})
			valueOfInputMap := reflect.ValueOf(x[0])
			indexToNext := reflect.ValueOf(x[1]).String()
			iter := valueOfInputMap.MapRange()
			for iter.Next() {
				keyStr := yakvm.NewAutoValue(iter.Key().Interface()).String()
				keysString = append(keysString, keyStr)
				keysMap[keyStr] = iter.Key().Interface()
			}
			sort.Strings(keysString)
			// map is the only
			for index, value := range keysString {
				if value == indexToNext {
					if index+1 == len(keysString) {
						return nil, nil
					} else {
						return keysMap[keysString[index+1]], valueOfInputMap.MapIndex(reflect.ValueOf(keysMap[keysString[index+1]])).Interface()
					}
				}
			}
			panic("`invalid key to 'next'`")
		}
		return nil, nil
	})

	Import("error", func(x any) {
		panic(x)
	})

}

var buildinLib = make(map[string]interface{})

func Import(name string, f interface{}) {
	buildinLib[name] = f
}

type LuaSnippetExecutor struct {
	sourceCode string
	engine     *Engine
	translator *luaast.LuaTranslator
}

func NewLuaSnippetExecutor(code string) *LuaSnippetExecutor {
	e := New()
	e.ImportLibs(buildinLib)
	return &LuaSnippetExecutor{sourceCode: code, engine: e, translator: &luaast.LuaTranslator{}}
}

func (l *LuaSnippetExecutor) Run() {
	err := l.engine.Eval(context.Background(), l.sourceCode)
	if err != nil {
		panic(fmt.Sprintf("\n==============\n%s\n==============\n", err.Error()))
	}
}

func (l *LuaSnippetExecutor) Debug() {
	l.engine.debug = true
	err := l.engine.Eval(context.Background(), l.sourceCode)
	if err != nil {
		panic(fmt.Sprintf("\n==============\n%s\n==============\n", err.Error()))
	}
}

// SmartRun SmartRun() will choose Run() or Debug() depending on the environment setting `LUA_DEBUG`
func (l *LuaSnippetExecutor) SmartRun() {
	if os.Getenv("LUA_DEBUG") != "" {
		l.Debug()
	} else {
		l.Run()
	}
}
