package user

import "reflect"

func mapTo[From, To any](from []From) []To {
	to := make([]To, 0, len(from))

	for _, from := range from {
		to = append(to, reflect.ValueOf(from).Convert(reflect.TypeOf(to).Elem()).Interface().(To))
	}

	return to
}
