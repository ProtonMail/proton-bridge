package safe

import "testing"

func TestValue(t *testing.T) {
	v := NewValue("foo")

	v.Load(func(data string) {
		if data != "foo" {
			t.Error("expected foo")
		}
	})

	v.Save("bar")

	v.Load(func(data string) {
		if data != "bar" {
			t.Error("expected bar")
		}
	})

	v.Mod(func(data *string) {
		*data = "baz"
	})

	v.Load(func(data string) {
		if data != "baz" {
			t.Error("expected baz")
		}
	})

	if LoadRet(v, func(data string) string {
		return data
	}) != "baz" {
		t.Error("expected baz")
	}
}
