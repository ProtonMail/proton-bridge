package tests

import (
	"fmt"
	"testing"

	"github.com/goccy/go-json"
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
