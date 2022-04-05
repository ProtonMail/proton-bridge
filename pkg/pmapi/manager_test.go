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

package pmapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

const testForceUpgradeBody = `{
	"Code":5003,
	"Error":"Upgrade!"
}`

const testTooManyAPIRequests = `{
	"Code":85131,
	"Error":"Too many recent API requests"
}`

func TestHandleTooManyRequests(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++

		if numCalls < 5 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("content-type", "application/json;charset=utf-8")
			fmt.Fprint(w, testTooManyAPIRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	m := New(Config{HostURL: ts.URL})

	m.SetRetryCount(5)

	// The call should succeed because the 5th retry should succeed (429s are retried).
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.NoError(t, err)

	// The server should be called 5 times.
	// The first four calls should return 429 and the last call should return 200.
	r.Equal(t, 5, numCalls)
}

func TestHandleUnprocessableEntity(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))

	m := New(Config{HostURL: ts.URL})

	m.SetRetryCount(5)

	// The call should fail because the first call should fail (422s are not retried).
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.EqualError(t, err, "422 Unprocessable Entity")
	// The server should be called 1 time.
	// The first call should return 422.
	r.Equal(t, 1, numCalls)
}

func TestHandleDialFailure(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusOK)
	}))

	// The failingRoundTripper will fail the first 5 times it is used.
	m := New(Config{HostURL: ts.URL})
	m.SetTransport(newFailingRoundTripper(5))
	m.SetRetryCount(5)

	// The call should succeed because the last retry should succeed (dial errors are retried).
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.NoError(t, err)

	// The server should be called 1 time.
	// The first 4 attempts don't reach the server.
	r.Equal(t, 1, numCalls)
}

func TestHandleTooManyDialFailures(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusOK)
	}))

	// The failingRoundTripper will fail the first 10 times it is used.
	// This is more than the number of retries we permit.
	// Thus, dials will fail.
	m := New(Config{HostURL: ts.URL})
	m.SetTransport(newFailingRoundTripper(10))
	m.SetRetryCount(5)

	// The call should fail because every dial will fail and we'll run out of retries.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.EqualError(t, err, "no internet connection")
	// The server should never be called.
	r.Equal(t, 0, numCalls)
}

func TestRetriesWithContextTimeout(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++

		if numCalls < 5 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Theoretically, this should succeed; on the fifth retry, we'll get StatusOK.
	m := New(Config{HostURL: ts.URL})
	m.SetRetryCount(5)

	// However, that will take ~0.5s, and we only allow 10ms in the context.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Thus, it will fail.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(ctx)
	r.EqualError(t, err, context.DeadlineExceeded.Error())
}

func TestObserveConnectionStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var onDown, onUp bool

	m := New(Config{HostURL: ts.URL})
	m.SetTransport(newFailingRoundTripper(10))
	m.SetRetryCount(5)
	m.AddConnectionObserver(NewConnectionObserver(func() { onDown = true }, func() { onUp = true }))

	// The call should fail because every dial will fail and we'll run out of retries.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.Error(t, err)
	r.False(t, onUp)
	r.True(t, onDown)

	onDown, onUp = false, false

	// The call should succeed because the last dial attempt will succeed.
	_, err = m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.NoError(t, err)
	r.True(t, onUp)
	r.False(t, onDown)
}

func TestReturnErrNoConnection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// We will fail more times than we retry, so requests should fail with ErrNoConnection.
	m := New(Config{HostURL: ts.URL})
	m.SetTransport(newFailingRoundTripper(10))
	m.SetRetryCount(5)

	// The call should fail because every dial will fail and we'll run out of retries.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.EqualError(t, err, "no internet connection")
}

func TestReturnErrUpgradeApplication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testForceUpgradeBody)
	}))

	m := New(Config{HostURL: ts.URL})

	// The call should fail because every call return force upgrade error.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	r.EqualError(t, err, ErrUpgradeApplication.Error())
}

type failingRoundTripper struct {
	http.RoundTripper

	fails, calls int
}

func newFailingRoundTripper(fails int) http.RoundTripper {
	return &failingRoundTripper{
		RoundTripper: http.DefaultTransport,
		fails:        fails,
	}
}

func (rt *failingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.calls++

	if rt.calls < rt.fails {
		return nil, errors.New("simulating network error")
	}

	return rt.RoundTripper.RoundTrip(req)
}
