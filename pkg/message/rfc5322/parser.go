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
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/sirupsen/logrus"
)

// ParseAddressList parses one or more valid RFC5322 (with RFC2047) addresses.
func ParseAddressList(input string) ([]*mail.Address, error) {
	if len(input) == 0 {
		return []*mail.Address{}, nil
	}

	l := parser.NewRFC5322Lexer(antlr.NewInputStream(input))
	p := parser.NewRFC5322Parser(antlr.NewCommonTokenStream(l, antlr.TokenDefaultChannel))
	w := &walker{}

	p.AddErrorListener(w)
	p.AddParseListener(&parseListener{rules: p.GetRuleNames()})

	antlr.ParseTreeWalkerDefault.Walk(w, p.AddressList())

	return w.res.([]*mail.Address), w.err
}

// ParseDateTime parses a valid RFC5322 date-time.
func ParseDateTime(input string) (time.Time, error) {
	if len(input) == 0 {
		return time.Time{}, nil
	}

	l := parser.NewRFC5322Lexer(antlr.NewInputStream(input))
	p := parser.NewRFC5322Parser(antlr.NewCommonTokenStream(l, antlr.TokenDefaultChannel))
	w := &walker{}

	p.AddErrorListener(w)
	p.AddParseListener(&parseListener{rules: p.GetRuleNames()})

	antlr.ParseTreeWalkerDefault.Walk(w, p.DateTime())

	return w.res.(time.Time), w.err
}

type parseListener struct {
	antlr.BaseParseTreeListener

	rules []string
}

func (l *parseListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	logrus.
		WithField("rule", l.rules[ctx.GetRuleIndex()]).
		WithField("text", ctx.GetText()).
		Trace("Entering rule")
}

func (l *parseListener) ExitEveryRule(ctx antlr.ParserRuleContext) {
	logrus.
		WithField("rule", l.rules[ctx.GetRuleIndex()]).
		WithField("text", ctx.GetText()).
		Trace("Exiting rule")
}
