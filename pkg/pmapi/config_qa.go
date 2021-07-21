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

// +build build_qa

package pmapi

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
)

func getRootURL() string {
	// This config allows to dynamically change ROOT URL.
	url := os.Getenv("PMAPI_ROOT_URL")
	if strings.HasPrefix(url, "http") {
		return url
	}
	if url != "" {
		return "https://" + url
	}
	return "https://api.protonmail.ch"
}

func newProxyDialerAndTransport(cfg Config) (*ProxyTLSDialer, http.RoundTripper) {
	transport := CreateTransportWithDialer(NewBasicTLSDialer(cfg))

	// TLS certificate of testing environment might be self-signed.
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return nil, transport
}
