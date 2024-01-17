package lowhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/h2non/filetype"
	"github.com/yaklang/yaklang/common/log"
	utils "github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
)

var (
	charsetRegexp         = regexp.MustCompile(`(?i)charset\s*=\s*"?\s*([^\s;\n\r"]+)`)
	metaCharsetRegexp     = regexp.MustCompile(`(?i)meta[^<>]*?charset\s*=\s*['"]?\s*([^\s;\n\r'"]+)`)
	mimeCharsetRegexp     = regexp.MustCompile(`(?i)content-type:\s*[^\n]*charset\s*=\s*['"]?\s*([^\s;\n\r'"]+)`)
	contentTypeRegexp     = regexp.MustCompile(`(?i)content-type:\s*([^\r\n]*)`)
	contentEncodingRegexp = regexp.MustCompile(`(?i)content-encoding:\s*\w*\r?\n`)
)

// var contentLengthRegexpCase = regexp.MustCompile(`(?i)(content-length:\s*\w*\d+\r?\n)`)
func metaCharsetChanger(raw []byte) []byte {
	if len(raw) <= 0 {
		return raw
	}
	// This is critical. You need to remove the matched content.
	buf := bytes.NewBuffer(nil)
	var slash [][2]int
	lastEnd := 0
	for _, va := range metaCharsetRegexp.FindAllSubmatchIndex(raw, -1) {
		if len(va) > 3 {
			slash = append(slash, [2]int{lastEnd, va[2]})
			lastEnd = va[3]
		}
	}
	slash = append(slash, [2]int{lastEnd, len(raw)})
	for _, slashIndex := range slash {
		buf.Write(raw[slashIndex[0]:slashIndex[1]])
		if slashIndex[1] < len(raw) {
			buf.WriteString("utf-8")
		}
	}
	return buf.Bytes()
}

func CharsetToUTF8(bodyRaw []byte, mimeType string, originCharset string) ([]byte, string) {
	if len(bodyRaw) <= 0 {
		return bodyRaw, mimeType
	}

	originMT, kv, _ := mime.ParseMediaType(mimeType)
	newKV := make(map[string]string)
	for k, v := range kv {
		newKV[k] = v
	}

	originMTLower := strings.ToLower(originMT)
	var checkingGB18030 bool
	if ret := strings.HasPrefix(originMTLower, "text/"); ret || strings.Contains(originMTLower, "script") {
		newKV["charset"] = "utf-8"
		checkingGB18030 = ret
	}
	fixedMIME := mime.FormatMediaType(originMT, newKV)
	if fixedMIME != "" {
		mimeType = fixedMIME
	}

	var handledChineseEncoding bool
	parseFromMIME := func() ([]byte, error) {
		if kv != nil && len(kv) > 0 {
			if charsetStr, ok := kv["charset"]; ok && !handledChineseEncoding {
				encodingIns, name := charset.Lookup(strings.ToLower(charsetStr))
				if encodingIns != nil {
					raw, err := encodingIns.NewDecoder().Bytes(bodyRaw)
					if err != nil {
						return nil, utils.Errorf("decode [%s] from mime type failed: %s", name, err)
					}
					if len(raw) > 0 {
						return raw, nil
					}
				}
			}
		}
		return nil, utils.Errorf("cannot detect charset from mime")
	}
	var encodeHandler encoding.Encoding
	switch originCharset {
	case "gbk", "gb18030":
		// If the encoding cannot be detected, see if 18030 conforms to
		replaced, _ := codec.GB18030ToUtf8(bodyRaw)
		if replaced != nil {
			handledChineseEncoding = true
			bodyRaw = replaced
		}
	default:
		encodeHandler, _ = charsetPrescan(bodyRaw)
		if encodeHandler == nil && checkingGB18030 && !utf8.Valid(bodyRaw) {
			//// If the encoding cannot be detected, see if 18030 conforms to
			//replaced, err := codec.GB18030ToUtf8(bodyRaw)
			//if err != nil {
			//	log.Debugf("gb18030 to utf8 failed: %v", err)
			//}
			//if replaced != nil {
			//	handledChineseEncoding = true
			//	bodyRaw = replaced
			//}
		}
	}

	raw, _ := parseFromMIME()
	if len(raw) > 0 {
		idxs := charsetRegexp.FindStringSubmatchIndex(mimeType)
		if len(idxs) > 3 {
			start, end := idxs[2], idxs[3]
			prefix, suffix := mimeType[:start], mimeType[end:]
			if encodeHandler != nil {
				raw = metaCharsetChanger(raw)
			}
			return raw, fmt.Sprintf("%v%v%v", prefix, "utf-8", suffix)
		}
		return raw, mimeType
	}

	if encodeHandler != nil {
		raw, err := encodeHandler.NewDecoder().Bytes(bodyRaw)
		if err != nil {
			return bodyRaw, mimeType
		}
		if len(raw) <= 0 {
			return bodyRaw, mimeType
		}
		return metaCharsetChanger(raw), mimeType
	}

	return bodyRaw, mimeType
}

func GetOverrideContentType(bodyPrescan []byte, contentType string) (overrideContentType string, originCharset string) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	if strings.Contains(strings.ToLower(contentType), "charset") {
		if _, params, _ := mime.ParseMediaType(strings.ToLower(contentType)); params != nil {
			var ok bool
			originCharset, ok = params["charset"]
			_ = ok
			_ = originCharset
		}
	}
	if bodyPrescan != nil && contentType != "" && !filetype.IsMIME(bodyPrescan, contentType) {
		actuallyMIME, err := filetype.Match(bodyPrescan)
		if err != nil {
			log.Debugf("detect bodyPrescan type failed: %v", err)
		}

		if actuallyMIME.MIME.Value != "" {
			log.Debugf("really content-type met: %s, origin: %v", actuallyMIME.MIME.Value, contentType)
			overrideContentType = actuallyMIME.MIME.Value
		}
	}
	return overrideContentType, originCharset
}

// FixHTTPResponse Try to repair the incoming response and return the repaired response, response body and error
// Example:
// ```
// fixedResponse, body, err = str.FixHTTPResponse(b"HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=gbk\r\n\r\n<html>Hello</html>")
// ```
func FixHTTPResponse(raw []byte) (rsp []byte, body []byte, _ error) {
	// log.Infof("response raw: \n%v", codec.EncodeBase64(raw))

	isChunked := false
	// These two are used to handle special cases of encoding.
	var contentEncoding string
	var contentType string
	headers, body := SplitHTTPHeadersAndBodyFromPacket(raw, func(line string) {
		if strings.HasPrefix(strings.ToLower(line), "content-type:") {
			_, contentType = SplitHTTPHeader(line)
		}
		// Determine the content
		line = strings.ToLower(line)
		if strings.HasPrefix(line, "transfer-encoding:") && utils.IContains(line, "chunked") {
			isChunked = true
		}
		if strings.HasPrefix(line, "content-encoding:") {
			contentEncoding = line
		}
	})
	if headers == "" {
		return nil, nil, utils.Errorf("error for parsing http response")
	}
	headerBytes := []byte(headers)

	bodyRaw := body
	if bodyRaw != nil && isChunked {
		unchunked, chunkErr := codec.HTTPChunkedDecode(bodyRaw)
		if unchunked != nil {
			bodyRaw = unchunked
		} else {
			if chunkErr == nil {
				bodyRaw = []byte{}
			}
		}
	}
	if contentEncoding != "" {
		decodedBodyRaw, fixed := ContentEncodingDecode(contentEncoding, bodyRaw)
		if decodedBodyRaw != nil && fixed {
			// contents get decoded
			headerBytes = RemoveCEHeaders(headerBytes)
			bodyRaw = decodedBodyRaw
		}
	}

	// . If the bodyRaw is an image, it will not be processed. How to determine if it is an image?
	skipped := false
	if len(bodyRaw) > 0 {
		if utils.IsImage(bodyRaw) {
			skipped = true
		}
	}

	// and take the first few hundred bytes to detect the type.
	var bodyPrescan []byte
	if len(bodyRaw) > 200 {
		bodyPrescan = bodyRaw[:200]
	} else {
		bodyPrescan = bodyRaw[:]
	}
	overrideContentType, originCharset := GetOverrideContentType(bodyPrescan, contentType)
	/*originCharset is lower!!!*/
	_ = originCharset

	// are all solved to deal with encoding issues.
	if bodyRaw != nil && !skipped {
		var mimeType string
		_, params, _ := mime.ParseMediaType(contentType)
		ctUTF8 := false
		if raw, ok := params["charset"]; ok {
			raw = strings.ToLower(raw)
			ctUTF8 = raw == "utf-8" || raw == "utf8"
		}

		if overrideContentType == "" {
			// If the types are consistent and no replacement is needed, then only the content-type and encoding issues are dealt with
			bodyRaw, mimeType = CharsetToUTF8(bodyRaw, contentType, originCharset)
			if mimeType != contentType {
				headerBytes = ReplaceMIMEType(headerBytes, mimeType)
			}
			// . It is Js, but does not include UTF8, it stands to reason that it should be added to UTF8
			if utils.IContains(mimeType, "javascript") && !ctUTF8 && len(bodyRaw) > 0 {
				// . Do not mess with this order. Wrong, you must first determine whether it is UTF8, and then determine the Chinese encoding
				if !codec.IsUtf8(bodyRaw) {
					if codec.IsGBK(bodyRaw) {
						decoded, err := codec.GbkToUtf8(bodyRaw)
						if err == nil && len(decoded) > 0 {
							bodyRaw = decoded
						}
					} else {
						matchResult, _ := codec.CharDetectBest(bodyRaw)
						if matchResult != nil {
							switch strings.ToLower(matchResult.Charset) {
							case "gbk", "gb2312", "gb-2312", "gb18030", "windows-1252", "gb-18030", "windows1252":
								decoded, err := codec.GB18030ToUtf8(bodyRaw)
								if err == nil && len(decoded) > 0 {
									bodyRaw = decoded
								}
							}
						}
					}
				}
			}
		} else {
			log.Infof("replace content-type to: %s", overrideContentType)
			headerBytes = ReplaceMIMEType(headerBytes, overrideContentType)
		}
	}

	return ReplaceHTTPPacketBodyEx(headerBytes, bodyRaw, false, true), bodyRaw, nil
}

func ReplaceMIMEType(headerBytes []byte, mimeType string) []byte {
	if mimeType == "" {
		return headerBytes
	}

	idxs := contentTypeRegexp.FindSubmatchIndex(headerBytes)
	if len(idxs) > 3 {
		buf := bytes.NewBuffer(nil)
		buf.Write(headerBytes[:idxs[2]])
		buf.WriteString(mimeType)
		buf.Write(headerBytes[idxs[3]:])
		return buf.Bytes()
	} else {
		return AppendHeaderToHTTPPacket(headerBytes, "Content-Type: "+mimeType)
	}
}

func RemoveCEHeaders(headerBytes []byte) []byte {
	return contentEncodingRegexp.ReplaceAll(headerBytes, []byte{})
}

// ReplaceBody Replace the body in the original HTTP request message with the specified body, and specify whether it is chunked. Return a new HTTP request message
// Example:
// ```
// poc.ReplaceBody(`POST / HTTP/1.1
// Host: example.com
// Content-Length: 11
//
// hello world`, "hello yak", false)
// ```
func ReplaceHTTPPacketBody(raw []byte, body []byte, chunk bool) (newHTTPRequest []byte) {
	return ReplaceHTTPPacketBodyEx(raw, body, chunk, false)
}

func ReplaceHTTPPacketBodyEx(raw []byte, body []byte, chunk bool, forceCL bool) []byte {
	// Remove the left blank characters
	raw = TrimLeftHTTPPacket(raw)
	reader := bufio.NewReader(bytes.NewBuffer(raw))
	firstLineBytes, err := utils.BufioReadLine(reader)
	if err != nil {
		return raw
	}

	headers := []string{
		string(firstLineBytes),
	}

	// . Next, parse various Header
	for {
		lineBytes, err := utils.BufioReadLine(reader)
		if err != nil && err != io.EOF {
			break
		}
		line := string(lineBytes)
		line = strings.TrimSpace(line)

		// Header. After parsing,
		if line == "" {
			break
		}

		lineLower := strings.ToLower(line)
		// to remove chunked
		if strings.HasPrefix(lineLower, "transfer-encoding:") && utils.IContains(line, "chunked") {
			continue
		}

		//if strings.HasPrefix(lineLower, "content-encoding:") {
		//	headers = append(headers, fmt.Sprintf("Content-Encoding: %v", "identity"))
		//	continue
		//}

		// Set content-length
		if strings.HasPrefix(lineLower, "content-length") {
			continue
		}
		headers = append(headers, line)
	}

	// Null body
	if body == nil {
		raw := strings.Join(headers, CRLF) + CRLF + CRLF
		return []byte(raw)
	}

	// chunked
	if chunk {
		headers = append(headers, "Transfer-Encoding: chunked")
		body = codec.HTTPChunkedEncode(body)
		buf := bytes.NewBuffer(nil)
		for _, header := range headers {
			buf.WriteString(header)
			buf.WriteString(CRLF)
		}
		buf.WriteString(CRLF)
		buf.Write(body)
		return buf.Bytes()
	}

	if len(body) > 0 || forceCL {
		headers = append(headers, fmt.Sprintf("Content-Length: %d", len(body)))
	}
	buf := new(bytes.Buffer)
	for _, header := range headers {
		buf.WriteString(header)
		buf.WriteString(CRLF)
	}
	buf.WriteString(CRLF)
	buf.Write(body)
	return buf.Bytes()
}

// ParseBytesToHTTPResponse parses byte arrays into HTTP responses.
// Example:
// ```
// res, err := str.ParseBytesToHTTPResponse(b"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
// ```
func ParseBytesToHTTPResponse(res []byte) (rspInst *http.Response, err error) {
	if len(res) <= 0 {
		return nil, utils.Errorf("empty http response")
	}
	rsp, err := utils.ReadHTTPResponseFromBytes(res, nil)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
