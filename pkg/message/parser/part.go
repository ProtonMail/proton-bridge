package parser

import (
	"errors"

	"github.com/emersion/go-message"
)

type Parts []*Part

type Part struct {
	Header   message.Header
	Body     []byte
	children Parts
}

func (p *Part) Part(n int) (part *Part, err error) {
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

func (p *Part) write(writer *message.Writer, w *Writer) (err error) {
	if len(p.children) > 0 {
		for _, child := range p.children {
			if err = child.writeAsChild(writer, w); err != nil {
				return
			}
		}
	}

	if _, err = writer.Write(p.Body); err != nil {
		return
	}

	return
}

func (p *Part) writeAsChild(writer *message.Writer, w *Writer) (err error) {
	if !w.shouldWrite(p) {
		return
	}

	childWriter, err := writer.CreatePart(p.Header)
	if err != nil {
		return
	}

	if err = p.write(childWriter, w); err != nil {
		return
	}

	return childWriter.Close()
}
