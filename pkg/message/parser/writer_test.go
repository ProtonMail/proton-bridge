package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserWrite(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	w := p.NewWriter()

	buf := new(bytes.Buffer)

	assert.NoError(t, w.Write(buf))
	assert.Equal(t, s("text_html_octet_attachment.eml"), crlf(buf.String()))
}

func TestParserWriteNoAttachments(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	w := p.
		NewWriter().
		WithCondition(func(p *Part) bool {
			// We don't write if the content disposition says it's an attachment.
			if disp, _, err := p.Header.ContentDisposition(); err == nil && disp == "attachment" {
				return false
			}

			return true
		})

	buf := new(bytes.Buffer)

	assert.NoError(t, w.Write(buf))
	assert.Equal(t, s("text_html.eml"), crlf(buf.String()))
}

func crlf(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
