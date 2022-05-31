// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// retryConnectionSleeps defines a smooth cool down in seconds.
var retryConnectionSleeps = []int{2, 5, 10, 30, 60} //nolint:gochecknoglobals

func (m *manager) pingUntilSuccess() {
	if m.isPingOngoing() {
		logrus.Debug("Ping already ongoing")
		return
	}
	m.pingingStarted()
	defer m.pingingStopped()

	attempt := 0
	for {
		ctx := ContextWithoutRetry(context.Background())
		err := m.testPing(ctx)
		if err == nil {
			return
		}

		waitTime := getRetryConnectionSleep(attempt)
		attempt++
		logrus.WithError(err).WithField("attempt", attempt).WithField("wait", waitTime).Debug("Connection (still) not available")
		time.Sleep(waitTime)
	}
}

func (m *manager) isPingOngoing() bool {
	m.pingMutex.RLock()
	defer m.pingMutex.RUnlock()

	return m.isPinging
}

func (m *manager) pingingStarted() {
	m.pingMutex.Lock()
	defer m.pingMutex.Unlock()
	m.isPinging = true
}

func (m *manager) pingingStopped() {
	m.pingMutex.Lock()
	defer m.pingMutex.Unlock()
	m.isPinging = false
}

func getRetryConnectionSleep(idx int) time.Duration {
	if idx >= len(retryConnectionSleeps) {
		idx = len(retryConnectionSleeps) - 1
	}
	sec := retryConnectionSleeps[idx]
	return time.Duration(sec) * time.Second
}

func (m *manager) testPing(ctx context.Context) error {
	if _, err := m.r(ctx).Get("/tests/ping"); err != nil {
		return err
	}
	return nil
}
