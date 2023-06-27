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

	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
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
	configStatusFile, err := getConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}
	data, err := loadConfigStatusFile(configStatusFile)
	if err != nil {
		return err
	}
	if data.DataV1.PendingSince.IsZero() {
		return fmt.Errorf("expected ConfigStatus pending but got success instead")
	}

	return nil
}

func (s *scenario) configStatusSucceedForUser(username string) error {
	configStatusFile, err := getConfigStatusFile(s.t, username)
	if err != nil {
		return err
	}
	data, err := loadConfigStatusFile(configStatusFile)
	if err != nil {
		return err
	}
	if !data.DataV1.PendingSince.IsZero() {
		return fmt.Errorf("expected ConfigStatus success but got pending since %s", data.DataV1.PendingSince)
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

func loadConfigStatusFile(filepath string) (configstatus.ConfigurationStatusData, error) {
	data := configstatus.ConfigurationStatusData{}
	if _, err := os.Stat(filepath); err != nil {
		return data, err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return data, err
	}
	defer func() { _ = f.Close() }()

	err = json.NewDecoder(f).Decode(&data)
	return data, err
}
