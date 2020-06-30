package parser

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestParser(t *testing.T, msg string) *Parser {
	r := f(msg)

	p, err := New(r)
	require.NoError(t, err)

	return p
}

func f(filename string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", filename))

	if err != nil {
		panic(err)
	}

	return f
}

func s(filename string) string {
	b, err := ioutil.ReadAll(f(filename))
	if err != nil {
		panic(err)
	}

	return string(b)
}
