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

package configstatus_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
	"github.com/stretchr/testify/require"
)

func TestConfigStatus_init_virgin(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)
	require.Equal(t, "1.0.0", config.Data.Metadata.Version)

	require.Equal(t, false, config.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config.Data.DataV1.LastProgress.IsZero())

	require.Equal(t, "", config.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config.Data.DataV1.ClickedLink)
	require.Equal(t, false, config.Data.DataV1.ReportSent)
	require.Equal(t, false, config.Data.DataV1.ReportClick)
	require.Equal(t, "", config.Data.DataV1.FailureDetails)
}

func TestConfigStatus_init_existing(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	var data = configstatus.ConfigurationStatusData{
		Metadata: configstatus.Metadata{Version: "1.0.0"},
		DataV1:   configstatus.DataV1{Autoconf: "Mr TBird"},
	}
	require.NoError(t, dumpConfigStatusInFile(&data, file))

	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config.Data.Metadata.Version)
	require.Equal(t, "Mr TBird", config.Data.DataV1.Autoconf)
}

func TestConfigStatus_init_bad_version(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	var data = configstatus.ConfigurationStatusData{
		Metadata: configstatus.Metadata{Version: "2.0.0"},
		DataV1:   configstatus.DataV1{Autoconf: "Mr TBird"},
	}
	require.NoError(t, dumpConfigStatusInFile(&data, file))

	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config.Data.Metadata.Version)
	require.Equal(t, "", config.Data.DataV1.Autoconf)
}

func TestConfigStatus_IsPending(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, true, config.IsPending())
	config.Data.DataV1.PendingSince = time.Time{}
	require.Equal(t, false, config.IsPending())
}

func TestConfigStatus_IsFromFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, false, config.IsFromFailure())
	config.Data.DataV1.FailureDetails = "test"
	require.Equal(t, true, config.IsFromFailure())
}

func TestConfigStatus_ApplySuccess(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, true, config.IsPending())
	require.NoError(t, config.ApplySuccess())
	require.Equal(t, false, config.IsPending())

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, true, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config2.Data.DataV1.ClickedLink)
	require.Equal(t, false, config2.Data.DataV1.ReportSent)
	require.Equal(t, false, config2.Data.DataV1.ReportClick)
	require.Equal(t, "", config2.Data.DataV1.FailureDetails)
}

func TestConfigStatus_ApplyFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)
	require.NoError(t, config.ApplySuccess())

	require.NoError(t, config.ApplyFailure("Big Failure"))
	require.Equal(t, true, config.IsFromFailure())
	require.Equal(t, true, config.IsPending())

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, false, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config2.Data.DataV1.ClickedLink)
	require.Equal(t, false, config2.Data.DataV1.ReportSent)
	require.Equal(t, false, config2.Data.DataV1.ReportClick)
	require.Equal(t, "Big Failure", config2.Data.DataV1.FailureDetails)
}

func TestConfigStatus_ApplyProgress(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, true, config.IsPending())
	require.Equal(t, true, config.Data.DataV1.LastProgress.IsZero())

	require.NoError(t, config.ApplyProgress())

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, false, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, false, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config2.Data.DataV1.ClickedLink)
	require.Equal(t, false, config2.Data.DataV1.ReportSent)
	require.Equal(t, false, config2.Data.DataV1.ReportClick)
	require.Equal(t, "", config2.Data.DataV1.FailureDetails)
}

func TestConfigStatus_RecordLinkClicked(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, uint64(0), config.Data.DataV1.ClickedLink)
	require.NoError(t, config.RecordLinkClicked(0))
	require.Equal(t, uint64(1), config.Data.DataV1.ClickedLink)
	require.NoError(t, config.RecordLinkClicked(1))
	require.Equal(t, uint64(3), config.Data.DataV1.ClickedLink)

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, false, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(3), config2.Data.DataV1.ClickedLink)
	require.Equal(t, false, config2.Data.DataV1.ReportSent)
	require.Equal(t, false, config2.Data.DataV1.ReportClick)
	require.Equal(t, "", config2.Data.DataV1.FailureDetails)
}

func TestConfigStatus_ReportClicked(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, false, config.Data.DataV1.ReportClick)
	require.NoError(t, config.ReportClicked())
	require.Equal(t, true, config.Data.DataV1.ReportClick)

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, false, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config2.Data.DataV1.ClickedLink)
	require.Equal(t, false, config2.Data.DataV1.ReportSent)
	require.Equal(t, true, config2.Data.DataV1.ReportClick)
	require.Equal(t, "", config2.Data.DataV1.FailureDetails)
}

func TestConfigStatus_ReportSent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dummy.json")
	config, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, false, config.Data.DataV1.ReportSent)
	require.NoError(t, config.ReportSent())
	require.Equal(t, true, config.Data.DataV1.ReportSent)

	config2, err := configstatus.LoadConfigurationStatus(file)
	require.NoError(t, err)

	require.Equal(t, "1.0.0", config2.Data.Metadata.Version)
	require.Equal(t, false, config2.Data.DataV1.PendingSince.IsZero())
	require.Equal(t, true, config2.Data.DataV1.LastProgress.IsZero())
	require.Equal(t, "", config2.Data.DataV1.Autoconf)
	require.Equal(t, uint64(0), config2.Data.DataV1.ClickedLink)
	require.Equal(t, true, config2.Data.DataV1.ReportSent)
	require.Equal(t, false, config2.Data.DataV1.ReportClick)
	require.Equal(t, "", config2.Data.DataV1.FailureDetails)
}

func dumpConfigStatusInFile(data *configstatus.ConfigurationStatusData, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(data)
}
