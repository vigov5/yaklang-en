package lowhttp

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"github.com/andybalholm/brotli"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"io"
	"io/ioutil"
)

func ContentEncodingDecode(contentEncoding string, bodyRaw []byte) (finalResult []byte, fixed bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("handle content-encoding decode failed! reason: %s", err)
			finalResult = bodyRaw
			fixed = false
		}
	}()

	switch true {
	case utils.IContains(contentEncoding, "gzip"):
		// Assume that chunked has been solved here
		if bytes.HasPrefix(bodyRaw, []byte{0x1f, 0x8b, 0x08}) {
			ungzipedRaw, err := gzip.NewReader(bytes.NewBuffer(bodyRaw[:]))
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				log.Warnf("uncompressed gzip failed: %s", err)
			}
			if ungzipedRaw != nil {
				raw, err := ioutil.ReadAll(ungzipedRaw)
				if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
					log.Errorf("read ungzip reader failed: %s", err)
				}
				if raw != nil {
					return raw, true
				}
			}
		}
		return bodyRaw, false
	case utils.IContains(contentEncoding, "br"):
		raw, err := ioutil.ReadAll(brotli.NewReader(bytes.NewBuffer(bodyRaw)))
		if err != nil {
			log.Errorf("read[brotli] decode failed: %s", err)
			return bodyRaw, false
		}
		return raw, true
	case utils.IContains(contentEncoding, "compress"):
		log.Errorf("Content-Encoding: compress is not supported")
		return bodyRaw, false
	case utils.IContains(contentEncoding, "deflate"):
		rawReader, err := zlib.NewReader(bytes.NewBuffer(bodyRaw))
		if err != nil {
			decodedBody, _ := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(bodyRaw)))
			if decodedBody != nil {
				return decodedBody, true
			}
			return bodyRaw, false
		}
		raw, err := ioutil.ReadAll(rawReader)
		if err != nil {
			return bodyRaw, false
		}
		return raw, true
	case utils.IContains(contentEncoding, "identity"):
		fallthrough
	case utils.IContains(contentEncoding, "*"):
		fallthrough
	default:
		return bodyRaw, false
	}
}
