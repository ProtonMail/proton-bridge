// Copyright (c) 2023 Proton AG
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
	"encoding/json"
	"fmt"
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
	return nil
}

func (s *scenario) internetIsTurnedOn() error {
	s.t.netCtl.SetCanDial(true)
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

	if haveKey := call.RequestHeader.Get(key); haveKey != value {
		return fmt.Errorf("have header %q, want %q", haveKey, value)
	}

	return nil
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
