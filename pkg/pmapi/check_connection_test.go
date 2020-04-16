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
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/dialer"
	"github.com/stretchr/testify/require"
)

const testServerPort = "18000"
const testRequestTimeout = 10 * time.Second

func TestMain(m *testing.M) {
	go startServer()
	time.Sleep(100 * time.Millisecond) // We need to wait till server is fully running.
	code := m.Run()
	os.Exit(code)
}

func startServer() {
	http.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	http.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	http.HandleFunc("/serverError", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})
	panic(http.ListenAndServe(":"+testServerPort, nil))
}

func TestCheckConnection(t *testing.T) {
	checkCheckConnection(t, "ok", "")
}

func TestCheckConnectionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	checkCheckConnection(t, "timeout", "Client.Timeout exceeded while awaiting headers")
}

func TestCheckConnectionServerError(t *testing.T) {
	checkCheckConnection(t, "serverError", "HTTP status code 500")
}

func checkCheckConnection(t *testing.T, path string, expectedErrMessage string) {
	client := dialer.DialTimeoutClient()
	client.Timeout = testRequestTimeout

	ch := make(chan error)

	go checkConnection(client, "http://localhost:"+testServerPort+"/"+path, ch)

	timeout := time.After(testRequestTimeout + time.Second)
	select {
	case err := <-ch:
		if expectedErrMessage == "" {
			require.NoError(t, err)
		} else {
			require.Error(t, err, expectedErrMessage)
		}
	case <-timeout:
		t.Error("checkConnection timeout failed")
	}
}
