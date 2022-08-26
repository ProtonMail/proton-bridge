package tests

import (
	"crypto/x509"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"gitlab.protontech.ch/go/liteapi/server/account"
)

func init() {
	key, err := crypto.GenerateKey("name", "email", "rsa", 1024)
	if err != nil {
		panic(err)
	}

	account.GenerateKey = func(name, email string, passphrase []byte, keyType string, bits int) (string, error) {
		encKey, err := key.Lock(passphrase)
		if err != nil {
			return "", err
		}

		return encKey.Armor()
	}

	template, err := certs.NewTLSTemplate()
	if err != nil {
		panic(err)
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		panic(err)
	}

	certs.GenerateCert = func(template *x509.Certificate) ([]byte, []byte, error) {
		return certPEM, keyPEM, nil
	}
}
