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

package notifications

import (
	"time"

	"github.com/ProtonMail/go-proton-api"
)

func generatedNotificationDisplayedMetric(status string, value int) proton.ObservabilityMetric {
	// Value cannot be zero or negative
	if value < 1 {
		value = 1
	}

	return proton.ObservabilityMetric{
		Name:      "bridge_remoteNotification_displayed_total",
		Version:   1,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": value,
			"Labels": map[string]string{
				"status": status,
			},
		},
	}
}

func GenerateReceivedMetric(count int) proton.ObservabilityMetric {
	return generatedNotificationDisplayedMetric("received", count)
}

func GenerateProcessedMetric(count int) proton.ObservabilityMetric {
	return generatedNotificationDisplayedMetric("processed", count)
}
