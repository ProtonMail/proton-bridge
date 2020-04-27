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

	"github.com/sirupsen/logrus"
)

type PinChecker struct {
	trustedPins []string
}

func NewPinChecker(trustedPins []string) PinChecker {
	return PinChecker{
		trustedPins: trustedPins,
	}
}

// CheckCertificate returns whether the connection presents a known TLS certificate.
func (p *PinChecker) CheckCertificate(conn net.Conn) error {
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

// ReportCertIssue reports a TLS key mismatch.
func (p *PinChecker) ReportCertIssue(host, port, datetime string, connState tls.ConnectionState, appVersion, userAgent string) {
	var certChain []string

	if len(connState.VerifiedChains) > 0 {
		certChain = marshalCert7468(connState.VerifiedChains[len(connState.VerifiedChains)-1])
	} else {
		certChain = marshalCert7468(connState.PeerCertificates)
	}

	report := NewTLSReport(host, port, connState.ServerName, certChain, p.trustedPins, appVersion)

	go postCertIssueReport(report, userAgent)
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
