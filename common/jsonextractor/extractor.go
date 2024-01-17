package jsonextractor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm/vmstack"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

import "github.com/tidwall/gjson"

var (
	reQuoted = regexp.MustCompile(`(?P<quoted>(\\x[0-9a-fA-F]{2}))`)
)

func FixJson(b []byte) []byte {
	// invalid character 'x' in string escape code
	b = reQuoted.ReplaceAllFunc(b, func(i []byte) []byte {
		raw, err := strconv.Unquote(`"` + string(i) + `"`)
		if err != nil || len(raw) <= 0 {
			return i
		}
		return []byte(fmt.Sprintf(`\u%04x`, raw[0]))
	})
	return b
}

func JsonValidObject(b []byte) ([]byte, bool) {
	if gjson.ValidBytes(b) {
		return b, true
	}

	r := gjson.ParseBytes(b)
	var buf []string
	if r.IsObject() {
		for k, v := range r.Map() {
			kJsonBytes, _ := json.Marshal(k)
			var kJson = string(kJsonBytes)
			if strings.HasPrefix(kJson, `"`) && strings.HasSuffix(kJson, `"`) {
				buf = append(buf, fmt.Sprintf(`%v: %s`, kJson, v.String()))
			} else {
				buf = append(buf, fmt.Sprintf(`"%v": %s`, kJson, v.String()))
			}
		}
	}

	if len(buf) > 0 {
		return []byte("{" + strings.Join(buf, ", ") + "}"), true
	}

	return nil, false
}

const (
	state_SingleQuoteString = "s-quote"
	state_DoubleQuoteString = "d-quote"
	state_BacktickString    = "b-quote"
	state_jsonObj           = "json-object"
	state_data              = "data"
	state_esExpr            = "es-expr"
	state_reset             = "reset"
)

func ExtractObjectIndexes(c string) [][2]int {
	scanner := bufio.NewScanner(bytes.NewBufferString(c))
	scanner.Split(bufio.ScanBytes)

	var index = -1
	var objectDepth = 0
	var objectDepthIndexTable = make(map[int]int)

	var results [][2]int
	stack := vmstack.New()
	pushState := func(i string) {
		if i == state_jsonObj {
			objectDepth++
			if _, existed := objectDepthIndexTable[objectDepth]; !existed {
				objectDepthIndexTable[objectDepth] = index
			}
		}
		stack.Push(i)
	}
	popState := func() {
		r := stack.Pop()
		if r != nil {
			raw, ok := r.(string)
			if ok && raw == state_jsonObj {
				// recording result
				ret, ok := objectDepthIndexTable[objectDepth]
				if ok && ret >= 0 {
					results = append(results, [2]int{objectDepthIndexTable[objectDepth], index + 1})
				}
				delete(objectDepthIndexTable, objectDepth)
				if objectDepth == 0 {
					objectDepthIndexTable = make(map[int]int)
				}
				objectDepth--
			}
		}
	}
	currentState := func() string {
		basicState := stack.Peek()
		if basicState == nil {
			return state_reset
		}
		return basicState.(string)
	}

	// startup stack state machine
	pushState(state_data)
	var last byte
	var ch byte
	for {
		last = ch
		if !scanner.Scan() {
			break
		}
		index++
		results := scanner.Bytes()
		if len(results) <= 0 {
			break
		}
		ch = results[0]

		switch currentState() {
		case state_data:
			switch ch {
			case '{':
				pushState(state_jsonObj)
				continue
			case '"':
				pushState(state_DoubleQuoteString)
				continue
			case '\'':
				pushState(state_SingleQuoteString)
				continue
			case '`':
				pushState(state_esExpr)
				continue
			}
		case state_jsonObj:
			switch ch {
			case '{':
				pushState(state_jsonObj)
				continue
			case '"':
				pushState(state_DoubleQuoteString)
				continue
			case '\'':
				pushState(state_SingleQuoteString)
				continue
			case '`':
				pushState(state_esExpr)
				continue
			case '}':
				popState()
				continue
			}
		case state_esExpr:
			switch ch {
			case '}':
				popState()
				continue
			}
		case state_DoubleQuoteString:
			switch ch {
			case '"':
				if last != '\\' {
					popState()
					continue
				}
			}
		case state_SingleQuoteString:
			switch ch {
			case '\'':
				if last != '\'' {
					popState()
					continue
				}
			}
		case state_BacktickString:
			/*
				This is very special, there are several situation needs to be dealt with
				`abc`
				`abc${"123" + `abc`}`
			*/
			switch ch {
			case '{':
				if last == '$' {
					// ${ , thinking this is expr
					pushState(state_esExpr)
					continue
				}
			case '`':
				if last != '\\' {
					popState()
					continue
				}
			}
		case state_reset:
			// empty state traceback, most likely it starts with the problematic
			//currentPair[0] = -1
			//currentPair[1] = -1
			//currentPair[2] = -1
			pushState(state_data)
		}
	}

	// shrinking result
	var blocks [][2]int
	var currentBlock = [2]int{-1, -1}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i][0] < results[j][0]
	})
	currentBlockIsJson := func() bool {
		if currentBlock[0] < 0 {
			return false
		}
		return json.Valid([]byte(c[currentBlock[0]:currentBlock[1]]))
	}
	for _, result := range results {
		retRaw := c[result[0]:result[1]]
		_, isJson := JsonValidObject([]byte(retRaw))
		// fmt.Printf("%v: idx: %v json: %v\n", retRaw, result, isJson)
		if currentBlock[0] < 0 {
			currentBlock[0], currentBlock[1] = result[0], result[1]
			continue
		}

		if result[0] >= currentBlock[0] && result[1] <= currentBlock[1] && currentBlockIsJson() {
			// included content
			continue
		} else {
			blocks = append(blocks, [2]int{currentBlock[0], currentBlock[1]})
			if isJson {
				currentBlock[0], currentBlock[1] = result[0], result[1]
			} else {
				blocks = append(blocks, [2]int{result[0], result[1]})
				currentBlock[0] = -1
				currentBlock[1] = -1
			}
		}
	}
	if currentBlock[0] < 0 {
		return blocks
	}
	return append(blocks, [2]int{currentBlock[0], currentBlock[1]})
}

func ExtractJSONWithRaw(raw string) (results []string, rawStr []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("extract json failed: %s", err)
		}
	}()
	var extraValid []string
	for _, obj := range ExtractObjectIndexes(raw) {
		jsonStr := raw[obj[0]:obj[1]]
		if ret, ok := JsonValidObject([]byte(jsonStr)); ok {
			if !json.Valid([]byte(jsonStr)) {
				rawStr = append(rawStr, jsonStr)
				// repaired JSON
				extraValid = append(extraValid, string(ret))
			} else {
				// Perfect JSON
				results = append(results, jsonStr)
			}
		} else {
			rawStr = append(rawStr, jsonStr)
		}
	}
	if len(extraValid) > 0 {
		results = append(results, extraValid...)
	}
	return
}

func ExtractStandardJSON(raw string) []string {
	jsonStr, _ := ExtractJSONWithRaw(raw)
	return jsonStr
}
