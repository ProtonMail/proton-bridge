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
	"os"
	"path/filepath"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
	"github.com/sirupsen/logrus"
)

func (s *scenario) configStatusFileExistForUser(username string) error {
	configStatusFile, err := getConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}
	if _, err := os.Stat(configStatusFile); err != nil {
		return err
	}
	return nil
}

func (s *scenario) configStatusIsPendingForUser(username string) error {
	data, err := loadConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}

	if data.DataV1.PendingSince.IsZero() {
		return fmt.Errorf("expected ConfigStatus pending but got success instead")
	}

	return nil
}

func (s *scenario) configStatusIsPendingWithFailureForUser(username string) error {
	data, err := loadConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}

	if data.DataV1.PendingSince.IsZero() {
		return fmt.Errorf("expected ConfigStatus pending but got success instead")
	}
	if data.DataV1.FailureDetails == "" {
		return fmt.Errorf("expected ConfigStatus pending with failure but got no failure instead")
	}

	return nil
}

func (s *scenario) configStatusSucceedForUser(username string) error {
	data, err := loadConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}

	if !data.DataV1.PendingSince.IsZero() {
		return fmt.Errorf("expected ConfigStatus success but got pending since %s", data.DataV1.PendingSince)
	}

	return nil
}

func (s *scenario) configStatusEventIsEventuallySendXTime(event string, number int) error {
	return eventually(func() error {
		err := s.checkEventSentForUser(event, number)
		logrus.WithError(err).Trace("Matching eventually")
		return err
	})
}

func (s *scenario) configStatusEventIsNotSendMoreThanXTime(event string, number int) error {
	if err := eventually(func() error {
		err := s.checkEventSentForUser(event, number+1)
		logrus.WithError(err).Trace("Matching eventually")
		return err
	}); err == nil {
		return fmt.Errorf("expected %s to be sent %d but catch %d", event, number, number+1)
	}
	return nil
}

func (s *scenario) forceConfigStatusProgressToBeSentForUser(username string) error {
	configStatusFile, err := getConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}

	data, err := loadConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}
	data.DataV1.PendingSince = time.Now().AddDate(0, 0, -2)
	data.DataV1.LastProgress = time.Now().AddDate(0, 0, -1)

	f, err := os.Create(configStatusFile)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(data)
}

func (s *scenario) checkEventSentForUser(event string, number int) error {
	calls, err := getLastTelemetryEventSent(s.t, event)
	if err != nil {
		return err
	}
	if len(calls) != number {
		return fmt.Errorf("expected %s to be sent %d but catch %d", event, number, len(calls))
	}
	return nil
}

func getConfigStatusFile(t *testCtx, username string) (string, error) {
	userID := t.getUserByName(username).getUserID()
	statsDir, err := t.locator.ProvideStatsPath()
	if err != nil {
		return "", fmt.Errorf("failed to get Statistics directory: %w", err)
	}
	return filepath.Join(statsDir, userID+".json"), nil
}

func loadConfigStatusFile(t *testCtx, username string) (configstatus.ConfigurationStatusData, error) {
	data := configstatus.ConfigurationStatusData{}

	configStatusFile, err := getConfigStatusFile(t, username)
	if err != nil {
		return data, err
	}

	if _, err := os.Stat(configStatusFile); err != nil {
		return data, err
	}

	f, err := os.Open(configStatusFile)
	if err != nil {
		return data, err
	}
	defer func() { _ = f.Close() }()

	err = json.NewDecoder(f).Decode(&data)
	return data, err
}

func getLastTelemetryEventSent(t *testCtx, event string) ([]server.Call, error) {
	var matches []server.Call

	calls, err := t.getAllCalls("POST", "/data/v1/stats")
	if err != nil {
		return matches, err
	}

	for _, call := range calls {
		var req proton.SendStatsReq
		if err := json.Unmarshal(call.RequestBody, &req); err != nil {
			continue
		}
		if req.Event == event {
			matches = append(matches, call)
		}
	}
	return matches, err
}
