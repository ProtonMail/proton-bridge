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

type qtext struct {
	value string
}

func (w *walker) EnterQtext(ctx *parser.QtextContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering qtext")

	w.enter(&qtext{
		value: ctx.GetText(),
	})
}

func (w *walker) ExitQtext(ctx *parser.QtextContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting qtext")

	type withQtext interface {
		withQtext(*qtext)
	}

	res := w.exit().(*qtext)

	if parent, ok := w.parent().(withQtext); ok {
		parent.withQtext(res)
	}
}
