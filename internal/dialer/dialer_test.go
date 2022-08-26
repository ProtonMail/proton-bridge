package dialer

import (
	"testing"

	"golang.org/x/net/http/httpproxy"
)

// skipIfProxyIsSet skips the tests if HTTPS proxy is set.
// Should be used for tests depending on proper certificate checks which
// is not possible under our CI setup.
func skipIfProxyIsSet(t *testing.T) {
	if httpproxy.FromEnvironment().HTTPSProxy != "" {
		t.SkipNow()
	}
}
