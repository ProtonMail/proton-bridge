package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveDir(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()

	// Create some files in from.
	if err := os.WriteFile(filepath.Join(from, "a"), []byte("a"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "b"), []byte("b"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(from, "c"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "c", "d"), []byte("d"), 0600); err != nil {
		t.Fatal(err)
	}

	// Move the files.
	if err := moveDir(from, to); err != nil {
		t.Fatal(err)
	}

	// Check that the files were moved.
	if _, err := os.Stat(filepath.Join(from, "a")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(to, "a")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(from, "b")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(to, "b")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(from, "c")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(to, "c")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(from, "c", "d")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(to, "c", "d")); err != nil {
		t.Fatal(err)
	}
}
