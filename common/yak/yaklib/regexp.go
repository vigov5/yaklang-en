package yaklib

import (
	"fmt"
	"regexp"

	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

// RegexpMatch uses regular expressions to try to match the string. If the match is successful, it returns true, otherwise it returns false.
// Example:
// ```
// str.RegexpMatch("^[a-z]+$", "abc") // true
// ```
func _strRegexpMatch(pattern string, s interface{}) bool {
	return _reMatch(pattern, s)
}

// Match uses regular expressions. Attempts to match a string, returning true if the match is successful, false otherwise
// Example:
// ```
// re.Match("^[a-z]+$", "abc") // true
// ```
func _reMatch(pattern string, s interface{}) bool {
	r, err := regexp.Compile(pattern)
	if err != nil {
		_diewith(utils.Errorf("compile[%v] failed: %v", pattern, err))
		return false
	}

	switch ret := s.(type) {
	case []byte:
		return r.Match(ret)
	case string:
		return r.MatchString(ret)
	default:
		_diewith(utils.Errorf("target: %v should be []byte or string", spew.Sdump(s)))
	}
	return false
}

// Find uses regular expressions to try to match strings. If the match is successful, it returns the first matched string. Otherwise, it returns an empty string.
// Example:
// ```
// re.Find("apple is an easy word", "^[a-z]+") // "apple"
// ```
func _find_extractByRegexp(origin interface{}, re string) string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return ""
	}
	return r.FindString(utils.InterfaceToString(origin))
}

// FindAll uses regular expressions to try to match a string, if If the match is successful, it returns all matched strings, otherwise it returns an empty string slice.
// Example:
// ```
// re.FindAll("Well,yakit is GUI client for yaklang", "yak[a-z]+") // ["yakit", "yaklang"]
// ```
func _findAll_extractByRegexp(origin interface{}, re string) []string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindAllString(utils.InterfaceToString(origin), -1)
}

// FindAllIndex uses regular expressions to try to match strings. If the match is successful, it returns the starting position of all matched strings. and the end position, otherwise it returns a two-dimensional slice of empty integers.
// Example:
// ```
// re.FindAllIndex("Well,yakit is GUI client for yaklang", "yak[a-z]+") // [[5, 10], [29, 36]]
// ```
func _findAllIndex_extractByRegexp(origin interface{}, re string) [][]int {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindAllStringIndex(utils.InterfaceToString(origin), -1)
}

// FindIndex uses regular expressions to try to match the string. If the match is successful, it returns an integer slice of length 2. The first element is the starting position and the second element is the end position. Otherwise returns an empty integer slice
// Example:
// ```
// re.FindIndex("Well,yakit is GUI client for yaklang", "yak[a-z]+") // [5, 10]
// ```
func _findIndex_extractByRegexp(origin interface{}, re string) []int {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindStringIndex(utils.InterfaceToString(origin))
}

// FindSubmatch uses regular expressions to try to match the string. If the match is successful, it returns the first matched string and the submatched string. Otherwise, it returns an empty string slice.
// Example:
// ```
// re.FindSubmatch("Well,yakit is GUI client for yaklang", "yak([a-z]+)") // ["yakit", "it"]
// ```
func _findSubmatch_extractByRegexp(origin interface{}, re string) []string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindStringSubmatch(utils.InterfaceToString(origin))
}

// FindSubmatchIndex uses regular expressions to try to match the string. If the match is successful, it returns the first matched string and the starting position and end position of the submatched string. Otherwise, it returns an empty integer slice.
// Example:
// ```
// re.FindSubmatchIndex("Well,yakit is GUI client for yaklang", "yak([a-z]+)") // [5, 10, 8, 10]
// ```
func _findSubmatchIndex_extractByRegexp(origin interface{}, re string) []int {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindStringSubmatchIndex(utils.InterfaceToString(origin))
}

// FindSubmatchAll uses regular expressions to try to match a string, if the match is successful returns all matched strings and submatched strings, otherwise returns a two-dimensional slice of empty string slices
// Example:
// ```
// // [["yakit", "it"], ["yaklang", "lang"]]
// re.FindSubmatchAll("Well,yakit is GUI client for yaklang", "yak([a-z]+)")
// ```
func _findSubmatchAll_extractByRegexp(origin interface{}, re string) [][]string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindAllStringSubmatch(utils.InterfaceToString(origin), -1)
}

// FindSubmatchAllIndex uses regular expressions to try to match the string. , if the match is successful, return the start and end positions of all matched strings and submatched strings, otherwise return a two-dimensional slice of empty integer slices
// Example:
// ```
// // [[5, 10, 8, 10], [29, 36, 32, 36]]
// re.FindSubmatchAllIndex("Well,yakit is GUI client for yaklang", "yak([a-z]+)")
// ```
func _findSubmatchAllIndex_extractByRegexp(origin interface{}, re string) [][]int {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return nil
	}
	return r.FindAllStringSubmatchIndex(utils.InterfaceToString(origin), -1)
}

// ReplaceAllWithFunc uses regular expression matching and replaces the string with a custom function, and returns the replacement. The string
// Example:
// ```
// // "yaklang is a programming language"
// re.ReplaceAllWithFunc("yakit is programming language", "yak([a-z]+)", func(s) {
// return "yaklang"
// })
// ```
func _replaceAllFunc_extractByRegexp(origin interface{}, re string, newStr func(string) string) string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return utils.InterfaceToString(origin)
	}
	return r.ReplaceAllStringFunc(utils.InterfaceToString(origin), newStr)
}

// ReplaceAll uses regular expressions to match and replace strings, and returns the replaced String
// Example:
// ```
// // "yaklang is a programming language"
// re.ReplaceAll("yakit is programming language", "yak([a-z]+)", "yaklang")
// ```
func _replaceAll_extractByRegexp(origin interface{}, re string, newStr interface{}) string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("compile %v failed: %s", re, err)
		return utils.InterfaceToString(origin)
	}
	return r.ReplaceAllString(utils.InterfaceToString(origin), utils.InterfaceToString(newStr))
}

// FindGroup uses regular expressions to match strings. If the match is successful, it returns a map whose key name is the named capture group in the regular expression and the key value is the matched string. Otherwise, an empty map is returned.
// Example:
// ```
// // {"0": "yakit", "other": "it"}
// re.FindGroup("Well,yakit is GUI client for yaklang", "yak(?P<other>[a-z]+)")
// ```
func reExtractGroups(i interface{}, re string) map[string]string {
	r, err := regexp.Compile(re)
	if err != nil {
		log.Error(err)
		return make(map[string]string)
	}
	matchIndex := map[int]string{}
	for _, name := range r.SubexpNames() {
		matchIndex[r.SubexpIndex(name)] = name
	}

	result := make(map[string]string)
	for index, value := range r.FindStringSubmatch(utils.InterfaceToString(i)) {
		name, ok := matchIndex[index]
		if !ok {
			name = fmt.Sprint(index)
		}
		result[name] = value
	}
	return result
}

// FindGroupAll uses the regular expression to match the string. If the match is successful, it returns a map slice whose key name is the named capture group in the regular expression and the key value is the matched string, otherwise an empty map slice is returned.
// Example:
// ```
// // [{"0": "yakit", "other": "it"}, {"0": "yaklang", "other": "lang"}]
// re.FindGroupAll("Well,yakit is GUI client for yaklang", "yak(?P<other>[a-z]+)")
// ```
func reExtractGroupsAll(i interface{}, raw string) []map[string]string {
	re, err := regexp.Compile(raw)
	if err != nil {
		log.Error(err)
		return nil
	}
	matchIndex := map[int]string{}
	for _, name := range re.SubexpNames() {
		matchIndex[re.SubexpIndex(name)] = name
	}

	var results []map[string]string
	for _, matches := range re.FindAllStringSubmatch(utils.InterfaceToString(i), -1) {
		result := make(map[string]string)
		for index, value := range matches {
			name, ok := matchIndex[index]
			if !ok {
				name = fmt.Sprint(index)
			}
			result[name] = value
		}
		results = append(results, result)
	}
	return results
}

// QuoteMeta returns a string that escapes all regular expression metacharacters in s. The result after
// Example:
// ```
// str.QuoteMeta("^[a-z]+$") // "\^\\[a-z\]\\+$"
// ```
func _quoteMeta(s string) string {
	return regexp.QuoteMeta(s)
}

// Compile parses the regular expression into a regular expression structure reference
// Example:
// ```
// re.Compile("^[a-z]+$")
// ```
func _compile(expr string) (*regexp.Regexp, error) {
	return regexp.Compile(expr)
}

// CompilePOSIX Parses the regular expression into a regular expression structure reference that conforms to POSIX ERE(egrep) syntax, and changes the matching semantics to left longest match
// Example:
// ```
// re.CompilePOSIX("^[a-z]+$")
// ```
func _compilePOSIX(expr string) (*regexp.Regexp, error) {
	return regexp.CompilePOSIX(expr)
}

// MustCompile Parses the regular expression into a regular expression object structure reference, if parsed Failure will cause a crash.
// Example:
// ```
// re.MustCompile("^[a-z]+$")
// ```
func _mustCompile(str string) *regexp.Regexp {
	return regexp.MustCompile(str)
}

// MustCompilePOSIX parses the regular expression into a POSIX regular expression structure reference. If the parsing fails, it will cause a crash.
// Example:
// ```
// re.MustCompilePOSIX("^[a-z]+$")
// ```
func _mustCompilePOSIX(str string) *regexp.Regexp {
	return regexp.MustCompilePOSIX(str)
}

var RegexpExport = map[string]interface{}{
	"QuoteMeta":        _quoteMeta,
	"Compile":          _compile,
	"CompilePOSIX":     _compilePOSIX,
	"MustCompile":      _mustCompile,
	"MustCompilePOSIX": _mustCompilePOSIX,

	"Match":                _reMatch,
	"Grok":                 Grok,
	"ExtractIPv4":          RegexpMatchIPv4,
	"ExtractIPv6":          RegexpMatchIPv6,
	"ExtractIP":            RegexpMatchIP,
	"ExtractEmail":         RegexpMatchEmail,
	"ExtractPath":          RegexpMatchPathParam,
	"ExtractTTY":           RegexpMatchTTY,
	"ExtractURL":           RegexpMatchURL,
	"ExtractHostPort":      RegexpMatchHostPort,
	"ExtractMac":           RegexpMatchMac,
	"Find":                 _find_extractByRegexp,
	"FindIndex":            _findIndex_extractByRegexp,
	"FindAll":              _findAll_extractByRegexp,
	"FindAllIndex":         _findAllIndex_extractByRegexp,
	"FindSubmatch":         _findSubmatch_extractByRegexp,
	"FindSubmatchIndex":    _findSubmatchIndex_extractByRegexp,
	"FindSubmatchAll":      _findSubmatchAll_extractByRegexp,
	"FindSubmatchAllIndex": _findSubmatchAllIndex_extractByRegexp,
	"FindGroup":            reExtractGroups,
	"FindGroupAll":         reExtractGroupsAll,
	"ReplaceAll":           _replaceAll_extractByRegexp,
	"ReplaceAllWithFunc":   _replaceAllFunc_extractByRegexp,
}
