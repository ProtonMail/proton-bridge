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

func (w *Writer) WithCondition(cond Condition) *Writer {
	w.cond = append(w.cond, cond)
	return w
}

func (w *Writer) Write(ww io.Writer) (err error) {
	msgWriter, err := message.CreateWriter(ww, w.root.Header)
	if err != nil {
		return
	}

	if err = w.root.write(msgWriter, w); err != nil {
		return
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
