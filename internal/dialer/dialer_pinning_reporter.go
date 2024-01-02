// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package dialer

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

type sentReport struct {
	r tlsReport
	t time.Time
}

type TLSReporter struct {
	hostURL     string
	appVersion  string
	userAgent   *useragent.UserAgent
	trustedPins []string
	sentReports []sentReport
}

func NewTLSReporter(hostURL, appVersion string, userAgent *useragent.UserAgent, trustedPins []string) *TLSReporter {
	return &TLSReporter{
		hostURL:     hostURL,
		appVersion:  appVersion,
		userAgent:   userAgent,
		trustedPins: trustedPins,
	}
}

// ReportCertIssue reports a TLS key mismatch.
func (r *TLSReporter) ReportCertIssue(remoteURI, host, port string, connState tls.ConnectionState) {
	var certChain []string

	if len(connState.VerifiedChains) > 0 {
		certChain = marshalCert7468(connState.VerifiedChains[len(connState.VerifiedChains)-1])
	} else {
		certChain = marshalCert7468(connState.PeerCertificates)
	}

	report := newTLSReport(host, port, connState.ServerName, certChain, r.trustedPins, r.appVersion)

	if !r.hasRecentlySentReport(report) {
		r.recordReport(report)

		if err := sendReport(report, r.userAgent.GetUserAgent(), r.appVersion, r.hostURL, remoteURI); err != nil {
			logrus.WithError(err).Error("Failed to send TLS pinning report")
		}
	}
}

// hasRecentlySentReport returns whether the report was already sent within the last 24 hours.
func (r *TLSReporter) hasRecentlySentReport(report tlsReport) bool {
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
func (r *TLSReporter) recordReport(report tlsReport) {
	r.sentReports = append(r.sentReports, sentReport{r: report, t: time.Now()})
}

func marshalCert7468(certs []*x509.Certificate) (pemCerts []string) {
	var buffer bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&buffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			logrus.WithError(err).Error("Failed to encode TLS certificate")
		}
		pemCerts = append(pemCerts, buffer.String())
		buffer.Reset()
	}

	return pemCerts
}
