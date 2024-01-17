package fuzztagx

import (
	"fmt"
	"github.com/yaklang/yaklang/common/fuzztagx/parser"
	"strconv"
	"strings"
	"testing"
)

func TestSyncRootTag(t *testing.T) {
	for _, test := range []struct {
		code   string
		expect []string
	}{
		{ // basic
			code:   "{{array(aaa|bbb|ccc)}}{{array(aaa|bbb|ccc)}}",
			expect: []string{"aaaaaa", "bbbbbb", "cccccc"},
		},
		{ // label synchronization settings
			code:   "{{array::1(aaa|bbb|ccc)}}{{array::1(aaa|bbb|ccc)}}",
			expect: []string{"aaaaaa", "bbbbbb", "cccccc"},
		},
		{ // label synchronization settings
			code:   "{{array(aaa|bbb|ccc)}}{{array::1(aaa|bbb|ccc)}}",
			expect: []string{"aaaaaa", "bbbbbb", "cccccc"},
		},
		{ // Only synchronize the outermost
			code:   "{{echo({{array(a|b|c)}}{{array(a|b|c)}})}}",
			expect: []string{"aa", "ba", "ca", "ab", "bb", "cb", "ac", "bc", "cc"},
		},
	} {
		iForRandStr := 0
		gener, err := NewGenerator(test.code, map[string]*parser.TagMethod{
			"echo": {
				Name: "echo",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					return []*parser.FuzzResult{parser.NewFuzzResultWithData(s)}, nil
				},
			},
			"repeat": {
				Name: "repeat",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					n, _ := strconv.Atoi(s)
					res := []*parser.FuzzResult{}
					for i := 0; i < n; i++ {
						res = append(res, parser.NewFuzzResultWithData(""))
					}
					return res, nil
				},
			},
			"array": {
				Name: "array",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					res := []*parser.FuzzResult{}
					for _, item := range strings.Split(s, "|") {
						res = append(res, parser.NewFuzzResultWithData(item))
					}
					return res, nil
				},
			},
			"randstr": {
				Name: "randstr",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					defer func() {
						iForRandStr++
					}()
					return []*parser.FuzzResult{parser.NewFuzzResultWithData(fmt.Sprint(iForRandStr))}, nil
				},
			},
			"f1": {
				Name: "f1",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					return []*parser.FuzzResult{parser.NewFuzzResultWithData(s + "1")}, nil
				},
			},
			"f2": {
				Name: "f2",
				Fun: func(s string) ([]*parser.FuzzResult, error) {
					return []*parser.FuzzResult{parser.NewFuzzResultWithData(s + "2")}, nil
				},
			},
		}, false, true)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; gener.Next(); i++ {
			if string(gener.Result().GetData()) != test.expect[i] {
				t.Fatal("error")
			}
		}
	}
}
