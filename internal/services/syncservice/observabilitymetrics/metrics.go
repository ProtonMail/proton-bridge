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

package observabilitymetrics

import (
	"time"

	"github.com/ProtonMail/go-proton-api"
)

const (
	errorCaseSchemaName      = "bridge_sync_message_build_errors_total"
	errorCaseSchemaVersion   = 1
	successCaseSchemaName    = "bridge_sync_message_build_success_total"
	successCaseSchemaVersion = 1
)

func generateStageBuildFailureObservabilityMetric(errorType string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      errorCaseSchemaName,
		Version:   errorCaseSchemaVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"errorType": errorType,
			},
		},
	}
}

func GenerateNoUnlockedKeyringMetric() proton.ObservabilityMetric {
	return generateStageBuildFailureObservabilityMetric("noUnlockedKeyring")
}

func GenerateFailedToBuildMetric() proton.ObservabilityMetric {
	return generateStageBuildFailureObservabilityMetric("failedToBuild")
}

// GenerateMessageBuiltSuccessMetric - Maybe this is incorrect, I'm not sure how metrics with no labels
// should be dealt with. The integration tests will tell us.
func GenerateMessageBuiltSuccessMetric() proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      successCaseSchemaName,
		Version:   successCaseSchemaVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value":  1,
			"Labels": map[string]string{},
		},
	}
}
