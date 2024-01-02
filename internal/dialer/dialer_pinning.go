// Copyright (c) 2024 Proton AG
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

package dialer

import (
	"context"
	"crypto/tls"
	"net"
)

// TrustedAPIPins contains trusted public keys of the protonmail API and proxies.
// NOTE: the proxy pins are the same for all proxy servers, guaranteed by infra team ;).
var TrustedAPIPins = []string{ //nolint:gochecknoglobals
	// api.protonmail.ch
	`pin-sha256="drtmcR2kFkM8qJClsuWgUzxgBkePfRCkRpqUesyDmeE="`, // current
	`pin-sha256="YRGlaY0jyJ4Jw2/4M8FIftwbDIQfh8Sdro96CeEel54="`, // hot backup
	`pin-sha256="AfMENBVvOS8MnISprtvyPsjKlPooqh8nMB/pvCrpJpw="`, // cold backup

	// protonmail.com
	// \todo remove when sure no one is using it.
	`pin-sha256="8joiNBdqaYiQpKskgtkJsqRxF7zN0C0aqfi8DacknnI="`, // current
	`pin-sha256="JMI8yrbc6jB1FYGyyWRLFTmDNgIszrNEMGlgy972e7w="`, // hot backup
	`pin-sha256="Iu44zU84EOCZ9vx/vz67/MRVrxF1IO4i4NIa8ETwiIY="`, // cold backup

	// proton.me
	`pin-sha256="CT56BhOTmj5ZIPgb/xD5mH8rY3BLo/MlhP7oPyJUEDo="`, // current
	`pin-sha256="35Dx28/uzN3LeltkCBQ8RHK0tlNSa2kCpCRGNp34Gxc="`, // hot backup
	`pin-sha256="qYIukVc63DEITct8sFT7ebIq5qsWmuscaIKeJx+5J5A="`, // col backup

	// proxies
	`pin-sha256="EU6TS9MO0L/GsDHvVc9D5fChYLNy5JdGYpJw0ccgetM="`, // main
	`pin-sha256="iKPIHPnDNqdkvOnTClQ8zQAIKG0XavaPkcEo0LBAABA="`, // backup 1
	`pin-sha256="MSlVrBCdL0hKyczvgYVSRNm88RicyY04Q2y5qrBt0xA="`, // backup 2
	`pin-sha256="C2UxW0T1Ckl9s+8cXfjXxlEqwAfPM4HiW2y3UdtBeCw="`, // backup 3
}

// TLSReportURI is the address where TLS reports should be sent.
const TLSReportURI = "https://reports.protonmail.ch/reports/tls"

// PinningTLSDialer wraps a TLSDialer to check fingerprints after connecting and
// to report errors if the fingerprint check fails.
type PinningTLSDialer struct {
	dialer     TLSDialer
	pinChecker PinChecker
	reporter   Reporter
	tlsIssueCh chan struct{}
}

// Reporter is used to report TLS issues.
type Reporter interface {
	ReportCertIssue(reportURI, host, port string, state tls.ConnectionState)
}

// PinChecker is used to check TLS keys of connections.
type PinChecker interface {
	CheckCertificate(conn net.Conn) error
}

// NewPinningTLSDialer constructs a new dialer which only returns TCP connections to servers
// which present known certificates.
// It checks pins using the given pinChecker and reports issues using the given reporter.
func NewPinningTLSDialer(dialer TLSDialer, reporter Reporter, pinChecker PinChecker) *PinningTLSDialer {
	return &PinningTLSDialer{
		dialer:     dialer,
		pinChecker: pinChecker,
		reporter:   reporter,
		tlsIssueCh: make(chan struct{}, 1),
	}
}

// DialTLSContext dials the given network/address, returning an error if the certificates don't match the trusted pins.
func (p *PinningTLSDialer) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := p.dialer.DialTLSContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if err := p.pinChecker.CheckCertificate(conn); err != nil {
		if tlsConn, ok := conn.(*tls.Conn); ok && p.reporter != nil {
			p.reporter.ReportCertIssue(TLSReportURI, host, port, tlsConn.ConnectionState())
		}

		p.tlsIssueCh <- struct{}{}

		return nil, err
	}

	return conn, nil
}

// GetTLSIssueCh returns a channel which notifies when a TLS issue is reported.
func (p *PinningTLSDialer) GetTLSIssueCh() <-chan struct{} {
	return p.tlsIssueCh
}
