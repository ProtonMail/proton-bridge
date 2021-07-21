// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// retryConnectionSleeps defines a smooth cool down in seconds.
	retryConnectionSleeps = []int{2, 5, 10, 30, 60} // nolint[gochecknoglobals]
)

func (m *manager) pingUntilSuccess() {
	attempt := 0
	for {
		err := m.testPing(context.Background())
		if err == nil {
			return
		}

		waitTime := getRetryConnectionSleep(attempt)
		attempt++
		logrus.WithError(err).WithField("attempt", attempt).WithField("wait", waitTime).Debug("Connection not available")
		time.Sleep(waitTime)
	}
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
