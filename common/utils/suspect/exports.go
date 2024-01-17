package suspect

import (
	"net/http"
	"reflect"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
)

// IsAlpha tries to convert the incoming parameters into a string, and then determines whether they are all composed of English letters.
// Example:
// ```
// str.IsAlpha("abc") // true
// str.IsAlpha("abc123") // false
// ```
func isAlpha(i interface{}) bool {
	return utils.MatchAllOfRegexp(i, `[a-zA-Z]+`)
}

// IsDigit tries to convert the incoming parameter into a string, and then determines its Whether they are all composed of numbers
// Example:
// ```
// str.IsDigit("123") // true
// str.IsDigit("abc123") // false
// ```
func isDigit(i interface{}) bool {
	return utils.MatchAllOfRegexp(i, `[0-9]+`)
}

// IsAlphaNum / IsAlNum tries to convert the incoming parameters into strings, and then determines whether they are all composed of
// Example:
// ```
// str.IsAlphaNum("abc123") // true
// str.IsAlphaNum("abc123!") // false
// ```
func isAlphaNum(i interface{}) bool {
	return utils.MatchAllOfRegexp(i, `[a-zA-Z0-9]+`)
}

// IsTLSServer attempts to access the incoming parameters. host, and then determine whether it is a TLS service. The first parameter is host, and zero or more parameters can be passed in later, which is the proxy address.
// Example:
// ```
// str.IsTLSServer("www.yaklang.com:443") // true
// str.IsTLSServer("www.yaklang.com:80") // false
// ```
func isTLSServer(addr string, proxies ...string) bool {
	return netx.IsTLSService(addr, proxies...)
}

// IsHttpURL Try to convert the incoming parameters into a string, and then guess whether it is a URL of the http(s) protocol
// Example:
// ```
// str.IsHttpURL("http://www.yaklang.com") // true
// str.IsHttpURL("https://www.yaklang.com") // true
// str.IsHttpURL("www.yaklang.com") // false
// ```
func isHttpURL(i interface{}) bool {
	return IsFullURL(i)
}

// IsUrlPath Try to convert the incoming parameter to String and then guess if it is a URL Path
// Example:
// ```
// str.IsUrlPath("/index.php") // true
// str.IsUrlPath("index.php") // false
// ```
func isUrlPath(i interface{}) bool {
	return IsURLPath(i)
}

// IsHtmlResponse Guess whether the incoming parameter is an original HTTP response message
// Example:
// ```
// str.IsHtmlResponse("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html></html>") // true
// resp, _ = str.ParseStringToHTTPResponse("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html></html>")
// str.IsHtmlResponse(resp) // true
// ```
func isHtmlResponse(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		rsp, err := lowhttp.ParseBytesToHTTPResponse([]byte(ret))
		if err != nil {
			log.Error(err)
			return false
		}
		return IsHTMLResponse(rsp)
	case []byte:
		rsp, err := lowhttp.ParseBytesToHTTPResponse(ret)
		if err != nil {
			log.Error(err)
			return false
		}
		return IsHTMLResponse(rsp)
	case *http.Response:
		return IsHTMLResponse(ret)
	default:
		log.Errorf("need []byte/string/*http.Response but got %s", reflect.TypeOf(ret))
		return false
	}
}

// IsServerError Guess whether the incoming parameter is a server error
// Example:
// ```
// str.IsServerError(`Fatal error: Uncaught Error: Call to undefined function sum() in F:\xampp\htdocs\test.php:7 Stack trace: #0 {main} thrown in <path> on line 7`) // true, this is the PHP error message
// ```
func isServerError(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return HaveServerError(utils.UnsafeStringToBytes(ret))
	case []byte:
		return HaveServerError(ret)
	default:
		return HaveServerError(utils.InterfaceToBytes(ret))
	}
}

// ExtractChineseIDCards tries to convert the incoming parameter into a string, and then extracts the ID number in the string.
// Example:
// ```
// str.ExtractChineseIDCards("Your ChineseID is: 110101202008027420?") // ["110101202008027420"]
// ```
func extractChineseIDCards(i interface{}) []string {
	switch ret := i.(type) {
	case string:
		return SearchChineseIDCards(utils.UnsafeStringToBytes(ret))
	case []byte:
		return SearchChineseIDCards(ret)
	default:
		return SearchChineseIDCards(utils.InterfaceToBytes(ret))
	}
}

// IsJsonResponse attempts to convert the incoming parameters into strings, and then guesses whether the incoming parameters are original HTTP response messages in JSON format. This is achieved by judging the Content-Type request header.
// Example:
// ```
// str.IsJsonResponse("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"code\": 0}") // true
// str.IsJsonResponse("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\nhello") // false
// ```
func isJsonResponse(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		rsp, err := lowhttp.ParseBytesToHTTPResponse(utils.UnsafeStringToBytes(ret))
		if err != nil {
			log.Error(err)
			return false
		}
		return IsJsonResponse(rsp)
	case []byte:
		rsp, err := lowhttp.ParseBytesToHTTPResponse(ret)
		if err != nil {
			log.Error(err)
			return false
		}
		return IsJsonResponse(rsp)
	case *http.Response:
		return IsJsonResponse(ret)
	default:
		log.Errorf("need []byte/string/*http.Response but got %s", reflect.TypeOf(ret))
		return false
	}
}

// IsRedirectParam guesses whether it is a redirection parameter based on the passed parameter name and parameter value.
// Example:
// ```
// str.IsRedirectParam("to","http://www.yaklang.com") // true because the parameter value is a complete URL
// str.IsRedirectParam("target","/index.php") // true, because the parameter value is a URL path and the parameter name is a common jump parameter name
// str.IsRedirectParam("id", "1") // false
// ```
func isRedirectParam(key, value string) bool {
	return BeUsedForRedirect(key, value)
}

// IsJSONPParam guesses whether it is a JSONP parameter based on the parameter name and value passed in.
// Example:
// ```
// str.IsJSONPParam("callback","jquery1.0.min.js") // true, because the parameter name is a common JSONP parameter name, and the parameter value is a common JS File name
// str.IsJSONPParam("f","jquery1.0.min.js") // true, because the parameter value is a common JS file name
// str.IsJSONPParam("id","1") // false
// ```
func isJSONPParam(key, value string) bool {
	return IsJSONPParam(key, value)
}

// composed of English letters and numbers. IsUrlParam guesses whether it is a URL parameter based on the incoming parameter name and parameter value.
// Example:
// ```
// str.IsUrlParam("url","http://www.yaklang.com") // true, because the parameter name is a common URL parameter name, and the parameter value is a complete URL
// str.IsUrlParam("id","1") // false
// ```
func isURLParam(key, value string) bool {
	return IsGenericURLParam(key, value)
}

// IsXmlParam guesses whether it is an XML parameter based on the incoming parameter name and parameter value
// Example:
// ```
// str.IsXmlParam("xml","<xml></xml>") // true, because the parameter name is common. XML parameter name, and the parameter value is a string in XML format
// str.IsXmlParam("X","<xml></xml>") // true because the parameter value is an XML format string
// str.IsXmlParam("id","1") // false
// ```
func isXMLParam(key, value string) bool {
	return IsXMLParam(key, value)
}

// IsSensitiveJson tries to convert the incoming parameter into a string, and then guesses whether it is sensitive JSON data
// Example:
// ```
// str.IsSensitiveJson(`{"password":"123456"}`) // true
// str.IsSensitiveJson(`{"uid": 10086}`) // true
// str.IsSensitiveJson(`{"id": 1}`) // false
// ```
func isSensitiveJson(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsSensitiveJSON(utils.UnsafeStringToBytes(ret))
	case []byte:
		return IsSensitiveJSON(ret)
	default:
		return IsSensitiveJSON(utils.InterfaceToBytes(ret))
	}
}

// IsSensitiveTokenField tries to convert the incoming parameters into strings, and then guesses whether they are tokens. Field
// Example:
// ```
// str.IsSensitiveTokenField("token") // true
// str.IsSensitiveTokenField("access_token") // true
// str.IsSensitiveTokenField("id") // false
// ```
func isSensitiveTokenField(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsTokenParam(ret)
	case []byte:
		return IsTokenParam(utils.UnsafeBytesToString(ret))
	default:
		return IsTokenParam(utils.InterfaceToString(ret))
	}
}

// IsPasswordField tries to convert the incoming parameter into a string, and then guesses whether it is a password field.
// Example:
// ```
// str.IsPasswordField("password") // true
// str.IsPasswordField("pwd") // true
// str.IsPasswordField("id") // false
// ```
func isPasswordField(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsPasswordKey(ret)
	case []byte:
		return IsPasswordKey(utils.UnsafeBytesToString(ret))
	default:
		return IsPasswordKey(utils.InterfaceToString(ret))
	}
}

// IsUsernameField Try to convert the incoming parameter into a string, Then guess whether it is the username field
// Example:
// ```
// str.IsUsernameField("username") // true
// str.IsUsernameField("user") // true
// str.IsUsernameField("id") // false
// ```
func isUsernameField(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsUsernameKey(ret)
	case []byte:
		return IsUsernameKey(utils.UnsafeBytesToString(ret))
	default:
		return IsUsernameKey(utils.InterfaceToString(ret))
	}
}

// IsSQLColumnField Try to convert the incoming parameter to a string and then guess if it is SQL query field
// Example:
// ```
// str.IsSQLColumnField("sort") // true
// str.IsSQLColumnField("order") // true
// str.IsSQLColumnField("id") // false
// ```
func isSQLColumnField(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsSQLColumnName(ret)
	case []byte:
		return IsSQLColumnName(utils.UnsafeBytesToString(ret))
	default:
		return IsSQLColumnName(utils.InterfaceToString(ret))
	}
}

// IsCaptchaField tries to convert the incoming parameter into a string, and then guesses whether it is Is the verification code field
// Example:
// ```
// str.IsCaptchaField("captcha") // true
// str.IsCaptchaField("code_img") // true
// str.IsCaptchaField("id") // false
// ```
func isCaptchaField(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsCaptchaKey(ret)
	case []byte:
		return IsCaptchaKey(utils.UnsafeBytesToString(ret))
	default:
		return IsCaptchaKey(utils.InterfaceToString(ret))
	}
}

// IsBase64Value tries to convert the incoming parameter into a string, and then guesses whether it is Base64 encoded data.
// Example:
// ```
// str.IsBase64Value("MTI=") // true
// str.IsBase64Value("123") // false
// ```
func isBase64Value(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsBase64(ret)
	case []byte:
		return IsBase64(utils.UnsafeBytesToString(ret))
	default:
		return IsBase64(utils.InterfaceToString(ret))
	}
}

// IsPlainBase64Value tries to convert the passed parameters into a string. Then guess whether it is Base64-encoded data. It has one more layer of judgment compared to IsBase64Value, that is, whether the base64-decoded data is a visible string.
// Example:
// ```
// str.IsPlainBase64Value("MTI=") // true
// str.IsPlainBase64Value("Aw==") // false
// ```
func isPlainBase64Value(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsBase64Password(ret)
	case []byte:
		return IsBase64Password(utils.UnsafeBytesToString(ret))
	default:
		return IsBase64Password(utils.InterfaceToString(ret))
	}
}

// IsMD5Value tries to convert the incoming parameter into a string, and then guesses whether it is MD5 Encoded data
// Example:
// ```
// str.IsMD5Value("202cb962ac59075b964b07152d234b70") // true
// str.IsMD5Value("123") // false
// ```
func isMD5Value(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsMD5Data(ret)
	case []byte:
		return IsMD5Data(utils.UnsafeBytesToString(ret))
	default:
		return IsMD5Data(utils.InterfaceToString(ret))
	}
}

// IsSha256Value Try to convert the incoming parameter to a string and then guess if it is SHA256 encoded data
// Example:
// ```
// str.IsSha256Value("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3") // true
// str.IsSha256Value("123") // false
// ```
func isSha256Value(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsSHA256Data(ret)
	case []byte:
		return IsSHA256Data(utils.UnsafeBytesToString(ret))
	default:
		return IsSHA256Data(utils.InterfaceToString(ret))
	}
}

// IsXmlRequest guesses whether the incoming parameter is a request header that is an original HTTP request message in XML format
// Example:
// ```
// str.IsXmlRequest("POST / HTTP/1.1\r\nContent-Type: application/xml\r\n\r\n<xml></xml>") // true
// str.IsXmlRequest("POST / HTTP/1.1\r\nContent-Type: text/html\r\n\r\n<html></html>") // false
// ```
func isXMLRequest(i interface{}) bool {
	switch ret := i.(type) {
	case []byte:
		return IsXMLRequest(ret)
	case string:
		return IsXMLRequest([]byte(ret))
	case *http.Request:
		raw, _ := utils.HttpDumpWithBody(i, true)
		return IsXMLRequest(raw)
	default:
		return false
	}
}

// IsXmlValue tries to convert the incoming parameter is a string, and then guesses whether it is XML format data.
// Example:
// ```
// str.IsXmlValue("<xml></xml>") // true
// str.IsXmlValue("<html></html>") // false
// ```
func isXmlValue(i interface{}) bool {
	switch ret := i.(type) {
	case string:
		return IsXMLString(ret)
	case []byte:
		return IsXMLBytes(ret)
	default:
		return false
	}
}

var GuessExports = map[string]interface{}{
	"IsAlpha":               isAlpha,
	"IsDigit":               isDigit,
	"IsAlphaNum":            isAlphaNum,
	"IsAlNum":               isAlphaNum,
	"IsTLSServer":           isTLSServer,
	"IsHttpURL":             IsFullURL,
	"IsUrlPath":             IsURLPath,
	"IsHtmlResponse":        isHtmlResponse,
	"IsServerError":         isServerError,
	"ExtractChineseIDCards": extractChineseIDCards,
	"IsJsonResponse":        isJsonResponse,
	"IsRedirectParam":       isRedirectParam,
	"IsJSONPParam":          isJSONPParam,
	"IsUrlParam":            isURLParam,
	"IsXmlParam":            isXMLParam,
	"IsSensitiveJson":       isSensitiveJson,
	"IsSensitiveTokenField": isSensitiveTokenField,
	"IsPasswordField":       isPasswordField,
	"IsUsernameField":       isUsernameField,
	"IsSQLColumnField":      isSQLColumnField,
	"IsCaptchaField":        isCaptchaField,
	"IsBase64Value":         isBase64Value,
	"IsPlainBase64Value":    isPlainBase64Value,
	"IsMD5Value":            isMD5Value,
	"IsSha256Value":         isSha256Value,
	"IsXmlRequest":          isXMLRequest,
	"IsXmlValue":            isXmlValue,
}
