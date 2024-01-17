package crawler

import (
	"net/http"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

type RequestIf interface {
	Url() string
	Request() *http.Request
	ResponseBody() []byte
	Response() (*http.Response, error)
	IsHttps() bool
	ResponseRaw() []byte
	RequestRaw() []byte
}

// Url Returns the URL string of the current request
// Example:
// ```
// req.Url()
// ```
func (r *Req) Url() string {
	if r.url != "" {
		return r.url
	}
	return r.request.URL.String()
}

// Request Returns the original request structure reference of the current request
// Example:
// ```
// req.Request()
// ```
func (r *Req) Request() *http.Request {
	reqIns, err := utils.ReadHTTPRequestFromBytes(r.requestRaw)
	if err != nil {
		log.Errorf("read request failed: %s", err)
	}
	return reqIns
}

// RequestRaw Returns the original request report of the current request Text
// Example:
// ```
// req.RequestRaw()
// ```
func (r *Req) RequestRaw() []byte {
	return r.requestRaw
}

// Response Returns the original response structure reference and error of the current request
// Example:
// ```
// resp, err = req.Response()
// ```
func (r *Req) Response() (*http.Response, error) {
	if r.response == nil {
		if r.err == nil {
			return nil, utils.Errorf("BUG: crawler.Req no response and error")
		}
		return nil, r.err
	}
	return r.response, nil
}

// ResponseRaw Returns the original response message of the current request
// Example:
// ```
// req.ResponseRaw()
// ```
func (r *Req) ResponseRaw() []byte {
	return r.responseRaw
}

// ResponseBody Returns the original response body of the current request
// Example:
// ```
// req.ResponseBody()
// ```
func (r *Req) ResponseBody() []byte {
	return r.responseBody
}

// IsHttps Returns whether the current request is an https request
// Example:
// ```
// req.IsHttps()
// ```
func (r *Req) IsHttps() bool {
	return r.https
}
