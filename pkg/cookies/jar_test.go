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

	jar, err := New(NewPersister(make(testPersister)))
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

type testPersister map[string]string

func (p testPersister) Set(key, value string) {
	p[key] = value
}

func (p testPersister) Get(key string) string {
	return p[key]
}
