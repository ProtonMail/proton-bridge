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

package parser

import "regexp"

type HandlerFunc func(*Part) error

type handler struct {
	typeRegExp, dispRegExp string
	fn                     HandlerFunc
}

func (h *handler) matchPart(p *Part) bool {
	return h.matchType(p) || h.matchDisp(p)
}

func (h *handler) matchType(p *Part) bool {
	if h.typeRegExp == "" {
		return false
	}

	t, _, err := p.Header.ContentType()
	if err != nil {
		t = ""
	}

	return regexp.MustCompile(h.typeRegExp).MatchString(t)
}

func (h *handler) matchDisp(p *Part) bool {
	if h.dispRegExp == "" {
		return false
	}

	disp, _, err := p.Header.ContentDisposition()
	if err != nil {
		disp = ""
	}

	return regexp.MustCompile(h.dispRegExp).MatchString(disp)
}
