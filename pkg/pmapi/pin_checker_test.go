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
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeClientInfoProvider struct {
	version, useragent string
}

func (c *fakeClientInfoProvider) GetAppVersion() string {
	return c.version
}

func (c *fakeClientInfoProvider) GetUserAgent() string {
	return c.useragent
}

func TestPinCheckerDoubleReport(t *testing.T) {
	reportCounter := 0

	reportServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportCounter++
	}))

	r := newTLSReporter(newPinChecker(TrustedAPIPins), &fakeClientInfoProvider{version: "3", useragent: "useragent"})

	// Report the same issue many times.
	for i := 0; i < 10; i++ {
		r.reportCertIssue(reportServer.URL, "myhost", "443", tls.ConnectionState{})
	}

	// We should only report once.
	assert.Eventually(t, func() bool {
		return reportCounter == 1
	}, time.Second, time.Millisecond)

	// If we then report something else many times.
	for i := 0; i < 10; i++ {
		r.reportCertIssue(reportServer.URL, "anotherhost", "443", tls.ConnectionState{})
	}

	// We should get a second report.
	assert.Eventually(t, func() bool {
		return reportCounter == 2
	}, time.Second, time.Millisecond)
}
