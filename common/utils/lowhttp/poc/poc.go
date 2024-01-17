package poc

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/yaklang/yaklang/common/utils/cli"

	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yak/yaklib/yakhttp"

	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
)

const (
	defaultWaitTime    = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

// for export
var (
	PoCOptWithSource       = WithSource
	PoCOptWithRuntimeId    = WithRuntimeId
	PoCOptWithFromPlugin   = WithFromPlugin
	PoCOptWithSaveHTTPFlow = WithSave
	PoCOptWithProxy        = WithProxy
)

type _pocConfig struct {
	Host                 string
	Port                 int
	ForceHttps           bool
	ForceHttp2           bool
	Timeout              time.Duration
	RetryTimes           int
	RetryInStatusCode    []int
	RetryNotInStatusCode []int
	RetryWaitTime        time.Duration
	RetryMaxWaitTime     time.Duration
	RedirectTimes        int
	NoRedirect           bool
	Proxy                []string
	FuzzParams           map[string][]string
	NoFixContentLength   bool
	JsRedirect           bool
	RedirectHandler      func(bool, []byte, []byte) bool
	Session              interface{} // session identifier such as the request message. You can use any object
	SaveHTTPFlow         bool
	Source               string
	Username             string
	Password             string

	// packetHandler
	PacketHandler []func([]byte) []byte

	// websocket opt
	// annotation: Is the Websocket connection enabled?
	Websocket bool

	// . This is the
	// parameter value for POST request. Parameter one is the bytes of data.
	// . The second parameter is the cancellation function. Calling it will forcibly disconnect the websocket.
	WebsocketHandler func(i []byte, cancel func())
	// , the means to obtain the Websocket client, if connected Successfully, the Websocket client here
	// . You can directly c.WriteText to write the data
	WebsocketClientHandler func(c *lowhttp.WebsocketClient)

	FromPlugin string
	RuntimeId  string
}

func (c *_pocConfig) ToLowhttpOptions() []lowhttp.LowhttpOpt {
	var opts []lowhttp.LowhttpOpt
	if c.Host != "" {
		opts = append(opts, lowhttp.WithHost(c.Host))
	}
	if c.Port != 0 {
		opts = append(opts, lowhttp.WithPort(c.Port))
	}
	if c.ForceHttps {
		opts = append(opts, lowhttp.WithHttps(c.ForceHttps))
	}
	opts = append(opts, lowhttp.WithHttp2(c.ForceHttp2))
	if c.Timeout > 0 {
		opts = append(opts, lowhttp.WithTimeout(c.Timeout))
	}
	if c.RetryTimes > 0 {
		opts = append(opts, lowhttp.WithRetryTimes(c.RetryTimes))
	}
	if len(c.RetryInStatusCode) > 0 {
		opts = append(opts, lowhttp.WithRetryInStatusCode(c.RetryInStatusCode))
	}
	if len(c.RetryNotInStatusCode) > 0 {
		opts = append(opts, lowhttp.WithRetryNotInStatusCode(c.RetryNotInStatusCode))
	}
	if c.RetryWaitTime > 0 {
		opts = append(opts, lowhttp.WithRetryWaitTime(c.RetryWaitTime))
	}
	if c.RetryMaxWaitTime > 0 {
		opts = append(opts, lowhttp.WithRetryMaxWaitTime(c.RetryMaxWaitTime))
	}
	if c.RedirectTimes > 0 {
		opts = append(opts, lowhttp.WithRedirectTimes(c.RedirectTimes))
	}
	if c.NoRedirect {
		opts = append(opts, lowhttp.WithRedirectTimes(0))
	}

	if c.Proxy != nil {
		opts = append(opts, lowhttp.WithProxy(c.Proxy...))
	}
	if c.FuzzParams != nil {
		log.Warnf("fuzz params is not nil, but not support now")
	}

	if c.NoFixContentLength {
		opts = append(opts, lowhttp.WithNoFixContentLength(c.NoFixContentLength))
	}
	if c.JsRedirect {
		opts = append(opts, lowhttp.WithJsRedirect(c.JsRedirect))
	}
	if c.RedirectHandler != nil {
		opts = append(opts, lowhttp.WithRedirectHandler(c.RedirectHandler))
	}
	if c.Session != nil {
		opts = append(opts, lowhttp.WithSession(c.Session))
	}
	if c.SaveHTTPFlow {
		opts = append(opts, lowhttp.WithSaveHTTPFlow(c.SaveHTTPFlow))
	}
	if c.Source != "" {
		opts = append(opts, lowhttp.WithSource(c.Source))
	}
	opts = append(opts, lowhttp.WithUsername(c.Username))
	opts = append(opts, lowhttp.WithPassword(c.Password))
	return opts
}

func NewDefaultPoCConfig() *_pocConfig {
	config := &_pocConfig{
		Host:                   "",
		Port:                   0,
		ForceHttps:             false,
		ForceHttp2:             false,
		Timeout:                15 * time.Second,
		RetryTimes:             0,
		RetryInStatusCode:      []int{},
		RetryNotInStatusCode:   []int{},
		RetryWaitTime:          defaultWaitTime,
		RetryMaxWaitTime:       defaultMaxWaitTime,
		RedirectTimes:          5,
		NoRedirect:             false,
		Proxy:                  nil,
		FuzzParams:             nil,
		NoFixContentLength:     false,
		JsRedirect:             false,
		RedirectHandler:        nil,
		Session:                nil,
		SaveHTTPFlow:           consts.GLOBAL_HTTP_FLOW_SAVE.IsSet(),
		Source:                 "",
		Websocket:              false,
		WebsocketHandler:       nil,
		WebsocketClientHandler: nil,
		PacketHandler:          make([]func([]byte) []byte, 0),
	}
	return config
}

type PocConfig func(c *_pocConfig)

// params is a request option parameter, used For using the passed in value when requesting, it should be noted that it can be convenient to use str.f() instead of
// Example:
// rsp, req, err = poc.HTTP(x`POST /post HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// {"key": "{{params(a)}}"}`, poc.params({"a":"bbb"})) // . The actual POST parameter sent is{"key": "bbb"}
func WithParams(i interface{}) PocConfig {
	return func(c *_pocConfig) {
		c.FuzzParams = utils.InterfaceToMap(i)
	}
}

// redirectHandler is a request option parameter used as a redirect processing function. If this option is set, the function will be called when redirecting. If the function returns true, the redirect will continue, otherwise it will not redirect. . The first parameter is whether to use https protocol, the second parameter is the original request message, and the third parameter is the original response message.
// Example:
// ```
// count = 3
// poc.Get("https://pie.dev/redirect/5", poc.redirectHandler(func(https, req, rsp) {
// count--
// return count >= 0
// })) // to pie.edv Request, use custom redirectHandler function, use count control, redirect up to 3 times
// ```
func WithRedirectHandler(i func(isHttps bool, req, rsp []byte) bool) PocConfig {
	return func(c *_pocConfig) {
		c.RedirectHandler = i
	}
}

// retryTimes is a request option parameter used to specify the number of retries when the request fails. It needs to be used with retryInStatusCode or retryNotInStatusCode to set the response code to retry
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.retryTimes(5), poc.retryInStatusCode(500, 502)) // initiates a request to example.com. If the response status code is 500 or 502, it will be retried, with a maximum of 5 retries.
// ```
func WithRetryTimes(t int) PocConfig {
	return func(c *_pocConfig) {
		c.RetryTimes = t
	}
}

// retryInStatusCode is a request option parameter, used to specify the To retry in the case of certain response status codes, you need to use
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.retryTimes(5), poc.retryInStatusCode(500, 502)) // initiates a request to example.com. If the response status code is 500 or 502, it will be retried, with a maximum of 5 retries.
// ```
func WithRetryInStatusCode(codes ...int) PocConfig {
	return func(c *_pocConfig) {
		c.RetryInStatusCode = codes
	}
}

// retryNotInStatusCode is a request option parameter used to specify retry without certain response status codes.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.retryTimes(5), poc.retryNotInStatusCode(200)) // to initiate a request to example.com, if the response status code is not equal to 200, retry, at most Retry
// ```
func WithRetryNotInStausCode(codes ...int) PocConfig {
	return func(c *_pocConfig) {
		c.RetryNotInStatusCode = codes
	}
}

// retryWaitTime is a request option parameter used to specify the minimum waiting time during retry. It needs to be used with retryTimes. The default is 0.1 seconds.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.retryTimes(5), poc.retryNotInStatusCode(200), poc.retryWaitTime(0.1)) // initiates a request to example.com. If the response status code is not If it is equal to 200, it will be retried, with a maximum of 5 retries, and a minimum wait of 0.1 seconds when retrying.
// ```
func WithRetryWaitTime(f float64) PocConfig {
	return func(c *_pocConfig) {
		c.RetryWaitTime = utils.FloatSecondDuration(f)
	}
}

// retryMaxWaitTime is a request option parameter, used to specify the maximum waiting time during retry. It needs to be used with retryTimes. The default is 2 seconds.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.retryTimes(5), poc.retryNotInStatusCode(200), poc.retryWaitTime(2)) // used to get the Websocket data. Initiate a request to example.com. If the response status code is not equal to 200 Then retry, up to 5 retries, and wait up to 2 seconds when retrying.
// ```
func WithRetryMaxWaitTime(f float64) PocConfig {
	return func(c *_pocConfig) {
		c.RetryMaxWaitTime = utils.FloatSecondDuration(f)
	}
}

// redirectTimes is a request option parameter used to specify the maximum number of redirects. The default is 5 times
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.redirectTimes(5)) // initiates a request to example.com. If the response is redirected to other links, the redirection will be automatically tracked up to 5 times.
// ```
func WithRedirectTimes(t int) PocConfig {
	return func(c *_pocConfig) {
		c.RedirectTimes = t
	}
}

// noFixContentLength is a request option parameter used to specify whether to repair the Content-Length field in the response message. The default is false and the Content-Length will be repaired automatically. Field
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.noFixContentLength()) // initiates a request to example.com. If the Content-Length field in the response message is incorrect or does not exist, also
// ```
func WithNoFixContentLength(b bool) PocConfig {
	return func(c *_pocConfig) {
		c.NoFixContentLength = b
	}
}

// noRedirect is a request option parameter, used to specify whether to track Redirect, the default is false, and the redirection will be automatically tracked.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.noRedirect()) // initiates a request to example.com. If the response is redirected to other links, the redirect will not be automatically tracked.
// ```
func WithNoRedirect(b bool) PocConfig {
	return func(c *_pocConfig) {
		c.NoRedirect = b
	}
}

// proxy is a request option parameter, used to specify the proxy used in the request. Multiple proxies can be specified. The system proxy
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.proxy("http://127.0.0.1:7890")) // example. com initiates a request, using http://127.0.0.1:7890 Proxy
// ```
func WithProxy(proxies ...string) PocConfig {
	return func(c *_pocConfig) {
		data := utils.StringArrayFilterEmpty(proxies)
		if len(data) > 0 {
			c.Proxy = proxies
		}
	}
}

// https is a request Option parameter, used to specify whether to use https protocol, the default is false, that is, use http protocol
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.https(true)) // initiates a request to example.com, using the https protocol.
// ```
func WithForceHTTPS(isHttps bool) PocConfig {
	return func(c *_pocConfig) {
		c.ForceHttps = isHttps
	}
}

// http2 is a request option parameter used to specify whether to use the http2 protocol. The default is false, which means using the http1 protocol
// Example:
// ```
// poc.Get("https://www.example.com", poc.http2(true), poc.https(true)) // should be repaired to initiate a request to www.example.com, using the http2 protocol
// ```
func WithForceHTTP2(isHttp2 bool) PocConfig {
	return func(c *_pocConfig) {
		c.ForceHttp2 = isHttp2
	}
}

// timeout is a request option parameter used to specify the read timeout, the default is 15 seconds
// Example:
// ```
// poc.Get("https://www.example.com", poc.timeout(15)) // initiates a request to www.baidu.com, and the read timeout is It is 15 seconds.
// ```
func WithTimeout(f float64) PocConfig {
	return func(c *_pocConfig) {
		c.Timeout = utils.FloatSecondDuration(f)
	}
}

// host is a request option parameter used to specify the actual requested host. If the request option is not set, the actual requested host will be determined based on the Host field in the original request message.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.host("yaklang.com")) // actually requests yaklang.com.
// ```
func WithHost(h string) PocConfig {
	return func(c *_pocConfig) {
		c.Host = h
	}
}

func WithRuntimeId(r string) PocConfig {
	return func(c *_pocConfig) {
		c.RuntimeId = r
	}
}

func WithFromPlugin(b string) PocConfig {
	return func(c *_pocConfig) {
		c.FromPlugin = b
	}
}

// websocket is a request option parameter used to allow the link to be upgraded to websocket. The request sent at this time should be a websocket handshake request.
// Example:
// ```
// rsp, req, err = poc.HTTP(`GET / HTTP/1.1
// Connection: Upgrade
// Upgrade: websocket
// Sec-Websocket-Version: 13
// Sec-Websocket-Extensions: permessage-deflate; client_max_window_bits
// Host: echo.websocket.events
// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7
// Sec-Websocket-Key: L31R1As+71fwuXqhwhABuA==`,
//
//	poc.proxy("http://127.0.0.1:7890"), poc.websocketFromServer(func(rsp, cancel) {
//		    dump(rsp)
//		}), poc.websocketOnClient(func(c) {
//		    c.WriteText("123")
//		}), poc.websocket(true),
//
// )
// time.Sleep(100)
// ```
func WithWebsocket(w bool) PocConfig {
	return func(c *_pocConfig) {
		c.Websocket = w
	}
}

// websocketOnClient is a request option parameter, which receives a callback function. This function has one parameter, which is the WebsocketClient structure. Through this structure, data can be sent to the server.
// Example:
// ```
// rsp, req, err = poc.HTTP(`GET / HTTP/1.1
// Connection: Upgrade
// Upgrade: websocket
// Sec-Websocket-Version: 13
// Sec-Websocket-Extensions: permessage-deflate; client_max_window_bits
// Host: echo.websocket.events
// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7
// Sec-Websocket-Key: L31R1As+71fwuXqhwhABuA==`,
//
//	poc.proxy("http://127.0.0.1:7890"), poc.websocketFromServer(func(rsp, cancel) {
//		    dump(rsp)
//		}), poc.websocketOnClient(func(c) {
//		    c.WriteText("123")
//		}), poc.websocket(true),
//
// )
// time.Sleep(100)
// ```
func WithWebsocketHandler(w func(i []byte, cancel func())) PocConfig {
	return func(c *_pocConfig) {
		c.WebsocketHandler = w
	}
}

// . websocketOnClient is a request option parameter. It receives a callback function. This function has a The parameter is the WebsocketClient structure through which data can be sent to the server.
// Example:
// ```
// rsp, req, err = poc.HTTP(`GET / HTTP/1.1
// Connection: Upgrade
// Upgrade: websocket
// Sec-Websocket-Version: 13
// Sec-Websocket-Extensions: permessage-deflate; client_max_window_bits
// Host: echo.websocket.events
// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7
// Sec-Websocket-Key: L31R1As+71fwuXqhwhABuA==`,
//
//	poc.proxy("http://127.0.0.1:7890"), poc.websocketFromServer(func(rsp, cancel) {
//		    dump(rsp)
//		}), poc.websocketOnClient(func(c) {
//		    c.WriteText("123")
//		}), poc.websocket(true),
//
// )
// time.Sleep(100)
// ```
func WithWebsocketClientHandler(w func(c *lowhttp.WebsocketClient)) PocConfig {
	return func(c *_pocConfig) {
		c.WebsocketClientHandler = w
	}
}

// port is a request option parameter used to specify the actual requested port. If this is not set, Request option, the actual requested port will be determined based on the Host field in the original request message.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.host("yaklang.com"), poc.port(443), poc.https(true)) // . In fact, it requests the 443 port
// ```
func WithPort(port int) PocConfig {
	return func(c *_pocConfig) {
		c.Port = port
	}
}

// jsRedirect is a request option parameter used to specify whether to track JS redirection. The default is false. That is, JS redirection will not be automatically tracked.
// Example:
// ```
// poc.HTTP(poc.BasicRequest(), poc.redirectTimes(5), poc.jsRedirect(true)) // initiates a request to www.baidu.com. If the response is repeated Directing to other links will also automatically track JS redirects, with a maximum of 5 redirects.
// ```
func WithJSRedirect(b bool) PocConfig {
	return func(c *_pocConfig) {
		c.JsRedirect = b
	}
}

// The identifier of the session, which can be used with any object.
// Example:
// ```
// poc.Get("https://pie.dev/cookies/set/AAA/BBB", poc.session("test")) // initiates the first request to pie.dev, which will set a cookie named AAA with a value of BBB.
// rsp, req, err = poc.Get("https://pie.dev/cookies", poc.session("test")) // initiates a second request to pie.dev. This request will output all cookies, and you can see the cookies set by the first request. Already exists
// ```
func WithSession(i interface{}) PocConfig {
	return func(c *_pocConfig) {
		c.Session = i
	}
}

// save is a request option parameter, used to specify whether to save the record of this request in the database. The default is true, which means it will be saved to the database.
// Example:
// ```
// poc.Get("https://exmaple.com", poc.save(true)) // with retryTimes to initiate a request to example.com, and the request will be saved in the database.
// ```
func WithSave(i bool) PocConfig {
	return func(c *_pocConfig) {
		c.SaveHTTPFlow = i
	}
}

// source is a request option parameter used to identify the source of the request when the request record is saved to the database.
// Example:
// ```
// poc.Get("https://exmaple.com", poc.save(true), poc.source("test")) // initiates a request to example.com and saves the request to the database, indicating that the source of the request is test. The
// ```
func WithSource(i string) PocConfig {
	return func(c *_pocConfig) {
		c.Source = i
	}
}

// replaceFirstLine is a request option parameter, used to change the request message and modify the first line (ie, request method, request path, protocol version)
// Example:
// ```
// poc.Get("https://exmaple.com", poc.replaceFirstLine("GET /test HTTP/1.1")) // . Initiate a request to example.com, modify the first line of the request message, and request/test path
// ```
func WithReplaceHttpPacketFirstLine(firstLine string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketFirstLine(packet, firstLine)
		},
		)
	}
}

// replaceMethod is a request option parameter, used to change the request message, modify the request method
// Example:
// ```
// poc.Options("https://exmaple.com", poc.replaceMethod("GET")) // Initiate a request to example.com, modify the request method to GET
// ```
func WithReplaceHttpPacketMethod(method string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketMethod(packet, method)
		},
		)
	}
}

// replaceHeader is a request option parameter, used to change the request message and modify the request header, if it does not exist.
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.replaceHeader("AAA", "BBB")) // initiates a request to pie.dev and modifies the value of the AAA request header to BBB. There is no AAA request header here, so the request header will be added.
// ```
func WithReplaceHttpPacketHeader(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketHeader(packet, key, value)
		},
		)
	}
}

// replaceHost is a request option parameter, used to change the request message and modify the Host request header. If it does not exist, it will be added. In fact, replaceHeader ("Host", the abbreviation of host),
// Example:
// ```
// poc.Get("https://yaklang.com/", poc.replaceHost("www.yaklang.com")) // initiates a request to yaklang.com, modifying the value of the Host request header to www.yaklang.com
// ```
func WithReplaceHttpPacketHost(host string) PocConfig {
	return WithReplaceHttpPacketHeader("Host", host)
}

// replaceBasicAuth is a request option parameter, used to change the request message and modify the Authorization request header to the ciphertext of the basic authentication. If it does not exist, it will be added. It is actually replaceHeader ("Authorization", codec.EncodeBase64(username + ":" , the abbreviation of
// Example:
// ```
// poc.Get("https://pie.dev/basic-auth/admin/password", poc.replaceBasicAuth("admin", "password")) // pie.dev initiates a request for basic authentication, and will get the 200 response status code
// ```
func WithReplaceHttpPacketBasicAuth(username, password string) PocConfig {
	return WithReplaceHttpPacketHeader("Authorization", "Basic "+codec.EncodeBase64(username+":"+password))
}

// replaceCookie is a request option parameter, used to change the request message and modify the Cookie request header. Value, if it does not exist, it will be increased.
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.replaceCookie("aaa", "bbb")) // initiates a request to pie.dev. There is no cookie value for aaa, so
// ```
func WithReplaceHttpPacketCookie(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketCookie(packet, key, value)
		},
		)
	}
}

// replaceBody is a request option parameter, used to change the request message and modify the request body content. , the first parameter is the modified request body content, and the second parameter is whether to transmit in chunks
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.replaceBody("a=b", false)) // to pie. .dev initiates a request and modifies the request body content to a=b
// ```
func WithReplaceHttpPacketBody(body []byte, chunk bool) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketBody(packet, body, chunk)
		},
		)
	}
}

// replacePath is a request option parameter, used to change the request message and modify the request path.
// Example:
// ```
// poc.Get("https://pie.dev/post", poc.replacePath("/get")) // to initiate a request to pie.dev, the actual request path is/get
// ```
func WithReplaceHttpPacketPath(path string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketPath(packet, path)
		},
		)
	}
}

func WithReplaceHttpPacketQueryParamRaw(rawQuery string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketQueryParamRaw(packet, rawQuery)
		},
		)
	}
}

// replaceQueryParam is a request option parameter used to change the request message. Modify GET request parameters, if they do not exist,
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.replaceQueryParam("a", "b")) // Initiate a request to pie.dev, add the GET request parameter a, the value is b
// ```
func WithReplaceHttpPacketQueryParam(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketQueryParam(packet, key, value)
		},
		)
	}
}

// replaceAllQueryParams is a request option parameter, used to change the request message and modify all GET request parameters. If it does not exist, it will be added. It receives a map[string]string type parameter, where key is the request parameter name, value is the request parameter value.
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.replaceAllQueryParams({"a":"b", "c":"d"})) // to initiate a request to pie.dev, add GET request parameter a with a value of b, add GET request parameter c with a value of d.
// ```
func WithReplaceAllHttpPacketQueryParams(values map[string]string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceAllHTTPPacketQueryParams(packet, values)
		},
		)
	}
}

// replacePostParam is A request option parameter, used to change the request message, modify the POST request parameter, if it does not exist,
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.replacePostParam("a", "b")) // initiates a request to pie.dev and deletes the POST request. Parameter a
// ```
func WithReplaceHttpPacketPostParam(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceHTTPPacketPostParam(packet, key, value)
		},
		)
	}
}

// replaceAllPostParams is a request option parameter, used to change the request message and modify all POST request parameters. If it does not exist, it will be added. It receives a map[string]string type parameter, where key is the POST request parameter name and value
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.replaceAllPostParams({"a":"b", "c":"d"})) // Initiate a request to pie.dev, add POST request parameter a, the value is b, POST request parameter c, the value is d
// ```
func WithReplaceAllHttpPacketPostParams(values map[string]string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.ReplaceAllHTTPPacketPostParams(packet, values)
		},
		)
	}
}

// appendHeader is a request option parameter, used to change the request message, add the request Header
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.appendHeader("AAA", "BBB")) // initiates a request to pie.dev and adds the value of the AAA request header to BBB
// ```
func WithAppendHeader(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketHeader(packet, key, value)
		},
		)
	}
}

func WithAppendHeaderIfNotExist(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketHeaderIfNotExist(packet, key, value)
		},
		)
	}
}

// appendHeaders is a request option parameter used to change the request message and add the request header
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.appendHeaders({"AAA": "BBB","CCC": "DDD"})) // initiates a request to pie.dev and adds the value of the AAA request header to BBB
// ```
func WithAppendHeaders(headers map[string]string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			for key, value := range headers {
				packet = lowhttp.AppendHTTPPacketHeader(packet, key, value)
			}
			return packet
		},
		)
	}
}

// appendCookie is a request option parameter used to change the request message and add a cookie request. The value
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.appendCookie("aaa", "bbb")) // initiates to pie.dev Request, add cookie key-value pair aaa:bbb
// ```
func WithAppendCookie(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketCookie(packet, key, value)
		},
		)
	}
}

// appendQueryParam is a request option parameter, used to change the request message. Add GET request parameter
// Example:
// ```
// poc.Get("https://pie.dev/get", poc.appendQueryParam("a", "b")) // Initiate a request to pie.dev, add the GET request parameter a, the value is b
// ```
func WithAppendQueryParam(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketQueryParam(packet, key, value)
		},
		)
	}
}

// appendPostParam is a request option parameter, used to change the request message, add the POST request parameter
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.appendPostParam("a", "b")) // initiates a request to pie.dev and deletes the POST request. Parameter a
// ```
func WithAppendPostParam(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketPostParam(packet, key, value)
		},
		)
	}
}

// appendPath is a request option parameter used to change the request message and add the request path after the existing request path.
// Example:
// ```
// poc.Get("https://yaklang.com/docs", poc.appendPath("/api/poc")) // initiates a request to yaklang.com. The actual request path is/docs/api/poc
// ```
func WithAppendHttpPacketPath(path string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketPath(packet, path)
		},
		)
	}
}

// appendFormEncoded is a request option parameter used to change the request message, add The form
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.appendFormEncoded("aaa", "bbb")) // initiates a request to pie.dev and adds a POST request form, where aaa is the key and bbb is the value.
// ```
func WithAppendHttpPacketFormEncoded(key, value string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketFormEncoded(packet, key, value)
		},
		)
	}
}

// appendUploadFile is a request option parameter, used to change the request message and add the uploaded file in the request body. The first parameter is the form name and the second parameter is the file name. , the third parameter is the file content, and the fourth parameter is an optional parameter, which is the file type (Content-Type).
// Example:
// ```
// poc.Post("https://pie.dev/post", poc.appendUploadFile("file", "phpinfo.php", "<?php phpinfo(); ?>", "image/jpeg"))// initiates a request to pie.dev and adds a POST request form. The file name is phpinfo.php, and the content is<?php phpinfo(); ?>, and the file type is image./jpeg
// ```
func WithAppendHttpPacketUploadFile(fieldName, fileName string, fileContent interface{}, contentType ...string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.AppendHTTPPacketUploadFile(packet, fieldName, fileName, fileContent, contentType...)
		},
		)
	}
}

// . deleteHeader is a request option parameter, used to change the request message and delete the request header.
// Example:
// ```
// poc.HTTP(`GET /get HTTP/1.1
// Content-Type: application/json
// AAA: BBB
// Host: pie.dev
//
// `, poc.deleteHeader("AAA"))// Initiate a request to pie.dev, delete the AAA request header
// ```
func WithDeleteHeader(key string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.DeleteHTTPPacketHeader(packet, key)
		},
		)
	}
}

// deleteCookie is a request option parameter used to change the request message and delete the value in the cookie.
// Example:
// ```
// poc.HTTP(`GET /get HTTP/1.1
// Content-Type: application/json
// Cookie: aaa=bbb; ccc=ddd
// Host: pie.dev
//
// `, poc.deleteCookie("aaa"))// Initiate a request to pie.dev, delete aaa in the cookie
// ```
func WithDeleteCookie(key string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.DeleteHTTPPacketCookie(packet, key)
		},
		)
	}
}

// deleteQueryParam is a request option parameter, used to change the request message, delete the GET request parameter
// Example:
// ```
// poc.HTTP(`GET /get?a=b&c=d HTTP/1.1
// Content-Type: application/json
// Host: pie.dev
//
// `, poc.deleteQueryParam("a")) // initiates a request to pie.dev, deletes the GET request parameter a
// ```
func WithDeleteQueryParam(key string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.DeleteHTTPPacketQueryParam(packet, key)
		},
		)
	}
}

// . deletePostParam is a request option parameter, used to change the request report. Text, delete the POST request parameter
// Example:
// ```
// poc.HTTP(`POST /post HTTP/1.1
// Content-Type: application/json
// Content-Length: 7
// Host: pie.dev
//
// a=b&c=d`, poc.deletePostParam("a")) // initiates a request to pie.dev, and deletes the POST request parameter a.
// ```
func WithDeletePostParam(key string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.DeleteHTTPPacketPostParam(packet, key)
		},
		)
	}
}

// deleteForm is a request option parameter, used to change the request message and delete the POST request form.
// Example:
// ```
// poc.HTTP(`POST /post HTTP/1.1
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
// --------------------------OFHnlKtUimimGcXvRSxgCZlIMAyDkuqsxeppbIFm--`, poc.deleteForm("aaa")) // initiates a request to pie.dev and deletes the POST request form aaa.
// ```
func WithDeleteForm(key string) PocConfig {
	return func(c *_pocConfig) {
		c.PacketHandler = append(c.PacketHandler, func(packet []byte) []byte {
			return lowhttp.DeleteHTTPPacketForm(packet, key)
		},
		)
	}
}

func fixPacketByConfig(packet []byte, config *_pocConfig) []byte {
	for _, fixFunc := range config.PacketHandler {
		packet = fixFunc(packet)
	}
	return packet
}

func handleUrlAndConfig(urlStr string, opts ...PocConfig) (*_pocConfig, error) {
	// poc module receives proxy and affects
	proxy := cli.CliStringSlice("proxy")
	config := NewDefaultPoCConfig()
	config.Proxy = proxy
	for _, opt := range opts {
		opt(config)
	}

	host, port, err := utils.ParseStringToHostPort(urlStr)
	if err != nil {
		return config, utils.Errorf("parse url failed: %s", err)
	}

	if port == 443 {
		config.ForceHttps = true
	}

	if config.Host == "" {
		config.Host = host
	}

	if config.Port == 0 {
		config.Port = port
	}

	if config.NoRedirect {
		config.RedirectTimes = 0
	}

	if config.RetryTimes < 0 {
		config.RetryTimes = 0
	}
	return config, nil
}

func handleRawPacketAndConfig(i interface{}, opts ...PocConfig) ([]byte, *_pocConfig, error) {
	var packet []byte
	switch ret := i.(type) {
	case string:
		packet = []byte(ret)
	case []byte:
		packet = ret
	case http.Request:
		r := &ret
		lowhttp.FixRequestHostAndPort(r)
		raw, err := utils.DumpHTTPRequest(r, true)
		if err != nil {
			return nil, nil, utils.Errorf("dump request out failed: %s", err)
		}
		packet = raw
	case *http.Request:
		lowhttp.FixRequestHostAndPort(ret)
		raw, err := utils.DumpHTTPRequest(ret, true)
		if err != nil {
			return nil, nil, utils.Errorf("dump request out failed: %s", err)
		}
		packet = raw
	case *yakhttp.YakHttpRequest:
		raw, err := utils.DumpHTTPRequest(ret.Request, true)
		if err != nil {
			return nil, nil, utils.Errorf("dump request out failed: %s", err)
		}
		packet = raw
	default:
		return nil, nil, utils.Errorf("cannot support: %s", reflect.TypeOf(i))
	}

	// poc module receives proxy and affects
	proxy := cli.CliStringSlice("proxy")
	config := NewDefaultPoCConfig()
	config.Proxy = proxy
	for _, opt := range opts {
		opt(config)
	}

	// . Modify the packet according to the config
	packet = fixPacketByConfig(packet, config)

	// first The data packet
	if config.FuzzParams != nil && len(config.FuzzParams) > 0 {
		packets, err := mutate.QuickMutate(string(packet), consts.GetGormProfileDatabase(), mutate.MutateWithExtraParams(config.FuzzParams))
		if err != nil {
			return nil, config, utils.Errorf("fuzz parameters failed: %v\n\nParams: \n%v", err, spew.Sdump(config.FuzzParams))
		}
		if len(packets) <= 0 {
			return nil, config, utils.Error("fuzzed packets empty!")
		}

		packet = []byte(packets[0])
	}

	u, err := lowhttp.ExtractURLFromHTTPRequestRaw(packet, config.ForceHttps)
	if err != nil {
		return nil, config, utils.Errorf("extract url failed: %s", err)
	}

	host, port, err := utils.ParseStringToHostPort(u.String())
	if err != nil {
		return nil, config, utils.Errorf("parse url failed: %s", err)
	}

	if port == 443 {
		config.ForceHttps = true
	}

	if config.Host == "" {
		config.Host = host
	}

	if config.Port == 0 {
		config.Port = port
	}

	if config.NoRedirect {
		config.RedirectTimes = 0
	}

	if config.RetryTimes < 0 {
		config.RetryTimes = 0
	}

	return packet, config, nil
}

func pochttp(packet []byte, config *_pocConfig) (*lowhttp.LowhttpResponse, error) {
	if config.Websocket {
		if config.Timeout <= 0 {
			config.Timeout = 15 * time.Second
		}
		wsCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		c, err := lowhttp.NewWebsocketClient(
			packet,
			lowhttp.WithWebsocketTLS(config.ForceHttps),
			lowhttp.WithWebsocketProxy(strings.Join(config.Proxy, ",")),
			lowhttp.WithWebsocketWithContext(wsCtx),
			lowhttp.WithWebsocketFromServerHandler(func(bytes []byte) {
				if config.WebsocketHandler != nil {
					config.WebsocketHandler(bytes, cancel)
				} else {
					spew.Dump(bytes)
				}
			}),
			lowhttp.WithWebsocketHost(config.Host),
			lowhttp.WithWebsocketPort(config.Port),
		)
		c.StartFromServer()
		if config.WebsocketClientHandler != nil {
			config.WebsocketClientHandler(c)
			c.Wait()
		}
		if err != nil {
			return nil, errors.Wrap(err, "websocket handshake failed")
		}
		return &lowhttp.LowhttpResponse{
			RawPacket: c.Response,
		}, nil
	}

	response, err := lowhttp.HTTP(
		lowhttp.WithHttps(config.ForceHttps),
		lowhttp.WithHost(config.Host),
		lowhttp.WithPort(config.Port),
		lowhttp.WithPacketBytes(packet),
		lowhttp.WithTimeout(config.Timeout),
		lowhttp.WithRetryTimes(config.RetryTimes),
		lowhttp.WithRetryInStatusCode(config.RetryInStatusCode),
		lowhttp.WithRetryNotInStatusCode(config.RetryNotInStatusCode),
		lowhttp.WithRetryWaitTime(config.RetryWaitTime),
		lowhttp.WithRetryMaxWaitTime(config.RetryMaxWaitTime),
		lowhttp.WithRedirectTimes(config.RedirectTimes),
		lowhttp.WithJsRedirect(config.JsRedirect),
		lowhttp.WithSession(config.Session),
		lowhttp.WithRedirectHandler(func(isHttps bool, req []byte, rsp []byte) bool {
			if config.RedirectHandler == nil {
				return true
			}
			return config.RedirectHandler(isHttps, req, rsp)
		}),
		lowhttp.WithNoFixContentLength(config.NoFixContentLength),
		lowhttp.WithHttp2(config.ForceHttp2),
		lowhttp.WithProxy(config.Proxy...),
		lowhttp.WithSaveHTTPFlow(config.SaveHTTPFlow),
		lowhttp.WithSource(config.Source),
		lowhttp.WithRuntimeId(config.RuntimeId),
		lowhttp.WithFromPlugin(config.FromPlugin),
	)
	return response, err
}

// HTTPEx is similar to HTTP. It sends a request and returns a response structure, request structure and error. Its first parameter can receive []byte, string, http.Request structure, and then it can receive zero to multiple request options, which are used to configure this request, such as setting a timeout or modifying the request. Message etc.
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// rsp, req, err = poc.HTTPEx(`GET / HTTP/1.1\r\nHost: www.yaklang.com\r\n\r\n`, poc.https(true), poc.replaceHeader("AAA", "BBB")) // sends a GET request based on HTTPS protocol to yaklang.com, and adds a request header AAA, whose value is BBB.
// desc(rsp) // View the available fields in the response structure
// ```
func HTTPEx(i interface{}, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	packet, config, err := handleRawPacketAndConfig(i, opts...)
	if err != nil {
		return nil, nil, err
	}
	response, err := pochttp(packet, config)
	if err != nil {
		return nil, nil, err
	}
	request, err := lowhttp.ParseBytesToHttpRequest(packet)
	if err != nil {
		return nil, nil, err
	}
	return response, request, nil
}

// BuildRequest is a tool function used to assist in building request messages. Its first parameter Can receive []byte, string, http.Request structure, then can receive zero to multiple request options, the options that modify the request message will be applied, and finally return the constructed request message
// Example:
// ```
// raw = poc.BuildRequest(poc.BasicRequest(), poc.https(true), poc.replaceHost("yaklang.com"), poc.replacePath("/docs/api/poc")) // . Construct a basic GET request, modify its Host to yaklang.com, and access the URI path to/docs/api/poc
// // raw = b"GET /docs/api/poc HTTP/1.1\r\nHost: www.yaklang.com\r\n\r\n"
// ```
func BuildRequest(i interface{}, opts ...PocConfig) []byte {
	packet, _, err := handleRawPacketAndConfig(i, opts...)
	if err != nil {
		log.Errorf("build request error: %s", err)
	}
	return packet
}

// HTTP sends the request and returns the original response message, original request message and error. Its first parameter can Receive []byte, string, http.Request structure, and then you can receive zero to multiple request options, which are used to configure this request, such as setting the timeout, or modifying the request message, etc.
// Example:
// ```
// poc.HTTP("GET / HTTP/1.1\r\nHost: www.yaklang.com\r\n\r\n", poc.https(true), poc.replaceHeader("AAA", "BBB")) // yaklang.com sends a request based on HTTPS GET request of the protocol, and add a request header AAA, whose value is BBB
// ```
func HTTP(i interface{}, opts ...PocConfig) (rsp []byte, req []byte, err error) {
	packet, config, err := handleRawPacketAndConfig(i, opts...)
	if err != nil {
		return nil, nil, err
	}
	response, err := pochttp(packet, config)
	return response.RawPacket, lowhttp.FixHTTPPacketCRLF(packet, config.NoFixContentLength), err
}

// Do sends a request with the specified request method to the specified URL and returns the response structure. Request structure and error. The first parameter is the request method, and the second parameter is the URL string. Next, you can receive zero to multiple request options, which are used to configure this request, such as setting a timeout. time, or modify the request message, etc.

// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Do("GET","https://yaklang.com", poc.https(true)) // sends an HTTPS protocol-based request to yaklang.com. GET request
// desc(rsp) // View the available fields in the response structure
// ```
func Do(method string, urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	config, err := handleUrlAndConfig(urlStr, opts...)
	if err != nil {
		return nil, nil, err
	}

	packet := lowhttp.UrlToRequestPacket(method, urlStr, nil, config.ForceHttps)
	packet = fixPacketByConfig(packet, config)

	response, err := pochttp(packet, config)
	if err != nil {
		return nil, nil, err
	}
	request, err := lowhttp.ParseBytesToHttpRequest(packet)
	if err != nil {
		return nil, nil, err
	}
	return response, request, nil
}

// with retryTimes. Get sends a GET request to the specified URL and returns the response structure, request structure and error. Its first parameter is the URL string, which can be received next. Zero to multiple request options, used to configure this request, such as setting a timeout, or modifying the request message, etc.
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Get("https://yaklang.com", poc.https(true)) // sends an HTTPS protocol-based request to yaklang.com. GET request
// desc(rsp) // View the available fields in the response structure
// ```
func DoGET(urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	return Do("GET", urlStr, opts...)
}

// Post sends to the specified URL. POST request and returns response structure, request structure and error. Its first parameter is a URL string, and then it can receive zero to multiple request options for configuring this request, such as setting a timeout. time, or modify the
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Post("https://yaklang.com", poc.https(true)) // to yaklang .com sends a POST request based on the HTTPS protocol.
// desc(rsp) // View the available fields in the response structure
// ```
func DoPOST(urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	return Do("POST", urlStr, opts...)
}

// Head sends a HEAD request to the specified URL and returns the response structure, request structure and error. Its first parameter is the URL string, and then it can receive zero to multiple request options for configuring this request. , such as setting the timeout, or modifying the request message, etc.
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Head("https://yaklang.com", poc.https(true)) // Send a HEAD request based on HTTPS protocol to yaklang.com
// desc(rsp) // View the available fields in the response structure
// ```
func DoHEAD(urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	return Do("HEAD", urlStr, opts...)
}

// Delete Send a DELETE request to the specified URL and return the response structure, request structure and error , its first parameter is a URL string, and then it can receive zero to multiple request options, which are used to configure this request, such as setting a timeout, or modifying the request message, etc.
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Delete("https://yaklang.com", poc.https(true)) // Send a DELETE request based on HTTPS protocol to yaklang.com
// desc(rsp) // View the available fields in the response structure
// ```
func DoDELETE(urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	return Do("DELETE", urlStr, opts...)
}

// Options Send an OPTIONS request to the specified URL and return the response structure, request structure and error. Its first parameter is the URL string, and then it can receive zero to Multiple request options, used to configure this request, such as setting the timeout, or modifying the request message, etc.
// will be added. You can use the desc function to view the available fields and methods in the structure.
// Example:
// ```
// poc.Options("https://yaklang.com", poc.https(true)) // sends an Options request based on the HTTPS protocol to yaklang.com.
// desc(rsp) // View the available fields in the response structure
// ```
func DoOPTIONS(urlStr string, opts ...PocConfig) (rspInst *lowhttp.LowhttpResponse, reqInst *http.Request, err error) {
	return Do("OPTIONS", urlStr, opts...)
}

// Websocket is actually equivalent to `poc.HTTP(..., poc.websocket(true))`, which is used to quickly send requests and establish websocket connections and return original response messages, original request messages and errors.
// Example:
// ```
// rsp, req, err = poc.Websocket(`GET / HTTP/1.1
// Connection: Upgrade
// Upgrade: websocket
// Sec-Websocket-Version: 13
// Sec-Websocket-Extensions: permessage-deflate; client_max_window_bits
// Host: echo.websocket.events
// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7
// Sec-Websocket-Key: L31R1As+71fwuXqhwhABuA==`,
//
//	poc.proxy("http://127.0.0.1:7890"), poc.websocketFromServer(func(rsp, cancel) {
//		    dump(rsp)
//		}), poc.websocketOnClient(func(c) {
//		    c.WriteText("123")
//		})
//
// )
// time.Sleep(100)
// ```
func DoWebSocket(raw interface{}, opts ...PocConfig) (rsp []byte, req []byte, err error) {
	opts = append(opts, WithWebsocket(true))
	return HTTP(raw, opts...)
}

// Split cuts the HTTP message and returns the response header and response body. Its first parameter is the original HTTP message. It can then receive zero to multiple callback functions. It calls back
// Example:
// ```
// poc.Split(`POST / HTTP/1.1
// Content-Type: application/json
// Host: www.example.com
//
// {"key": "value"}`, func(header) {
// dump(header)
// })
// ```
func split(raw []byte, hook ...func(line string)) (headers string, body []byte) {
	return lowhttp.SplitHTTPHeadersAndBodyFromPacket(raw, hook...)
}

// FixHTTPRequest Try to parse the incoming Repair the HTTP request message and return the repaired request
// Example:
// ```
// poc.FixHTTPRequest(b"GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
// ```
func fixHTTPRequest(raw []byte) []byte {
	return lowhttp.FixHTTPRequest(raw)
}

// FixHTTPResponse attempts to repair the incoming HTTP response message and returns the repaired response
// Example:
// ```
// poc.FixHTTPResponse(b"HTTP/1.1 200 OK\nContent-Length: 5\n\nhello")
// ```
func fixHTTPResponse(r []byte) []byte {
	rsp, _, _ := lowhttp.FixHTTPResponse(r)
	return rsp
}

// CurlToHTTPRequest Try to convert the curl command into an HTTP request message, the return value is bytes, That is, the converted HTTP request message
// Example:
// ```
// poc.CurlToHTTPRequest("curl -X POST -d 'a=b&c=d' http://example.com")
// ```
func curlToHTTPRequest(command string) (req []byte) {
	raw, err := lowhttp.CurlToHTTPRequest(command)
	if err != nil {
		log.Errorf(`CurlToHTTPRequest failed: %s`, err)
	}
	return raw
}

// . HTTPRequestToCurl attempts to convert the HTTP request message into a curl command. The first parameter is whether to use HTTPS, and the second parameter is the HTTP request message, and its return value is string, that is, the converted curl command
// Example:
// ```
// poc.HTTPRequestToCurl(true, "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
// ```
func httpRequestToCurl(https bool, raw any) (curlCommand string) {
	cmd, err := lowhttp.GetCurlCommand(https, utils.InterfaceToBytes(raw))
	if err != nil {
		log.Errorf(`http2curl.GetCurlCommand(req): %v`, err)
		return ""
	}
	return cmd.String()
}

var PoCExports = map[string]interface{}{
	"HTTP":          HTTP,
	"HTTPEx":        HTTPEx,
	"BasicRequest":  lowhttp.BasicRequest,
	"BasicResponse": lowhttp.BasicResponse,
	"BuildRequest":  BuildRequest,
	"Get":           DoGET,
	"Post":          DoPOST,
	"Head":          DoHEAD,
	"Delete":        DoDELETE,
	"Options":       DoOPTIONS,
	"Do":            Do,
	// websocket, you can directly reuse HTTP parameters.
	"Websocket": DoWebSocket,

	// options
	"host":                 WithHost,
	"port":                 WithPort,
	"retryTimes":           WithRetryTimes,
	"retryInStatusCode":    WithRetryInStatusCode,
	"retryNotInStatusCode": WithRetryNotInStausCode,
	"retryWaitTime":        WithRetryWaitTime,
	"retryMaxWaitTime":     WithRetryMaxWaitTime,
	"redirectTimes":        WithRedirectTimes,
	"noRedirect":           WithNoRedirect,
	"jsRedirect":           WithJSRedirect,
	"redirectHandler":      WithRedirectHandler,
	"https":                WithForceHTTPS,
	"http2":                WithForceHTTP2,
	"params":               WithParams,
	"proxy":                WithProxy,
	"timeout":              WithTimeout,
	"noFixContentLength":   WithNoFixContentLength,
	"session":              WithSession,
	"save":                 WithSave,
	"source":               WithSource,
	"websocket":            WithWebsocket,
	"websocketFromServer":  WithWebsocketHandler,
	"websocketOnClient":    WithWebsocketClientHandler,

	"replaceFirstLine":      WithReplaceHttpPacketFirstLine,
	"replaceMethod":         WithReplaceHttpPacketMethod,
	"replaceHeader":         WithReplaceHttpPacketHeader,
	"replaceHost":           WithReplaceHttpPacketHost,
	"replaceBasicAuth":      WithReplaceHttpPacketBasicAuth,
	"replaceCookie":         WithReplaceHttpPacketCookie,
	"replaceBody":           WithReplaceHttpPacketBody,
	"replaceAllQueryParams": WithReplaceAllHttpPacketQueryParams,
	"replaceAllPostParams":  WithReplaceAllHttpPacketPostParams,
	"replaceQueryParam":     WithReplaceHttpPacketQueryParam,
	"replacePostParam":      WithReplaceHttpPacketPostParam,
	"replacePath":           WithReplaceHttpPacketPath,
	"appendHeader":          WithAppendHeader,
	"appendHeaders":         WithAppendHeaders,
	"appendCookie":          WithAppendCookie,
	"appendQueryParam":      WithAppendQueryParam,
	"appendPostParam":       WithAppendPostParam,
	"appendPath":            WithAppendHttpPacketPath,
	"appendFormEncoded":     WithAppendHttpPacketFormEncoded,
	"appendUploadFile":      WithAppendHttpPacketUploadFile,
	"deleteHeader":          WithDeleteHeader,
	"deleteCookie":          WithDeleteCookie,
	"deleteQueryParam":      WithDeleteQueryParam,
	"deletePostParam":       WithDeletePostParam,
	"deleteForm":            WithDeleteForm,

	// split
	"Split":           split,
	"FixHTTPRequest":  fixHTTPRequest,
	"FixHTTPResponse": fixHTTPResponse,

	// packet helper
	"ReplaceBody":              lowhttp.ReplaceHTTPPacketBody,
	"FixHTTPPacketCRLF":        lowhttp.FixHTTPPacketCRLF,
	"HTTPPacketForceChunked":   lowhttp.HTTPPacketForceChunked,
	"ParseBytesToHTTPRequest":  lowhttp.ParseBytesToHttpRequest,
	"ParseBytesToHTTPResponse": lowhttp.ParseBytesToHTTPResponse,
	"ParseUrlToHTTPRequestRaw": lowhttp.ParseUrlToHttpRequestRaw,

	"ReplaceHTTPPacketMethod":         lowhttp.ReplaceHTTPPacketMethod,
	"ReplaceHTTPPacketFirstLine":      lowhttp.ReplaceHTTPPacketFirstLine,
	"ReplaceHTTPPacketHeader":         lowhttp.ReplaceHTTPPacketHeader,
	"ReplaceHTTPPacketBody":           lowhttp.ReplaceHTTPPacketBodyFast,
	"ReplaceHTTPPacketCookie":         lowhttp.ReplaceHTTPPacketCookie,
	"ReplaceHTTPPacketHost":           lowhttp.ReplaceHTTPPacketHost,
	"ReplaceHTTPPacketBasicAuth":      lowhttp.ReplaceHTTPPacketBasicAuth,
	"ReplaceAllHTTPPacketQueryParams": lowhttp.ReplaceAllHTTPPacketQueryParams,
	"ReplaceAllHTTPPacketPostParams":  lowhttp.ReplaceAllHTTPPacketPostParams,
	"ReplaceHTTPPacketQueryParam":     lowhttp.ReplaceHTTPPacketQueryParam,
	"ReplaceHTTPPacketPostParam":      lowhttp.ReplaceHTTPPacketPostParam,
	"ReplaceHTTPPacketPath":           lowhttp.ReplaceHTTPPacketPath,
	"AppendHTTPPacketHeader":          lowhttp.AppendHTTPPacketHeader,
	"AppendHTTPPacketCookie":          lowhttp.AppendHTTPPacketCookie,
	"AppendHTTPPacketQueryParam":      lowhttp.AppendHTTPPacketQueryParam,
	"AppendHTTPPacketPostParam":       lowhttp.AppendHTTPPacketPostParam,
	"AppendHTTPPacketPath":            lowhttp.AppendHTTPPacketPath,
	"AppendHTTPPacketFormEncoded":     lowhttp.AppendHTTPPacketFormEncoded,
	"AppendHTTPPacketUploadFile":      lowhttp.AppendHTTPPacketUploadFile,
	"DeleteHTTPPacketHeader":          lowhttp.DeleteHTTPPacketHeader,
	"DeleteHTTPPacketCookie":          lowhttp.DeleteHTTPPacketCookie,
	"DeleteHTTPPacketQueryParam":      lowhttp.DeleteHTTPPacketQueryParam,
	"DeleteHTTPPacketPostParam":       lowhttp.DeleteHTTPPacketPostParam,
	"DeleteHTTPPacketForm":            lowhttp.DeleteHTTPPacketForm,

	"GetAllHTTPPacketQueryParams": lowhttp.GetAllHTTPRequestQueryParams,
	"GetAllHTTPPacketPostParams":  lowhttp.GetAllHTTPRequestPostParams,
	"GetHTTPPacketQueryParam":     lowhttp.GetHTTPRequestQueryParam,
	"GetHTTPPacketPostParam":      lowhttp.GetHTTPRequestPostParam,
	"GetHTTPPacketCookieValues":   lowhttp.GetHTTPPacketCookieValues,
	"GetHTTPPacketCookieFirst":    lowhttp.GetHTTPPacketCookieFirst,
	"GetHTTPPacketCookie":         lowhttp.GetHTTPPacketCookie,
	"GetHTTPPacketContentType":    lowhttp.GetHTTPPacketContentType,
	"GetHTTPPacketCookies":        lowhttp.GetHTTPPacketCookies,
	"GetHTTPPacketCookiesFull":    lowhttp.GetHTTPPacketCookiesFull,
	"GetHTTPPacketHeaders":        lowhttp.GetHTTPPacketHeaders,
	"GetHTTPPacketHeadersFull":    lowhttp.GetHTTPPacketHeadersFull,
	"GetHTTPPacketHeader":         lowhttp.GetHTTPPacketHeader,
	"GetHTTPPacketBody":           lowhttp.GetHTTPPacketBody,
	"GetHTTPPacketFirstLine":      lowhttp.GetHTTPPacketFirstLine,
	"GetStatusCodeFromResponse":   lowhttp.GetStatusCodeFromResponse,
	"GetHTTPRequestMethod":        lowhttp.GetHTTPRequestMethod,
	"GetHTTPRequestPath":          lowhttp.GetHTTPRequestPath,
	// ext for path
	"GetHTTPRequestPathWithoutQuery": lowhttp.GetHTTPRequestPathWithoutQuery,
	// extract url
	"GetUrlFromHTTPRequest": lowhttp.GetUrlFromHTTPRequest,

	"CurlToHTTPRequest": curlToHTTPRequest,
	"HTTPRequestToCurl": httpRequestToCurl,
	"IsResponse":        lowhttp.IsResp,
}
