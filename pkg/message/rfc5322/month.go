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

type month struct {
	value time.Month
}

func (w *walker) EnterMonth(ctx *parser.MonthContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering month")

	var m time.Month

	switch strings.ToLower(ctx.GetText()) {
	case "jan":
		m = time.January
	case "feb":
		m = time.February
	case "mar":
		m = time.March
	case "apr":
		m = time.April
	case "may":
		m = time.May
	case "jun":
		m = time.June
	case "jul":
		m = time.July
	case "aug":
		m = time.August
	case "sep":
		m = time.September
	case "oct":
		m = time.October
	case "nov":
		m = time.November
	case "dec":
		m = time.December
	default:
		w.err = errors.New("no such month")
	}

	w.enter(&month{
		value: m,
	})
}

func (w *walker) ExitMonth(ctx *parser.MonthContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting month")

	type withMonth interface {
		withMonth(*month)
	}

	res := w.exit().(*month)

	if parent, ok := w.parent().(withMonth); ok {
		parent.withMonth(res)
	}
}
