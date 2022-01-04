// Copyright (c) 2022 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package settings

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testPrefFilePath = "/tmp/pref.json"

func TestLoadNoKeyValueStore(t *testing.T) {
	pref := newTestEmptyKeyValueStore(t)
	require.Equal(t, "", pref.Get("key"))
}

func TestLoadBadKeyValueStore(t *testing.T) {
	require.NoError(t, ioutil.WriteFile(testPrefFilePath, []byte("{\"key\":\"value"), 0700))
	pref := newKeyValueStore(testPrefFilePath)
	require.Equal(t, "", pref.Get("key"))
}

func TestKeyValueStoreGet(t *testing.T) {
	pref := newTestKeyValueStore(t)
	require.Equal(t, "value", pref.Get("str"))
	require.Equal(t, "42", pref.Get("int"))
	require.Equal(t, "true", pref.Get("bool"))
	require.Equal(t, "t", pref.Get("falseBool"))
}

func TestKeyValueStoreGetInt(t *testing.T) {
	pref := newTestKeyValueStore(t)
	require.Equal(t, 0, pref.GetInt("str"))
	require.Equal(t, 42, pref.GetInt("int"))
	require.Equal(t, 0, pref.GetInt("bool"))
	require.Equal(t, 0, pref.GetInt("falseBool"))
}

func TestKeyValueStoreGetBool(t *testing.T) {
	pref := newTestKeyValueStore(t)
	require.Equal(t, false, pref.GetBool("str"))
	require.Equal(t, false, pref.GetBool("int"))
	require.Equal(t, true, pref.GetBool("bool"))
	require.Equal(t, false, pref.GetBool("falseBool"))
}

func TestKeyValueStoreSetDefault(t *testing.T) {
	pref := newTestEmptyKeyValueStore(t)
	pref.setDefault("key", "value")
	pref.setDefault("key", "othervalue")
	require.Equal(t, "value", pref.Get("key"))
}

func TestKeyValueStoreSet(t *testing.T) {
	pref := newTestEmptyKeyValueStore(t)
	pref.Set("str", "value")
	checkSavedKeyValueStore(t, "{\n\t\"str\": \"value\"\n}")
}

func TestKeyValueStoreSetInt(t *testing.T) {
	pref := newTestEmptyKeyValueStore(t)
	pref.SetInt("int", 42)
	checkSavedKeyValueStore(t, "{\n\t\"int\": \"42\"\n}")
}

func TestKeyValueStoreSetBool(t *testing.T) {
	pref := newTestEmptyKeyValueStore(t)
	pref.SetBool("trueBool", true)
	pref.SetBool("falseBool", false)
	checkSavedKeyValueStore(t, "{\n\t\"falseBool\": \"false\",\n\t\"trueBool\": \"true\"\n}")
}

func newTestEmptyKeyValueStore(t *testing.T) *keyValueStore {
	require.NoError(t, os.RemoveAll(testPrefFilePath))
	return newKeyValueStore(testPrefFilePath)
}

func newTestKeyValueStore(t *testing.T) *keyValueStore {
	require.NoError(t, ioutil.WriteFile(testPrefFilePath, []byte("{\"str\":\"value\",\"int\":\"42\",\"bool\":\"true\",\"falseBool\":\"t\"}"), 0700))
	return newKeyValueStore(testPrefFilePath)
}

func checkSavedKeyValueStore(t *testing.T, expected string) {
	data, err := ioutil.ReadFile(testPrefFilePath)
	require.NoError(t, err)
	require.Equal(t, expected, string(data))
}
