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
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
)

const genericHeartbeatSchemaName = "bridge_generic_user_heartbeat_total"
const genericHeartbeatVersion = 2

type heartbeatData struct {
	receivedSyncError      bool
	receivedEventLoopError bool
	receivedOtherError     bool
	receivedGluonError     bool
}

func (d *distinctionUtility) resetHeartbeatData() {
	d.heartbeatData.receivedSyncError = false
	d.heartbeatData.receivedOtherError = false
	d.heartbeatData.receivedEventLoopError = false
	d.heartbeatData.receivedGluonError = false
}

func (d *distinctionUtility) updateHeartbeatData(errType DistinctionErrorTypeEnum) {
	d.withUpdateHeartbeatDataLock(func() {
		//nolint:exhaustive
		switch errType {
		case SyncError:
			d.heartbeatData.receivedSyncError = true
		case EventLoopError:
			d.heartbeatData.receivedEventLoopError = true
		case GluonMessageError, GluonImapError, GluonOtherError:
			d.heartbeatData.receivedGluonError = true
		}
	})
}

func (d *distinctionUtility) runHeartbeat() {
	go func() {
		defer async.HandlePanic(d.panicHandler)
		defer d.heartbeatTicker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-d.heartbeatTicker.C:
				d.sendHeartbeat()
			}
		}
	}()
}

func (d *distinctionUtility) withUpdateHeartbeatDataLock(fn func()) {
	d.heartbeatDataLock.Lock()
	defer d.heartbeatDataLock.Unlock()
	fn()
}

// sendHeartbeat - will only send a heartbeat if there is an authenticated client
// otherwise we might end up polluting the cache and therefore our metrics.
func (d *distinctionUtility) sendHeartbeat() {
	d.withUpdateHeartbeatDataLock(func() {
		d.sendMetricsWithGuard(d.generateHeartbeatUserMetric())
		d.resetHeartbeatData()
	})
}

func formatBool(value bool) string {
	return fmt.Sprintf("%t", value)
}

// generateHeartbeatUserMetric creates the heartbeat user metric and includes the relevant data.
func (d *distinctionUtility) generateHeartbeatUserMetric() proton.ObservabilityMetric {
	return generateHeartbeatMetric(
		d.getUserPlanSafe(),
		d.getEmailClientUserAgent(),
		getEnabled(d.settingsGetter.GetProxyAllowed()),
		getEnabled(d.getBetaAccessEnabled()),
		formatBool(d.heartbeatData.receivedOtherError),
		formatBool(d.heartbeatData.receivedSyncError),
		formatBool(d.heartbeatData.receivedEventLoopError),
		formatBool(d.heartbeatData.receivedGluonError),
	)
}

func generateHeartbeatMetric(plan, mailClient, dohEnabled, betaAccess, otherError, syncError, eventLoopError, gluonError string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      genericHeartbeatSchemaName,
		Version:   genericHeartbeatVersion,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"plan":                   plan,
				"mailClient":             mailClient,
				"dohEnabled":             dohEnabled,
				"betaAccessEnabled":      betaAccess,
				"receivedOtherError":     otherError,
				"receivedSyncError":      syncError,
				"receivedEventLoopError": eventLoopError,
				"receivedGluonError":     gluonError,
			},
		},
	}
}
