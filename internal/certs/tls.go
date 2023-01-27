// Copyright (c) 2023 Proton AG
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

package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
)

// ErrTLSCertExpiresSoon is returned when the TLS certificate is about to expire.
var ErrTLSCertExpiresSoon = fmt.Errorf("TLS certificate will expire soon")

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

// GenerateCert generates a new TLS certificate and returns it as PEM.
var GenerateCert = func(template *x509.Certificate) ([]byte, []byte, error) { //nolint:gochecknoglobals
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create certificate")
	}

	certPEM := new(bytes.Buffer)

	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, err
	}

	keyPEM := new(bytes.Buffer)

	if err := pem.Encode(keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), keyPEM.Bytes(), nil
}

// GetConfig tries to load TLS config or generate new one which is then returned.
func GetConfig(certPEM, keyPEM []byte) (*tls.Config, error) {
	c, err := tls.X509KeyPair(certPEM, keyPEM)
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
		ClientAuth:   tls.RequestClientCert,
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
	}, nil
}
