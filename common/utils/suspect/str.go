package suspect

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"unicode"

	"github.com/yaklang/yaklang/common/utils/lowhttp"
)

var (
	maybeURLRegexStr     = `(url)|(addr)|(link)|(download)|(src)|(service)|(target)`
	maybeURLRegex        = regexp.MustCompile(`(?i)` + maybeURLRegexStr)
	maybeRedirectRegex   = regexp.MustCompile(`(?i)(next)|(goto)|(target)|(return)|(location)|(r_url)|` + maybeURLRegexStr)
	maybeJSONPRegex      = regexp.MustCompile(`(?i)(callback)|(jsonp)|(^cb$)|(function)`)
	maybeJSONPValue      = regexp.MustCompile(`(?i)(jquery\d)|(callback)|(jsonp)`)
	sensitiveJSONKeyList = []string{"uid", "user_id", "uin", "uname", "user", "username", "user_name", "nick", "unick",
		"phone", "mobile", "ip", "email", "password", "ticket", "secret", "token"}
	sensitiveJSONKeyRegex *regexp.Regexp
	maybeTokenRegex       = regexp.MustCompile("(?i)(token)|(secret)|(key$)|(^tk$)|(^ak$)")
	commonURLPathExtRegex = regexp.MustCompile(`(?i)(\.php)|(\.jsp)|(\.asp)|(\.aspx)|(\.html)|(\.action)|(\.do)`)
	maybeSQLColumnNameKey = regexp.MustCompile(`(?i)(sort)|(order)|(field)|(column)`)
	maybePasswordKeyRegex = regexp.MustCompile("(?i)(password)|(pass_word)|(pass_code)|(passcode)|(passw)|(pwd)|(psw)|(psd)|(pswd)|(passwd)|(mima)|(txtmm)|(yhmm)|(pass$)")
	maybeUsernameKeyRegex = regexp.MustCompile(`(?i)(^name$)|(uname)|(^uid$)|(^uin$)|(account)|(user_id)|(userid)|(txtuser$)|(nick)|(user$)`)
	maybeCaptchaKeyRegex  = regexp.MustCompile(`(?i)(captcha)|(vcode)|(v_code)|(yzm)|(yanzhengma)`)

	maybeServerErrorPageKeyword = regexp.MustCompile(`(?i)(stack\s?trace)|(exception)|(error)|(panic)|(warning)|(notice)`)
	// / or C:\ is the same. If there is an absolute path, you need to match
	maybePythonStackTraceRegex = regexp.MustCompile(`File "((/[a-zA-Z])|([c-zC-Z]:\\))(.+?).py"(, line \d+,)? in`)
	maybeJVMStackTraceRegex    = regexp.MustCompile(`(?s)Exception.+([aA]t )?(.+?)\((.+?.(java|kt):\d+)|(Native Method)\)`)
	maybeJSStackTraceRegex     = regexp.MustCompile(`at (.+?)\(((/[a-zA-Z])|([c-zC-Z]:\\))(.+?).js:\d+:\d+\)`)
	maybePHPErrorRegex         = regexp.MustCompile(`(?i)(Fatal error:.+?\.php:\d+)|(Notice:.+?\.php on line \d+)|(Warning:.+?\.php on line \d+)`)
	maybeCSStackTraceRegex     = regexp.MustCompile(`at (.+?) in (((/[a-zA-Z])|([c-zC-Z]:\\))(.+?).cs:line \d+)`)
	maybeGOStackTraceRegex     = regexp.MustCompile(`((/[a-zA-Z])|([c-zC-Z]:\\))(.+?).go:\d+ \+0x`)

	maybeChinaIDCardNumberRegex = regexp.MustCompile(`\b[1-9]\d{5}(19|20)\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]\b`)
	maybeBase64Regex            = regexp.MustCompile(`^[a-zA-Z0-9+/=]{4,}$`)
	maybeMD5Regex               = regexp.MustCompile(`^[0-9a-z]{32}$`)
	maybeSHA256Regex            = regexp.MustCompile(`^[0-9a-z]{64}$`)
)

func init() {
	keyList := make([]string, 0, 4*len(sensitiveJSONKeyList))
	for _, v := range sensitiveJSONKeyList {
		keyList = append(keyList, fmt.Sprintf(`(%s:)`, v),
			fmt.Sprintf(`('%s':)`, v),
			fmt.Sprintf(`(\\'%s\\':)`, v),
			fmt.Sprintf(`("%s":)`, v),
			fmt.Sprintf(`(\\"%s\\":)`, v))
	}
	s := `(?i)` + strings.Join(keyList, "|")
	sensitiveJSONKeyRegex = regexp.MustCompile(s)
}

func GetSensitiveKeyList() []string {
	// https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice
	return append(sensitiveJSONKeyList[:0:0], sensitiveJSONKeyList...)
}

// . Guess whether it is based on the name of the key. Detection key based on redirect parameter
func BeUsedForRedirect(key string, value interface{}) bool {
	return maybeRedirectRegex.MatchString(key) || IsURLPath(value) || IsFullURL(value)
}

func IsJSONPParam(key string, value interface{}) bool {
	return maybeJSONPRegex.MatchString(key) || maybeJSONPValue.MatchString(fmt.Sprint(value))
}

func IsGenericURLParam(key string, value interface{}) bool {
	return maybeURLRegex.MatchString(key) && (IsURLPath(value) || IsFullURL(value))
}

func IsSensitiveJSON(data []byte) bool {
	// 检测 key: 'key': "key": \'key\': \"key\": Several forms of keys. If it is the following keyword, it is considered to be sensitive information. The path starting with
	return sensitiveJSONKeyRegex.Match(data)
}

func IsTokenParam(key string) bool {
	return maybeTokenRegex.MatchString(key)
}

func IsSQLColumnName(s string) bool {
	return maybeSQLColumnNameKey.MatchString(s)
}

func IsPasswordKey(key string) bool {
	return maybePasswordKeyRegex.MatchString(key)
}

func IsUsernameKey(key string) bool {
	if maybeUsernameKeyRegex.MatchString(key) {
		return true
	}
	key = strings.ToLower(key)
	for _, f1 := range []string{"name"} {
		for _, f2 := range []string{"user", "login", "nick", "account", "auth"} {
			if strings.Contains(key, f1) && strings.Contains(key, f2) {
				return true
			}
		}
	}
	return false
}

func IsCaptchaKey(key string) bool {
	if maybeCaptchaKeyRegex.MatchString(key) {
		return true
	}
	key = strings.ToLower(key)
	for _, f1 := range []string{"code"} {
		for _, f2 := range []string{"check", "chk", "validate", "verify", "verification", "validation",
			"img", "image", "pic", "auth", "rand", "confirm"} {
			if strings.Contains(key, f1) && strings.Contains(key, f2) {
				return true
			}
		}
	}
	return false
}

func IsBase64(s string) bool {
	return utils.IsBase64(s)
}

func IsBase64Password(s string) bool {
	if !IsBase64(s) {
		return false
	}
	password, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	for _, item := range string(password) {
		if unicode.IsControl(item) {
			return false
		}
	}
	return true
}

func IsMD5Data(s string) bool {
	return maybeMD5Regex.MatchString(s)
}

func IsSHA256Data(s string) bool {
	return maybeSHA256Regex.MatchString(s)
}

func IsXMLRequest(raw []byte) bool {
	raw = lowhttp.FixHTTPRequest(raw)
	_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(raw)
	req, err := lowhttp.ParseBytesToHttpRequest(raw)
	if err != nil {
		return false
	}

	if req.Method != "POST" {
		return false
	}

	contentType := req.Header.Get("Content-Type")

	if strings.Contains(contentType, "xml") {
		return true
	}

	if strings.Contains(contentType, "octet-stream") {
		reader := io.LimitReader(bytes.NewBuffer(body), 128)
		b, err := ioutil.ReadAll(reader)
		if err == nil && IsXMLBytes(b) {
			return true
		}
	}
	return false
}

var (
	maybeXMLKey = regexp.MustCompile(`(?i)(xml)`)
	maybeXML    = regexp.MustCompile(`(?i)(<\?xml.*>)|(<.*>.*</.*>)`)
)

func IsXMLString(data string) bool {
	return maybeXML.MatchString(data)
}

func IsXMLBytes(data []byte) bool {
	return maybeXML.Match(data)
}

func IsXMLParam(key string, value interface{}) bool {
	if maybeXMLKey.MatchString(key) || IsXMLString(fmt.Sprint(value)) {
		return true
	}
	return false
}
