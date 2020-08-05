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
	p.children = append(p.children, child)
}

func (p *Part) isUTF8() bool {
	return utf8.Valid(p.Body)
}

// TODO: Do we then need to set charset to utf-8? What if it's embedded in html?
func (p *Part) convertToUTF8() error {
	t, params, err := p.Header.ContentType()
	if err != nil {
		return err
	}

	var decoder *encoding.Decoder

	if knownCharset, ok := params["charset"]; !ok {
		encoding, _, _ := charset.DetermineEncoding(p.Body, t)
		decoder = encoding.NewDecoder()
	} else if decoder, err = pmmime.SelectDecoder(knownCharset); err != nil {
		return err
	}

	if p.Body, err = decoder.Bytes(p.Body); err != nil {
		return err
	}

	params["charset"] = "utf-8"
	p.Header.SetContentType(t, params)

	return nil
}
