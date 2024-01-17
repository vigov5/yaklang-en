package parser

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/utils"
	"testing"
)

func TestIndexString(t *testing.T) {
	res := IndexAllSubstrings("{{=aa=}}", "{{", "{{=", "}}", "=}}") //A little bug
	spew.Dump(res)
}
func TestEscaper_Unescape(t *testing.T) {
	chars := []string{"{{", "("} // Single and double characters
	escaper := NewDefaultEscaper(chars...)
	for _, testcase := range [][2]string{
		// Test boundary
		{
			`\%s\%s`,
			`%s%s`,
		},
		{
			`\%saaa\%s`,
			`%saaa%s`,
		},
		{
			`aa\%saaa\%saa`,
			`aa%saaa%saa`,
		},
		//Escape undefined characters (it should be the characters themselves after escaping)
		{
			`aa\a%saaa\a%saa`,
			`aaa%saaaa%saa`,
		},
		// Escape escape characters
		{
			`aa\\%saaa\\%saa`,
			`aa\%saaa\%saa`,
		},
		{
			`aa\\\%saaa\\\%saa`,
			`aa\%saaa\%saa`,
		},
	} {
		s1 := fmt.Sprintf(testcase[0], utils.InterfaceToSliceInterface(chars)...)
		s2 := fmt.Sprintf(testcase[1], utils.InterfaceToSliceInterface(chars)...)
		res, err := escaper.Unescape(s1)
		if err != nil {
			t.Fatal(err)
		}
		if res != s2 {
			t.Fatal(errors.New(spew.Sprintf("unescape string `%s` error", s1)))
		}
	}
	res, err := escaper.Unescape(`\\`)
	if err != nil {
		t.Fatal(err)
	}
	if res != `\` {
		t.Fatal(errors.New(spew.Sprintf("unescape string `%s` error", `\\`)))
	}
}
func TestAutoEscape(t *testing.T) {
	chars := []string{"{{", "}}", "(", ")"} // Single and double characters
	escaper := NewDefaultEscaper(chars...)
	for _, testcase := range [][2]string{
		{
			"{{asd())", // Test boundary
			`\{{asd\(\)\)`,
		},
		{
			"asd{{asd())aaa",
			`asd\{{asd\(\)\)aaa`,
		},
		{
			`asd\{{asd\())aaa`,
			`asd\\\{{asd\\\(\)\)aaa`,
		},
	} {
		res := escaper.Escape(testcase[0])
		if testcase[1] != res {
			t.Fatal(spew.Sprintf("expect: %s, got: %s", testcase[1], res))
		}
	}
}
