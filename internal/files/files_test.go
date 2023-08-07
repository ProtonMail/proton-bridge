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

package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveDir(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()

	// Create some files in from.
	if err := os.WriteFile(filepath.Join(from, "a"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "b"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(from, "c"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "c", "d"), []byte("d"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Move the files.
	if err := MoveDir(from, to); err != nil {
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
