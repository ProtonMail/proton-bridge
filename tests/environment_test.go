package tests

import (
	"fmt"
	"strings"
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

func (s *scenario) theValueOfTheHeaderInTheRequestToIs(key, path, value string) error {
	call, err := s.t.getLastCall(path)
	if err != nil {
		return err
	}

	if haveKey := call.Request.Header.Get(key); haveKey != value {
		return fmt.Errorf("have header %q, want %q", haveKey, value)
	}

	return nil
}
