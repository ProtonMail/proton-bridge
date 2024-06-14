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

package telemetry

type RepairData struct {
	MeasurementGroup string
	Event            string
	Values           map[string]string
	Dimensions       map[string]string
}

func NewRepairTriggerData() RepairData {
	return RepairData{
		MeasurementGroup: "bridge.any.repair",
		Event:            "repair_trigger",
		Values:           map[string]string{},
		Dimensions:       map[string]string{},
	}
}

func NewRepairDeferredTriggerData() RepairData {
	return RepairData{
		MeasurementGroup: "bridge.any.repair",
		Event:            "repair_deferred_trigger",
		Values:           map[string]string{},
		Dimensions:       map[string]string{},
	}
}
