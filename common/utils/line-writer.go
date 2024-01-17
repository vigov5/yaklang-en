package utils

import (
	"io/fs"
	"os"
)

type fileLineWriter struct {
	*os.File
	flag      int
	firstLine bool
}

func NewFileLineWriter(file string, flag int, perm fs.FileMode) (*fileLineWriter, error) {
	f, err := os.OpenFile(file, flag, perm)
	if err != nil {
		return nil, err
	}
	// When appending, if the last character is not a newline character, add a newline character
	if flag&os.O_APPEND != 0 {
		state, err := f.Stat()
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 1)
		f.ReadAt(buf, state.Size()-1)
		if buf[0] != '\n' {
			_, err := f.WriteString("\n")
			if err != nil {
				return nil, err
			}
		}
	}

	return &fileLineWriter{File: f, flag: flag, firstLine: true}, nil
}

func (w *fileLineWriter) WriteLine(line []byte) (n int, err error) {
	if w.firstLine {
		w.firstLine = false
	} else {
		n, err = w.File.WriteString("\n")
	}
	n, err = w.File.Write(line)
	if err != nil {
		return
	}
	return
}

func (w *fileLineWriter) WriteLineString(line string) (n int, err error) {
	if w.firstLine {
		w.firstLine = false
	} else {
		n, err = w.File.WriteString("\n")
	}
	n, err = w.File.WriteString(line)
	if err != nil {
		return
	}
	return
}
