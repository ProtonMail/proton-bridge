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

import "time"

type ConfigProgressValues struct {
	NbDay          int `json:"nb_day"`
	NbDaySinceLast int `json:"nb_day_since_last"`
}

type ConfigProgressData struct {
	MeasurementGroup string
	Event            string
	Values           ConfigProgressValues
	Dimensions       struct{}
}

type ConfigProgressBuilder struct{}

func (*ConfigProgressBuilder) New(config *ConfigurationStatus) ConfigProgressData {
	config.DataLock.RLock()
	defer config.DataLock.RUnlock()

	return ConfigProgressData{
		MeasurementGroup: "bridge.any.configuration",
		Event:            "bridge_config_progress",
		Values: ConfigProgressValues{
			NbDay:          numberOfDay(time.Now(), config.Data.DataV1.PendingSince),
			NbDaySinceLast: numberOfDay(time.Now(), config.Data.DataV1.LastProgress),
		},
	}
}

func numberOfDay(now, prev time.Time) int {
	if now.IsZero() || prev.IsZero() {
		return 1
	}
	if now.Year() > prev.Year() {
		if now.YearDay() > prev.YearDay() {
			return 365 + (now.YearDay() - prev.YearDay())
		}
		return (prev.YearDay() + now.YearDay()) - 365
	} else if now.YearDay() > prev.YearDay() {
		return now.YearDay() - prev.YearDay()
	}
	return 0
}
