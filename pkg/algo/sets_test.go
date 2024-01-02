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

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type T struct {
	k, v int
}

func TestSetIntersection(t *testing.T) {
	keysAreEqual := func(a, b interface{}) bool {
		return a.(T).k == b.(T).k //nolint:forcetypeassert
	}

	type args struct {
		a  interface{}
		b  interface{}
		eq func(a, b interface{}) bool
	}

	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "integer sets",
			args: args{a: []int{1, 2, 3}, b: []int{3, 4, 5}, eq: func(a, b interface{}) bool { return a == b }},
			want: []int{3},
		},
		{
			name: "string sets",
			args: args{a: []string{"1", "2", "3"}, b: []string{"3", "4", "5"}, eq: func(a, b interface{}) bool { return a == b }},
			want: []string{"3"},
		},
		{
			name: "custom comp, only compare on keys, prefer first set if keys are the same",
			args: args{a: []T{{k: 1, v: 1}, {k: 2, v: 2}}, b: []T{{k: 2, v: 1234}, {k: 3, v: 3}}, eq: keysAreEqual},
			want: []T{{k: 2, v: 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// using cmp.Equal because it handles the interfaces correctly; testify/assert doesn't
			// treat these as equal because their types are different ([]interface vs []int)
			if got := SetIntersection(tt.args.a, tt.args.b, tt.args.eq); cmp.Equal(got, tt.want) {
				t.Errorf("SetIntersection() = %v, want %v", got, tt.want)
			}
		})
	}
}
