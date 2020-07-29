package parser

import (
	"errors"

	"github.com/emersion/go-message"
)

type Part struct {
	Header   message.Header
	Body     []byte
	parent   *Part
	children []*Part
}

func (p *Part) Part(n int) (part *Part, err error) {
	if len(p.children) < n {
		return nil, errors.New("no such part")
	}

	return p.children[n-1], nil
}

func (p *Part) Parts() (n int) {
	return len(p.children)
}

func (p *Part) Parent() *Part {
	return p.parent
}

func (p *Part) Siblings() []*Part {
	if p.parent == nil {
		return nil
	}

	siblings := []*Part{}

	for _, sibling := range p.parent.children {
		if sibling != p {
			siblings = append(siblings, sibling)
		}
	}

	return siblings
}

func (p *Part) AddChild(child *Part) {
	p.children = append(p.children, child)
}

func (p *Part) visit(w *Walker) (err error) {
	hdl := p.getHandler(w)

	if err = hdl.handleEnter(w, p); err != nil {
		return
	}

	for _, child := range p.children {
		if err = child.visit(w); err != nil {
			return
		}
	}

	return hdl.handleExit(w, p)
}

func (p *Part) getHandler(w *Walker) handler {
	if dispHandler := w.getDispHandler(p); dispHandler != nil {
		return dispHandler
	}

	return w.getTypeHandler(p)
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
