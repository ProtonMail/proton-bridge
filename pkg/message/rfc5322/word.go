// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package rfc5322

import (
	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type word struct {
	value string
}

func (w *word) withAtom(atom *atom) {
	w.value = atom.value
}

func (w *word) withQuotedString(quotedString *quotedString) {
	w.value = quotedString.value
}

func (w *word) withEncodedWord(encodedWord *encodedWord) {
	w.value = encodedWord.value
}

func (w *walker) EnterWord(ctx *parser.WordContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering word")

	w.enter(&word{
		value: ctx.GetText(),
	})
}

func (w *walker) ExitWord(ctx *parser.WordContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting word")

	type withWord interface {
		withWord(*word)
	}

	res := w.exit().(*word)

	if parent, ok := w.parent().(withWord); ok {
		parent.withWord(res)
	}
}
