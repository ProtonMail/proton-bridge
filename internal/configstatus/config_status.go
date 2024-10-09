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

package configstatus

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/sirupsen/logrus"
)

const version = "1.0.0"

func LoadConfigurationStatus(filepath string) (*ConfigurationStatus, error) {
	status := ConfigurationStatus{
		FilePath: filepath,
		DataLock: safe.NewRWMutex(),
		Data:     &ConfigurationStatusData{},
	}

	if _, err := os.Stat(filepath); err == nil {
		if err := status.Load(); err == nil {
			return &status, nil
		}
		logrus.WithError(err).Warn("Cannot load configuration status file. Reset it.")
	}

	status.Data.init()
	if err := status.Save(); err != nil {
		return &status, err
	}
	return &status, nil
}

func (status *ConfigurationStatus) Load() error {
	bytes, err := os.ReadFile(status.FilePath)
	if err != nil {
		return err
	}

	var metadata MetadataOnly
	if err := json.Unmarshal(bytes, &metadata); err != nil {
		return err
	}

	if metadata.Metadata.Version != version {
		return fmt.Errorf("unsupported configstatus file version %s", metadata.Metadata.Version)
	}

	return json.Unmarshal(bytes, status.Data)
}

func (status *ConfigurationStatus) Save() error {
	temp := status.FilePath + "_temp"
	f, err := os.Create(temp) //nolint:gosec
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(status.Data)
	if err := f.Close(); err != nil {
		logrus.WithError(err).Error("Error while closing configstatus file.")
	}
	if err != nil {
		return err
	}

	return os.Rename(temp, status.FilePath)
}

func (status *ConfigurationStatus) IsPending() bool {
	status.DataLock.RLock()
	defer status.DataLock.RUnlock()

	return !status.Data.DataV1.PendingSince.IsZero()
}

func (status *ConfigurationStatus) isPendingSinceMin() int {
	if min := int(time.Since(status.Data.DataV1.PendingSince).Minutes()); min > 0 { //nolint:predeclared
		return min
	}
	return 0
}

func (status *ConfigurationStatus) IsFromFailure() bool {
	status.DataLock.RLock()
	defer status.DataLock.RUnlock()

	return status.Data.DataV1.FailureDetails != ""
}

func (status *ConfigurationStatus) ApplySuccess() error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	status.Data.init()
	status.Data.DataV1.PendingSince = time.Time{}
	return status.Save()
}

func (status *ConfigurationStatus) ApplyFailure(err string) error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	status.Data.init()
	status.Data.DataV1.FailureDetails = err
	return status.Save()
}

func (status *ConfigurationStatus) ApplyProgress() error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	status.Data.DataV1.LastProgress = time.Now()
	return status.Save()
}

func (status *ConfigurationStatus) RecordLinkClicked(link uint64) error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	if !status.Data.hasLinkClicked(link) {
		status.Data.setClickedLink(link)
		return status.Save()
	}
	return nil
}

func (status *ConfigurationStatus) ReportClicked() error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	if !status.Data.DataV1.ReportClick {
		status.Data.DataV1.ReportClick = true
		return status.Save()
	}
	return nil
}

func (status *ConfigurationStatus) ReportSent() error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	if !status.Data.DataV1.ReportSent {
		status.Data.DataV1.ReportSent = true
		return status.Save()
	}
	return nil
}

func (status *ConfigurationStatus) AutoconfigUsed(client string) error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()

	if client != status.Data.DataV1.Autoconf {
		status.Data.DataV1.Autoconf = client
		return status.Save()
	}
	return nil
}

func (status *ConfigurationStatus) Remove() error {
	status.DataLock.Lock()
	defer status.DataLock.Unlock()
	return os.Remove(status.FilePath)
}

func (data *ConfigurationStatusData) init() {
	data.Metadata = Metadata{
		Version: version,
	}
	data.DataV1.PendingSince = time.Now()
	data.DataV1.LastProgress = time.Time{}
	data.DataV1.Autoconf = ""
	data.DataV1.ClickedLink = 0
	data.DataV1.ReportSent = false
	data.DataV1.ReportClick = false
	data.DataV1.FailureDetails = ""
}

func (data *ConfigurationStatusData) setClickedLink(pos uint64) {
	data.DataV1.ClickedLink |= 1 << pos
}

func (data *ConfigurationStatusData) hasLinkClicked(pos uint64) bool {
	val := data.DataV1.ClickedLink & (1 << pos)
	return val > 0
}

func (data *ConfigurationStatusData) clickedLinkToString() string {
	var str = ""
	var first = true
	for i := 0; i < 64; i++ {
		if data.hasLinkClicked(uint64(i)) { //nolint:gosec // disable G115
			if !first {
				str += ","
			} else {
				first = false
				str += "["
			}
			str += strconv.Itoa(i)
		}
	}
	if str != "" {
		str += "]"
	}
	return str
}
