package crawler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
)

var (
	ExcludedSuffix = []string{
		".css",
		".jpg", ".jpeg", ".png",
		".mp3", ".mp4",
		".flv", ".aac", ".ogg",
		".svg", "ico", ".gif",
		".doc", "docx", ".pptx",
		".ppt", ".pdf",
	}
	ExcludedMIME = []string{
		"image/*",
		"audio/*", "video/*", "*octet-stream*",
		"application/ogg", "application/pdf", "application/msword",
		"application/x-ppt", "video/avi", "application/x-ico",
		"*zip",
	}
)

type header struct {
	Key   string
	Value string
}

type cookie struct {
	cookie        *http.Cookie
	allowOverride bool
}

func (c *Config) init() {
	c.BasicAuth = false
	c.concurrent = 20
	c.maxRedirectTimes = 5
	c.connectTimeout = 10 * time.Second
	c.userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"
	c.maxDepth = 5
	c.maxBodySize = 10 * 1024 * 1024
	c.maxCountOfLinks = 10000
	c.maxCountOfRequest = 1000
	c.disallowSuffix = ExcludedSuffix
	c.disallowMIMEType = ExcludedMIME
	c.startFromParentPath = true
	c.maxRetryTimes = 3
	c.allowMethod = []string{"GET", "POST"}
	c.allowUrlRegexp = make(map[string]*regexp.Regexp)
	c.forbiddenUrlRegexp = make(map[string]*regexp.Regexp)
	c.forbiddenUrlRegexp["(?i)logout"] = regexp.MustCompile(`(?i)logout`)
	c.forbiddenUrlRegexp["(?i)setup"] = regexp.MustCompile(`(?i)setup`)
	c.forbiddenUrlRegexp["(?i)/exit"] = regexp.MustCompile(`(?i)/exit`)
	c.forbiddenUrlRegexp[`\?[^\s]*C=\w*;O=\w*`] = regexp.MustCompile(`\?\S*C=\w*;O=\w*`)
}

type Config struct {
	// basic authentication is implemented through execReq.
	BasicAuth    bool
	AuthUsername string
	AuthPassword string

	// Transport.
	proxies          []string
	concurrent       int
	maxRedirectTimes int
	connectTimeout   time.Duration // 10s
	// tlsConfig        *tls.Config

	// UA
	userAgent string

	// maximum depth,
	maxDepth int // 5

	// maxBodySize is implemented in preReq, and
	maxBodySize int // 10 * 1024 * 1024

	// Limit
	disallowSuffix []string

	// mime Limit
	disallowMIMEType []string

	// Request maximum limit
	maxCountOfRequest int // 1000

	// maxCountOfLinks limits
	maxCountOfLinks int // 2000

	//
	maxRetryTimes int // 3

	startFromParentPath bool     // true
	allowMethod         []string // GET / POST

	//
	allowDomains    []glob.Glob
	forbiddenDomain []glob.Glob

	allowUrlRegexp     map[string]*regexp.Regexp
	forbiddenUrlRegexp map[string]*regexp.Regexp // (?i(logout))

	headers []*header
	cookie  []*cookie

	onRequest func(req *Req)
	onLogin   func(req *Req)

	extractionRules func(*Req) []interface{}
	// appended links
	extraPathForEveryPath     []string
	extraPathForEveryRootPath []string

	// cache, do not
	_cachedOpts []lowhttp.LowhttpOpt

	// runtime id
	runtimeID string
}

var configMutex = new(sync.Mutex)

func (c *Config) GetLowhttpConfig() []lowhttp.LowhttpOpt {
	configMutex.Lock()
	defer configMutex.Unlock()

	if len(c._cachedOpts) > 0 {
		return c._cachedOpts
	}

	var opts []lowhttp.LowhttpOpt
	opts = append(opts, lowhttp.WithProxy(c.proxies...))
	if c.AuthUsername != "" || c.AuthPassword != "" {
		opts = append(opts, lowhttp.WithUsername(c.AuthUsername), lowhttp.WithPassword(c.AuthPassword))
	}
	if c.maxRedirectTimes > 0 {
		opts = append(opts, lowhttp.WithRedirectTimes(c.maxRedirectTimes))
	}
	if c.connectTimeout > 0 {
		opts = append(opts, lowhttp.WithTimeout(c.connectTimeout))
	}
	if c.maxRetryTimes > 0 {
		opts = append(opts, lowhttp.WithRetryTimes(c.maxRetryTimes))
	}
	c._cachedOpts = opts
	return opts
}

func (c *Config) CheckShouldBeHandledURL(u *url.URL) bool {
	pass := false

	// As long as there is one that passes,
	if len(c.allowDomains) > 0 {
		pass = false
		for _, g := range c.allowDomains {
			if g.Match(u.Hostname()) {
				pass = true
				break
			}
		}
		if !pass {
			return false
		}
	}

	// As long as there is a blacklist that does not pass, it will not pass.
	pass = true
	for _, g := range c.forbiddenDomain {
		if g.Match(u.Hostname()) {
			pass = false
			break
		}
	}
	if !pass {
		return false
	}

	// As long as there is a URL Whitelist
	if len(c.allowUrlRegexp) > 0 {
		pass = false
		for _, g := range c.allowUrlRegexp {
			if g.MatchString(u.String()) {
				pass = true
				break
			}
		}
		if !pass {
			return false
		}
	}

	// As long as there is one that does not pass the blacklist, it cannot pass
	pass = true
	for _, g := range c.forbiddenUrlRegexp {
		if g.MatchString(u.String()) {
			pass = false
			break
		}
	}
	if !pass {
		return false
	}

	if len(c.disallowSuffix) > 0 {
		ext := filepath.Ext(u.Path)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if utils.StringSliceContain(c.disallowSuffix, ext) {
			return false
		}
	}

	return true
}

type ConfigOpt func(c *Config)

// disallowSuffix is an option function used to specify the suffix when crawling. Blacklist
// Example:
// ```
// crawler.Start("https://example.com", crawler.disallowSuffix(".css", ".jpg", ".png")) // will not crawl css, jpg, and png files during crawling.
// ```
func WithDisallowSuffix(d []string) ConfigOpt {
	return func(c *Config) {
		c.disallowSuffix = d
	}
}

func WithDisallowMIMEType(d []string) ConfigOpt {
	return func(c *Config) {
		c.disallowMIMEType = d
	}
}

// basicAuth is an option function used to specify the basic authentication user name and password that should be automatically filled in when crawling.
// Example:
// ```
// crawler.Start("https://example.com", crawler.basicAuth("admin", "admin"))
// ```
func WithBasicAuth(user, pass string) ConfigOpt {
	return func(c *Config) {
		c.BasicAuth = true
		c.AuthUsername = user
		c.AuthPassword = pass
	}
}

// proxy is an option function used to specify the proxy
// Example:
// ```
// crawler.Start("https://example.com", crawler.proxy("http://127.0.0.1:8080"))
// ```
func WithProxy(proxies ...string) ConfigOpt {
	return func(c *Config) {
		c.proxies = proxies
	}
}

// concurrent is an option function, used to specify the number of concurrency when crawling, the default is 20
// Example:
// ```
// crawler.Start("https://example.com", crawler.concurrent(10))
// ```
func WithConcurrent(concurrent int) ConfigOpt {
	return func(c *Config) {
		c.concurrent = concurrent
	}
}

// maxRedirect is an option function, used for Specify the maximum number of redirects when crawling, the default is 5
// Example:
// ```
// crawler.Start("https://example.com", crawler.maxRedirect(10))
// ```
func WithMaxRedirectTimes(maxRedirectTimes int) ConfigOpt {
	return func(c *Config) {
		c.maxRedirectTimes = maxRedirectTimes
	}
}

// domainInclude is an option function, used for Specify the domain name whitelist when crawling.
// domain allows the use of glob syntax, such as*.example.com
// Example:
// ```
// crawler.Start("https://example.com", crawler.domainInclude("*.example.com"))
// ```
func WithDomainWhiteList(domain string) ConfigOpt {
	var pattern string
	if !strings.HasPrefix(domain, "*") {
		pattern = "*" + domain
	}
	p, err := glob.Compile(pattern)
	if err != nil {
		log.Errorf("limit domain[%v] failed: %v", domain, err)
		return func(c *Config) {
			for _, i := range HostToWildcardGlobs(domain) {
				c.allowDomains = append(c.allowDomains, i)
			}
		}
	}

	return func(c *Config) {
		for _, i := range HostToWildcardGlobs(domain) {
			c.allowDomains = append(c.allowDomains, i)
		}
		c.allowDomains = append(c.allowDomains, p)
	}
}

// domainExclude is an option function used to specify the domain name blacklist when crawling.
// domain allows the use of glob syntax, such as*.example.com
// Example:
// ```
// crawler.Start("https://example.com", crawler.domainExclude("*.baidu.com"))
// ```
func WithDomainBlackList(domain string) ConfigOpt {
	var pattern string
	if !strings.HasPrefix(domain, "*") {
		pattern = "*" + domain
	}
	p, err := glob.Compile(pattern)
	if err != nil {
		log.Errorf("limit domain[%v] failed: %v", domain, err)
		return func(c *Config) {}
	}
	return func(c *Config) {
		c.forbiddenDomain = append(c.forbiddenDomain, p)
	}
}

func WithExtraSuffixForEveryPath(path ...string) ConfigOpt {
	return func(c *Config) {
		c.extraPathForEveryPath = path
	}
}

func WithExtraSuffixForEveryRootPath(path ...string) ConfigOpt {
	return func(c *Config) {
		c.extraPathForEveryPath = path
	}
}

// urlRegexpInclude is an option function used to specify the URL regular whitelist when crawling.
// Example:
// ```
// crawler.Start("https://example.com", crawler.urlRegexpInclude(`\.html`))
// ```
func WithUrlRegexpWhiteList(re string) ConfigOpt {
	p, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("limit url regexp[%v] whitelist failed: %v", re, err)
		return func(c *Config) {}
	}
	return func(c *Config) {
		c.allowUrlRegexp[re] = p
	}
}

// urlRegexpExclude is an option function used to specify the URL regular blacklist when crawling
// Example:
// ```
// crawler.Start("https://example.com", crawler.urlRegexpExclude(`\.jpg`))
// ```
func WithUrlRegexpBlackList(re string) ConfigOpt {
	p, err := regexp.Compile(re)
	if err != nil {
		log.Errorf("limit url regexp[%v] blacklist failed: %v", re, err)
		return func(c *Config) {}
	}
	return func(c *Config) {
		c.forbiddenUrlRegexp[re] = p
	}
}

// connectTimeout is an option function used to specify the connection timeout when crawling, the default is 10s
// Example:
// ```
// crawler.Start("https://example.com", crawler.connectTimeout(5))
// ```
func WithConnectTimeout(f float64) ConfigOpt {
	return func(c *Config) {
		c.connectTimeout = time.Duration(int64(f * float64(time.Second)))
	}
}

// . responseTimeout is an option function used to specify the response timeout when crawling. The default is 10s.
// ! Not implemented
// Example:
// ```
// crawler.Start("https://example.com", crawler.responseTimeout(5))
// ```
func WithResponseTimeout(f float64) ConfigOpt {
	return func(c *Config) {
		// dummy, ignore it
	}
}

// userAgent. It is an option function used to specify crawling time. The User-Agent
// Example:
// ```
// crawler.Start("https://example.com", crawler.userAgent("yaklang-crawler"))
// ```
func WithUserAgent(ua string) ConfigOpt {
	return func(c *Config) {
		c.userAgent = ua
	}
}

// will pass maxDepth is an option function used to specify the maximum depth when crawling, the default For 5
// Example:
// ```
// crawler.Start("https://example.com", crawler.maxDepth(10))
// ```
func WithMaxDepth(depth int) ConfigOpt {
	return func(c *Config) {
		c.maxDepth = depth
	}
}

// bodySize is an option function used to specify when crawling. The maximum response body size, the default is 10MB.
// Example:
// ```
// crawler.Start("https://example.com", crawler.bodySize(1024 * 1024))
// ```
func WithBodySize(size int) ConfigOpt {
	return func(c *Config) {
		c.maxBodySize = size
	}
}

// maxRequest is an option function, used to specify the maximum number of requests when crawling, the default is 1000
// Example:
// ```
// crawler.Start("https://example.com", crawler.maxRequest(10000))
// ```
func WithMaxRequestCount(limit int) ConfigOpt {
	return func(c *Config) {
		c.maxCountOfRequest = limit
	}
}

// maxUrls Yes An option function used to specify the maximum number of links when crawling. The default is 10000.
// Example:
// ```
// crawler.Start("https://example.com", crawler.maxUrls(20000))
// ```
func WithMaxUrlCount(limit int) ConfigOpt {
	return func(c *Config) {
		c.maxCountOfLinks = limit
	}
}

// maxRetry is an option function used to specify the maximum number of retries during crawling. The default is 3.
// Example:
// ```
// crawler.Start("https://example.com", crawler.maxRetry(10))
// ```
func WithMaxRetry(limit int) ConfigOpt {
	return func(c *Config) {
		c.maxRetryTimes = limit
	}
}

// in handleReqResult. forbiddenFromParent is an option function used to specify whether to prohibit requests from the root path when crawling. The default is false.
// when crawling. For a starting URL, if it does not start from the root path And it is not prohibited to initiate requests from the root path, then the crawler will start crawling
// Example:
// ```
// crawler.Start("https://example.com/a/b/c", crawler.forbiddenFromParent(false)) // This will start from https://example.com/ from its root path. Start crawling
// ```
func WithForbiddenFromParent(b bool) ConfigOpt {
	return func(c *Config) {
		c.startFromParentPath = !b
	}
}

// header is an option function, used to specify the request header when crawling.
// Example:
// ```
// crawler.Start("https://example.com", crawler.header("User-Agent", "yaklang-crawler"))
// ```
func WithHeader(k, v string) ConfigOpt {
	return func(c *Config) {
		c.headers = append(c.headers, &header{
			Key:   k,
			Value: v,
		})
	}
}

// urlExtractor is an option function, which receives a function as a parameter, used to add additional link extraction rules for crawlers.
// Example:
// ```
// crawler.Start("https://example.com", crawler.urlExtractor(func(req) {
// through preReq. Try to write your own rules to extract additional links from the response body (req.Response() or req.ResponseRaw()). Configuration in
// })
// ```
func WithUrlExtractor(f func(*Req) []interface{}) ConfigOpt {
	return func(c *Config) {
		c.extractionRules = f
	}
}

// cookie is an option function. Cookie used to specify the crawler
// Example:
// ```
// crawler.Start("https://example.com", crawler.cookie("key", "value"))
// ```
func WithFixedCookie(k, v string) ConfigOpt {
	return func(c *Config) {
		c.cookie = append(c.cookie, &cookie{
			cookie: &http.Cookie{
				Name:  k,
				Value: v,
			},
			allowOverride: false,
		})
	}
}

func WithOnRequest(f func(req *Req)) ConfigOpt {
	return func(c *Config) {
		c.onRequest = f
	}
}

// autoLogin is an option function used to specify the automatic number of retries during crawling. Fill in the possible login form
// Example:
// ```
// crawler.Start("https://example.com", crawler.autoLogin("admin", "admin"))
// ```
func WithAutoLogin(username, password string, flags ...string) ConfigOpt {
	return func(c *Config) {
		c.onLogin = func(req *Req) {
			if !utils.IContains(req.Request().Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
				return
			}

			_, body := lowhttp.SplitHTTPHeadersAndBodyFromPacket(req.requestRaw)
			if body == nil {
				log.Errorf("auto login failed... body empty")
				return
			}

			values, err := url.ParseQuery(string(body))
			if err != nil {
				log.Errorf("parse body to query kvs failed: %s", err)
				return
			}

			values.Set(req.maybeLoginUsername, username)
			values.Set(req.maybeLoginPassword, password)

			valueStr := strings.TrimSpace(values.Encode())
			if c.maxRetryTimes <= 0 {
				c.maxRetryTimes = 1
			}
			req.request.GetBody = func() (io.ReadCloser, error) {
				req.request.ContentLength = int64(len(valueStr))
				return ioutil.NopCloser(bytes.NewBufferString(valueStr)), nil
			}
			req.request.Body, _ = req.request.GetBody()

			opts := c.GetLowhttpConfig()
			req.requestRaw, _ = utils.DumpHTTPRequest(req.request, true)
			opts = append(opts, lowhttp.WithPacketBytes(req.requestRaw), lowhttp.WithHttps(req.IsHttps()))

			rspIns, err := lowhttp.HTTP(opts...)
			if err != nil {
				req.err = err
				return
			}
			for _, cookieItem := range lowhttp.ExtractCookieJarFromHTTPResponse(rspIns.RawPacket) {
				c.cookie = append(c.cookie, &cookie{cookie: cookieItem, allowOverride: false})
			}
			req.responseRaw = rspIns.RawPacket
			req.response, _ = utils.ReadHTTPResponseFromBytes(rspIns.RawPacket, req.request)
		}
	}
}

func WithRuntimeID(id string) ConfigOpt {
	return func(c *Config) {
		c.runtimeID = id
	}
}
