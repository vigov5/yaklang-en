package netx

import (
	"context"
	"crypto/tls"
	"github.com/ReneKroon/ttlcache"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"strings"
	"time"
)

var isTlsCached = ttlcache.NewCache()

func IsTLSService(addr string, proxies ...string) bool {
	result, ok := isTlsCached.Get(addr)
	if ok {
		return result.(bool)
	}

	isHttps := false
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := DialTCPTimeout(5*time.Second, addr, proxies...)
	if err == nil {
		defer conn.Close()
		host, _, _ := utils.ParseStringToHostPort(addr)
		loopBack := utils.IsLoopback(host)
		tlsConn := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionSSL30,
			MaxVersion:         tls.VersionTLS13,
			ServerName:         host,
		})

		err = tlsConn.HandshakeContext(ctx)
		if err == nil {
			isHttps = true // The handshake is successful, set isHttps to true
			//// Get the connection status
			//state := tlsConn.ConnectionState()
			//// Print the cipher suite used
			//log.Infof("Cipher Suite: %s\n", tls.CipherSuiteName(state.CipherSuite))
		} else {
			log.Infof("TLS handshake failed: %v", err)
			// Check whether the error message contains a specific TLS error
			if strings.Contains(err.Error(), "handshake failure") || strings.Contains(err.Error(), "protocol version not supported") || strings.HasSuffix(err.Error(), "unsupported elliptic curve") {
				isHttps = true
			}
		}

		// Set the cache according to the value of isHttps
		if !loopBack {
			isTlsCached.SetWithTTL(addr, isHttps, 30*time.Second)
		}
	}

	return isHttps
}
