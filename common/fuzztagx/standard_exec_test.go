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
}

// 同步渲染数量测试
func TestSyncRender(t *testing.T) {
	for i, testcase := range [][2]any{
		{
			"{{echo::1({{list(aaa|ccc)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
			3,
		},
		{
			"{{echo::1({{list(aaa|ccc|ddd)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
			3,
		},
		{
			"{{echo::1({{list(aaa|ccc|ddd|eee)}})}}{{echo::1({{list(aaa|ccc|ddd)}})}}",
			4,
		},
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
}

// 畸形测试
func TestDeformityTag(t *testing.T) {
	for _, v := range [][]string{
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
		//{"{{xx12:-_(____)____)}}[[[[}}", "{{xx12:-_(____)____)}}[[[[}}"}, // {{xx12:-_(____)____)}}应该被正确解析
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
		//{"{{echo(123123\\))}}", "123123)"}, // 括号不需要转义
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

// Tag名后换行
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

// 转义
func TestEscape(t *testing.T) {
	for _, v := range [][]string{
		{"{{echo(\\{{)}})}}", "{{)}}"},
		{"\\{{echo(1)}}", "\\1"},                   // 标签外不转义
		{"\\{{echo(1)\\}}", "\\{{echo(1)\\}}"},     // {{之后开始tag语法，需要转义，\}}转义后不能作为标签闭合符号，导致标签解析失败，原文输出
		{"\\{{echo(1)\\}}}}", "\\{{echo(1)\\}}}}"}, // 标签解析成功，但由于标签内数据`echo(1)}`编译失败，导致原文输出

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
			t.Fatal("error")
		}
	}
}

func TestExecutePrefixTag(t *testing.T) {
	var m = map[string]func(string) []string{
		"expr:a": func(s string) []string {
			return []string{s}
		},
	}
	testData := []string{
		//`}}`,
		`x"{{int(a)}}"`,
		"base64(111)",
		" base64(111)",
		" base64(111) ",
		" base64(base64(111)) ",
		"base64(111))))",
		"base64((111)",
	}
	for _, d := range testData {
		println(d)
		res, err := ExecuteWithStringHandler(fmt.Sprintf(`{{expr:a(%s)}}`, d), m)
		if err != nil {
			panic(utils.Errorf("test data [%v] error: %v", d, err))
		}
		if len(res) == 0 {
			panic("generate error")
		}
		if res[0] != d {
			panic("TestExecutePrefixTag failed")
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
		{"{{raw::raw({{aaa()}})}}", "{{aaa()}}"},
		{"aaa{{raw::raw({{aaa()}})}}aaa{{repeat(3)}}", []string{"aaa{{aaa()}}aaa", "aaa{{aaa()}}aaa", "aaa{{aaa()}}aaa"}},
		{"{{randstr::rep()}}{{repeat(10)}}", func(s []string) bool {
			return true
		}},
		{"{{randstr()}}{{repeat(10)}}", func(s []string) bool {
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