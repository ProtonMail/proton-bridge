// Copyright (c) 2024 Proton AG
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

package algo

import "reflect"

// SetIntersection complexity: O(n^2), could be better but this is simple enough.
func SetIntersection(a, b interface{}, eq func(a, b interface{}) bool) []interface{} {
	set := make([]interface{}, 0)
	av := reflect.ValueOf(a)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		if contains(b, el, eq) {
			set = append(set, el)
		}
	}

	return set
}

func contains(a, e interface{}, eq func(a, b interface{}) bool) bool {
	v := reflect.ValueOf(a)

	for i := 0; i < v.Len(); i++ {
		if eq(v.Index(i).Interface(), e) {
			return true
		}
	}

	return false
}
