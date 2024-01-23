package yaklib

import (
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils/tlsutils"
)

// GenerateRSA1024KeyPair Generates a 1024-bit size RSA public and private key pair, returns PEM format public key and private key with error
// Example:
// ```
// pub, pri, err := tls.GenerateRSA1024KeyPair()
// ```
func generateRSA1024KeyPair() ([]byte, []byte, error) {
	return tlsutils.RSAGenerateKeyPair(1024)
}

// GenerateRSA2048KeyPair Generates a 2048-bit size RSA public and private key pair, returns the PEM format public and private key with an error
// Example:
// ```
// pub, pri, err := tls.GenerateRSA2048KeyPair()
// ```
func generateRSA2048KeyPair() ([]byte, []byte, error) {
	return tlsutils.RSAGenerateKeyPair(2048)
}

// GenerateRSA4096KeyPair Generates a 4096-bit RSA public and private key pair, returns the PEM format public key and private key with an error
// Example:
// ```
// pub, pri, err := tls.GenerateRSA4096KeyPair()
// ```
func generateRSA4096KeyPair() ([]byte, []byte, error) {
	return tlsutils.RSAGenerateKeyPair(4096)
}

// GenerateRootCA generates root certificate and private key based on name, returns PEM format certificate and private key with error
// Example:
// ```
// cert, key, err := tls.GenerateRootCA("yaklang.io")
// ```
func generateRootCA(commonName string) (ca []byte, key []byte, err error) {
	return tlsutils.GenerateSelfSignedCertKeyWithCommonName(commonName, "", nil, nil)
}

var TlsExports = map[string]interface{}{
	"GenerateRSAKeyPair":       tlsutils.RSAGenerateKeyPair,
	"GenerateRSA1024KeyPair":   generateRSA1024KeyPair,
	"GenerateRSA2048KeyPair":   generateRSA2048KeyPair,
	"GenerateRSA4096KeyPair":   generateRSA4096KeyPair,
	"GenerateSM2KeyPair":       tlsutils.SM2GenerateKeyPair,
	"GenerateRootCA":           generateRootCA,
	"SignX509ServerCertAndKey": tlsutils.SignServerCrtNKey,
	"SignX509ClientCertAndKey": tlsutils.SignClientCrtNKey,
	"SignServerCertAndKey":     tlsutils.SignServerCrtNKeyWithoutAuth,
	"SignClientCertAndKey":     tlsutils.SignClientCrtNKeyWithoutAuth,
	"Inspect":                  netx.TLSInspect,
	"EncryptWithPkcs1v15":      tlsutils.PemPkcs1v15Encrypt,
	"DecryptWithPkcs1v15":      tlsutils.PemPkcs1v15Decrypt,
}
