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
	"errors"
	"unicode/utf8"

	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/emersion/go-message"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
)

type Parts []*Part

type Part struct {
	Header   message.Header
	Body     []byte
	children Parts
}

func (p *Part) Child(n int) (part *Part, err error) {
	if len(p.children) < n {
		return nil, errors.New("no such part")
	}

	return p.children[n-1], nil
}

func (p *Part) Children() Parts {
	return p.children
}

func (p *Part) AddChild(child *Part) {
	if p.isMultipartMixed() {
		p.children = append(p.children, child)
	} else {
		root := &Part{
			Header:   getContentHeaders(p.Header),
			Body:     p.Body,
			children: p.children,
		}

		p.Body = nil
		p.children = Parts{root, child}
		stripContentHeaders(&p.Header)
		p.Header.Set("Content-Type", "multipart/mixed")
	}
}

func (p *Part) ConvertToUTF8() error {
	if utf8.Valid(p.Body) {
		return nil
	}

	t, params, err := p.Header.ContentType()
	if err != nil {
		return err
	}

	decoder := selectSuitableDecoder(p, t, params)

	if p.Body, err = decoder.Bytes(p.Body); err != nil {
		return err
	}

	// HELP: Is this okay? What about when the charset is embedded in structured text type eg html/xml?
	if params == nil {
		params = make(map[string]string)
	}

	params["charset"] = "utf-8"

	p.Header.SetContentType(t, params)

	return nil
}

func selectSuitableDecoder(p *Part, t string, params map[string]string) *encoding.Decoder {
	if charset, ok := params["charset"]; ok {
		if decoder, err := pmmime.SelectDecoder(charset); err == nil {
			return decoder
		}
	}

	encoding, _, _ := charset.DetermineEncoding(p.Body, t)

	return encoding.NewDecoder()
}

func (p *Part) is7BitClean() bool {
	for _, b := range p.Body {
		if b > 1<<7 {
			return false
		}
	}

	return true
}

func (p *Part) isMultipartMixed() bool {
	t, _, err := p.Header.ContentType()
	if err != nil {
		return false
	}

	return t == "multipart/mixed"
}

func getContentHeaders(header message.Header) message.Header {
	var res message.Header

	res.Set("Content-Type", header.Get("Content-Type"))
	res.Set("Content-Disposition", header.Get("Content-Disposition"))
	res.Set("Content-Transfer-Encoding", header.Get("Content-Transfer-Encoding"))

	return res
}

func stripContentHeaders(header *message.Header) {
	header.Del("Content-Type")
	header.Del("Content-Disposition")
	header.Del("Content-Transfer-Encoding")
}
