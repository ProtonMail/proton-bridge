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
	"strconv"
	"time"
)

type ConfigRecoveryValues struct {
	Duration int `json:"duration"`
}

type ConfigRecoveryDimensions struct {
	Autoconf       string `json:"autoconf"`
	ReportClick    string `json:"report_click"`
	ReportSent     string `json:"report_sent"`
	ClickedLink    string `json:"clicked_link"`
	FailureDetails string `json:"failure_details"`
}

type ConfigRecoveryData struct {
	MeasurementGroup string
	Event            string
	Values           ConfigRecoveryValues
	Dimensions       ConfigRecoveryDimensions
}

type ConfigRecoveryBuilder struct{}

func (*ConfigRecoveryBuilder) New(data *ConfigurationStatusData) ConfigRecoveryData {
	return ConfigRecoveryData{
		MeasurementGroup: "bridge.any.configuration",
		Event:            "bridge_config_recovery",
		Values: ConfigRecoveryValues{
			Duration: int(time.Since(data.DataV1.PendingSince).Minutes()),
		},
		Dimensions: ConfigRecoveryDimensions{
			Autoconf:       data.DataV1.Autoconf,
			ReportClick:    strconv.FormatBool(data.DataV1.ReportClick),
			ReportSent:     strconv.FormatBool(data.DataV1.ReportSent),
			ClickedLink:    data.clickedLinkToString(),
			FailureDetails: data.DataV1.FailureDetails,
		},
	}
}
