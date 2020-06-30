package parser

import (
	"github.com/emersion/go-message"
)

type Part struct {
	Header   message.Header
	Body     []byte
	children []*Part
}

func (p *Part) visit(w *Walker) (err error) {
	if err = p.handle(w); err != nil {
		return
	}

	for _, child := range p.children {
		if err = child.visit(w); err != nil {
			return
		}
	}

	return
}

func (p *Part) getTypeHandler(w *Walker) (hdl PartHandler) {
	t, _, err := p.Header.ContentType()
	if err != nil {
		return
	}

	return w.typeHandlers[t]
}

func (p *Part) getDispHandler(w *Walker) (hdl DispHandler) {
	t, _, err := p.Header.ContentDisposition()
	if err != nil {
		return
	}

	return w.dispHandlers[t]
}

func (p *Part) handle(w *Walker) (err error) {
	typeHandler := p.getTypeHandler(w)
	dispHandler := p.getDispHandler(w)
	defaultHandler := w.defaultHandler

	switch {
	case dispHandler != nil && typeHandler != nil:
		return dispHandler(p, typeHandler)

	case dispHandler != nil && typeHandler == nil:
		return dispHandler(p, defaultHandler)

	case dispHandler == nil && typeHandler != nil:
		return typeHandler(p)

	default:
		return defaultHandler(p)
	}
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
