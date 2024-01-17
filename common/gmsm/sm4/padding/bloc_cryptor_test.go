package padding

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/yaklang/yaklang/common/gmsm/sm4"
	"testing"
)

func TestP7BlockDecrypt(t *testing.T) {
	src := bytes.Repeat([]byte{7}, 16)

	srcIn := bytes.NewBuffer(src)
	encOut := bytes.NewBuffer(make([]byte, 0, 1024))

	key := make([]byte, 16)
	iv := make([]byte, 16)
	_, _ = rand.Read(key)
	_, _ = rand.Read(iv)
	fmt.Printf("key: %02X\n", key)
	fmt.Printf("iv : %02X\n", iv)
	block, err := sm4.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypter := cipher.NewCBCEncrypter(block, iv)

	err = P7BlockEnc(encrypter, srcIn, encOut)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Original text: %02X\n", src)
	fmt.Printf("Encryption: %02X\n", encOut.Bytes())

	decrypter := cipher.NewCBCDecrypter(block, iv)
	decOut := bytes.NewBuffer(make([]byte, 0, 1024))
	err = P7BlockDecrypt(decrypter, encOut, decOut)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Decryption: %02X\n", decOut.Bytes())
	if !bytes.Equal(src, decOut.Bytes()) {
		t.Fatalf("Actual decryption result: %02X, Expected result: %02X", decOut.Bytes(), src)
	}
}
