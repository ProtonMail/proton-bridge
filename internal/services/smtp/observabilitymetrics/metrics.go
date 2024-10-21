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
	smtpErrorsSchemaName    = "bridge_smtp_errors_total"
	smtpErrorsSchemaVersion = 1

	smtpSendSuccessSchemaName    = "bridge_smtp_send_success_total"
	smtpSendSuccessSchemaVersion = 1
)

func generateSMTPErrorObservabilityMetric(errorType string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      smtpErrorsSchemaName,
		Version:   smtpErrorsSchemaVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"errorType": errorType,
			},
		},
	}
}

func GenerateFailedGetParentID() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedGetParentId")
}

func GenerateUnsupportedMIMEType() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("unsupportedMIMEType")
}

func GenerateFailedCreateDraft() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedToCreateDraft")
}

func GenerateFailedCreateAttachments() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedCreateAttachments")
}

func GenerateFailedToGetRecipients() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedGetRecipients")
}

func GenerateFailedCreatePackages() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedCreatePackages")
}

func GenerateFailedSendDraft() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedSendDraft")
}

func GenerateFailedDeleteFromDrafts() proton.ObservabilityMetric {
	return generateSMTPErrorObservabilityMetric("failedDeleteFromDrafts")
}

func GenerateSMTPSendSuccess() proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      smtpSendSuccessSchemaName,
		Version:   smtpSendSuccessSchemaVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value":  1,
			"Labels": map[string]string{},
		},
	}
}
