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
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/sirupsen/logrus"
)

type year struct {
	value int
}

func (w *walker) EnterYear(ctx *parser.YearContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering year")

	var text string

	for _, digit := range ctx.AllDigit() {
		text += digit.GetText()
	}

	val, err := strconv.Atoi(text)
	if err != nil {
		w.err = err
	}

	// NOTE: 2-digit years are obsolete but let's just have some simple handling anyway.
	if len(text) == 2 {
		if val > time.Now().Year()%100 {
			val += 1900
		} else {
			val += 2000
		}
	}

	w.enter(&year{
		value: val,
	})
}

func (w *walker) ExitYear(ctx *parser.YearContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting year")

	type withYear interface {
		withYear(*year)
	}

	res := w.exit().(*year)

	if parent, ok := w.parent().(withYear); ok {
		parent.withYear(res)
	}
}
