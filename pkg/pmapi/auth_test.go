package pmapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func TestAutomaticAuthRefresh(t *testing.T) {
	var wantAuth = &pmapi.Auth{
		UID:          "testUID",
		AccessToken:  "testAcc",
		RefreshToken: "testRef",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(wantAuth); err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(mux)

	var gotAuth *pmapi.Auth

	// Create a new client.
	c := pmapi.New(pmapi.Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(-time.Second))

	// Register an auth handler.
	c.AddAuthHandler(func(auth *pmapi.Auth) error { gotAuth = auth; return nil })

	// Make a request with an access token that already expired one second ago.
	if _, err := c.GetAddresses(context.Background()); err != nil {
		t.Fatal("got unexpected error", err)
	}

	// The auth callback should have been called.
	if *gotAuth != *wantAuth {
		t.Fatal("got unexpected auth", gotAuth)
	}
}

func Test401AuthRefresh(t *testing.T) {
	var wantAuth = &pmapi.Auth{
		UID:          "testUID",
		AccessToken:  "testAcc",
		RefreshToken: "testRef",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(wantAuth); err != nil {
			panic(err)
		}
	})

	var call int

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		call++

		if call == 1 {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	ts := httptest.NewServer(mux)

	var gotAuth *pmapi.Auth

	// Create a new client.
	c := pmapi.New(pmapi.Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(time.Hour))

	// Register an auth handler.
	c.AddAuthHandler(func(auth *pmapi.Auth) error { gotAuth = auth; return nil })

	// The first request will fail with 401, triggering a refresh and retry.
	if _, err := c.GetAddresses(context.Background()); err != nil {
		t.Fatal("got unexpected error", err)
	}

	// The auth callback should have been called.
	if *gotAuth != *wantAuth {
		t.Fatal("got unexpected auth", gotAuth)
	}
}

func Test401RevokedAuth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	ts := httptest.NewServer(mux)

	c := pmapi.New(pmapi.Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh.
	// The retry will also fail with 401, returning an error.
	_, err := c.GetAddresses(context.Background())
	if err == nil {
		t.Fatal("expected error, instead got", err)
	}

	if !errors.Is(err, pmapi.ErrUnauthorized) {
		t.Fatal("expected error to be ErrUnauthorized, instead got", err)
	}
}
