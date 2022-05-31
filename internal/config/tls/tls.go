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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package tls

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
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type TLS struct {
	settingsPath string
}

func New(settingsPath string) *TLS {
	return &TLS{
		settingsPath: settingsPath,
	}
}

// NewTLSTemplate creates a new TLS template certificate with a random serial number.
func NewTLSTemplate() (*x509.Certificate, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate serial number")
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CH"},
			Organization:       []string{"Proton AG"},
			OrganizationalUnit: []string{"Proton Mail"},
			CommonName:         "127.0.0.1",
		},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(20 * 365 * 24 * time.Hour),
	}, nil
}

var ErrTLSCertExpiresSoon = fmt.Errorf("TLS certificate will expire soon")

// getTLSCertPath returns path to certificate; used for TLS servers (IMAP, SMTP).
func (t *TLS) getTLSCertPath() string {
	return filepath.Join(t.settingsPath, "cert.pem")
}

// getTLSKeyPath returns path to private key; used for TLS servers (IMAP, SMTP).
func (t *TLS) getTLSKeyPath() string {
	return filepath.Join(t.settingsPath, "key.pem")
}

// HasCerts returns whether TLS certs have been generated.
func (t *TLS) HasCerts() bool {
	if _, err := os.Stat(t.getTLSCertPath()); err != nil {
		return false
	}

	if _, err := os.Stat(t.getTLSKeyPath()); err != nil {
		return false
	}

	return true
}

// GenerateCerts generates certs from the given template.
func (t *TLS) GenerateCerts(template *x509.Certificate) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	certOut, err := os.Create(t.getTLSCertPath())
	if err != nil {
		return err
	}
	defer certOut.Close() //nolint:errcheck,gosec

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(t.getTLSKeyPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer keyOut.Close() //nolint:errcheck,gosec

	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

// GetConfig tries to load TLS config or generate new one which is then returned.
func (t *TLS) GetConfig() (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(t.getTLSCertPath(), t.getTLSKeyPath())
	if err != nil {
		return nil, errors.Wrap(err, "failed to load keypair")
	}

	c.Leaf, err = x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse certificate")
	}

	if time.Now().Add(31 * 24 * time.Hour).After(c.Leaf.NotAfter) {
		return nil, ErrTLSCertExpiresSoon
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(c.Leaf)

	//nolint:gosec  // We need to support older TLS versions for AppleMail and Outlook
	return &tls.Config{
		Certificates: []tls.Certificate{c},
		ServerName:   "127.0.0.1",
		ClientAuth:   tls.VerifyClientCertIfGiven,
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
	}, nil
}
