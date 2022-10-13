// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

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
