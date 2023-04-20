package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/cucumber/godog"
	"github.com/sirupsen/logrus"
)

func (s *scenario) bridgeEventuallySendsTheFollowingHeartbeat(text *godog.DocString) error {
	return eventually(func() error {
		err := s.bridgeSendsTheFollowingHeartbeat(text)
		logrus.WithError(err).Trace("Matching eventually")
		return err
	})
}

func (s *scenario) bridgeSendsTheFollowingHeartbeat(text *godog.DocString) error {
	var wantHeartbeat telemetry.HeartbeatData
	err := json.Unmarshal([]byte(text.Content), &wantHeartbeat)
	if err != nil {
		return err
	}

	return matchHeartbeat(s.t.heartbeat.heartbeat, wantHeartbeat)
}

func (s *scenario) bridgeNeedsToSendHeartbeat() error {
	last := s.t.heartbeat.GetLastHeartbeatSent()
	if !isAnotherDay(last, time.Now()) {
		return fmt.Errorf("heartbeat already sent at %s", last)
	}
	return nil
}

func (s *scenario) bridgeDoNotNeedToSendHeartbeat() error {
	last := s.t.heartbeat.GetLastHeartbeatSent()
	if isAnotherDay(last, time.Now()) {
		return fmt.Errorf("heartbeat needs to be sent - last %s", last)
	}
	return nil
}

func (s *scenario) heartbeatIsNotwhitelisted() error {
	s.t.heartbeat.rejectSend()
	return nil
}

func matchHeartbeat(have, want telemetry.HeartbeatData) error {
	if have == (telemetry.HeartbeatData{}) {
		return errors.New("no heartbeat send (yet)")
	}

	// Ignore rollout number
	want.Dimensions.Rollout = have.Dimensions.Rollout

	if have != want {
		return fmt.Errorf("missing heartbeat: have %#v, want %#v", have, want)
	}

	return nil
}

func isAnotherDay(last, now time.Time) bool {
	return now.Year() > last.Year() || (now.Year() == last.Year() && now.YearDay() > last.YearDay())
}
