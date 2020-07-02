package parser

import (
	"io"
	"io/ioutil"

	"github.com/emersion/go-message"
)

// TODO: Set this to something that handles charsets.
func init() { // nolint[gochecknoinits]
	message.CharsetReader = func(string, io.Reader) (io.Reader, error) {
		panic("not implemented")
	}
}

type Parser struct {
	stack []*Part
	root  *Part
}

func New(r io.Reader) (p *Parser, err error) {
	p = new(Parser)

	if err = p.parse(r); err != nil {
		return
	}

	return
}

func (p *Parser) NewWalker() *Walker {
	return newWalker(p.root)
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
		if part, err = part.Part(n); err != nil {
			return
		}
	}

	return
}

func (p *Parser) parse(r io.Reader) (err error) {
	e, err := message.Read(r)
	if err != nil {
		return
	}

	return p.parseEntity(e)
}

func (p *Parser) enter() {
	p.stack = append(p.stack, &Part{})
}

func (p *Parser) exit() {
	var built *Part

	p.stack, built = p.stack[:len(p.stack)-1], p.stack[len(p.stack)-1]

	if len(p.stack) > 0 {
		p.top().children = append(p.top().children, built)
	} else {
		p.root = built
	}
}

func (p *Parser) top() *Part {
	return p.stack[len(p.stack)-1]
}

func (p *Parser) withHeader(h message.Header) {
	p.top().Header = h
}

func (p *Parser) withBody(bytes []byte) {
	p.top().Body = bytes
}

func (p *Parser) parseEntity(e *message.Entity) (err error) {
	p.enter()
	defer p.exit()

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
