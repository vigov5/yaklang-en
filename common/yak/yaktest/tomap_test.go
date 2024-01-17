package yaktest

import (
	"fmt"
	"testing"
)

func TestRun_TOMAP(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "Test x.ConvertToMap map[string]string",
			Src:  fmt.Sprintf(`result = x.ConvertToMap({"aaaa": "127.0.0.1"}); assert(result["default"] == undefined)`),
		},
		{
			Name: "Test x.ConvertToMap map[string][]byte",
			Src:  fmt.Sprintf(`result = x.ConvertToMap({"aaaa": []byte("asdfasdfasd")}); assert(result["default"] == undefined)`),
		},
		{
			Name: "Test x.ConvertToMap map[string]int",
			Src:  fmt.Sprintf(`result = x.ConvertToMap({"aaaa": 123});ret = str.Join(result["aaaa"], ""); println(ret); assert(ret == "123")`),
		},
		{
			Name: "Test x.ConvertToMap map[string]float",
			Src:  fmt.Sprintf(`result = x.ConvertToMap({"aaaa": 123.1});ret = str.Join(result["aaaa"], ""); println(ret); assert(ret == "123.1")`),
		},
	}

	Run("x.ConvertToMap usability testing", t, cases...)
}
