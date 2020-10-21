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
	"fmt"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type angleAddr struct {
	address string
}

func (a *angleAddr) withAddrSpec(addrSpec *addrSpec) {
	a.address = fmt.Sprintf("%v@%v", addrSpec.localPart, addrSpec.domain)
}

func (a *angleAddr) withObsAngleAddr(obsAngleAddr *obsAngleAddr) {
	a.address = obsAngleAddr.address
}

func (w *walker) EnterAngleAddr(ctx *parser.AngleAddrContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering angleAddr")
	w.enter(&angleAddr{})
}

func (w *walker) ExitAngleAddr(ctx *parser.AngleAddrContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting angleAddr")

	type withAngleAddr interface {
		withAngleAddr(*angleAddr)
	}

	res := w.exit().(*angleAddr)

	if parent, ok := w.parent().(withAngleAddr); ok {
		parent.withAngleAddr(res)
	}
}
