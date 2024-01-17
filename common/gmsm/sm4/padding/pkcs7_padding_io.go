package padding

import (
	"bytes"
	"errors"
	"io"
)

// PKCS7PaddingReader Input stream conforming to PKCS#7 padding
type PKCS7PaddingReader struct {
	fIn       io.Reader
	padding   io.Reader
	blockSize int
	readed    int64
	eof       bool
	eop       bool
}

// NewPKCS7PaddingReader Create PKCS7 padding Reader
// in: Input stream
// blockSize: Block size
func NewPKCS7PaddingReader(in io.Reader, blockSize int) *PKCS7PaddingReader {
	return &PKCS7PaddingReader{
		fIn:       in,
		padding:   nil,
		eof:       false,
		eop:       false,
		blockSize: blockSize,
	}
}

func (p *PKCS7PaddingReader) Read(buf []byte) (int, error) {
	/*
		based on the length that has been read - read Get the file
			- The file length is sufficient, directly return
			- Insufficient
		- Read n bytes, remaining Requires m bytes
		- read from padding and append to buff.
			- EOF returns directly, the entire Reader end
	*/
	// All have been read
	if p.eof && p.eop {
		return 0, io.EOF
	}

	var n, off = 0, 0
	var err error
	if !p.eof {
		// Read the file
		n, err = p.fIn.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			// Error return
			return 0, err
		}
		p.readed += int64(n)
		if errors.Is(err, io.EOF) {
			// Mark the end of the file
			p.eof = true
		}
		if n == len(buf) {
			// is long enough to directly return
			return n, nil
		}
		// The file length is insufficient, create Padding
		p.newPadding()
		// . If the length is not enough, ask for
		off = n
	}

	if !p.eop {
		// Read the stream
		var n2 = 0
		n2, err = p.padding.Read(buf[off:])
		n += n2
		if errors.Is(err, io.EOF) {
			p.eop = true
		}
	}
	return n, err
}

// Create new Padding
func (p *PKCS7PaddingReader) newPadding() {
	if p.padding != nil {
		return
	}
	size := p.blockSize - int(p.readed%int64(p.blockSize))
	padding := bytes.Repeat([]byte{byte(size)}, size)
	p.padding = bytes.NewReader(padding)
}

// PKCS7PaddingWriter In line with the input stream removed by PKCS#7, the last packet The filling will be removed based on the filling condition.
type PKCS7PaddingWriter struct {
	cache     *bytes.Buffer // The buffer area
	swap      []byte        // Temporary swap area
	out       io.Writer     // Output position
	blockSize int           // points Block size
}

// NewPKCS7PaddingWriter PKCS#7 Padding Writer Can remove padding
func NewPKCS7PaddingWriter(out io.Writer, blockSize int) *PKCS7PaddingWriter {
	cache := bytes.NewBuffer(make([]byte, 0, 1024))
	swap := make([]byte, 1024)
	return &PKCS7PaddingWriter{out: out, blockSize: blockSize, cache: cache, swap: swap}
}

// Write Keep a padding size of data, and write the rest to the output
func (p *PKCS7PaddingWriter) Write(buff []byte) (n int, err error) {
	// Write cache
	n, err = p.cache.Write(buff)
	if err != nil {
		return 0, err
	}
	if p.cache.Len() > p.blockSize {
		// Read out the part that exceeds one packet length and write it into the actual out
		size := p.cache.Len() - p.blockSize
		_, _ = p.cache.Read(p.swap[:size])
		_, err = p.out.Write(p.swap[:size])
		if err != nil {
			return 0, err
		}
	}
	return n, err

}

// from Padding. Final remove the padding and write the last block
func (p *PKCS7PaddingWriter) Final() error {
	// After Write, cache will only retain one Block length data
	b := p.cache.Bytes()
	length := len(b)
	if length != p.blockSize {
		return errors.New("Illegal PKCS7 padding")
	}
	if length == 0 {
		return nil
	}
	unpadding := int(b[length-1])
	if unpadding > p.blockSize || unpadding == 0 {
		return errors.New("Illegal PKCS7 padding")
	}
	_, err := p.out.Write(b[:(length - unpadding)])
	return err
}
