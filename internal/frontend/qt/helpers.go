// Copyright (c) 2022 Proton AG
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

//go:build build_qt
// +build build_qt

package qt

import (
	"regexp"
	"strings"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

// getCursorPos returns current mouse position to be able to use in QML
func getCursorPos() *core.QPoint {
	return gui.QCursor_Pos()
}

// newQByteArrayFromString is a wrapper for new QByteArray from string.
func newQByteArrayFromString(name string) *core.QByteArray {
	return core.NewQByteArray2(name, len(name))
}

var (
	reMultiSpaces     = regexp.MustCompile(`\s{2,}`)
	reStartWithSymbol = regexp.MustCompile(`^[.,/#!$@%^&*;:{}=\-_` + "`" + `~()]`)
)

// getInitials based on webapp implementation:
// https://github.com/ProtonMail/WebClients/blob/55d96a8b4afaaa4372fc5f1ef34953f2070fd7ec/packages/shared/lib/helpers/string.ts#L145
func getInitials(fullName string) string {
	words := strings.Split(
		reMultiSpaces.ReplaceAllString(fullName, " "),
		" ",
	)

	n := 0
	for _, word := range words {
		if !reStartWithSymbol.MatchString(word) {
			words[n] = word
			n++
		}
	}

	if n == 0 {
		return "?"
	}

	initials := words[0][0:1]
	if n != 1 {
		initials += words[n-1][0:1]
	}
	return strings.ToUpper(initials)
}
