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

package updater

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.1.0", false)

	versionMap := VersionMap{
		"stable": VersionInfo{
			Version:           semver.MustParse("1.5.0"),
			MinAuto:           semver.MustParse("1.4.0"),
			Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
			RolloutProportion: 1.0,
		},
	}

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
	).Return(mustMarshal(t, versionMap), nil)

	version, err := updater.Check()

	assert.Equal(t, semver.MustParse("1.5.0"), version.Version)
	assert.NoError(t, err)
}

func TestCheckEarlyAccess(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.1.0", true)

	versionMap := VersionMap{
		"stable": VersionInfo{
			Version:           semver.MustParse("1.5.0"),
			MinAuto:           semver.MustParse("1.0.0"),
			Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
			RolloutProportion: 1.0,
		},
		"early": VersionInfo{
			Version:           semver.MustParse("1.6.0"),
			MinAuto:           semver.MustParse("1.0.0"),
			Package:           "https://protonmail.com/download/bridge/update_1.6.0_linux.tgz",
			RolloutProportion: 1.0,
		},
	}

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
	).Return(mustMarshal(t, versionMap), nil)

	version, err := updater.Check()

	assert.Equal(t, semver.MustParse("1.6.0"), version.Version)
	assert.NoError(t, err)
}

func TestCheckBadSignature(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.2.0", false)

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
	).Return(nil, errors.New("bad signature"))

	_, err := updater.Check()

	assert.Error(t, err)
}

func TestIsUpdateApplicable(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.4.0", false)

	versionOld := VersionInfo{
		Version:           semver.MustParse("1.3.0"),
		MinAuto:           semver.MustParse("1.3.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.3.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	assert.Equal(t, false, updater.IsUpdateApplicable(versionOld))

	versionEqual := VersionInfo{
		Version:           semver.MustParse("1.4.0"),
		MinAuto:           semver.MustParse("1.3.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.4.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	assert.Equal(t, false, updater.IsUpdateApplicable(versionEqual))

	versionNew := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.3.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	assert.Equal(t, true, updater.IsUpdateApplicable(versionNew))
}

func TestCanInstall(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.4.0", false)

	versionManual := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.5.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	assert.Equal(t, false, updater.CanInstall(versionManual))

	versionAuto := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.3.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	assert.Equal(t, true, updater.CanInstall(versionAuto))
}

func TestInstallUpdate(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.4.0", false)

	latestVersion := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.4.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		latestVersion.Package,
		latestVersion.Package+".sig",
	).Return([]byte("tgz_data_here"), nil)

	err := updater.InstallUpdate(latestVersion)

	assert.NoError(t, err)
}

func TestInstallUpdateBadSignature(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.4.0", false)

	latestVersion := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.4.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		latestVersion.Package,
		latestVersion.Package+".sig",
	).Return(nil, errors.New("bad signature"))

	err := updater.InstallUpdate(latestVersion)

	assert.Error(t, err)
}

func TestInstallUpdateAlreadyOngoing(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	cm := mocks.NewMockManager(c)

	updater := newTestUpdater(cm, "1.4.0", false)

	updater.installer = &fakeInstaller{delay: 2 * time.Second}

	latestVersion := VersionInfo{
		Version:           semver.MustParse("1.5.0"),
		MinAuto:           semver.MustParse("1.4.0"),
		Package:           "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		RolloutProportion: 1.0,
	}

	cm.EXPECT().DownloadAndVerify(
		gomock.Any(),
		latestVersion.Package,
		latestVersion.Package+".sig",
	).Return([]byte("tgz_data_here"), nil)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		assert.NoError(t, updater.InstallUpdate(latestVersion))
		wg.Done()
	}()

	// Wait for the installation to begin.
	time.Sleep(time.Second)

	err := updater.InstallUpdate(latestVersion)
	if assert.Error(t, err) {
		assert.Equal(t, ErrOperationOngoing, err)
	}

	wg.Wait()
}

func newTestUpdater(manager pmapi.Manager, curVer string, earlyAccess bool) *Updater {
	return New(
		manager,
		&fakeInstaller{},
		newFakeSettings(0.5, earlyAccess),
		nil,
		semver.MustParse(curVer),
		"bridge", "linux",
	)
}

type fakeInstaller struct {
	bad   bool
	delay time.Duration
}

func (i *fakeInstaller) InstallUpdate(version *semver.Version, r io.Reader) error {
	if i.bad {
		return errors.New("bad install")
	}

	time.Sleep(i.delay)

	return nil
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	require.NoError(t, err)

	return b
}

type fakeSettings struct {
	*settings.Settings
}

// newFakeSettings creates a temporary folder for files.
func newFakeSettings(rollout float64, earlyAccess bool) *fakeSettings {
	dir, err := ioutil.TempDir("", "test-settings")
	if err != nil {
		panic(err)
	}

	s := &fakeSettings{Settings: settings.New(dir)}

	s.SetFloat64(settings.RolloutKey, rollout)

	if earlyAccess {
		s.Set(settings.UpdateChannelKey, string(EarlyChannel))
	} else {
		s.Set(settings.UpdateChannelKey, string(StableChannel))
	}

	return s
}
