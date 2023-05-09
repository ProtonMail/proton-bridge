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

package tests

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/golang/mock/gomock"
)

func (s *scenario) bridgeStarts() error {
	return s.t.startBridge()
}

func (s *scenario) bridgeRestarts() error {
	if err := s.t.stopBridge(); err != nil {
		return err
	}

	return s.t.startBridge()
}

func (s *scenario) bridgeStops() error {
	return s.t.stopBridge()
}

func (s *scenario) bridgeVersionIsAndTheLatestAvailableVersionIsReachableFrom(current, latest, minAuto string) error {
	s.t.version = semver.MustParse(current)
	s.t.mocks.Updater.SetLatestVersion(semver.MustParse(latest), semver.MustParse(minAuto))
	return nil
}

func (s *scenario) theAPIRequiresBridgeVersion(version string) error {
	s.t.api.SetMinAppVersion(semver.MustParse(version))
	return nil
}

func (s *scenario) theUserChangesTheIMAPPortTo(port int) error {
	return s.t.bridge.SetIMAPPort(context.Background(), port)
}

func (s *scenario) theUserChangesTheSMTPPortTo(port int) error {
	return s.t.bridge.SetSMTPPort(context.Background(), port)
}

func (s *scenario) theUserSetsTheAddressModeOfUserTo(user, mode string) error {
	switch mode {
	case "split":
		return s.t.bridge.SetAddressMode(context.Background(), s.t.getUserByName(user).getUserID(), vault.SplitMode)

	case "combined":
		return s.t.bridge.SetAddressMode(context.Background(), s.t.getUserByName(user).getUserID(), vault.CombinedMode)

	default:
		return fmt.Errorf("unknown address mode %q", mode)
	}
}

func (s *scenario) theUserChangesTheDefaultKeychainApplication() error {
	return s.t.bridge.SetKeychainApp("CustomKeychainApp")
}

func (s *scenario) theUserChangesTheGluonPath() error {
	gluonDir, err := os.MkdirTemp(s.t.dir, "gluon")
	if err != nil {
		return err
	}

	return s.t.bridge.SetGluonDir(context.Background(), gluonDir)
}

func (s *scenario) theUserDeletesTheGluonFiles() error {
	if path, err := s.t.locator.ProvideGluonDataPath(); err != nil {
		return fmt.Errorf("failed to get gluon Data path: %w", err)
	} else if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove gluon Data path: %w", err)
	}

	return s.theUserDeletesTheGluonCache()
}

func (s *scenario) theUserDeletesTheGluonCache() error {
	if path, err := s.t.locator.ProvideGluonDataPath(); err != nil {
		return fmt.Errorf("failed to get gluon Cache path: %w", err)
	} else if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove gluon Cache path: %w", err)
	}
	return nil
}

func (s *scenario) theUserHasDisabledAutomaticUpdates() error {
	var started bool

	if s.t.bridge == nil {
		if err := s.t.startBridge(); err != nil {
			return err
		}

		started = true
	}
	if err := s.t.bridge.SetAutoUpdate(false); err != nil {
		return err
	}

	if started {
		if err := s.t.stopBridge(); err != nil {
			return err
		}
	}
	return nil
}

func (s *scenario) theUserHasDisabledAutomaticStart() error {
	return s.t.bridge.SetAutostart(false)
}

func (s *scenario) theUserHasEnabledAlternativeRouting() error {
	s.t.expectProxyCtlAllowProxy()
	return s.t.bridge.SetProxyAllowed(true)
}

func (s *scenario) theUserSetIMAPModeToSSL() error {
	return s.t.bridge.SetIMAPSSL(context.Background(), true)
}

func (s *scenario) theUserSetSMTPModeToSSL() error {
	return s.t.bridge.SetSMTPSSL(context.Background(), true)
}

func (s *scenario) theUserReportsABug() error {
	return s.t.bridge.ReportBug(context.Background(), "osType", "osVersion", "description", "username", "email", "client", false)
}

func (s *scenario) bridgeSendsAConnectionUpEvent() error {
	if event := s.t.events.await(events.ConnStatusUp{}, 30*time.Second); event == nil {
		return errors.New("expected connection up event, got none")
	}

	return nil
}

func (s *scenario) bridgeSendsAConnectionDownEvent() error {
	if event := s.t.events.await(events.ConnStatusDown{}, 30*time.Second); event == nil {
		return errors.New("expected connection down event, got none")
	}

	return nil
}

func (s *scenario) bridgeSendsADeauthEventForUser(username string) error {
	event, ok := awaitType(s.t.events, events.UserDeauth{}, 30*time.Second)
	if !ok {
		return errors.New("expected deauth event, got none")
	}

	if wantUserID := s.t.getUserByName(username).getUserID(); event.UserID != wantUserID {
		return fmt.Errorf("expected deauth event for user %s, got %s", wantUserID, event.UserID)
	}

	return nil
}

func (s *scenario) bridgeSendsAnAddressCreatedEventForUser(username string) error {
	event, ok := awaitType(s.t.events, events.UserAddressCreated{}, 30*time.Second)
	if !ok {
		return errors.New("expected address created event, got none")
	}

	if wantUserID := s.t.getUserByName(username).getUserID(); event.UserID != wantUserID {
		return fmt.Errorf("expected address created event for user %s, got %s", wantUserID, event.UserID)
	}

	return nil
}

func (s *scenario) bridgeSendsAnAddressDeletedEventForUser(username string) error {
	event, ok := awaitType(s.t.events, events.UserAddressDeleted{}, 30*time.Second)
	if !ok {
		return errors.New("expected address deleted event, got none")
	}

	if wantUserID := s.t.getUserByName(username).getUserID(); event.UserID != wantUserID {
		return fmt.Errorf("expected address deleted event for user %s, got %s", wantUserID, event.UserID)
	}

	return nil
}

func (s *scenario) bridgeSendsSyncStartedAndFinishedEventsForUser(username string) error {
	startEvent, ok := awaitType(s.t.events, events.SyncStarted{}, 30*time.Second)
	if !ok {
		return errors.New("expected sync started event, got none")
	}

	if wantUserID := s.t.getUserByName(username).getUserID(); startEvent.UserID != wantUserID {
		return fmt.Errorf("expected sync started event for user %s, got %s", wantUserID, startEvent.UserID)
	}

	finishEvent, ok := awaitType(s.t.events, events.SyncFinished{}, 30*time.Second)
	if !ok {
		return errors.New("expected sync finished event, got none")
	}

	if wantUserID := s.t.getUserByName(username).getUserID(); finishEvent.UserID != wantUserID {
		return fmt.Errorf("expected sync finished event for user %s, got %s", wantUserID, finishEvent.UserID)
	}

	return nil
}

func (s *scenario) bridgeSendsAnUpdateNotAvailableEvent() error {
	if event := s.t.events.await(events.UpdateNotAvailable{}, 30*time.Second); event == nil {
		return errors.New("expected update not available event, got none")
	}

	return nil
}

func (s *scenario) bridgeSendsAnUpdateAvailableEventForVersion(version string) error {
	event, ok := awaitType(s.t.events, events.UpdateAvailable{}, 30*time.Second)
	if !ok {
		return errors.New("expected update available event, got none")
	}

	if !event.Compatible {
		return errors.New("expected update event to be installable")
	}

	if !event.Version.Version.Equal(semver.MustParse(version)) {
		return fmt.Errorf("expected update event for version %s, got %s", version, event.Version.Version)
	}

	return nil
}

func (s *scenario) bridgeSendsAManualUpdateEventForVersion(version string) error {
	event, ok := awaitType(s.t.events, events.UpdateAvailable{}, 30*time.Second)
	if !ok {
		return errors.New("expected update available event, got none")
	}

	if event.Compatible {
		return errors.New("expected update event to not be installable")
	}

	if !event.Version.Version.Equal(semver.MustParse(version)) {
		return fmt.Errorf("expected update event for version %s, got %s", version, event.Version.Version)
	}

	return nil
}

func (s *scenario) bridgeSendsAnUpdateInstalledEventForVersion(version string) error {
	event, ok := awaitType(s.t.events, events.UpdateInstalled{}, 30*time.Second)
	if !ok {
		return errors.New("expected update installed event, got none")
	}

	if !event.Version.Version.Equal(semver.MustParse(version)) {
		return fmt.Errorf("expected update installed event for version %s, got %s", version, event.Version.Version)
	}

	return nil
}

func (s *scenario) bridgeSendsAForcedUpdateEvent() error {
	if event := s.t.events.await(events.UpdateForced{}, 30*time.Second); event == nil {
		return errors.New("expected update forced event, got none")
	}

	return nil
}

func (s *scenario) bridgeReportsMessage(message string) error {
	s.t.reporter.removeMatchingRecords(
		gomock.Eq(false),
		gomock.Eq(message),
		gomock.Any(),
		1,
	)
	return nil
}

func (s *scenario) bridgeTelemetryFeatureEnabled() error {
	return s.checkTelemetry(true)
}

func (s *scenario) bridgeTelemetryFeatureDisabled() error {
	return s.checkTelemetry(false)
}

func (s *scenario) checkTelemetry(expect bool) error {
	res := s.t.bridge.IsTelemetryAvailable()
	if res != expect {
		return fmt.Errorf("expected telemetry feature %v but got %v ", expect, res)
	}
	return nil
}

func (s *scenario) theUserHidesAllMail() error {
	return s.t.bridge.SetShowAllMail(false)
}

func (s *scenario) theUserShowsAllMail() error {
	return s.t.bridge.SetShowAllMail(true)
}

func (s *scenario) theUserDisablesTelemetryInBridgeSettings() error {
	return s.t.bridge.SetTelemetryDisabled(true)
}

func (s *scenario) theUserEnablesTelemetryInBridgeSettings() error {
	return s.t.bridge.SetTelemetryDisabled(false)
}

func (s *scenario) networkPortIsBusy(port int) {
	if listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port)); err == nil { // we ignore errors. Most likely port is already busy.
		s.t.dummyListeners = append(s.t.dummyListeners, listener)
	}
}

func (s *scenario) networkPortRangeIsBusy(startPort, endPort int) {
	if startPort > endPort {
		startPort, endPort = endPort, startPort
	}

	for port := startPort; port <= endPort; port++ {
		s.networkPortIsBusy(port)
	}
}

func (s *scenario) bridgeIMAPPortIs(expectedPort int) error {
	actualPort := s.t.bridge.GetIMAPPort()
	if actualPort != expectedPort {
		return fmt.Errorf("expected IMAP port to be %v but got %v", expectedPort, actualPort)
	}

	return nil
}

func (s *scenario) bridgeSMTPPortIs(expectedPort int) error {
	actualPort := s.t.bridge.GetSMTPPort()
	if actualPort != expectedPort {
		return fmt.Errorf("expected SMTP port to be %v but got %v", expectedPort, actualPort)
	}

	return nil
}
