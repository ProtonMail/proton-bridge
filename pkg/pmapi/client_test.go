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

package pmapi

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testClientConfig = &ClientConfig{
	AppVersion:       "GoPMAPI_1.0.14",
	ClientID:         "demoapp",
	FirstReadTimeout: 500 * time.Millisecond,
	MinSpeed:         256,
}

func newTestClient(cm *ClientManager) *Client {
	return cm.GetClient("tester")
}

func TestClient_Do(t *testing.T) {
	const testResBody = "Hello World!"

	var receivedReq *http.Request
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedReq = r
		fmt.Fprint(w, testResBody)
	}))
	defer s.Close()

	req, err := c.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal("Expected no error while creating request, got:", err)
	}

	res, err := c.Do(req, true)
	if err != nil {
		t.Fatal("Expected no error while executing request, got:", err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Expected no error while reading response, got:", err)
	}
	require.Nil(t, res.Body.Close())

	if string(b) != testResBody {
		t.Fatalf("Invalid response body: expected %v, got %v", testResBody, string(b))
	}

	h := receivedReq.Header
	if h.Get("x-pm-appversion") != testClientConfig.AppVersion {
		t.Fatalf("Invalid app version header: expected %v, got %v", testClientConfig.AppVersion, h.Get("x-pm-appversion"))
	}
	if h.Get("x-pm-apiversion") != fmt.Sprintf("%v", Version) {
		t.Fatalf("Invalid api version header: expected %v, got %v", Version, h.Get("x-pm-apiversion"))
	}
	if h.Get("x-pm-uid") != "" {
		t.Fatalf("Expected no uid header when not authenticated, got %v", h.Get("x-pm-uid"))
	}
	if h.Get("Authorization") != "" {
		t.Fatalf("Expected no authentication header when not authenticated, got %v", h.Get("Authorization"))
	}
}

func TestClient_DoRetryAfter(t *testing.T) {
	testStart := time.Now()
	secondAttemptTime := time.Now()

	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			w.Header().Set("content-type", "application/json;charset=utf-8")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return ""
		},
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			w.Header().Set("content-type", "application/json;charset=utf-8")
			w.WriteHeader(http.StatusOK)
			secondAttemptTime = time.Now()
			return "/HTTP_200.json"
		},
	)
	defer finish()

	require.Nil(t, c.SendSimpleMetric("some_category", "some_action", "some_label"))
	waitedTime := secondAttemptTime.Sub(testStart)
	isInRange := 1*time.Second < waitedTime && waitedTime <= 11*time.Second
	require.True(t, isInRange, "Waited time: %v", waitedTime)
}

type slowTransport struct {
	transport      http.RoundTripper
	firstBodySleep time.Duration
}

func (t *slowTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if err == nil {
		resp.Body = &slowReadCloser{
			req:            req,
			readCloser:     resp.Body,
			firstBodySleep: t.firstBodySleep,
		}
	}
	return resp, err
}

type slowReadCloser struct {
	req            *http.Request
	readCloser     io.ReadCloser
	firstBodySleep time.Duration
}

func (r *slowReadCloser) Read(p []byte) (n int, err error) {
	// Normally timeout is processed by Read function.
	// It's hard to test slow connection; we need to manually
	// check when context is Done, because otherwise timeout
	// happens only during failed Read which will not happen
	// in this artificial environment.
	select {
	case <-r.req.Context().Done():
		return 0, context.Canceled
	case <-time.After(r.firstBodySleep):
	}
	return r.readCloser.Read(p)
}

func (r *slowReadCloser) Close() error {
	return r.readCloser.Close()
}

func TestClient_FirstReadTimeout(t *testing.T) {
	requestTimeout := testClientConfig.FirstReadTimeout + 1*time.Second

	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			return "/HTTP_200.json"
		},
	)
	defer finish()

	c.hc.Transport = &slowTransport{
		transport:      c.hc.Transport,
		firstBodySleep: requestTimeout,
	}

	started := time.Now()
	err := c.SendSimpleMetric("some_category", "some_action", "some_label")
	require.Error(t, err, "cannot reach the server")
	require.True(t, time.Since(started) < requestTimeout, "Actual waited time: %v", time.Since(started))
}

func TestClient_MinSpeedTimeout(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeSlow(2*time.Second),
	)
	defer finish()

	err := c.SendSimpleMetric("some_category", "some_action", "some_label")
	require.Error(t, err, "cannot reach the server")
}

func TestClient_MinSpeedNoTimeout(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeSlow(500*time.Millisecond),
	)
	defer finish()

	err := c.SendSimpleMetric("some_category", "some_action", "some_label")
	require.Nil(t, err)
}

func routeSlow(delay time.Duration) func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
	return func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("{\"code\":1000,\"key\":\""))
		for chunk := 1; chunk <= 10; chunk++ {
			// We need to write enough bytes which enforce flushing data
			// because writer used by httptest does not implement Flusher.
			for i := 1; i <= 10000; i++ {
				_, _ = w.Write([]byte("a"))
			}
			time.Sleep(delay)
		}
		_, _ = w.Write([]byte("\"}"))
		return ""
	}
}
