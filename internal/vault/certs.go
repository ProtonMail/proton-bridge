// Copyright (c) 2026 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// GetBridgeTLSCert returns the PEM-encoded certificate for the bridge.
// If CertPEMPath is set, it will attempt to read the certificate from the file.
// Otherwise, or on read/validation failure, it will return the certificate from the vault.
func (vault *Vault) GetBridgeTLSCert() ([]byte, []byte) {
	certs := vault.getSafe().Certs

	if certPath, keyPath := certs.CustomCertPath, certs.CustomKeyPath; certPath != "" && keyPath != "" {
		if certPEM, keyPEM, err := readPEMCert(certPath, keyPath); err == nil {
			return certPEM, keyPEM
		}

		logrus.Error("Failed to read certificate from file, using default")
	}

	return certs.Bridge.Cert, certs.Bridge.Key
}

// SetBridgeTLSCertPath sets the path to PEM-encoded certificates for the bridge.
func (vault *Vault) SetBridgeTLSCertPath(certPath, keyPath string) error {
	if _, _, err := readPEMCert(certPath, keyPath); err != nil {
		return fmt.Errorf("invalid certificate: %w", err)
	}

	return vault.modSafe(func(data *Data) {
		data.Certs.CustomCertPath = certPath
		data.Certs.CustomKeyPath = keyPath
	})
}

// SetBridgeTLSCertKey sets the path to PEM-encoded certificates for the bridge.
func (vault *Vault) SetBridgeTLSCertKey(cert, key []byte) error {
	return vault.modSafe(func(data *Data) {
		data.Certs.Bridge.Cert = cert
		data.Certs.Bridge.Key = key
	})
}

// HasCustomBridgeTLSCert reports whether Bridge is configured to load its TLS certificate from custom files.
func (vault *Vault) HasCustomBridgeTLSCert() bool {
	certs := vault.getSafe().Certs

	return certs.CustomCertPath != "" && certs.CustomKeyPath != ""
}

// SetBridgeTLSCertForAddress generates and stores a managed Bridge TLS certificate for address.
func (vault *Vault) SetBridgeTLSCertForAddress(address string) error {
	addresses := BridgeTLSCertAddresses(address)

	logrus.WithFields(logrus.Fields{
		"advertiseAddress": address,
		"certAddresses":    addresses,
	}).Warn("Generating bridge TLS certificate")

	template, err := certs.NewTLSTemplate(addresses...)
	if err != nil {
		return err
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		return err
	}

	return vault.SetBridgeTLSCertKey(certPEM, keyPEM)
}

var (
	lookupDNS    = lookupBridgeTLSCertAddressesDNS
	lookupSystem = lookupBridgeTLSCertAddressesSystem
)

// BridgeTLSCertAddresses returns the advertised addresses covered by the managed Bridge TLS certificate.
func BridgeTLSCertAddresses(address string) []string {
	addresses := []string{address}

	if _, err := netip.ParseAddr(address); err == nil {
		return addresses
	}

	if !strings.Contains(address, ".") {
		addresses = append(addresses, address+".local")
	}

	resolvedAddresses := lookupDNS(address)
	if len(resolvedAddresses) == 0 {
		resolvedAddresses = lookupSystem(address)
	}

	addresses = append(addresses, resolvedAddresses...)

	addresses = uniqueStrings(addresses)
	logrus.WithFields(logrus.Fields{
		"advertiseAddress": address,
		"certAddresses":    addresses,
	}).Warn("Resolved bridge TLS certificate addresses")

	return addresses
}

func lookupBridgeTLSCertAddressesDNS(address string) []string {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		logrus.WithError(err).WithField("advertiseAddress", address).Debug("Failed to load resolver config for bridge TLS certificate addresses")
		return nil
	}

	var addresses []string
	client := dns.Client{Timeout: 2 * time.Second}

	for _, candidate := range dnsLookupCandidates(address, config.Search) {
		resolved, ok := queryDNSCandidate(client, config, candidate)
		if ok {
			addresses = append(addresses, candidate)
		}

		addresses = append(addresses, resolved...)
	}

	return uniqueStrings(addresses)
}

func dnsLookupCandidates(address string, search []string) []string {
	candidates := []string{address}

	if !strings.Contains(address, ".") {
		for _, domain := range search {
			domain = strings.TrimSuffix(domain, ".")
			if domain != "" {
				candidates = append(candidates, address+"."+domain)
			}
		}
	}

	return uniqueStrings(candidates)
}

func queryDNSCandidate(client dns.Client, config *dns.ClientConfig, candidate string) ([]string, bool) {
	var addresses []string
	var found bool

	for _, questionType := range []uint16{dns.TypeA, dns.TypeAAAA, dns.TypeCNAME} {
		answers, err := queryDNS(client, config, candidate, questionType)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"certAddress": candidate,
				"queryType":   dns.TypeToString[questionType],
			}).Debug("Failed DNS query for bridge TLS certificate address")
			continue
		}

		for _, answer := range answers {
			switch record := answer.(type) {
			case *dns.A:
				found = true
				addresses = append(addresses, record.A.String())
			case *dns.AAAA:
				found = true
				addresses = append(addresses, record.AAAA.String())
			case *dns.CNAME:
				found = true
				addresses = append(addresses, strings.TrimSuffix(record.Target, "."))
			}
		}
	}

	return uniqueStrings(addresses), found
}

func queryDNS(client dns.Client, config *dns.ClientConfig, name string, questionType uint16) ([]dns.RR, error) {
	question := new(dns.Msg)
	question.SetQuestion(dns.Fqdn(name), questionType)

	for _, server := range config.Servers {
		response, _, err := client.Exchange(question, net.JoinHostPort(server, config.Port))
		if err != nil {
			continue
		}

		if response.Rcode == dns.RcodeSuccess {
			return response.Answer, nil
		}
	}

	return nil, fmt.Errorf("no successful DNS response")
}

func lookupBridgeTLSCertAddressesSystem(address string) []string {
	var addresses []string

	if cname, err := net.LookupCNAME(address); err == nil {
		addresses = append(addresses, strings.TrimSuffix(cname, "."))
	} else {
		logrus.WithError(err).WithField("advertiseAddress", address).Debug("Failed to resolve bridge TLS certificate canonical name")
	}

	if ips, err := net.LookupIP(address); err == nil {
		for _, ip := range ips {
			addresses = append(addresses, ip.String())
		}
	} else {
		logrus.WithError(err).WithField("advertiseAddress", address).Debug("Failed to resolve bridge TLS certificate IP addresses")
	}

	return uniqueStrings(addresses)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		if value == "" {
			continue
		}

		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func readPEMCert(certPEMPath, keyPEMPath string) ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(filepath.Clean(certPEMPath))
	if err != nil {
		return nil, nil, err
	}

	keyPEM, err := os.ReadFile(filepath.Clean(keyPEMPath))
	if err != nil {
		return nil, nil, err
	}

	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return nil, nil, err
	}

	return certPEM, keyPEM, nil
}
