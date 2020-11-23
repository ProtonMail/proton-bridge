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

package updater

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	versionMap := VersionMap{
		"live": VersionInfo{
			Version: semver.MustParse("1.5.0"),
			MinAuto: semver.MustParse("1.4.0"),
			Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
			Rollout: 1.0,
		},
	}

	client.EXPECT().DownloadAndVerify(
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
		gomock.Any(),
	).Return(bytes.NewReader(mustMarshal(t, versionMap)), nil)

	client.EXPECT().Logout()

	updateCh := make(chan VersionInfo)

	defer updater.Watch(
		time.Minute,
		func(update VersionInfo) error {
			updateCh <- update
			return nil
		},
		func(err error) {
			t.Fatal(err)
		},
	)()

	assert.Equal(t, semver.MustParse("1.5.0"), (<-updateCh).Version)
}

func TestWatchIgnoresCurrentVersion(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.5.0")

	versionMap := VersionMap{
		"live": VersionInfo{
			Version: semver.MustParse("1.5.0"),
			MinAuto: semver.MustParse("1.4.0"),
			Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
			Rollout: 1.0,
		},
	}

	client.EXPECT().DownloadAndVerify(
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
		gomock.Any(),
	).Return(bytes.NewReader(mustMarshal(t, versionMap)), nil)

	client.EXPECT().Logout()

	updateCh := make(chan VersionInfo)

	defer updater.Watch(
		time.Minute,
		func(update VersionInfo) error {
			updateCh <- update
			return nil
		},
		func(err error) {
			t.Fatal(err)
		},
	)()

	select {
	case <-updateCh:
		t.Fatal("We shouldn't update because we are already up to date")
	case <-time.After(1500 * time.Millisecond):
	}
}

func TestWatchIgnoresVerionsThatRequireManualUpdate(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	versionMap := VersionMap{
		"live": VersionInfo{
			Version: semver.MustParse("1.5.0"),
			MinAuto: semver.MustParse("1.5.0"),
			Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
			Rollout: 1.0,
		},
	}

	client.EXPECT().DownloadAndVerify(
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
		gomock.Any(),
	).Return(bytes.NewReader(mustMarshal(t, versionMap)), nil)

	client.EXPECT().Logout()

	updateCh := make(chan VersionInfo)

	defer updater.Watch(
		time.Minute,
		func(update VersionInfo) error {
			updateCh <- update
			return nil
		},
		func(err error) {
			t.Fatal(err)
		},
	)()

	select {
	case <-updateCh:
		t.Fatal("We shouldn't update because this version requires a manual update")
	case <-time.After(1500 * time.Millisecond):
	}
}

func TestWatchBadSignature(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	client.EXPECT().DownloadAndVerify(
		updater.getVersionFileURL(),
		updater.getVersionFileURL()+".sig",
		gomock.Any(),
	).Return(nil, errors.New("bad signature"))

	client.EXPECT().Logout()

	updateCh := make(chan VersionInfo)
	errorsCh := make(chan error)

	defer updater.Watch(
		time.Minute,
		func(update VersionInfo) error {
			updateCh <- update
			return nil
		},
		func(err error) {
			errorsCh <- err
		},
	)()

	assert.Error(t, <-errorsCh)
}

func TestInstallUpdate(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	latestVersion := VersionInfo{
		Version: semver.MustParse("1.5.0"),
		MinAuto: semver.MustParse("1.4.0"),
		Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		Rollout: 1.0,
	}

	client.EXPECT().DownloadAndVerify(
		latestVersion.Package,
		latestVersion.Package+".sig",
		gomock.Any(),
	).Return(bytes.NewReader([]byte("tgz_data_here")), nil)

	client.EXPECT().Logout()

	assert.NoError(t, updater.InstallUpdate(latestVersion))
}

func TestInstallUpdateBadSignature(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	latestVersion := VersionInfo{
		Version: semver.MustParse("1.5.0"),
		MinAuto: semver.MustParse("1.4.0"),
		Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		Rollout: 1.0,
	}

	client.EXPECT().DownloadAndVerify(
		latestVersion.Package,
		latestVersion.Package+".sig",
		gomock.Any(),
	).Return(nil, errors.New("bad signature"))

	client.EXPECT().Logout()

	assert.Error(t, updater.InstallUpdate(latestVersion))
}

func TestInstallUpdateAlreadyOngoing(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	client := mocks.NewMockClient(c)

	updater := newTestUpdater(client, "1.4.0")

	updater.installer = &fakeInstaller{delay: 2 * time.Second}

	latestVersion := VersionInfo{
		Version: semver.MustParse("1.5.0"),
		MinAuto: semver.MustParse("1.4.0"),
		Package: "https://protonmail.com/download/bridge/update_1.5.0_linux.tgz",
		Rollout: 1.0,
	}

	client.EXPECT().DownloadAndVerify(
		latestVersion.Package,
		latestVersion.Package+".sig",
		gomock.Any(),
	).Return(bytes.NewReader([]byte("tgz_data_here")), nil)

	client.EXPECT().Logout()

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

func newTestUpdater(client *mocks.MockClient, curVer string) *Updater {
	return New(
		&fakeClientProvider{client: client},
		&fakeInstaller{},
		nil,
		semver.MustParse(curVer),
		"bridge", "linux",
		0.5,
	)
}

type fakeClientProvider struct {
	client *mocks.MockClient
}

func (p *fakeClientProvider) GetAnonymousClient() pmapi.Client {
	return p.client
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
