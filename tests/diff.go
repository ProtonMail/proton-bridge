package tests

import (
	"encoding/json"
	"reflect"

	"github.com/bradenaw/juniper/xslices"
)

func ToAny(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	var a any

	if err := json.Unmarshal(b, &a); err != nil {
		panic(err)
	}

	return a
}

func IsSub(outer, inner any) bool {
	if outer == nil && inner != nil {
		return IsSub(reflect.Zero(reflect.TypeOf(inner)).Interface(), inner)
	}

	if outer != nil && inner == nil {
		return IsSub(reflect.Zero(reflect.TypeOf(outer)).Interface(), outer)
	}

	switch inner := inner.(type) {
	case map[string]any:
		outer, ok := outer.(map[string]any)
		if !ok {
			return false
		}

		return isSubMap(outer, inner)

	case []any:
		outer, ok := outer.([]any)
		if !ok {
			return false
		}

		return isSubSlice(outer, inner)

	default:
		if reflect.TypeOf(outer) != reflect.TypeOf(inner) {
			return false
		}

		if reflect.DeepEqual(outer, inner) {
			return true
		}

		return reflect.DeepEqual(reflect.Zero(reflect.TypeOf(inner)).Interface(), inner)
	}
}

func isSubMap(outer, inner map[string]any) bool {
	for k, v := range inner {
		w, ok := outer[k]
		if !ok {
			for _, w := range outer {
				if IsSub(w, inner) {
					return true
				}
			}
		}

		if !IsSub(w, v) {
			return false
		}
	}

	return true
}

func isSubSlice(outer, inner []any) bool {
	if len(inner) != len(outer) {
		return false
	}

	for _, v := range inner {
		if xslices.IndexFunc(outer, func(outer any) bool {
			return IsSub(outer, v)
		}) < 0 {
			return false
		}
	}

	return true
}
