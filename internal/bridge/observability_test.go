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

package bridge_test

import (
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestBridge_Observability(t *testing.T) {
	testMetric := proton.ObservabilityMetric{
		Name:      "test1",
		Version:   1,
		Timestamp: time.Now().Unix(),
		Data:      nil,
	}

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		throttlePeriod := time.Millisecond * 500
		observability.ModifyThrottlePeriod(throttlePeriod)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, getErr(bridge.LoginFull(ctx, username, password, nil, nil)))

			bridge.PushObservabilityMetric(testMetric)
			time.Sleep(time.Millisecond * 50) // Wait for the metric to be sent
			require.Equal(t, 1, len(s.GetObservabilityStatistics().Metrics))

			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * 5) // Minor delay between each so our tests aren't flaky
				bridge.PushObservabilityMetric(testMetric)
			}
			// We should still have only 1 metric sent as the throttleDuration has not passed
			require.Equal(t, 1, len(s.GetObservabilityStatistics().Metrics))

			// Wait for throttle duration to pass; we should have our remaining metrics posted
			time.Sleep(throttlePeriod)
			require.Equal(t, 11, len(s.GetObservabilityStatistics().Metrics))

			// Wait for the throttle duration to reset; i.e. so we have enough time to send a request immediately
			time.Sleep(throttlePeriod)
			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * 5)
				bridge.PushObservabilityMetric(testMetric)
			}
			// We should only have one additional metric sent immediately
			require.Equal(t, 12, len(s.GetObservabilityStatistics().Metrics))

			// Wait for the others to be sent
			time.Sleep(throttlePeriod)
			require.Equal(t, 21, len(s.GetObservabilityStatistics().Metrics))

			// Spam the endpoint a bit
			for i := 0; i < 300; i++ {
				if i < 200 {
					time.Sleep(time.Millisecond * 10)
				}
				bridge.PushObservabilityMetric(testMetric)
			}

			// Ensure we've sent all metrics
			time.Sleep(throttlePeriod)

			observabilityStats := s.GetObservabilityStatistics()
			require.Equal(t, 321, len(observabilityStats.Metrics))

			// Verify that each request had a throttleDuration time difference between each request
			for i := 0; i < len(observabilityStats.RequestTime)-1; i++ {
				tOne := observabilityStats.RequestTime[i]
				tTwo := observabilityStats.RequestTime[i+1]
				require.True(t, tTwo.Sub(tOne).Abs() > throttlePeriod)
			}
		})
	})
}
