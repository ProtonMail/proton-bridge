package focus

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestFocus_Raise(t *testing.T) {
	// Start the focus service.
	service, err := NewService(semver.MustParse("1.2.3"))
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

func TestFocus_Version(t *testing.T) {
	// Start the focus service.
	_, err := NewService(semver.MustParse("1.2.3"))
	require.NoError(t, err)

	// Try to dial it, it should succeed.
	version, ok := TryVersion()
	require.True(t, ok)
	require.Equal(t, "1.2.3", version.String())
}
