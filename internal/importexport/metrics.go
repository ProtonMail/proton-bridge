// Copyright (c) 2020 Proton Technologies AG
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

package importexport

import (
	"strconv"

	"github.com/ProtonMail/proton-bridge/internal/metrics"
)

type metricsManager struct {
	ie       *ImportExport
	category metrics.Category
}

func newImportMetricsManager(ie *ImportExport) *metricsManager {
	return &metricsManager{
		ie:       ie,
		category: metrics.Import,
	}
}

func newExportMetricsManager(ie *ImportExport) *metricsManager {
	return &metricsManager{
		ie:       ie,
		category: metrics.Export,
	}
}

func (m *metricsManager) Load(numberOfMailboxes int) {
	label := strconv.Itoa(numberOfMailboxes)
	m.ie.SendMetric(metrics.New(m.category, metrics.TransferLoad, metrics.Label(label)))
}

func (m *metricsManager) Start() {
	m.ie.SendMetric(metrics.New(m.category, metrics.TransferStart, metrics.NoLabel))
}

func (m *metricsManager) Complete() {
	m.ie.SendMetric(metrics.New(m.category, metrics.TransferComplete, metrics.NoLabel))
}

func (m *metricsManager) Cancel() {
	m.ie.SendMetric(metrics.New(m.category, metrics.TransferCancel, metrics.NoLabel))
}

func (m *metricsManager) Fail() {
	m.ie.SendMetric(metrics.New(m.category, metrics.TransferFail, metrics.NoLabel))
}
