package tests

import (
	"crypto/x509"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
)

var (
	preCompPGPKey  *crypto.Key
	preCompCertPEM []byte
	preCompKeyPEM  []byte
)

func FastGenerateKey(name, email string, passphrase []byte, keyType string, bits int) (string, error) {
	encKey, err := preCompPGPKey.Lock(passphrase)
	if err != nil {
		return "", err
	}

	return encKey.Armor()
}

func FastGenerateCert(template *x509.Certificate) ([]byte, []byte, error) {
	return preCompCertPEM, preCompKeyPEM, nil
}

func init() {
	key, err := crypto.GenerateKey("name", "email", "rsa", 1024)
	if err != nil {
		panic(err)
	}

	template, err := certs.NewTLSTemplate()
	if err != nil {
		panic(err)
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		panic(err)
	}

	preCompPGPKey = key
	preCompCertPEM = certPEM
	preCompKeyPEM = keyPEM
}
