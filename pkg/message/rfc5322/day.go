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
	"strconv"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type day struct {
	value int
}

func (w *walker) EnterDay(ctx *parser.DayContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering day")

	var text string

	for _, digit := range ctx.AllDigit() {
		text += digit.GetText()
	}

	val, err := strconv.Atoi(text)
	if err != nil {
		w.err = err
	}

	w.enter(&day{
		value: val,
	})
}

func (w *walker) ExitDay(ctx *parser.DayContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting day")

	type withDay interface {
		withDay(*day)
	}

	res := w.exit().(*day)

	if parent, ok := w.parent().(withDay); ok {
		parent.withDay(res)
	}
}
