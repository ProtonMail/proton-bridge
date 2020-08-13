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

package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJar(t *testing.T) {
	testCookies := []testCookie{
		{"TestName1", "TestValue1"},
		{"TestName2", "TestValue2"},
		{"TestName3", "TestValue3"},
	}

	ts := getTestServer(t, testCookies...)
	defer ts.Close()

	jar, err := NewCookieJar(make(testGetterSetter))
	require.NoError(t, err)

	client := &http.Client{Jar: jar}

	setRes, err := client.Get(ts.URL + "/set")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, setRes.Body.Close())

	getRes, err := client.Get(ts.URL + "/get")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, getRes.Body.Close())
}

type testCookie struct {
	name, value string
}

func getTestServer(t *testing.T, wantCookies ...testCookie) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/set", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, cookie := range wantCookies {
			http.SetCookie(w, &http.Cookie{
				Name:  cookie.name,
				Value: cookie.value,
			})
		}

		w.WriteHeader(http.StatusOK)
	}))

	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Len(t, r.Cookies(), len(wantCookies))

		for k, v := range r.Cookies() {
			assert.Equal(t, wantCookies[k].name, v.Name)
			assert.Equal(t, wantCookies[k].value, v.Value)
		}

		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(mux)
}

type testGetterSetter map[string]string

func (p testGetterSetter) Set(key, value string) {
	p[key] = value
}

func (p testGetterSetter) Get(key string) string {
	return p[key]
}
