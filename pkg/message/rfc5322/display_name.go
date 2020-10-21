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

type displayName struct {
	words []string
}

func (n *displayName) withWord(word *word) {
	n.words = append(n.words, word.value)
}

func (w *walker) EnterDisplayName(ctx *parser.DisplayNameContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering displayName")
	w.enter(&displayName{})
}

func (w *walker) ExitDisplayName(ctx *parser.DisplayNameContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting displayName")

	type withDisplayName interface {
		withDisplayName(*displayName)
	}

	res := w.exit().(*displayName)

	if parent, ok := w.parent().(withDisplayName); ok {
		parent.withDisplayName(res)
	}
}
