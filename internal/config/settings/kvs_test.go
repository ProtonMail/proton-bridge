// Copyright (c) 2022 Proton AG
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

package settings

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadNoKeyValueStore(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestEmptyKeyValueStore(r)
	defer clean()

	r.Equal("", pref.Get("key"))
}

func TestLoadBadKeyValueStore(t *testing.T) {
	r := require.New(t)
	path, clean := newTmpFile(r)
	defer clean()

	r.NoError(ioutil.WriteFile(path, []byte("{\"key\":\"MISSING_QUOTES"), 0o700))
	pref := newKeyValueStore(path)
	r.Equal("", pref.Get("key"))
}

func TestKeyValueStor(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestKeyValueStore(r)
	defer clean()

	r.Equal("value", pref.Get("str"))
	r.Equal("42", pref.Get("int"))
	r.Equal("true", pref.Get("bool"))
	r.Equal("t", pref.Get("falseBool"))
}

func TestKeyValueStoreGetInt(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestKeyValueStore(r)
	defer clean()

	r.Equal(0, pref.GetInt("str"))
	r.Equal(42, pref.GetInt("int"))
	r.Equal(0, pref.GetInt("bool"))
	r.Equal(0, pref.GetInt("falseBool"))
}

func TestKeyValueStoreGetBool(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestKeyValueStore(r)
	defer clean()

	r.Equal(false, pref.GetBool("str"))
	r.Equal(false, pref.GetBool("int"))
	r.Equal(true, pref.GetBool("bool"))
	r.Equal(false, pref.GetBool("falseBool"))
}

func TestKeyValueStoreSetDefault(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestEmptyKeyValueStore(r)
	defer clean()

	pref.setDefault("key", "value")
	pref.setDefault("key", "othervalue")
	r.Equal("value", pref.Get("key"))
}

func TestKeyValueStoreSet(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestEmptyKeyValueStore(r)
	defer clean()

	pref.Set("str", "value")
	checkSavedKeyValueStore(r, pref.path, "{\n\t\"str\": \"value\"\n}")
}

func TestKeyValueStoreSetInt(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestEmptyKeyValueStore(r)
	defer clean()

	pref.SetInt("int", 42)
	checkSavedKeyValueStore(r, pref.path, "{\n\t\"int\": \"42\"\n}")
}

func TestKeyValueStoreSetBool(t *testing.T) {
	r := require.New(t)
	pref, clean := newTestEmptyKeyValueStore(r)
	defer clean()

	pref.SetBool("trueBool", true)
	pref.SetBool("falseBool", false)
	checkSavedKeyValueStore(r, pref.path, "{\n\t\"falseBool\": \"false\",\n\t\"trueBool\": \"true\"\n}")
}

func newTmpFile(r *require.Assertions) (path string, clean func()) {
	tmpfile, err := ioutil.TempFile("", "pref.*.json")
	r.NoError(err)
	defer r.NoError(tmpfile.Close())

	return tmpfile.Name(), func() {
		r.NoError(os.Remove(tmpfile.Name()))
	}
}

func newTestEmptyKeyValueStore(r *require.Assertions) (*keyValueStore, func()) {
	path, clean := newTmpFile(r)
	return newKeyValueStore(path), clean
}

func newTestKeyValueStore(r *require.Assertions) (*keyValueStore, func()) {
	path, clean := newTmpFile(r)
	r.NoError(ioutil.WriteFile(path, []byte("{\"str\":\"value\",\"int\":\"42\",\"bool\":\"true\",\"falseBool\":\"t\"}"), 0o700))
	return newKeyValueStore(path), clean
}

func checkSavedKeyValueStore(r *require.Assertions, path, expected string) {
	data, err := ioutil.ReadFile(path)
	r.NoError(err)
	r.Equal(expected, string(data))
}
