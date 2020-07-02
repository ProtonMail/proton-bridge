package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPart(t *testing.T) {
	p := newTestParser(t, "complex_structure.eml")

	wantParts := map[string]string{
		"":        "multipart/mixed",
		"1":       "text/plain",
		"2":       "application/octet-stream",
		"3":       "multipart/mixed",
		"3.1":     "text/plain",
		"3.2":     "application/octet-stream",
		"4":       "multipart/mixed",
		"4.1":     "image/gif",
		"4.2":     "multipart/mixed",
		"4.2.1":   "text/plain",
		"4.2.2":   "multipart/alternative",
		"4.2.2.1": "text/plain",
		"4.2.2.2": "text/html",
	}

	for partNumber, wantContType := range wantParts {
		part, err := p.Part(getPartNumber(partNumber))
		require.NoError(t, err)

		contType, _, err := part.Header.ContentType()
		require.NoError(t, err)
		assert.Equal(t, wantContType, contType)
	}
}

func getPartNumber(s string) (part []int) {
	if s == "" {
		return
	}

	for _, number := range strings.Split(s, ".") {
		i64, err := strconv.ParseInt(number, 10, 64)
		if err != nil {
			panic(err)
		}

		part = append(part, int(i64))
	}

	return
}
