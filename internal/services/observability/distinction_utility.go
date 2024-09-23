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
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

var updateInterval = time.Minute * 5 //nolint:gochecknoglobals

type observabilitySender interface {
	addMetricsIfClients(metric ...proton.ObservabilityMetric)
}

// distinctionUtility - used to discern whether X number of events stem from Y number of users.
type distinctionUtility struct {
	ctx context.Context

	panicHandler async.PanicHandler

	lastSentMap map[DistinctionErrorTypeEnum]time.Time // Ensures we don't step over the limit of one user update every 5 mins.

	observabilitySender observabilitySender
	settingsGetter      settingsGetter

	userPlanUnsafe string
	userPlanLock   sync.Mutex

	heartbeatData     heartbeatData
	heartbeatDataLock sync.Mutex
	heartbeatTicker   *time.Ticker
}

func newDistinctionUtility(ctx context.Context, panicHandler async.PanicHandler, observabilitySender observabilitySender) *distinctionUtility {
	distinctionUtility := &distinctionUtility{
		ctx: ctx,

		panicHandler: panicHandler,

		lastSentMap: createLastSentMap(),

		observabilitySender: observabilitySender,

		userPlanUnsafe: planUnknown,

		heartbeatData:   heartbeatData{},
		heartbeatTicker: time.NewTicker(updateInterval),
	}

	return distinctionUtility
}

// sendMetricsWithGuard - schedules the metrics to be sent only if there are authenticated clients.
func (d *distinctionUtility) sendMetricsWithGuard(metrics ...proton.ObservabilityMetric) {
	if d.observabilitySender == nil {
		return
	}

	d.observabilitySender.addMetricsIfClients(metrics...)
}

func (d *distinctionUtility) setSettingsGetter(getter settingsGetter) {
	d.settingsGetter = getter
}

// checkAndUpdateLastSentMap - checks whether we have sent a relevant user update metric
// within the last 5 minutes.
func (d *distinctionUtility) checkAndUpdateLastSentMap(key DistinctionErrorTypeEnum) bool {
	curTime := time.Now()
	val, ok := d.lastSentMap[key]
	if !ok {
		d.lastSentMap[key] = curTime
		return true
	}

	if val.Add(updateInterval).Before(curTime) {
		d.lastSentMap[key] = curTime
		return true
	}

	return false
}

// generateUserMetric creates the relevant user update metric based on its type
// and the relevant settings. In the future this will need to be expanded to support multiple
// versions of the metric if we ever decide to change them.
func (d *distinctionUtility) generateUserMetric(
	metricType DistinctionErrorTypeEnum,
) proton.ObservabilityMetric {
	schemaName, ok := errorSchemaMap[metricType]
	if !ok {
		return proton.ObservabilityMetric{}
	}

	return generateUserMetric(schemaName, d.getUserPlanSafe(),
		d.getEmailClientUserAgent(),
		getEnabled(d.getProxyAllowed()),
		getEnabled(d.getBetaAccessEnabled()),
	)
}

func generateUserMetric(schemaName, plan, mailClient, dohEnabled, betaAccess string) proton.ObservabilityMetric {
	return proton.ObservabilityMetric{
		Name:      schemaName,
		Version:   1,
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"plan":              plan,
				"mailClient":        mailClient,
				"dohEnabled":        dohEnabled,
				"betaAccessEnabled": betaAccess,
			},
		},
	}
}

func (d *distinctionUtility) generateDistinctMetrics(errType DistinctionErrorTypeEnum, metrics ...proton.ObservabilityMetric) []proton.ObservabilityMetric {
	d.updateHeartbeatData(errType)

	if d.checkAndUpdateLastSentMap(errType) {
		metrics = append(metrics, d.generateUserMetric(errType))
	}
	return metrics
}

func (d *distinctionUtility) getEmailClientUserAgent() string {
	ua := ""
	if d.settingsGetter != nil {
		ua = d.settingsGetter.GetCurrentUserAgent()
	}

	return matchUserAgent(ua)
}

func (d *distinctionUtility) getBetaAccessEnabled() bool {
	if d.settingsGetter == nil {
		return false
	}
	return d.settingsGetter.GetUpdateChannel() == updater.EarlyChannel
}

func (d *distinctionUtility) getProxyAllowed() bool {
	if d.settingsGetter == nil {
		return false
	}
	return d.settingsGetter.GetProxyAllowed()
}
