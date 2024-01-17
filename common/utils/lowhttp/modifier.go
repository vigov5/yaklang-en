package lowhttp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/samber/lo"

	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/jsonpath"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func IsHeader(headerLine, wantHeader string) bool {
	return strings.HasPrefix(strings.ToLower(headerLine), strings.ToLower(wantHeader)+":")
}

func IsChunkedHeaderLine(line string) bool {
	k, v := SplitHTTPHeader(line)
	if utils.IContains(k, "transfer-encoding") && utils.IContains(v, "chunked") {
		return true
	}
	return false
}

func SetHTTPPacketUrl(packet []byte, rawURL string) []byte {
	var buf bytes.Buffer
	var header []string
	isChunked := false
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return packet
	}

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			buf.WriteString(method + " " + parsed.RequestURI() + " " + proto)
			buf.WriteString(CRLF)
			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			if IsHeader(line, "Host") {
				line = fmt.Sprintf("Host: %s", parsed.Host)
			}
			header = append(header, line)
			return line
		},
	)

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// ReplaceHTTPPacketFirstLine Yes An auxiliary, used to change the request message, modify the first line (i.e. request method, request path, protocol version)
// Example:
// ```
// poc.ReplaceHTTPPacketFirstLine(`GET / HTTP/1.1
// Host: Example.com
// `, "GET /test HTTP/1.1")) // . Initiate a request to example.com, modify the first line of the request message, and request/test path
// ```
func ReplaceHTTPPacketFirstLine(packet []byte, firstLine string) []byte {
	var isChunked bool
	header := []string{firstLine}
	_, body := SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
		header = append(header, line)
		return line
	})
	return ReplaceHTTPPacketBody([]byte(strings.Join(header, CRLF)+CRLF), body, isChunked)
}

// ReplaceHTTPPacketMethod is an auxiliary function, used to change the request message, modify the request method
// Example:
// ```
// poc.ReplaceHTTPPacketMethod(poc.BasicRequest(), "OPTIONS") // Modify the request method to OPTIONS
// ```
func ReplaceHTTPPacketMethod(packet []byte, newMethod string) []byte {
	var buf bytes.Buffer
	var header []string
	var (
		isChunked      = false
		isPost         = strings.ToUpper(newMethod) == "POST"
		hasContentType = false
	)

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			method = newMethod
			buf.WriteString(method + " " + requestUri + " " + proto)
			buf.WriteString(CRLF)

			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			header = append(header, line)
			if IsHeader(line, "Content-Type") {
				hasContentType = true
			}

			return line
		},
	)

	// fix content-type
	if isPost && !hasContentType {
		header = append(header, "Content-Type: application/x-www-form-urlencoded")
	}

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// ReplaceHTTPPacketPath is an auxiliary function used to change the request message and modify the request path
// Example:
// ```
// poc.ReplaceHTTPPacketPath(poc.BasicRequest(), "/get") // Modify the request path to/get
// ```
func ReplaceHTTPPacketPath(packet []byte, p string) []byte {
	var isChunked bool
	var buf bytes.Buffer
	var header []string

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			defer func() {
				buf.WriteString(method + " " + requestUri + " " + proto)
				buf.WriteString(CRLF)
			}()

			// handle requestUri
			u, _ := url.Parse(requestUri)
			if u == nil { // invalid url
				return nil
			}

			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			u.Path = p
			requestUri = u.String()

			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			header = append(header, line)
			return line
		},
	)

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// AppendHTTPPacketPath is an auxiliary function used to change the request message. Add the request path
// Example:
// ```
// poc.AppendHTTPPacketPath(`GET /docs HTTP/1.1
// Host: yaklang.com
// `, "/api/poc")) // Initiates a request to example.com, the actual request Change the path to/docs/api/poc
// ```
func AppendHTTPPacketPath(packet []byte, p string) []byte {
	var isChunked bool
	var buf bytes.Buffer
	var header []string

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			defer func() {
				buf.WriteString(method + " " + requestUri + " " + proto)
				buf.WriteString(CRLF)
			}()

			// handle requestUri
			u, _ := url.Parse(requestUri)
			if u == nil { // invalid url
				return nil
			}

			u.Path = path.Join(u.Path, p)
			requestUri = u.String()

			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			header = append(header, line)
			return line
		},
	)

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

func handleHTTPPacketQueryParam(packet []byte, noAutoEncode bool, callback func(*QueryParams)) []byte {
	var isChunked bool
	var buf bytes.Buffer
	var header []string

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			defer func() {
				buf.WriteString(method + " " + requestUri + " " + proto)
				buf.WriteString(CRLF)
			}()

			urlIns := ForceStringToUrl(requestUri)
			u := NewQueryParams(urlIns.RawQuery).DisableAutoEncode(noAutoEncode)
			callback(u)
			urlIns.RawQuery = u.Encode()
			requestUri = urlIns.String()
			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			header = append(header, line)
			return line
		},
	)

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// ReplaceAllHTTPPacketQueryParams is an auxiliary function, used To change the request message, modify all GET request parameters. If they do not exist, they will be added. It receives a map[string]string type parameter, where key is the request parameter name and value is the request parameter value.
// Example:
// ```
// poc.ReplaceAllHTTPPacketQueryParams(poc.BasicRequest(), {"a":"b", "c":"d"}) // Add GET request parameter a, the value is b, add GET request parameter c, the value is d
// ```
func ReplaceAllHTTPPacketQueryParams(packet []byte, values map[string]string) []byte {
	return handleHTTPPacketQueryParam(packet, false, func(q *QueryParams) {
		// clear all values
		shouldRemove := make(map[string]struct{})
		shouldReplace := make(map[string]string)
		for _, item := range q.Items {
			_, ok := values[item.Key]
			if !ok {
				shouldRemove[item.Key] = struct{}{}
			} else {
				shouldReplace[item.Key] = values[item.Key]
			}
		}

		for k := range shouldRemove {
			q.Remove(k)
		}
		var extraItem []*QueryParamItem
		for k, v := range values {
			_, ok := shouldReplace[k]
			if ok {
				q.Set(k, v)
			} else {
				extraItem = append(extraItem, &QueryParamItem{Key: k, Value: v})
			}
		}

		if len(extraItem) > 0 {
			sort.SliceStable(extraItem, func(i, j int) bool {
				return extraItem[i].Key < extraItem[j].Key
			})
			lo.ForEach(extraItem, func(item *QueryParamItem, _ int) {
				q.Set(item.Key, item.Value)
			})
		}
	})
}

// ReplaceHTTPPacketQueryParam is an auxiliary function used to change the request message and modify the GET request parameters. If it does not exist,
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("GET", "https://pie.dev/get")
// poc.ReplaceHTTPPacketQueryParam(raw, "a", "b") // Add GET request parameter a, the value is b
// ```
func ReplaceHTTPPacketQueryParam(packet []byte, key, value string) []byte {
	return handleHTTPPacketQueryParam(packet, false, func(q *QueryParams) {
		q.Set(key, value)
	})
}

func ReplaceHTTPPacketQueryParamWithoutEncoding(packet []byte, key, value string) []byte {
	return handleHTTPPacketQueryParam(packet, true, func(q *QueryParams) {
		q.Set(key, value)
	})
}

func ReplaceHTTPPacketQueryParamRaw(packet []byte, rawQuery string) []byte {
	var isChunked bool
	var buf bytes.Buffer
	var header []string

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			defer func() {
				buf.WriteString(method + " " + requestUri + " " + proto)
				buf.WriteString(CRLF)
			}()

			urlIns := ForceStringToUrl(requestUri)
			urlIns.RawQuery = rawQuery
			requestUri = urlIns.String()
			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			header = append(header, line)
			return line
		},
	)

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// AppendHTTPPacketQueryParam is an auxiliary function used to change the request message and add GET request parameters
// Example:
// ```
// poc.AppendHTTPPacketQueryParam(poc.BasicRequest(), "a", "b") // Add GET request parameter a, the value is b
// ```
func AppendHTTPPacketQueryParam(packet []byte, key, value string) []byte {
	return handleHTTPPacketQueryParam(packet, false, func(q *QueryParams) {
		q.Add(key, value)
	})
}

func AppendAllHTTPPacketQueryParam(packet []byte, Params map[string][]string) []byte {
	for key, values := range Params {
		for _, value := range values {
			if value == "" {
				continue
			}
			packet = handleHTTPPacketQueryParam(packet, false, func(q *QueryParams) {
				q.Add(key, value)
			})
		}
	}
	return packet
}

// DeleteHTTPPacketQueryParam is an auxiliary function, used to change the request message, delete the GET request parameter
// Example:
// ```
// poc.DeleteHTTPPacketQueryParam(`GET /get?a=b&c=d HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// `, "a") // Delete the GET request parameter a
// ```
func DeleteHTTPPacketQueryParam(packet []byte, key string) []byte {
	return handleHTTPPacketQueryParam(packet, false, func(q *QueryParams) {
		q.Del(key)
	})
}

func handleHTTPPacketPostParam(packet []byte, callback func(url.Values)) []byte {
	var isChunked bool

	headersRaw, bodyRaw := SplitHTTPPacket(packet, nil, nil)
	bodyString := unsafe.String(unsafe.SliceData(bodyRaw), len(bodyRaw))
	values, err := url.ParseQuery(bodyString)
	if err == nil {
		callback(values)
		// values.Set(key, value)
		newBody := values.Encode()
		bodyRaw = unsafe.Slice(unsafe.StringData(newBody), len(newBody))
	}

	return ReplaceHTTPPacketBody([]byte(headersRaw), bodyRaw, isChunked)
}

// ReplaceAllHTTPPacketPostParams is an auxiliary function. Function, used to change the request message and modify all POST request parameters. If they do not exist, they will be added. It receives a map[string]string type parameter, where key is the POST request parameter name and value is the POST request parameter value.
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("POST", "https://pie.dev/post")
// poc.ReplaceAllHTTPPacketPostParams(raw, {"a":"b", "c":"d"}) // will be added. Add POST request parameter a, value is b, POST request parameter c, The value is d.
// ```
func ReplaceAllHTTPPacketPostParams(packet []byte, values map[string]string) []byte {
	return handleHTTPPacketPostParam(packet, func(q url.Values) {
		// clear all values
		for k := range q {
			q.Del(k)
		}

		for k, v := range values {
			q.Set(k, v)
		}
	})
}

// ReplaceHTTPPacketPostParam is an auxiliary function used to change the request message and modify the POST request parameters. If it does not exist,
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("POST", "https://pie.dev/post")
// poc.ReplaceHTTPPacketPostParam(raw, "a", "b") // Add POST request parameter a with value b
// ```
func ReplaceHTTPPacketPostParam(packet []byte, key, value string) []byte {
	return handleHTTPPacketPostParam(packet, func(q url.Values) {
		q.Set(key, value)
	})
}

// AppendHTTPPacketPostParam is an auxiliary function, used to change the request message, add POST request parameters
// Example:
// ```
// poc.AppendHTTPPacketPostParam(poc.BasicRequest(), "a", "b") // initiates a request to pie.dev and deletes the POST request. Parameter a
// ```
func AppendHTTPPacketPostParam(packet []byte, key, value string) []byte {
	return handleHTTPPacketPostParam(packet, func(q url.Values) {
		q.Add(key, value)
	})
}

// DeleteHTTPPacketPostParam is an auxiliary function used to change the request message, delete the POST request parameter
// Example:
// ```
// poc.DeleteHTTPPacketPostParam(`POST /post HTTP/1.1
// Content-Type: application/json
// Content-Length: 7
// Host: pie.dev
//
// a=b&c=d`, "a") // deletes the POST request parameter a.
// ```
func DeleteHTTPPacketPostParam(packet []byte, key string) []byte {
	return handleHTTPPacketPostParam(packet, func(q url.Values) {
		q.Del(key)
	})
}

// ReplaceHTTPPacketHeader is an auxiliary function used to change the request message and modify the request header. If it does not exist, it will be added.
// Example:
// ```
// poc.ReplaceHTTPPacketHeader(poc.BasicRequest(),"AAA", "BBB") // Modifies the value of the AAA request header For BBB, there is no AAA request header, so the request header will be added.
// ```
func ReplaceHTTPPacketHeader(packet []byte, headerKey string, headerValue any) []byte {
	var firstLine string
	var header []string
	var handled bool
	isChunked := IsChunkedHeaderLine(headerKey + ": " + utils.InterfaceToString(headerValue))
	_, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		firstLine = method + " " + requestUri + " " + proto
		return nil
	}, func(proto string, code int, codeMsg string) error {
		if codeMsg == "" {
			firstLine = proto + " " + fmt.Sprint(code)
		} else {
			firstLine = proto + " " + fmt.Sprint(code) + " " + codeMsg
		}
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
		if k, _ := SplitHTTPHeader(line); k == headerKey {
			handled = true
			header = append(header, headerKey+": "+utils.InterfaceToString(headerValue))
			return line
		}
		header = append(header, line)
		return line
	})
	if !handled {
		header = append(header, headerKey+": "+utils.InterfaceToString(headerValue))
	}
	var buf bytes.Buffer
	buf.WriteString(firstLine)
	buf.WriteString(CRLF)
	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// ReplaceHTTPPacketHost is an auxiliary function, used to change the request message, modify the Host The request header will be added if it does not exist. It is actually ReplaceHTTPPacketHeader ("Host", the abbreviation of host),
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("GET", "https://yaklang.com")
// poc.ReplaceHTTPPacketHost(raw, "www.yaklang.com") // Modifies the value of the Host request header to www.yaklang.com
// ```
func ReplaceHTTPPacketHost(packet []byte, host string) []byte {
	return ReplaceHTTPPacketHeader(packet, "Host", host)
}

// ReplaceHTTPPacketBasicAuth is an auxiliary function used to change the request message and modify the Authorization request. The header is the ciphertext of the basic authentication. If it does not exist, it will be added. It is actually ReplaceHTTPPacketHeader("Authorization", codec.EncodeBase64(username + ":" , the abbreviation of
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("GET", "https://pie.dev/basic-auth/admin/password")
// poc.ReplaceHTTPPacketBasicAuth(raw, "admin", "password") // Modify the Authorization request header
// ```
func ReplaceHTTPPacketBasicAuth(packet []byte, username, password string) []byte {
	return ReplaceHTTPPacketHeader(packet, "Authorization", "Basic "+codec.EncodeBase64(username+":"+password))
}

// AppendHTTPPacketHeader is an auxiliary function used to change the request message and add the request header.
// Example:
// ```
// poc.AppendHTTPPacketHeader(poc.BasicRequest(), "AAA", "BBB") // Add AAA The value of the request header is BBB.
// ```
func AppendHTTPPacketHeader(packet []byte, headerKey string, headerValue any) []byte {
	var firstLine string
	var header []string
	var isChunked bool
	_, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		firstLine = method + " " + requestUri + " " + proto
		return nil
	}, func(proto string, code int, codeMsg string) error {
		if codeMsg == "" {
			firstLine = proto + " " + fmt.Sprint(code)
		} else {
			firstLine = proto + " " + fmt.Sprint(code) + " " + codeMsg
		}
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
		header = append(header, line)
		return line
	})
	header = append(header, headerKey+": "+utils.InterfaceToString(headerValue))
	var buf bytes.Buffer
	buf.WriteString(firstLine)
	buf.WriteString(CRLF)
	buf.WriteString(strings.Join(header, CRLF))
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// DeleteHTTPPacketHeader is an auxiliary function, used to change the request message and delete the request header.
// Example:
// ```
// poc.DeleteHTTPPacketHeader(`GET /get HTTP/1.1
// Content-Type: application/json
// AAA: BBB
// Host: pie.dev
//
// `, "AAA") // Delete the AAA request header
// ```
func DeleteHTTPPacketHeader(packet []byte, headerKey string) []byte {
	var firstLine string
	var header []string
	var isChunked bool
	_, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		firstLine = method + " " + requestUri + " " + proto
		return nil
	}, func(proto string, code int, codeMsg string) error {
		if codeMsg == "" {
			firstLine = proto + " " + fmt.Sprint(code)
		} else {
			firstLine = proto + " " + fmt.Sprint(code) + " " + codeMsg
		}
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}

		if k, _ := SplitHTTPHeader(line); k == headerKey {
			return ""
		}
		header = append(header, line)
		return line
	})
	var buf bytes.Buffer
	buf.WriteString(firstLine)
	buf.WriteString(CRLF)
	buf.WriteString(strings.Join(header, CRLF))
	return ReplaceHTTPPacketBody(buf.Bytes(), body, false)
}

// ReplaceHTTPPacketCookie is an auxiliary function used to change the request message and modify the value in the Cookie request header. If it does not exist,
// Example:
// ```
// poc.ReplaceHTTPPacketCookie(poc.BasicRequest(), p"aaa", "bbb") // Modify the cookie value. Since there is no aaa cookie value here,
// ```
func ReplaceHTTPPacketCookie(packet []byte, key string, value any) []byte {
	var isReq bool
	var isRsp bool
	handled := false
	var isChunked bool
	header, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		isReq = true
		return nil
	}, func(proto string, code int, codeMsg string) error {
		isRsp = true
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}

		if !isReq && !isRsp {
			return line
		}

		k, cookieRaw := SplitHTTPHeader(line)
		if (strings.ToLower(k) == "cookie" && isReq) || (strings.ToLower(k) == "set-cookie" && isRsp) {
			existed := ParseCookie(cookieRaw)
			if len(existed) <= 0 {
				return line
			}
			cookie := make([]*http.Cookie, len(existed))
			for index, c := range existed {
				if c.Name == key {
					handled = true
					c.Value = utils.InterfaceToString(value)
				}
				cookie[index] = c
			}
			return k + ": " + CookiesToString(cookie)
		}
		return line
	})
	data := ReplaceHTTPPacketBody([]byte(header), body, isChunked)
	if handled {
		return data
	}
	return AppendHTTPPacketCookie(data, key, value)
}

// AppendHTTPPacketCookie is an auxiliary function used to change the request message and add the value in the Cookie request header.
// Example:
// ```
// poc.AppendHTTPPacketCookie(poc.BasicRequest(), "aaa", "bbb") // Add cookie key-value pair aaa:bbb
// ```
func AppendHTTPPacketCookie(packet []byte, key string, value any) []byte {
	var isReq bool
	var added bool
	var isRsp bool
	var isChunked bool
	header, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		isReq = true
		return nil
	}, func(proto string, code int, codeMsg string) error {
		isRsp = true
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}

		if !isReq && !isRsp {
			return line
		}

		if added {
			return line
		}

		k, cookieRaw := SplitHTTPHeader(line)
		k = strings.ToLower(k)
		if (k == "cookie" && isReq) || (k == "set-cookie" && isRsp) {
			existed := ParseCookie(cookieRaw)
			existed = append(existed, &http.Cookie{Name: key, Value: utils.InterfaceToString(value)})
			added = true
			return k + ": " + CookiesToString(existed)
		}

		return line
	})
	if !added {
		if isReq {
			header = strings.Trim(header, CRLF) + CRLF + "Cookie: " + CookiesToString([]*http.Cookie{
				{Name: key, Value: utils.InterfaceToString(value)},
			})
		}
		if isRsp {
			header = strings.Trim(header, CRLF) + CRLF + "Set-Cookie: " + CookiesToString([]*http.Cookie{
				{Name: key, Value: utils.InterfaceToString(value)},
			})
		}
	}
	return ReplaceHTTPPacketBody([]byte(header), body, isChunked)
}

// will be added. DeleteHTTPPacketCookie is an auxiliary function used to change the request message and delete the value in the cookie.
// Example:
// ```
// poc.DeleteHTTPPacketCookie(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: aaa=bbb; ccc=ddd
// Host: pie.dev
//
// `, "aaa") // Delete aaa in the cookie
// ```
func DeleteHTTPPacketCookie(packet []byte, key string) []byte {
	var isReq bool
	var isRsp bool
	var isChunked bool
	header, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		isReq = true
		return nil
	}, func(proto string, code int, codeMsg string) error {
		isRsp = true
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
		if !isReq && !isRsp {
			return line
		}

		k, cookieRaw := SplitHTTPHeader(line)
		k = strings.ToLower(k)

		if (k == "cookie" && isReq) || (k == "set-cookie" && isRsp) {
			existed := ParseCookie(cookieRaw)
			existed = funk.Filter(existed, func(cookie *http.Cookie) bool {
				return cookie.Name != key
			}).([]*http.Cookie)
			return k + ": " + CookiesToString(existed)
		}

		return line
	})
	return ReplaceHTTPPacketBody([]byte(header), body, isChunked)
}

func handleHTTPRequestForm(packet []byte, fixMethod bool, fixContentType bool, callback func(string, *multipart.Reader, *multipart.Writer) bool) []byte {
	var header []string
	var (
		buf             bytes.Buffer
		bodyBuf         bytes.Buffer
		requestMethod                     = ""
		hasContentType                    = false
		isChunked                         = false
		isFormDataPost                    = false
		fixBody                           = false
		multipartWriter *multipart.Writer = multipart.NewWriter(&bodyBuf)
	)
	// not handle response
	if bytes.HasPrefix(packet, []byte("HTTP/")) {
		return packet
	}

	_, body := SplitHTTPPacket(packet,
		func(method string, requestUri string, proto string) error {
			requestMethod = method
			// rewrite method
			if fixMethod {
				method = "POST"
			}

			buf.WriteString(method + " " + requestUri + " " + proto)
			buf.WriteString(CRLF)

			return nil
		},
		nil,
		func(line string) string {
			if !isChunked {
				isChunked = IsChunkedHeaderLine(line)
			}
			if IsHeader(line, "Content-Type") {
				hasContentType = true
				_, v := SplitHTTPHeader(line)
				d, params, _ := mime.ParseMediaType(v)

				if d == "multipart/form-data" {
					isFormDataPost = true
					// try to get boundary
					if boundary, ok := params["boundary"]; ok {
						multipartWriter.SetBoundary(boundary)
					}
				} else if fixContentType {
					// rewrite content-type
					line = fmt.Sprintf("Content-Type: %s", multipartWriter.FormDataContentType())
				}
			}
			header = append(header, line)
			return line
		},
	)

	if isFormDataPost {
		// multipart reader
		multipartReader := multipart.NewReader(bytes.NewReader(body), multipartWriter.Boundary())
		// append form
		fixBody = callback(requestMethod, multipartReader, multipartWriter)
	} else {
		// rewrite body
		fixBody = callback(requestMethod, nil, multipartWriter)
	}
	multipartWriter.Close()
	if fixBody {
		body = bodyBuf.Bytes()
	}

	if fixContentType && !hasContentType {
		header = append(header, fmt.Sprintf("Content-Type: %s", multipartWriter.FormDataContentType()))
	}

	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// AppendHTTPPacketFormEncoded is an auxiliary function used to change the request message and add the form in the request body.
// Example:
// ```
// poc.AppendHTTPPacketFormEncoded(`POST /post HTTP/1.1
// Host: pie.dev
// Content-Type: multipart/form-data; boundary=------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm
// Content-Length: 203
//
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm
// Content-Disposition: form-data; name="aaa"
//
// bbb
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm--`, "ccc", "ddd") // Add POST request form, where ccc is the key and ddd is the value
// ```
func AppendHTTPPacketFormEncoded(packet []byte, key, value string) []byte {
	return handleHTTPRequestForm(packet, true, true, func(_ string, multipartReader *multipart.Reader, multipartWriter *multipart.Writer) bool {
		if multipartReader != nil {
			// copy part
			for {
				part, err := multipartReader.NextPart()
				if err != nil {
					break
				}
				partWriter, err := multipartWriter.CreatePart(part.Header)
				if err != nil {
					break
				}
				_, err = io.Copy(partWriter, part)
				if err != nil {
					break
				}
			}
		}
		// append form
		if multipartWriter != nil {
			multipartWriter.WriteField(key, value)
		}
		return true
	})
}

// AppendHTTPPacketUploadFile is an auxiliary function used to change the request message and add the request body The uploaded file, the first parameter is the original request message, the second parameter is the form name, the third parameter is the file name, the fourth parameter is the file content, the fifth parameter is an optional parameter, which is the file Type (Content-Type)
// Example:
// ```
// _, raw, _ = poc.ParseUrlToHTTPRequestRaw("POST", "https://pie.dev/post")
// poc.AppendHTTPPacketUploadFile(raw, "file", "phpinfo.php", "<?php phpinfo(); ?>", "image/jpeg")) // adds a POST request form with the file name phpinfo.php and the content is<?php phpinfo(); ?>, and the file type is image./jpeg
// ```
func AppendHTTPPacketUploadFile(packet []byte, fieldName, fileName string, fileContent interface{}, contentType ...string) []byte {
	hasContentType := len(contentType) > 0

	return handleHTTPRequestForm(packet, true, true, func(_ string, multipartReader *multipart.Reader, multipartWriter *multipart.Writer) bool {
		if multipartReader != nil {
			// copy part
			for {
				part, err := multipartReader.NextPart()
				if err != nil {
					break
				}
				partWriter, err := multipartWriter.CreatePart(part.Header)
				if err != nil {
					break
				}
				_, err = io.Copy(partWriter, part)
				if err != nil {
					break
				}
			}
		}
		// append upload file
		if multipartWriter != nil {
			var content []byte
			h := make(textproto.MIMEHeader)
			contentDisposition := fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldName))
			if fileName != "" {
				contentDisposition += fmt.Sprintf(`;filename="%s"`, escapeQuotes(fileName))
			}
			h.Set("Content-Disposition", contentDisposition)

			guessContentType := "application/octet-stream"
			if hasContentType {
				guessContentType = contentType[0]
			}

			switch r := fileContent.(type) {
			case string:
				content = unsafe.Slice(unsafe.StringData(r), len(r))
			case []byte:
				content = r
			case io.Reader:
				r.Read(content)
			}
			if !hasContentType {
				guessContentType = http.DetectContentType(content)
			}

			if guessContentType != "" {
				h.Set("Content-Type", guessContentType)
			}

			partWriter, err := multipartWriter.CreatePart(h)
			if err == nil {
				partWriter.Write(content)
			}
		}
		return true
	})
}

// DeleteHTTPPacketForm is an auxiliary function used to change the request message, delete the POST request form
// Example:
// ```
// poc.DeleteHTTPPacketForm(`POST /post HTTP/1.1
// Host: pie.dev
// Content-Type: multipart/form-data; boundary=------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm
// Content-Length: 308
//
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm
// Content-Disposition: form-data; name="aaa"
//
// bbb
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm
// Content-Disposition: form-data; name="ccc"
//
// ddd
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm--`, "aaa") // Delete the POST request form aaa
// ```
func DeleteHTTPPacketForm(packet []byte, key string) []byte {
	return handleHTTPRequestForm(packet, false, false, func(method string, multipartReader *multipart.Reader, multipartWriter *multipart.Writer) bool {
		if strings.ToUpper(method) != "POST" {
			return false
		}

		if multipartReader != nil {
			// copy part
			for {
				part, err := multipartReader.NextPart()
				if err != nil {
					break
				}

				// skip part if key matched
				if part.FormName() == key {
					continue
				}

				partWriter, err := multipartWriter.CreatePart(part.Header)
				if err != nil {
					break
				}
				_, err = io.Copy(partWriter, part)
				if err != nil {
					break
				}
			}
			return true
		}
		return false
	})
}

func GetParamsFromBody(contentType string, body []byte) (params map[string]string, useRaw bool, err error) {
	var (
		mediaType         string
		contentTYpeParams map[string]string
	)
	params = make(map[string]string)
	// This is to handle complex json/xml situation
	handleUnmarshalValues := func(v any) ([]string, []string) {
		var keys, values []string
		ref := reflect.ValueOf(v)
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			arrayLen := ref.Len()
			if arrayLen > 0 {
				return []string{""}, []string{utils.InterfaceToString(ref.Index(arrayLen - 1).Interface())}
			}
		case reflect.Map:
			refKeys := ref.MapKeys()
			if len(refKeys) > 0 {
				for _, refKeys := range refKeys {
					keys = append(keys, utils.InterfaceToString(refKeys.Interface()))
					values = append(values, utils.InterfaceToString(ref.MapIndex(refKeys).Interface()))
				}
				return keys, values
			}
		case reflect.Float32, reflect.Float64:
			floatV, ok := v.(float64)
			if ok && floatV == float64(int(floatV)) {
				v = int(floatV)
			}
		}
		return []string{""}, []string{utils.InterfaceToString(v)}
	}
	handleUnmarshalResults := func(tempMap map[string]any) map[string]string {
		params := make(map[string]string, len(tempMap))
		for k, v := range tempMap {
			extraKeys, extraValues := handleUnmarshalValues(v)
			for i, key := range extraKeys {
				if key == "" {
					params[k] = extraValues[i]
					continue
				}
				params[fmt.Sprintf("%s[%s]", k, key)] = extraValues[i]
			}
		}
		return params
	}

	mediaType, contentTYpeParams, err = mime.ParseMediaType(contentType)
	if err != nil {
		return
	}
	if mediaType == "application/x-www-form-urlencoded" {
		var values url.Values
		values, err = url.ParseQuery(utils.UnsafeBytesToString(body))
		if err == nil {
			for k, v := range values {
				params[k] = v[0]
			}
		}
	} else if mediaType == "multipart/form-data" {
		boundary, ok := contentTYpeParams["boundary"]
		if !ok {
			err = utils.Error("boundary not found")
			return
		}
		mr := multipart.NewReader(bytes.NewReader(body), boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, false, err
			}

			// Check whether part is a form field
			if part.FormName() != "" {
				content, err := io.ReadAll(part)
				if err != nil && err != io.EOF {
					return nil, false, err
				}
				params[part.FormName()] = utils.UnsafeBytesToString(content)
			}
		}
	} else if strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		var tempMap map[string]any
		err = json.Unmarshal(body, &tempMap)
		if err == nil {
			params = handleUnmarshalResults(tempMap)
		}
	} else if strings.HasPrefix(strings.ToLower(contentType), "application/xml") {
		tempMap := utils.XmlLoads(body)
		if err == nil {
			params = handleUnmarshalResults(tempMap)
		}
	} else {
		// . This flag bit is used to mark whether to call Or use the original body directly, which is used by default
		useRaw = true
	}

	return
}

func AppendHTTPPacketHeaderIfNotExist(packet []byte, headerKey string, headerValue any) []byte {
	var firstLine string
	var header []string
	var exist bool
	isChunked := IsChunkedHeaderLine(headerKey + ": " + utils.InterfaceToString(headerValue))
	_, body := SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		firstLine = method + " " + requestUri + " " + proto
		return nil
	}, func(proto string, code int, codeMsg string) error {
		if codeMsg == "" {
			firstLine = proto + " " + fmt.Sprint(code)
		} else {
			firstLine = proto + " " + fmt.Sprint(code) + " " + codeMsg
		}
		return nil
	}, func(line string) string {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
		if k, _ := SplitHTTPHeader(line); k == headerKey {
			exist = true
		}
		header = append(header, line)
		return line
	})
	if !exist {
		header = append(header, headerKey+": "+utils.InterfaceToString(headerValue))
	}
	var buf bytes.Buffer
	buf.WriteString(firstLine)
	buf.WriteString(CRLF)
	for _, line := range header {
		buf.WriteString(line)
		buf.WriteString(CRLF)
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, isChunked)
}

// GetHTTPPacketCookieValues will be returned here. It is an auxiliary function used to obtain the Cookie value in the request message. Its return value is []string. This is because Cookie may have multiple values with the same key name.
// Example:
// ```
// poc.GetHTTPPacketCookieValues(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c
// Host: pie.dev
//
// `, "a") // after the existing request path to get the Cookie value with the key name a. Here it will return ["b", "c"]
// ```
func GetHTTPPacketCookieValues(packet []byte, key string) (cookieValues []string) {
	var val []string
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if k, cookieRaw := SplitHTTPHeader(line); k == "Cookie" || k == "cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				if e.Name == key {
					val = append(val, e.Value)
				}
			}
		}

		if k, cookieRaw := SplitHTTPHeader(line); strings.ToLower(k) == "set-cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				if e.Name == key {
					val = append(val, e.Value)
				}
			}
		}

		return line
	})
	return val
}

// GetHTTPPacketCookieFirst will be added. Is an auxiliary function used to get the Cookie value in the request message, and its return value is string
// Example:
// ```
// poc.GetHTTPPacketCookieFirst(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; c=d
// Host: pie.dev
//
// `, "a") // gets the Cookie value with key name a, which will be returned here."b"
// ```
func GetHTTPPacketCookieFirst(packet []byte, key string) (cookieValue string) {
	ret := GetHTTPPacketCookieValues(packet, key)
	if len(ret) > 0 {
		return ret[0]
	}
	return ""
}

// GetUrlFromHTTPRequest is an auxiliary function used to obtain the URL in the request message, and its return value is string
// Example:
// ```
// poc.GetUrlFromHTTPRequest("https", `GET /get HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// `) // gets the URL."https://pie.dev/get"
func GetUrlFromHTTPRequest(scheme string, packet []byte) (url string) {
	if scheme == "" {
		scheme = "http"
	}
	u, err := ExtractURLFromHTTPRequestRaw(packet, strings.HasPrefix(strings.ToLower(scheme), "https"))
	if err != nil {
		return ""
	}
	u.Scheme = scheme
	return u.String()
}

// GetHTTPPacketCookie is an auxiliary function, used to obtain the Cookie value in the request message. Its return value is string.
// Example:
// ```
// poc.GetHTTPPacketCookie(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; c=d
// Host: pie.dev
//
// `, "a") // gets the Cookie value with key name a, which will be returned here."b"
// ```
func GetHTTPPacketCookie(packet []byte, key string) (cookieValue string) {
	return GetHTTPPacketCookieFirst(packet, key)
}

// GetHTTPPacketContentType is an auxiliary function used to obtain the Content-Type request header in the request message, and its return value is string
// Example:
// ```
// poc.GetHTTPPacketContentType(`POST /post HTTP/1.1
// Content-Type: application/json
// COntent-Length: 7
// Host: pie.dev
//
// a=b&c=d`) // Gets the Content-Type request header
// ```
func GetHTTPPacketContentType(packet []byte) (contentType string) {
	var val string
	fetched := false
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if fetched {
			return line
		}
		if k, v := SplitHTTPHeader(line); strings.ToLower(k) == "content-type" {
			fetched = true
			val = v
		}
		return line
	})
	return val
}

// GetHTTPPacketCookies Is an auxiliary function used to obtain all Cookie values in the request message, and its return value is map[string]string
// Example:
// ```
// poc.GetHTTPPacketCookies(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; c=d
// Host: pie.dev
//
// `) // gets all cookie values. Here Will return{"a":"b", "c":"d"}
// ```
func GetHTTPPacketCookies(packet []byte) (cookies map[string]string) {
	val := make(map[string]string)
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if k, cookieRaw := SplitHTTPHeader(line); k == "Cookie" || k == "cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				val[e.Name] = e.Value
			}
		}

		if k, cookieRaw := SplitHTTPHeader(line); strings.ToLower(k) == "set-cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				val[e.Name] = e.Value
			}
		}

		return line
	})
	return val
}

// GetHTTPPacketCookiesFull is an auxiliary function used to obtain all Cookie values in the request message, and its return value is map[ string][]string, this is because the cookie may have multiple values with the same key name.
// Example:
// ```
// poc.GetHTTPPacketCookiesFull(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Host: pie.dev
//
// `) // gets all cookie values. Here Will return{"a":["b", "c"], "c":["d"]}
// ```
func GetHTTPPacketCookiesFull(packet []byte) (cookies map[string][]string) {
	val := make(map[string][]string)
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if k, cookieRaw := SplitHTTPHeader(line); k == "Cookie" || k == "cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				if _, ok := val[e.Name]; !ok {
					val[e.Name] = make([]string, 0)
				}
				val[e.Name] = append(val[e.Name], e.Value)
			}
		}

		if k, cookieRaw := SplitHTTPHeader(line); strings.ToLower(k) == "set-cookie" {
			existed := ParseCookie(cookieRaw)
			for _, e := range existed {
				if _, ok := val[e.Name]; !ok {
					val[e.Name] = make([]string, 0)
				}
				val[e.Name] = append(val[e.Name], e.Value)
			}
		}
		return line
	})
	return val
}

// GetHTTPPacketHeaders is an auxiliary function used to obtain all request headers in the request message, and its return value is map[string]string
// Example:
// ```
// poc.GetHTTPPacketCookiesFull(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Host: pie.dev
//
// `) // will be returned here to get all requests. header,{"Content-Type": "application/json", "Cookie": "a=b; a=c; c=d", "Host": "pie.dev"}
// ```
func GetHTTPPacketHeaders(packet []byte) (headers map[string]string) {
	val := make(map[string]string)
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if k, v := SplitHTTPHeader(line); k != "" {
			val[k] = v
		}
		return line
	})
	return val
}

// GetHTTPPacketHeadersFull is an auxiliary function used to obtain The return value of all request headers in the request message is map[string][]string. This is because the request header may have multiple values with the same key name.
// Example:
// ```
// poc.GetHTTPPacketHeadersFull(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Cookie: e=f
// Host: pie.dev
//
// `) // will be returned here to get all requests. header,{"Content-Type": ["application/json"], "Cookie": []"a=b; a=c; c=d", "e=f"], "Host": ["pie.dev"]}
// ```
func GetHTTPPacketHeadersFull(packet []byte) (headers map[string][]string) {
	val := make(map[string][]string)
	SplitHTTPPacket(packet, nil, nil, func(line string) string {
		if k, v := SplitHTTPHeader(line); k != "" {
			if _, ok := val[k]; !ok {
				val[k] = make([]string, 0)
			}
			val[k] = append(val[k], v)
		}
		return line
	})
	return val
}

// GetHTTPPacketHeaders is an auxiliary function used to obtain the request header specified in the request message. Its return value is string
// Example:
// ```
// poc.GetHTTPPacketCookiesFull(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Host: pie.dev
//
// `) // Get the Content-Type request header, which will return"application/json"
// ```
func GetHTTPPacketHeader(packet []byte, key string) (header string) {
	ret := GetHTTPPacketHeaders(packet)
	if ret == nil {
		return ""
	}

	fuzzResult := make(map[string]string)
	for headerKey, value := range ret {
		if key == headerKey {
			return value
		}
		if strings.ToLower(key) == strings.ToLower(headerKey) {
			fuzzResult[key] = value
		}
	}
	if len(fuzzResult) > 0 {
		return fuzzResult[key]
	}
	return ""
}

// GetHTTPPacketQueryParam is an auxiliary function used to obtain The GET request parameter specified in the request message, its return value is string
// Example:
// ```
// poc.GetHTTPPacketQueryParam(`GET /get?a=b&c=d HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// `, "a") // Get the value of GET request parameter a
// ```
func GetHTTPRequestQueryParam(packet []byte, key string) (paramValue string) {
	vals := GetHTTPRequestQueryParamFull(packet, key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// GetHTTPPacketBody is an auxiliary function, used to get the request body in the request message , its return value is bytes
// Example:
// ```
// poc.GetHTTPPacketBody(`POST /post HTTP/1.1
// Content-Type: application/json
// COntent-Length: 7
// Host: pie.dev
//
// a=b&c=d`) // Get the request header, here is b"a=b&c=d"
// ```
func GetHTTPPacketBody(packet []byte) (body []byte) {
	_, body = SplitHTTPHeadersAndBodyFromPacket(packet)
	return body
}

// will be added. GetHTTPPacketPostParam is an auxiliary function used to obtain the POST request parameters specified in the request message. Its return value is string
// Example:
// ```
// poc.GetHTTPPacketPostParam(`POST /post HTTP/1.1
// Content-Type: application/json
// COntent-Length: 7
// Host: pie.dev
//
// a=b&c=d`, "a") // Get POST request parameters The value of a
// ```
func GetHTTPRequestPostParam(packet []byte, key string) (paramValue string) {
	vals := GetHTTPRequestPostParamFull(packet, key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func GetHTTPPacketJSONValue(packet []byte, key string) any {
	return jsonpath.Find(GetHTTPPacketBody(packet), "$."+key)
}

func GetHTTPPacketJSONPath(packet []byte, key string) any {
	return jsonpath.Find(GetHTTPPacketBody(packet), key)
}

func GetHTTPRequestPostParamFull(packet []byte, key string) []string {
	body := GetHTTPPacketBody(packet)
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil
	}
	v, ok := vals[key]
	if ok {
		return v
	}
	return nil
}

func GetHTTPRequestQueryParamFull(packet []byte, key string) []string {
	u, err := ExtractURLFromHTTPRequestRaw(packet, false)
	if err != nil {
		return nil
	}
	val := u.Query()
	vals, ok := val[key]
	if ok {
		return vals
	}
	return []string{}
}

// GetAllHTTPPacketPostParams is an auxiliary function used to obtain all POST request parameters in the request message. Its return value is map[string]string, where the key is the parameter name. , the value is the parameter value
// Example:
// ```
// poc.GetAllHTTPPacketPostParams(`POST /post HTTP/1.1
// Content-Type: application/json
// COntent-Length: 7
// Host: pie.dev
//
// a=b&c=d`) // Get all POST request parameters
// ```
func GetAllHTTPRequestPostParams(packet []byte) (params map[string]string) {
	body := GetHTTPPacketBody(packet)
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil
	}
	ret := make(map[string]string)
	for k, v := range vals {
		ret[k] = v[len(v)-1]
	}
	return ret
}

// GetAllHTTPPacketQueryParams is an auxiliary function used to obtain all GET request parameters in the request message, and its return value is map[string]string, where the key is the parameter name and the value is the parameter value
// Example:
// ```
// poc.GetAllHTTPPacketQueryParams(`GET /get?a=b&c=d HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// `) // Get all GET request parameters
// ```
func GetAllHTTPRequestQueryParams(packet []byte) (params map[string]string) {
	u, err := ExtractURLFromHTTPRequestRaw(packet, false)
	if err != nil {
		return nil
	}
	vals := u.Query()
	ret := make(map[string]string)
	for k, v := range vals {
		ret[k] = v[len(v)-1]
	}
	return ret
}

// GetStatusCodeFromResponse is an auxiliary function used to obtain the status code in the response message, and its return value is int
// Example:
// ```
// poc.GetStatusCodeFromResponse(`HTTP/1.1 200 OK
// Content-Length: 5
//
// hello`) // gets the status code in the response message, and 200 will be returned here.
// ```
func GetStatusCodeFromResponse(packet []byte) (statusCode int) {
	SplitHTTPPacket(packet, nil, func(proto string, code int, codeMsg string) error {
		statusCode = code
		return nil
	})
	return statusCode
}

// GetHTTPRequestPathWithoutQuery is an auxiliary function used to obtain the path in the response message. The return value is string, which does not contain query.
// Example:
// ```
// poc.GetHTTPRequestPathWithoutQuery("GET /a/bc.html?a=1 HTTP/1.1\r\nHost: www.example.com\r\n\r\n") // /a/bc.html
// ```
func GetHTTPRequestPathWithoutQuery(packet []byte) (path string) {
	return strings.Split(GetHTTPRequestPath(packet), "?")[0]
}

// GetHTTPRequestPath is an auxiliary function. Used to obtain the path in the response message. The return value is a string, including query.
// Example:
// ```
// poc.GetHTTPRequestPath("GET /a/bc.html?a=1 HTTP/1.1\r\nHost: www.example.com\r\n\r\n") // /a/bc.html?a=1
// ```
func GetHTTPRequestPath(packet []byte) (path string) {
	SplitHTTPPacket(packet, func(method string, requestUri string, proto string) error {
		path = requestUri
		return io.EOF
	}, nil)
	return path
}

// GetHTTPRequestMethod is an auxiliary function used to get the request method in the request message, and its return value is string
// Example:
// ```
// poc.GetHTTPRequestMethod(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Host: pie.dev
//
// `) // Get the request method, which will return"GET"
// ```
func GetHTTPRequestMethod(packet []byte) (method string) {
	SplitHTTPPacket(packet, func(m string, _ string, _ string) error {
		method = m
		return utils.Error("normal")
	}, nil)
	return method
}

// . GetHTTPPacketFirstLine is an auxiliary function used to get the value of the first line in the HTTP message. Its return value is string, string, string.
// will be returned here. In the request message, its three return values are: request method, request URI, and protocol version.
// In the response message, its three return values are: protocol version, status code, status code description
// Example:
// ```
// poc.GetHTTPPacketFirstLine(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: a=b; a=c; c=d
// Host: pie.dev
//
// `) // gets the request method, request URI, and protocol version."GET", "/get", "HTTP/1.1"
// ```
func GetHTTPPacketFirstLine(packet []byte) (string, string, string) {
	packet = TrimLeftHTTPPacket(packet)
	reader := bufio.NewReader(bytes.NewBuffer(packet))
	var err error
	firstLineBytes, err := utils.BufioReadLine(reader)
	if err != nil {
		return "", "", ""
	}
	firstLineBytes = TrimSpaceHTTPPacket(firstLineBytes)

	var headers []string
	headers = append(headers, string(firstLineBytes))
	if bytes.HasPrefix(firstLineBytes, []byte("HTTP/")) {
		// response
		proto, code, codeMsg, _ := utils.ParseHTTPResponseLine(string(firstLineBytes))
		return proto, fmt.Sprint(code), codeMsg
	} else {
		// request
		method, requestURI, proto, _ := utils.ParseHTTPRequestLine(string(firstLineBytes))
		return method, requestURI, proto
	}
}

// ReplaceHTTPPacketBody is an auxiliary function used to change the request message and modify the request. body content, the first parameter is the modified request body content, and the second parameter is whether to transmit in chunks.
// Example:
// ```
// poc.ReplaceHTTPPacketBody(poc.BasicRequest(), "a=b") // Modify The content of the request body is a=b
// ```
func ReplaceHTTPPacketBodyFast(packet []byte, body []byte) []byte {
	var isChunked bool
	SplitHTTPHeadersAndBodyFromPacket(packet, func(line string) {
		if !isChunked {
			isChunked = IsChunkedHeaderLine(line)
		}
	})
	return ReplaceHTTPPacketBody(packet, body, isChunked)
}
