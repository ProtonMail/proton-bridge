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

package user

import (
	"encoding/json"
	"os"
	"time"
)

type ConfigurationStatusData struct {
	PendingSince   time.Time `json:"auto_update"`
	LastProgress   time.Time `json:"last_progress"`
	Autoconf       string    `json:"auto_conf"`
	ClickedLink    uint64    `json:"clicked_link"`
	ReportSent     bool      `json:"report_sent"`
	ReportClick    bool      `json:"report_click"`
	FailureDetails string    `json:"failure_details"`
}

type ConfigurationStatus struct {
	FilePath string
	Data     ConfigurationStatusData
}

func LoadConfigurationStatus(filepath string) (*ConfigurationStatus, error) {
	status := ConfigurationStatus{
		FilePath: filepath,
		Data:     ConfigurationStatusData{},
	}
	if _, err := os.Stat(filepath); err == nil {
		if err := status.Data.load(filepath); err == nil {
			return &status, nil
		}
	} else {
		status.Data.init()
		if err := status.save(); err == nil {
			return &status, nil
		}
	}
	return &status, nil
}

func (status *ConfigurationStatus) Success() error {
	status.Data.init()
	status.Data.PendingSince = time.Time{}
	return status.save()
}

func (status *ConfigurationStatus) Failure(err string) error {
	status.Data.init()
	status.Data.FailureDetails = err
	return status.save()
}

func (status *ConfigurationStatus) Progress() error {
	status.Data.LastProgress = time.Now()
	return status.save()
}

func (status *ConfigurationStatus) RecordLinkClicked(link uint) error {
	if !status.Data.hasLinkClicked(link) {
		status.Data.setClickedLink(link)
		return status.save()
	}
	return nil
}

func (status *ConfigurationStatus) save() error {
	return status.Data.save(status.FilePath)
}

func (data *ConfigurationStatusData) init() {
	data.PendingSince = time.Now()
	data.LastProgress = time.Time{}
	data.Autoconf = ""
	data.ClickedLink = 0
	data.ReportSent = false
	data.ReportClick = false
	data.FailureDetails = ""
}

func (data *ConfigurationStatusData) load(filepath string) error {
	f, err := os.Open(filepath) // nolint: gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewDecoder(f).Decode(data)
}

func (data *ConfigurationStatusData) save(filepath string) error {
	f, err := os.Create(filepath) // nolint: gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(data)
}

func (data *ConfigurationStatusData) setClickedLink(pos uint) {
	data.ClickedLink |= 1 << pos
}

func (data *ConfigurationStatusData) hasLinkClicked(pos uint) bool {
	val := data.ClickedLink & (1 << pos)
	return val > 0
}
