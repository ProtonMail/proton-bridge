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
	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/sirupsen/logrus"
)

type encodedWord struct {
	value string
}

func (w *walker) EnterEncodedWord(ctx *parser.EncodedWordContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering encodedWord")

	word, err := pmmime.WordDec.Decode(ctx.GetText())
	if err != nil {
		word = ctx.GetText()
	}

	w.enter(&encodedWord{
		value: word,
	})
}

func (w *walker) ExitEncodedWord(ctx *parser.EncodedWordContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting encodedWord")

	type withEncodedWord interface {
		withEncodedWord(*encodedWord)
	}

	res := w.exit().(*encodedWord)

	if parent, ok := w.parent().(withEncodedWord); ok {
		parent.withEncodedWord(res)
	}
}
