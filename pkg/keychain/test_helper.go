// Copyright (c) 2024 Proton AG
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

package keychain

import (
	"sync"

	"github.com/docker/docker-credential-helpers/credentials"
)

type TestHelper map[string]*credentials.Credentials

func NewTestKeychainsList() *List {
	keychainHelper := NewTestHelper()
	helpers := make(Helpers)
	helpers["mock"] = func(string) (credentials.Helper, error) { return keychainHelper, nil }
	var list = List{helpers: helpers, defaultHelper: "mock", locker: &sync.Mutex{}}
	return &list
}

func NewTestHelper() TestHelper {
	return make(TestHelper)
}

func (h TestHelper) Add(creds *credentials.Credentials) error {
	h[creds.ServerURL] = creds
	return nil
}

func (h TestHelper) Delete(url string) error {
	delete(h, url)
	return nil
}

func (h TestHelper) Get(url string) (string, string, error) {
	creds := h[url]

	return creds.Username, creds.Secret, nil
}

func (h TestHelper) List() (map[string]string, error) {
	list := make(map[string]string)

	for url, creds := range h {
		list[url] = creds.Username
	}

	return list, nil
}
