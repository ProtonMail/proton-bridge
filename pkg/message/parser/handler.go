// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package parser

import (
	"regexp"
)

type HandlerFunc func(*Part) error

type handler struct {
	typeRegExp, dispRegExp *regexp.Regexp
	fn                     HandlerFunc
}

func (h *handler) matchPart(p *Part) bool {
	return h.matchType(p) || h.matchDisp(p)
}

func (h *handler) matchPartSkipAttachment(p *Part) bool {
	return !p.IsAttachment() && h.matchPart(p)
}

func (h *handler) matchType(p *Part) bool {
	if h.typeRegExp == nil {
		return false
	}

	t, _, err := p.ContentType()
	if err != nil {
		t = ""
	}

	return h.typeRegExp.MatchString(t)
}

func (h *handler) matchDisp(p *Part) bool {
	if h.dispRegExp == nil {
		return false
	}

	disp, _, err := p.ContentDisposition()
	if err != nil {
		disp = ""
	}

	return h.dispRegExp.MatchString(disp)
}
