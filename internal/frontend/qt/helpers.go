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

// +build !nogui

package qt

import (
	"bufio"
	"os"
	"os/exec"
	"time"

	"github.com/therecipe/qt/core"
)

// NewQByteArrayFromString is a wrapper for new QByteArray from string.
func NewQByteArrayFromString(name string) *core.QByteArray {
	return core.NewQByteArray2(name, len(name))
}

// NewQVariantString is a wrapper for QVariant alocator String.
func NewQVariantString(data string) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantStringArray is a wrapper for QVariant alocator String Array.
func NewQVariantStringArray(data []string) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantBool is a wrapper for QVariant alocator Bool.
func NewQVariantBool(data bool) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantInt is a wrapper for QVariant alocator Int.
func NewQVariantInt(data int) *core.QVariant {
	return core.NewQVariant1(data)
}

// Pause is used to show GUI tests.
func Pause() {
	time.Sleep(500 * time.Millisecond)
}

// PauseLong is used to diplay GUI tests.
func PauseLong() {
	time.Sleep(3 * time.Second)
}

// FIXME: Not working in test...
func WaitForEnter() {
	log.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func FcMatchSans() (family string) {
	family = "DejaVu Sans"
	fcMatch, err := exec.Command("fc-match", "-f", "'%{family}'", "sans-serif").Output()
	if err == nil {
		return string(fcMatch)
	}
	return
}
