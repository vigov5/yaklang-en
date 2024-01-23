package yakhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"

	"github.com/corpix/uarand"
	"github.com/davecgh/go-spew/spew"
)

var getDefaultHTTPClient = utils.NewDefaultHTTPClient

var ClientPool sync.Map

func GetClient(session interface{}) *http.Client {
	var client *http.Client

	if iClient, ok := ClientPool.Load(session); !ok {
		client = getDefaultHTTPClient()
		ClientPool.Store(session, client)
	} else {
		client = iClient.(*http.Client)
	}

	return client
}

// dump Get the original message referenced by the specified request structure reference or response structure reference, return the original message and error
// Example:
// ```
// req, err = http.NewRequest("GET", "http://www.yaklang.com", http.timeout(10))
// reqRaw, err = http.dump(req)
// rsp, err = http.Do(req)
// rspRaw, err = http.dump(rsp)
// ```
func dump(i interface{}) ([]byte, error) {
	return dumpWithBody(i, true)
}

// dumphead to obtain the original message header referenced by the specified request structure or response structure, and return the original message header and error
// Example:
// ```
// req, err = http.NewRequest("GET", "http://www.yaklang.com", http.timeout(10))
// reqHeadRaw, err = http.dumphead(req)
// rsp, err = http.Do(req)
// rspHeadRaw, err = http.dumphead(rsp)
// ```
func dumphead(i interface{}) ([]byte, error) {
	return dumpWithBody(i, false)
}

func dumpWithBody(i interface{}, body bool) ([]byte, error) {
	if body {
		isReq, raw, err := _dumpWithBody(i, body)
		if err != nil {
			return nil, err
		}
		if isReq {
			return lowhttp.FixHTTPPacketCRLF(raw, false), nil
		} else {
			raw, _, err := lowhttp.FixHTTPResponse(raw)
			if err != nil {
				return nil, err
			}
			return raw, nil
		}
	}
	_, raw, err := _dumpWithBody(i, body)
	return raw, err
}

func _dumpWithBody(i interface{}, body bool) (isReq bool, _ []byte, _ error) {
	if i == nil {
		return false, nil, utils.Error("nil interface for http.dump")
	}

	switch ret := i.(type) {
	case *http.Request:
		raw, err := utils.DumpHTTPRequest(ret, body)
		return true, raw, err
	case http.Request:
		return _dumpWithBody(&ret, body)
	case *http.Response:
		raw, err := utils.DumpHTTPResponse(ret, body)
		return false, raw, err
	case http.Response:
		return _dumpWithBody(&ret, body)
	case YakHttpResponse:
		return _dumpWithBody(ret.Response, body)
	case *YakHttpResponse:
		if ret == nil {
			return false, nil, utils.Error("nil YakHttpResponse for http.dump")
		}
		return _dumpWithBody(ret.Response, body)
	case YakHttpRequest:
		return _dumpWithBody(ret.Request, body)
	case *YakHttpRequest:
		return _dumpWithBody(ret.Request, body)
	default:
		return false, nil, utils.Errorf("error type for http.dump, Type: [%v]", reflect.TypeOf(i))
	}
}

// show obtains the original message specified by the request structure reference or response structure reference and outputs it to the standard output
// Example:
// ```
// req, err = http.NewRequest("GET", "http://www.yaklang.com", http.timeout(10))
// http.show(req)
// rsp, err = http.Do(req)
// http.show(rsp)
// ```
func httpShow(i interface{}) {
	rsp, err := dumpWithBody(i, true)
	if err != nil {
		log.Errorf("show failed: %s", err)
		return
	}
	fmt.Println(string(rsp))
}

// showhead. Get the original message header of the specified request structure reference or response structure reference and output it to the standard output.
// Example:
// ```
// req, err = http.NewRequest("GET", "http://www.yaklang.com", http.timeout(10))
// http.showhead(req)
// rsp, err = http.Do(req)
// http.showhead(rsp)
// ```
func showhead(i interface{}) {
	rsp, err := dumphead(i)
	if err != nil {
		log.Errorf("show failed: %s", err)
		return
	}
	fmt.Println(string(rsp))
}

type YakHttpRequest struct {
	*http.Request

	timeout    time.Duration
	proxies    func(req *http.Request) (*url.URL, error)
	redirector func(i *http.Request, reqs []*http.Request) bool
	session    interface{}
}

type HttpOption func(req *YakHttpRequest)

// timeout is a request option. Parameters, used to set the request timeout, in seconds.
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.timeout(10))
// ```
func timeout(f float64) HttpOption {
	return func(req *YakHttpRequest) {
		req.timeout = utils.FloatSecondDuration(f)
	}
}

// Raw Generate the request structure reference based on the original request message, return the request structure reference and error
// Note that this function will only generate a request structure reference and will not initiate a request
// ! Deprecated, Use poc.HTTP or poc.HTTPEx instead of
// Example:
// ```
// req, err = http.Raw("GET / HTTP/1.1\r\nHost: www.yaklang.com\r\n\r\n")
// ```
func rawRequest(i interface{}) (*http.Request, error) {
	var rawReq string
	switch ret := i.(type) {
	case []byte:
		rawReq = string(ret)
	case string:
		rawReq = ret
	case *http.Request:
		return ret, nil
	case http.Request:
		return &ret, nil
	case *YakHttpRequest:
		return ret.Request, nil
	case YakHttpRequest:
		return ret.Request, nil
	default:
		return nil, utils.Errorf("not a valid type: %v for req: %v", reflect.TypeOf(i), spew.Sdump(i))
	}

	return lowhttp.ParseStringToHttpRequest(rawReq)
}

// proxy is a request option parameter, used to set the proxy for one or more requests. When requesting, Find an available proxy according to the order. Use
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.proxy("http://127.0.0.1:7890", "http://127.0.0.1:8083"))
// ```
func WithProxy(values ...string) HttpOption {
	return func(req *YakHttpRequest) {
		values = utils.StringArrayFilterEmpty(values)
		if len(values) <= 0 {
			return
		}
		req.proxies = func(req *http.Request) (*url.URL, error) {
			return url.Parse(values[0])
		}
	}
}

// NewRequest generates a request structure reference based on the specified method and URL, and returns the request structure reference and error. Its first parameter is the URL, and then it can receive zero to multiple request options for Configure this request, such as setting the timeout, etc.
// Note that this function will only generate a request structure reference and will not initiate a request
// ! Deprecated.
// Example:
// ```
// req, err = http.NewRequest("GET", "http://www.yaklang.com", http.timeout(10))
// ```
func NewHttpNewRequest(method, url string, opts ...HttpOption) (*YakHttpRequest, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	rawReq := &YakHttpRequest{
		Request: req,
	}
	for _, op := range opts {
		op(rawReq)
	}
	return rawReq, nil
}

// GetAllBody Get the original response message
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com")
// raw = http.GetAllBody(rsp)
// ```
func GetAllBody(raw interface{}) []byte {
	switch r := raw.(type) {
	case *http.Response:
		if r == nil {
			return nil
		}

		if r.Body == nil {
			return nil
		}

		rspRaw, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil
		}
		return rspRaw
	case *YakHttpResponse:
		return GetAllBody(r.Response)
	default:
		log.Errorf("unsupported GetAllBody for %v", reflect.TypeOf(raw))
		return nil
	}
}

// params is a request option parameter, used to add/Specify GET parameters, which will URL encode the parameters
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.params("a=b"), http.params("c=d"))
// ```
func GetParams(i interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		req.URL.RawQuery = utils.UrlJoinParams(req.URL.RawQuery, i)
	}
}

// postparams is a request option parameter, used to add/Specifies the POST parameter, which will convert the parameters to the URL Encoding
// Example:
// ```
// rsp, err = http.Post("http://www.yaklang.com", http.postparams("a=b"), http.postparams("c=d"))
// ```
func PostParams(i interface{}) HttpOption {
	return Body(utils.UrlJoinParams("", i))
}

// Do Send a request based on the constructed request structure reference, return the response structure reference and error
// ! Deprecated.
// Example:
// ```
// req, err = http.Raw("GET / HTTP/1.1\r\nHost: www.yaklang.com\r\n\r\n")
// rsp, err = http.Do(req)
// ```
func Do(i interface{}) (*http.Response, error) {
	switch ret := i.(type) {
	case *http.Request:
		return Do(&YakHttpRequest{Request: ret})
	case http.Request:
		return Do(&YakHttpRequest{Request: &ret})
	case *YakHttpRequest:
	default:
		return nil, utils.Errorf("not a valid type: %v for req: %v", reflect.TypeOf(i), spew.Sdump(i))
	}
	req, _ := i.(*YakHttpRequest)

	var client *http.Client
	if req.session != nil {
		client = GetClient(req.session)
	} else {
		client = getDefaultHTTPClient()
	}

	httpTr := client.Transport.(*http.Transport)
	httpTr.Proxy = req.proxies
	if req.timeout > 0 {
		client.Timeout = req.timeout
	}
	client.Transport = httpTr
	if req.redirector != nil {
		client.CheckRedirect = func(i *http.Request, via []*http.Request) error {
			if !req.redirector(i, via) {
				return utils.Errorf("user aborted...")
			} else {
				return nil
			}
		}
	}
	return client.Do(req.Request)
}

// uarand returns a random User -Agent
// Example:
// ```
// ua = http.uarand()
// ```
func _getuarand() string {
	return uarand.GetRandom()
}

// header is a request option parameter, used to add/Specify the request header
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.header("AAA", "BBB"))
// ```
func Header(key, value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		req.Header.Set(fmt.Sprint(key), fmt.Sprint(value))
	}
}

// useragent is a request option parameter, used to specify the requested User-Agent
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.ua("yaklang-http"))
// ```
func UserAgent(value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		req.Header.Set("User-Agent", fmt.Sprint(value))
	}
}

// fakeua is a request option parameter, used to randomly specify the requested User-Agent Agent
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.fakeua())
// ```
func FakeUserAgent() HttpOption {
	return func(req *YakHttpRequest) {
		req.Header.Set("User-Agent", _getuarand())
	}
}

// redirect is a request option parameter, which receives the redirection processing function for custom redirection Orientation processing logic, returning true means continuing redirection, returning false means terminating redirection.
// redirection processing function is the current request structure reference, and the second parameter is the previous request structure reference
// Example:
// ```
// rsp, err = http.Get("http://pie.dev/redirect/3", http.redirect(func(r, vias) bool { return true })
// ```
func RedirectHandler(c func(r *http.Request, vias []*http.Request) bool) HttpOption {
	return func(req *YakHttpRequest) {
		req.redirector = c
	}
}

// noredirect is a request option parameter used to prohibit redirection
// Example:
// ```
// rsp, err = http.Get("http://pie.dev/redirect/3", http.noredirect())
// ```
func NoRedirect() HttpOption {
	return func(req *YakHttpRequest) {
		req.redirector = func(r *http.Request, vias []*http.Request) bool {
			return false
		}
	}
}

// header is a request option parameter, used to set Cookie
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.Cookie("a=b; c=d"))
// ```
func Cookie(value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		req.Header.Set("Cookie", fmt.Sprint(value))
	}
}

// body is a request option parameter, used to specify the request body in JSON format
// It will JSON serialize the incoming value, and then set the serialized value as the request body
// Example:
// ```
// rsp, err = http.Post("https://pie.dev/post", http.header("Content-Type", "application/json"), http.json({"a": "b", "c": "d"}))
// ```
func JsonBody(value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		body, err := json.Marshal(value)
		if err != nil {
			log.Errorf("yak http.json cannot marshal json: %v\n  ORIGIN: %v\n", err, string(spew.Sdump(value)))
			return
		}
		Body(body)(req)
		Header("Content-Type", "application/json")(req)
	}
}

// body is a request option Parameter, used to specify the request body.
// Example:
// ```
// rsp, err = http.Post("https://pie.dev/post", http.body("a=b&c=d"))
// ```
func Body(value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		var rc *bytes.Buffer
		switch ret := value.(type) {
		case string:
			rc = bytes.NewBufferString(ret)
		case []byte:
			rc = bytes.NewBuffer(ret)
		case io.Reader:
			all, err := ioutil.ReadAll(ret)
			if err != nil {
				return
			}
			rc = bytes.NewBuffer(all)
		default:
			rc = bytes.NewBufferString(fmt.Sprint(ret))
		}
		if rc != nil {
			req.ContentLength = int64(len(rc.Bytes()))
			buf := rc.Bytes()
			req.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
			req.Body, _ = req.GetBody()
		}
	}
}

// session is a request option parameter, used to specify the session based on the incoming value. Using the same value will use the same session, and the same session will Automatically reuse Cookie
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.session("request1"))
// ```
func Session(value interface{}) HttpOption {
	return func(req *YakHttpRequest) {
		req.session = value
	}
}

// Request initiates a request according to the specified URL. Its first parameter is the URL, and then it can receive zero to multiple request options. Used to configure this request, such as setting the request body, setting the timeout, etc.
// Return the response structure reference and error
// ! Deprecated, use poc.Do instead of
// Example:
// ```
// rsp, err = http.Request("POST","http://pie.dev/post", http.body("a=b&c=d"), http.timeout(10))
// ```
func httpRequest(method, url string, options ...HttpOption) (*YakHttpResponse, error) {
	reqRaw, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req := &YakHttpRequest{Request: reqRaw}
	for _, opt := range options {
		opt(req)
	}
	// client reuse to implement session management
	var client *http.Client
	if req.session != nil {
		client = GetClient(req.session)
	} else {
		client = getDefaultHTTPClient()
	}

	httpTr := client.Transport.(*http.Transport)
	httpTr.Proxy = req.proxies
	if req.timeout > 0 {
		client.Timeout = req.timeout
	}
	client.Transport = httpTr
	rspRaw, err := client.Do(req.Request)
	if err != nil {
		return nil, err
	}
	return &YakHttpResponse{Response: rspRaw}, nil
}

// Get Initiate a GET request according to the specified URL, its first parameter is URL, and then can receive zero to multiple request options, which are used to configure this request, such as setting timeout, etc.
// Return the response structure reference and error
// ! Deprecated, use poc.Get instead of
// Example:
// ```
// rsp, err = http.Get("http://www.yaklang.com", http.timeout(10))
// ```
func _get(url string, options ...HttpOption) (*YakHttpResponse, error) {
	return httpRequest("GET", url, options...)
}

// Post Based on the specified URL Initiate a POST request, its first parameter is the URL, and then you can receive zero to multiple request options, which are used to configure this request, such as setting the request body, setting the timeout, etc.
// Return the response structure reference and error
// ! Deprecated, use poc.Post instead
// Example:
// ```
// rsp, err = http.Post("http://pie.dev/post", http.body("a=b&c=d"), http.timeout(10))
// ```
func _post(url string, options ...HttpOption) (*YakHttpResponse, error) {
	return httpRequest("POST", url, options...)
}

type YakHttpResponse struct {
	*http.Response
}

func (y *YakHttpResponse) Json() interface{} {
	data := y.Data()
	if data == "" {
		return nil
	}
	var i interface{}
	err := json.Unmarshal([]byte(data), &i)
	if err != nil {
		log.Errorf("parse %v to json failed: %v", strconv.Quote(data), err)
		return ""
	}
	return i
}

func (y *YakHttpResponse) Data() string {
	if y.Response == nil {
		log.Error("response empty")
		return ""
	}

	if y.Response.Body == nil {
		return ""
	}

	body, _ := ioutil.ReadAll(y.Response.Body)
	y.Response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

func (y *YakHttpResponse) GetHeader(key string) string {
	return y.Response.Header.Get(key)
}

func (y *YakHttpResponse) Raw() []byte {
	raw, _ := dumpWithBody(y, true)
	return raw
}

var HttpExports = map[string]interface{}{
	// to obtain the native Raw request package
	"Raw": rawRequest,

	// shortcut
	"Get":     _get,
	"Post":    _post,
	"Request": httpRequest,

	// Do and Request combination Initiate a request
	"Do":         Do,
	"NewRequest": NewHttpNewRequest,

	// Get the response content response
	"GetAllBody": GetAllBody,

	// debugging information
	"dump":     dump,
	"show":     httpShow,
	"dumphead": dumphead,
	"showhead": showhead,

	// ua
	"ua":        UserAgent,
	"useragent": UserAgent,
	"fakeua":    FakeUserAgent,

	// header
	"header": Header,

	// cookie
	"cookie": Cookie,

	// body
	"body": Body,

	// json
	"json": JsonBody,

	// urlencode params are different from body, this will encode
	// params For get request
	// data Request
	"params":     GetParams,
	"postparams": PostParams,

	// proxy
	"proxy": WithProxy,

	// timeout
	"timeout": timeout,

	// redirect
	"redirect":   RedirectHandler,
	"noredirect": NoRedirect,

	// session
	"session": Session,

	"uarand": _getuarand,
}
