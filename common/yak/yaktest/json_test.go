package yaktest

import "testing"

func TestJson(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "JSON package test failed: judge ARRAY 1",
			Src:  `j, err := json.New("[1,2,3,4,5,6]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package Test failed: judge ARRAY 2",
			Src:  `j, err := json.New("[1,2,\"3\",4,5,6]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge ARRAY 3",
			Src:  `j, err := json.New("[1]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge ARRAY 4",
			Src:  `j, err := json.New("[{\"1\": 123}, 2]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge ARRAY 5",
			Src:  `j, err := json.New("[1,2,[2,3]]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge ARRAY 6",
			Src:  `j, err := json.New("[1,4,[2,3],{\"a\": 123}, null]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge ARRAY 7",
			Src:  `j, err := json.New("[1,2,3,4,5,6]"); die(err); if (!j.IsSlice()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge Nil",
			Src: `j, err := json.New("null"); die(err); 
if (!j.IsNull()){
die("error for json array test")
}`,
		},
		{
			Name: "JSON package test failed: judge Object 1",
			Src:  `j, err := json.New("{}"); die(err); if (!j.IsMap()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge Object 2",
			Src:  `j, err := json.New("{\"ab\": 123}"); die(err); if (!j.IsMap()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge Object 3",
			Src:  `j, err := json.New("{\"b\": \"basdf\", \"a\": null}"); die(err); if (!j.IsMap()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge Object 4",
			Src:  `j, err := json.New("{\"b\": 12, \"basdf\": 123}"); die(err); if (!j.IsMap()) {die("error for json array test")}`,
		},
		{
			Name: "JSON package test failed: judge Object 5",
			Src:  `j, err := json.New("{\"1\": [1,2,3]}"); die(err); if (!j.IsMap()) {die("error for json array test")}`,
		},
	}

	Run("Yak json test", t, cases...)
}
