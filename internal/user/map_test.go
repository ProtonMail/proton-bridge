package user

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	type Key int

	type Value string

	type Data struct {
		key   Key
		value Value
	}

	m := newOrdMap(
		func(d Data) Key { return d.key },
		func(d Data) Value { return d.value },
		func(a, b Data) bool { return a.key < b.key },
		Data{key: 1, value: "a"},
		Data{key: 2, value: "b"},
		Data{key: 3, value: "c"},
	)

	// Insert some new data.
	m.insert(Data{key: 4, value: "d"})
	m.insert(Data{key: 5, value: "e"})

	// Delete some data.
	require.Equal(t, Value("c"), m.delete(3))
	require.Equal(t, Value("a"), m.delete(1))
	require.Equal(t, Value("e"), m.delete(5))

	// Check the remaining keys and values are correct.
	require.Equal(t, []Key{2, 4}, m.keys())
	require.Equal(t, []Value{"b", "d"}, m.values())

	// Overwrite some data.
	m.insert(Data{key: 2, value: "two"})
	m.insert(Data{key: 4, value: "four"})

	// Check the remaining keys and values are correct.
	require.Equal(t, []Key{2, 4}, m.keys())
	require.Equal(t, []Value{"two", "four"}, m.values())
}
