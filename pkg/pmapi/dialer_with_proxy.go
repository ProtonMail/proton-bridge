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

package pmapi

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// TLSReport is inspired by https://tools.ietf.org/html/rfc7469#section-3.
type TLSReport struct {
	//  DateTime of observed pin validation in time.RFC3339 format.
	DateTime string `json:"date-time"`

	// Hostname to which the UA made original request that failed pin validation.
	Hostname string `json:"hostname"`

	// Port to which the UA made original request that failed pin validation.
	Port int `json:"port"`

	// EffectiveExpirationDate for noted pins in time.RFC3339 format.
	EffectiveExpirationDate string `json:"effective-expiration-date"`

	// IncludeSubdomains indicates whether or not the UA has noted the
	// includeSubDomains directive for the Known Pinned Host.
	IncludeSubdomains bool `json:"include-subdomains"`

	// NotedHostname indicates the hostname that the UA noted when it noted
	// the Known Pinned Host. This field allows operators to understand why
	// Pin Validation was performed for, e.g., foo.example.com when the
	// noted Known Pinned Host was example.com with includeSubDomains set.
	NotedHostname string `json:"noted-hostname"`

	// ServedCertificateChain is the certificate chain, as served by
	// the Known Pinned Host during TLS session setup.  It is provided as an
	// array of strings; each string pem1, ... pemN is the Privacy-Enhanced
	// Mail (PEM) representation of each X.509 certificate as described in
	// [RFC7468].
	ServedCertificateChain []string `json:"served-certificate-chain"`

	// ValidatedCertificateChain is the certificate chain, as
	// constructed by the UA during certificate chain verification.  (This
	// may differ from the served-certificate-chain.)  It is provided as an
	// array of strings; each string pem1, ... pemN is the PEM
	// representation of each X.509 certificate as described in [RFC7468].
	// UAs that build certificate chains in more than one way during the
	// validation process SHOULD send the last chain built.  In this way,
	// they can avoid keeping too much state during the validation process.
	ValidatedCertificateChain []string `json:"validated-certificate-chain"`

	// The known-pins are the Pins that the UA has noted for the Known
	// Pinned Host.  They are provided as an array of strings with the
	// syntax: known-pin = token "=" quoted-string
	// e.g.:
	// ```
	// "known-pins": [
	//   'pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="',
	//   "pin-sha256=\"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=\""
	// ]
	// ```
	KnownPins []string `json:"known-pins"`

	// AppVersion is used to set `x-pm-appversion` json format from datatheorem/TrustKit.
	AppVersion string `json:"app-version"`
}

// ErrTLSMatch indicates that no TLS fingerprint match could be found.
var ErrTLSMatch = fmt.Errorf("TLS fingerprint match not found")

// DialerWithPinning will provide dial function which checks the fingerprints of public cert
// received from contacted server. If no match found among know pinse it will report using
// ReportCertIssueLocal.
type DialerWithPinning struct {
	// isReported will stop reporting if true.
	isReported bool

	// report stores known pins.
	report TLSReport

	// When reportURI is not empty the tls issue report will be send to this URI.
	reportURI string

	// ReportCertIssueLocal is used send signal to application about certificate issue.
	// It is used only if set.
	ReportCertIssueLocal func()

	// proxyManager manages API proxies.
	proxyManager *proxyManager

	// A logger for logging messages.
	log logrus.FieldLogger
}

func NewDialerWithPinning(reportURI string, report TLSReport) *DialerWithPinning {
	log := logrus.WithField("pkg", "pmapi/tls-pinning")

	proxyManager := newProxyManager(dohProviders, proxyQuery)

	return &DialerWithPinning{
		isReported:   false,
		reportURI:    reportURI,
		report:       report,
		proxyManager: proxyManager,
		log:          log,
	}
}

func NewPMAPIPinning(appVersion string) *DialerWithPinning {
	return NewDialerWithPinning(
		"https://reports.protonmail.ch/reports/tls",
		TLSReport{
			EffectiveExpirationDate:   time.Now().Add(365 * 24 * 60 * 60 * time.Second).Format(time.RFC3339),
			IncludeSubdomains:         false,
			ValidatedCertificateChain: []string{},
			ServedCertificateChain:    []string{},
			AppVersion:                appVersion,

			// NOTE: the proxy pins are the same for all proxy servers, guaranteed by infra team ;)
			KnownPins: []string{
				`pin-sha256="drtmcR2kFkM8qJClsuWgUzxgBkePfRCkRpqUesyDmeE="`, // current
				`pin-sha256="YRGlaY0jyJ4Jw2/4M8FIftwbDIQfh8Sdro96CeEel54="`, // hot
				`pin-sha256="AfMENBVvOS8MnISprtvyPsjKlPooqh8nMB/pvCrpJpw="`, // cold
				`pin-sha256="EU6TS9MO0L/GsDHvVc9D5fChYLNy5JdGYpJw0ccgetM="`, // proxy main
				`pin-sha256="iKPIHPnDNqdkvOnTClQ8zQAIKG0XavaPkcEo0LBAABA="`, // proxy backup 1
				`pin-sha256="MSlVrBCdL0hKyczvgYVSRNm88RicyY04Q2y5qrBt0xA="`, // proxy backup 2
				`pin-sha256="C2UxW0T1Ckl9s+8cXfjXxlEqwAfPM4HiW2y3UdtBeCw="`, // proxy backup 3
			},
		},
	)
}

func (p *DialerWithPinning) reportCertIssue(connState tls.ConnectionState) {
	p.isReported = true

	if p.ReportCertIssueLocal != nil {
		go p.ReportCertIssueLocal()
	}

	if p.reportURI != "" {
		p.report.NotedHostname = connState.ServerName
		p.report.ServedCertificateChain = marshalCert7468(connState.PeerCertificates)

		if len(connState.VerifiedChains) > 0 {
			p.report.ServedCertificateChain = marshalCert7468(
				connState.VerifiedChains[len(connState.VerifiedChains)-1],
			)
		}

		go p.reportCertIssueRemote()
	}
}

func (p *DialerWithPinning) reportCertIssueRemote() {
	b, err := json.Marshal(p.report)
	if err != nil {
		p.log.Errorf("marshal request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", p.reportURI, bytes.NewReader(b))
	if err != nil {
		p.log.Errorf("create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", CurrentUserAgent)
	req.Header.Set("x-pm-apiversion", strconv.Itoa(Version))
	req.Header.Set("x-pm-appversion", p.report.AppVersion)

	p.log.Debugf("report req: %+v\n", req)

	c := &http.Client{}
	res, err := c.Do(req)
	p.log.Debugf("res: %+v\nerr: %v", res, err)
	if err != nil {
		return
	}
	_, _ = ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		p.log.Errorf("response status: %v", res.Status)
	}
	_ = res.Body.Close()
}

func certFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return fmt.Sprintf(`pin-sha256=%q`, base64.StdEncoding.EncodeToString(hash[:]))
}

func marshalCert7468(certs []*x509.Certificate) (pemCerts []string) {
	var buffer bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&buffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			logrus.WithField("pkg", "pmapi/tls-pinning").Errorf("encoding TLS cert: %v", err)
		}
		pemCerts = append(pemCerts, buffer.String())
		buffer.Reset()
	}

	return pemCerts
}

func (p *DialerWithPinning) TransportWithPinning() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialTLS:               p.dialAndCheckFingerprints,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Minute,
		ExpectContinueTimeout: 500 * time.Millisecond,

		// GODT-126: this was initially 10s but logs from users showed a significant number
		// were hitting this timeout, possibly due to flaky wifi taking >10s to reconnect.
		// Bumping to 30s for now to avoid this problem.
		ResponseHeaderTimeout: 30 * time.Second,

		// If we allow up to 30 seconds for response headers, it is reasonable to allow up
		// to 30 seconds for the TLS handshake to take place.
		TLSHandshakeTimeout: 30 * time.Second,
	}
}

// dialAndCheckFingerprint to set as http.Transport.DialTLS.
//
// * note that when DialTLS is not nil the Transport.TLSClientConfig and Transport.TLSHandshakeTimeout are ignored.
// * dialAndCheckFingerprints fails if certificate is not valid (not signed by authority or not matching hostname).
// * dialAndCheckFingerprints will pass if certificate pin does not have a match, but will send notification using
//   p.ReportCertIssueLocal() and p.reportCertIssueRemote() if they are not nil.
func (p *DialerWithPinning) dialAndCheckFingerprints(network, address string) (conn net.Conn, err error) {
	// If DoH is enabled, we hardfail on fingerprint mismatches.
	if globalIsDoHAllowed() && p.isReported {
		return nil, ErrTLSMatch
	}

	// Try to dial the given address but use a proxy if necessary.
	if conn, err = p.dialWithProxyFallback(network, address); err != nil {
		return
	}

	// If cert issue was already reported, we don't want to check fingerprints anymore.
	if p.isReported {
		return nil, ErrTLSMatch
	}

	// Check the cert fingerprint to ensure it is known.
	if err = p.checkFingerprints(conn); err != nil {
		p.log.WithError(err).Error("Error checking cert fingerprints")
		return
	}

	return
}

// dialWithProxyFallback tries to dial the given address but falls back to alternative proxies if need be.
func (p *DialerWithPinning) dialWithProxyFallback(network, address string) (conn net.Conn, err error) {
	var host, port string
	if host, port, err = net.SplitHostPort(address); err != nil {
		return
	}

	// Try to dial, and if it succeeds, then just return.
	if conn, err = p.dial(network, address); err == nil {
		return
	}

	// If DoH is not allowed, give up. Or, if we are dialing something other than the API
	// (e.g. we dial protonmail.com/... to check for updates), there's also no point in
	// continuing since a proxy won't help us reach that.
	if !globalIsDoHAllowed() || host != stripProtocol(GlobalGetRootURL()) {
		return
	}

	// Find a new proxy.
	var proxy string
	if proxy, err = p.proxyManager.findProxy(); err != nil {
		return
	}

	// Switch to the proxy.
	p.log.WithField("proxy", proxy).Debug("Switching to proxy")
	p.proxyManager.useProxy(proxy)

	// Retry dial with proxy.
	return p.dial(network, net.JoinHostPort(proxy, port))
}

// dial returns a connection to the given address using the given network.
func (p *DialerWithPinning) dial(network, address string) (conn net.Conn, err error) {
	var port string
	if p.report.Hostname, port, err = net.SplitHostPort(address); err != nil {
		return
	}
	if p.report.Port, err = strconv.Atoi(port); err != nil {
		return
	}
	p.report.DateTime = time.Now().Format(time.RFC3339)

	dialer := &net.Dialer{Timeout: 10 * time.Second}

	// If we are not dialing the standard API then we should skip cert verification checks.
	var tlsConfig *tls.Config = nil
	if address != stripProtocol(globalOriginalURL) {
		tlsConfig = &tls.Config{InsecureSkipVerify: true} // nolint[gosec]
	}

	return tls.DialWithDialer(dialer, network, address, tlsConfig)
}

func (p *DialerWithPinning) checkFingerprints(conn net.Conn) (err error) {
	if !checkTLSCerts {
		return
	}

	connState := conn.(*tls.Conn).ConnectionState()

	hasFingerprintMatch := false
	for _, peerCert := range connState.PeerCertificates {
		fingerprint := certFingerprint(peerCert)

		for i, pin := range p.report.KnownPins {
			if pin == fingerprint {
				hasFingerprintMatch = true

				if i != 0 {
					p.log.Warnf("Matched fingerprint (%q) was not primary pinned key (was key #%d)", fingerprint, i)
				}

				break
			}
		}

		if hasFingerprintMatch {
			break
		}
	}

	if !hasFingerprintMatch {
		p.reportCertIssue(connState)
		return ErrTLSMatch
	}

	return err
}
