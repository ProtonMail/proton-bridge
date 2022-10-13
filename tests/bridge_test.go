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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
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
	return s.t.bridge.SetIMAPPort(port)
}

func (s *scenario) theUserChangesTheSMTPPortTo(port int) error {
	return s.t.bridge.SetSMTPPort(port)
}

func (s *scenario) theUserSetsTheAddressModeOfTo(user, mode string) error {
	switch mode {
	case "split":
		return s.t.bridge.SetAddressMode(context.Background(), s.t.getUserID(user), vault.SplitMode)

	case "combined":
		return s.t.bridge.SetAddressMode(context.Background(), s.t.getUserID(user), vault.CombinedMode)

	default:
		return fmt.Errorf("unknown address mode %q", mode)
	}
}

func (s *scenario) theUserChangesTheGluonPath() error {
	gluonDir, err := os.MkdirTemp(s.t.dir, "gluon")
	if err != nil {
		return err
	}

	return s.t.bridge.SetGluonDir(context.Background(), gluonDir)
}

func (s *scenario) theUserDeletesTheGluonFiles() error {
	path, err := s.t.locator.ProvideGluonPath()
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
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

func (s *scenario) theUserReportsABug() error {
	return s.t.bridge.ReportBug(context.Background(), "osType", "osVersion", "description", "username", "email", "client", false)
}

func (s *scenario) bridgeSendsAConnectionUpEvent() error {
	return try(s.t.connStatusCh, 5*time.Second, func(event events.Event) error {
		if event, ok := event.(events.ConnStatusUp); !ok {
			return fmt.Errorf("expected connection up event, got %T", event)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAConnectionDownEvent() error {
	return try(s.t.connStatusCh, 5*time.Second, func(event events.Event) error {
		if event, ok := event.(events.ConnStatusDown); !ok {
			return fmt.Errorf("expected connection down event, got %T", event)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsADeauthEventForUser(username string) error {
	return try(s.t.deauthCh, 5*time.Second, func(event events.UserDeauth) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected deauth event for user with ID %s, got %s", wantUserID, event.UserID)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAnAddressCreatedEventForUser(username string) error {
	return try(s.t.addrCreatedCh, 60*time.Second, func(event events.UserAddressCreated) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected user address created event for user with ID %s, got %s", wantUserID, event.UserID)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAnAddressDeletedEventForUser(username string) error {
	return try(s.t.addrDeletedCh, 60*time.Second, func(event events.UserAddressDeleted) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected user address deleted event for user with ID %s, got %s", wantUserID, event.UserID)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsSyncStartedAndFinishedEventsForUser(username string) error {
	if err := get(s.t.syncStartedCh, func(event events.SyncStarted) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected sync started event for user with ID %s, got %s", wantUserID, event.UserID)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to get sync started event: %w", err)
	}

	if err := get(s.t.syncFinishedCh, func(event events.SyncFinished) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected sync finished event for user with ID %s, got %s", wantUserID, event.UserID)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to get sync finished event: %w", err)
	}

	return nil
}

func (s *scenario) bridgeSendsAnUpdateNotAvailableEvent() error {
	return try(s.t.updateCh, 5*time.Second, func(event events.Event) error {
		if event, ok := event.(events.UpdateNotAvailable); !ok {
			return fmt.Errorf("expected update not available event, got %T", event)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAnUpdateAvailableEventForVersion(version string) error {
	return try(s.t.updateCh, 5*time.Second, func(event events.Event) error {
		updateEvent, ok := event.(events.UpdateAvailable)
		if !ok {
			return fmt.Errorf("expected update available event, got %T", event)
		}

		if !updateEvent.CanInstall {
			return errors.New("expected update event to be installable")
		}

		if !updateEvent.Version.Version.Equal(semver.MustParse(version)) {
			return fmt.Errorf("expected update event for version %s, got %s", version, updateEvent.Version.Version)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAManualUpdateEventForVersion(version string) error {
	return try(s.t.updateCh, 5*time.Second, func(event events.Event) error {
		updateEvent, ok := event.(events.UpdateAvailable)
		if !ok {
			return fmt.Errorf("expected manual update event, got %T", event)
		}

		if updateEvent.CanInstall {
			return errors.New("expected update event to not be installable")
		}

		if !updateEvent.Version.Version.Equal(semver.MustParse(version)) {
			return fmt.Errorf("expected update event for version %s, got %s", version, updateEvent.Version.Version)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAnUpdateInstalledEventForVersion(version string) error {
	return try(s.t.updateCh, 5*time.Second, func(event events.Event) error {
		updateEvent, ok := event.(events.UpdateInstalled)
		if !ok {
			return fmt.Errorf("expected update installed event, got %T", event)
		}

		if !updateEvent.Version.Version.Equal(semver.MustParse(version)) {
			return fmt.Errorf("expected update event for version %s, got %s", version, updateEvent.Version.Version)
		}

		return nil
	})
}

func (s *scenario) bridgeSendsAForcedUpdateEvent() error {
	return try(s.t.forcedUpdateCh, 5*time.Second, func(event events.UpdateForced) error {
		return nil
	})
}

func try[T any](inCh *queue.QueuedChannel[T], wait time.Duration, fn func(T) error) error {
	select {
	case event := <-inCh.GetChannel():
		return fn(event)

	case <-time.After(wait):
		return errors.New("timeout waiting for event")
	}
}

func get[T any](inCh *queue.QueuedChannel[T], fn func(T) error) error {
	return fn(<-inCh.GetChannel())
}
