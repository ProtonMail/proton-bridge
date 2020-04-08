// Copyright (c) 2020 Proton Technologies AG
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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testPrefFilePath = "/tmp/pref.json"

func shutdownTestPreferences() {
	_ = os.RemoveAll(testPrefFilePath)
}

func TestLoadNoPreferences(t *testing.T) {
	pref := newTestEmptyPreferences(t)
	require.Equal(t, "", pref.Get("key"))
}

func TestLoadBadPreferences(t *testing.T) {
	require.NoError(t, ioutil.WriteFile(testPrefFilePath, []byte("{\"key\":\"value"), 0700))
	pref := NewPreferences(testPrefFilePath)
	require.Equal(t, "", pref.Get("key"))
}

func TestPreferencesGet(t *testing.T) {
	pref := newTestPreferences(t)
	require.Equal(t, "value", pref.Get("str"))
	require.Equal(t, "42", pref.Get("int"))
	require.Equal(t, "true", pref.Get("bool"))
	require.Equal(t, "t", pref.Get("falseBool"))
}

func TestPreferencesGetInt(t *testing.T) {
	pref := newTestPreferences(t)
	require.Equal(t, 0, pref.GetInt("str"))
	require.Equal(t, 42, pref.GetInt("int"))
	require.Equal(t, 0, pref.GetInt("bool"))
	require.Equal(t, 0, pref.GetInt("falseBool"))
}

func TestPreferencesGetBool(t *testing.T) {
	pref := newTestPreferences(t)
	require.Equal(t, false, pref.GetBool("str"))
	require.Equal(t, false, pref.GetBool("int"))
	require.Equal(t, true, pref.GetBool("bool"))
	require.Equal(t, false, pref.GetBool("falseBool"))
}

func TestPreferencesSetDefault(t *testing.T) {
	pref := newTestEmptyPreferences(t)
	pref.SetDefault("key", "value")
	pref.SetDefault("key", "othervalue")
	require.Equal(t, "value", pref.Get("key"))
}

func TestPreferencesSet(t *testing.T) {
	pref := newTestEmptyPreferences(t)
	pref.Set("str", "value")
	checkSavedPreferences(t, "{\"str\":\"value\"}")
}

func TestPreferencesSetInt(t *testing.T) {
	pref := newTestEmptyPreferences(t)
	pref.SetInt("int", 42)
	checkSavedPreferences(t, "{\"int\":\"42\"}")
}

func TestPreferencesSetBool(t *testing.T) {
	pref := newTestEmptyPreferences(t)
	pref.SetBool("trueBool", true)
	pref.SetBool("falseBool", false)
	checkSavedPreferences(t, "{\"falseBool\":\"false\",\"trueBool\":\"true\"}")
}

func newTestEmptyPreferences(t *testing.T) *Preferences {
	require.NoError(t, os.RemoveAll(testPrefFilePath))
	return NewPreferences(testPrefFilePath)
}

func newTestPreferences(t *testing.T) *Preferences {
	require.NoError(t, ioutil.WriteFile(testPrefFilePath, []byte("{\"str\":\"value\",\"int\":\"42\",\"bool\":\"true\",\"falseBool\":\"t\"}"), 0700))
	return NewPreferences(testPrefFilePath)
}

func checkSavedPreferences(t *testing.T, expected string) {
	data, err := ioutil.ReadFile(testPrefFilePath)
	require.NoError(t, err)
	require.Equal(t, expected+"\n", string(data))
}
