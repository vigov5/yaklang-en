package yaklib

import (
	"context"
	"net/http"

	"github.com/yaklang/yaklang/common/crep"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/utils/lowhttp/httpctx"
)

var MitmExports = map[string]interface{}{
	"Start":  startMitm,
	"Bridge": startBridge,

	"maxContentLength":     mitmMaxContentLength,
	"isTransparent":        mitmConfigIsTransparent,
	"context":              mitmConfigContext,
	"host":                 mitmConfigHost,
	"callback":             mitmConfigCallback,
	"hijackHTTPRequest":    mitmConfigHijackHTTPRequest,
	"hijackHTTPResponse":   mitmConfigHijackHTTPResponse,
	"hijackHTTPResponseEx": mitmConfigHijackHTTPResponseEx,
	"wscallback":           mitmConfigWSCallback,
	"wsforcetext":          mitmConfigWSForceTextFrame,
	"rootCA":               mitmConfigCertAndKey,
	"useDefaultCA":         mitmConfigUseDefault,
}

// Start Start A MITM (man-in-the-middle) proxy server, its first parameter is the port, and then it can receive zero to more option functions, which are used to affect the behavior of the middle-man proxy server.
// will be called. If the CA certificate and CA certificate are not specified. private key, then the built-in certificate and private key
// Example:
// ```
// mitm.Start(8080, mitm.host("127.0.0.1"), mitm.callback(func(isHttps, urlStr, req, rsp) { http.dump(req); http.dump(rsp)  })) // starts a man-in-the-middle proxy server and prints requests and responses to the standard output. Whether
// ```
func startMitm(
	port int,
	opts ...MitmConfigOpt,
) error {
	return startBridge(port, "", opts...)
}

type mitmConfig struct {
	ctx                context.Context
	host               string
	callback           func(isHttps bool, urlStr string, r *http.Request, rsp *http.Response)
	wsForceTextFrame   bool
	wscallback         func(data []byte, isRequest bool) interface{}
	mitmCert, mitmPkey []byte
	useDefaultMitmCert bool
	maxContentLength   int

	// is enabled. Transparent hijacking
	isTransparent            bool
	hijackRequest            func(isHttps bool, urlStr string, req []byte, forward func([]byte), reject func())
	hijackWebsocketDataFrame func(isHttps bool, urlStr string, req []byte, forward func([]byte), reject func())
	hijackResponse           func(isHttps bool, urlStr string, rsp []byte, forward func([]byte), reject func())
	hijackResponseEx         func(isHttps bool, urlStr string, req, rsp []byte, forward func([]byte), reject func())
}

type MitmConfigOpt func(config *mitmConfig)

// isTransparent is an option function used to specify whether the middleman proxy server turns on transparent hijacking mode. The default is false
// . When transparent mode is enabled, all traffic will be forwarded by default, and all callback functions will be ignored.
// Example:
// ```
// mitm.Start(8080, mitm.isTransparent(true))
// ```
func mitmConfigIsTransparent(b bool) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.isTransparent = b
	}
}

var MitmConfigContext = mitmConfigContext

// context is an option function that specifies the context of the man-in-the-middle proxy server
// Example:
// ```
// mitm.Start(8080, mitm.context(context.Background()))
// ```
func mitmConfigContext(ctx context.Context) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.ctx = ctx
	}
}

// useDefaultCA is an option function used to specify Whether the middleman proxy server uses built-in certificates and private keys, the default is true
// Default certificate and private key path: ~/yakit-projects/yak-mitm-ca.crt and ~/yakit-projects/yak-mitm-ca.key
// Example:
// ```
// mitm.Start(8080, mitm.useDefaultCA(true))
// ```
func mitmConfigUseDefault(t bool) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.useDefaultMitmCert = t
	}
}

// host is an option function used to specify the listening address of the middleman proxy server. The default is empty, that is, listening to all network cards
// Example:
// ```
// mitm.Start(8080, mitm.host("127.0.0.1"))
// ```
func mitmConfigHost(host string) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.host = host
	}
}

// callback is an option function used to specify the callback function of the middleman proxy server. When the request and response are received, Calling this callback function
// Example:
// ```
// mitm.Start(8080, mitm.callback(func(isHttps, urlStr, req, rsp) { http.dump(req); http.dump(rsp)  }))
// ```
func mitmConfigCallback(f func(bool, string, *http.Request, *http.Response)) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.callback = f
	}
}

// hijackHTTPRequest is an option function used to specify the request hijacking function of the middleman proxy server. When the request is received, the callback function
// . By calling the fourth parameter of this callback function, the request content can be modified. By calling this callback The fifth parameter of the function can discard the request
// Example:
// ```
// mitm.Start(8080, mitm.hijackHTTPRequest(func(isHttps, urlStr, req, modified, dropped) {
// // will be called to add an additional request header
// req = poc.ReplaceHTTPPacketHeader(req, "AAA", "BBB")
// modified(req)
// }
// ))
// ```
func mitmConfigHijackHTTPRequest(h func(isHttps bool, u string, req []byte, modified func([]byte), dropped func())) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.hijackRequest = h
	}
}

// hijackHTTPResponse is an option function used to specify the response hijacking function of the middleman proxy server. When receiving After receiving the response, the callback function
// By calling the fourth parameter of the callback function, the response content can be modified, and by calling the fifth parameter of the callback function, the response can be discarded
// Example:
// ```
// mitm.Start(8080, mitm.hijackHTTPResponse(func(isHttps, urlStr, rsp, modified, dropped) {
// // Modify the response body to hijacked
// rsp = poc.ReplaceBody(rsp, b"hijacked", false)
// modified(rsp)
// }
// ))
// ```
func mitmConfigHijackHTTPResponse(h func(isHttps bool, u string, rsp []byte, modified func([]byte), dropped func())) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.hijackResponse = h
	}
}

// will be used. hijackHTTPResponseEx is an option function used to specify the response hijacking function of the middleman proxy server. When the response is received, the callback function
// By calling the fifth parameter of this callback function, the response content can be modified, by calling the sixth parameter of this callback function Parameters, you can discard the response
// The difference between it and hijackHTTPResponse is that it can obtain the original request message
// Example:
// ```
// mitm.Start(8080, mitm.hijackHTTPResponseEx(func(isHttps, urlStr, req, rsp, modified, dropped) {
// // Modify the response body to hijacked
// rsp = poc.ReplaceBody(rsp, b"hijacked", false)
// modified(rsp)
// }
// ))
// ```
func mitmConfigHijackHTTPResponseEx(h func(bool, string, []byte, []byte, func([]byte), func())) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.hijackResponseEx = h
	}
}

// wsforcetext is an option function used to force the specified middle-man proxy. The data frame hijacked by the servers websocket is converted into a text frame, the default is false
// ! Deprecated
// Example:
// ```
// mitm.Start(8080, mitm.wsforcetext(true))
// ```
func mitmConfigWSForceTextFrame(b bool) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.wsForceTextFrame = b
	}
}

// wscallback is an option function used to specify the websocket hijacking function of the middleman proxy server when a websocket request or response is received After that, the callback function
// The first parameter of the callback function is the content of the request or response
// when receiving requests and responses. The second parameter is a Boolean value for Indicates whether the content is a request or a response, true indicates a request, false indicates a response
// Through the return value of this callback function, the content of the request or response can be modified
// Example:
// ```
// mitm.Start(8080, mitm.wscallback(func(data, isRequest) { println(data); return data }))
// ```
func mitmConfigWSCallback(f func([]byte, bool) interface{}) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.wscallback = f
	}
}

// rootCA is an option function used to specify the root certificate and private key of the middleman proxy server
// Example:
// ```
// mitm.Start(8080, mitm.rootCA(cert, key))
// ```
func mitmConfigCertAndKey(cert, key []byte) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.mitmCert = cert
		config.mitmPkey = key
	}
}

// will be called. maxContentLength is an option function used to Specify the maximum request and response content length of the middleman proxy server, the default is 10MB
// Example:
// ```
// mitm.Start(8080, mitm.maxContentLength(100 * 1000 * 1000))
// ```
func mitmMaxContentLength(i int) MitmConfigOpt {
	return func(config *mitmConfig) {
		config.maxContentLength = i
	}
}

// Bridge starts a MITM (man-in-the-middle) proxy server. Its first parameter is the port, the second parameter is the downstream proxy server address, and then it can receive zero to more option functions for Affects the behavior of the middleman proxy server
// Bridge is similar to Start, but slightly different. Bridge can specify the downstream proxy server address, and by default will print to the standard output
// will be called. If the CA certificate and CA certificate are not specified. private key, then the built-in certificate and private key
// Example:
// ```
// mitm.Bridge(8080, "", mitm.host("127.0.0.1"), mitm.callback(func(isHttps, urlStr, req, rsp) { http.dump(req); http.dump(rsp)  })) // starts a man-in-the-middle proxy server and prints requests and responses to the standard output. Whether
// ```
func startBridge(
	port interface{},
	downstreamProxy string,
	opts ...MitmConfigOpt,
) error {
	config := &mitmConfig{
		ctx:                context.Background(),
		host:               "",
		callback:           nil,
		mitmCert:           nil,
		mitmPkey:           nil,
		useDefaultMitmCert: true,
		maxContentLength:   10 * 1000 * 1000,
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.host == "" {
		config.host = "127.0.0.1"
	}

	if config.mitmPkey == nil || config.mitmCert == nil {
		if !config.useDefaultMitmCert {
			return utils.Errorf("empty root CA, please use tls to generate or use mitm.useDefaultCA(true) to allow buildin ca.")
		}
		log.Infof("mitm proxy use the default cert and key")
	}

	if config.isTransparent && downstreamProxy != "" {
		log.Errorf("mitm.Bridge cannot be 'isTransparent'")
	}

	if config.ctx == nil {
		config.ctx = context.Background()
	}

	server, err := crep.NewMITMServer(
		crep.MITM_SetWebsocketHijackMode(true),
		crep.MITM_SetForceTextFrame(config.wsForceTextFrame),
		crep.MITM_SetWebsocketRequestHijackRaw(func(req []byte, r *http.Request, rspIns *http.Response, t int64) []byte {
			var i interface{}
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
				}
			}()

			if config.wscallback != nil {
				i = config.wscallback(req, true)
				req = utils.InterfaceToBytes(i)
			}

			return req
		}),
		crep.MITM_SetWebsocketResponseHijackRaw(func(rsp []byte, r *http.Request, rspIns *http.Response, t int64) []byte {
			var i interface{}
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
				}
			}()

			if config.wscallback != nil {
				i = config.wscallback(rsp, false)
				rsp = utils.InterfaceToBytes(i)
			}

			return rsp
		}),
		crep.MITM_SetHijackedMaxContentLength(config.maxContentLength),
		crep.MITM_SetTransparentHijackMode(config.isTransparent),
		crep.MITM_SetTransparentMirrorHTTP(func(isHttps bool, r *http.Request, rsp *http.Response) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
				}
			}()

			urlIns, err := lowhttp.ExtractURLFromHTTPRequest(r, isHttps)
			if urlIns == nil {
				log.Errorf("parse to url instance failed...")
				return
			}

			if config.callback != nil {
				config.callback(isHttps, urlIns.String(), r, rsp)
				return
			}

			println("RECV request:", urlIns.String())
			println("REQUEST: ")
			raw, err := utils.HttpDumpWithBody(r, false)
			if err != nil {
				println("Parse Request Failed: %s")
			}
			println(string(raw))
			println("RESPONSE: ")
			raw, err = utils.HttpDumpWithBody(rsp, false)
			if err != nil {
				println("Parse Response Failed: %s")
			}
			println(string(raw))
			println("-----------------------------")
		}),
		crep.MITM_SetHTTPResponseMirror(func(isHttps bool, u string, r *http.Request, rsp *http.Response, remoteAddr string) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
				}
			}()

			if config.callback != nil {
				config.callback(isHttps, u, r, rsp)
				return
			}

			urlIns, _ := lowhttp.ExtractURLFromHTTPRequest(r, isHttps)
			if urlIns == nil {
				log.Errorf("parse to url instance failed...")
				return
			}

			println("RECV request:", urlIns.String())
			println("REQUEST: ")
			raw, err := utils.HttpDumpWithBody(r, false)
			if err != nil {
				println("Parse Request Failed: %s")
			}
			println(string(raw))
			println("RESPONSE: ")
			raw, err = utils.HttpDumpWithBody(rsp, false)
			if err != nil {
				println("Parse Response Failed: %s")
			}
			println(string(raw))
			println("-----------------------------")
		}),
		crep.MITM_SetDownstreamProxy(downstreamProxy),
		crep.MITM_SetCaCertAndPrivKey(config.mitmCert, config.mitmPkey),
		crep.MITM_SetHTTPRequestHijackRaw(func(isHttps bool, reqIns *http.Request, req []byte) []byte {
			if config.hijackRequest == nil {
				return req
			}

			if reqIns.Method == "CONNECT" {
				return req
			}

			req = lowhttp.FixHTTPRequest(req)
			urlStrIns, _ := lowhttp.ExtractURLFromHTTPRequestRaw(req, isHttps)
			after := req
			isDropped := utils.NewBool(false)
			config.hijackRequest(isHttps, urlStrIns.String(), req, func(bytes []byte) {
				after = bytes
			}, func() {
				isDropped.Set()
			})
			if isDropped.IsSet() {
				return nil
			}
			return lowhttp.FixHTTPRequest(after)
		}),
		crep.MITM_SetHTTPResponseHijackRaw(func(isHttps bool, req *http.Request, rspInstance *http.Response, rsp []byte, remoteAddr string) []byte {
			if config.hijackResponse == nil && config.hijackResponseEx == nil {
				return rsp
			}

			if req.Method == "CONNECT" {
				return rsp
			}

			fixedResp, _, _ := lowhttp.FixHTTPResponse(rsp)
			if fixedResp == nil {
				fixedResp = rsp
			}
			urlStrIns, err := lowhttp.ExtractURLFromHTTPRequest(req, isHttps)
			if err != nil {
				log.Errorf("extract url from httpRequest failed: %s", err)
			}
			after := fixedResp
			isDropped := utils.NewBool(false)
			if config.hijackResponse != nil {
				config.hijackResponse(isHttps, urlStrIns.String(), fixedResp, func(bytes []byte) {
					after = bytes
				}, func() {
					isDropped.Set()
				})
			}

			if config.hijackResponseEx != nil {
				reqRaw := httpctx.GetRequestBytes(req)
				if reqRaw != nil {
					reqRaw, _ = utils.HttpDumpWithBody(req, true)
					if reqRaw != nil && !httpctx.GetContextBoolInfoFromRequest(req, httpctx.REQUEST_CONTEXT_KEY_RequestIsStrippedGzip) {
						reqRaw = lowhttp.DeletePacketEncoding(reqRaw)
					}
				}
				config.hijackResponseEx(isHttps, urlStrIns.String(), reqRaw, fixedResp, func(bytes []byte) {
					after = bytes
				}, func() {
					isDropped.Set()
				})
			}
			if isDropped.IsSet() {
				return nil
			}

			return after
		}),
	)
	if err != nil {
		return utils.Errorf("create mitm server failed: %s", err)
	}
	err = server.Serve(config.ctx, utils.HostPort(config.host, port))
	if err != nil {
		log.Errorf("server mitm failed: %s", err)
		return err
	}
	return nil
}
