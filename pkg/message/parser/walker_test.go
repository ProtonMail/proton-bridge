package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalker(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	allBodies := [][]byte{}

	walker := p.NewWalker().
		RegisterDefaultHandler(func(p *Part) (err error) {
			if p.Body != nil {
				allBodies = append(allBodies, p.Body)
			}
			return
		})

	assert.NoError(t, walker.Walk())
	assert.ElementsMatch(t, [][]byte{
		[]byte("<html><body>This is body of <b>HTML mail</b> with attachment</body></html>"),
		[]byte("if you are reading this, hi!"),
	}, allBodies)
}

func TestWalkerTypeHandler(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	html := [][]byte{}

	walker := p.NewWalker().
		RegisterContentTypeHandler("text/html", func(p *Part) (err error) {
			html = append(html, p.Body)
			return
		})

	assert.NoError(t, walker.Walk())
	assert.ElementsMatch(t, [][]byte{
		[]byte("<html><body>This is body of <b>HTML mail</b> with attachment</body></html>"),
	}, html)
}

func TestWalkerDispositionHandler(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	attachments := [][]byte{}

	walker := p.NewWalker().
		RegisterContentDispositionHandler("attachment", func(p *Part) (err error) {
			attachments = append(attachments, p.Body)
			return
		})

	assert.NoError(t, walker.Walk())
	assert.ElementsMatch(t, [][]byte{
		[]byte("if you are reading this, hi!"),
	}, attachments)
}

func TestWalkerDispositionAndTypeHandler_TypeDefinedFirst(t *testing.T) {
	p := newTestParser(t, "text_html_octet_attachment.eml")

	var typeCalled, dispCalled bool

	walker := p.NewWalker().
		RegisterContentTypeHandler("application/octet-stream", func(p *Part) (err error) {
			typeCalled = true
			return
		}).
		RegisterContentDispositionHandler("attachment", func(p *Part) (err error) {
			dispCalled = true
			return
		})

	assert.NoError(t, walker.Walk())
	assert.True(t, typeCalled)
	assert.False(t, dispCalled)
}
