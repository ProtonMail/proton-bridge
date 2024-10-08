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

package observability

import (
	"github.com/ProtonMail/go-proton-api"
)

type Adapter struct {
	sender Sender
}

func NewAdapter(sender Sender) *Adapter {
	return &Adapter{sender: sender}
}

// VerifyAndParseGenericMetrics parses a metric provided as an interface into a proton.ObservabilityMetric type.
// It's exported as it is also used in integration tests.
func VerifyAndParseGenericMetrics(metric map[string]interface{}) (bool, proton.ObservabilityMetric) {
	name, ok := metric["Name"].(string)
	if !ok {
		return false, proton.ObservabilityMetric{}
	}

	version, ok := metric["Version"].(int)
	if !ok {
		return false, proton.ObservabilityMetric{}
	}

	timestamp, ok := metric["Timestamp"].(int64)
	if !ok {
		return false, proton.ObservabilityMetric{}
	}

	data, ok := metric["Data"]
	if !ok {
		return false, proton.ObservabilityMetric{}
	}

	return true, proton.ObservabilityMetric{
		Name:      name,
		Version:   version,
		Timestamp: timestamp,
		Data:      data,
	}
}

func (adapter *Adapter) AddMetrics(metrics ...map[string]interface{}) {
	var typedMetrics []proton.ObservabilityMetric

	for _, metric := range metrics {
		if ok, m := VerifyAndParseGenericMetrics(metric); ok {
			typedMetrics = append(typedMetrics, m)
		}
	}

	if len(typedMetrics) > 0 {
		adapter.sender.AddMetrics(typedMetrics...)
	}
}

func (adapter *Adapter) AddDistinctMetrics(errType interface{}, metrics ...map[string]interface{}) {
	errTypeInt, ok := errType.(int)
	if !ok {
		return
	}

	var typedMetrics []proton.ObservabilityMetric
	for _, metric := range metrics {
		if ok, m := VerifyAndParseGenericMetrics(metric); ok {
			typedMetrics = append(typedMetrics, m)
		}
	}

	if len(typedMetrics) > 0 {
		adapter.sender.AddDistinctMetrics(DistinctionErrorTypeEnum(errTypeInt), typedMetrics...)
	}
}
