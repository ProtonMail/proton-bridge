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

package tests

import (
	"context"
	"fmt"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/evtloopmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/syncmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	smtpMetrics "github.com/ProtonMail/proton-bridge/v3/internal/services/smtp/observabilitymetrics"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice/observabilitymetrics"
)

// userHeartbeatPermutationsObservability corresponds to bridge_generic_user_heartbeat_total_v1.schema.json.
func (s *scenario) userHeartbeatPermutationsObservability(username string) error {
	const batchSize = 1000
	metrics := observability.GenerateAllHeartbeatMetricPermutations()
	metricLen := len(metrics)

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		for i := 0; i < len(metrics); i += batchSize {
			end := i + batchSize
			if end > metricLen {
				end = metricLen
			}

			batch := proton.ObservabilityBatch{Metrics: metrics[i:end]}
			if err := c.SendObservabilityBatch(ctx, batch); err != nil {
				return err
			}
		}

		return nil
	})
}

// userDistinctionMetricsPermutationsObservability corresponds to:
//   - bridge_sync_errors_users_total_v1.schema.json
//   - bridge_gluon_imap_errors_users_total_v1.schema.json
//   - bridge_gluon_message_errors_users_total_v1.schema.json
//   - bridge_gluon_other_errors_users_total_v1.schema.json
//   - bridge_event_loop_events_errors_users_total_v1.schema.json.
//   - bridge_smtp_errors_users_total_v1.schema.json
func (s *scenario) userDistinctionMetricsPermutationsObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: observability.GenerateAllUsedDistinctionMetricPermutations()}
	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// syncFailureMessageEventsObservability corresponds to bridge_sync_message_event_failures_total_v1.schema.json.
func (s *scenario) syncFailureMessageEventsObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			syncmsgevents.GenerateSyncFailureCreateMessageEventMetric(),
			syncmsgevents.GenerateSyncFailureDeleteMessageEventMetric(),
		},
	}
	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// eventLoopFailureMessageEventsObservability corresponds to bridge_event_loop_message_event_failures_total_v1.schema.json.
func (s *scenario) eventLoopFailureMessageEventsObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			evtloopmsgevents.GenerateMessageEventFailedToBuildDraft(),
			evtloopmsgevents.GenerateMessageEventFailedToBuildMessage(),
			evtloopmsgevents.GenerateMessageEventFailureCreateMessageMetric(),
			evtloopmsgevents.GenerateMessageEventFailureDeleteMessageMetric(),
			evtloopmsgevents.GenerateMessageEventFailureUpdateMetric(),
			evtloopmsgevents.GenerateMessageEventUpdateChannelDoesNotExist(),
		},
	}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// syncFailureMessageBuiltObservability corresponds to bridge_sync_message_event_failures_total_v1.schema.json.
func (s *scenario) syncFailureMessageBuiltObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			observabilitymetrics.GenerateNoUnlockedKeyringMetric(),
			observabilitymetrics.GenerateFailedToBuildMetric(),
		},
	}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// syncSuccessMessageBuiltObservability corresponds to bridge_sync_message_build_success_total_v1.schema.json.
func (s *scenario) syncSuccessMessageBuiltObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			observabilitymetrics.GenerateMessageBuiltSuccessMetric(),
		},
	}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// testGluonErrorObservabilityMetrics corresponds to bridge_gluon_errors_total_v1.schema.json.
func (s *scenario) testGluonErrorObservabilityMetrics(username string) error {
	allMetrics := observability.GenerateAllGluonMetrics()

	parsedMetrics := []proton.ObservabilityMetric{}
	for _, el := range allMetrics {
		ok, parsedMetric := observability.VerifyAndParseGenericMetrics(el)
		if !ok {
			return fmt.Errorf("failed to parse generic gluon metric")
		}
		parsedMetrics = append(parsedMetrics, parsedMetric)
	}

	batch := proton.ObservabilityBatch{Metrics: parsedMetrics}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// SMTPErrorObservabilityMetrics corresponds to bridge_smtp_errors_total_v1.schema.json.
func (s *scenario) SMTPErrorObservabilityMetrics(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			smtpMetrics.GenerateFailedGetParentID(),
			smtpMetrics.GenerateUnsupportedMIMEType(),
			smtpMetrics.GenerateFailedCreateDraft(),
			smtpMetrics.GenerateFailedCreateAttachments(),
			smtpMetrics.GenerateFailedCreatePackages(),
			smtpMetrics.GenerateFailedToGetRecipients(),
			smtpMetrics.GenerateFailedSendDraft(),
			smtpMetrics.GenerateFailedDeleteFromDrafts(),
		},
	}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

func (s *scenario) SMTPSendSuccessObservabilityMetric(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: []proton.ObservabilityMetric{
			smtpMetrics.GenerateSMTPSendSuccess(),
		},
	}

	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}
