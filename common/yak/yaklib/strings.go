package yaklib

import (
	"encoding/json"
	"github.com/yaklang/yaklang/common/xhtml"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/yaklang/yaklang/common/domainextractor"
	"github.com/yaklang/yaklang/common/filter"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/jsonextractor"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/utils/network"
	"github.com/yaklang/yaklang/common/utils/suspect"
)

// IndexAny Returns the index of the first occurrence of any character of chars in string s, if chars does not exist in the string , then return -1
// Example:
// ```
// str.IndexAny("Hello world", "world") // 2 because l occurs first in third character
// str.IndexAny("Hello World", "Yak") // -1
// ```
func IndexAny(s string, chars string) int {
	return strings.IndexAny(s, chars)
}

// StartsWith / HasPrefix determines whether string s begins with prefix
// Example:
// ```
// str.StartsWith("Hello Yak", "Hello") // true
// str.StartsWith("Hello Yak", "Yak") // false
// ```
func StartsWith(s string, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// EndsWith / HasSuffix. Determine whether the string s ends with suffix.
// Example:
// ```
// str.EndsWith("Hello Yak", "Yak") // true
// str.EndsWith("Hello Yak", "Hello") // false
// ```
func EndsWith(s string, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Title returns the titled version of string s, that is, the first letters of all words are capitalized
// Example:
// ```
// str.Title("hello yak") // Hello Yak
// ```
func Title(s string) string {
	return strings.Title(s)
}

// Join connects the elements in i with d. If passed in The parameter is not a string, it will be automatically converted to a string, and then connected with d. If the connection fails, the string form of i will be returned.
// Example:
// ```
// str.Join([]string{"hello", "yak"}, " ") // hello yak
// str.Join([]int{1, 2, 3}, " ") // 1 2 3
// ```
func Join(i interface{}, d interface{}) (defaultResult string) {
	s := utils.InterfaceToString(d)
	defaultResult = utils.InterfaceToString(i)
	defer func() {
		recover()
	}()
	defaultResult = strings.Join(funk.Map(i, func(element interface{}) string {
		return utils.InterfaceToString(element)
	}).([]string), s)
	return
}

// TrimLeft Returns a string that removes all characters in the cutset string on the left side of string s
// Example:
// ```
// str.TrimLeft("Hello Yak", "H") // ello Yak
// str.TrimLeft("HelloYak", "Hello") // Yak
// ```
func TrimLeft(s string, cutset string) string {
	return strings.TrimLeft(s, cutset)
}

// that reads data from string s TrimPrefix Returns a string with the prefix removed from string s
// Example:
// ```
// str.TrimPrefix("Hello Yak", "Hello") //  Yak
// str.TrimPrefix("HelloYak", "Hello") // Yak
// ```
func TrimPrefix(s string, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// TrimRight Returns a string that removes all characters in the cutset string on the right side of string s
// Example:
// ```
// str.TrimRight("Hello Yak", "k") // Hello Ya
// str.TrimRight("HelloYak", "Yak") // Hello
// ```
func TrimRight(s string, cutset string) string {
	return strings.TrimRight(s, cutset)
}

// TrimSuffix Returns the string
// Example:
// ```
// str.TrimSuffix("Hello Yak", "ak") // Hello Y
// str.TrimSuffix("HelloYak", "Yak") // Hello
// ```
func TrimSuffix(s string, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// Trim returns a string that removes all characters in the cutset string on both sides of string s.
// Example:
// ```
// str.Trim("Hello Yak", "Hk") // ello Ya
// str.Trim("HelloYakHello", "Hello") // Yak
// ```
func Trim(s string, cutset string) string {
	return strings.Trim(s, cutset)
}

// TrimSpace Returns A string with all whitespace characters on both sides of the string s removed.
// Example:
// ```
// str.TrimSpace(" \t\n Hello Yak \n\t\r\n") // Hello Yak
// ```
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Split Splits string s into string slices according to sep
// Example:
// ```
// str.Split("Hello Yak", " ") // [Hello", "Yak"]
// ```
func Split(s string, sep string) []string {
	return strings.Split(s, sep)
}

// SplitAfter Split the string s into string slices according to sep, but each element will retain sep.
// Example:
// ```
// str.SplitAfter("Hello-Yak", "-") // [Hello-", "Yak"]
// ```
func SplitAfter(s string, sep string) []string {
	return strings.SplitAfter(s, sep)
}

// SplitAfterN Split the string s into string slices according to sep, but each element will retain sep, and can be divided into n elements at most
// Example:
// ```
// str.SplitAfterN("Hello-Yak-and-World", "-", 2) // [Hello-", "Yak-and-World"]
// ```
func SplitAfterN(s string, sep string, n int) []string {
	return strings.SplitAfterN(s, sep, n)
}

// SplitN Compares the string s according to sep. Split into string slices of up to n elements
// Example:
// ```
// str.SplitN("Hello-Yak-and-World", "-", 2) // [Hello", "Yak-and-World"]
// ```
func SplitN(s string, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

// ToLower Returns the lowercase form of string s
// Example:
// ```
// str.ToLower("HELLO YAK") // hello yak
// ```
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper Returns the uppercase version of string s
// Example:
// ```
// str.ToUpper("hello yak") // HELLO YAK
// ```
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// Repeat Returns a string that repeats string s count times
// Example:
// ```
// str.Repeat("hello", 3) // hellohellohello
// ```
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// ToTitle Returns the titled version of string s, where all Unicode letters are converted to their uppercase
// Example:
// ```
// str.ToTitle("hello yak") // HELLO YAK
// ```
func ToTitle(s string) string {
	return strings.ToTitle(s)
}

// Contains Judgment string Whether s contains substr
// Example:
// ```
// str.Contains("hello yakit", "yak") // true
// ```
func Contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}

// ReplaceAll Returns a string that replaces all old strings in string s with new strings
// Example:
// ```
// str.ReplaceAll("hello yak", "yak", "yakit") // hello yakit
// ```
func ReplaceAll(s string, old string, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Replace returns the prefix of string s String that replaces n old strings with new strings
// Example:
// ```
// str.Replace("hello yak", "l", "L", 1) // heLlo yak
// ```
func Replace(s string, old string, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// NewReader Returns a*Reader
// Example:
// ```
// r = str.NewReader("hello yak")
// buf = make([]byte, 256)
// _, err = r.Read(buf)
// die(err)
// println(sprintf("%s", buf)) // hello yak
// ```
func NewReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

// if the conversion fails Index returns the index of the first occurrence of substr in string s. If substr does not exist in the string, it returns -1.
// Example:
// ```
// str.Index("hello yak", "yak") // 6
// str.Index("hello world", "yak") // -1
// ```
func Index(s string, substr string) int {
	return strings.Index(s, substr)
}

// Count returns the number of occurrences of substr in string s.
// Example:
// ```
// str.Count("hello yak", "l") // 2
// ```
func Count(s string, substr string) int {
	return strings.Count(s, substr)
}

// Compare Compares each character in the strings a and b one by one according to the order of the ascii code table. If a==b, returns 0. If a<b, then return -1, if a>b of replacement, 1 will be returned.
// Example:
// ```
// str.Compare("hello yak", "hello yak") // 0
// str.Compare("hello yak", "hello") // 1
// str.Compare("hello", "hello yak") // -1
// ```
func Compare(a string, b string) int {
	return strings.Compare(a, b)
}

// ContainsAny Determines whether the string s contains any character in chars
// Example:
// ```
// str.ContainsAny("hello yak", "ly") // true
// str.ContainsAny("hello yak", "m") // false
// ```
func ContainsAny(s string, chars string) bool {
	return strings.ContainsAny(s, chars)
}

// EqualFold Determines whether strings s and t are equal, ignoring case
// Example:
// ```
// str.EqualFold("hello Yak", "HELLO YAK") // true
// ```
func EqualFold(s string, t string) bool {
	unicode.IsSpace('a')
	return strings.EqualFold(s, t)
}

// Fields. Returns a string slice that divides the string s according to whitespace characters ('\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0).
// Example:
// ```
// str.Fields("hello world\nhello yak\tand\vyakit") // [hello", "world", "hello", "yak", "and", "yakit"]
// ```
func Fields(s string) []string {
	return strings.Fields(s)
}

// IndexByte Returns the index of the first character equal to c in string s. If c does not exist in the string, returns -1
// Example:
// ```
// str.IndexByte("hello yak", 'y') // 6
// str.IndexByte("hello yak", 'm') // -1
// ```
func IndexByte(s string, c byte) int {
	return strings.IndexByte(s, c)
}

// LastIndex Returns a string The index of the last occurrence of substr in s. If substr does not exist in the string, -1 is returned.
// Example:
// ```
// str.LastIndex("hello yak", "l") // 3
// str.LastIndex("hello yak", "m") // -1
// ```
func LastIndex(s string, substr string) int {
	return strings.LastIndex(s, substr)
}

// LastIndexAny. Return characters. The index of the last occurrence of any character of chars in string s. If chars does not exist in the string, -1 is returned.
// Example:
// ```
// str.LastIndexAny("hello yak", "ly") // 6
// str.LastIndexAny("hello yak", "m") // -1
// ```
func LastIndexAny(s string, chars string) int {
	return strings.LastIndexAny(s, chars)
}

// with the suffix suffix removed from the string s. LastIndexByte Returns the index of the last character equal to c in the string s. If the character If c does not exist in the string, then -1 is returned.
// Example:
// ```
// str.LastIndexByte("hello yak", 'l') // 3
// str.LastIndexByte("hello yak", 'm') // -1
// ```
func LastIndexByte(s string, c byte) int {
	return strings.LastIndexByte(s, c)
}

// ToValidUTF8 returns the invalid UTF-8 in string s. If the encoding is replaced with the string
// Example:
// ```
//
// str.ToValidUTF8("hello yak", "?") // hello yak
// ```
func ToValidUTF8(s string, replacement string) string {
	return strings.ToValidUTF8(s, replacement)
}

// ExtractJson Try to extract the JSON in the string and repair it Return
// Example:
// ```
// str.ExtractJson("hello yak") // []
// str.ExtractJson(`{"hello": "yak"}`) // [{"hello": "yak"}]
// ```
func extractValidJson(i interface{}) []string {
	return jsonextractor.ExtractStandardJSON(utils.InterfaceToString(i))
}

// ExtractJsonWithRaw Try to extract the JSON in the string and return it, the first return value returns the repaired JSON string Array, the second return value returns the original JSON string array (if repair fails)
// Example:
// ```
// str.ExtractJsonWithRaw("hello yak") // [], []
// str.ExtractJsonWithRaw(`{"hello": "yak"}`) // [{"hello": "yak"}], []
// ```
func extractJsonEx(i interface{}) ([]string, []string) {
	return jsonextractor.ExtractJSONWithRaw(utils.InterfaceToString(i))
}

// ExtractDomain Try to extract the JSON in the string Domain name and returns
// Example:
// ```
// str.ExtractDomain("hello yak") // []
// str.ExtractDomain("hello yaklang.com or yaklang.io") // ["yaklang.com", "yaklang.io"]
// ```
func extractDomain(i interface{}) []string {
	return domainextractor.ExtractDomains(utils.InterfaceToString(i))
}

// ExtractRootDomain Try to extract the root domain name in the string and return
// Example:
// ```
// str.ExtractRootDomain("hello yak") // []
// str.ExtractRootDomain("hello www.yaklang.com or www.yaklang.io") // ["yaklang.com", "yaklang.io"]
// ```
func extractRootDomain(i interface{}) []string {
	return domainextractor.ExtractRootDomains(utils.InterfaceToString(i))
}

// ExtractTitle Try to parse the incoming string into HTML and extract the title (title tag). Return
// Example:
// ```
// str.ExtractTitle("hello yak") // ""
// str.ExtractTitle("<title>hello yak</title>") // "hello yak"
// ```
func extractTitle(i interface{}) string {
	return utils.ExtractTitleFromHTMLTitle(utils.InterfaceToString(i), "")
}

// PathJoin concatenates the incoming file paths and returns
// Example:
// ```
// str.PathJoin("/var", "www", "html") // in *unix: "/var/www/html"    in Windows: \var\www\html
// ```
func pathJoin(elem ...string) (newPath string) {
	return filepath.Join(elem...)
}

// ToJsonIndentStr Convert v to a formatted JSON string and return, or the empty string
// Example:
// ```
// str.ToJsonIndentStr({"hello":"yak"}) // {"hello": "yak"}
// ```
func toJsonIndentStr(d interface{}) string {
	raw, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return ""
	}
	return string(raw)
}

var StringsExport = map[string]interface{}{
	// Basic string tools
	"IndexAny":       IndexAny,
	"StartsWith":     StartsWith,
	"EndsWith":       EndsWith,
	"Title":          Title,
	"Join":           Join,
	"TrimLeft":       TrimLeft,
	"TrimPrefix":     TrimPrefix,
	"TrimRight":      TrimRight,
	"TrimSuffix":     TrimSuffix,
	"Trim":           Trim,
	"TrimSpace":      TrimSpace,
	"Split":          Split,
	"SplitAfter":     SplitAfter,
	"SplitAfterN":    SplitAfterN,
	"SplitN":         SplitN,
	"ToLower":        ToLower,
	"ToUpper":        ToUpper,
	"HasPrefix":      StartsWith,
	"HasSuffix":      EndsWith,
	"Repeat":         Repeat,
	"ToTitleSpecial": strings.ToTitleSpecial,
	"ToTitle":        ToTitle,
	"Contains":       Contains,
	"ReplaceAll":     ReplaceAll,
	"Replace":        Replace,
	"NewReader":      strings.NewReader,
	"Index":          Index,
	"Count":          Count,
	"Compare":        Compare,
	"ContainsAny":    ContainsAny,
	"EqualFold":      EqualFold,
	"Fields":         Fields,
	"IndexByte":      IndexByte,
	"LastIndex":      LastIndex,
	"LastIndexAny":   LastIndexAny,
	"LastIndexByte":  LastIndexByte,
	"ToLowerSpecial": strings.ToLowerSpecial,
	"ToUpperSpecial": strings.ToUpperSpecial,
	"ToValidUTF8":    ToValidUTF8,

	// The unique
	"RandStr":                utils.RandStringBytes,
	"Random":                 xhtml.RandomUpperAndLower,
	"f":                      _sfmt,
	"SplitAndTrim":           utils.PrettifyListFromStringSplited,
	"StringSliceContains":    utils.StringSliceContain,
	"StringSliceContainsAll": utils.StringSliceContainsAll,
	"RemoveRepeat":           utils.RemoveRepeatStringSlice,
	"RandSecret":             utils.RandSecret,
	"IsStrongPassword":       utils.IsStrongPassword,
	"ExtractStrContext":      utils.ExtractStrContextByKeyword,

	// Supports url and host:port Parsed into Host Port
	"CalcSimilarity":                    utils.CalcSimilarity,
	"CalcTextMaxSubStrStability":        utils.CalcTextSubStringStability,
	"CalcSSDeepStability":               utils.CalcSSDeepStability,
	"CalcSimHashStability":              utils.CalcSimHashStability,
	"CalcSimHash":                       utils.SimHash,
	"CalcSSDeep":                        utils.SSDeepHash,
	"ParseStringToHostPort":             utils.ParseStringToHostPort,
	"IsIPv6":                            utils.IsIPv6,
	"IsIPv4":                            utils.IsIPv4,
	"StringContainsAnyOfSubString":      utils.StringContainsAnyOfSubString,
	"ExtractHost":                       utils.ExtractHost,
	"ExtractDomain":                     extractDomain,
	"ExtractRootDomain":                 extractRootDomain,
	"ExtractJson":                       extractValidJson,
	"ExtractJsonWithRaw":                extractJsonEx,
	"LowerAndTrimSpace":                 utils.StringLowerAndTrimSpace,
	"HostPort":                          utils.HostPort,
	"ParseStringToHTTPRequest":          lowhttp.ParseStringToHttpRequest,
	"SplitHostsToPrivateAndPublic":      utils.SplitHostsToPrivateAndPublic,
	"ParseBytesToHTTPRequest":           lowhttp.ParseBytesToHttpRequest,
	"ParseStringToHTTPResponse":         lowhttp.ParseStringToHTTPResponse,
	"ParseBytesToHTTPResponse":          lowhttp.ParseBytesToHTTPResponse,
	"FixHTTPResponse":                   lowhttp.FixHTTPResponse,
	"ExtractBodyFromHTTPResponseRaw":    lowhttp.ExtractBodyFromHTTPResponseRaw,
	"FixHTTPRequest":                    lowhttp.FixHTTPRequest,
	"ExtractURLFromHTTPRequestRaw":      lowhttp.ExtractURLFromHTTPRequestRaw,
	"ExtractURLFromHTTPRequest":         lowhttp.ExtractURLFromHTTPRequest,
	"ExtractTitle":                      extractTitle,
	"SplitHTTPHeadersAndBodyFromPacket": lowhttp.SplitHTTPHeadersAndBodyFromPacket,
	"MergeUrlFromHTTPRequest":           lowhttp.MergeUrlFromHTTPRequest,
	"ReplaceHTTPPacketBody":             lowhttp.ReplaceHTTPPacketBody,

	"ParseStringToHosts":              utils.ParseStringToHosts,
	"ParseStringToPorts":              utils.ParseStringToPorts,
	"ParseStringToUrls":               utils.ParseStringToUrls,
	"ParseStringToUrlsWith3W":         utils.ParseStringToUrlsWith3W,
	"ParseStringToCClassHosts":        network.ParseStringToCClassHosts,
	"ParseStringUrlToWebsiteRootPath": utils.ParseStringUrlToWebsiteRootPath,
	"ParseStringUrlToUrlInstance":     utils.ParseStringUrlToUrlInstance,
	"UrlJoin":                         utils.UrlJoin,
	"IPv4ToCClassNetwork":             utils.GetCClassByIPv4,
	"ParseStringToLines":              utils.ParseStringToLines,
	"PathJoin":                        pathJoin,
	"Grok":                            Grok,
	"JsonToMapList":                   JsonToMapList,
	// "JsonStreamToMapList":             JsonStreamToMapList,
	"JsonToMap":       JsonToMap,
	"ParamsGetOr":     ParamsGetOr,
	"ToJsonIndentStr": toJsonIndentStr,

	"NewFilter":            filter.NewFilter,
	"RemoveDuplicatePorts": filter.RemoveDuplicatePorts,
	"FilterPorts":          filter.FilterPorts,

	"RegexpMatch": _strRegexpMatch,

	"MatchAllOfRegexp":    utils.MatchAllOfRegexp,
	"MatchAllOfGlob":      utils.MatchAllOfGlob,
	"MatchAllOfSubString": utils.MatchAllOfSubString,
	"MatchAnyOfRegexp":    utils.MatchAnyOfRegexp,
	"MatchAnyOfGlob":      utils.MatchAnyOfGlob,
	"MatchAnyOfSubString": utils.MatchAnyOfSubString,

	"IntersectString":     funk.IntersectString,
	"Intersect":           funk.IntersectString,
	"Subtract":            funk.SubtractString,
	"ToStringSlice":       utils.InterfaceToStringSlice,
	"VersionGreater":      utils.VersionGreater,
	"VersionGreaterEqual": utils.VersionGreaterEqual,
	"VersionEqual":        utils.VersionEqual,
	"VersionLessEqual":    utils.VersionLessEqual,
	"VersionLess":         utils.VersionLess,
	"VersionCompare":      utils.VersionCompare,
	"Cut":                 strings.Cut,
	"CutPrefix":           strings.CutPrefix,
	"CutSuffix":           strings.CutSuffix,
}

func init() {
	for k, v := range suspect.GuessExports {
		StringsExport[k] = v
	}
}
