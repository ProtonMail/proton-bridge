// Copyright (c) 2025 Proton AG
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

package bridge_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	bridgePkg "github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater/versioncompare"
	"github.com/elastic/go-sysinfo/types"
	"github.com/stretchr/testify/require"
)

// NOTE: we always assume the highest version is always the first in the release json array

func Test_Update_BetaEligible(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			err := bridge.SetUpdateChannel(updater.EarlyChannel)
			require.NoError(t, err)

			bridge.SetCurrentVersionTest(semver.MustParse("2.1.1"))

			expectedRelease := updater.Release{
				ReleaseCategory:   updater.EarlyAccessReleaseCategory,
				Version:           semver.MustParse("2.1.2"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           &semver.Version{},
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				expectedRelease,
			}}

			go func() {
				time.Sleep(1 * time.Second)
				mocks.Updater.SetLatestVersion(updaterData)
				bridge.CheckForUpdates()
			}()

			select {
			case update := <-updateCh:
				require.Equal(t, events.UpdateInstalled{
					Release: expectedRelease,
					Silent:  true,
				}, update)
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for update")
			}
		})
	})
}

func Test_Update_Stable(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			err := bridge.SetUpdateChannel(updater.StableChannel)
			require.NoError(t, err)

			bridge.SetCurrentVersionTest(semver.MustParse("2.1.1"))

			expectedRelease := updater.Release{
				ReleaseCategory:   updater.StableReleaseCategory,
				Version:           semver.MustParse("2.1.3"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           &semver.Version{},
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.EarlyAccessReleaseCategory,
					Version:           semver.MustParse("2.1.4"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           &semver.Version{},
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedRelease,
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateInstalled{
				Release: expectedRelease,
				Silent:  true,
			}, <-updateCh)
		})
	})
}

func Test_Update_CurrentReleaseNewest(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{})
			defer done()

			err := bridge.SetUpdateChannel(updater.StableChannel)
			require.NoError(t, err)

			bridge.SetCurrentVersionTest(semver.MustParse("2.1.5"))

			expectedRelease := updater.Release{
				ReleaseCategory:   updater.StableReleaseCategory,
				Version:           semver.MustParse("2.1.3"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           &semver.Version{},
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.EarlyAccessReleaseCategory,
					Version:           semver.MustParse("2.1.4"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           &semver.Version{},
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedRelease,
			}}

			mocks.Updater.SetLatestVersion(updaterData)
			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateNotAvailable{}, <-updateCh)
		})
	})
}

func Test_Update_NotRolledOutYet(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetUpdateChannel(updater.EarlyChannel))
			bridge.SetCurrentVersionTest(semver.MustParse("2.0.0"))
			require.NoError(t, bridge.SetRolloutPercentageTest(1.0))

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.1.5"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 0.5,
					MinAuto:           &semver.Version{},
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.1.4"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 0.5,
					MinAuto:           &semver.Version{},
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{})
			defer done()

			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateNotAvailable{}, <-updateCh)
		})
	})
}

func Test_Update_CheckOSVersion_NoUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetAutoUpdate(true))
			require.NoError(t, bridge.SetUpdateChannel(updater.StableChannel))

			currentBridgeVersion := semver.MustParse("2.1.5")
			bridge.SetCurrentVersionTest(currentBridgeVersion)

			// Override the OS version check
			bridge.SetHostVersionGetterTest(func(_ types.Host) string {
				return "10.0.0"
			})

			updateNotAvailableCh, done := bridge.GetEvents(events.UpdateNotAvailable{})
			defer done()

			updateCh, updateChDone := bridge.GetEvents(events.UpdateInstalled{})
			defer updateChDone()

			expectedRelease := updater.Release{
				ReleaseCategory: updater.StableReleaseCategory,
				Version:         semver.MustParse("2.4.0"),
				SystemVersion: versioncompare.SystemVersion{
					Minimum: "12.0.0",
					Maximum: "13.0.0",
				},
				RolloutProportion: 1.0,
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				expectedRelease,
				{
					ReleaseCategory: updater.StableReleaseCategory,
					Version:         semver.MustParse("2.3.0"),
					SystemVersion: versioncompare.SystemVersion{
						Minimum: "10.1.0",
						Maximum: "11.5",
					},
					RolloutProportion: 1.0,
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			if runtime.GOOS == "darwin" {
				require.Equal(t, events.UpdateNotAvailable{}, <-updateNotAvailableCh)
			} else {
				require.Equal(t, events.UpdateInstalled{
					Release: expectedRelease,
					Silent:  true,
				}, <-updateCh)
			}
		})
	})
}

func Test_Update_CheckOSVersion_HasUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetAutoUpdate(true))
			require.NoError(t, bridge.SetUpdateChannel(updater.StableChannel))

			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			currentBridgeVersion := semver.MustParse("2.1.5")
			bridge.SetCurrentVersionTest(currentBridgeVersion)

			// Override the OS version check
			bridge.SetHostVersionGetterTest(func(_ types.Host) string {
				return "10.0.0"
			})

			expectedUpdateRelease := updater.Release{
				ReleaseCategory: updater.StableReleaseCategory,
				Version:         semver.MustParse("2.2.0"),
				SystemVersion: versioncompare.SystemVersion{
					Minimum: "10.0.0",
					Maximum: "10.1.12",
				},
				RolloutProportion: 1.0,
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			expectedUpdateReleaseWindowsLinux := updater.Release{
				ReleaseCategory: updater.StableReleaseCategory,
				Version:         semver.MustParse("2.4.0"),
				SystemVersion: versioncompare.SystemVersion{
					Minimum: "12.0.0",
				},
				RolloutProportion: 1.0,
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				expectedUpdateReleaseWindowsLinux,
				{
					ReleaseCategory: updater.StableReleaseCategory,
					Version:         semver.MustParse("2.3.0"),
					SystemVersion: versioncompare.SystemVersion{
						Minimum: "11.0.0",
					},
					RolloutProportion: 1.0,
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedUpdateRelease,
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.1.0"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			if runtime.GOOS == "darwin" {
				require.Equal(t, events.UpdateInstalled{
					Release: expectedUpdateRelease,
					Silent:  true,
				}, <-updateCh)
			} else {
				require.Equal(t, events.UpdateInstalled{
					Release: expectedUpdateReleaseWindowsLinux,
					Silent:  true,
				}, <-updateCh)
			}
		})
	})
}

func Test_Update_UpdateFromMinVer_UpdateAvailable(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetAutoUpdate(true))
			require.NoError(t, bridge.SetUpdateChannel(updater.StableChannel))

			currentBridgeVersion := semver.MustParse("2.1.5")
			bridge.SetCurrentVersionTest(currentBridgeVersion)

			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			expectedUpdateRelease := updater.Release{
				ReleaseCategory:   updater.StableReleaseCategory,
				Version:           semver.MustParse("2.2.0"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           currentBridgeVersion,
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.3.0"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.2.1"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.2.1"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.2.0"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedUpdateRelease,
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateInstalled{
				Release: expectedUpdateRelease,
				Silent:  true,
			}, <-updateCh)
		})
	})
}

// Test_Update_UpdateFromMinVer_NoCompatibleVersionForceManual -
// if we have an update, but we don't satisfy minVersion, a manual update to the highest possible version should be performed.
func Test_Update_UpdateFromMinVer_NoCompatibleVersionForceManual(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetAutoUpdate(true))
			require.NoError(t, bridge.SetUpdateChannel(updater.StableChannel))

			currentBridgeVersion := semver.MustParse("2.1.5")
			bridge.SetCurrentVersionTest(currentBridgeVersion)

			updateCh, done := bridge.GetEvents(events.UpdateAvailable{})
			defer done()

			expectedUpdateRelease := updater.Release{
				ReleaseCategory:   updater.StableReleaseCategory,
				Version:           semver.MustParse("2.3.0"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           semver.MustParse("2.2.1"),
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.2.1"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.2.0"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				{
					ReleaseCategory:   updater.StableReleaseCategory,
					Version:           semver.MustParse("2.2.0"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.1.6"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedUpdateRelease,
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateAvailable{
				Release:    expectedUpdateRelease,
				Silent:     false,
				Compatible: false,
			}, <-updateCh)
		})
	})
}

// Test_Update_UpdateFromMinVer_NoCompatibleVersionForceManual_BetaMismatch - only Beta updates are available
// nor do we satisfy the minVersion, we can't do anything in this case.
func Test_Update_UpdateFromMinVer_NoCompatibleVersionForceManual_BetaMismatch(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridgePkg.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridgePkg.Bridge, mocks *bridgePkg.Mocks) {
			require.NoError(t, bridge.SetAutoUpdate(true))
			require.NoError(t, bridge.SetUpdateChannel(updater.StableChannel))

			currentBridgeVersion := semver.MustParse("2.1.5")
			bridge.SetCurrentVersionTest(currentBridgeVersion)

			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{})
			defer done()

			expectedUpdateRelease := updater.Release{
				ReleaseCategory:   updater.EarlyAccessReleaseCategory,
				Version:           semver.MustParse("2.3.0"),
				SystemVersion:     versioncompare.SystemVersion{},
				RolloutProportion: 1.0,
				MinAuto:           semver.MustParse("2.2.1"),
				File: []updater.File{
					{
						URL:        "RANDOM_INSTALLER_URL",
						Identifier: updater.InstallerIdentifier,
					},
					{
						URL:        "RANDOM_PACKAGE_URL",
						Identifier: updater.PackageIdentifier,
					},
				},
			}

			updaterData := updater.VersionInfo{Releases: []updater.Release{
				{
					ReleaseCategory:   updater.EarlyAccessReleaseCategory,
					Version:           semver.MustParse("2.2.1"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.2.0"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				{
					ReleaseCategory:   updater.EarlyAccessReleaseCategory,
					Version:           semver.MustParse("2.2.0"),
					SystemVersion:     versioncompare.SystemVersion{},
					RolloutProportion: 1.0,
					MinAuto:           semver.MustParse("2.1.6"),
					File: []updater.File{
						{
							URL:        "RANDOM_INSTALLER_URL",
							Identifier: updater.InstallerIdentifier,
						},
						{
							URL:        "RANDOM_PACKAGE_URL",
							Identifier: updater.PackageIdentifier,
						},
					},
				},
				expectedUpdateRelease,
			}}

			mocks.Updater.SetLatestVersion(updaterData)

			bridge.CheckForUpdates()

			require.Equal(t, events.UpdateNotAvailable{}, <-updateCh)
		})
	})
}
