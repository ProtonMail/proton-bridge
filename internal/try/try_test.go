package try

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTry(t *testing.T) {
	res, err := CatchVal(func() (string, error) {
		return "foo", nil
	})
	require.NoError(t, err)
	require.Equal(t, "foo", res)
}

func TestTryCatch(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			return nil
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}

func TestTryCatchError(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			return fmt.Errorf("catch error")
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}

func TestTryPanic(t *testing.T) {
	res, err := CatchVal(
		func() (string, error) {
			panic("oops")
		},
		func() error {
			return nil
		},
	)
	require.ErrorContains(t, err, "panic: oops")
	require.Zero(t, res)
}

func TestTryCatchPanic(t *testing.T) {
	tryErr := fmt.Errorf("oops")

	res, err := CatchVal(
		func() (string, error) {
			return "", tryErr
		},
		func() error {
			panic("oops")
		},
	)
	require.ErrorIs(t, err, tryErr)
	require.Zero(t, res)
}
