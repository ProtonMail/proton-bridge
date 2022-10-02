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
	s.t.dialer.SetCanDial(false)
	return nil
}

func (s *scenario) internetIsTurnedOn() error {
	s.t.dialer.SetCanDial(true)
	return nil
}

func (s *scenario) theUserAgentIs(userAgent string) error {
	if haveUserAgent := s.t.bridge.GetCurrentUserAgent(); haveUserAgent != userAgent {
		return fmt.Errorf("have user agent %q, want %q", haveUserAgent, userAgent)
	}

	return nil
}

func (s *scenario) theHeaderInTheRequestToHasSetTo(method, path, key, value string) error {
	call, err := s.t.getLastCall(method, path)
	if err != nil {
		return err
	}

	if haveKey := call.Header.Get(key); haveKey != value {
		return fmt.Errorf("have header %q, want %q", haveKey, value)
	}

	return nil
}

func (s *scenario) theBodyInTheRequestToIs(method, path string, value *godog.DocString) error {
	call, err := s.t.getLastCall(method, path)
	if err != nil {
		return err
	}

	var body, want map[string]any

	if err := json.Unmarshal(call.Body, &body); err != nil {
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
