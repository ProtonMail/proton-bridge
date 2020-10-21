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
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type localPart struct {
	value string
}

func (p *localPart) withDotAtom(dotAtom *dotAtom) {
	p.value = dotAtom.value
}

func (p *localPart) withQuotedString(quotedString *quotedString) {
	p.value = quotedString.value
}

func (p *localPart) withObsLocalPart(obsLocalPart *obsLocalPart) {
	p.value = strings.Join(obsLocalPart.words, ".")
}

func (w *walker) EnterLocalPart(ctx *parser.LocalPartContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering localPart")
	w.enter(&localPart{})
}

func (w *walker) ExitLocalPart(ctx *parser.LocalPartContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting localPart")

	type withLocalPart interface {
		withLocalPart(*localPart)
	}

	res := w.exit().(*localPart)

	if parent, ok := w.parent().(withLocalPart); ok {
		parent.withLocalPart(res)
	}
}
