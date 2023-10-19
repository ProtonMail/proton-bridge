// Copyright (c) 2023 Proton AG
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

package configstatus_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
	"github.com/stretchr/testify/require"
)

func TestConfigurationAbort_default(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	var builder = configstatus.ConfigAbortBuilder{}
	req := builder.New(config)

	require.Equal(t, "bridge.any.configuration", req.MeasurementGroup)
	require.Equal(t, "bridge_config_abort", req.Event)
	require.Equal(t, 0, req.Values.Duration)
	require.Equal(t, "false", req.Dimensions.ReportClick)
	require.Equal(t, "false", req.Dimensions.ReportSent)
	require.Equal(t, "", req.Dimensions.ClickedLink)
}

func TestConfigurationAbort_fed(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	var data = configstatus.ConfigurationStatusData{
		Metadata: configstatus.Metadata{Version: "1.0.0"},
		DataV1: configstatus.DataV1{
			PendingSince:   time.Now().Add(-10 * time.Minute),
			LastProgress:   time.Time{},
			Autoconf:       "Mr TBird",
			ClickedLink:    42,
			ReportSent:     false,
			ReportClick:    true,
			FailureDetails: "Not an error",
		},
	}
	require.NoError(t, dumpConfigStatusInFile(&data, file))

	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	var builder = configstatus.ConfigAbortBuilder{}
	req := builder.New(config)

	require.Equal(t, "bridge.any.configuration", req.MeasurementGroup)
	require.Equal(t, "bridge_config_abort", req.Event)
	require.Equal(t, 10, req.Values.Duration)
	require.Equal(t, "true", req.Dimensions.ReportClick)
	require.Equal(t, "false", req.Dimensions.ReportSent)
	require.Equal(t, "[1,3,5]", req.Dimensions.ClickedLink)
}
