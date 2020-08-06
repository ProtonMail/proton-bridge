// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package parser

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/emersion/go-message"
)

type Parser struct {
	stack []*Part
	root  *Part
}

func New(b []byte) (*Parser, error) {
	p := new(Parser)

	entity, err := message.Read(bytes.NewReader(b))
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, err
	}

	if err := p.parseEntity(entity); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Parser) NewWalker() *Walker {
	return newWalker(p.root)
}

func (p *Parser) NewVisitor(defaultRule VisitorRule) *Visitor {
	return newVisitor(p.root, defaultRule)
}

func (p *Parser) NewWriter() *Writer {
	return newWriter(p.root)
}

func (p *Parser) Root() *Part {
	return p.root
}

func (p *Parser) Part(number []int) (part *Part, err error) {
	part = p.root

	for _, n := range number {
		if part, err = part.Child(n); err != nil {
			return
		}
	}

	return
}

func (p *Parser) beginPart() {
	p.stack = append(p.stack, &Part{})
}

func (p *Parser) endPart() {
	var part *Part

	p.stack, part = p.stack[:len(p.stack)-1], p.stack[len(p.stack)-1]

	if len(p.stack) > 0 {
		p.top().children = append(p.top().children, part)
	} else {
		p.root = part
	}
}

func (p *Parser) top() *Part {
	if len(p.stack) == 0 {
		return nil
	}

	return p.stack[len(p.stack)-1]
}

func (p *Parser) withHeader(h message.Header) {
	p.top().Header = h
}

func (p *Parser) withBody(bytes []byte) {
	p.top().Body = bytes
}

func (p *Parser) parseEntity(e *message.Entity) error {
	p.beginPart()
	defer p.endPart()

	p.withHeader(e.Header)

	if mr := e.MultipartReader(); mr != nil {
		return p.parseMultipart(mr)
	}

	return p.parsePart(e)
}

func (p *Parser) parsePart(e *message.Entity) (err error) {
	bytes, err := ioutil.ReadAll(e.Body)
	if err != nil {
		return
	}

	p.withBody(bytes)

	return
}

func (p *Parser) parseMultipart(r message.MultipartReader) (err error) {
	for {
		var child *message.Entity

		if child, err = r.NextPart(); err != nil {
			return ignoreEOF(err)
		}

		if err = p.parseEntity(child); err != nil {
			return
		}
	}
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}

	return err
}
