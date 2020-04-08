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

package args

import (
	"os"
	"strings"
)

// FilterProcessSerialNumberFromArgs removes additional flag from MacOS. More info ProcessSerialNumber
// http://mirror.informatimago.com/next/developer.apple.com/documentation/Carbon/Reference/Process_Manager/prmref_main/data_type_5.html#//apple_ref/doc/uid/TP30000208/C001951
func FilterProcessSerialNumberFromArgs() {
	tmp := os.Args[:0]
	for _, arg := range os.Args {
		if !strings.Contains(arg, "-psn_") {
			tmp = append(tmp, arg)
		}
	}
	os.Args = tmp
}
