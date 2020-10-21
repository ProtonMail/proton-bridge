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
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type zone struct {
	location *time.Location
}

func (z *zone) withOffset(offset *offset) {
	z.location = time.FixedZone(offset.rep, offset.value)
}

func (z *zone) withObsZone(obsZone *obsZone) {
	z.location = obsZone.location
}

func (w *walker) EnterZone(ctx *parser.ZoneContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering zone")

	w.enter(&zone{
		location: time.UTC,
	})
}

func (w *walker) ExitZone(ctx *parser.ZoneContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting zone")

	type withZone interface {
		withZone(*zone)
	}

	res := w.exit().(*zone)

	if parent, ok := w.parent().(withZone); ok {
		parent.withZone(res)
	}
}
