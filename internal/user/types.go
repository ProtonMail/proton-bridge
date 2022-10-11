package user

import (
	"fmt"
	"reflect"
)

func mapTo[From, To any](from []From) []To {
	to := make([]To, 0, len(from))

	for _, from := range from {
		val, ok := reflect.ValueOf(from).Convert(reflect.TypeOf(to).Elem()).Interface().(To)
		if !ok {
			panic(fmt.Sprintf("cannot convert %T to %T", from, *new(To)))
		}

		to = append(to, val)
	}

	return to
}
