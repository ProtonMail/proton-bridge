package parser

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestParser(t *testing.T, msg string) *Parser {
	r := f(msg)

	p, err := New(r)
	require.NoError(t, err)

	return p
}

func TestParserSpecifiedLatin1Charset(t *testing.T) {
	p := newTestParser(t, "text_plain_latin1.eml")

	checkBodies(t, p, "ééééééé")
}

func TestParserUnspecifiedLatin1Charset(t *testing.T) {
	p := newTestParser(t, "text_plain_unknown_latin1.eml")

	checkBodies(t, p, "ééééééé")
}

func TestParserSpecifiedLatin2Charset(t *testing.T) {
	p := newTestParser(t, "text_plain_latin2.eml")

	checkBodies(t, p, "řšřšřš")
}

func TestParserEmbeddedLatin2Charset(t *testing.T) {
	p := newTestParser(t, "text_html_embedded_latin2_encoding.eml")

	checkBodies(t, p, `<html><head><meta charset="ISO-8859-2"></head><body>latin2 řšřš</body></html>`)
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

func checkBodies(t *testing.T, p *Parser, wantBodies ...string) {
	var partBodies, expectedBodies [][]byte

	require.NoError(t, p.NewWalker().RegisterDefaultHandler(func(p *Part) (err error) {
		if p.Body != nil {
			partBodies = append(partBodies, p.Body)
		}

		return
	}).Walk())

	for _, body := range wantBodies {
		expectedBodies = append(expectedBodies, []byte(body))
	}

	assert.ElementsMatch(t, expectedBodies, partBodies)
}
