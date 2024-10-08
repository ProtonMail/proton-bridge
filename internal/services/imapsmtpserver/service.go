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

package imapsmtpserver

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	bridgesmtp "github.com/ProtonMail/proton-bridge/v3/internal/services/smtp"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

// Service manages the IMAP & SMTP servers and their listeners.
type Service struct {
	requests *cpc.CPC

	imapServer   *gluon.Server
	imapListener net.Listener

	smtpServer   *smtp.Server
	smtpListener net.Listener
	smtpAccounts *bridgesmtp.Accounts

	smtpSettings   SMTPSettingsProvider
	imapSettings   IMAPSettingsProvider
	eventPublisher events.EventPublisher
	panicHandler   async.PanicHandler
	reporter       reporter.Reporter

	log   *logrus.Entry
	tasks *async.Group

	uidValidityGenerator imap.UIDValidityGenerator
	telemetry            Telemetry

	observabilitySender observability.Sender
}

func NewService(
	ctx context.Context,
	smtpSettings SMTPSettingsProvider,
	imapSettings IMAPSettingsProvider,
	eventPublisher events.EventPublisher,
	panicHandler async.PanicHandler,
	reporter reporter.Reporter,
	uidValidityGenerator imap.UIDValidityGenerator,
	telemetry Telemetry,
	observabilitySender observability.Sender,
) *Service {
	return &Service{
		requests:     cpc.NewCPC(),
		smtpAccounts: bridgesmtp.NewAccounts(),

		panicHandler:         panicHandler,
		reporter:             reporter,
		smtpSettings:         smtpSettings,
		imapSettings:         imapSettings,
		eventPublisher:       eventPublisher,
		log:                  logrus.WithField("service", "server-manager"),
		tasks:                async.NewGroup(ctx, panicHandler),
		uidValidityGenerator: uidValidityGenerator,
		telemetry:            telemetry,

		observabilitySender: observabilitySender,
	}
}

func (sm *Service) Init(ctx context.Context, group *async.Group, subscription events.Subscription) error {
	imapServer, err := sm.createIMAPServer(ctx)
	if err != nil {
		return err
	}

	smtpServer := sm.createSMTPServer()

	sm.imapServer = imapServer
	sm.smtpServer = smtpServer

	group.Once(func(ctx context.Context) {
		logging.DoAnnotated(ctx, func(ctx context.Context) {
			sm.run(ctx, subscription)
		}, logging.Labels{
			"service": "server-manager",
		})
	})

	if err := sm.serveIMAP(ctx); err != nil {
		sm.log.WithError(err).Error("Failed to start IMAP server on bridge start")
		sm.imapListener = nil
	}

	if err := sm.serveSMTP(ctx); err != nil {
		sm.log.WithError(err).Error("Failed to start SMTP server on bridge start")
		sm.smtpListener = nil
	}

	return nil
}

func (sm *Service) CloseServers(ctx context.Context) error {
	defer sm.requests.Close()
	_, err := sm.requests.Send(ctx, &smRequestClose{})

	return err
}

func (sm *Service) RestartIMAP(ctx context.Context) error {
	_, err := sm.requests.Send(ctx, &smRequestRestartIMAP{})

	return err
}

func (sm *Service) RestartSMTP(ctx context.Context) error {
	_, err := sm.requests.Send(ctx, &smRequestRestartSMTP{})

	return err
}

func (sm *Service) AddIMAPUser(
	ctx context.Context,
	connector connector.Connector,
	addrID string,
	idProvider imapservice.GluonIDProvider,
	syncStateProvider syncservice.StateProvider,
) error {
	_, err := sm.requests.Send(ctx, &smRequestAddIMAPUser{
		connector:         connector,
		addrID:            addrID,
		idProvider:        idProvider,
		syncStateProvider: syncStateProvider,
	})

	return err
}

func (sm *Service) SetGluonDir(ctx context.Context, gluonDir string) error {
	_, err := sm.requests.Send(ctx, &smRequestSetGluonDir{
		dir: gluonDir,
	})

	return err
}

func (sm *Service) RemoveIMAPUser(ctx context.Context, deleteData bool, provider imapservice.GluonIDProvider, addrID ...string) error {
	_, err := sm.requests.Send(ctx, &smRequestRemoveIMAPUser{
		withData:   deleteData,
		addrID:     addrID,
		idProvider: provider,
	})

	return err
}

func (sm *Service) AddSMTPAccount(ctx context.Context, service *bridgesmtp.Service) error {
	_, err := sm.requests.Send(ctx, &smRequestAddSMTPAccount{account: service})

	return err
}

func (sm *Service) RemoveSMTPAccount(ctx context.Context, service *bridgesmtp.Service) error {
	_, err := sm.requests.Send(ctx, &smRequestRemoveSMTPAccount{account: service})

	return err
}

func (sm *Service) run(ctx context.Context, subscription events.Subscription) {
	eventSub := subscription.Add()
	defer subscription.Remove(eventSub)

	for {
		select {
		case <-ctx.Done():
			sm.handleClose(ctx)
			return

		case evt := <-eventSub.GetChannel():
			switch evt.(type) {
			case events.ConnStatusDown:
				sm.log.Info("Server Manager, network down stopping listeners")
				if err := sm.closeSMTPServer(ctx); err != nil {
					sm.log.WithError(err).Error("Failed to close SMTP server")
				}

				if err := sm.stopIMAPListener(ctx); err != nil {
					sm.log.WithError(err)
				}
			case events.ConnStatusUp:
				sm.log.Info("Server Manager, network up starting listeners")
				sm.handleLoadedUserCountChange(ctx)
			}

		case request, ok := <-sm.requests.ReceiveCh():
			if !ok {
				return
			}

			switch r := request.Value().(type) {
			case *smRequestClose:
				sm.handleClose(ctx)
				request.Reply(ctx, nil, nil)
				return

			case *smRequestRestartSMTP:
				err := sm.restartSMTP(ctx)
				request.Reply(ctx, nil, err)

			case *smRequestRestartIMAP:
				err := sm.restartIMAP(ctx)
				request.Reply(ctx, nil, err)

			case *smRequestAddIMAPUser:
				err := sm.handleAddIMAPUser(ctx, r.connector, r.addrID, r.idProvider, r.syncStateProvider)
				request.Reply(ctx, nil, err)
				if err == nil {
					sm.handleLoadedUserCountChange(ctx)
				}

			case *smRequestRemoveIMAPUser:
				err := sm.handleRemoveIMAPUser(ctx, r.withData, r.idProvider, r.addrID...)
				request.Reply(ctx, nil, err)
				if err == nil {
					sm.handleLoadedUserCountChange(ctx)
				}

			case *smRequestSetGluonDir:
				err := sm.handleSetGluonDir(ctx, r.dir)
				request.Reply(ctx, nil, err)

			case *smRequestAddSMTPAccount:
				sm.log.WithField("user", r.account.UserID()).Debug("Adding SMTP Account")
				sm.smtpAccounts.AddAccount(r.account)
				request.Reply(ctx, nil, nil)

			case *smRequestRemoveSMTPAccount:
				sm.log.WithField("user", r.account.UserID()).Debug("Removing SMTP Account")
				sm.smtpAccounts.RemoveAccount(r.account)
				request.Reply(ctx, nil, nil)
			}
		}
	}
}

func (sm *Service) handleLoadedUserCountChange(ctx context.Context) {
	sm.log.Infof("Validating Listener State")
	if sm.imapListener == nil {
		if err := sm.serveIMAP(ctx); err != nil {
			sm.log.WithError(err).Error("Failed to start IMAP server")
		}
	}

	if sm.smtpListener == nil {
		if err := sm.restartSMTP(ctx); err != nil {
			sm.log.WithError(err).Error("Failed to start SMTP server")
		}
	}
}

func (sm *Service) handleClose(ctx context.Context) {
	// Close the IMAP server.
	if err := sm.closeIMAPServer(ctx); err != nil {
		sm.log.WithError(err).Error("Failed to close IMAP server")
	}

	// Close the SMTP server.
	if err := sm.closeSMTPServer(ctx); err != nil {
		sm.log.WithError(err).Error("Failed to close SMTP server")
	}

	// Cancel and wait needs to be called here since the SMTP server does not have a way to exit
	// the task on context cancellation. Therefor we need to wait here after we issued a close request.
	sm.tasks.CancelAndWait()
}

func (sm *Service) handleAddIMAPUser(ctx context.Context,
	connector connector.Connector,
	addrID string,
	idProvider imapservice.GluonIDProvider,
	syncStateProvider syncservice.StateProvider,
) error {
	// Due to the many different error exits, performer user count change at this stage rather we split the incrementing
	// of users from the logic.
	return sm.handleAddIMAPUserImpl(ctx, connector, addrID, idProvider, syncStateProvider)
}

func (sm *Service) handleAddIMAPUserImpl(ctx context.Context,
	connector connector.Connector,
	addrID string,
	idProvider imapservice.GluonIDProvider,
	syncStateProvider syncservice.StateProvider,
) error {
	if sm.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	log := sm.log.WithFields(logrus.Fields{
		"addrID": addrID,
	})
	log.Info("Adding user to imap server")

	if gluonID, ok := idProvider.GetGluonID(addrID); ok {
		log.WithField("gluonID", gluonID).Info("Loading existing IMAP user")

		// Load the user, checking whether the DB was newly created.
		isNew, err := sm.imapServer.LoadUser(ctx, connector, gluonID, idProvider.GluonKey())
		if err != nil {
			return fmt.Errorf("failed to load IMAP user: %w", err)
		}

		if isNew {
			// If the DB was newly created, clear the sync status; gluon's DB was not found.
			sm.log.Warn("IMAP user DB was newly created, clearing sync status")

			// Remove the user from IMAP so we can clear the sync status.
			if err := sm.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				return fmt.Errorf("failed to remove IMAP user: %w", err)
			}

			// Clear the sync status -- we need to resync all messages.
			if err := syncStateProvider.ClearSyncStatus(ctx); err != nil {
				return fmt.Errorf("failed to clear sync status: %w", err)
			}

			// Add the user back to the IMAP server.
			if isNew, err := sm.imapServer.LoadUser(ctx, connector, gluonID, idProvider.GluonKey()); err != nil {
				return fmt.Errorf("failed to add IMAP user: %w", err)
			} else if isNew {
				panic("IMAP user should already have a database")
			}
		} else {
			status, err := syncStateProvider.GetSyncStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get sync status: %w", err)
			}

			if !status.HasLabels {
				// Otherwise, the DB already exists -- if the labels are not yet synced, we need to re-create the DB.
				if err := sm.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
					return fmt.Errorf("failed to remove old IMAP user: %w", err)
				}

				if err := idProvider.RemoveGluonID(addrID, gluonID); err != nil {
					return fmt.Errorf("failed to remove old IMAP user ID: %w", err)
				}

				gluonID, err := sm.imapServer.AddUser(ctx, connector, idProvider.GluonKey())
				if err != nil {
					return fmt.Errorf("failed to add IMAP user: %w", err)
				}

				if err := idProvider.SetGluonID(addrID, gluonID); err != nil {
					return fmt.Errorf("failed to set IMAP user ID: %w", err)
				}

				log.WithField("gluonID", gluonID).Info("Re-created IMAP user")
			}
		}
	} else {
		log.Info("Creating new IMAP user")

		// GODT-3003: Ensure previous IMAP sync state is cleared if we run into code path after vault corruption.
		if err := syncStateProvider.ClearSyncStatus(ctx); err != nil {
			return fmt.Errorf("failed to reset sync status: %w", err)
		}

		gluonID, err := sm.imapServer.AddUser(ctx, connector, idProvider.GluonKey())
		if err != nil {
			return fmt.Errorf("failed to add IMAP user: %w", err)
		}

		if err := idProvider.SetGluonID(addrID, gluonID); err != nil {
			return fmt.Errorf("failed to set IMAP user ID: %w", err)
		}

		log.WithField("gluonID", gluonID).Info("Created new IMAP user")
	}

	return nil
}

func (sm *Service) handleRemoveIMAPUser(ctx context.Context, withData bool, idProvider imapservice.GluonIDProvider, addrIDs ...string) error {
	if sm.imapServer == nil {
		return fmt.Errorf("no imap server instance running")
	}

	sm.log.WithFields(logrus.Fields{
		"withData":  withData,
		"addresses": addrIDs,
	}).Debug("Removing IMAP user")

	for _, addrID := range addrIDs {
		gluonID, ok := idProvider.GetGluonID(addrID)
		if !ok {
			sm.log.Warnf("Could not find Gluon ID for addrID %v", addrID)
			continue
		}

		if err := sm.imapServer.RemoveUser(ctx, gluonID, withData); err != nil {
			return fmt.Errorf("failed to remove IMAP user: %w", err)
		}

		if withData {
			if err := idProvider.RemoveGluonID(addrID, gluonID); err != nil {
				return fmt.Errorf("failed to remove IMAP user ID: %w", err)
			}
		}
	}

	return nil
}

func (sm *Service) createIMAPServer(ctx context.Context) (*gluon.Server, error) {
	gluonDataDir, err := sm.imapSettings.DataDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon Database directory: %w", err)
	}

	server, err := newIMAPServer(
		sm.imapSettings.CacheDirectory(),
		gluonDataDir,
		sm.imapSettings.Version(),
		sm.imapSettings.TLSConfig(),
		sm.reporter,
		sm.imapSettings.LogClient(),
		sm.imapSettings.LogServer(),
		sm.imapSettings.EventPublisher(),
		sm.tasks,
		sm.uidValidityGenerator,
		sm.panicHandler,
		sm.observabilitySender,
	)
	if err == nil {
		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerCreated{})
	}

	return server, err
}

func (sm *Service) createSMTPServer() *smtp.Server {
	return newSMTPServer(sm.smtpAccounts, sm.smtpSettings)
}

func (sm *Service) closeSMTPServer(ctx context.Context) error {
	// We close the listener ourselves even though it's also closed by smtpServer.Close().
	// This is because smtpServer.Serve() is called in a separate goroutine and might be executed
	// after we've already closed the server. However, go-smtp has a bug; it blocks on the listener
	// even after the server has been closed. So we close the listener ourselves to unblock it.

	if sm.smtpListener != nil {
		sm.log.Info("Closing SMTP Listener")
		if err := sm.smtpListener.Close(); err != nil {
			return fmt.Errorf("failed to close SMTP listener: %w", err)
		}

		sm.smtpListener = nil
	}

	if sm.smtpServer != nil {
		sm.log.Info("Closing SMTP server")
		if err := sm.smtpServer.Close(); err != nil {
			sm.log.WithError(err).Debug("Failed to close SMTP server (expected -- we close the listener ourselves)")
		}

		sm.smtpServer = nil

		sm.eventPublisher.PublishEvent(ctx, events.SMTPServerStopped{})
	}

	return nil
}

func (sm *Service) closeIMAPServer(ctx context.Context) error {
	if sm.imapListener != nil {
		sm.log.Info("Closing IMAP Listener")

		if err := sm.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}

		sm.imapListener = nil

		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerStopped{})
	}

	if sm.imapServer != nil {
		sm.log.Info("Closing IMAP server")
		if err := sm.imapServer.Close(ctx); err != nil {
			return fmt.Errorf("failed to close IMAP server: %w", err)
		}

		sm.imapServer = nil

		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerClosed{})
	}

	return nil
}

func (sm *Service) restartIMAP(ctx context.Context) error {
	sm.log.Info("Restarting IMAP server")

	if sm.imapListener != nil {
		if err := sm.imapListener.Close(); err != nil {
			return fmt.Errorf("failed to close IMAP listener: %w", err)
		}

		sm.imapListener = nil

		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerStopped{})
	}

	return sm.serveIMAP(ctx)
}

func (sm *Service) restartSMTP(ctx context.Context) error {
	sm.log.Info("Restarting SMTP server")

	if err := sm.closeSMTPServer(ctx); err != nil {
		return fmt.Errorf("failed to close SMTP: %w", err)
	}

	sm.eventPublisher.PublishEvent(ctx, events.SMTPServerStopped{})

	sm.smtpServer = newSMTPServer(sm.smtpAccounts, sm.smtpSettings)

	return sm.serveSMTP(ctx)
}

func (sm *Service) serveSMTP(ctx context.Context) error {
	port, err := func() (int, error) {
		sm.log.WithFields(logrus.Fields{
			"port": sm.smtpSettings.Port(),
			"ssl":  sm.smtpSettings.UseSSL(),
		}).Info("Starting SMTP server")

		smtpListener, err := newListener(sm.smtpSettings.Port(), sm.smtpSettings.UseSSL(), sm.smtpSettings.TLSConfig())
		if err != nil {
			return 0, fmt.Errorf("failed to create SMTP listener: %w", err)
		}

		sm.smtpListener = smtpListener

		sm.tasks.Once(func(context.Context) {
			if err := sm.smtpServer.Serve(smtpListener); err != nil {
				sm.log.WithError(err).Info("SMTP server stopped")
			}
		})

		if err := sm.smtpSettings.SetPort(getPort(smtpListener.Addr())); err != nil {
			return 0, fmt.Errorf("failed to store SMTP port in vault: %w", err)
		}

		return getPort(smtpListener.Addr()), nil
	}()

	if err != nil {
		sm.eventPublisher.PublishEvent(ctx, events.SMTPServerError{
			Error: err,
		})

		return err
	}

	sm.eventPublisher.PublishEvent(ctx, events.SMTPServerReady{
		Port: port,
	})

	return nil
}

func (sm *Service) serveIMAP(ctx context.Context) error {
	port, err := func() (int, error) {
		if sm.imapServer == nil {
			return 0, fmt.Errorf("no IMAP server instance running")
		}

		sm.log.WithFields(logrus.Fields{
			"port": sm.imapSettings.Port(),
			"ssl":  sm.imapSettings.UseSSL(),
		}).Info("Starting IMAP server")

		imapListener, err := newListener(sm.imapSettings.Port(), sm.imapSettings.UseSSL(), sm.imapSettings.TLSConfig())
		if err != nil {
			return 0, fmt.Errorf("failed to create IMAP listener: %w", err)
		}

		sm.imapListener = imapListener

		if err := sm.imapServer.Serve(ctx, sm.imapListener); err != nil {
			return 0, fmt.Errorf("failed to serve IMAP: %w", err)
		}

		if err := sm.imapSettings.SetPort(getPort(imapListener.Addr())); err != nil {
			return 0, fmt.Errorf("failed to store IMAP port in vault: %w", err)
		}

		return getPort(imapListener.Addr()), nil
	}()

	if err != nil {
		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerError{
			Error: err,
		})

		return err
	}

	sm.eventPublisher.PublishEvent(ctx, events.IMAPServerReady{
		Port: port,
	})

	return nil
}

func (sm *Service) stopIMAPListener(ctx context.Context) error {
	sm.log.Info("Stopping IMAP listener")
	if sm.imapListener != nil {
		if err := sm.imapListener.Close(); err != nil {
			return err
		}

		sm.imapListener = nil

		sm.eventPublisher.PublishEvent(ctx, events.IMAPServerStopped{})
	}

	return nil
}

func (sm *Service) handleSetGluonDir(ctx context.Context, newGluonDir string) error {
	currentGluonDir := sm.imapSettings.CacheDirectory()
	newGluonDir = filepath.Join(newGluonDir, "gluon")
	if newGluonDir == currentGluonDir {
		return fmt.Errorf("new gluon dir is the same as the old one")
	}

	if err := sm.closeIMAPServer(ctx); err != nil {
		return fmt.Errorf("failed to close IMAP: %w", err)
	}

	if err := moveGluonCacheDir(sm.imapSettings, currentGluonDir, newGluonDir); err != nil {
		sm.log.WithError(err).Error("failed to move GluonCacheDir")

		if err := sm.imapSettings.SetCacheDirectory(currentGluonDir); err != nil {
			return fmt.Errorf("failed to revert GluonCacheDir: %w", err)
		}

		return err
	}

	sm.telemetry.SetCacheLocation(newGluonDir)

	imapServer, err := sm.createIMAPServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new IMAP server: %w", err)
	}

	sm.imapServer = imapServer

	if err := sm.serveIMAP(ctx); err != nil {
		return fmt.Errorf("failed to serve IMAP: %w", err)
	}

	return nil
}

type smRequestClose struct{}

type smRequestRestartIMAP struct{}

type smRequestRestartSMTP struct{}

type smRequestAddIMAPUser struct {
	connector         connector.Connector
	addrID            string
	idProvider        imapservice.GluonIDProvider
	syncStateProvider syncservice.StateProvider
}

type smRequestRemoveIMAPUser struct {
	withData   bool
	addrID     []string
	idProvider imapservice.GluonIDProvider
}

type smRequestSetGluonDir struct {
	dir string
}

type smRequestAddSMTPAccount struct {
	account *bridgesmtp.Service
}

type smRequestRemoveSMTPAccount struct {
	account *bridgesmtp.Service
}
