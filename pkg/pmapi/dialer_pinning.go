// Copyright (c) 2021 Proton Technologies AG
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

package pmapi

import (
	"crypto/tls"
	"net"

	"github.com/sirupsen/logrus"
)

// PinningTLSDialer wraps a TLSDialer to check fingerprints after connecting and
// to report errors if the fingerprint check fails.
type PinningTLSDialer struct {
	dialer TLSDialer

	// pinChecker is used to check TLS keys of connections.
	pinChecker *pinChecker

	// tlsIssueNotifier is used to notify something when there is a TLS issue.
	tlsIssueNotifier func()

	reporter *tlsReporter

	// A logger for logging messages.
	log logrus.FieldLogger
}

// NewPinningTLSDialer constructs a new dialer which only returns tcp connections to servers
// which present known certificates.
// If enabled, it reports any invalid certificates it finds.
func NewPinningTLSDialer(dialer TLSDialer) *PinningTLSDialer {
	return &PinningTLSDialer{
		dialer:     dialer,
		pinChecker: newPinChecker(TrustedAPIPins),
		log:        logrus.WithField("pkg", "pmapi/tls-pinning"),
	}
}

func (p *PinningTLSDialer) SetTLSIssueNotifier(notifier func()) {
	p.tlsIssueNotifier = notifier
}

func (p *PinningTLSDialer) EnableRemoteTLSIssueReporting(cm *ClientManager) {
	p.reporter = newTLSReporter(p.pinChecker, cm)
}

// DialTLS dials the given network/address, returning an error if the certificates don't match the trusted pins.
func (p *PinningTLSDialer) DialTLS(network, address string) (net.Conn, error) {
	conn, err := p.dialer.DialTLS(network, address)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if err := p.pinChecker.checkCertificate(conn); err != nil {
		if p.tlsIssueNotifier != nil {
			go p.tlsIssueNotifier()
		}

		if tlsConn, ok := conn.(*tls.Conn); ok && p.reporter != nil {
			p.reporter.reportCertIssue(
				TLSReportURI,
				host,
				port,
				tlsConn.ConnectionState(),
			)
		}

		return nil, err
	}

	return conn, nil
}
