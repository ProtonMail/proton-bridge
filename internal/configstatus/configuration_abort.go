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
)

type ConfigAbortValues struct {
	Duration int `json:"duration"`
}

type ConfigAbortDimensions struct {
	ReportClick string `json:"report_click"`
	ReportSent  string `json:"report_sent"`
	ClickedLink string `json:"clicked_link"`
}

type ConfigAbortData struct {
	MeasurementGroup string
	Event            string
	Values           ConfigSuccessValues
	Dimensions       ConfigSuccessDimensions
}

type ConfigAbortBuilder struct{}

func (*ConfigAbortBuilder) New(config *ConfigurationStatus) ConfigAbortData {
	config.DataLock.RLock()
	defer config.DataLock.RUnlock()

	return ConfigAbortData{
		MeasurementGroup: "bridge.any.configuration",
		Event:            "bridge_config_abort",
		Values: ConfigSuccessValues{
			Duration: config.isPendingSinceMin(),
		},
		Dimensions: ConfigSuccessDimensions{
			ReportClick: strconv.FormatBool(config.Data.DataV1.ReportClick),
			ReportSent:  strconv.FormatBool(config.Data.DataV1.ReportSent),
			ClickedLink: config.Data.clickedLinkToString(),
		},
	}
}
