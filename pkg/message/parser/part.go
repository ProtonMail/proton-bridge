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
	"bytes"
	"errors"
	"mime"
	"unicode/utf8"

	pmmime "github.com/ProtonMail/proton-bridge/v3/pkg/mime"
	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
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

const utf8Charset = "UTF-8"

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

func (p *Part) ContentDisposition() (string, map[string]string, error) {
	return pmmime.ParseMediaType(p.Header.Get("Content-Disposition"))
}

func (p *Part) HasContentID() bool {
	return len(p.Header.Get("content-id")) != 0
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

func (p *Part) InsertChild(index int, child *Part) {
	if p.isMultipartMixedOrRelated() {
		p.children = slices.Insert(p.children, index, child)
	} else {
		p.AddChild(child)
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

	params["charset"] = utf8Charset

	p.Header.SetContentType(t, params)

	return nil
}

func (p *Part) ConvertMetaCharset() error {
	doc, err := html.Parse(bytes.NewReader(p.Body))
	if err != nil {
		return err
	}

	// Override charset to UTF-8 in meta headers only if needed.
	var metaModified = false
	goquery.NewDocumentFromNode(doc).Find("meta").Each(func(_ int, sel *goquery.Selection) {
		if val, ok := sel.Attr("content"); ok {
			t, params, err := pmmime.ParseMediaType(val)
			if err != nil {
				logrus.WithField("pkg", "parser").WithError(err).Error("Meta tag parsing fails.")
				return
			}

			if charset, ok := params["charset"]; ok && charset != utf8Charset {
				params["charset"] = utf8Charset
				sel.SetAttr("content", mime.FormatMediaType(t, params))
				metaModified = true
			}
		}

		if charset, ok := sel.Attr("charset"); ok && charset != utf8Charset {
			sel.SetAttr("charset", utf8Charset)
			metaModified = true
		}
	})

	// Override the body part only if modification was applied
	// as html.render will sanitise the html headers.
	if metaModified {
		buf := new(bytes.Buffer)

		if err := html.Render(buf, doc); err != nil {
			return err
		}

		p.Body = buf.Bytes()
	}
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

func (p *Part) isMultipartMixedOrRelated() bool {
	t, _, err := p.ContentType()
	if err != nil {
		return false
	}

	return t == "multipart/mixed" || t == "multipart/related"
}

func (p *Part) IsAttachment() bool {
	disp, _, err := p.ContentDisposition()
	if err != nil {
		disp = ""
	}
	return disp == "attachment"
}

func getContentHeaders(header message.Header) message.Header {
	var res message.Header

	if contentType := header.Get("Content-Type"); contentType != "" {
		res.Set("Content-Type", contentType)
	}

	if contentDisposition := header.Get("Content-Disposition"); contentDisposition != "" {
		res.Set("Content-Disposition", contentDisposition)
	}

	if contentTransferEncoding := header.Get("Content-Transfer-Encoding"); contentTransferEncoding != "" {
		res.Set("Content-Transfer-Encoding", contentTransferEncoding)
	}

	return res
}

func stripContentHeaders(header *message.Header) {
	header.Del("Content-Type")
	header.Del("Content-Disposition")
	header.Del("Content-Transfer-Encoding")
}
