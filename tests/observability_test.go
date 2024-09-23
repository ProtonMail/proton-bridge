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

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/evtloopmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/syncmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice/observabilitymetrics"
)

// userHeartbeatPermutationsObservability - corresponds to bridge_generic_user_heartbeat_total_v1.schema.json.
func (s *scenario) userHeartbeatPermutationsObservability(username string) error {
	metrics := observability.GenerateAllHeartbeatMetricPermutations()
	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		batch := proton.ObservabilityBatch{Metrics: metrics}
		return c.SendObservabilityBatch(ctx, batch)
	})
}

// userDistinctionMetricsPermutationsObservability - corresponds to:
// bridge_sync_errors_users_total_v1.schema.json
// bridge_event_loop_events_errors_users_total_v1.schema.json.
func (s *scenario) userDistinctionMetricsPermutationsObservability(username string) error {
	batch := proton.ObservabilityBatch{
		Metrics: observability.GenerateAllUsedDistinctionMetricPermutations()}
	return s.t.withClientPass(context.Background(), username, s.t.getUserByName(username).userPass, func(ctx context.Context, c *proton.Client) error {
		err := c.SendObservabilityBatch(ctx, batch)
		return err
	})
}

// syncFailureMessageEventsObservability - corresponds to bridge_sync_message_event_failures_total_v1.schema.json.
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

// eventLoopFailureMessageEventsObservability - corresponds to bridge_event_loop_message_event_failures_total_v1.schema.json.
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

// syncFailureMessageBuiltObservability - corresponds to bridge_sync_message_event_failures_total_v1.schema.json.
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

// syncSuccessMessageBuiltObservability - corresponds to bridge_sync_message_build_success_total_v1.schema.json.
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
