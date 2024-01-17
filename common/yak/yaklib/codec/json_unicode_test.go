package codec

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestJsonUnicodeDecode(t *testing.T) {
	var a = JsonUnicodeEncode("hello ab")
	spew.Dump(a)
	println(a)
	var result = JsonUnicodeDecode(a)
	if result != "hello ab" {
		panic("unicode decode failed")
	}
}
