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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_IsSub(t *testing.T) {
	tests := []struct {
		outer string
		inner string
		want  bool
	}{
		{
			outer: `{}`,
			inner: `{}`,
			want:  true,
		},
		{
			outer: `{"a": 1}`,
			inner: `{"a": 1}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": 2}`,
			inner: `{"a": 1}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": 2}`,
			inner: `{"a": 1, "c": 3}`,
			want:  false,
		},
		{
			outer: `{"a": 1, "b": {"c": 2}}`,
			inner: `{"c": 2}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": 3}}`,
			inner: `{"c": 2}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": 3}}`,
			inner: `{"c": 2, "d": 3}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": 3}}`,
			inner: `{"c": 2, "e": 3}`,
			want:  false,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": "ignore"}}`,
			inner: `{"a": 1, "b": {"c": 2, "d": ""}}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": null}}`,
			inner: `{"a": 1, "b": {"c": 2, "d": null}}`,
			want:  true,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": ["1"]}}`,
			inner: `{"a": 1, "b": {"c": 2, "d": []}}`,
			want:  false,
		},
		{
			outer: `{"a": 1, "b": {"c": 2, "d": []}}`,
			inner: `{"a": 1, "b": {"c": 2, "d": null}}`,
			want:  true,
		},
		{
			outer: `{"a": []}`,
			inner: `{"a": []}`,
			want:  true,
		},
		{
			outer: `{"a": [1, 2]}`,
			inner: `{"a": [1, 2]}`,
			want:  true,
		},
		{
			outer: `{"a": [1, 3]}`,
			inner: `{"a": [1, 2]}`,
			want:  false,
		},
		{
			outer: `{"a": [1, 2, 3]}`,
			inner: `{"a": [1, 2]}`,
			want:  false,
		},
		{
			outer: `{"a": null}`,
			inner: `{"a": []}`,
			want:  true,
		},
		{
			outer: `{"a": []}`,
			inner: `{"a": null}`,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v vs %v", tt.inner, tt.outer), func(t *testing.T) {
			var outerMap, innerMap map[string]any

			if err := json.Unmarshal([]byte(tt.outer), &outerMap); err != nil {
				t.Fatal(err)
			}

			if err := json.Unmarshal([]byte(tt.inner), &innerMap); err != nil {
				t.Fatal(err)
			}

			if got := IsSub(outerMap, innerMap); got != tt.want {
				t.Errorf("isSub() = %v, want %v", got, tt.want)
			}
		})
	}
}
