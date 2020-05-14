package pmapi

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPinCheckerDoubleReport(t *testing.T) {
	reportCounter := 0

	reportServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportCounter++
	}))

	pc := newPinChecker(TrustedAPIPins)

	// Report the same issue many times.
	for i := 0; i < 10; i++ {
		pc.reportCertIssue(reportServer.URL, "myhost", "443", tls.ConnectionState{}, "3", "useragent")
	}

	// We should only report once.
	assert.Eventually(t, func() bool {
		return reportCounter == 1
	}, time.Second, time.Millisecond)

	// If we then report something else many times.
	for i := 0; i < 10; i++ {
		pc.reportCertIssue(reportServer.URL, "anotherhost", "443", tls.ConnectionState{}, "3", "useragent")
	}

	// We should get a second report.
	assert.Eventually(t, func() bool {
		return reportCounter == 2
	}, time.Second, time.Millisecond)
}
