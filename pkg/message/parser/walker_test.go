// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

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

func TestWalkerTypeHandler_excludingAttachment(t *testing.T) {
	p := newTestParser(t, "forwarding_html_attachment.eml")

	html := [][]byte{}
	plain := [][]byte{}

	walker := p.NewWalker().
		RegisterContentTypeHandler("text/html", func(p *Part) (err error) {
			html = append(html, p.Body)
			return
		}).
		RegisterContentTypeHandler("text/plain", func(p *Part) (err error) {
			plain = append(plain, p.Body)
			return
		})

	assert.NoError(t, walker.WalkSkipAttachment())
	assert.Equal(t, 1, len(plain))
	assert.Equal(t, 0, len(html))
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
		RegisterContentTypeHandler("application/octet-stream", func(_ *Part) (err error) {
			typeCalled = true
			return
		}).
		RegisterContentDispositionHandler("attachment", func(_ *Part) (err error) {
			dispCalled = true
			return
		})

	assert.NoError(t, walker.Walk())
	assert.True(t, typeCalled)
	assert.False(t, dispCalled)
}
