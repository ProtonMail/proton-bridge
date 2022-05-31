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
	"bytes"
	"errors"
	"mime"
	"unicode/utf8"

	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"
	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
)

type Parts []*Part

type Part struct {
	Header   message.Header
	Body     []byte
	children Parts
}

func (p *Part) ContentType() (string, map[string]string, error) {
	t, params, err := p.Header.ContentType()
	if err != nil {
		// go-message's implementation of ContentType() doesn't handle duplicate parameters
		// e.g. Content-Type: text/plain; charset=utf-8; charset=UTF-8
		// so if it fails, we try again with pmmime's implementation, which does.
		t, params, err = pmmime.ParseMediaType(p.Header.Get("Content-Type"))
	}

	return t, params, err
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
	logrus.Trace("Converting part to utf-8")

	t, params, err := p.ContentType()
	if err != nil {
		return err
	}

	decoder := selectSuitableDecoder(p, t, params)

	if p.Body, err = decoder.Bytes(p.Body); err != nil {
		return err
	}

	if params == nil {
		params = make(map[string]string)
	}

	params["charset"] = "UTF-8"

	p.Header.SetContentType(t, params)

	return nil
}

func (p *Part) ConvertMetaCharset() error {
	doc, err := html.Parse(bytes.NewReader(p.Body))
	if err != nil {
		return err
	}

	goquery.NewDocumentFromNode(doc).Find("meta").Each(func(n int, sel *goquery.Selection) {
		if val, ok := sel.Attr("content"); ok {
			t, params, err := pmmime.ParseMediaType(val)
			if err != nil {
				return
			}

			params["charset"] = "UTF-8"

			sel.SetAttr("content", mime.FormatMediaType(t, params))
		}

		if _, ok := sel.Attr("charset"); ok {
			sel.SetAttr("charset", "UTF-8")
		}
	})

	buf := new(bytes.Buffer)

	if err := html.Render(buf, doc); err != nil {
		return err
	}

	p.Body = buf.Bytes()

	return nil
}

func selectSuitableDecoder(p *Part, t string, params map[string]string) *encoding.Decoder {
	if charset, ok := params["charset"]; ok {
		logrus.WithField("charset", charset).Trace("The part has a specified charset")

		if decoder, err := pmmime.SelectDecoder(charset); err == nil {
			logrus.Trace("The charset is known; decoder has been selected")
			return decoder
		}

		logrus.Warn("The charset is unknown; no decoder could be selected")
	}

	if utf8.Valid(p.Body) {
		logrus.Trace("The part is already valid utf-8, returning noop encoder")
		return encoding.Nop.NewDecoder()
	}

	encoding, name, _ := charset.DetermineEncoding(p.Body, t)

	logrus.WithField("name", name).Warn("Determined encoding by reading body")

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
	t, _, err := p.ContentType()
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
