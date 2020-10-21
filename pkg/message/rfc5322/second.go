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

type second struct {
	value int
}

func (w *walker) EnterSecond(ctx *parser.SecondContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Entering second")

	var text string

	for _, digit := range ctx.AllDigit() {
		text += digit.GetText()
	}

	val, err := strconv.Atoi(text)
	if err != nil {
		w.err = err
	}

	w.enter(&second{
		value: val,
	})
}

func (w *walker) ExitSecond(ctx *parser.SecondContext) {
	logrus.WithField("text", ctx.GetText()).Trace("Exiting second")

	type withSecond interface {
		withSecond(*second)
	}

	res := w.exit().(*second)

	if parent, ok := w.parent().(withSecond); ok {
		parent.withSecond(res)
	}
}
