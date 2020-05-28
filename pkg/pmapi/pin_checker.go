// Copyright (c) 2020 Proton Technologies AG
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
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

type pinChecker struct {
	trustedPins []string
	sentReports []sentReport
}

type sentReport struct {
	r tlsReport
	t time.Time
}

func newPinChecker(trustedPins []string) pinChecker {
	return pinChecker{
		trustedPins: trustedPins,
	}
}

// checkCertificate returns whether the connection presents a known TLS certificate.
func (p *pinChecker) checkCertificate(conn net.Conn) error {
	connState := conn.(*tls.Conn).ConnectionState()

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

// reportCertIssue reports a TLS key mismatch.
func (p *pinChecker) reportCertIssue(remoteURI, host, port string, connState tls.ConnectionState, appVersion, userAgent string) {
	var certChain []string

	if len(connState.VerifiedChains) > 0 {
		certChain = marshalCert7468(connState.VerifiedChains[len(connState.VerifiedChains)-1])
	} else {
		certChain = marshalCert7468(connState.PeerCertificates)
	}

	r := newTLSReport(host, port, connState.ServerName, certChain, p.trustedPins, appVersion)

	if !p.hasRecentlySentReport(r) {
		p.recordReport(r)
		go r.sendReport(remoteURI, userAgent)
	}
}

// hasRecentlySentReport returns whether the report was already sent within the last 24 hours.
func (p *pinChecker) hasRecentlySentReport(report tlsReport) bool {
	var validReports []sentReport

	for _, r := range p.sentReports {
		if time.Since(r.t) < 24*time.Hour {
			validReports = append(validReports, r)
		}
	}

	p.sentReports = validReports

	for _, r := range p.sentReports {
		if cmp.Equal(report, r.r) {
			return true
		}
	}

	return false
}

// recordReport records the given report and the current time so we can check whether we recently sent this report.
func (p *pinChecker) recordReport(r tlsReport) {
	p.sentReports = append(p.sentReports, sentReport{r: r, t: time.Now()})
}

func marshalCert7468(certs []*x509.Certificate) (pemCerts []string) {
	var buffer bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&buffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			logrus.WithField("pkg", "pmapi/tls-pinning").WithError(err).Error("Failed to encode TLS certificate")
		}
		pemCerts = append(pemCerts, buffer.String())
		buffer.Reset()
	}

	return pemCerts
}
