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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

func (s *scenario) itSucceeds() error {
	if err := s.t.getLastError(); err != nil {
		return fmt.Errorf("expected nil, got error %v", err)
	}

	return nil
}

func (s *scenario) itFails() error {
	if err := s.t.getLastError(); err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	return nil
}

func (s *scenario) itFailsWithError(wantErr string) error {
	err := s.t.getLastError()
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}

	if haveErr := err.Error(); !strings.Contains(haveErr, wantErr) {
		return fmt.Errorf("expected error %q, got %q", wantErr, haveErr)
	}

	return nil
}

func (s *scenario) internetIsTurnedOff() error {
	s.t.netCtl.SetCanDial(false)
	t, ok := (*s.t.rt).(*http.Transport)
	if ok {
		t.CloseIdleConnections()
	}
	return nil
}

func (s *scenario) internetIsTurnedOn() error {
	s.t.netCtl.SetCanDial(true)
	t, ok := (*s.t.rt).(*http.Transport)
	if ok {
		t.CloseIdleConnections()
	}
	return nil
}

func (s *scenario) theUserAgentIs(userAgent string) error {
	return eventually(func() error {
		if haveUserAgent := s.t.bridge.GetCurrentUserAgent(); haveUserAgent != userAgent {
			return fmt.Errorf("have user agent %q, want %q", haveUserAgent, userAgent)
		}

		return nil
	})
}

func (s *scenario) theHeaderInTheRequestToHasSetTo(method, path, key, value string) error {
	call, err := s.t.getLastCall(method, path)
	if err != nil {
		return err
	}

	if haveValue := call.RequestHeader.Get(key); haveValue != value {
		return fmt.Errorf("header field %q have %q, want %q", key, haveValue, value)
	}

	return nil
}

func (s *scenario) theHeaderInTheMultipartRequestToHasSetTo(method, path, key, value string) error {
	req, err := s.getLastCallMultipartForm(method, path)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}
	if haveValue := req.FormValue(key); haveValue != value {
		return fmt.Errorf("header field %q have %q, want %q", key, haveValue, value)
	}
	return nil
}

func (s *scenario) checkParsedMultipartFormForFile(method, path, file string, hasFile bool) error {
	req, err := s.getLastCallMultipartForm(method, path)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	if _, ok := req.MultipartForm.File[file]; hasFile != ok {
		return fmt.Errorf("Multipart file in bug report is %t, want it to be %t", ok, hasFile)
	}

	return nil
}

func (s *scenario) theHeaderInTheMultipartRequestToHasFile(method, path, file string) error {
	return s.checkParsedMultipartFormForFile(method, path, file, true)
}

func (s *scenario) theHeaderInTheMultipartRequestToHasNoFile(method, path, file string) error {
	return s.checkParsedMultipartFormForFile(method, path, file, false)
}

func (s *scenario) getLastCallMultipartForm(method, path string) (*http.Request, error) {
	// We have to exclude HTTP-Overrides to avoid race condition with the creating and sending of the draft message.
	call, err := s.t.getLastCallExcludingHTTPOverride(method, path)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if _, err := buf.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", call.Method, call.URL.Path)); err != nil {
		return nil, fmt.Errorf("failed to write request line: %w", err)
	}

	if err := call.RequestHeader.Write(buf); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := buf.WriteString("\r\n"); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := buf.Write(call.RequestBody); err != nil {
		return nil, fmt.Errorf("failed to write body: %w", err)
	}

	req, err := http.ReadRequest(bufio.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	if err := req.ParseMultipartForm(1 << 10); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}
	return req, nil
}

func (s *scenario) theBodyInTheRequestToIs(method, path string, value *godog.DocString) error {
	// We have to exclude HTTP-Overrides to avoid race condition with the creating and sending of the draft message.
	call, err := s.t.getLastCallExcludingHTTPOverride(method, path)
	if err != nil {
		return err
	}

	var body, want map[string]any

	if err := json.Unmarshal(call.RequestBody, &body); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value.Content), &want); err != nil {
		return err
	}

	if !IsSub(body, want) {
		return fmt.Errorf("have body %v, want %v", body, want)
	}

	return nil
}

func (s *scenario) theBodyInTheResponseToIs(method, path string, value *godog.DocString) error {
	// We have to exclude HTTP-Overrides to avoid race condition with the creating and sending of the draft message.
	call, err := s.t.getLastCallExcludingHTTPOverride(method, path)
	if err != nil {
		return err
	}

	var body, want map[string]any

	if err := json.Unmarshal(call.ResponseBody, &body); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value.Content), &want); err != nil {
		return err
	}

	if !IsSub(body, want) {
		return fmt.Errorf("have body %v, want %v", body, want)
	}

	return nil
}

func (s *scenario) theMessageUsedKeyForSending(address string) error {
	addrID := s.t.getUserByAddress(address).getAddrID(address)

	call, err := s.t.getLastCallExcludingHTTPOverride("POST", "/mail/v4/messages")
	if err != nil {
		return err
	}

	var body, want map[string]any

	if err := json.Unmarshal(call.ResponseBody, &body); err != nil {
		return err
	}

	want = map[string]any{
		"Message": map[string]any{
			"AddressID": addrID,
		},
	}

	if !IsSub(body, want) {
		return fmt.Errorf("have body %v, want %v", body, want)
	}

	return nil
}
