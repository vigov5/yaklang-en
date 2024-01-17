package lowhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	utils "github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
)

var (
	_contentLengthRE    = regexp.MustCompile(`(?i)Content-Length:(\s+)?(\d+)?\r?\n?`)
	_transferEncodingRE = regexp.MustCompile(`(?i)Transfer-Encoding:(\s+)?.*?(chunked).*?\r?\n?`)
	fetchBoundaryRegexp = regexp.MustCompile(`boundary\s?=\s?([^;]+)`)
)

// HTTPPacketForceChunked Forces the body of an HTTP message to chunked encoding
// Example:
// ```
// poc.HTTPPacketForceChunked(`POST / HTTP/1.1
// Host: example.com
// Content-Length: 11
//
// hello world`)
// ```
func HTTPPacketForceChunked(raw []byte) []byte {
	header, body := SplitHTTPHeadersAndBodyFromPacket(raw)
	return ReplaceHTTPPacketBodyEx([]byte(header), body, true, false)
}

func AppendHeaderToHTTPPacket(raw []byte, line string) []byte {
	header, body := SplitHTTPHeadersAndBodyFromPacket(raw)
	header = strings.TrimRight(header, "\r\n") + CRLF + strings.TrimSpace(line) + CRLF + CRLF
	return []byte(header + string(body))
}

// FixHTTPPacketCRLF fixes the CRLF problem of an HTTP message (normal messages end with \r\n at the end of each line, but some messages may have \n). If noFixLength is true, then Content-Length will not be repaired, otherwise it will try to repair Content-Length
// Example:
// ```
// poc.FixHTTPPacketCRLF(`POST / HTTP/1.1
// Host: example.com
// Content-Length: 11
//
// hello world`, false)
// ```
func FixHTTPPacketCRLF(raw []byte, noFixLength bool) []byte {
	// Remove the left blank characters
	raw = TrimLeftHTTPPacket(raw)
	if raw == nil || len(raw) == 0 {
		return nil
	}
	var isMultipart bool
	var haveChunkedHeader bool
	var haveContentLength bool
	var isRequest bool
	var isResponse bool
	var contentLengthIsNotRecommanded bool

	plrand := fmt.Sprintf("[[REPLACE_CONTENT_LENGTH:%v]]", utils.RandStringBytes(20))
	plrandHandled := false
	header, body := SplitHTTPPacket(
		raw,
		func(m, u, proto string) error {
			isRequest = true
			contentLengthIsNotRecommanded = utils.ShouldRemoveZeroContentLengthHeader(m)
			return nil
		},
		func(proto string, code int, codeMsg string) error {
			isResponse = true
			return nil
		},
		func(line string) string {
			key, value := SplitHTTPHeader(line)
			keyLower := strings.ToLower(key)
			valLower := strings.ToLower(value)
			if !isMultipart && keyLower == "content-type" && strings.HasPrefix(valLower, "multipart/form-data") {
				isMultipart = true
			}
			if !haveContentLength && strings.ToLower(key) == "content-length" {
				haveContentLength = true
				if noFixLength {
					return line
				}
				return fmt.Sprintf(`%v: %v`, key, plrand)
			}
			if !haveChunkedHeader && keyLower == "transfer-encoding" && strings.Contains(valLower, "chunked") {
				haveChunkedHeader = true
			}
			return line
		},
	)

	// cl te existed at the same time, handle smuggle!
	smuggleCase := isRequest && haveContentLength && haveChunkedHeader
	_ = smuggleCase

	// applying patch to restore CRLF at body
	// if `raw` has CRLF at body end (by design HTTP smuggle) and `noFixContentLength` is true
	//if bytes.HasSuffix(raw, []byte(CRLF+CRLF)) && noFixLength && len(body) > 0 && !smuggleCase {
	//	body = append(body, []byte(CRLF+CRLF)...)
	//}

	_ = isResponse
	handleChunked := haveChunkedHeader && !haveContentLength
	var restBody []byte
	if handleChunked {
		// chunked body is very complex
		// if multiRequest: extract and remove body suffix
		var bodyDecode []byte
		bodyDecode, restBody = codec.HTTPChunkedDecodeWithRestBytes(body)
		if len(bodyDecode) > 0 {
			readLen := len(body) - len(restBody)
			body = body[:readLen]
		}
	}

	/* boundary fix */
	if isRequest && isMultipart {
		boundary, fixed := FixMultipartBody(body)
		if boundary != "" {
			header = string(ReplaceMIMEType([]byte(header), "multipart/form-data; boundary="+boundary))
			body = fixed
		}
	}

	if !noFixLength && !haveChunkedHeader {
		if haveContentLength {
			// have CL
			// only cl && no chunked && fix length
			// fix content-length
			header = strings.Replace(header, plrand, strconv.Itoa(len(body)), 1)
		} else {
			bodyLength := len(body)
			if bodyLength > 0 {
				// no CL
				// fix content-length
				header = strings.TrimRight(header, CRLF)
				header += fmt.Sprintf("\r\nContent-Length: %v\r\n\r\n", bodyLength)
			} else {
				if !contentLengthIsNotRecommanded {
					header = strings.TrimRight(header, CRLF)
					header += "\r\nContent-Length: 0\r\n\r\n"
				}
			}
		}
		plrandHandled = true
	}

	if !plrandHandled && haveContentLength && !noFixLength {
		header = strings.Replace(header, plrand, strconv.Itoa(len(body)), 1)
	}

	var buf bytes.Buffer
	buf.Write([]byte(header))
	if len(body) > 0 {
		buf.Write(body)
	}
	if len(restBody) > 0 {
		buf.Write(restBody)
	}
	return buf.Bytes()
}

// FixHTTPRequest Try to parse the incoming Repair the HTTP request message and return the repaired request
// Example:
// ```
// str.FixHTTPRequest(b"GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
// ```
func FixHTTPRequest(raw []byte) []byte {
	return FixHTTPPacketCRLF(raw, false)
}

func DeletePacketEncoding(raw []byte) []byte {
	var encoding string
	var isChunked bool
	var buf bytes.Buffer
	_, body := SplitHTTPPacket(raw, func(method string, requestUri string, proto string) error {
		buf.WriteString(method + " " + requestUri + " " + proto + CRLF)
		return nil
	}, func(proto string, code int, codeMsg string) error {
		buf.WriteString(proto + " " + strconv.Itoa(code) + " " + codeMsg + CRLF)
		return nil
	}, func(line string) string {
		k, v := SplitHTTPHeader(line)
		ret := strings.ToLower(k)
		if ret == "content-encoding" {
			encoding = v
			return ""
		} else if ret == "transfer-encoding" && utils.IContains(v, "chunked") {
			isChunked = true
			return ""
		}
		buf.WriteString(line + CRLF)
		return line
	})
	buf.WriteString(CRLF)

	if isChunked {
		decBody, _ := codec.HTTPChunkedDecode(body)
		if len(decBody) > 0 {
			body = decBody
		}
	}

	decResult, fixed := ContentEncodingDecode(encoding, body)
	if fixed && len(decResult) > 0 {
		body = decResult
	}
	return ReplaceHTTPPacketBody(buf.Bytes(), body, false)
}

func ConvertHTTPRequestToFuzzTag(i []byte) []byte {
	var boundary string // If the packet is uploaded, the boundary will not be empty
	header, body := SplitHTTPHeadersAndBodyFromPacket(i, func(line string) {
		k, v := SplitHTTPHeader(strings.TrimSpace(line))
		switch strings.ToLower(k) {
		case "content-type":
			ctVal, params, _ := mime.ParseMediaType(v)
			if ctVal == "multipart/form-data" && params != nil && len(params) > 0 {
				boundary = params["boundary"]
			}
		}
	})

	if boundary != "" {
		// When uploading files
		reader := multipart.NewReader(bytes.NewBuffer(body), boundary)

		// Repair the data packet
		var buf bytes.Buffer
		fixedBody := multipart.NewWriter(&buf)
		fixedBody.SetBoundary(boundary)
		for {
			part, err := reader.NextRawPart()
			if err != nil {
				break
			}
			w, err := fixedBody.CreatePart(part.Header)
			if err != nil {
				log.Errorf("write part to new part failed: %s", err)
				continue
			}

			body, err := ioutil.ReadAll(part)
			if err != nil {
				log.Errorf("copy multipart-stream failed: %s", err)
			}
			if utf8.Valid(body) {
				w.Write(body)
			} else {
				w.Write([]byte(ToUnquoteFuzzTag(body)))
			}
		}
		fixedBody.Close()
		body = buf.Bytes()
		return ReplaceHTTPPacketBody([]byte(header), body, false)
	}

	if utf8.Valid(body) {
		return i
	}
	body = []byte(ToUnquoteFuzzTag(body))
	return ReplaceHTTPPacketBody([]byte(header), body, false)
}

const (
	printableMin = 32
	printableMax = 126
)

func ToUnquoteFuzzTag(i []byte) string {
	if utf8.Valid(i) {
		return string(i)
	}

	buf := bytes.NewBufferString(`{{unquote("`)
	for _, b := range i {
		if b >= printableMin && b <= printableMax {
			switch b {
			case '(':
				buf.WriteString(`\x28`)
			case ')':
				buf.WriteString(`\x29`)
			case '}':
				buf.WriteString(`\x7d`)
			case '{':
				buf.WriteString(`\x7b`)
			case '\\':
				buf.WriteString(`\\`)
			case '"':
				buf.WriteString(`\"`)
			default:
				buf.WriteByte(b)
			}
		} else {
			buf.WriteString(fmt.Sprintf(`\x%02x`, b))
		}
	}
	buf.WriteString(`")}}`)
	return buf.String()
}

//func FixHTTPRequest(raw []byte) []byte {
//	// Remove the left blank characters
//	raw = TrimLeftHTTPPacket(raw)
//
//	// fixes unreasonable headers.
//	if bytes.Contains(raw, []byte("Transfer-Encoding: chunked")) {
//		headers, body := SplitHTTPHeadersAndBodyFromPacket(raw)
//		headersRaw := fixInvalidHTTPHeaders([]byte(headers))
//		raw = append(headersRaw, body...)
//	}
//	raw = AddConnectionClosed(raw)
//
//	//
//	reader := bufio.NewReader(bytes.NewBuffer(raw))
//	firstLineBytes, _, err := reader.ReadLine()
//	if err != nil {
//		return raw
//	}
//
//	var headers = []string{
//		string(firstLineBytes),
//	}
//
//	// . Next, parse various Header
//	isChunked := false
//	for {
//		lineBytes, _, err := reader.ReadLine()
//		if err != nil && err != io.EOF {
//			return raw
//		}
//		line := string(lineBytes)
//		line = strings.TrimSpace(line)
//
//		// Header. After parsing,
//		if line == "" {
//			break
//		}
//
//		if strings.HasPrefix(line, "Transfer-Encoding:") && strings.Contains(line, "chunked") {
//			isChunked = true
//		}
//		headers = append(headers, line)
//	}
//	restBody, _ := ioutil.ReadAll(reader)
//
//	// Remove the original \r\n
//	if bytes.HasSuffix(restBody, []byte("\r\n\r\n")) {
//		restBody = restBody[:len(restBody)-4]
//	}
//	if bytes.HasSuffix(restBody, []byte("\n\n")) {
//		restBody = restBody[:len(restBody)-2]
//	}
//
//	// . Repair the content-length
//	var index = -1
//	emptyBody := bytes.TrimSpace(restBody) == nil
//	if emptyBody {
//		restBody = nil
//	}
//	contentLength := len(restBody)
//	if !isChunked && contentLength > 0 {
//		for i, r := range headers {
//			if strings.HasPrefix(r, "Content-Length:") {
//				index = i
//			}
//		}
//		if index < 0 {
//			headers = append(headers, fmt.Sprintf("Content-Length: %v", contentLength))
//		} else {
//			headers[index] = fmt.Sprintf("Content-Length: %v", contentLength)
//		}
//	}
//
//	var finalRaw []byte
//	// Add a new ending delimiter
//	if emptyBody {
//		finalRaw = []byte(strings.Join(headers, CRLF) + CRLF + CRLF)
//	} else {
//		finalRaw = []byte(strings.Join(headers, CRLF) + CRLF + CRLF + string(restBody) + CRLF + CRLF)
//	}
//
//	return finalRaw
//}

// ParseStringToHTTPRequest Parse the string into HTTP request
// Example:
// ```
// req, err = str.ParseStringToHTTPRequest("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
// ```
func ParseStringToHttpRequest(raw string) (*http.Request, error) {
	return ParseBytesToHttpRequest([]byte(raw))
}

var (
	contentTypeChineseCharset = regexp.MustCompile(`(?i)charset\s*=\s*['"]?(.*?)(gb[^'^"^\s]+)['"]?`)                      // 2 gkxxxx
	charsetInMeta             = regexp.MustCompile(`(?i)<\s*meta.*?(charset|content)\s*=\s*['"]?(.*?)(gb[^'^"^\s]+)['"]?`) // 3 gbxxx
)

// ParseUrlToHTTPRequestRaw. Parse the URL into the original HTTP request message and return whether it is HTTPS, the original request message and the error
// Example:
// ```
// ishttps, raw, err = poc.ParseUrlToHTTPRequestRaw("GET", "https://yaklang.com")
// ```
func ParseUrlToHttpRequestRaw(method string, i interface{}) (isHttps bool, req []byte, err error) {
	urlStr := utils.InterfaceToString(i)
	reqInst, err := http.NewRequest(strings.ToUpper(method), urlStr, http.NoBody)
	if err != nil {
		return false, nil, err
	}
	reqInst.Header.Set("User-Agent", consts.DefaultUserAgent)
	bytes, err := utils.HttpDumpWithBody(reqInst, true)
	return strings.HasPrefix(strings.ToLower(urlStr), "https://"), bytes, err
}

func CopyRequest(r *http.Request) *http.Request {
	if r == nil {
		return nil
	}
	raw, err := utils.HttpDumpWithBody(r, true)
	if err != nil {
		log.Warnf("copy request && Dump failed: %s", err)
	}
	if raw == nil {
		return nil
	}
	result, err := ParseBytesToHttpRequest(raw)
	if err != nil {
		log.Warnf("copy request && ParseBytesToHttpRequest failed: %s", err)
	}
	return result
}

// ParseBytesToHTTPRequest Parse the byte array into HTTP request
// Example:
// ```
// req, err := str.ParseBytesToHTTPRequest(b"GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
// ```
func ParseBytesToHttpRequest(raw []byte) (reqInst *http.Request, err error) {
	fixed := FixHTTPPacketCRLF(raw, false)
	if fixed == nil {
		return nil, io.EOF
	}
	req, readErr := utils.ReadHTTPRequestFromBytes(fixed)
	if readErr != nil {
		log.Errorf("read [standard] httpRequest failed: %s", readErr)
		return nil, readErr
	}
	return req, nil
}

func ExtractBoundaryFromBody(raw interface{}) string {
	bodyStr := strings.TrimSpace(utils.InterfaceToString(raw))
	if strings.HasPrefix(bodyStr, "--") && strings.HasSuffix(bodyStr, "--") {
		sc := bufio.NewScanner(bytes.NewBufferString(bodyStr))
		sc.Split(bufio.ScanLines)
		if !sc.Scan() {
			return ""
		}
		prefixWithBoundary := sc.Text()
		if strings.HasPrefix(prefixWithBoundary, "--") {
			return prefixWithBoundary[2:]
		}
		return ""
	}
	return ""
}

var (
	ReadHTTPRequestFromBytes        = utils.ReadHTTPRequestFromBytes
	ReadHTTPRequestFromBufioReader  = utils.ReadHTTPRequestFromBufioReader
	ReadHTTPResponseFromBytes       = utils.ReadHTTPResponseFromBytes
	ReadHTTPResponseFromBufioReader = utils.ReadHTTPResponseFromBufioReader
)

func ExtractStatusCodeFromResponse(raw []byte) int {
	var statusCode int
	SplitHTTPPacket(raw, nil, func(proto string, code int, codeMsg string) error {
		statusCode = code
		return utils.Error("abort")
	})
	return statusCode
}
