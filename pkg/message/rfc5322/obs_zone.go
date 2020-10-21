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
	"errors"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type obsZone struct {
	location *time.Location
}

func (w *walker) EnterObsZone(ctx *parser.ObsZoneContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering obsZone")

	loc := time.UTC

	switch strings.ToLower(ctx.GetText()) {
	case "ut":
		loc = time.FixedZone(ctx.GetText(), 0)
	case "utc":
		loc = time.FixedZone(ctx.GetText(), 0)
	case "gmt":
		loc = time.FixedZone(ctx.GetText(), 0)
	case "est":
		loc = time.FixedZone(ctx.GetText(), -5*60*60)
	case "edt":
		loc = time.FixedZone(ctx.GetText(), -4*60*60)
	case "cst":
		loc = time.FixedZone(ctx.GetText(), -6*60*60)
	case "cdt":
		loc = time.FixedZone(ctx.GetText(), -5*60*60)
	case "mst":
		loc = time.FixedZone(ctx.GetText(), -7*60*60)
	case "mdt":
		loc = time.FixedZone(ctx.GetText(), -6*60*60)
	case "pst":
		loc = time.FixedZone(ctx.GetText(), -8*60*60)
	case "pdt":
		loc = time.FixedZone(ctx.GetText(), -7*60*60)
	default:
		w.err = errors.New("bad timezone")
	}

	w.enter(&obsZone{
		location: loc,
	})
}

func (w *walker) ExitObsZone(ctx *parser.ObsZoneContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting obsZone")

	type withObsZone interface {
		withObsZone(*obsZone)
	}

	res := w.exit().(*obsZone)

	if parent, ok := w.parent().(withObsZone); ok {
		parent.withObsZone(res)
	}
}
