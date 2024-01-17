package websvr

import (
	"crypto/tls"
	x "crypto/x509"
	"github.com/yaklang/yaklang/common/gmsm/gmtls"
	"github.com/yaklang/yaklang/common/gmsm/x509"
	"io/ioutil"
)

const (
	rsaCertPath     = "./certs/rsa_sign.cer"
	rsaKeyPath      = "./certs/rsa_sign_key.pem"
	RSACaCertPath   = "./certs/RSA_CA.cer"
	RSAAuthCertPath = "./certs/rsa_auth_cert.cer"
	RSAAuthKeyPath  = "./certs/rsa_auth_key.pem"
	SM2CaCertPath   = "./certs/SM2_CA.cer"
	SM2AuthCertPath = "./certs/sm2_auth_cert.cer"
	SM2AuthKeyPath  = "./certs/sm2_auth_key.pem"
	sm2SignCertPath = "./certs/sm2_sign_cert.cer"
	sm2SignKeyPath  = "./certs/sm2_sign_key.pem"
	sm2EncCertPath  = "./certs/sm2_enc_cert.cer"
	sm2EncKeyPath   = "./certs/sm2_enc_key.pem"
)

// RSA configuration
func loadRsaConfig() (*gmtls.Config, error) {
	cert, err := gmtls.LoadX509KeyPair(rsaCertPath, rsaKeyPath)
	if err != nil {
		return nil, err
	}
	return &gmtls.Config{Certificates: []gmtls.Certificate{cert}}, nil
}

// SM2 configuration
func loadSM2Config() (*gmtls.Config, error) {
	sigCert, err := gmtls.LoadX509KeyPair(sm2SignCertPath, sm2SignKeyPath)
	if err != nil {
		return nil, err
	}
	encCert, err := gmtls.LoadX509KeyPair(sm2EncCertPath, sm2EncKeyPath)
	if err != nil {
		return nil, err
	}
	return &gmtls.Config{
		GMSupport:    &gmtls.GMSupport{},
		Certificates: []gmtls.Certificate{sigCert, encCert},
	}, nil
}

// Switch GMSSL/TSL
func loadAutoSwitchConfig() (*gmtls.Config, error) {
	rsaKeypair, err := gmtls.LoadX509KeyPair(rsaCertPath, rsaKeyPath)
	if err != nil {
		return nil, err
	}
	sigCert, err := gmtls.LoadX509KeyPair(sm2SignCertPath, sm2SignKeyPath)
	if err != nil {
		return nil, err
	}
	encCert, err := gmtls.LoadX509KeyPair(sm2EncCertPath, sm2EncKeyPath)
	if err != nil {
		return nil, err

	}
	return gmtls.NewBasicAutoSwitchConfig(&sigCert, &encCert, &rsaKeypair)
}

// Two-way identity authentication server configuration
func loadServerMutualTLCPAuthConfig() (*gmtls.Config, error) {
	// Signing key pair/Certificate and encryption key pair/certificate
	sigCert, err := gmtls.LoadX509KeyPair(sm2SignCertPath, sm2SignKeyPath)
	if err != nil {
		return nil, err
	}
	encCert, err := gmtls.LoadX509KeyPair(sm2EncCertPath, sm2EncKeyPath)
	if err != nil {
		return nil, err

	}

	// Trusted root certificate
	certPool := x509.NewCertPool()
	cacert, err := ioutil.ReadFile(SM2CaCertPath)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(cacert)

	return &gmtls.Config{
		GMSupport:    gmtls.NewGMSupport(),
		Certificates: []gmtls.Certificate{sigCert, encCert},
		ClientCAs:    certPool,
		ClientAuth:   gmtls.RequireAndVerifyClientCert,
	}, nil
}

// Require client identity authentication
func loadAutoSwitchConfigClientAuth() (*gmtls.Config, error) {
	config, err := loadAutoSwitchConfig()
	if err != nil {
		return nil, err
	}
	// Settings require client certificate request, identifying the need for client identity authentication
	config.ClientAuth = gmtls.RequireAndVerifyClientCert
	return config, nil
}

// Get client server two-way identity authentication configuration
func bothAuthConfig() (*gmtls.Config, error) {
	// Trusted root certificate
	certPool := x509.NewCertPool()
	cacert, err := ioutil.ReadFile(SM2CaCertPath)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(cacert)
	authKeypair, err := gmtls.LoadX509KeyPair(SM2AuthCertPath, SM2AuthKeyPath)
	if err != nil {
		return nil, err
	}
	return &gmtls.Config{
		GMSupport:          &gmtls.GMSupport{},
		RootCAs:            certPool,
		Certificates:       []gmtls.Certificate{authKeypair},
		InsecureSkipVerify: false,
	}, nil

}

// Obtaining one-way identity authentication (only authenticating the server) Configuring
func singleSideAuthConfig() (*gmtls.Config, error) {
	// Trusted root certificate
	certPool := x509.NewCertPool()
	cacert, err := ioutil.ReadFile(SM2CaCertPath)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(cacert)

	return &gmtls.Config{
		GMSupport: &gmtls.GMSupport{},
		RootCAs:   certPool,
	}, nil
}

// Get client server two-way identity authentication configuration
func rsaBothAuthConfig() (*tls.Config, error) {
	// Trusted root certificate
	certPool := x.NewCertPool()
	cacert, err := ioutil.ReadFile(RSACaCertPath)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(cacert)
	authKeypair, err := tls.LoadX509KeyPair(RSAAuthCertPath, RSAAuthKeyPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MaxVersion:         tls.VersionTLS12,
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{authKeypair},
		InsecureSkipVerify: false,
	}, nil

}

// Obtaining one-way identity authentication (only authenticating the server) Configuring
func rsaSingleSideAuthConfig() (*tls.Config, error) {
	// Trusted root certificate
	certPool := x.NewCertPool()
	cacert, err := ioutil.ReadFile(RSACaCertPath)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(cacert)

	return &tls.Config{
		MaxVersion: tls.VersionTLS12,
		RootCAs:    certPool,
	}, nil
}
