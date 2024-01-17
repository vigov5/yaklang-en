package coreplugin

import (
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

func TestGRPCMUSTPASS_SQL(t *testing.T) {
	client, err := yakgrpc.NewLocalClient()
	if err != nil {
		panic(err)
	}

	pluginName := "Heuristic SQL injection detection"
	vul := VulInfo{
		Path: []string{
			"/user/by-id-safe?id=1",
			"/user/cookie-id",
			"/user/id?id=1",
			"/user/id-json?id=%7B%22uid%22%3A1%2C%22id%22%3A%221%22%7D",
			"/user/id-b64-json?id=eyJ1aWQiOjEsImlkIjoiMSJ9",
			"/user/name?name=admin",
			"/user/id-error?id=1",
			"/user/name/like?name=a",
			"/user/name/like/2?name=a",
			"/user/name/like/b64j?data=eyJuYW1lYjY0aiI6ImEifQ%3D%3D",
		},
		Headers: []*ypb.KVPair{{
			Key:   "Cookie",
			Value: "ID=1",
		}},
		ExpectedResult: map[string]int{
			//"Parameter: id No closed boundary detected":                         1,
			//"Suspected SQL injection: [ Parameter: Number [id] Unbounded closure]":                        4,
			"exists based on UNION SQL injection: [Parameter name: id Value: [1]]": 4,
			//"Suspected SQL injection: [Parameter: string [name] single quote closed]":                     1,
			"exists based on UNION SQL injection: [Parameter name: name Value: [admin]": 1,
			//"Suspected SQL injection: [Parameter: number [ID] enclosed in double quotes]":                        1,
			"exists based on UNION SQL injection: [Parameter name: ID value: [1]]": 1,
			//"Suspected SQL injection: [Parameter: string [name] like injection (%' )】":              2,
			"exists based on UNION SQL injection: [Parameter name: name value :[a]]": 1,
			//"Suspected SQL injection: [Parameter: String [data] like injection (%' )】":              1,
			"There may be error-based SQL injection: [Parameter name: id Original value: [1]] Guess database Type: MySQL": 1,
		},
		StrictMode: false,
	}

	//vul10 := VulInfo{
	//	Path:           "/user/name/like/b64?nameb64=%59%51%3d%3d",
	//	ExpectedResult: map[string]int{"Suspected SQL injection: [Parameter: string [name] like injection (%' )】": 3},
	//}
	Must(CoreMitmPlugTest(pluginName, server, vul, client, t), "SQL plug-in does not meet expectations for SQL injection detection results")
}
