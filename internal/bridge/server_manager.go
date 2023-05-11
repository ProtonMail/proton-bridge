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
	"fmt"
	"net"
	"path/filepath"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

// ServerManager manages the IMAP & SMTP servers and their listeners.
type ServerManager struct {
	requests *cpc.CPC

	imapServer   *gluon.Server
	imapListener net.Listener

	smtpServer   *smtp.Server
	smtpListener net.Listener

	loadedUserCount int
}

func newServerManager() *ServerManager {
	return &ServerManager{
		requests: cpc.NewCPC(),
	}
}

func (sm *ServerManager) Init(bridge *Bridge) error {
	imapServer, err := createIMAPServer(bridge)
	if err != nil {
		return err
	}

	smtpServer := createSMTPServer(bridge)

	sm.imapServer = imapServer
	sm.smtpServer = smtpServer

	bridge.tasks.Once(func(ctx context.Context) {
		logging.DoAnnotated(ctx, func(ctx context.Context) {
			sm.run(ctx, bridge)
		}, logging.Labels{
			"service": "server-manager",
		})
	})

	return nil
}

func (sm *ServerManager) CloseServers(ctx context.Context) error {
	defer sm.requests.Close()
	_, err := sm.requests.Send(ctx, &smRequestClose{})

	return err
}

func (sm *ServerManager) RestartIMAP(ctx context.Context) error {
	_, err := sm.requests.Send(ctx, &smRequestRestartIMAP{})

	return err
}

func (sm *ServerManager) RestartSMTP(ctx context.Context) error {
	_, err := sm.requests.Send(ctx, &smRequestRestartSMTP{})

	return err
}

func (sm *ServerManager) AddIMAPUser(ctx context.Context, user *user.User) error {
	_, err := sm.requests.Send(ctx, &smRequestAddIMAPUser{user: user})

	return err
}

func (sm *ServerManager) RemoveIMAPUser(ctx context.Context, user *user.User, withData bool) error {
	_, err := sm.requests.Send(ctx, &smRequestRemoveIMAPUser{
		user:     user,
		withData: withData,
	})

	return err
}

func (sm *ServerManager) SetGluonDir(ctx context.Context, gluonDir string) error {
	_, err := sm.requests.Send(ctx, &smRequestSetGluonDir{
		dir: gluonDir,
	})

	return err
}

func (sm *ServerManager) AddGluonUser(ctx context.Context, conn connector.Connector, passphrase []byte) (string, error) {
	reply, err := cpc.SendTyped[string](ctx, sm.requests, &smRequestAddGluonUser{
		conn:       conn,
		passphrase: passphrase,
	})

	return reply, err
}

func (sm *ServerManager) RemoveGluonUser(ctx context.Context, gluonID string) error {
	_, err := sm.requests.Send(ctx, &smRequestRemoveGluonUser{
		userID: gluonID,
	})

	return err
}

func (sm *ServerManager) run(ctx context.Context, bridge *Bridge) {
	eventCh, cancel := bridge.GetEvents()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			sm.handleClose(ctx, bridge)
			return

		case evt := <-eventCh:
			switch evt.(type) {
			case events.ConnStatusDown:
				logrus.Info("Server Manager, network down stopping listeners")
				if err := sm.closeSMTPServer(bridge); err != nil {
					logrus.WithError(err).Error("Failed to close SMTP server")
				}

				if err := sm.stopIMAPListener(bridge); err != nil {
					logrus.WithError(err)
				}
			case events.ConnStatusUp:
				logrus.Info("Server Manager, network up starting listeners")
				sm.handleLoadedUserCountChange(ctx, bridge)
			}

		case request, ok := <-sm.requests.ReceiveCh():
			if !ok {
				return
			}

			switch r := request.Value().(type) {
			case *smRequestClose:
				sm.handleClose(ctx, bridge)
				request.Reply(ctx, nil, nil)
				return

			case *smRequestRestartSMTP:
				err := sm.restartSMTP(bridge)
				request.Reply(ctx, nil, err)

			case *smRequestRestartIMAP:
				err := sm.restartIMAP(ctx, bridge)
				request.Reply(ctx, nil, err)

			case *smRequestAddIMAPUser:
				err := sm.handleAddIMAPUser(ctx, r.user)
				request.Reply(ctx, nil, err)
				if err == nil {
					sm.loadedUserCount++
					sm.handleLoadedUserCountChange(ctx, bridge)
				}

			case *smRequestRemoveIMAPUser:
				err := sm.handleRemoveIMAPUser(ctx, r.user, r.withData)
				request.Reply(ctx, nil, err)
				if err == nil {
					sm.loadedUserCount--
					sm.handleLoadedUserCountChange(ctx, bridge)
				}

			case *smRequestSetGluonDir:
				err := sm.handleSetGluonDir(ctx, bridge, r.dir)
				request.Reply(ctx, nil, err)

			case *smRequestAddGluonUser:
				id, err := sm.handleAddGluonUser(ctx, r.conn, r.passphrase)
				request.Reply(ctx, id, err)

			case *smRequestRemoveGluonUser:
				err := sm.handleRemoveGluonUser(ctx, r.userID)
				request.Reply(ctx, nil, err)
			}
		}
	}
}

func (sm *ServerManager) handleLoadedUserCountChange(ctx context.Context, bridge *Bridge) {
	logrus.Infof("Validating Listener State %v", sm.loadedUserCount)
	if sm.shouldStartServers() {
		if sm.imapListener == nil {
			if err := sm.serveIMAP(ctx, bridge); err != nil {
				logrus.WithError(err).Error("Failed to start IMAP server")
			}
		}

		if sm.smtpListener == nil {
			if err := sm.restartSMTP(bridge); err != nil {
				logrus.WithError(err).Error("Failed to start SMTP server")
			}
		}
	} else {
		if sm.imapListener != nil {
			if err := sm.stopIMAPListener(bridge); err != nil {
				logrus.WithError(err).Error("Failed to stop IMAP server")
			}
		}

		if sm.smtpListener != nil {
			if err := sm.closeSMTPServer(bridge); err != nil {
				logrus.WithError(err).Error("Failed to stop SMTP server")
			}
		}
	}
}

func (sm *ServerManager) handleClose(ctx context.Context, bridge *Bridge) {
	// Close the IMAP server.
	if err := sm.closeIMAPServer(ctx, bridge); err != nil {
		logrus.WithError(err).Error("Failed to close IMAP server")
	}

	// Close the SMTP server.
	if err := sm.closeSMTPServer(bridge); err != nil {
		logrus.WithError(err).Error("Failed to close SMTP server")
	}
}

func (sm *ServerManager) handleAddIMAPUser(ctx context.Context, user *user.User) error {
	if sm.imapServer == nil {
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
			isNew, err := sm.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey())
			if err != nil {
				return fmt.Errorf("failed to load IMAP user: %w", err)
			}

			if isNew {
				// If the DB was newly created, clear the sync status; gluon's DB was not found.
				logrus.Warn("IMAP user DB was newly created, clearing sync status")

				// Remove the user from IMAP so we can clear the sync status.
				if err := sm.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
					return fmt.Errorf("failed to remove IMAP user: %w", err)
				}

				// Clear the sync status -- we need to resync all messages.
				if err := user.ClearSyncStatus(); err != nil {
					return fmt.Errorf("failed to clear sync status: %w", err)
				}

				// Add the user back to the IMAP server.
				if isNew, err := sm.imapServer.LoadUser(ctx, imapConn, gluonID, user.GluonKey()); err != nil {
					return fmt.Errorf("failed to add IMAP user: %w", err)
				} else if isNew {
					panic("IMAP user should already have a database")
				}
			} else if status := user.GetSyncStatus(); !status.HasLabels {
				// Otherwise, the DB already exists -- if the labels are not yet synced, we need to re-create the DB.
				if err := sm.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
					return fmt.Errorf("failed to remove old IMAP user: %w", err)
				}

				if err := user.RemoveGluonID(addrID, gluonID); err != nil {
					return fmt.Errorf("failed to remove old IMAP user ID: %w", err)
				}

				gluonID, err := sm.imapServer.AddUser(ctx, imapConn, user.GluonKey())
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

			gluonID, err := sm.imapServer.AddUser(ctx, imapConn, user.GluonKey())
			if err != nil {
				return fmt.Errorf("failed to add IMAP user: %w", err)
			}

			if err := user.SetGluonID(addrID, gluonID); err != nil {
				return fmt.Errorf("failed to set IMAP user ID: %w", err)
			}

			log.WithField("gluonID", gluonID).Info("Created new IMAP user")
		}
	}

	// Trigger a sync for the user, if needed.
	user.TriggerSync()

	return nil
}

func (sm *ServerManager) handleRemoveIMAPUser(ctx context.Context, user *user.User, withData bool) error {
	if sm.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	logrus.WithFields(logrus.Fields{
		"userID":   user.ID(),
		"withData": withData,
	}).Debug("Removing IMAP user")

	for addrID, gluonID := range user.GetGluonIDs() {
		if err := sm.imapServer.RemoveUser(ctx, gluonID, withData); err != nil {
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

func createIMAPServer(bridge *Bridge) (*gluon.Server, error) {
	gluonDataDir, err := bridge.GetGluonDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon Database directory: %w", err)
	}

	return newIMAPServer(
		bridge.vault.GetGluonCacheDir(),
		gluonDataDir,
		bridge.curVersion,
		bridge.tlsConfig,
		bridge.reporter,
		bridge.logIMAPClient,
		bridge.logIMAPServer,
		bridge.imapEventCh,
		bridge.tasks,
		bridge.uidValidityGenerator,
		bridge.panicHandler,
	)
}

func createSMTPServer(bridge *Bridge) *smtp.Server {
	return newSMTPServer(bridge, bridge.tlsConfig, bridge.logSMTP)
}

func (sm *ServerManager) closeSMTPServer(bridge *Bridge) error {
	// We close the listener ourselves even though it's also closed by smtpServer.Close().
	// This is because smtpServer.Serve() is called in a separate goroutine and might be executed
	// after we've already closed the server. However, go-smtp has a bug; it blocks on the listener
	// even after the server has been closed. So we close the listener ourselves to unblock it.

	if sm.smtpListener != nil {
		logrus.Info("Closing SMTP Listener")
		if err := sm.smtpListener.Close(); err != nil {
			return fmt.Errorf("failed to close SMTP listener: %w", err)
		}

		sm.smtpListener = nil
	}

	if sm.smtpServer != nil {
		logrus.Info("Closing SMTP server")
		if err := sm.smtpServer.Close(); err != nil {
			logrus.WithError(err).Debug("Failed to close SMTP server (expected -- we close the listener ourselves)")
		}

		sm.smtpServer = nil

		bridge.publish(events.SMTPServerStopped{})
	}

	return nil
}

func (sm *ServerManager) closeIMAPServer(ctx context.Context, bridge *Bridge) error {
	if sm.imapListener != nil {
		logrus.Info("Closing IMAP Listener")

		if err := sm.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}

		sm.imapListener = nil

		bridge.publish(events.IMAPServerStopped{})
	}

	if sm.imapServer != nil {
		logrus.Info("Closing IMAP server")
		if err := sm.imapServer.Close(ctx); err != nil {
			return fmt.Errorf("failed to close IMAP server: %w", err)
		}

		sm.imapServer = nil
	}

	return nil
}

func (sm *ServerManager) restartIMAP(ctx context.Context, bridge *Bridge) error {
	logrus.Info("Restarting IMAP server")

	if sm.imapListener != nil {
		if err := sm.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}

		sm.imapListener = nil

		bridge.publish(events.IMAPServerStopped{})
	}

	if sm.shouldStartServers() {
		return sm.serveIMAP(ctx, bridge)
	}

	return nil
}

func (sm *ServerManager) restartSMTP(bridge *Bridge) error {
	logrus.Info("Restarting SMTP server")

	if err := sm.closeSMTPServer(bridge); err != nil {
		return fmt.Errorf("failed to close SMTP: %w", err)
	}

	bridge.publish(events.SMTPServerStopped{})

	sm.smtpServer = newSMTPServer(bridge, bridge.tlsConfig, bridge.logSMTP)

	if sm.shouldStartServers() {
		return sm.serveSMTP(bridge)
	}

	return nil
}

func (sm *ServerManager) serveSMTP(bridge *Bridge) error {
	port, err := func() (int, error) {
		logrus.WithFields(logrus.Fields{
			"port": bridge.vault.GetSMTPPort(),
			"ssl":  bridge.vault.GetSMTPSSL(),
		}).Info("Starting SMTP server")

		smtpListener, err := newListener(bridge.vault.GetSMTPPort(), bridge.vault.GetSMTPSSL(), bridge.tlsConfig)
		if err != nil {
			return 0, fmt.Errorf("failed to create SMTP listener: %w", err)
		}

		sm.smtpListener = smtpListener

		bridge.tasks.Once(func(context.Context) {
			if err := sm.smtpServer.Serve(smtpListener); err != nil {
				logrus.WithError(err).Info("SMTP server stopped")
			}
		})

		if err := bridge.vault.SetSMTPPort(getPort(smtpListener.Addr())); err != nil {
			return 0, fmt.Errorf("failed to store SMTP port in vault: %w", err)
		}

		return getPort(smtpListener.Addr()), nil
	}()

	if err != nil {
		bridge.publish(events.SMTPServerError{
			Error: err,
		})

		return err
	}

	bridge.publish(events.SMTPServerReady{
		Port: port,
	})

	return nil
}

func (sm *ServerManager) serveIMAP(ctx context.Context, bridge *Bridge) error {
	port, err := func() (int, error) {
		if sm.imapServer == nil {
			return 0, fmt.Errorf("no IMAP server instance running")
		}

		logrus.WithFields(logrus.Fields{
			"port": bridge.vault.GetIMAPPort(),
			"ssl":  bridge.vault.GetIMAPSSL(),
		}).Info("Starting IMAP server")

		imapListener, err := newListener(bridge.vault.GetIMAPPort(), bridge.vault.GetIMAPSSL(), bridge.tlsConfig)
		if err != nil {
			return 0, fmt.Errorf("failed to create IMAP listener: %w", err)
		}

		sm.imapListener = imapListener

		if err := sm.imapServer.Serve(ctx, sm.imapListener); err != nil {
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

func (sm *ServerManager) stopIMAPListener(bridge *Bridge) error {
	logrus.Info("Stopping IMAP listener")
	if sm.imapListener != nil {
		if err := sm.imapListener.Close(); err != nil {
			return err
		}

		sm.imapListener = nil

		bridge.publish(events.IMAPServerStopped{})
	}

	return nil
}

func (sm *ServerManager) handleSetGluonDir(ctx context.Context, bridge *Bridge, newGluonDir string) error {
	return safe.RLockRet(func() error {
		currentGluonDir := bridge.GetGluonCacheDir()
		newGluonDir = filepath.Join(newGluonDir, "gluon")
		if newGluonDir == currentGluonDir {
			return fmt.Errorf("new gluon dir is the same as the old one")
		}

		if err := sm.closeIMAPServer(context.Background(), bridge); err != nil {
			return fmt.Errorf("failed to close IMAP: %w", err)
		}

		sm.loadedUserCount = 0

		if err := bridge.moveGluonCacheDir(currentGluonDir, newGluonDir); err != nil {
			logrus.WithError(err).Error("failed to move GluonCacheDir")

			if err := bridge.vault.SetGluonDir(currentGluonDir); err != nil {
				return fmt.Errorf("failed to revert GluonCacheDir: %w", err)
			}
		}

		bridge.heartbeat.SetCacheLocation(newGluonDir)

		gluonDataDir, err := bridge.GetGluonDataDir()
		if err != nil {
			return fmt.Errorf("failed to get Gluon Database directory: %w", err)
		}

		imapServer, err := newIMAPServer(
			bridge.vault.GetGluonCacheDir(),
			gluonDataDir,
			bridge.curVersion,
			bridge.tlsConfig,
			bridge.reporter,
			bridge.logIMAPClient,
			bridge.logIMAPServer,
			bridge.imapEventCh,
			bridge.tasks,
			bridge.uidValidityGenerator,
			bridge.panicHandler,
		)
		if err != nil {
			return fmt.Errorf("failed to create new IMAP server: %w", err)
		}

		sm.imapServer = imapServer
		for _, bridgeUser := range bridge.users {
			if err := sm.handleAddIMAPUser(ctx, bridgeUser); err != nil {
				return fmt.Errorf("failed to add users to new IMAP server: %w", err)
			}
			sm.loadedUserCount++
		}

		if sm.shouldStartServers() {
			if err := sm.serveIMAP(ctx, bridge); err != nil {
				return fmt.Errorf("failed to serve IMAP: %w", err)
			}
		}

		return nil
	}, bridge.usersLock)
}

func (sm *ServerManager) handleAddGluonUser(ctx context.Context, conn connector.Connector, passphrase []byte) (string, error) {
	if sm.imapServer == nil {
		return "", fmt.Errorf("no imap server instance running")
	}

	return sm.imapServer.AddUser(ctx, conn, passphrase)
}

func (sm *ServerManager) handleRemoveGluonUser(ctx context.Context, userID string) error {
	if sm.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	return sm.imapServer.RemoveUser(ctx, userID, true)
}

func (sm *ServerManager) shouldStartServers() bool {
	return sm.loadedUserCount >= 1
}

type smRequestClose struct{}

type smRequestRestartIMAP struct{}

type smRequestRestartSMTP struct{}

type smRequestAddIMAPUser struct {
	user *user.User
}

type smRequestRemoveIMAPUser struct {
	user     *user.User
	withData bool
}

type smRequestSetGluonDir struct {
	dir string
}

type smRequestAddGluonUser struct {
	conn       connector.Connector
	passphrase []byte
}

type smRequestRemoveGluonUser struct {
	userID string
}
