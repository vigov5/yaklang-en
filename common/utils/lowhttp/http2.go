package lowhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const transportDefaultConnFlow = 1 << 30
const defaultHeaderTableSize = 65536
const defaultStreamReceiveWindowSize = 1 << 20
const defaultMaxFrameSize = 1 << 14
const defaultMaxConcurrentStreamSize = 100
const defaultMaxHeaderListSize = 10 << 18

// requires cc.wmu be held
func h2FramerWriteHeaders(frame *http2.Framer, streamID uint32, endStream bool, maxFrameSize uint32, hdrs []byte) error {
	first := true // first frame written (HEADERS is first, then CONTINUATION)
	for len(hdrs) > 0 {
		chunk := hdrs
		if len(chunk) > int(maxFrameSize) {
			chunk = chunk[:maxFrameSize]
		}
		hdrs = hdrs[len(chunk):]
		endHeaders := len(hdrs) == 0
		if first {
			err := frame.WriteHeaders(http2.HeadersFrameParam{
				StreamID:      streamID,
				BlockFragment: chunk,
				EndStream:     endStream,
				EndHeaders:    endHeaders,
			})
			first = false
			if err != nil {
				return err
			}
		} else {
			err := frame.WriteContinuation(streamID, endHeaders, chunk)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func HTTPRequestToHTTP2(schema string, host string, conn net.Conn, raw []byte, noFixContentLength bool) ([]byte, error) {
	bw, br := bufio.NewWriter(conn), bufio.NewReader(conn)
	frame := http2.NewFramer(bw, br)
	frame.MaxHeaderListSize = 10 << 20

	// Write preface.
	_, err := bw.Write([]byte(http2.ClientPreface))
	if err != nil {
		return nil, utils.Errorf("yak.h2 preface failed: %s", err)
	}

	// . Write negotiation configuration settings
	// Most of these configurations come from
	err = frame.WriteSettings([]http2.Setting{
		{ID: http2.SettingEnablePush, Val: 0},
		{ID: http2.SettingInitialWindowSize, Val: transportDefaultConnFlow}, // is the default value.
		{ID: http2.SettingMaxConcurrentStreams, Val: 100},
		//{ID: http2.SettingMaxHeaderListSize, Val: 1000},
	}...)
	if err != nil {
		return nil, utils.Errorf("yak.h2 write setting failed")
	}

	// writes WindowUpdate.
	err = frame.WriteWindowUpdate(0, transportDefaultConnFlow)
	if err != nil {
		return nil, utils.Errorf("yak.h2 write windows update")
	}
	bw.Flush()

	// next processes the header in hpack form.
	var hbuf bytes.Buffer
	henc := hpack.NewEncoder(&hbuf)
	var requestHeaders []hpack.HeaderField
	f := func(k, v string) {
		requestHeaders = append(requestHeaders, hpack.HeaderField{Name: k, Value: v})
	}

	var methodReq = http.MethodGet
	// parse request
	_, body := SplitHTTPHeadersAndBodyFromPacketEx(raw, func(method string, requestUri string, proto string) error {
		f(":authority", host)
		if method == "" {
			method = http.MethodGet
		}
		methodReq = method
		f(":method", method)
		if !utils.AsciiEqualFold(method, "CONNECT") {
			f(":path", requestUri)
			f(":scheme", schema)
		}
		return nil
	}, func(line string) {
		result := strings.SplitN(line, ":", 2)
		if len(result) == 1 {
			f(strings.ToLower(result[0]), "")
		} else if len(result) == 2 {
			key := strings.ToLower(result[0])
			value := strings.TrimLeft(result[1], " ")
			switch key {
			case "host": // :authority
				var targetIndex = -1
				for index, h := range requestHeaders {
					if h.Name == ":authority" {
						targetIndex = index
						break
					}
				}
				if targetIndex >= 0 {
					requestHeaders[targetIndex].Value = value
				}
			case "content-length":
				// Need to automatically repair
				if noFixContentLength {
					f("content-length", value)
				}
			case "connection", "proxy-connection",
				"transfer-encoding", "upgrade",
				"keep-alive": // Logically speaking, these shouldnt appear, but who knows?

				/**
				case "cookie":
					// Per 8.1.2.5 To allow for better compression efficiency, the
					// Cookie header field MAY be split into separate header fields,
					// each with one or more cookie-pairs.
					//      . V1ll4n: However, multiple cookie pairs can also take effect without conflict. You can choose not to process it, then it will not be processed. !
					for {
						p := strings.IndexByte(value, ';')
						if p < 0 {
							if len(value) > 0 {
								f(key, value)
							}
							break
						}
						f("cookie", value[:p])
						p++
						for p+1 <= len(value) && value[p] == ' ' {
							p++
						}
						value = value[p:]
					}

				**/
			default:
				f(key, value)
			}
		}
	})
	cl := int64(len(body))
	if ShouldSendReqContentLength(methodReq, cl) && !noFixContentLength {
		f("content-length", strconv.FormatInt(cl, 10))
	}
	for _, h := range requestHeaders {
		henc.WriteField(h)
	}

	streamId := uint32(1)
	// streamId defaults to 1. Anyway, the connection will be disconnected.
	// cl. If it is 0, it means there is no follow-up. , the stream will not end
	// maxFrameSize just use the default value, the standard library is this
	endRequestStream := cl <= 0
	err = h2FramerWriteHeaders(frame, streamId, endRequestStream, 16<<10, hbuf.Bytes())
	if err != nil {
		return nil, utils.Errorf("yak.h2 framer write headers failed: %s", err)
	}

	// Clear the buffer
	err = bw.Flush()
	if err != nil {
		log.Warnf("h2 conn flush(header) error: %s", err)
	}
	if cl > 0 {
		_ = frame.WriteData(streamId, true, body)
		err = bw.Flush()
		if err != nil {
			log.Warnf("h2 conn flush(data) error: %s", err)
		}
	}

	var statusCode int
	var headers []hpack.HeaderField
	var responseBody bytes.Buffer
	// processes response information
	var pseudoHeaders []hpack.HeaderField
	var endStream bool
	var lastError []error

	// windowUpdate
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		frameResponse, err := frame.ReadFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnf("yak.h2(client) frame read failed: %s", err)
			lastError = append(lastError, err)
		}
		if frameResponse == nil {
			break
		}
		// log.Infof("h2 recv frametype: %v data: %v", reflect.TypeOf(frameResponse), frameResponse)
		switch ret := frameResponse.(type) {
		case *http2.HeadersFrame:
			hpackHeader := ret.HeaderBlockFragment()
			parsedHeaders, err := hpack.NewDecoder(4096*4, func(f hpack.HeaderField) {

			}).DecodeFull(hpackHeader)
			if err != nil {
				log.Errorf("yak.h2 hpack decode err: %s", err)
			}
			for _, h := range parsedHeaders {
				if h.IsPseudo() {
					pseudoHeaders = append(pseudoHeaders, h)
					if utils.AsciiEqualFold(h.Name[1:], "status") {
						statusCode, _ = strconv.Atoi(h.Value)
					}
					continue
				}
				headers = append(headers, h)
			}
			if ret.StreamEnded() {
				endStream = true
			}
		case *http2.DataFrame:
			responseBody.Write(ret.Data())
			if ret.StreamEnded() {
				endStream = true
			}
			frame.WriteWindowUpdate(0, transportDefaultConnFlow)
			frame.WriteWindowUpdate(ret.StreamID, transportDefaultConnFlow)
		case *http2.SettingsFrame:
			if !ret.IsAck() {
				frame.WriteSettingsAck()
			}
		default:
		}

		if endStream {
			break
		}
	}

	if statusCode <= 0 {
		if len(lastError) > 0 {
			return nil, utils.Errorf("yak.h2 read frame failed: %v", lastError)
		}
		return nil, utils.Error("yak.h2 cannot read statuscode")
	}

	firstLine := fmt.Sprintf("HTTP/2.0 %v %v", statusCode, http.StatusText(statusCode))
	var headerLine = make([]string, len(headers)+1)
	headerLine[0] = firstLine
	for index, h := range headers {
		headerLine[index+1] = fmt.Sprintf("%v: %v", h.Name, h.Value)
	}
	rawPacket := append([]byte(strings.Join(headerLine, CRLF)+CRLF+CRLF), responseBody.Bytes()...)
	fixedPacket, _, err := FixHTTPResponse(rawPacket)
	return fixedPacket, err

}

func HTTP2RequestToHTTP(framer *http2.Framer) ([]byte, error) {
	var statusCode int
	var headers []hpack.HeaderField
	var requestBody bytes.Buffer
	// processes response information
	var pseudoHeaders []hpack.HeaderField
	var endStream bool
	var lastError []error

	// windowUpdate
	for {
		frameRequest, err := framer.ReadFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnf("yak.h2 frame read failed: %s", err)
			lastError = append(lastError, err)
		}
		if frameRequest == nil {
			break
		}
		// log.Infof("h2 recv frametype: %v data: %v", reflect.TypeOf(frameResponse), frameResponse)
		switch ret := frameRequest.(type) {
		case *http2.HeadersFrame:
			hpackHeader := ret.HeaderBlockFragment()
			parsedHeaders, err := hpack.NewDecoder(4096*4, func(f hpack.HeaderField) {

			}).DecodeFull(hpackHeader)
			if err != nil {
				log.Errorf("yak.h2 hpack decode err: %s", err)
			}
			for _, h := range parsedHeaders {
				if h.IsPseudo() {
					pseudoHeaders = append(pseudoHeaders, h)
					if utils.AsciiEqualFold(h.Name[1:], "status") {
						statusCode, _ = strconv.Atoi(h.Value)
					}
					continue
				}
				headers = append(headers, h)
			}
			if ret.StreamEnded() {
				endStream = true
			}
		case *http2.DataFrame:
			requestBody.Write(ret.Data())
			if ret.StreamEnded() {
				endStream = true
			}
		case *http2.SettingsFrame:
			// ignore setting frame leave this for remote server to handle
		default:
		}

		if endStream {
			break
		}
	}

	if statusCode <= 0 {
		if len(lastError) > 0 {
			return nil, utils.Errorf("yak.h2 read frame failed: %v", lastError)
		}
		return nil, utils.Error("yak.h2 cannot read statuscode")
	}

	firstLine := fmt.Sprintf("HTTP/2.0 %v %v", statusCode, http.StatusText(statusCode))
	var headerLine = make([]string, len(headers)+1)
	headerLine[0] = firstLine
	for index, h := range headers {
		headerLine[index+1] = fmt.Sprintf("%v: %v", h.Name, h.Value)
	}
	spew.Dump(ReplaceHTTPPacketBody([]byte(strings.Join(headerLine, CRLF)+CRLF+CRLF), requestBody.Bytes(), false), nil)
	return ReplaceHTTPPacketBody([]byte(strings.Join(headerLine, CRLF)+CRLF+CRLF), requestBody.Bytes(), false), nil

}

func HTTP2ResponseToHTTP(frame *http2.Frame) ([]byte, error) {
	return nil, nil

}

//// requires cc.wmu be held
//func h2FramerWriteHeaders(frame *http2.Framer, streamID uint32, endStream bool, maxFrameSize int, hdrs []byte) error {
//	first := true // first frame written (HEADERS is first, then CONTINUATION)
//	for len(hdrs) > 0 {
//		chunk := hdrs
//		if len(chunk) > maxFrameSize {
//			chunk = chunk[:maxFrameSize]
//		}
//		hdrs = hdrs[len(chunk):]
//		endHeaders := len(hdrs) == 0
//		if first {
//			err := frame.WriteHeaders(http2.HeadersFrameParam{
//				StreamID:      streamID,
//				BlockFragment: chunk,
//				EndStream:     endStream,
//				EndHeaders:    endHeaders,
//			})
//			first = false
//			if err != nil {
//				return err
//			}
//		} else {
//			err := frame.WriteContinuation(streamID, endHeaders, chunk)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

type windowSizeControl struct {
	windowMutex   *sync.Mutex
	windowSize    int64
	windowChanged *sync.Cond
}

func newControl(size int64) *windowSizeControl {
	return &windowSizeControl{
		windowMutex:   new(sync.Mutex),
		windowSize:    size,
		windowChanged: new(sync.Cond),
	}
}

func (wc *windowSizeControl) increaseWindowSize(i int64) {
	wc.windowMutex.Lock()
	defer wc.windowMutex.Unlock()

	var notify = false
	if wc.windowSize <= 0 {
		notify = true
	}
	wc.windowSize += i
	if wc.windowSize > 0 && notify {
		wc.windowChanged.Broadcast()
	}
}

func (wc *windowSizeControl) decreaseWindowSize(i int64) {
	wc.windowMutex.Lock()
	wc.windowSize -= i
	defer wc.windowMutex.Unlock()
	if wc.windowSize <= 0 {
		log.Infof("window size is %v, waiting for window size changed", wc.windowSize)
		wc.windowChanged.Wait()
	}
}
