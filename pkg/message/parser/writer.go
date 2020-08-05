package parser

import (
	"io"

	"github.com/emersion/go-message"
)

type Writer struct {
	root *Part
	cond []Condition
}

type Condition func(p *Part) bool

func newWriter(root *Part) *Writer {
	return &Writer{
		root: root,
	}
}

// WithCondition allows setting a condition when parts should be written.
// Parts are passed to each condition set and if any condition returns false,
// the part is not written.
// This initially seemed like a good idea but is now kinda useless.
func (w *Writer) WithCondition(cond Condition) *Writer {
	w.cond = append(w.cond, cond)
	return w
}

func (w *Writer) Write(ww io.Writer) error {
	if w.shouldFilter(w.root) {
		w.root.Header.Add("Content-Transfer-Encoding", "base64")
	}

	msgWriter, err := message.CreateWriter(ww, w.root.Header)
	if err != nil {
		return err
	}

	if err := w.write(msgWriter, w.root); err != nil {
		return err
	}

	return msgWriter.Close()
}

func (w *Writer) shouldWrite(p *Part) bool {
	for _, cond := range w.cond {
		if !cond(p) {
			return false
		}
	}

	return true
}

func (w *Writer) shouldFilter(p *Part) bool {
	encoding := p.Header.Get("Content-Transfer-Encoding")

	if encoding != "" && encoding == "quoted-printable" || encoding == "base64" {
		return false
	}

	for _, b := range p.Body {
		if uint8(b) > 1<<7 {
			return true
		}
	}

	return false
}

func (w *Writer) write(writer *message.Writer, p *Part) error {
	if len(p.children) > 0 {
		for _, child := range p.children {
			if err := w.writeAsChild(writer, child); err != nil {
				return err
			}
		}
	}

	if _, err := writer.Write(p.Body); err != nil {
		return err
	}

	return nil
}

func (w *Writer) writeAsChild(writer *message.Writer, p *Part) error {
	if !w.shouldWrite(p) {
		return nil
	}

	if w.shouldFilter(p) {
		p.Header.Add("Content-Transfer-Encoding", "base64")
	}

	childWriter, err := writer.CreatePart(p.Header)
	if err != nil {
		return err
	}

	if err := w.write(childWriter, p); err != nil {
		return err
	}

	return childWriter.Close()
}
