// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package pmapi

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
)

// ErrTLSMismatch indicates that no TLS fingerprint match could be found.
var ErrTLSMismatch = errors.New("no TLS fingerprint match found")

type pinChecker struct {
	trustedPins []string
}

func newPinChecker(trustedPins []string) *pinChecker {
	return &pinChecker{
		trustedPins: trustedPins,
	}
}

// checkCertificate returns whether the connection presents a known TLS certificate.
func (p *pinChecker) checkCertificate(conn net.Conn) error {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return errors.New("connection is not a TLS connection")
	}

	connState := tlsConn.ConnectionState()

	for _, peerCert := range connState.PeerCertificates {
		fingerprint := certFingerprint(peerCert)

		for _, pin := range p.trustedPins {
			if pin == fingerprint {
				return nil
			}
		}
	}

	return ErrTLSMismatch
}

func certFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return fmt.Sprintf(`pin-sha256=%q`, base64.StdEncoding.EncodeToString(hash[:]))
}
