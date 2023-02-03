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

package bridge

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/proton-bridge/v3/internal/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

const (
	defaultClientName    = "UnknownClient"
	defaultClientVersion = "0.0.1"
)

func (bridge *Bridge) serveIMAP() error {
	port, err := func() (int, error) {
		if bridge.imapServer == nil {
			return 0, fmt.Errorf("no IMAP server instance running")
		}

		logrus.Info("Starting IMAP server")

		imapListener, err := newListener(bridge.vault.GetIMAPPort(), bridge.vault.GetIMAPSSL(), bridge.tlsConfig)
		if err != nil {
			return 0, fmt.Errorf("failed to create IMAP listener: %w", err)
		}

		bridge.imapListener = imapListener

		if err := bridge.imapServer.Serve(context.Background(), bridge.imapListener); err != nil {
			return 0, fmt.Errorf("failed to serve IMAP: %w", err)
		}

		if err := bridge.vault.SetIMAPPort(getPort(imapListener.Addr())); err != nil {
			return 0, fmt.Errorf("failed to store IMAP port in vault: %w", err)
		}

		return getPort(imapListener.Addr()), nil
	}()

	if err != nil {
		bridge.publish(events.IMAPServerError{
			Error: err,
		})

		return err
	}

	bridge.publish(events.IMAPServerReady{
		Port: port,
	})

	return nil
}

func (bridge *Bridge) restartIMAP() error {
	logrus.Info("Restarting IMAP server")

	if bridge.imapListener != nil {
		if err := bridge.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}

		bridge.publish(events.IMAPServerStopped{})
	}

	return bridge.serveIMAP()
}

func (bridge *Bridge) closeIMAP(ctx context.Context) error {
	logrus.Info("Closing IMAP server")

	if bridge.imapServer != nil {
		if err := bridge.imapServer.Close(ctx); err != nil {
			return fmt.Errorf("failed to close IMAP server: %w", err)
		}

		bridge.imapServer = nil
	}

	if bridge.imapListener != nil {
		if err := bridge.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}
	}

	bridge.publish(events.IMAPServerStopped{})

	return nil
}

// addIMAPUser connects the given user to gluon.
func (bridge *Bridge) addIMAPUser(ctx context.Context, user *user.User) error {
	if bridge.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	imapConn, err := user.NewIMAPConnectors()
	if err != nil {
		return fmt.Errorf("failed to create IMAP connectors: %w", err)
	}

	for addrID, imapConn := range imapConn {
		log := logrus.WithFields(logrus.Fields{
			"userID": user.ID(),
			"addrID": addrID,
		})

		if gluonID, ok := user.GetGluonID(addrID); ok {
			log.WithField("gluonID", gluonID).Info("Loading existing IMAP user")

			// Load the user, checking whether the DB was newly created.
			isNew, err := bridge.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey())
			if err != nil {
				return fmt.Errorf("failed to load IMAP user: %w", err)
			}

			if isNew {
				// If the DB was newly created, clear the sync status; gluon's DB was not found.
				logrus.Warn("IMAP user DB was newly created, clearing sync status")

				if err := user.ClearSyncStatus(); err != nil {
					return fmt.Errorf("failed to clear sync status: %w", err)
				}
			} else if status := user.GetSyncStatus(); !status.HasLabels {
				// Otherwise, the DB already exists -- if the labels are not yet synced, we need to re-create the DB.
				if err := bridge.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
					return fmt.Errorf("failed to remove old IMAP user: %w", err)
				}

				if err := user.RemoveGluonID(addrID, gluonID); err != nil {
					return fmt.Errorf("failed to remove old IMAP user ID: %w", err)
				}

				gluonID, err := bridge.imapServer.AddUser(ctx, imapConn, user.GluonKey())
				if err != nil {
					return fmt.Errorf("failed to add IMAP user: %w", err)
				}

				if err := user.SetGluonID(addrID, gluonID); err != nil {
					return fmt.Errorf("failed to set IMAP user ID: %w", err)
				}

				log.WithField("gluonID", gluonID).Info("Re-created IMAP user")
			}
		} else {
			log.Info("Creating new IMAP user")

			gluonID, err := bridge.imapServer.AddUser(ctx, imapConn, user.GluonKey())
			if err != nil {
				return fmt.Errorf("failed to add IMAP user: %w", err)
			}

			if err := user.SetGluonID(addrID, gluonID); err != nil {
				return fmt.Errorf("failed to set IMAP user ID: %w", err)
			}

			log.WithField("gluonID", gluonID).Info("Created new IMAP user")
		}
	}

	user.TriggerSync()
	return nil
}

// removeIMAPUser disconnects the given user from gluon, optionally also removing its files.
func (bridge *Bridge) removeIMAPUser(ctx context.Context, user *user.User, withData bool) error {
	if bridge.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	logrus.WithFields(logrus.Fields{
		"userID":   user.ID(),
		"withData": withData,
	}).Debug("Removing IMAP user")

	for addrID, gluonID := range user.GetGluonIDs() {
		if err := bridge.imapServer.RemoveUser(ctx, gluonID, withData); err != nil {
			return fmt.Errorf("failed to remove IMAP user: %w", err)
		}

		if withData {
			if err := user.RemoveGluonID(addrID, gluonID); err != nil {
				return fmt.Errorf("failed to remove IMAP user ID: %w", err)
			}
		}
	}

	return nil
}

func (bridge *Bridge) handleIMAPEvent(event imapEvents.Event) {
	switch event := event.(type) {
	case imapEvents.UserAdded:
		for labelID, count := range event.Counts {
			logrus.WithFields(logrus.Fields{
				"gluonID": event.UserID,
				"labelID": labelID,
				"count":   count,
			}).Info("Received mailbox message count")
		}

	case imapEvents.SessionAdded:
		if !bridge.identifier.HasClient() {
			bridge.identifier.SetClient(defaultClientName, defaultClientVersion)
		}

	case imapEvents.IMAPID:
		logrus.WithFields(logrus.Fields{
			"sessionID": event.SessionID,
			"name":      event.IMAPID.Name,
			"version":   event.IMAPID.Version,
		}).Info("Received IMAP ID")

		if event.IMAPID.Name != "" && event.IMAPID.Version != "" {
			bridge.identifier.SetClient(event.IMAPID.Name, event.IMAPID.Version)
		}

	case imapEvents.LoginFailed:
		logrus.WithFields(logrus.Fields{
			"sessionID": event.SessionID,
			"username":  event.Username,
		}).Info("Received IMAP login failure notification")
		bridge.publish(events.IMAPLoginFailed{Username: event.Username})
	}
}

func getGluonDir(encVault *vault.Vault) (string, error) {
	if err := os.MkdirAll(encVault.GetGluonCacheDir(), 0o700); err != nil {
		return "", fmt.Errorf("failed to create gluon dir: %w", err)
	}

	return encVault.GetGluonCacheDir(), nil
}

func ApplyGluonCachePathSuffix(basePath string) string {
	return filepath.Join(basePath, "backend", "store")
}

func ApplyGluonConfigPathSuffix(basePath string) string {
	return filepath.Join(basePath, "backend", "db")
}

func newIMAPServer(
	gluonCacheDir, gluonConfigDir string,
	version *semver.Version,
	tlsConfig *tls.Config,
	reporter reporter.Reporter,
	logClient, logServer bool,
	eventCh chan<- imapEvents.Event,
	tasks *async.Group,
	uidValidityGenerator imap.UIDValidityGenerator,
) (*gluon.Server, error) {
	gluonCacheDir = ApplyGluonCachePathSuffix(gluonCacheDir)
	gluonConfigDir = ApplyGluonConfigPathSuffix(gluonConfigDir)

	logrus.WithFields(logrus.Fields{
		"gluonStore": gluonCacheDir,
		"gluonDB":    gluonConfigDir,
		"version":    version,
		"logClient":  logClient,
		"logServer":  logServer,
	}).Info("Creating IMAP server")

	if logClient || logServer {
		log := logrus.WithField("protocol", "IMAP")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	var imapClientLog io.Writer

	if logClient {
		imapClientLog = logging.NewIMAPLogger()
	} else {
		imapClientLog = io.Discard
	}

	var imapServerLog io.Writer

	if logServer {
		imapServerLog = logging.NewIMAPLogger()
	} else {
		imapServerLog = io.Discard
	}

	imapServer, err := gluon.New(
		gluon.WithTLS(tlsConfig),
		gluon.WithDataDir(gluonCacheDir),
		gluon.WithDatabaseDir(gluonConfigDir),
		gluon.WithStoreBuilder(new(storeBuilder)),
		gluon.WithLogger(imapClientLog, imapServerLog),
		getGluonVersionInfo(version),
		gluon.WithReporter(reporter),
		gluon.WithUIDValidityGenerator(uidValidityGenerator),
	)
	if err != nil {
		return nil, err
	}

	tasks.Once(func(ctx context.Context) {
		async.ForwardContext(ctx, eventCh, imapServer.AddWatcher())
	})

	tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, imapServer.GetErrorCh(), func(err error) {
			logrus.WithError(err).Error("IMAP server error")
		})
	})

	return imapServer, nil
}

func getGluonVersionInfo(version *semver.Version) gluon.Option {
	return gluon.WithVersionInfo(
		int(version.Major()),
		int(version.Minor()),
		int(version.Patch()),
		constants.FullAppName,
		"TODO",
		"TODO",
	)
}

type storeBuilder struct{}

func (*storeBuilder) New(path, userID string, passphrase []byte) (store.Store, error) {
	return store.NewOnDiskStore(
		filepath.Join(path, userID),
		passphrase,
	)
}

func (*storeBuilder) Delete(path, userID string) error {
	return os.RemoveAll(filepath.Join(path, userID))
}
