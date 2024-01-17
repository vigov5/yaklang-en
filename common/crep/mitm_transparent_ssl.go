package crep

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	setHijackRequestLock = new(sync.Mutex)
	fallbackHttpFrame    = []byte(`HTTP/1.1 200 OK
Content-Length: 11

origin_fail

`)
)

func (m *MITMServer) ServeTransparentTLS(ctx context.Context, addr string) error {
	if m.mitmConfig == nil {
		return utils.Errorf("mitm config empty")
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return utils.Errorf("listen tcp://%v failed; %s", addr, err)
	}

	addrVerbose := fmt.Sprintf("tcp://%v", addr)
	go func() {
		log.Infof("start to server transparent mitm server: tcp://%v", addr)
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Errorf("%v accept new conn failed: %s", addrVerbose, err)
				return
			}

			log.Infof("recv tcp conn from %v", conn.RemoteAddr().String())
			_ = conn
			go func() {
				defer conn.Close()

				log.Infof("start check tls/http connection... for %s", conn.RemoteAddr().String())
				err := m.handleHTTPS(ctx, conn, addr)
				if err != nil {
					log.Errorf("handle conn from [%s] failed: %s", conn.RemoteAddr().String(), err)
					return
				}
			}()
		}
	}()

	select {
	case <-ctx.Done():
		return l.Close()
	}
}

func (m *MITMServer) handleHTTPS(ctx context.Context, conn net.Conn, origin string) error {
	pc := utils.NewPeekableNetConn(conn)
	raw, err := pc.Peek(1)
	if err != nil {
		return utils.Errorf("peek [%s] failed: %s", conn.RemoteAddr(), err)
	}

	log.Infof("peek one char[%#v] for %v", raw, conn.RemoteAddr().String())

	var isHttps = utils.NewAtomicBool()
	var httpConn net.Conn
	var sni string
	switch raw[0] {
	case 0x16: // https
		log.Infof("serving https for: %s", conn.RemoteAddr().String())
		tconn := tls.Server(pc, m.mitmConfig.TLS())
		err := tconn.Handshake()
		if err != nil {
			return utils.Errorf("tls handshake failed: %s", err)
		}
		log.Infof("conn: %s handshake finished", conn.RemoteAddr().String())
		httpConn = tconn
		sni = tconn.ConnectionState().ServerName
		isHttps.Set()
	default: // http
		log.Infof("start to serve http for %s", conn.RemoteAddr().String())
		httpConn = pc
		isHttps.UnSet()
	}

	//log.Infof("parse req http finished: %v", spew.Sdump(req))
	if httpConn == nil {
		return nil
	}

	if sni == "" {
		return utils.Errorf("SNI empty...")
	}

	log.Infof("start to handle http request for %s", conn.RemoteAddr().String())
	var readerBuffer bytes.Buffer
	var reqReader = io.TeeReader(httpConn, &readerBuffer)
	firstRequest, err := utils.ReadHTTPRequestFromBufioReader(bufio.NewReader(reqReader))
	if err != nil {
		return utils.Errorf("read request failed: %s for %s", err, conn.RemoteAddr().String())
	}
	log.Infof("read request finished for %v", httpConn.RemoteAddr())

	var fakeUrl string
	if isHttps.IsSet() {
		fakeUrl = fmt.Sprintf("https://%v", firstRequest.Host)
	} else {
		fakeUrl = fmt.Sprintf("http://%v", firstRequest.Host)
	}

	// Set timeout and context. Control
	var timeout time.Duration = 30 * time.Second
	var ctxDDL time.Time
	if ddl, ok := ctx.Deadline(); ok {
		ctxDDL = ddl
		timeout = ddl.Sub(time.Now())
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
	}

	host, port, err := utils.ParseStringToHostPort(fakeUrl)
	if err != nil {
		return utils.Errorf("cannot identify target[%s]: %s", fakeUrl, err)
	}

	originHost := host
	if (!utils.IsIPv4(host)) && (!utils.IsIPv6(utils.FixForParseIP(host))) {
		log.Infof("start to handle dns items for %v", host)
		cachedTarget, ok := m.dnsCache.Load(host)
		if !ok {
			target := netx.LookupFirst(host, netx.WithTimeout(timeout), netx.WithDNSServers(m.DNSServers...))
			if target == "" {
				//httpConn.Write(fallbackHttpFrame)
				return utils.Errorf("cannot query dns host[%s]", host)
			}
			log.Infof("dns query finished for %v: results: [%#v]", host, target)
			host = target
			m.dnsCache.Store(host, target)
		} else {
			_h := cachedTarget.(string)
			log.Infof("dns cache matched: %v -> %v", host, _h)
			host = _h
		}
	}
	target := utils.HostPort(host, port)

	// If it is a loopback, a custom content
	if utils.HostPort(host, port) == origin {
		log.Infof("lookback: %s", origin)
		httpConn.Write(fallbackHttpFrame)
		return nil
	}

	log.Infof("start to connect remote addr: %v", target)
	var dialer = &net.Dialer{
		Timeout:  timeout,
		Deadline: ctxDDL,
	}

	var remoteConn net.Conn
	if !isHttps.IsSet() {
		log.Infof("tcp connect to %s", target)
		remoteConn, err = dialer.Dial("tcp", target)
		if err != nil {
			return utils.Errorf("remote tcp://%v failed to dial: %s", utils.HostPort(host, port), err)
		}
	} else {
		log.Infof("tcp+tls connect to %s", target)
		remoteConn, err = tls.DialWithDialer(dialer, "tcp", target, &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionSSL30, // nolint[:staticcheck]
			MaxVersion:         tls.VersionTLS13,
			ServerName:         originHost,
		})
		if err != nil {
			return utils.Errorf("remote tcp+tls://%v failed: %s", target, err)
		}
	}
	defer remoteConn.Close()

	// The following is the forwarding mode, without hijacking.
	if m.transparentHijackMode == nil || !m.transparentHijackMode.IsSet() {
		// In transparent mode, all callbacks will not take effect.
		_, err = remoteConn.Write(readerBuffer.Bytes())
		if err != nil {
			return utils.Errorf("write first http.Request raw []byte failed: %s", err)
		}
		if firstRequest.Body != nil {
			n, _ := io.Copy(remoteConn, firstRequest.Body)
			log.Errorf("request have body len: %v", n)
		}

		var wg = new(sync.WaitGroup)
		wg.Add(2)
		log.Infof("start to do transparent traffic")
		go func() {
			defer wg.Done()
			defer remoteConn.Close()
			io.Copy(remoteConn, httpConn)
		}()

		go func() {
			defer wg.Done()
			defer remoteConn.Close()
			io.Copy(httpConn, remoteConn)
		}()
		defer log.Infof("finished conn from %s to %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
		wg.Wait()
	} else {
		// How to interact with the network next?
		// transparent mode, the callback will not take effect until hijacking is enabled.

		// hijacks the first request.
		var reqBytes = readerBuffer.Bytes()
		if m.transparentHijackRequest == nil && m.transparentHijackRequestManager == nil {
			// When you do not hijack the request, directly Write, do not wait for all
			log.Infof("write first request for %v", remoteConn.RemoteAddr().String())
			_, err = remoteConn.Write(reqBytes)
			if err != nil {
				return utils.Errorf("write first http.Request raw []byte failed: %s", err)
			}
			if firstRequest.Body != nil {
				n, _ := io.Copy(remoteConn, firstRequest.Body)
				log.Errorf("request have body len: %v", n)
			}
			log.Infof("write first request finished for %v", remoteConn.RemoteAddr().String())
		} else {
			// hijacking scenario handles the first packet
			if firstRequest.Body != nil {
				_, _ = ioutil.ReadAll(firstRequest.Body)
			}

			reqBytes = readerBuffer.Bytes()
			if m.transparentHijackRequest != nil {
				reqBytes = m.transparentHijackRequest(isHttps.IsSet(), reqBytes)
			}

			if m.transparentHijackRequestManager != nil {
				reqBytes = m.transparentHijackRequestManager.Hijacked(isHttps.IsSet(), reqBytes)
			}

			remoteConn.Write(reqBytes)

		}

		var rspRaw bytes.Buffer
		// parsing response
		var responseReader = io.TeeReader(remoteConn, &rspRaw)

		// If the response is not hijacked, the speed will be guaranteed as much as you read and write.
		if m.transparentHijackResponse == nil {
			responseReader = io.TeeReader(responseReader, httpConn)
		}

		// to build a response. This response is very critical. When
		rsp, err := http.ReadResponse(bufio.NewReader(responseReader), firstRequest)
		if err != nil {
			return utils.Errorf("read response for req[%v]->%v failed: %s", firstRequest.URL.String(), remoteConn.RemoteAddr(), err)
		}

		// Parse the Body. This body is from remote -> local
		// If the details are not hijacked, just read it normally. Dont worry too much.
		// is hijacked, reading will not be written directly, and manual httpConn.Write is required. After
		if rsp.Body != nil {
			rspBody, _ := ioutil.ReadAll(rsp.Body)
			if len(rspBody) > 0 {
				log.Infof("rsp body found length: %v", len(rspBody))
			}
		}
		log.Info("first req and rsp recv finished!")

		var rspBytes = rspRaw.Bytes()
		// hijacking response, you need to write httpConn manually, but it must be read before it can be hijacked, so the speed may be affected.
		if m.transparentHijackResponse != nil {
			rspBytes = m.transparentHijackResponse(isHttps.IsSet(), rspBytes)
			_, err = httpConn.Write(rspBytes)
			if err != nil {
				return utils.Errorf("feedback response bytes from [%s] to [%s] failed: %s",
					remoteConn.RemoteAddr().String(), httpConn.RemoteAddr().String(), err,
				)
			}
		}

		if m.transparentOriginMirror != nil {
			go m.transparentOriginMirror(isHttps.IsSet(), readerBuffer.Bytes(), rspRaw.Bytes())
		}

		if m.transparentHijackedMirror != nil {
			go m.transparentHijackedMirror(isHttps.IsSet(), reqBytes, rspBytes)
		}

		if rsp.Close {
			return nil
		}

		for {
			// Read request
			var reqRaw bytes.Buffer

			// Here are some useless characters that do not comply with the HTTP protocol prefix request.
			var buf = make([]byte, 1)
			for {
				_, err := httpConn.Read(buf)
				if err != nil {
					return utils.Errorf("httpConn read failed: %s", err)
				}
				if len(buf) > 0 {
					firstByte := buf[0]
					if ('A' <= firstByte && 'Z' >= firstByte) || (firstByte >= 91 && firstByte <= 122) {
						break
					} else {
						continue
					}
				}
			}

			var reqReader = io.TeeReader( // Read http from local If the .Request comes out with a
				io.MultiReader(bytes.NewReader(buf), httpConn),
				&reqRaw,
			)

			// If you do not hijack, read and forward as much as you want.
			if m.transparentHijackRequest == nil && m.transparentHijackRequestManager == nil {
				reqReader = io.TeeReader(reqReader, remoteConn)
			}
			req, err := utils.ReadHTTPRequestFromBufioReader(bufio.NewReader(reqReader))
			if err != nil {
				return utils.Errorf("read http request from: %s failed: %s", httpConn.RemoteAddr().String(), err)
			}

			// The purpose of this is to read the body buffer. If the request is hijacked, it will be written to remoteConn synchronously.
			// . If there is no hijacking request here, nothing strange will happen. Just read it out and collect the result in reqRaw.
			if req.Body != nil {
				_, _ = ioutil.ReadAll(req.Body)
			}

			// hijacking request, this reqRaw must It contains body (if possible)
			var reqBytes = reqRaw.Bytes()
			switch true {
			case m.transparentHijackRequest != nil:
				reqBytes = m.transparentHijackRequest(isHttps.IsSet(), reqBytes)
				_, err = remoteConn.Write(reqBytes)
				if err != nil {
					return utils.Errorf("write http request from [%v] to [%s] failed: %s", httpConn.RemoteAddr().String(), remoteConn.RemoteAddr().String(), err)
				}
			case m.transparentHijackRequestManager != nil:
				reqBytes = m.transparentHijackRequestManager.Hijacked(isHttps.IsSet(), reqBytes)
				_, err := remoteConn.Write(reqBytes)
				if err != nil {
					return utils.Errorf("write http request from [%v] to [%s] failed: %s", httpConn.RemoteAddr().String(), remoteConn.RemoteAddr().String(), err)
				}
			}

			// Read the response
			var rspRaw bytes.Buffer
			var remoteResponseReader = io.TeeReader(remoteConn, &rspRaw)
			if m.transparentHijackResponse == nil {
				remoteResponseReader = io.TeeReader(remoteResponseReader, httpConn)
			}
			rsp, err := http.ReadResponse(bufio.NewReader(remoteResponseReader), req)
			if err != nil {
				return utils.Errorf("read http response from: %s failed: %s", remoteConn.RemoteAddr().String(), err)
			}

			// to be read. Similar to the above code, this is to read the buffer out of
			if rsp.Body != nil {
				_, _ = ioutil.ReadAll(rsp.Body)
			}

			// mirror traffic will be returned. This traffic has not been hijacked!
			if m.transparentOriginMirror != nil {
				go m.transparentOriginMirror(isHttps.IsSet(), reqRaw.Bytes(), rspRaw.Bytes())
			}

			// Hijack return result
			// The hijacking here does not automatically write to httpConn, so it needs to be written manually. This is a synchronization operation, and the performance bottleneck is here.
			var rspBytes = rspRaw.Bytes()
			if m.transparentHijackResponse != nil {
				rspBytes = m.transparentHijackResponse(isHttps.IsSet(), rspBytes)
				_, err = httpConn.Write(rspBytes)
				if err != nil {
					return utils.Errorf("write http response from [%s] to [%s] failed: %s", remoteConn.RemoteAddr().String(), httpConn.RemoteAddr().String(), err)
				}
			}

			// hijacking Mirror traffic
			if m.transparentHijackedMirror != nil {
				go m.transparentHijackedMirror(isHttps.IsSet(), reqBytes, rspBytes)
			}

			// Current req/rsp is processed, and the response requires closing. Information must be provided before closing. Transmit back
			if rsp.Close {
				return nil
			}
		}
	}

	return nil
}
