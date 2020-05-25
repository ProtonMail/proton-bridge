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
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	proxyUseDuration         = 24 * time.Hour
	proxyLookupWait          = 5 * time.Second
	proxyCacheRefreshTimeout = 20 * time.Second
	proxyDoHTimeout          = 20 * time.Second
	proxyCanReachTimeout     = 20 * time.Second
	proxyQuery               = "dMFYGSLTQOJXXI33ONVQWS3BOMNUA.protonpro.xyz"
)

var dohProviders = []string{ //nolint[gochecknoglobals]
	"https://dns11.quad9.net/dns-query",
	"https://dns.google/dns-query",
}

// proxyProvider manages known proxies.
type proxyProvider struct {
	// dohLookup is used to look up the given query at the given DoH provider, returning the TXT records>
	dohLookup func(ctx context.Context, query, provider string) (urls []string, err error)

	providers  []string // List of known doh providers.
	query      string   // The query string used to find proxies.
	proxyCache []string // All known proxies, cached in case DoH providers are unreachable.

	cacheRefreshTimeout time.Duration
	dohTimeout          time.Duration
	canReachTimeout     time.Duration

	lastLookup time.Time // The time at which we last attempted to find a proxy.
}

// newProxyProvider creates a new proxyProvider that queries the given DoH providers
// to retrieve DNS records for the given query string.
func newProxyProvider(providers []string, query string) (p *proxyProvider) { // nolint[unparam]
	p = &proxyProvider{
		providers:           providers,
		query:               query,
		cacheRefreshTimeout: proxyCacheRefreshTimeout,
		dohTimeout:          proxyDoHTimeout,
		canReachTimeout:     proxyCanReachTimeout,
	}

	// Use the default DNS lookup method; this can be overridden if necessary.
	p.dohLookup = p.defaultDoHLookup

	return
}

// findReachableServer returns a working API server (either proxy or standard API).
func (p *proxyProvider) findReachableServer() (proxy string, err error) {
	logrus.Debug("Trying to find a reachable server")

	if time.Now().Before(p.lastLookup.Add(proxyLookupWait)) {
		return "", errors.New("not looking for a proxy, too soon")
	}

	p.lastLookup = time.Now()

	// We use a waitgroup to wait for both
	//  a) the check whether the API is reachable, and
	//  b) the DoH queries.
	// This is because the Alternative Routes v2 spec says:
	//  Call the GET /test/ping route on normal API domain (same time as DoH requests and wait until all have finished)
	var wg sync.WaitGroup
	var apiReachable bool

	wg.Add(2)

	go func() {
		defer wg.Done()
		apiReachable = p.canReach(rootURL)
	}()

	go func() {
		defer wg.Done()
		err = p.refreshProxyCache()
	}()

	wg.Wait()

	if apiReachable {
		proxy = rootURL
		return
	}

	if err != nil {
		return
	}

	for _, url := range p.proxyCache {
		if p.canReach(url) {
			proxy = url
			return
		}
	}

	return "", errors.New("no reachable server could be found")
}

// refreshProxyCache loads the latest proxies from the known providers.
// If the process takes longer than proxyCacheRefreshTimeout, an error is returned.
func (p *proxyProvider) refreshProxyCache() error {
	logrus.Info("Refreshing proxy cache")

	ctx, cancel := context.WithTimeout(context.Background(), p.cacheRefreshTimeout)
	defer cancel()

	resultChan := make(chan []string)

	go func() {
		for _, provider := range p.providers {
			if proxies, err := p.dohLookup(ctx, p.query, provider); err == nil {
				resultChan <- proxies
				return
			}
		}
	}()

	select {
	case result := <-resultChan:
		p.proxyCache = result
		return nil

	case <-ctx.Done():
		return errors.New("timed out while refreshing proxy cache")
	}
}

// canReach returns whether we can reach the given url.
func (p *proxyProvider) canReach(url string) bool {
	logrus.WithField("url", url).Debug("Trying to ping proxy")

	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	}

	dialer := NewPinningTLSDialer(NewBasicTLSDialer())

	pinger := resty.New().
		SetHostURL(url).
		SetTimeout(p.canReachTimeout).
		SetTransport(CreateTransportWithDialer(dialer))

	if _, err := pinger.R().Get("/tests/ping"); err != nil {
		logrus.WithField("proxy", url).WithError(err).Warn("Failed to ping proxy")
		return false
	}

	return true
}

// defaultDoHLookup is the default implementation of the proxy manager's DoH lookup.
// It looks up DNS TXT records for the given query URL using the given DoH provider.
// It returns a list of all found TXT records.
// If the whole process takes more than proxyDoHTimeout then an error is returned.
func (p *proxyProvider) defaultDoHLookup(ctx context.Context, query, dohProvider string) (data []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.dohTimeout)
	defer cancel()

	dataChan, errChan := make(chan []string), make(chan error)

	go func() {
		// Build new DNS request in RFC1035 format.
		dnsRequest := new(dns.Msg).SetQuestion(dns.Fqdn(query), dns.TypeTXT)

		// Pack the DNS request message into wire format.
		rawRequest, err := dnsRequest.Pack()
		if err != nil {
			errChan <- errors.Wrap(err, "failed to pack DNS request")
			return
		}

		// Encode wire-format DNS request message as base64url (RFC4648) without padding chars.
		encodedRequest := base64.RawURLEncoding.EncodeToString(rawRequest)

		// Make DoH request to the given DoH provider.
		rawResponse, err := resty.New().R().SetContext(ctx).SetQueryParam("dns", encodedRequest).Get(dohProvider)
		if err != nil {
			errChan <- errors.Wrap(err, "failed to make DoH request")
			return
		}

		// Unpack the DNS response.
		dnsResponse := new(dns.Msg)
		if err = dnsResponse.Unpack(rawResponse.Body()); err != nil {
			errChan <- errors.Wrap(err, "failed to unpack DNS response")
			return
		}

		// Pick out the TXT answers.
		for _, answer := range dnsResponse.Answer {
			if t, ok := answer.(*dns.TXT); ok {
				data = append(data, t.Txt...)
			}
		}

		dataChan <- data
	}()

	select {
	case data = <-dataChan:
		logrus.WithField("data", data).Info("Received TXT records")
		return

	case err = <-errChan:
		logrus.WithField("provider", dohProvider).WithError(err).Error("Failed to query DNS records")
		return

	case <-ctx.Done():
		logrus.WithField("provider", dohProvider).Error("Timed out querying DNS records")
		return []string{}, errors.New("timed out querying DNS records")
	}
}
