package safe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlice(t *testing.T) {
	s := NewSlice(1, 2, 3, 4, 5)

	{
		var have []int

		s.Iter(func(val int) {
			have = append(have, val)
		})

		require.Equal(t, []int{1, 2, 3, 4, 5}, have)
	}

	s.Append(6)
	s.Delete(3)

	{
		var have []int

		s.Iter(func(val int) {
			have = append(have, val)
		})

		require.Equal(t, []int{1, 2, 4, 5, 6}, have)
	}
}
