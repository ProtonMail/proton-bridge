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

package configstatus

import (
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
)

const ProgressCheckInterval = time.Hour

type Metadata struct {
	Version string `json:"version"`
}

type MetadataOnly struct {
	Metadata Metadata `json:"metadata"`
}

type DataV1 struct {
	PendingSince   time.Time `json:"pending_since"`
	LastProgress   time.Time `json:"last_progress"`
	Autoconf       string    `json:"auto_conf"`
	ClickedLink    uint64    `json:"clicked_link"`
	ReportSent     bool      `json:"report_sent"`
	ReportClick    bool      `json:"report_click"`
	FailureDetails string    `json:"failure_details"`
}

type ConfigurationStatusData struct {
	Metadata Metadata `json:"metadata"`
	DataV1   DataV1   `json:"dataV1"`
}

type ConfigurationStatus struct {
	FilePath string
	DataLock safe.RWMutex

	Data *ConfigurationStatusData
}
