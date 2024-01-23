package fuzztagx

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/utils"
	"strconv"
	"strings"
	"testing"
)

var testMap = map[string]func(string) []string{
	"echo": func(i string) []string {
		return []string{i}
	},
	"array": func(i string) []string {
		return strings.Split(i, "|")
	},
	"get1": func(i string) []string {
		return []string{"1"}
	},
	"list": func(s string) []string {
		return strings.Split(s, "|")
	},
	"int": func(i string) []string {
		return funk.Map(utils.ParseStringToPorts(i), func(i int) string {
			return strconv.Itoa(i)
		}).([]string)
	},
	"panic": func(s string) []string {
		panic(s)
		return nil
	},
}

// . Synchronous rendering quantity test
func TestSyncRender(t *testing.T) {
	for i, testcase := range [][2]any{
		//{
		//	"{{echo::1({{list(aaa|ccc)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
		//	3,
		//},
		//{
		//	"{{echo::1({{list(aaa|ccc|ddd)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
		//	3,
		//},
		//{
		//	"{{echo::1({{list(aaa|ccc|ddd|eee)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
		//	4,
		//},
		{
			"{{echo::3({{list(aaa|ccc|ddd)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
			9,
		},
		{
			"{{echo({{list(aaa|ccc|ddd)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
			9,
		},
		{
			"{{echo({{list(aaa|ccc|ddd)}})}}{{echo({{list(aaa|ccc|ddd)}})}}",
			9,
		},
	} {
		result, err := ExecuteWithStringHandler(testcase[0].(string), testMap)
		if err != nil {
			panic(err)
		}
		if len(result) != testcase[1].(int) {
			t.Fatal(utils.Errorf("testcase %d error,got: length %d, expect length: %d", i, len(result), testcase[1].(int)))
		}
	}
	result, err := ExecuteWithStringHandler("{{echo::1({{array::1(a|b)}})}}", testMap)
	if err != nil {
		panic(err)
	}
	if result[0] != "a" || result[1] != "b" {
		panic("test sync render error")
	}
}

// Malformation test
func TestDeformityTag(t *testing.T) {
	for _, v := range [][]string{
		{"{{echo(${<{{echo(a)}}})}}", "${<a}"},
		{"{{echo({{{echo(a)}}})}}", "{a}"},
		{"{{echo({{{echo(a)}}}})}}", "{a}}"},
		{`{{echo(\{{1{{echo(a)}}}})}}`, "{{1a}}"}, // why
		{"{{get1(1-29)}}", "1"},
		{"{{i$$$$$nt(1-29)}}", "{{i$$$$$nt(1-29)}}"},
		{"{{xx12}}", ""},
		{"{{xx12:}}", ""},
		{"{{xx12:-_}}", ""},
		{"{{xx12:-_[[[[}}", "{{xx12:-_[[[[}}"},
		{"{{xx12:-_[}}[[[}}", "{{xx12:-_[}}[[[}}"},
		{"{{xx12:-_}}[[[[}}", "[[[[}}"},
		{"{{xx12:-_(1)}}[[[[}}", "[[[[}}"},
		{"{{xx12:-_:::::::(2)}}[[[[}}", "[[[[}}"},
		{"{{xx12:-_()}}[[[[}}", "[[[[}}"},
		//{"{{xx12:-_(____)____)}}[[[[}}", "{{xx12:-_(____)____)}}[[[[}}"}, // {{xx12:-_(____)____)}}should be parsed correctly.
		{"{{xx12:-_(____\\)____)}}[[[[}}", "[[[[}}"},
		{"{{xx12:-_(____\\)} }____)}}{[[[[}}", "{[[[[}}"},
		{"{{xx12:-_(____)} }}____)}}[[[[}}", "{{xx12:-_(____)} }}____)}}[[[[}}"},
		{"{{xx12:-_(____\\)} }____)}}{{[[[[}}", "{{[[[[}}"},
		{"{{xx12:-_(____\\)} }____)}}{{1[[[[}}", "{{1[[[[}}"},
		//{"{{xx12:-_(____\\)} }__)__)}}{{1[[[[}}", "{{xx12:-_(____\\)} }__)__)}}{{1[[[[}}"},
		{"{{xx12:-_(____\\)} }__\\)__)}}{{1[[[[}}", "{{1[[[[}}"},
		{"{{{{1[[[[}}", "{{{{1[[[[}}"},
		{"{{{{get1}}{{1[[[[}}", "{{1{{1[[[[}}"},
		{"{{i{{get1}}nt(1-2)}}", ""},
		{"{{", "{{"},
		//{"{{echo(123123\\))}}", "123123)"}, // brackets do not need to be escaped.
		//{"{{print(list{\\())}}", "{{print(list{\\())}}"},
		//{"{{print(list{\\(\\))}}", ""},
		{"{{{echo(123)}}", "{123"},
		// {"{{i{{get1}}n{{get1}}t(1-2)}}", "{{i1nt(1-2)}}"},
	} {
		t, r := v[0], v[1]
		spew.Dump(t)
		result, err := ExecuteWithStringHandler(t, testMap)
		if err != nil {
			panic(err)
		}
		if len(result) <= 0 {
			panic(1)
		}
		if result[0] != r {
			m := fmt.Sprintf("got: %v expect: %v", strconv.Quote(result[0]), strconv.Quote(r))
			panic(m)
		}
	}
}

// Newline after the tag name
func TestNewLineAfterTagName(t *testing.T) {
	var m = map[string]func(string) []string{
		"s": func(s string) []string {
			return []string{s + "a"}
		},
	}

	res, err := ExecuteWithStringHandler(`{{s 
() }}`, m)
	spew.Dump(res)
	if err != nil {
		panic(err)
	}
	if len(res) < 1 || res[0] != "a" {
		panic("exec with new line error")
	}
}

func TestExecuteBug1(t *testing.T) {
	var m = map[string]func(string) []string{
		"int": func(s string) []string {
			return []string{s}
		},
	}

	res, err := ExecuteWithStringHandler(`{{int::aaa(1)}} {{int::aaa(1)}} {{int::aaa(1)}}`, m)
	spew.Dump(res)
	if err != nil {
		panic(err)
	}
	if len(res) < 1 || res[0] != "1 1 1" {
		panic("error")
	}
}

// Escape
func TestEscape(t *testing.T) {
	for _, v := range [][]string{
		{"{{echo(\\{{)}})}}", "{{)}}"},
		{"\\{{echo(1)}}", "\\1"},                   // Do not escape outside tags
		{"\\{{echo(1)\\}}", "\\{{echo(1)\\}}"},     // {{Then start the tag syntax, which needs to be escaped. \}}. After escaping, it cannot be used as a label closing symbol, causing label parsing to fail. The original text output
		{"\\{{echo(1)\\}}}}", "\\{{echo(1)\\}}}}"}, // The tag parsing is successful, but because the data in the tag is `echo(1)}`Compilation fails, resulting in original text output
		//{`{{echo({{echo(\\\\)}})}}`, `\\`},         // Multi-layer Tag nested escaping
		{`{{echo({{echo(\\)}})}}`, `\\`}, // \not escaped.
		{`{{echo(C:\Abc\tmp)}}`, `C:\Abc\tmp`},
	} {
		res, err := ExecuteWithStringHandler(v[0], map[string]func(string2 string) []string{
			"echo": func(s string) []string {
				return []string{s}
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res[0] != v[1] {
			t.Fatal(spew.Sprintf("expect: %s, got: %s", v[1], res[0]))
		}
	}
}

func TestMagicLabel(t *testing.T) {
	checkSameString := func(s []string) bool {
		set := utils.NewSet[string]()
		for _, v := range s {
			set.Add(v)
		}
		return len(set.List()) == 1
	}
	_ = checkSameString
	for _, v := range [][]any{
		{"{{randstr()}}{{repeat(10)}}", func(s []string) bool {
			return true
		}},
		{"{{randstr::dyn()}}{{repeat(10)}}", func(s []string) bool {
			return len(s) == 10 && s[0] != s[1]
		}},
		{"{{array::1(a|b)}}{{array::1(a|b|c)}}", []string{"aa", "bb", "c"}},
		{"{{array::1::rep(a|b)}}{{array::1(a|b|c)}}", []string{"aa", "bb", "bc"}},
		{"{{array::1(a|b|c)}}{{array::1::rep(a|b)}}", []string{"aa", "bb", "cb"}},
	} {
		t, r := v[0], v[1]
		spew.Dump(t)
		result, err := ExecuteWithStringHandler(t.(string), map[string]func(string) []string{
			"array": func(s string) []string {
				return strings.Split(s, "|")
			},
			"raw": func(s string) []string {
				return []string{s}
			},
			"randstr": func(s string) []string {
				return []string{utils.RandStringBytes(10)}
			},
			"repeat": func(s string) []string {
				res := make([]string, 0)
				n, err := strconv.Atoi(s)
				if err != nil {
					return res
				}

				for range make([]int, n) {
					res = append(res, "")
				}
				return res
			},
		})
		if err != nil {
			panic(err)
		}
		spew.Dump(result)
		switch ret := r.(type) {
		case string:
			if result[0] != r {
				m := fmt.Sprintf("got: %v expect: %v", strconv.Quote(result[0]), strconv.Quote(ret))
				panic(m)
			}
		case []string:
			if len(result) != len(ret) {
				panic("check failed")
			}
			for i, v := range result {
				if v != ret[i] {
					panic("check failed")
				}
			}
		case func([]string) bool:
			if !ret(result) {
				panic("check failed")
			}
		default:
			panic("unknown type")
		}
	}
}

func TestRawTag(t *testing.T) {
	for _, v := range [][]string{
		{"{{=asdasd=}}", "asdasd"},                                  // Regular
		{`\{{=hello{{=hello\{{=world=}}`, `\{{=hellohello{{=world`}, // Test raw tag escape
	} {
		res, err := ExecuteWithStringHandler(v[0], map[string]func(string2 string) []string{
			"echo": func(s string) []string {
				return []string{s}
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res[0] != v[1] {
			t.Fatal(spew.Sprintf("expect: %s, got: %s", v[1], res[0]))
		}
	}
}
func TestMutiTag(t *testing.T) {
	for _, v := range [][]string{
		{"{{echo({{={{echo()}}=}})}}", "{{echo()}}"}, // Regular
		//{`{{echo({{=}}=}})}}`, `}}`}, // Test nesting (raw tags should block all syntax)
	} {
		res, err := ExecuteWithStringHandler(v[0], map[string]func(string2 string) []string{
			"echo": func(s string) []string {
				return []string{s}
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res[0] != v[1] {
			t.Fatal(spew.Sprintf("expect: %s, got: %s", v[1], res[0]))
		}
	}
}

// Test tag execution errors
func TestErrors(t *testing.T) {
	// Several cases of execution errors: Tag compilation error (return to original text ), function name not found (generated empty?), internal execution error of the function, continue to generate
	res, err := ExecuteWithStringHandler("{{panic(error}}", testMap)
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "{{panic(error}}" {
		t.Fatal("expect `{{panic(error}}`")
	}

	res, err = ExecuteWithStringHandler("{{aaa}}", testMap)
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "" {
		t.Fatal("expect ``")
	}

	res, err = ExecuteWithStringHandler("{{echo(a{{panic(error)}}b)}}", testMap)
	if res[0] != "ab" {
		t.Fatal("expect `ab`")
	}
}
