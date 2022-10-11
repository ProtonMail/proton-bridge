package safe

import "testing"

func TestSafe_Map(t *testing.T) {
	m := NewMap(func(a, b string, data map[string]string) bool {
		return a < b
	})

	m.Set("a", "b")

	if !m.Has("a") {
		t.Fatal("expected to have key")
	}

	if m.Has("b") {
		t.Fatal("expected not to have key")
	}

	m.Set("b", "c")

	if !m.Has("b") {
		t.Fatal("expected to have key")
	}

	if !m.HasFunc(func(key string, val string) bool {
		return key == "b"
	}) {
		t.Fatal("expected to have key")
	}

	if !m.Get("b", func(val string) {
		if val != "c" {
			t.Fatal("expected to have value")
		}
	}) {
		t.Fatal("expected to have key")
	}

	if !m.Index(0, func(key string, val string) {
		if key != "a" || val != "b" {
			t.Fatal("expected to have key and value")
		}
	}) {
		t.Fatal("expected to have index")
	}

	if !m.Index(1, func(key string, val string) {
		if key != "b" || val != "c" {
			t.Fatal("expected to have key and value")
		}
	}) {
		t.Fatal("expected to have index")
	}

	if m.Index(2, func(key string, val string) {
		t.Fatal("expected not to have index")
	}) {
		t.Fatal("expected not to have index")
	}

	if !m.GetDelete("b", func(val string) {
		if val != "c" {
			t.Fatal("expected to have value")
		}
	}) {
		t.Fatal("expected to have key")
	}

	if m.Has("b") {
		t.Fatal("expected not to have key")
	}

	if m.GetDelete("b", func(val string) {
		t.Fatal("expected not to have value")
	}) {
		t.Fatal("expected not to have key")
	}

	if !m.Index(0, func(key string, val string) {
		if key != "a" || val != "b" {
			t.Fatal("expected to have key and value")
		}
	}) {
		t.Fatal("expected to have index")
	}

	m.Values(func(val []string) {
		if len(val) != 1 {
			t.Fatal("expected to have values")
		}

		if val[0] != "b" {
			t.Fatal("expected to have value")
		}
	})
}
