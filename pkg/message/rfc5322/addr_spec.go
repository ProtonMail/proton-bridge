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

type addrSpec struct {
	localPart, domain string
}

func (a *addrSpec) withLocalPart(localPart *localPart) {
	a.localPart = localPart.value
}

func (a *addrSpec) withDomain(domain *domain) {
	a.domain = domain.value
}

func (a *addrSpec) withPort(port *port) {
	a.domain += ":" + port.value
}

func (w *walker) EnterAddrSpec(ctx *parser.AddrSpecContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering addrSpec")
	w.enter(&addrSpec{})
}

func (w *walker) ExitAddrSpec(ctx *parser.AddrSpecContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting addrSpec")

	type withAddrSpec interface {
		withAddrSpec(*addrSpec)
	}

	res := w.exit().(*addrSpec)

	if parent, ok := w.parent().(withAddrSpec); ok {
		parent.withAddrSpec(res)
	}
}
