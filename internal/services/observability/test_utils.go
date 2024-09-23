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

func GenerateAllUsedDistinctionMetricPermutations() []proton.ObservabilityMetric {
	planValues := []string{
		planUnknown,
		planOther,
		planBusiness,
		planIndividual,
		planGroup}
	mailClientValues := []string{
		emailAgentAppleMail,
		emailAgentOutlook,
		emailAgentThunderbird,
		emailAgentOther,
		emailAgentUnknown,
	}
	enabledValues := []string{
		getEnabled(true), getEnabled(false),
	}

	var metrics []proton.ObservabilityMetric

	for _, schemaName := range errorSchemaMap {
		for _, plan := range planValues {
			for _, mailClient := range mailClientValues {
				for _, dohEnabled := range enabledValues {
					for _, betaAccess := range enabledValues {
						metrics = append(metrics, generateUserMetric(schemaName, plan, mailClient, dohEnabled, betaAccess))
					}
				}
			}
		}
	}
	return metrics
}

func GenerateAllHeartbeatMetricPermutations() []proton.ObservabilityMetric {
	planValues := []string{
		planUnknown,
		planOther,
		planBusiness,
		planIndividual,
		planGroup}
	mailClientValues := []string{
		emailAgentAppleMail,
		emailAgentOutlook,
		emailAgentThunderbird,
		emailAgentOther,
		emailAgentUnknown,
	}
	enabledValues := []string{
		getEnabled(true), getEnabled(false),
	}

	trueFalseValues := []string{
		"true", "false",
	}

	var metrics []proton.ObservabilityMetric
	for _, plan := range planValues {
		for _, mailClient := range mailClientValues {
			for _, dohEnabled := range enabledValues {
				for _, betaAccess := range enabledValues {
					for _, receivedOtherError := range trueFalseValues {
						for _, receivedSyncError := range trueFalseValues {
							for _, receivedEventLoopError := range trueFalseValues {
								metrics = append(metrics,
									generateHeartbeatMetric(plan,
										mailClient,
										dohEnabled,
										betaAccess,
										receivedOtherError,
										receivedSyncError,
										receivedEventLoopError,
									),
								)
							}
						}
					}
				}
			}
		}
	}
	return metrics
}
