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

package evtloopmsgevents

import (
	"time"

	"github.com/ProtonMail/go-proton-api"
)

const (
	messageEventErrorCaseSchemaName    = "bridge_event_loop_message_event_failures_total"
	messageEventErrorCaseSchemaVersion = 1
)

func generateMessageEventFailureObservabilityMetric(eventType string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      messageEventErrorCaseSchemaName,
		Version:   messageEventErrorCaseSchemaVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"eventType": eventType,
			},
		},
	}
}

func GenerateMessageEventFailureCreateMessageMetric() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("createMessageEvent")
}

func GenerateMessageEventFailureDeleteMessageMetric() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("deleteMessageEvent")
}

func GenerateMessageEventFailureUpdateMetric() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("updateEvent")
}

func GenerateMessageEventFailedToBuildMessage() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("failedToBuildMessage")
}

func GenerateMessageEventFailedToBuildDraft() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("failedToBuildDraft")
}

func GenerateMessageEventUpdateChannelDoesNotExist() proton.ObservabilityMetric {
	return generateMessageEventFailureObservabilityMetric("messageUpdateChannelDoesNotExist")
}
