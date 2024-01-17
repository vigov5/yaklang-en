package suspect

import (
	"bytes"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// ref: https://portswigger.net/blog/json-hijacking-for-the-modern-web
//
// Judgment logic
// 1. get method
// 2. There are callback, cb, jsonp parameter
// 3. (nosniff = true && content-type = js) || (nosniff = false && content-type maybe js)
// 4. Cannot be {, <, [, " at the beginning
// 5. Contains (or =
// 6. Important! Contains sensitive data, username, ip, etc.
// 7. This function Used for preliminary screening (Check function), the specific vulnerability is determined in the jsonp package
func IsSensitiveJSONP(reqRaw []byte, rspRaw []byte) bool {
	freq, _ := mutate.NewFuzzHTTPRequest(reqRaw)
	if freq != nil {
		if len(freq.GetGetQueryParams()) <= 0 {
			return false
		}
	}

	resp, err := lowhttp.ParseBytesToHTTPResponse(rspRaw)
	if err != nil {
		return false
	}

	contentType := strings.TrimSpace(strings.ToLower(resp.Header.Get("Content-Type")))
	// https://github.com/chromium/chromium/blob/fc262dcd403c74cf3e22896f32d9723ba463f0b6/third_party/blink/common/mime_util/mime_util.cc#L42
	jsContentTypes := []string{
		"application/ecmascript",
		"application/javascript",
		"application/x-ecmascript",
		"application/x-javascript",
		"text/ecmascript",
		"text/javascript",
		"text/javascript1.0",
		"text/javascript1.1",
		"text/javascript1.2",
		"text/javascript1.3",
		"text/javascript1.4",
		"text/javascript1.5",
		"text/jscript",
		"text/livescript",
		"text/x-ecmascript",
		"text/x-javascript",
	}
	maybeJSContentTypes := []string{
		"text/html",
		"text/plain",
		"application/json",
		"text/json",
	}

	if !utils.StringHasPrefix(contentType, jsContentTypes) {
		// other content-types under nosniff cannot be executed
		if resp.Header.Get("X-Content-Type-Options") == "nosniff" {
			return false
		}
		// adds some content that is common when there is no nosniff- type
		if contentType != "" && !utils.StringHasPrefix(contentType, maybeJSContentTypes) {
			return false
		}
	}

	_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(rspRaw)
	rest := bytes.TrimLeft(body, "\t\n\v\f\r \x85\xa0")
	if len(rest) <= 0 {
		return false
	}
	switch rest[0] {
	case '{', '<', '[', '"':
		return false
	}
	return IsSensitiveJSON(body)
}

// IsHTMLResponse determines whether the response is in html format
// 1. response content-type
// 2. check fist 500 bytes
func IsHTMLResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return true
	}
	if resp.Header.Get("X-Content-Type-Options") == "nosniff" {
		return false
	}

	var right = 0
	var raw []byte
	var err error
	if resp.Body != nil {
		raw, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		right = len(raw) - 1
	}
	if right <= 0 {
		return false
	}
	if right > 500 {
		right = 500
	}
	toBeTested := raw[:right]
	if bytes.Contains(toBeTested, []byte("<html")) ||
		bytes.Contains(toBeTested, []byte("<head")) ||
		bytes.Contains(toBeTested, []byte("<body")) {
		return true
	}
	return false
}

func HaveServerError(body []byte) bool {
	bodyStr := string(body)
	if !maybeServerErrorPageKeyword.MatchString(bodyStr) {
		return false
	}
	bodyStr = html.UnescapeString(bodyStr)
	for _, regex := range []*regexp.Regexp{maybePHPErrorRegex, maybeJVMStackTraceRegex, maybePythonStackTraceRegex,
		maybeJSStackTraceRegex, maybeCSStackTraceRegex, maybeGOStackTraceRegex} {
		if regex.MatchString(bodyStr) {
			return true
		}
	}
	return false
}

func SearchChineseIDCards(data []byte) []string {
	coefficient := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	code := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	ret := make([]string, 0, 10)
	m := maybeChinaIDCardNumberRegex.FindAllSubmatch(data, 10)
	// https://github.com/afanti-com/utils-go/blob/master/idCardNo/idCardNo.go
	// in the query. Make sure it is an ID number to avoid false positives.
	for _, item := range m {
		if len(item) >= 1 {
			number := item[0]
			if number[len(number)-1] == 'x' {
				number[len(number)-1] = 'X'
			}
			sum := 0
			for i := 0; i < 17; i++ {
				sum += int(number[i]-byte('0')) * coefficient[i]
			}
			if code[sum%11] == number[17] {
				ret = append(ret, string(number))
			}
		}
	}
	return ret
}

func IsJsonResponse(resp *http.Response) bool {
	mayBeJSONType := []string{
		"application/json",
		"text/json",
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		return utils.StringArrayContains(mayBeJSONType, contentType)
	}
	return false
}

func IsJsonResponseRaw(resp []byte) bool {
	resp, _, err := lowhttp.FixHTTPResponse(resp)
	if err != nil {
		return false
	}
	r, err := lowhttp.ParseBytesToHTTPResponse(resp)
	if err != nil {
		return false
	}
	return IsJsonResponse(r)
}
