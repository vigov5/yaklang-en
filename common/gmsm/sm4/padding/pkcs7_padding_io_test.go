package padding

import (
	"bytes"
	"io"
	"testing"
)

// Test P7 fill Reader
func TestPaddingFileReader_Read(t *testing.T) {
	srcIn := bytes.NewBuffer(bytes.Repeat([]byte{'A'}, 16))
	p := NewPKCS7PaddingReader(srcIn, 16)

	tests := []struct {
		name    string
		buf     []byte
		want    int
		wantErr error
	}{
		{"Read file 1B", make([]byte, 1), 1, nil},
		{"Cross-read 15B File 1B", make([]byte, 16), 16, nil},
		{"Fill read 3B", make([]byte, 3), 3, nil},
		{"Exceed padding and read 16B", make([]byte, 16), 12, nil},
		{"End of file 16B", make([]byte, 16), 0, io.EOF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Read(tt.buf)
			if err != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Read() read = %v, but %v is needed", got, tt.want)
			}
		})
	}
}

// Test P7 padding Writer
func TestPKCS7PaddingWriter_Write(t *testing.T) {
	src := []byte{
		0, 1, 2, 3, 4, 5, 6, 7,
	}
	paddedSrc := append(src, bytes.Repeat([]byte{0x08}, 8)...)
	reader := bytes.NewReader(paddedSrc)
	out := bytes.NewBuffer(make([]byte, 0, 64))
	writer := NewPKCS7PaddingWriter(out, 8)

	for {
		buf := make([]byte, 3)
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		if err != nil {
			t.Fatal(err)
		}
	}
	err := writer.Final()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out.Bytes(), src) {
		t.Fatalf("After removing padding, the actual value is %02X, expect to remove padding The result is %02X", out.Bytes(), src)
	}
}
