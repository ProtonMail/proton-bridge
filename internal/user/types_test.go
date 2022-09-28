package user

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToType(t *testing.T) {
	type myString string

	// Slices of different types are not equal.
	require.NotEqual(t, []myString{"a", "b", "c"}, []string{"a", "b", "c"})

	// But converting them to the same type makes them equal.
	require.Equal(t, []myString{"a", "b", "c"}, mapTo[string, myString]([]string{"a", "b", "c"}))

	// The conversion can happen in the other direction too.
	require.Equal(t, []string{"a", "b", "c"}, mapTo[myString, string]([]myString{"a", "b", "c"}))
}
