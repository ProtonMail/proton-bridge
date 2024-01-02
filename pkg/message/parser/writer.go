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
	"io"

	"github.com/emersion/go-message"
)

type Writer struct {
	root *Part
}

func newWriter(root *Part) *Writer {
	return &Writer{
		root: root,
	}
}

func (w *Writer) Write(ww io.Writer) error {
	if !w.root.is7BitClean() {
		w.root.Header.Set("Content-Transfer-Encoding", "base64")
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
	if !p.is7BitClean() {
		p.Header.Set("Content-Transfer-Encoding", "base64")
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
