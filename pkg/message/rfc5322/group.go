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
	"net/mail"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type group struct {
	addresses []*mail.Address
}

func (g *group) withGroupList(groupList *groupList) {
	g.addresses = append(g.addresses, groupList.addresses...)
}

func (w *walker) EnterGroup(ctx *parser.GroupContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering group")
	w.enter(&group{})
}

func (w *walker) ExitGroup(ctx *parser.GroupContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting group")

	type withGroup interface {
		withGroup(*group)
	}

	res := w.exit().(*group)

	if parent, ok := w.parent().(withGroup); ok {
		parent.withGroup(res)
	}
}
