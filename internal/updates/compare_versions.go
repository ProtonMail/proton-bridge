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

package updates

import (
	"regexp"
	"strconv"
	"strings"
)

var nonVersionChars = regexp.MustCompile(`([^0-9.]+)`) //nolint[gochecknoglobals]

// sanitizeVersion returns only numbers and periods.
func sanitizeVersion(version string) string {
	return nonVersionChars.ReplaceAllString(version, "")
}

// Result can be false positive, but must not be false negative.
// Assuming
//  * dot separated integers format e.g. "A.B.C.…" where A,B,C,… are integers
//  * `1.1` == `1.1.0` (i.e. first is not newer)
//  * `1.1.1` > `1.1` (i.e. first is newer)
func isFirstVersionNewer(first, second string) (firstIsNewer bool, err error) {
	first = sanitizeVersion(first)
	second = sanitizeVersion(second)

	firstIsNewer, err = false, nil
	if first == second {
		return
	}

	firstIsNewer = true
	var firstArr, secondArr []int
	if firstArr, err = versionStrToInts(first); err != nil {
		return
	}
	if secondArr, err = versionStrToInts(second); err != nil {
		return
	}

	verLength := max(len(firstArr), len(secondArr))
	firstArr = appendZeros(firstArr, verLength)
	secondArr = appendZeros(secondArr, verLength)

	for i := 0; i < verLength; i++ {
		if firstArr[i] == secondArr[i] {
			continue
		}
		return firstArr[i] > secondArr[i], nil
	}
	return false, nil
}

func versionStrToInts(version string) (intArr []int, err error) {
	strArr := strings.Split(version, ".")
	intArr = make([]int, len(strArr))
	for index, item := range strArr {
		if item == "" {
			intArr[index] = 0
			continue
		}
		intArr[index], err = strconv.Atoi(item)
		if err != nil {
			return
		}
	}
	return
}

func appendZeros(ints []int, newsize int) []int {
	size := len(ints)
	if size >= newsize {
		return ints
	}
	zeros := make([]int, newsize-size)
	return append(ints, zeros...)
}

func max(ints ...int) (max int) {
	max = ints[0]
	for _, a := range ints {
		if max < a {
			max = a
		}
	}
	return
}
