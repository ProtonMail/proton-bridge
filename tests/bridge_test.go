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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/kb"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/cucumber/godog"
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

type testBugReport struct {
	request bridge.ReportBugReq
	bridge  *bridge.Bridge
}

func newTestBugReport(br *bridge.Bridge) *testBugReport {
	request := bridge.ReportBugReq{
		OSType:      "osType",
		OSVersion:   "osVersion",
		Title:       "title",
		Description: "description",
		Username:    "username",
		Email:       "email@pm.me",
		EmailClient: "client",
		IncludeLogs: false,
	}
	return &testBugReport{
		request: request,
		bridge:  br,
	}
}

func (r *testBugReport) report() error {
	if r.request.IncludeLogs == true {
		data := []byte("Test log file.\n")
		logName := "20231031_122940334_bri_000_v3.6.1+qa_br-178.log"
		logPath, err := r.bridge.GetLogsPath()
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(logPath, logName), data, 0o600); err != nil {
			return err
		}
	}
	return r.bridge.ReportBug(context.Background(), &r.request)
}

func (s *scenario) theUserReportsABug() error {
	return newTestBugReport(s.t.bridge).report()
}

func (s *scenario) theUserReportsABugWithSingleHeaderChange(key, value string) error {
	bugReport := newTestBugReport(s.t.bridge)

	switch key {
	case "osType":
		bugReport.request.OSType = value
	case "osVersion":
		bugReport.request.OSVersion = value
	case "Title":
		bugReport.request.Title = value
	case "Description":
		bugReport.request.Description = value
	case "Username":
		bugReport.request.Username = value
	case "Email":
		bugReport.request.Email = value
	case "Client":
		bugReport.request.EmailClient = value
	case "IncludeLogs":
		att, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse bug report attachment preferences: %w", err)
		}
		bugReport.request.IncludeLogs = att
	default:
		return fmt.Errorf("Wrong header (\"%s\") is being checked", key)
	}

	return bugReport.report()
}

func (s *scenario) theUserReportsABugWithDetails(value *godog.DocString) error {
	bugReport := newTestBugReport(s.t.bridge)
	if err := json.Unmarshal([]byte(value.Content), &bugReport.request); err != nil {
		return fmt.Errorf("cannot parse bug report details: %w", err)
	}
	return bugReport.report()
}

func (s *scenario) theDescriptionInputProvidesKnowledgeBaseArticles(description string, value *godog.DocString) error {
	var wantSuggestions kb.ArticleList
	if err := json.Unmarshal([]byte(value.Content), &wantSuggestions); err != nil {
		return fmt.Errorf("Cannot parse wanted suggestions: %w", err)
	}

	haveSuggestions, err := s.t.bridge.GetKnowledgeBaseSuggestions(description)
	if err != nil {
		return err
	}

	for i := 0; i < len(wantSuggestions); i++ {
		if haveSuggestions[i].URL != wantSuggestions[i].URL {
			return fmt.Errorf("Description \"%v\" has URL: \"%v\", want: \"%v\"", description, haveSuggestions[i].URL, wantSuggestions[i].URL)
		}

		if haveSuggestions[i].Title != wantSuggestions[i].Title {
			return fmt.Errorf("Description \"%v\" has Title: \"%v\", want: \"%v\"", description, haveSuggestions[i].Title, wantSuggestions[i].Title)
		}
	}

	return nil
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
	wantUserID := s.t.getUserByName(username).getUserID()
	for {
		startEvent, ok := awaitType(s.t.events, events.SyncStarted{}, 30*time.Second)
		if !ok {
			return errors.New("expected sync started event, got none")
		}

		// There can be multiple sync events, and some might not be for this user
		if startEvent.UserID != wantUserID {
			continue
		}

		break
	}
	for {
		finishEvent, ok := awaitType(s.t.events, events.SyncFinished{}, 30*time.Second)
		if !ok {
			return errors.New("expected sync finished event, got none")
		}

		if wantUserID := s.t.getUserByName(username).getUserID(); finishEvent.UserID == wantUserID {
			return nil
		}
	}
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
	return eventually(func() error {
		res := s.t.bridge.IsTelemetryAvailable(context.Background())
		if res != expect {
			return fmt.Errorf("expected telemetry feature %v but got %v ", expect, res)
		}
		return nil
	})
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
