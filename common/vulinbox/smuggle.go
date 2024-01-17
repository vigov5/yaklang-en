package vulinbox

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"io"
	"net"
	"net/http"
	"strings"
)

func (s *VulinServer) registerPipelineNSmuggle() {
	smugglePort := utils.GetRandomAvailableTCPPort()
	pipelinePort := utils.GetRandomAvailableTCPPort()

	pipelineNSmuggleSubroute := s.router.PathPrefix("/http/protocol").Name("HTTP CDN and Pipeline Security").Subrouter()
	go func() {
		err := Smuggle(context.Background(), smugglePort)
		if err != nil && err != io.EOF {
			log.Error(err)
		}
	}()
	err := utils.WaitConnect(utils.HostPort("127.0.0.1", smugglePort), 3)
	if err != nil {
		log.Error(err)
		return
	}
	addRouteWithVulInfo(pipelineNSmuggleSubroute, &VulInfo{
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Location", "http://"+utils.HostPort("127.0.0.1", smugglePort))
			writer.WriteHeader(302)
		},
		Path:  `/smuggle`,
		Title: "HTTP request smuggling case: HTTP Smuggle",
	})

	go func() {
		err := Pipeline(context.Background(), pipelinePort)
		if err != nil && err != io.EOF {
			log.Error(err)
		}
	}()
	err = utils.WaitConnect(utils.HostPort("127.0.0.1", pipelinePort), 3)
	if err != nil && err != io.EOF {
		log.Error(err)
		return
	}
	addRouteWithVulInfo(pipelineNSmuggleSubroute, &VulInfo{
		Path: `/pipeline`,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Location", "http://"+utils.HostPort("127.0.0.1", pipelinePort))
			writer.WriteHeader(302)
		},
		Title: "HTTP Pipeline normal case (control group, not a vulnerability)",
	})
}

func Pipeline(ctx context.Context, port int) error {
	if port <= 0 {
		port = utils.GetRandomAvailableTCPPort()
	}
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel
	log.Infof("start to listen Pipeline server proxy: %v", port)
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			cancel()
			lis.Close()
		}
	}()
	defer lis.Close()

	ordinaryRequest := lowhttp.FixHTTPRequest([]byte(`HTTP/1.1 200 OK
Server: ReverseProxy for restriction admin in VULINBOX!
Content-Type: text/html; charset=utf-8
`))
	ordinaryRequest = lowhttp.ReplaceHTTPPacketBody(
		ordinaryRequest, UnsafeRender(
			"HTTP/1.1 Pipeline", []byte(`
in HTTP/1.1, the default Connection: keep-alive is set, <br>
In this case, if the client sends multiple requests, the server will serialize these requests.<br>
that reads the Body according to Transfer-Encoding: chunked That is to say , the next request will be made only when the previous request is completed. The<br>
<br>
Generally speaking, this does not cause security problems
`)),
		false,
	)

	handleRequest := func(reader *bufio.Reader, writer *bufio.Writer) {
		for {
			req, err := utils.ReadHTTPRequestFromBufioReader(reader)
			if err != nil {
				if err != io.EOF {
					log.Error(err)
				}
				return
			}
			log.Infof("method: %v, url: %v", req.Method, req.URL)
			raw, _ := io.ReadAll(req.Body)
			if len(raw) > 0 {
				spew.Dump(raw)
			}
			writer.Write(ordinaryRequest)
			writer.Flush()
			if req.Close {
				log.Info("close connection")
				return
			}
		}
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Error(err)
			break
		}
		br := bufio.NewReader(conn)
		bw := bufio.NewWriter(conn)
		go func() {
			handleRequest(br, bw)
			conn.Close()
		}()
	}
	return nil
}

func Smuggle(ctx context.Context, port int) error {
	if port <= 0 {
		port = utils.GetRandomAvailableTCPPort()
	}
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel
	log.Infof("start to listen Pipeline server proxy: %v", port)
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			cancel()
			lis.Close()
		}
	}()
	defer lis.Close()

	ordinaryRequest := lowhttp.FixHTTPRequest([]byte(`HTTP/1.1 200 OK
Server: ReverseProxy for restriction admin in VULINBOX!
Content-Type: text/html; charset=utf-8
`))
	ordinaryRequest = lowhttp.ReplaceHTTPPacketBody(
		ordinaryRequest, UnsafeRender(
			"HTTP/1.1 Smuggle", []byte(`
In the business scenario where the proxy server reversely proxies the real server,<br>
if The proxy server does not correctly handle the priority of Transfer-Encoding: chunked and Content-Length: ... <br>
Then the hacker may construct a special request that has inconsistent parsing results in the proxy server and the real server. ,<br>
that reads the Body according to Content-Length, thus causing security problems.<br>
<br>
For example, in the server of this port: the proxy server will think that this is an HTTP/1.1 request,<br>
proxy server will read the data in the Body and splice it and pass it to the real server.<br>
And the real server will think that this is an HTTP/1.1s request, the request is divided incorrectly, and two requests are read at the same time, causing security issues. In<br>
<br>
`)),
		false,
	)

	onlyHandleContentLength := func(reader *bufio.Reader, writer *bufio.Writer) ([]byte, error) {
		var proxied bytes.Buffer
		fistLine, err := utils.BufioReadLine(reader)
		if err != nil {
			return nil, err
		}
		proxied.WriteString(string(fistLine) + "\r\n")
		log.Infof("first line: %v", string(fistLine))
		var bodySize int
		for {
			line, err := utils.BufioReadLine(reader)
			if err != nil {
				log.Error(err)
				break
			}
			if string(line) == "" {
				break
			}
			k, v := lowhttp.SplitHTTPHeader(string(line))
			if strings.ToLower(k) == "content-length" {
				bodySize = codec.Atoi(v)
				log.Infof("content-length: %v", bodySize)
			} else {
				proxied.WriteString(string(line) + "\r\n")
			}
		}
		proxied.WriteString("\r\n")
		if bodySize > 0 {
			body := make([]byte, bodySize)
			_, _ = io.ReadFull(reader, body)
			proxied.Write(body)
		}
		return proxied.Bytes(), nil
	}
	handleRequest := func(ret []byte, writer *bufio.Writer) {
		for {
			chunked := false
			_, body := lowhttp.SplitHTTPPacket(ret, func(method string, requestUri string, proto string) error {
				return nil
			}, nil, func(line string) string {
				k, v := lowhttp.SplitHTTPHeader(line)
				if strings.ToLower(k) == "transfer-encoding" && utils.IContains(v, "chunked") {
					chunked = true
				}
				return line
			})
			_, pth, _ := lowhttp.GetHTTPPacketFirstLine(ret)
			switch pth {
			case "/":
				writer.Write(ordinaryRequest)
			default:
				writer.WriteString("HTTP/1.1 404 Not Found\r\nContent-Length: 9\r\n\r\nnot found")
			}
			writer.Flush()
			if chunked {
				var before []byte
				_ = before
				var ok bool
				before, ret, ok = bytes.Cut(body, []byte("0\r\n\r\n"))
				if ok {
					raw, err := codec.HTTPChunkedDecode(append(before, []byte("0\r\n\r\n")...))
					if err != nil {
						log.Error(err)
						return
					}
					spew.Dump(raw)
					continue
				}
			}
			return
		}
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Error(err)
			break
		}
		br := bufio.NewReader(conn)
		bw := bufio.NewWriter(conn)
		go func() {
			defer conn.Close()
			proxies, err := onlyHandleContentLength(br, bw)
			if err != nil {
				log.Error(err)
				return
			}
			fmt.Println(string(proxies))
			spew.Dump(proxies)
			handleRequest(proxies, bw)
			utils.FlushWriter(bw)
			println("=====================================")
		}()
	}
	return nil
}
