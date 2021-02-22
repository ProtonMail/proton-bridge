package pmapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func TestHandleTooManyRequests(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++

		if numCalls < 5 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))

	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// The call should succeed because the 5th retry should succeed (429s are retried).
	if _, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background()); err != nil {
		t.Fatal("got unexpected error", err)
	}

	// The server should be called 5 times.
	// The first four calls should return 429 and the last call should return 200.
	if numCalls != 5 {
		t.Fatal("expected numCalls to be 5, instead got", numCalls)
	}
}

func TestHandleUnprocessableEntity(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))

	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// The call should fail because the first call should fail (422s are not retried).
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	if err == nil {
		t.Fatal("expected error, instead got", err)
	}

	// API-side errors get ErrAPIFailure
	if !errors.Is(err, pmapi.ErrAPIFailure) {
		t.Fatal("expected error to be ErrAPIFailure, instead got", err)
	}

	// The server should be called 1 time.
	// The first call should return 422.
	if numCalls != 1 {
		t.Fatal("expected numCalls to be 1, instead got", numCalls)
	}
}

func TestHandleDialFailure(t *testing.T) {
	var numCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusOK)
	}))

	// The failingRoundTripper will fail the first 5 times it is used.
	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set a custom transport.
	m.SetTransport(newFailingRoundTripper(5))

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// The call should succeed because the last retry should succeed (dial errors are retried).
	if _, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background()); err != nil {
		t.Fatal("got unexpected error", err)
	}

	// The server should be called 1 time.
	// The first 4 attempts don't reach the server.
	if numCalls != 1 {
		t.Fatal("expected numCalls to be 1, instead got", numCalls)
	}
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
	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set a custom transport.
	m.SetTransport(newFailingRoundTripper(10))

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// The call should fail because every dial will fail and we'll run out of retries.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	if err == nil {
		t.Fatal("expected error, instead got", err)
	}

	if !errors.Is(err, pmapi.ErrNoConnection) {
		t.Fatal("expected error to be ErrNoConnection, instead got", err)
	}

	// The server should never be called.
	if numCalls != 0 {
		t.Fatal("expected numCalls to be 0, instead got", numCalls)
	}
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
	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// However, that will take ~5s, and we only allow 1s in the context.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Thus, it will fail.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(ctx)
	if err == nil {
		t.Fatal("expected error, instead got", err)
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expected error to be DeadlineExceeded, instead got", err)
	}
}

func TestObserveConnectionStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var onDown, onUp bool

	m := pmapi.New(pmapi.Config{HostURL: ts.URL})

	// Set a custom transport.
	m.SetTransport(newFailingRoundTripper(10))

	// Set the retry count to 5.
	m.SetRetryCount(5)

	// Add a connection observer.
	m.AddConnectionObserver(pmapi.NewConnectionObserver(func() { onDown = true }, func() { onUp = true }))

	// The call should fail because every dial will fail and we'll run out of retries.
	if _, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background()); err == nil {
		t.Fatal("expected error, instead got", err)
	}

	if onDown != true || onUp == true {
		t.Fatal("expected onDown to have been called and onUp to not have been called")
	}

	onDown, onUp = false, false

	// The call should succeed because the last dial attempt will succeed.
	if _, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background()); err != nil {
		t.Fatal("got unexpected error", err)
	}

	if onDown == true || onUp != true {
		t.Fatal("expected onUp to have been called and onDown to not have been called")
	}
}

func TestReturnErrNoConnection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// We will fail more times than we retry, so requests should fail with ErrNoConnection.
	m := pmapi.New(pmapi.Config{HostURL: ts.URL})
	m.SetTransport(newFailingRoundTripper(10))
	m.SetRetryCount(5)

	// The call should fail because every dial will fail and we'll run out of retries.
	_, err := m.NewClient("", "", "", time.Now().Add(time.Hour)).GetAddresses(context.Background())
	if err == nil {
		t.Fatal("expected error, instead got", err)
	}

	if !errors.Is(err, pmapi.ErrNoConnection) {
		t.Fatal("expected error to be ErrNoConnection, instead got", err)
	}
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
