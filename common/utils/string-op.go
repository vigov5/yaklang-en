package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
)

// . SplitAndTrim splits the string s into string slices according to sep, and removes the leading and trailing whitespace of each string. Character
// Example:
// ```
// str.SplitAndTrim(" hello yak ", " ") // ["hello", "yak"]
// ```
func PrettifyListFromStringSplited(Raw string, sep string) (targets []string) {
	targetsRaw := strings.Split(Raw, sep)
	for _, tRaw := range targetsRaw {
		r := strings.TrimSpace(tRaw)
		if len(r) > 0 {
			targets = append(targets, r)
		}
	}
	return
}

func PrettifyShrinkJoin(sep string, s ...string) string {
	var buf bytes.Buffer
	var count = 0
	var existedHashMap = make(map[string]struct{})
	for _, element := range s {
		for _, i := range PrettifyListFromStringSplited(element, sep) {
			if i == "" {
				continue
			}

			_, ok := existedHashMap[i]
			if ok {
				continue
			}
			existedHashMap[i] = struct{}{}
			if count == 0 {
				buf.WriteString(i)
			} else {
				buf.WriteString(sep + i)
			}
			count++
		}
	}
	return buf.String()
}

func PrettifyJoin(sep string, s ...string) string {
	var buf bytes.Buffer
	var count = 0
	for _, i := range s {
		if i == "" {
			continue
		}

		if count == 0 {
			buf.WriteString(i)
		} else {
			buf.WriteString(sep + i)
		}
		count++
	}
	return buf.String()
}

// PrettifyListFromStringSplitEx split string using given sep if no sep given sep = []string{",", "|"}
func PrettifyListFromStringSplitEx(Raw string, sep ...string) (targets []string) {
	if len(sep) <= 0 {
		sep = []string{",", "|"}
	}
	patternStr := ""
	for _, v := range sep {
		if len(v) > 0 {
			patternStr += regexp.QuoteMeta(string(v[0])) + "|"
		}
	}
	if len(patternStr) > 0 {
		patternStr = patternStr[:len(patternStr)-1]
	}

	var targetsRaw []string
	re, err := regexp.Compile(patternStr)
	if err != nil {
		log.Warn(err)
		return targetsRaw
	}
	targetsRaw = re.Split(Raw, -1)
	for _, tRaw := range targetsRaw {
		r := strings.TrimSpace(tRaw)
		if len(r) > 0 {
			targets = append(targets, r)
		}
	}
	return
}

func ToLowerAndStrip(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// StringSliceContains Determines whether the string slice s contains raw. For non-string slices, It will try to convert its elements into strings and then determine whether it contains
// Example:
// ```
// str.StringSliceContains(["hello", "yak"], "yak") // true
// str.StringSliceContains([1, 2, 3], "4") // false
// ```
func StringSliceContain(s interface{}, raw string) (result bool) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	haveResult := false
	switch ret := s.(type) {
	case []string:
		for _, i := range ret {
			if i == raw {
				return true
			}
		}
		return false
	}
	funk.ForEach(s, func(i interface{}) {
		if haveResult {
			return
		}
		if InterfaceToString(i) == raw {
			haveResult = true
		}
	})
	return haveResult
}

// StringContainsAnyOfSubString Determines whether the string s contains any substring in subs
// Example:
// ```
// str.StringContainsAnyOfSubString("hello yak", ["yak", "world"]) // true
// ```
func StringContainsAnyOfSubString(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func StringContainsAllOfSubString(s string, subs []string) bool {
	if len(subs) <= 0 {
		return false
	}
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func IStringContainsAnyOfSubString(s string, subs []string) bool {
	for _, sub := range subs {
		if IContains(s, sub) {
			return true
		}
	}
	return false
}

func ConvertToStringSlice(raw ...interface{}) (r []string) {
	for _, e := range raw {
		r = append(r, fmt.Sprintf("%v", e))
	}
	return
}

func ChanStringToSlice(c chan string) (result []string) {
	for l := range c {
		result = append(result, l)
	}
	return
}

var (
	cStyleCharPRegexp, _ = regexp.Compile(`\\((x[0-9abcdef]{2})|([0-9]{1,3}))`)
)

func ParseCStyleBinaryRawToBytes(raw []byte) []byte {
	// like "\\x12" => "\x12"
	return cStyleCharPRegexp.ReplaceAllFunc(raw, func(i []byte) []byte {
		if bytes.HasPrefix(i, []byte("\\x")) {
			if len(i) == 4 {
				rawChar := string(i[2:])
				charInt, err := strconv.ParseInt("0x"+string(rawChar), 0, 16)
				if err != nil {
					return i
				}
				return []byte{byte(charInt)}
			} else {
				return i
			}
		} else if bytes.HasPrefix(raw, []byte("\\")) {
			if len(i) > 1 && len(i) <= 4 {
				rawChar := string(i[1:])
				charInt, err := strconv.ParseInt(string(rawChar), 10, 8)
				if err != nil {
					return i
				}
				return []byte{byte(charInt)}
			} else {
				return i
			}
		}
		return i
	})
}

var GbkToUtf8 = codec.GbkToUtf8
var Utf8ToGbk = codec.Utf8ToGbk

func ParseStringToVisible(raw interface{}) string {
	var s = InterfaceToString(raw)
	s = EscapeInvalidUTF8Byte([]byte(s))
	//s = strings.ReplaceAll(s, "\x20", "\\x20")
	s = strings.ReplaceAll(s, "\x0b", "\\v")
	r, err := regexp.Compile(`\s`)
	if err != nil {
		return s
	}
	return r.ReplaceAllStringFunc(s, func(s string) string {
		var result = strconv.Quote(s)
		for strings.HasPrefix(result, "\"") {
			result = result[1:]
		}
		for strings.HasSuffix(result, "\"") {
			result = result[:len(result)-1]
		}
		return result
	})
}

func EscapeInvalidUTF8Byte(s []byte) string {
	// . The result returned by this operation is not equivalent to the original string
	ret := make([]rune, 0, len(s)+20)
	start := 0
	for {
		r, size := utf8.DecodeRune(s[start:])
		if r == utf8.RuneError {
			// Description is empty
			if size == 0 {
				break
			} else {
				// is not rune
				ret = append(ret, []rune(fmt.Sprintf("\\x%02x", s[start]))...)
			}
		} else {
			// is not a control character such as line break
			if unicode.IsControl(r) && !unicode.IsSpace(r) {
				ret = append(ret, []rune(fmt.Sprintf("\\x%02x", r))...)
			} else {
				// Normal character
				ret = append(ret, r)
			}
		}
		start += size
	}
	return string(ret)
}

var GBKSafeString = codec.GBKSafeString

func LastLine(s []byte) []byte {
	s = bytes.TrimSpace(s)
	scanner := bufio.NewScanner(bytes.NewReader(s))
	scanner.Split(bufio.ScanLines)

	var lastLine = s
	for scanner.Scan() {
		lastLine = scanner.Bytes()
	}

	return lastLine
}

func RemoveUnprintableChars(raw string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(raw))
	scanner.Split(bufio.ScanBytes)

	var buf = bytes.NewBufferString("")
	for scanner.Scan() {
		c := scanner.Bytes()[0]

		if c <= 0x7e && c >= 0x20 {
			buf.WriteByte(c)
		} else {
			buf.WriteString(`\x` + fmt.Sprintf("%02x", c))
		}
	}

	return buf.String()
}

func RemoveUnprintableCharsWithReplace(raw string, handle func(i byte) string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(raw))
	scanner.Split(bufio.ScanBytes)

	var r []byte
	for scanner.Scan() {
		c := scanner.Bytes()[0]

		if c <= 0x7e && c >= 0x20 {
			r = append(r, c)
		} else {
			r = append(r, []byte(handle(c))...)
		}
	}

	return string(r)
}

func RemoveUnprintableCharsWithReplaceItem(raw string) string {
	return RemoveUnprintableCharsWithReplace(raw, func(i byte) string {
		return fmt.Sprintf("__HEX_%v__", codec.EncodeToHex([]byte{i}))
	})
}

func RemoveRepeatedWithStringSlice(slice []string) []string {
	r := map[string]interface{}{}
	for _, s := range slice {
		r[s] = 1
	}

	var r2 []string
	for k, _ := range r {
		r2 = append(r2, k)
	}
	return r2
}

var (
	titleRegexp = regexp.MustCompile(`(?is)\<title\>(.*?)\</?title\>`)
)

func ExtractTitleFromHTMLTitle(s string, defaultValue string) string {
	var title string
	l := titleRegexp.FindString(s)
	if len(l) > 15 {
		title = EscapeInvalidUTF8Byte([]byte(l))[7 : len(l)-8]
	}
	titleRunes := []rune(title)
	if len(titleRunes) > 128 {
		title = string(titleRunes[0:128]) + "..."
	}

	if title == "" {
		return defaultValue
	}

	return title
}
