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

package imapservice

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type EventProvider interface {
	userevents.Subscribable
	RewindEventID(ctx context.Context, eventID string) error
}

type Telemetry interface {
	useridentity.Telemetry
	SendConfigStatusSuccess(ctx context.Context)
	ReportConfigStatusFailure(errDetails string)
}

type GluonIDProvider interface {
	GetGluonID(addrID string) (string, bool)
	GetGluonIDs() map[string]string
	SetGluonID(addrID, gluonID string) error
	RemoveGluonID(addrID, gluonID string) error
	GluonKey() []byte
}

type Service struct {
	log *logrus.Entry
	cpc *cpc.CPC

	client        APIClient
	identityState *rwIdentity
	labels        *rwLabels
	addressMode   usertypes.AddressMode

	subscription *userevents.EventChanneledSubscriber

	gluonIDProvider GluonIDProvider
	eventProvider   EventProvider
	serverManager   IMAPServerManager
	eventPublisher  events.EventPublisher

	telemetry    Telemetry
	panicHandler async.PanicHandler
	sendRecorder *sendrecorder.SendRecorder
	reporter     reporter.Reporter

	eventSubscription events.Subscription
	eventWatcher      *watcher.Watcher[events.Event]
	connectors        map[string]*Connector
	maxSyncMemory     uint64
	showAllMail       bool

	syncHandler        *syncservice.Handler
	syncUpdateApplier  *SyncUpdateApplier
	syncMessageBuilder *SyncMessageBuilder
	syncStateProvider  *SyncState
	syncReporter       *syncReporter

	syncConfigPath     string
	lastHandledEventID string
	isSyncing          bool
}

func NewService(
	client APIClient,
	identityState *useridentity.State,
	gluonIDProvider GluonIDProvider,
	eventProvider EventProvider,
	serverManager IMAPServerManager,
	eventPublisher events.EventPublisher,
	bridgePassProvider useridentity.BridgePassProvider,
	keyPassProvider useridentity.KeyPassProvider,
	panicHandler async.PanicHandler,
	sendRecorder *sendrecorder.SendRecorder,
	telemetry Telemetry,
	reporter reporter.Reporter,
	addressMode usertypes.AddressMode,
	subscription events.Subscription,
	syncConfigDir string,
	maxSyncMemory uint64,
	showAllMail bool,
) *Service {
	subscriberName := fmt.Sprintf("imap-%v", identityState.User.ID)

	log := logrus.WithFields(logrus.Fields{
		"user":    identityState.User.ID,
		"service": "imap",
	})
	rwIdentity := newRWIdentity(identityState, bridgePassProvider, keyPassProvider)

	syncUpdateApplier := NewSyncUpdateApplier()
	syncMessageBuilder := NewSyncMessageBuilder(rwIdentity)
	syncReporter := newSyncReporter(identityState.User.ID, eventPublisher, time.Second)

	return &Service{
		cpc:           cpc.NewCPC(),
		client:        client,
		log:           log,
		identityState: rwIdentity,
		labels:        newRWLabels(),
		addressMode:   addressMode,

		gluonIDProvider: gluonIDProvider,
		serverManager:   serverManager,
		eventProvider:   eventProvider,
		eventPublisher:  eventPublisher,

		subscription: userevents.NewEventSubscriber(subscriberName),

		panicHandler: panicHandler,
		sendRecorder: sendRecorder,
		telemetry:    telemetry,
		reporter:     reporter,

		connectors:    make(map[string]*Connector),
		maxSyncMemory: maxSyncMemory,

		eventWatcher:      subscription.Add(events.IMAPServerCreated{}),
		eventSubscription: subscription,
		showAllMail:       showAllMail,

		syncUpdateApplier:  syncUpdateApplier,
		syncMessageBuilder: syncMessageBuilder,
		syncReporter:       syncReporter,
		syncConfigPath:     getSyncConfigPath(syncConfigDir, identityState.User.ID),
	}
}

func (s *Service) Start(
	ctx context.Context,
	group *orderedtasks.OrderedCancelGroup,
	syncRegulator syncservice.Regulator,
	lastEventID string,
) error {
	s.lastHandledEventID = lastEventID
	{
		syncStateProvider, err := NewSyncState(s.syncConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load sync state: %w", err)
		}

		s.syncStateProvider = syncStateProvider
	}

	s.syncHandler = syncservice.NewHandler(syncRegulator, s.client, s.identityState.UserID(), s.syncStateProvider, s.log, s.panicHandler)

	// Get user labels
	apiLabels, err := s.client.GetLabels(ctx, proton.LabelTypeSystem, proton.LabelTypeFolder, proton.LabelTypeLabel)
	if err != nil {
		return fmt.Errorf("failed to get labels: %w", err)
	}

	s.labels.SetLabels(apiLabels)

	{
		connectors, err := s.buildConnectors()
		if err != nil {
			s.log.WithError(err).Error("Failed to build connectors")
			return err
		}
		s.connectors = connectors
	}

	if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
		return err
	}

	group.Go(ctx, s.identityState.identity.User.ID, "imap-service", s.run)
	return nil
}

func (s *Service) SetAddressMode(ctx context.Context, mode usertypes.AddressMode) error {
	_, err := s.cpc.Send(ctx, &setAddressModeReq{mode: mode})

	return err
}

func (s *Service) Resync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &resyncReq{})

	return err
}

func (s *Service) CancelSync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &cancelSyncReq{})

	return err
}

func (s *Service) ResumeSync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &resumeSyncReq{})

	return err
}

func (s *Service) OnBadEvent(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &onBadEventReq{})

	return err
}

func (s *Service) OnBadEventResync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &onBadEventResyncReq{})

	return err
}

func (s *Service) OnLogout(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &onLogoutReq{})

	return err
}

func (s *Service) ShowAllMail(ctx context.Context, v bool) error {
	_, err := s.cpc.Send(ctx, &showAllMailReq{v: v})

	return err
}

func (s *Service) GetLabels(ctx context.Context) (map[string]proton.Label, error) {
	return cpc.SendTyped[map[string]proton.Label](ctx, s.cpc, &getLabelsReq{})
}

func (s *Service) GetSyncFailedMessageIDs(ctx context.Context) ([]string, error) {
	return cpc.SendTyped[[]string](ctx, s.cpc, &getSyncFailedMessagesReq{})
}

func (s *Service) Close() {
	for _, c := range s.connectors {
		c.StateClose()
	}

	s.connectors = make(map[string]*Connector)
}

func (s *Service) HandleRefreshEvent(ctx context.Context, _ proton.RefreshFlag) error {
	s.log.Debug("handling refresh event")

	if err := s.identityState.Write(func(identity *useridentity.State) error {
		return identity.OnRefreshEvent(ctx)
	}); err != nil {
		s.log.WithError(err).Error("Failed to apply refresh event to identity state")
		return err
	}

	s.cancelSync()

	if err := s.removeConnectorsFromServer(ctx, s.connectors, true); err != nil {
		return err
	}

	if err := s.syncStateProvider.ClearSyncStatus(ctx); err != nil {
		return fmt.Errorf("failed to clear sync status:%w", err)
	}

	if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
		return err
	}

	s.startSyncing()

	return nil
}

func (s *Service) HandleUserEvent(_ context.Context, user *proton.User) error {
	s.log.Debug("handling user event")

	return s.identityState.Write(func(identity *useridentity.State) error {
		identity.OnUserEvent(*user)
		return nil
	})
}

func (s *Service) run(ctx context.Context) { //nolint gocyclo
	s.log.Info("Starting IMAP Service")
	defer s.log.Info("Exiting IMAP Service")

	defer s.cpc.Close()
	defer s.eventSubscription.Remove(s.eventWatcher)
	defer s.syncHandler.Close()

	s.startSyncing()

	eventHandler := userevents.EventHandler{
		UserHandler:    s,
		AddressHandler: s,
		RefreshHandler: s,
		LabelHandler:   s,
		MessageHandler: s,
	}

	syncEventHandler := s.newSyncEventHandler()

	s.eventProvider.Subscribe(s.subscription)
	defer s.eventProvider.Unsubscribe(s.subscription)

	for {
		select {
		case <-ctx.Done():
			return

		case req, ok := <-s.cpc.ReceiveCh():
			if !ok {
				continue
			}
			switch r := req.Value().(type) {
			case *setAddressModeReq:
				err := s.setAddressMode(ctx, r.mode)
				req.Reply(ctx, nil, err)

			case *resyncReq:
				s.log.Info("Received resync request, handling as refresh event")
				err := s.HandleRefreshEvent(ctx, 0)
				req.Reply(ctx, nil, err)
				s.log.Info("Resync reply sent, handling as refresh event")

			case *cancelSyncReq:
				s.log.Info("Cancelling sync")
				s.syncHandler.Cancel()
				req.Reply(ctx, nil, nil)

			case *resumeSyncReq:
				s.log.Info("Resuming sync")
				// Cancel previous run, if any, just in case.
				s.cancelSync()
				s.startSyncing()
				req.Reply(ctx, nil, nil)
			case *getLabelsReq:
				labels := s.labels.GetLabelMap()
				req.Reply(ctx, labels, nil)

			case *onBadEventReq:
				err := s.removeConnectorsFromServer(ctx, s.connectors, false)
				req.Reply(ctx, nil, err)

			case *onBadEventResyncReq:
				err := s.addConnectorsToServer(ctx, s.connectors)
				req.Reply(ctx, nil, err)

			case *onLogoutReq:
				err := s.removeConnectorsFromServer(ctx, s.connectors, false)
				req.Reply(ctx, nil, err)

			case *showAllMailReq:
				req.Reply(ctx, nil, nil)
				s.setShowAllMail(r.v)

			case *getSyncFailedMessagesReq:
				status, err := s.syncStateProvider.GetSyncStatus(ctx)
				if err != nil {
					req.Reply(ctx, nil, fmt.Errorf("failed to get sync status: %w", err))
					continue
				}

				req.Reply(ctx, maps.Keys(status.FailedMessages), nil)

			default:
				s.log.Error("Received unknown request")
			}

		case err, ok := <-s.syncHandler.OnSyncFinishedCH():
			{
				if !ok {
					continue
				}

				if err != nil {
					s.log.WithError(err).Error("Sync failed")
					continue
				}

				s.log.Info("Sync complete, starting API event stream")
				if err := s.eventProvider.RewindEventID(ctx, s.lastHandledEventID); err != nil {
					if errors.Is(err, context.Canceled) {
						continue
					}

					s.log.WithError(err).Error("Failed to rewind event service")
					s.eventPublisher.PublishEvent(ctx, events.UserBadEvent{
						UserID:     s.identityState.UserID(),
						OldEventID: "",
						NewEventID: "",
						EventInfo:  "",
						Error:      fmt.Errorf("failed to rewind event loop: %w", err),
					})
				}

				s.isSyncing = false
			}

		case request, ok := <-s.syncUpdateApplier.requestCh:
			if !ok {
				continue
			}

			updates, err := request(ctx, s.addressMode, s.connectors)

			if err := s.syncUpdateApplier.reply(ctx, updates, err); err != nil {
				if !errors.Is(err, context.Canceled) {
					s.log.WithError(err).Error("unexpected error during sync update reply")
				}
				return
			}

		case e, ok := <-s.subscription.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(event proton.Event) error {
				if s.isSyncing {
					if err := syncEventHandler.OnEvent(ctx, event); err != nil {
						return err
					}

					// We need to reset the sync if we receive a refresh event during a sync and update
					// the last event id to avoid problems.
					if event.Refresh&proton.RefreshMail != 0 {
						s.lastHandledEventID = event.EventID
					}

					return nil
				}

				if err := eventHandler.OnEvent(ctx, event); err != nil {
					return err
				}

				s.lastHandledEventID = event.EventID

				return nil
			})
		case e, ok := <-s.eventWatcher.GetChannel():
			if !ok {
				continue
			}

			if _, ok := e.(events.IMAPServerCreated); ok {
				if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
					s.log.WithError(err).Error("Failed to add connector to server after created")
				}
			}
		}
	}
}

func (s *Service) buildConnectors() (map[string]*Connector, error) {
	connectors := make(map[string]*Connector)

	if s.addressMode == usertypes.AddressModeCombined {
		addr, err := s.identityState.GetPrimaryAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to build connector for combined mode: %w", err)
		}

		connectors[addr.ID] = NewConnector(
			addr.ID,
			s.client,
			s.labels,
			s.identityState,
			s.addressMode,
			s.sendRecorder,
			s.panicHandler,
			s.telemetry,
			s.showAllMail,
		)

		return connectors, nil
	}

	for _, addr := range s.identityState.GetAddresses() {
		connectors[addr.ID] = NewConnector(
			addr.ID,
			s.client,
			s.labels,
			s.identityState,
			s.addressMode,
			s.sendRecorder,
			s.panicHandler,
			s.telemetry,
			s.showAllMail,
		)
	}

	return connectors, nil
}

func (s *Service) rebuildConnectors() error {
	newConnectors, err := s.buildConnectors()
	if err != nil {
		return err
	}

	for _, c := range s.connectors {
		c.StateClose()
	}

	s.connectors = newConnectors

	return nil
}

func (s *Service) addConnectorsToServer(ctx context.Context, connectors map[string]*Connector) error {
	addedConnectors := make([]string, 0, len(connectors))
	for _, c := range connectors {
		if err := s.serverManager.AddIMAPUser(ctx, c, c.addrID, s.gluonIDProvider, s.syncStateProvider); err != nil {
			s.log.WithError(err).Error("Failed to add connect to imap server")

			if err := s.serverManager.RemoveIMAPUser(ctx, false, s.gluonIDProvider, addedConnectors...); err != nil {
				s.log.WithError(err).Error("Failed to remove previously added connectors after failure")
			}
		}
		addedConnectors = append(addedConnectors, c.addrID)
	}

	return nil
}

func (s *Service) removeConnectorsFromServer(ctx context.Context, connectors map[string]*Connector, deleteData bool) error {
	addrIDs := make([]string, 0, len(connectors))

	for _, c := range connectors {
		addrIDs = append(addrIDs, c.addrID)
	}

	if err := s.serverManager.RemoveIMAPUser(ctx, deleteData, s.gluonIDProvider, addrIDs...); err != nil {
		return fmt.Errorf("failed to remove gluon users from server: %w", err)
	}

	return nil
}

func (s *Service) setAddressMode(ctx context.Context, mode usertypes.AddressMode) error {
	if s.addressMode == mode {
		return nil
	}

	s.addressMode = mode
	if mode == usertypes.AddressModeSplit {
		s.log.Info("Setting Split Address Mode")
	} else {
		s.log.Info("Setting Combined Address Mode")
	}

	s.cancelSync()

	if err := s.removeConnectorsFromServer(ctx, s.connectors, true); err != nil {
		return err
	}

	if err := s.syncStateProvider.ClearSyncStatus(ctx); err != nil {
		return fmt.Errorf("failed to clear sync status:%w", err)
	}

	if err := s.rebuildConnectors(); err != nil {
		return fmt.Errorf("failed to rebuild connectors: %w", err)
	}

	if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
		return err
	}

	s.startSyncing()

	return nil
}

func (s *Service) setShowAllMail(v bool) {
	if s.showAllMail == v {
		return
	}

	s.showAllMail = v

	for _, c := range s.connectors {
		c.ShowAllMail(v)
	}
}

func (s *Service) startSyncing() {
	s.isSyncing = true
	s.syncHandler.Execute(s.syncReporter, s.labels.GetLabelMap(), s.syncUpdateApplier, s.syncMessageBuilder, syncservice.DefaultRetryCoolDown)
}

func (s *Service) cancelSync() {
	s.syncHandler.CancelAndWait()
	s.isSyncing = false
}

type resyncReq struct{}

type cancelSyncReq struct{}

type resumeSyncReq struct{}

type getLabelsReq struct{}

type onBadEventReq struct{}

type onBadEventResyncReq struct{}

type onLogoutReq struct{}

type showAllMailReq struct{ v bool }

type setAddressModeReq struct {
	mode usertypes.AddressMode
}

type getSyncFailedMessagesReq struct{}

func getSyncConfigPath(path string, userID string) string {
	return filepath.Join(path, fmt.Sprintf("sync-%v", userID))
}
