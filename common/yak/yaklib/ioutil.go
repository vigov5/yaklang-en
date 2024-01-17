package yaklib

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/yaklang/yaklang/common/utils"
)

// exists. ReadAll reads all bytes in the Reader, and returns the read The data and error
// Example:
// ```
// data, err = ioutil.ReadAll(reader)
// ```
func _readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// ReadFile Reads all contents in the specified file and returns read Incoming data and errors
// Example:
// ```
// // . Assume that the file /tmp/test.txt, the content is "hello yak"
// data, err = ioutil.ReadFile("/tmp/test.txt") // data = b"hello yak", err = nil
// ```
func _readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadEvery1s Reads the Reader once per second until EOF is read or the callback function returns false
// Example:
// ```
// r, w, err = io.Pipe() // creates a pipe, returns a reading end and a writing end and an error
// die(err)
// go func{
// for {
// w.WriteString("hello yak\n")
// time.Sleep(1)
// }
// }
// io.ReadEvery1s(context.New(), r, func(data) {
// println(string(data))
// return true
// })
// ```
func _readEvery1s(c context.Context, reader io.Reader, f func([]byte) bool) {
	utils.ReadWithContextTickCallback(c, reader, f, 1*time.Second)
}

// LimitReader returns a Reader that reads bytes from r, but returns EOF after n bytes have been read
// Example:
// ```
// lr = io.LimitReader(reader, 1024)
// ```
func _limitReader(r io.Reader, n int64) io.Reader {
	return io.LimitReader(r, n)
}

// TeeReader returns a Reader that reads bytes from r and writes the read bytes into w.
// This Reader is usually used to save a copy of the data that has been read.
// Example:
// ```
// tr = io.TeeReader(reader, buf)
// io.ReadAll(tr)
// // . Now buf also saves all the data read in the reader.
// ```
func _teeReader(r io.Reader, w io.Writer) io.Reader {
	return io.TeeReader(r, w)
}

// MultiReader Returns a Reader that reads data from multiple Readers
// Example:
// ```
// mr = io.MultiReader(reader1, reader2) // reads mr, that is, reads the data in reader1 and reader2 in order.
// io.ReadAll(mr)
// ```
func _multiReader(readers ...io.Reader) io.Reader {
	return io.MultiReader(readers...)
}

// NopCloser returns a ReadCloser, which reads data from r and implements an empty Close method.
// Example:
// ```
// r = io.NopCloser(reader)
// r.Close() // Do nothing
// ```
func _nopCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}

// Pipe creates a pipe and returns a read Get end and a write end and error
// Example:
// ```
// r, w, err = os.Pipe()
// die(err)
//
//	go func {
//	    w.WriteString("hello yak")
//	    w.Close()
//	}
//
// bytes, err = io.ReadAll(r)
// die(err)
// dump(bytes)
// ```
func _ioPipe() (r *os.File, w *os.File, err error) {
	return os.Pipe()
}

// Copy copies the data in the reader to the writer. , until EOF is read or an error occurs, the number of bytes copied and the error are returned.
// Example:
// ```
// n, err = io.Copy(writer, reader)
// ```
func _copy(writer io.Writer, reader io.Reader) (written int64, err error) {
	return io.Copy(writer, reader)
}

// CopyN Copy the data in the reader to the writer until EOF is read or n bytes are copied, Return the number of bytes copied and error
// Example:
// ```
// n, err = io.CopyN(writer, reader, 1024)
// ```
func _copyN(writer io.Writer, reader io.Reader, n int64) (written int64, err error) {
	return io.CopyN(writer, reader, n)
}

// WriteString writes the string s into the writer, and returns the number of bytes written and the error
// Example:
// ```
// n, err = io.WriteString(writer, "hello yak")
// ```
func _writeString(writer io.Writer, s string) (n int, err error) {
	return io.WriteString(writer, s)
}

// ReadStable steadily reads data from the reader until EOF or timeout is read, and returns the read data
// Example:
// ```
// data = io.ReadStable(reader, 60)
// ```
func _readStable(reader io.Reader, float float64) []byte {
	return utils.StableReader(reader, utils.FloatSecondDuration(float), 10*1024*1024)
}

// . Discard is a writer. , it discards all written data.
var Discard = ioutil.Discard

// EOF is an error, indicating that EOF was read.
var EOF = io.EOF

var IoExports = map[string]interface{}{
	"ReadAll":     _readAll,
	"ReadFile":    _readFile,
	"ReadEvery1s": _readEvery1s,

	// inherits from io
	"LimitReader": _limitReader,
	"TeeReader":   _teeReader,
	"MultiReader": _multiReader,
	"NopCloser":   _nopCloser,
	"Pipe":        _ioPipe,
	"Copy":        _copy,
	"CopyN":       _copyN,
	//"NewSectionReader": io.NewSectionReader,
	//"ReadFull":         io.ReadFull,
	//"ReadAtLeast":      io.ReadAtLeast,
	"WriteString": _writeString,
	"ReadStable":  _readStable,

	"Discard": Discard,
	"EOF":     EOF,
}
