// Copyright (c) 2022 Proton AG
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
	"io"

	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	stack []*Part
	root  *Part
}

func New(r io.Reader) (*Parser, error) {
	p := new(Parser)

	entity, err := message.Read(newEndOfMailTrimmer(r))
	if err != nil {
		switch {
		case message.IsUnknownCharset(err):
			logrus.WithError(err).Warning("Message has an unknown charset")
		case message.IsUnknownEncoding(err):
			logrus.WithError(err).Warning("Message has an unknown encoding")
		default:
			return nil, err
		}
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

// Section returns the message part referred to by the given section. A section
// is zero or more integers. For example, section 1.2.3 will return the third
// part of the second part of the first part of the message.
func (p *Parser) Section(section []int) (part *Part, err error) {
	part = p.root

	for _, n := range section {
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
	bytes, err := io.ReadAll(e.Body)
	if err != nil {
		return
	}

	p.withBody(bytes)

	return
}

func (p *Parser) parseMultipart(r message.MultipartReader) (err error) {
	for {
		var child *message.Entity

		if child, err = r.NextPart(); err != nil && !message.IsUnknownCharset(err) {
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
