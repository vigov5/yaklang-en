package lowhttp

import (
	"bufio"
	"bytes"
	"container/list"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp/httpctx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	H2 = "h2"
	H1 = "http/1.1"
)

var (
	DefaultLowHttpConnPool = &LowHttpConnPool{
		maxIdleConn:        100,
		maxIdleConnPerHost: 2,
		connCount:          0,
		idleConnTimeout:    90 * time.Second,
		idleConn:           make(map[string][]*persistConn),
		keepAliveTimeout:   30 * time.Second,
	}
	errServerClosedIdle = errors.New("conn pool: server closed idle connection")
)

func NewDefaultHttpConnPool() *LowHttpConnPool {
	return &LowHttpConnPool{
		maxIdleConn:        100,
		maxIdleConnPerHost: 2,
		connCount:          0,
		idleConnTimeout:    90 * time.Second,
		idleConn:           make(map[string][]*persistConn),
		keepAliveTimeout:   30 * time.Second,
	}
}

type LowHttpConnPool struct {
	idleConnMux        sync.RWMutex              //Idle connection access lock
	maxIdleConn        int                       //Maximum total connections
	maxIdleConnPerHost int                       //Maximum connection for a single host
	connCount          int                       //Existing connection counter
	idleConn           map[string][]*persistConn //Idle connection
	idleConnTimeout    time.Duration             //Connection expiration time
	idleLRU            connLRU                   //Connection pool LRU
	keepAliveTimeout   time.Duration
}

// Take out an idle connection
// want Retrieve an available connection and take this connection out of the connection pool
func (l *LowHttpConnPool) getIdleConn(key connectKey, opts ...netx.DialXOption) (*persistConn, error) {
	//Try to obtain a reused connection
	if oldPc, ok := l.getFromConn(key); ok {
		return oldPc, nil
	}
	//If there is no reuse connection, create a new connection
	pConn, err := newPersistConn(key, l, opts...)
	if err != nil {
		return nil, err
	}
	return pConn, nil
}

func (l *LowHttpConnPool) getFromConn(key connectKey) (oldPc *persistConn, getConn bool) {
	l.idleConnMux.Lock()
	defer l.idleConnMux.Unlock()
	getConn = false
	var oldTime time.Time
	if l.idleConnTimeout > 0 {
		oldTime = time.Now().Add(-l.idleConnTimeout)
	}

	//Take out one from the connection pool. Connection
	if connList, ok := l.idleConn[key.hash()]; ok {
		if key.scheme == H2 { // h2. No need to remove the connection.
			for len(connList) > 0 {
				oldPc = connList[len(connList)-1]

				//Check whether the obtained connection is available
				canUse := !(oldPc.alt.readGoAway || oldPc.alt.closed || oldPc.alt.full)
				if canUse {
					getConn = true
					break
				}
				connList = connList[:len(connList)-1]
			}
		} else {
			for len(connList) > 0 {
				oldPc = connList[len(connList)-1]
				//Check whether the acquired connection is idle and timeout, if After timeout, remove the next
				tooOld := !oldTime.Before(oldPc.idleAt)
				if !tooOld {
					l.idleLRU.remove(oldPc)
					connList = connList[:len(connList)-1]
					getConn = true
					break
				}
				oldPc.Conn.Close()
				l.idleLRU.remove(oldPc)
				connList = connList[:len(connList)-1]
			}
			if len(connList) > 0 {
				l.idleConn[key.hash()] = connList
			} else {
				delete(l.idleConn, key.hash())
			}
		}
	}
	return
}

func (l *LowHttpConnPool) putIdleConn(conn *persistConn) error {
	cacheKeyHash := conn.cacheKey.hash()
	l.idleConnMux.Lock()
	defer l.idleConnMux.Unlock()
	//If it exceeds the maximum number of connections that a single host can have, it will directly give up adding the connection.
	if len(l.idleConn[cacheKeyHash]) >= l.maxIdleConnPerHost {
		return nil
	}

	//Add a connection to the connection pool, convert the connection status, refresh the idle time
	conn.idleAt = time.Now()
	if l.idleConnTimeout > 0 { //Determine the idle time, if it is 0, there is no limit
		if conn.closeTimer != nil {
			conn.closeTimer.Reset(l.idleConnTimeout)
		} else {
			conn.closeTimer = time.AfterFunc(l.idleConnTimeout, conn.removeConn)
		}
	}

	if l.connCount >= l.maxIdleConn {
		oldPconn := l.idleLRU.removeOldest()
		err := l.removeConnLocked(oldPconn)
		if err != nil {
			return err
		}
	}
	l.idleConn[cacheKeyHash] = append(l.idleConn[cacheKeyHash], conn)
	conn.markReused()
	return nil
}

// Delete an idle connection from the pool in an environment with write lock
func (l *LowHttpConnPool) removeConnLocked(pConn *persistConn) error {
	if pConn.closeTimer != nil {
		pConn.closeTimer.Stop()
	}
	key := pConn.cacheKey.hash()
	connList := l.idleConn[pConn.cacheKey.hash()]
	pConn.Conn.Close()
	switch len(connList) {
	case 0:
		return nil
	case 1:
		if connList[0] == pConn {
			delete(l.idleConn, key)
		}
	default:
		for i, v := range connList {
			if v != pConn {
				continue
			}
			copy(connList[i:], connList[i+1:])
			l.idleConn[key] = connList[:len(connList)-1]
			break
		}
	}
	return nil
}

// Long connection
type persistConn struct {
	alt      *http2ClientConn
	net.Conn //conn body
	mu       sync.Mutex
	p        *LowHttpConnPool //Connect to the corresponding connection pool
	cacheKey connectKey       //Connection pool cache key
	isProxy  bool             //Whether to use a proxy
	alive    bool             //Survival judgment
	sawEOF   bool             //Whether the connection is EOF

	idleAt               time.Time                 //Enter the idle time
	closeTimer           *time.Timer               //Close the timer
	dialOption           []netx.DialXOption        //dial option
	br                   *bufio.Reader             // from conn
	bw                   *bufio.Writer             // to conn
	reqCh                chan requestAndResponseCh //Read pipe
	writeCh              chan writeRequest         //Write Enter the pipeline
	closeCh              chan struct{}             //Close the signal
	writeErrCh           chan error                //write error signal
	serverStartTime      time.Time                 //Response time
	numExpectedResponses int                       //Expected number of responses
	reused               bool                      //Try to obtain the reuse connection
	closed               error                     //Reason for connection closure

	inPool bool
	isIdle bool

	//debug info
	wPacket []packetInfo
	rPacket []packetInfo
}

type requestAndResponseCh struct {
	reqPacket   []byte
	ch          chan responseInfo
	reqInstance *http.Request
	option      *LowhttpExecConfig
	writeErrCh  chan error
	//respCh
}

type responseInfo struct {
	resp      *http.Response
	respBytes []byte
	err       error
	info      httpInfo
}

type httpInfo struct {
	ServerTime time.Duration
}

type writeRequest struct {
	reqPacket   []byte
	ch          chan error
	reqInstance *http.Request
}

type packetInfo struct {
	localPort string
	packet    []byte
}

type persistConnWriter struct {
	pc *persistConn
}

func (w persistConnWriter) Write(p []byte) (n int, err error) {
	n, err = w.pc.Conn.Write(p)
	return
}

func (w persistConnWriter) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = io.Copy(w.pc.Conn, r)
	return
}

type bodyEOFSignal struct {
	body         io.ReadCloser
	mu           sync.Mutex        // guards following 4 fields
	closed       bool              // whether Close has been called
	rerr         error             // sticky Read error
	fn           func(error) error // err will be nil on Read io.EOF
	earlyCloseFn func() error      // optional alt Close func used if io.EOF not seen
}

var errReadOnClosedResBody = errors.New("http: read on closed response body")

func (es *bodyEOFSignal) Read(p []byte) (n int, err error) {
	es.mu.Lock()
	closed, rerr := es.closed, es.rerr
	es.mu.Unlock()
	if closed {
		return 0, errReadOnClosedResBody
	}
	if rerr != nil {
		return 0, rerr
	}

	n, err = es.body.Read(p)
	if err != nil {
		es.mu.Lock()
		defer es.mu.Unlock()
		if es.rerr == nil {
			es.rerr = err
		}
		err = es.condfn(err)
	}
	return
}

func (es *bodyEOFSignal) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	if es.closed {
		return nil
	}
	es.closed = true
	if es.earlyCloseFn != nil && es.rerr != io.EOF {
		return es.earlyCloseFn()
	}
	err := es.body.Close()
	return es.condfn(err)
}

// caller must hold es.mu.
func (es *bodyEOFSignal) condfn(err error) error {
	if es.fn == nil {
		return err
	}
	err = es.fn(err)
	es.fn = nil
	return err
}

func newPersistConn(key connectKey, pool *LowHttpConnPool, opt ...netx.DialXOption) (*persistConn, error) {
	needProxy := len(key.proxy) > 0
	opt = append(opt, netx.DialX_WithKeepAlive(pool.keepAliveTimeout))
	newConn, err := netx.DialX(key.addr, opt...)
	if err != nil {
		return nil, err
	}
	if key.https && key.scheme == H2 {
		if newConn.(*tls.Conn).ConnectionState().NegotiatedProtocol == H1 { //Downgrade
			key.scheme = H1
		}
	}

	// Initialize the connection
	pc := &persistConn{
		Conn:                 newConn,
		mu:                   sync.Mutex{},
		p:                    pool,
		cacheKey:             key,
		isProxy:              needProxy,
		sawEOF:               false,
		idleAt:               time.Time{},
		closeTimer:           nil,
		dialOption:           opt,
		reqCh:                make(chan requestAndResponseCh, 1),
		writeCh:              make(chan writeRequest, 1),
		closeCh:              make(chan struct{}, 1),
		writeErrCh:           make(chan error, 1),
		serverStartTime:      time.Time{},
		wPacket:              make([]packetInfo, 0),
		rPacket:              make([]packetInfo, 0),
		numExpectedResponses: 0,
	}

	if key.scheme == H2 {
		pc.h2Conn()
		go pc.alt.readLoop()
		if err = pc.alt.preface(); err == nil {
			err = pool.putIdleConn(pc)
			if err != nil {
				return nil, err
			}
			return pc, nil
		}
		newH1Conn, err := netx.DialX(key.addr, opt...) // Downgrade
		if err != nil {
			return nil, err
		}
		pc.alt = nil
		pc.Conn = newH1Conn
		pc.cacheKey.scheme = H1
		return pc, nil // After downgrading, the connection pool should not be used because it was an unexpected request. Compatible, no need to reuse
	}

	pc.br = bufio.NewReader(pc)
	pc.bw = bufio.NewWriter(persistConnWriter{pc})

	//Start the read and write cycle
	go pc.writeLoop()
	go pc.readLoop()
	return pc, nil
}

func (pc *persistConn) h2Conn() {
	newH2Conn := &http2ClientConn{
		conn:              pc.Conn,
		mu:                new(sync.Mutex),
		streams:           make(map[uint32]*http2ClientStream),
		currentStreamID:   1,
		idleTimeout:       pc.p.idleConnTimeout,
		maxFrameSize:      defaultMaxFrameSize,
		initialWindowSize: defaultStreamReceiveWindowSize,
		headerListMaxSize: defaultHeaderTableSize,
		connWindowControl: newControl(defaultStreamReceiveWindowSize),
		maxStreamsCount:   defaultMaxConcurrentStreamSize,
		fr:                http2.NewFramer(pc.Conn, bufio.NewReader(pc.Conn)),
		frWriteMutex:      new(sync.Mutex),
		hDec:              hpack.NewDecoder(defaultHeaderTableSize, nil),
		closeCond:         sync.NewCond(new(sync.Mutex)),
		preFaceCond:       sync.NewCond(new(sync.Mutex)),
		http2StreamPool: &sync.Pool{
			New: func() interface{} {
				return new(http2ClientStream)
			},
		},
	}

	newH2Conn.idleTimer = time.AfterFunc(newH2Conn.idleTimeout, func() {
		newH2Conn.closed = true
	})
	pc.alt = newH2Conn
}

func (pc *persistConn) readLoop() {
	defer func() {
		if pc.reused {
			pc.removeConn()
		}
	}()

	tryPutIdleConn := func() bool {
		err := pc.p.putIdleConn(pc)
		if err != nil {
			return false
		}
		return true
	}

	eofc := make(chan struct{})
	defer close(eofc) // unblock reader on errors

	var rc requestAndResponseCh

	count := 0
	alive := true
	firstAuth := true
	for alive {
		// if failed, handle it (re-conn / or abandoned)
		_ = pc.Conn.SetReadDeadline(time.Time{})
		_, err := pc.br.Peek(1)

		// Check whether there is a response that needs to be returned. If not, you can return directly without returning data to the pipe (err).
		if pc.numExpectedResponses == 0 {
			if err == io.EOF {
				pc.closeConn(errServerClosedIdle)
			} else {
				pc.closeConn(err)
			}
			return
		}
		info := httpInfo{ServerTime: time.Since(pc.serverStartTime)}

		if firstAuth {
			rc = <-pc.reqCh
		}

		if err != nil { //A flagged error needs to be returned to the main process, which is used by the main process to determine whether to retry
			if errors.Is(err, io.EOF) {
				pc.sawEOF = true
			}
			rc.ch <- responseInfo{err: connPoolReadFromServerError{err: err}}
			return
		}

		var resp *http.Response

		var stashRequest = rc.reqInstance
		if stashRequest == nil {
			stashRequest = new(http.Request)
		}
		// peek is executed, so we can read without timeout
		// for long time chunked supported
		_ = pc.Conn.SetReadDeadline(time.Time{})
		resp, err = utils.ReadHTTPResponseFromBufioReaderConn(pc.br, pc.Conn, stashRequest)
		if resp != nil {
			resp.Request = nil
		}

		if firstAuth && resp != nil && resp.StatusCode == http.StatusUnauthorized {
			if authHeader := IGetHeader(resp, "WWW-Authenticate"); len(authHeader) > 0 {
				if auth := GetHttpAuth(authHeader[0], rc.option); auth != nil {
					authReq, err := auth.Authenticate(pc.Conn, rc.option)
					if err == nil {
						pc.writeCh <- writeRequest{reqPacket: authReq,
							ch:          rc.writeErrCh,
							reqInstance: rc.reqInstance,
						}
					}
					firstAuth = false
					continue
				}
			}
		}

		count++
		var responseRaw bytes.Buffer
		var respPacket []byte
		var respClose bool
		if resp != nil {
			respClose = resp.Close
			respPacket = httpctx.GetBareResponseBytes(stashRequest)
		}
		if len(respPacket) > 0 {
			responseRaw.Write(respPacket)
		}

		if err != nil || respClose {
			if responseRaw.Len() >= len(respPacket) { // If there is data proof inside TeaReader, It proves that there is response data, but the parsing fails
				// continue read 5 seconds, to receive rest data
				// ignore error, treat as bad conn
				timeout := 5 * time.Second
				if respClose {
					timeout = 1 * time.Second //If http close, only wait for 1 second
				}
				restBytes, _ := utils.ReadUntilStable(pc.br, pc.Conn, timeout, 300*time.Millisecond)
				pc.sawEOF = true // Abandoned connection
				if len(restBytes) > 0 {
					responseRaw.Write(restBytes)
					respPacket = responseRaw.Bytes()
				}
			}
		}

		if len(respPacket) > 0 {
			httpctx.SetBareResponseBytesForce(stashRequest, respPacket) // . Forcibly modify the original response packet.
			err = nil
		}

		pc.mu.Lock()
		pc.numExpectedResponses-- //Reduce the number of expected responses
		pc.mu.Unlock()

		rc.ch <- responseInfo{resp: resp, respBytes: respPacket, info: info, err: err}
		firstAuth = true
		alive = alive &&
			!pc.sawEOF &&
			tryPutIdleConn()
	}
}

func (pc *persistConn) writeLoop() {
	count := 0
	for {
		select {
		case wr := <-pc.writeCh:
			count++
			_, err := pc.bw.Write(wr.reqPacket)
			if err == nil {
				err = pc.bw.Flush()
				pc.serverStartTime = time.Now()
			}
			wr.ch <- err //to exec.go
			if err != nil {
				pc.writeErrCh <- err
				//log.Infof("!!!conn [%v] connect [%v] has be writed [%d]", pc.Conn.LocalAddr(), pc.cacheKey.addr, count)
				return
			}
			pc.mu.Lock()
			pc.numExpectedResponses++
			pc.mu.Unlock()
		case <-pc.closeCh:
			//log.Infof("!!!conn [%v] connect [%v] has be writed [%d]", pc.Conn.LocalAddr(), pc.cacheKey.addr, count)
			return
		}
	}
}

func (pc *persistConn) closeConn(err error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if pc.closed != nil {
		return
	}
	if err == nil {
		err = errors.New("lowhttp: conn pool unknown error")
	}
	pc.Conn.Close()
	pc.closed = err
	close(pc.closeCh)
}

func (pc *persistConn) Close() error {
	return pc.p.putIdleConn(pc)
}

func (pc *persistConn) removeConn() {
	l := pc.p
	l.idleConnMux.Lock()
	defer l.idleConnMux.Unlock()
	err := l.removeConnLocked(pc)
	if err != nil {
		log.Error(err)
	}
}

func (pc *persistConn) Read(b []byte) (n int, err error) {
	n, err = pc.Conn.Read(b)
	if err == io.EOF {
		pc.sawEOF = true
	}
	return
}

// markReused Indicates that this connection has been reused
func (pc *persistConn) markReused() {
	pc.mu.Lock()
	pc.reused = true
	pc.mu.Unlock()
}

func (pc *persistConn) shouldRetryRequest(err error) bool {
	//todo H2 processing
	if !pc.reused {
		//If the initial connection fails, there will be no retry.
		return false
	}
	var connPoolReadFromServerError connPoolReadFromServerError
	if errors.As(err, &connPoolReadFromServerError) {
		//Server error other than EOF, retry
		return true
	}
	//todo Idempotent request
	if errors.Is(err, errServerClosedIdle) {
		// Abandon the connection.
		return true
	}
	return false // Conservatively do not retry
}

type connPoolReadFromServerError struct {
	err error
}

func (e connPoolReadFromServerError) Unwrap() error { return e.err }

func (e connPoolReadFromServerError) Error() string {
	return fmt.Sprintf("lowhttp: conn pool failed to read from server: %v", e.err)
}

type connectKey struct {
	proxy        []string //Usable proxies
	scheme, addr string   //Protocol and target address
	https        bool
	gmTls        bool
}

func (c connectKey) hash() string {
	return utils.CalcSha1(c.proxy, c.scheme, c.addr, c.https, c.gmTls)
}

type connLRU struct {
	ll *list.List // list.Element.Value type is of *persistConn
	m  map[*persistConn]*list.Element
}

// . Add a new connection to the doubly linked list of LRU.
func (cl *connLRU) add(pc *persistConn) {
	if cl.ll == nil {
		cl.ll = list.New()
		cl.m = make(map[*persistConn]*list.Element)
	}
	ele := cl.ll.PushFront(pc)
	if _, ok := cl.m[pc]; ok {
		panic("persistConn was already in LRU")
	}
	cl.m[pc] = ele
}

// Move LRU after using a connection
func (cl *connLRU) use(pc *persistConn) {
	if cl.ll == nil {
		cl.ll = list.New()
		cl.m = make(map[*persistConn]*list.Element)
	}
	ele, ok := cl.m[pc]
	if !ok {
		panic("persistConn is not already in LRU")
	}
	cl.ll.MoveToFront(ele)
}

// Remove the connection that should be deleted from LRU
func (cl *connLRU) removeOldest() *persistConn {
	ele := cl.ll.Back()
	pc := ele.Value.(*persistConn)
	cl.ll.Remove(ele)
	delete(cl.m, pc)
	return pc
}

// Delete an element in an LRU linked list
func (cl *connLRU) remove(pc *persistConn) {
	if ele, ok := cl.m[pc]; ok {
		cl.ll.Remove(ele)
		delete(cl.m, pc)
	}
}

// Get cached Length.
func (cl *connLRU) len() int {
	return len(cl.m)
}
