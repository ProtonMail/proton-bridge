package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
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

func (s *scenario) theUserChangesTheGluonPath() error {
	gluonDir, err := os.MkdirTemp(s.t.dir, "gluon")
	if err != nil {
		return err
	}

	return s.t.bridge.SetGluonDir(context.Background(), gluonDir)
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
	return try(s.t.userDeauthCh, 5*time.Second, func(event events.UserDeauth) error {
		if wantUserID := s.t.getUserID(username); wantUserID != event.UserID {
			return fmt.Errorf("expected deauth event for user with ID %s, got %s", wantUserID, event.UserID)
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

func try[T any](inCh <-chan T, wait time.Duration, fn func(T) error) error {
	select {
	case event := <-inCh:
		return fn(event)

	case <-time.After(wait):
		return errors.New("timeout waiting for event")
	}
}

func get[T any](inCh <-chan T, fn func(T) error) error {
	return fn(<-inCh)
}
