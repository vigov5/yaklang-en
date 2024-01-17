package padding

import (
	"crypto/cipher"
	"io"
)

// P7BlockDecrypt decrypts the ciphertext, and removes PKCS#7 padding
// decrypter: block decryptor
// in: ciphertext input stream
// out: plaintext output stream
func P7BlockDecrypt(decrypter cipher.BlockMode, in io.Reader, out io.Writer) error {
	bufIn := make([]byte, 1024)
	bufOut := make([]byte, 1024)
	p7Out := NewPKCS7PaddingWriter(out, decrypter.BlockSize())
	for {
		n, err := in.Read(bufIn)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		decrypter.CryptBlocks(bufOut, bufIn[:n])
		_, err = p7Out.Write(bufOut[:n])
		if err != nil {
			return err
		}
	}
	return p7Out.Final()
}

// P7BlockEnc fills the original text with PKCS#7 padding mode, and encrypts the output
// encrypter: block encryptor
// in: plain text input stream
// out: cipher text output stream
func P7BlockEnc(encrypter cipher.BlockMode, in io.Reader, out io.Writer) error {
	bufIn := make([]byte, 1024)
	bufOut := make([]byte, 1024)
	p7In := NewPKCS7PaddingReader(in, encrypter.BlockSize())
	for {
		n, err := p7In.Read(bufIn)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		encrypter.CryptBlocks(bufOut, bufIn[:n])
		_, err = out.Write(bufOut[:n])
		if err != nil {
			return err
		}
	}
	return nil
}
