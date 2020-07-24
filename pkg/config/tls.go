// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type tlsConfiger interface {
	GetTLSCertPath() string
	GetTLSKeyPath() string
}

var tlsTemplate = x509.Certificate{ //nolint[gochecknoglobals]
	SerialNumber: big.NewInt(-1),
	Subject: pkix.Name{
		Country:            []string{"CH"},
		Organization:       []string{"Proton Technologies AG"},
		OrganizationalUnit: []string{"ProtonMail"},
		CommonName:         "127.0.0.1",
	},
	KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	BasicConstraintsValid: true,
	IsCA:                  true,
	IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().Add(20 * 365 * 24 * time.Hour),
}

var ErrTLSCertExpireSoon = fmt.Errorf("TLS certificate will expire soon")

// GetTLSConfig tries to load TLS config or generate new one which is then returned.
func GetTLSConfig(cfg tlsConfiger) (tlsConfig *tls.Config, err error) {
	certPath := cfg.GetTLSCertPath()
	keyPath := cfg.GetTLSKeyPath()
	tlsConfig, err = loadTLSConfig(certPath, keyPath)
	if err != nil {
		log.WithError(err).Warn("Cannot load cert, generating a new one")
		tlsConfig, err = generateTLSConfig(certPath, keyPath)
		if err != nil {
			return
		}

		if runtime.GOOS == "darwin" {
			if err := exec.Command( // nolint[gosec]
				"execute-with-privileges",
				"/usr/bin/security",
				"add-trusted-cert",
				"-r", "trustRoot",
				"-p", "ssl",
				"-k", "/Library/Keychains/System.keychain",
				certPath,
			).Run(); err != nil {
				log.WithError(err).Error("Failed to add cert to system keychain")
			}
		}
	}

	tlsConfig.ServerName = "127.0.0.1"
	tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(tlsConfig.Certificates[0].Leaf)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.ClientCAs = caCertPool

	/* This is deprecated:
	 * SA1019: tlsConfig.BuildNameToCertificate is deprecated:
	 * NameToCertificate only allows associating a single certificate with a given name.
	 * Leave that field nil to let the library select the first compatible chain from Certificates.
	 */
	tlsConfig.BuildNameToCertificate() // nolint[staticcheck]

	return tlsConfig, err
}

func loadTLSConfig(certPath, keyPath string) (tlsConfig *tls.Config, err error) {
	c, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return
	}

	c.Leaf, err = x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return
	}

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{c},
	}

	if time.Now().Add(31 * 24 * time.Hour).After(c.Leaf.NotAfter) {
		err = ErrTLSCertExpireSoon
		return
	}
	return
}

// See https://golang.org/src/crypto/tls/generate_cert.go
func generateTLSConfig(certPath, keyPath string) (tlsConfig *tls.Config, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("failed to generate private key: %s", err)
		return
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		err = fmt.Errorf("failed to generate serial number: %s", err)
		return
	}

	tlsTemplate.SerialNumber = serialNumber
	derBytes, err := x509.CreateCertificate(rand.Reader, &tlsTemplate, &tlsTemplate, &priv.PublicKey, priv)
	if err != nil {
		err = fmt.Errorf("failed to create certificate: %s", err)
		return
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return
	}
	defer certOut.Close() //nolint[errcheck]
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer keyOut.Close() //nolint[errcheck]
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return
	}

	return loadTLSConfig(certPath, keyPath)
}
