package yaklib

import (
	"bufio"
	"bytes"
	"io"
	"reflect"

	"github.com/yaklang/yaklang/common/utils"
)

// NewBuffer creates a new Buffer structure reference, which helps us process strings
// Buffer also implements the Reader and Writer interfaces.
// Commonly used The Buffer methods are: Bytes, String, Read, Write, WriteString, WriteByte, Reset
// Example:
// ```
// buffer = bufio.NewBuffer() // or you can also use io.NewBuffer("hello yak") to initialize a Buffer
// buffer.WriteString("hello yak")
// data, err = io.ReadAll(buffer) // data = b"hello yak", err = nil
// ```
func _newBuffer(b ...[]byte) *bytes.Buffer {
	buffer := &bytes.Buffer{}
	if len(b) > 0 {
		buffer.Write(b[0])
	}
	return buffer
}

// NewReader creates a new BufioReader according to the incoming Reader. The structure reference
// based on the incoming Reader and Writer. Commonly used BufioReader methods are: Read, ReadByte , ReadBytes, ReadLine, ReadString, Reset
// Example:
// ```
// reader = bufio.NewReader(os.Stdin)
// ```
func _newReader(i interface{}) (*bufio.Reader, error) {
	if rd, ok := i.(io.Reader); ok {
		return bufio.NewReader(rd), nil
	} else {
		return nil, utils.Errorf("not support type: %v", reflect.TypeOf(i))
	}
}

// NewReaderSize Based on the passed-in Writer Reader creates a new BufioReader structure reference with a cache size of size
// based on the incoming Reader and Writer. Commonly used BufioReader methods are: Read, ReadByte , ReadBytes, ReadLine, ReadString, Reset
// Example:
// ```
// reader = bufio.NewReaderSize(os.Stdin, 1024)
// ```
func _newReaderSize(i interface{}, size int) (*bufio.Reader, error) {
	if rd, ok := i.(io.Reader); ok {
		return bufio.NewReaderSize(rd, size), nil
	} else {
		return nil, utils.Errorf("not support type: %v", reflect.TypeOf(i))
	}
}

// NewWriter Creates a new BufioWriter structure reference based on the passed-in Writer
// The commonly used BufioWriter methods are: Write, WriteByte, WriteString, Reset, Flush
// Example:
// ```
// writer, err = bufio.NewWriter(os.Stdout)
// writer.WriteString("hello yak")
// writer.Flush()
// ```
func _newWriter(i interface{}) (*bufio.Writer, error) {
	if wd, ok := i.(io.Writer); ok {
		return bufio.NewWriter(wd), nil
	} else {
		return nil, utils.Errorf("not support type: %v", reflect.TypeOf(i))
	}
}

// NewWriterSize Creates a new BufioWriter structure reference based on the passed-in Writer, and its cache size is size
// The commonly used BufioWriter methods are: Write, WriteByte, WriteString, Reset, Flush
// Example:
// ```
// writer, err = bufio.NewWriterSize(os.Stdout, 1024)
// writer.WriteString("hello yak")
// writer.Flush()
// ```
func _newWriterSize(i interface{}, size int) (*bufio.Writer, error) {
	if wd, ok := i.(io.Writer); ok {
		return bufio.NewWriterSize(wd, size), nil
	} else {
		return nil, utils.Errorf("not support type: %v", reflect.TypeOf(i))
	}
}

// NewReadWriter creates a new BufioReadWriter structure reference
// BufioReadWriter can call the methods of BufioReader and BufioWriter at the same time
// Example:
// ```
// rw, err = bufio.NewReadWriter(os.Stdin, os.Stdout)
// ```
func _newReadWriter(i, i2 interface{}) (*bufio.ReadWriter, error) {
	var (
		rd  *bufio.Reader
		wd  *bufio.Writer
		err error
	)

	rd, err = _newReader(i)
	if err != nil {
		return nil, err
	}
	wd, err = _newWriter(i2)
	if err != nil {
		return nil, err
	}

	return bufio.NewReadWriter(rd, wd), nil
}

// NewScanner creates a new Scanner structure reference based on the incoming Reader.
// Commonly used Scanner methods are: Scan, Text, Err, Split, SplitFunc
// Example:
// ```
// buf = bufio.NewBuffer("hello yak\nhello yakit")
// scanner, err = bufio.NewScanner(buf)
// for scanner.Scan() {
// println(scanner.Text())
// }
// ```
func _newScanner(i interface{}) (*bufio.Scanner, error) {
	if rd, ok := i.(io.Reader); ok {
		return bufio.NewScanner(rd), nil
	} else {
		return nil, utils.Errorf("not support type: %v", reflect.TypeOf(i))
	}
}

var BufioExport = map[string]interface{}{
	"NewBuffer":     _newBuffer,
	"NewReader":     _newReader,
	"NewReaderSize": _newReaderSize,
	"NewWriter":     _newWriter,
	"NewWriterSize": _newWriterSize,
	"NewReadWriter": _newReadWriter,
	"NewScanner":    _newScanner,
}
