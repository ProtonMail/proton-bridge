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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	r "github.com/stretchr/testify/require"
)

var (
	colRed     = "\033[1;31m"
	colNon     = "\033[0;39m"
	reHTTPCode = regexp.MustCompile(`(HTTP|get|post|put|delete)_(\d{3}).*.json`)
)

func newTestConfig(url string) Config {
	return Config{
		HostURL:    url,
		AppVersion: "GoPMAPI_1.0.14",
	}
}

// newTestClient is old function and should be replaced everywhere by newTestClientCallbacks.
func newTestClient(h http.Handler) (*httptest.Server, Client) {
	s := httptest.NewServer(h)

	return s, newManager(newTestConfig(s.URL)).NewClient(testUID, testAccessToken, testRefreshToken, time.Now().Add(time.Hour))
}

func newTestClientCallbacks(tb testing.TB, callbacks ...func(testing.TB, http.ResponseWriter, *http.Request) string) (func(), Client) {
	reqNum := 0
	_, file, line, _ := runtime.Caller(1)
	file = filepath.Base(file)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNum++
		if reqNum > len(callbacks) {
			fmt.Printf(
				"%s:%d: %sServer was requested %d times which is more requests than expected %d times%s\n\n",
				file, line, colRed, reqNum, len(callbacks), colNon,
			)
			tb.FailNow()
		}
		response := callbacks[reqNum-1](tb, w, r)
		if response != "" {
			writeJSONResponsefromFile(tb, w, response, reqNum-1)
		}
	}))

	finish := func() {
		server.CloseClientConnections() // Closing without waiting for finishing requests.
		if reqNum != len(callbacks) {
			fmt.Printf(
				"%s:%d: %sServer was requested %d times but expected to be %d times%s\n\n",
				file, line, colRed, reqNum, len(callbacks), colNon,
			)
			tb.Error("server failed")
		}
	}

	return finish, newManager(newTestConfig(server.URL)).NewClient(testUID, testAccessToken, testRefreshToken, time.Now().Add(time.Hour))
}

func checkMethodAndPath(r *http.Request, method, path string) error {
	var result *multierror.Error
	if err := checkHeader(r.Header, "x-pm-appversion", "GoPMAPI_1.0.14"); err != nil {
		result = multierror.Append(result, err)
	}
	if r.Method != method {
		err := fmt.Errorf("Invalid request method expected %v, got %v", method, r.Method)
		result = multierror.Append(result, err)
	}
	if r.URL.RequestURI() != path {
		err := fmt.Errorf("Invalid request path expected %v, got %v", path, r.URL.RequestURI())
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

func writeJSONResponsefromFile(tb testing.TB, w http.ResponseWriter, response string, reqNum int) {
	if match := reHTTPCode.FindAllSubmatch([]byte(response), -1); len(match) != 0 {
		httpCode, err := strconv.Atoi(string(match[0][len(match[0])-1]))
		r.NoError(tb, err)
		w.WriteHeader(httpCode)
	}
	f, err := os.Open("./testdata/routes/" + response)
	r.NoError(tb, err)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	w.Header().Set("x-test-pmapi-response", fmt.Sprintf("%s:%d", tb.Name(), reqNum))
	_, err = io.Copy(w, f)
	r.NoError(tb, err)
}

func checkHeader(h http.Header, field, exp string) error {
	val := h.Get(field)
	if val != exp {
		msg := "wrong field %s expected %q but have %q"
		return fmt.Errorf(msg, field, exp, val)
	}
	return nil
}

func isAuthReq(r *http.Request, uid, token string) error { //nolint:unparam   always retrieves testUID
	if err := checkHeader(r.Header, "x-pm-uid", uid); err != nil {
		return err
	}
	if err := checkHeader(r.Header, "authorization", "Bearer "+token); err != nil { //nolint:revive   can return the error right away but this is easier to read
		return err
	}
	return nil
}
