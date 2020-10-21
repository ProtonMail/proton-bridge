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

type nameAddr struct {
	name, address string
}

func (a *nameAddr) withDisplayName(displayName *displayName) {
	a.name = strings.Join(displayName.words, " ")
}

func (a *nameAddr) withAngleAddr(angleAddr *angleAddr) {
	a.address = angleAddr.address
}

func (w *walker) EnterNameAddr(ctx *parser.NameAddrContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering nameAddr")
	w.enter(&nameAddr{})
}

func (w *walker) ExitNameAddr(ctx *parser.NameAddrContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting nameAddr")

	type withNameAddr interface {
		withNameAddr(*nameAddr)
	}

	res := w.exit().(*nameAddr)

	if parent, ok := w.parent().(withNameAddr); ok {
		parent.withNameAddr(res)
	}
}
