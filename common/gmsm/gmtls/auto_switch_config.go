package gmtls

// NewBasicAutoSwitchConfig Returns a GMSSL/TLS automatic switching configuration
//
// sm2SigCert: SM2 signature key pair, certificate
// sm2EncCert: SM2 encryption key pair, certificate
// stdCert: RSA/ECC Standard key pairs and certificates
//
// return: the most basic Config object
func NewBasicAutoSwitchConfig(sm2SigCert, sm2EncCert, stdCert *Certificate) (*Config, error) {
	fncGetSignCertKeypair := func(info *ClientHelloInfo) (*Certificate, error) {
		gmFlag := false
		// Check whether the supported protocol contains GMSSL
		for _, v := range info.SupportedVersions {
			if v == VersionGMSSL {
				gmFlag = true
				break
			}
		}

		if gmFlag {
			return sm2SigCert, nil
		} else {
			return stdCert, nil
		}
	}

	fncGetEncCertKeypair := func(info *ClientHelloInfo) (*Certificate, error) {
		return sm2EncCert, nil
	}
	support := NewGMSupport()
	support.EnableMixMode()
	return &Config{
		GMSupport:        support,
		GetCertificate:   fncGetSignCertKeypair,
		GetKECertificate: fncGetEncCertKeypair,
	}, nil
}
