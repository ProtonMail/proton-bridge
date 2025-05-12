// Copyright (c) 2025 Proton AG
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

package gluonmetrics

import (
	"time"

	"github.com/ProtonMail/go-proton-api"
)

const (
	newIMAPConnectionThresholdExceededSchemaName = "bridge_imap_recently_opened_connections_total"
	newIMAPConnectionThresholdExceededVersion    = 1
)

func GenerateNewOpenedIMAPConnectionsExceedThreshold(emailClient, totalOpenIMAPConnectionCount, newlyOpenedIMAPConnectionCount string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      newIMAPConnectionThresholdExceededSchemaName,
		Version:   newIMAPConnectionThresholdExceededVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"mailClient":                                   emailClient,
				"numberOfOpenIMAPConnectionsBuckets":           totalOpenIMAPConnectionCount,
				"numberOfRecentlyOpenedIMAPConnectionsBuckets": newlyOpenedIMAPConnectionCount,
			},
		},
	}
}
