package yaklib

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/yaklang/yaklang/common/jsonextractor"
	"github.com/yaklang/yaklang/common/jsonpath"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

type jsonConfig struct {
	prefix string
	indent string
}

type JsonOpt func(opt *jsonConfig)

var JsonExports = map[string]interface{}{
	"New":        _yakJson,
	"Marshal":    _jsonMarshal,
	"dumps":      _jsonDumps,
	"loads":      _jsonLoad,
	"withPrefix": _withPrefix,
	"withIndent": _withIndent,

	// This is the JSONPath module
	"Find":          jsonpath.Find,
	"FindPath":      jsonpath.FindFirst,
	"ReplaceAll":    jsonpath.ReplaceAll,
	"ExtractJSON":   jsonextractor.ExtractStandardJSON,
	"ExtractJSONEx": jsonextractor.ExtractJSONWithRaw,
}

func NewJsonConfig() *jsonConfig {
	return &jsonConfig{
		prefix: "",
		indent: "",
	}
}

// withPrefix Set the prefix for JSON dumps
// Example:
// ```
// v = json.dumps({"a": "b", "c": "d"}, withPrefix="  ")
// ```
func _withPrefix(prefix string) JsonOpt {
	return func(opt *jsonConfig) {
		opt.prefix = prefix
	}
}

// . withIndent Set the indent when JSON dumps
// Example:
// ```
// v = json.dumps({"a": "b", "c": "d"}, withIndent="  ")
// ```
func _withIndent(indent string) JsonOpt {
	return func(opt *jsonConfig) {
		opt.indent = indent
	}
}

// Marshal Convert an object to JSON bytes, return the converted bytes and error
// Example:
// ```
// v, err = json.Marshal({"a": "b", "c": "d"})
// // v = b"{"a": "b", "c": "d"}"
// ```
func _jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// dumps Convert an object to a JSON string, return the converted string
// It can also receive zero to multiple request option functions for configuring the conversion process and controlling the converted Indentation, prefix, etc.
// Example:
// ```
// v = json.dumps({"a": "b", "c": "d"})
// ```
func _jsonDumps(raw interface{}, opts ...JsonOpt) string {
	config := NewJsonConfig()
	for _, opt := range opts {
		opt(config)
	}

	bytes, err := json.MarshalIndent(raw, config.prefix, config.indent)
	if err != nil {
		log.Errorf("parse error: %v", err)
		return ""
	}
	return string(bytes)
}

// loads Convert a JSON string to an object and return the converted object
// Example:
// ```
// v = json.loads(`{"a": "b", "c": "d"}`)
// ```
func _jsonLoad(raw interface{}, opts ...JsonOpt) interface{} {
	// . There is no load option in opts, so
	var i interface{}
	defaultValue := make(map[string]interface{})

	str := utils.InterfaceToString(raw)
	str = strings.TrimSpace(str)
	err := json.Unmarshal([]byte(str), &i)
	if err != nil {
		// Try to decode
		if strings.Contains(err.Error(), `character 'x'`) {
			fixed := string(jsonextractor.FixJson([]byte(str)))
			if fixed != "" {
				str = fixed
			}
			err := json.Unmarshal([]byte(str), &i)
			if err == nil {
				return i
			}
		}

		// If JSON decoding fails, try to fix it
		if strings.HasPrefix(str, "{") {
			fixed, ok := jsonextractor.JsonValidObject([]byte(str))
			if ok {
				err := json.Unmarshal([]byte(fixed), &i)
				if err == nil {
					return i
				}
			}
		}
		log.Error(err)
		return defaultValue
	}
	return i
}

type yakJson struct {
	origin     interface{}
	jsonObject interface{}
}

// Determine if it is a map/object {}
func (y *yakJson) IsObject() bool {
	return y.jsonObject != nil && reflect.TypeOf(y.jsonObject).Kind() == reflect.Map
}

func (y *yakJson) IsMap() bool {
	return y.IsObject()
}

// Determine whether it is []
func (y *yakJson) IsSlice() bool {
	return y.jsonObject != nil && ((reflect.TypeOf(y.jsonObject).Kind() == reflect.Slice) ||
		(reflect.TypeOf(y.jsonObject).Kind() == reflect.Array))
}

func (y *yakJson) IsArray() bool {
	return y.IsSlice()
}

// is not processed here for the time being. determine whether it is null
func (y *yakJson) IsNil() bool {
	return y.jsonObject == nil
}

func (y *yakJson) IsNull() bool {
	return y.IsNil()
}

// determines whether it is string
func (y *yakJson) IsString() bool {
	return y.jsonObject != nil && (reflect.TypeOf(y.jsonObject).Kind() == reflect.String)
}

// Determine if it is a number
func (y *yakJson) IsNumber() bool {
	return y.jsonObject != nil && (reflect.TypeOf(y.jsonObject).Kind() == reflect.Float64 ||
		reflect.TypeOf(y.jsonObject).Kind() == reflect.Int ||
		reflect.TypeOf(y.jsonObject).Kind() == reflect.Int64 ||
		reflect.TypeOf(y.jsonObject).Kind() == reflect.Uint64 ||
		reflect.TypeOf(y.jsonObject).Kind() == reflect.Float32 ||
		reflect.TypeOf(y.jsonObject).Kind() == reflect.Int)
}

func (y *yakJson) Value() interface{} {
	return y.jsonObject
}

// New Creates and returns a new JSON object with an error based on the incoming value
// Example:
// ```
// v, err = json.New("foo")
// v, err = json.New(b"bar")
// v, err = json.New({"a": "b", "c": "d"})
// ```
func _yakJson(i interface{}) (*yakJson, error) {
	j := &yakJson{}

	var raw interface{}
	j.origin = i

	switch ret := i.(type) {
	case []byte:
		err := json.Unmarshal(ret, &raw)
		if err != nil {
			return nil, err
		}
	case string:
		err := json.Unmarshal([]byte(ret), &raw)
		if err != nil {
			return nil, err
		}
	default:
		rawBytes, err := json.Marshal(ret)
		if err != nil {
			return nil, utils.Errorf("marshal input{%#v} failed: %v", ret, err)
		}

		err = json.Unmarshal(rawBytes, &raw)
		if err != nil {
			return nil, err
		}
	}
	j.jsonObject = raw

	return j, nil
}

// Find use JSONPath Finds and returns all values in JSON
// Example:
// ```
// v = json.Find(`{"a":"a1","c":{"a":"a2"}}`, "$..a") // v = [a1, a2]
// ```
func _jsonpathFind(json interface{}, jsonPath string) interface{} {
	return jsonpath.Find(json, jsonPath)
}

// FindPath Use JSONPath to find and retrieve the first value in JSON
// Example:
// ```
//
//	v = json.Find(`{"a":"a1","c":{"a":"a2"}}`, "$..a") // v = a1
//
// ```
func _jsonpathFindPath(json interface{}, jsonPath string) interface{} {
	return jsonpath.FindFirst(json, jsonPath)
}

// ReplaceAll Use JSONPath to replace the JSON All values, return the replaced JSON map
// Example:
// ```
// v = json.ReplaceAll(`{"a":"a1","c":{"a":"a2"}}`, "$..a", "b") // v = {"a":"b","c":{"a":"b"}}
// ```
func _jsonpathReplaceAll(json interface{}, jsonPath string, replaceValue interface{}) map[string]interface{} {
	return jsonpath.ReplaceAll(json, jsonPath, replaceValue)
}

// ExtractJSON Extract all repaired JSON strings from a piece of text
// Example:
// ```
// v = json.ExtractJSON(`Here is your result: {"a": "b"} and {"c": "d"}`)
// // v = ["{"a": "b"}", "{"c": "d"}"]
// ```
func _jsonpathExtractJSON(raw string) []string {
	return jsonextractor.ExtractStandardJSON(raw)
}

// ExtractJSONEx From a piece of text Extract all the repaired JSON strings from it, and return the repaired JSON string and the pre-repaired JSON string.
func _jsonpathExtractJSONEx(raw string) (results []string, rawStr []string) {
	return jsonextractor.ExtractJSONWithRaw(raw)
}
