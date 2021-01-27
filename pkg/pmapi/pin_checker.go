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
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

type pinChecker struct {
	trustedPins []string
}

type sentReport struct {
	r tlsReport
	t time.Time
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

type clientConfigProvider interface {
	GetClientConfig() *ClientConfig
}

type tlsReporter struct {
	cm          clientConfigProvider
	p           *pinChecker
	sentReports []sentReport
}

func newTLSReporter(p *pinChecker, cm clientConfigProvider) *tlsReporter {
	return &tlsReporter{
		cm: cm,
		p:  p,
	}
}

// reportCertIssue reports a TLS key mismatch.
func (r *tlsReporter) reportCertIssue(remoteURI, host, port string, connState tls.ConnectionState) {
	var certChain []string

	if len(connState.VerifiedChains) > 0 {
		certChain = marshalCert7468(connState.VerifiedChains[len(connState.VerifiedChains)-1])
	} else {
		certChain = marshalCert7468(connState.PeerCertificates)
	}

	cfg := r.cm.GetClientConfig()

	report := newTLSReport(host, port, connState.ServerName, certChain, r.p.trustedPins, cfg.AppVersion)

	if !r.hasRecentlySentReport(report) {
		r.recordReport(report)
		go report.sendReport(remoteURI, cfg.UserAgent)
	}
}

// hasRecentlySentReport returns whether the report was already sent within the last 24 hours.
func (r *tlsReporter) hasRecentlySentReport(report tlsReport) bool {
	var validReports []sentReport

	for _, r := range r.sentReports {
		if time.Since(r.t) < 24*time.Hour {
			validReports = append(validReports, r)
		}
	}

	r.sentReports = validReports

	for _, r := range r.sentReports {
		if cmp.Equal(report, r.r) {
			return true
		}
	}

	return false
}

// recordReport records the given report and the current time so we can check whether we recently sent this report.
func (r *tlsReporter) recordReport(report tlsReport) {
	r.sentReports = append(r.sentReports, sentReport{r: report, t: time.Now()})
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
