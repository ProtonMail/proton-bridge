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
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/stretchr/testify/assert"
)

func TestTLSReporter_DoubleReport(t *testing.T) {
	reportCounter := 0

	reportServer := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		reportCounter++
	}))

	r := NewTLSReporter("hostURL", "appVersion", useragent.New(), TrustedAPIPins)

	// Report the same issue many times.
	for i := 0; i < 10; i++ {
		r.ReportCertIssue(reportServer.URL, "myhost", "443", tls.ConnectionState{})
	}

	// We should only report once.
	assert.Eventually(t, func() bool {
		return reportCounter == 1
	}, time.Second, time.Millisecond)

	// If we then report something else many times.
	for i := 0; i < 10; i++ {
		r.ReportCertIssue(reportServer.URL, "anotherhost", "443", tls.ConnectionState{})
	}

	// We should get a second report.
	assert.Eventually(t, func() bool {
		return reportCounter == 2
	}, time.Second, time.Millisecond)
}
