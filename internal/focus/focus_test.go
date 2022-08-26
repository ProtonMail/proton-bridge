package focus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFocusRaise(t *testing.T) {
	// Start the focus service.
	service, err := NewService()
	require.NoError(t, err)

	// Try to dial it, it should succeed.
	require.True(t, TryRaise())

	// The service should report a raise call.
	<-service.GetRaiseCh()

	// Stop the service.
	service.Close()

	// Try to dial it, it should fail.
	require.False(t, TryRaise())
}
