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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJarGetSet(t *testing.T) {
	ts := getTestServer(t, []testCookie{
		{"TestName1", "TestValue1", 3600},
		{"TestName2", "TestValue2", 3600},
		{"TestName3", "TestValue3", 3600},
	})
	defer ts.Close()

	client := getClientWithJar(t, make(testGetterSetter))

	// Hit a server that sets some cookies.
	setRes, err := client.Get(ts.URL + "/set")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, setRes.Body.Close())

	// Hit a server that checks the cookies are there.
	getRes, err := client.Get(ts.URL + "/get")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, getRes.Body.Close())
}

func TestJarLoad(t *testing.T) {
	ts := getTestServer(t, []testCookie{
		{"TestName1", "TestValue1", 3600},
		{"TestName2", "TestValue2", 3600},
		{"TestName3", "TestValue3", 3600},
	})
	defer ts.Close()

	// This will be our "persistent storage" from which the cookie jar should load cookies.
	gs := make(testGetterSetter)

	// This client saves cookies to persistent storage.
	oldClient := getClientWithJar(t, gs)

	// Hit a server that sets some cookies.
	setRes, err := oldClient.Get(ts.URL + "/set")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, setRes.Body.Close())

	// This client loads cookies from persistent storage.
	newClient := getClientWithJar(t, gs)

	// Hit a server that checks the cookies are there.
	getRes, err := newClient.Get(ts.URL + "/get")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, getRes.Body.Close())
}

func TestJarExpiry(t *testing.T) {
	ts := getTestServer(t, []testCookie{
		{"TestName1", "TestValue1", 3600},
		{"TestName2", "TestValue2", 1},
		{"TestName3", "TestValue3", 3600},
	})
	defer ts.Close()

	// This will be our "persistent storage" from which the cookie jar should load cookies.
	gs := make(testGetterSetter)

	// This client saves cookies to persistent storage.
	oldClient := getClientWithJar(t, gs)

	// Hit a server that sets some cookies.
	setRes, err := oldClient.Get(ts.URL + "/set")
	if err != nil {
		t.FailNow()
	}
	require.NoError(t, setRes.Body.Close())

	// Wait until the second cookie expires.
	time.Sleep(2 * time.Second)

	// Load a client, which will clear out expired cookies.
	_ = getClientWithJar(t, gs)

	assert.Contains(t, gs["cookies"], "TestName1")
	assert.NotContains(t, gs["cookies"], "TestName2")
	assert.Contains(t, gs["cookies"], "TestName3")
}

type testCookie struct {
	name, value string
	maxAge      int
}

func getClientWithJar(t *testing.T, gs GetterSetter) *http.Client {
	jar, err := NewCookieJar(gs)
	require.NoError(t, err)

	return &http.Client{Jar: jar}
}

func getTestServer(t *testing.T, wantCookies []testCookie) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/set", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, cookie := range wantCookies {
			http.SetCookie(w, &http.Cookie{
				Name:   cookie.name,
				Value:  cookie.value,
				MaxAge: cookie.maxAge,
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
