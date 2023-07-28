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
	"fmt"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
)

type EventProvider interface {
	userevents.Subscribable
	userevents.EventController
}

type Telemetry interface {
	useridentity.Telemetry
	SendConfigStatusSuccess(ctx context.Context)
}

type GluonIDProvider interface {
	GetGluonID(addrID string) (string, bool)
	GetGluonIDs() map[string]string
	SetGluonID(addrID, gluonID string) error
	RemoveGluonID(addrID, gluonID string) error
	GluonKey() []byte
}

type SyncStateProvider interface {
	AddFailedMessageID(messageID string) error
	RemFailedMessageID(messageID string) error
	GetSyncStatus() vault.SyncStatus
	ClearSyncStatus() error
	SetHasLabels(bool) error
	SetHasMessages(bool) error
	SetLastMessageID(messageID string) error
}

type Service struct {
	log *logrus.Entry
	cpc *cpc.CPC

	client        APIClient
	identityState *rwIdentity
	labels        *rwLabels
	addressMode   usertypes.AddressMode

	refreshSubscriber *userevents.RefreshChanneledSubscriber
	addressSubscriber *userevents.AddressChanneledSubscriber
	userSubscriber    *userevents.UserChanneledSubscriber
	messageSubscriber *userevents.MessageChanneledSubscriber
	labelSubscriber   *userevents.LabelChanneledSubscriber

	gluonIDProvider   GluonIDProvider
	syncStateProvider SyncStateProvider
	eventProvider     EventProvider
	serverManager     IMAPServerManager
	eventPublisher    events.EventPublisher

	telemetry    Telemetry
	panicHandler async.PanicHandler
	sendRecorder *sendrecorder.SendRecorder
	reporter     reporter.Reporter

	eventSubscription events.Subscription
	eventWatcher      *watcher.Watcher[events.Event]
	connectors        map[string]*Connector
	maxSyncMemory     uint64
	showAllMail       bool
}

func NewService(
	client APIClient,
	identityState *useridentity.State,
	gluonIDProvider GluonIDProvider,
	syncStateProvider SyncStateProvider,
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
	maxSyncMemory uint64,
	showAllMail bool,
) *Service {
	subscriberName := fmt.Sprintf("imap-%v", identityState.User.ID)

	return &Service{
		cpc: cpc.NewCPC(),
		log: logrus.WithFields(logrus.Fields{
			"user":    identityState.User.ID,
			"service": "imap",
		}),
		client:        client,
		identityState: newRWIdentity(identityState, bridgePassProvider, keyPassProvider),
		labels:        newRWLabels(),
		addressMode:   addressMode,

		gluonIDProvider:   gluonIDProvider,
		serverManager:     serverManager,
		syncStateProvider: syncStateProvider,
		eventProvider:     eventProvider,
		eventPublisher:    eventPublisher,

		refreshSubscriber: userevents.NewRefreshSubscriber(subscriberName),
		addressSubscriber: userevents.NewAddressSubscriber(subscriberName),
		userSubscriber:    userevents.NewUserSubscriber(subscriberName),
		messageSubscriber: userevents.NewMessageSubscriber(subscriberName),
		labelSubscriber:   userevents.NewLabelSubscriber(subscriberName),

		panicHandler: panicHandler,
		sendRecorder: sendRecorder,
		telemetry:    telemetry,
		reporter:     reporter,

		connectors:    make(map[string]*Connector),
		maxSyncMemory: maxSyncMemory,

		eventWatcher:      subscription.Add(events.IMAPServerCreated{}),
		eventSubscription: subscription,
		showAllMail:       showAllMail,
	}
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) error {
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
	_, err := s.cpc.Send(ctx, &cancelSyncReq{})

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

func (s *Service) Close() {
	for _, c := range s.connectors {
		c.StateClose()
	}

	s.connectors = make(map[string]*Connector)
}

func (s *Service) run(ctx context.Context) { //nolint gocyclo
	s.log.Info("Starting IMAP Service")
	defer s.log.Info("Exiting IMAP Service")

	defer s.cpc.Close()
	defer s.eventSubscription.Remove(s.eventWatcher)

	syncHandler := newSyncHandler(ctx, s.panicHandler)
	defer syncHandler.Close()

	syncHandler.launch(s)

	subscription := userevents.Subscription{
		User:     s.userSubscriber,
		Refresh:  s.refreshSubscriber,
		Address:  s.addressSubscriber,
		Labels:   s.labelSubscriber,
		Messages: s.messageSubscriber,
	}

	s.eventProvider.Subscribe(subscription)
	defer s.eventProvider.Unsubscribe(subscription)

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
				err := s.setAddressMode(ctx, syncHandler, r.mode)
				req.Reply(ctx, nil, err)

			case *resyncReq:
				s.log.Info("Received resync request, handling as refresh event")
				err := s.onRefreshEvent(ctx, syncHandler)
				req.Reply(ctx, nil, err)
				s.log.Info("Resync reply sent, handling as refresh event")

			case *cancelSyncReq:
				s.log.Info("Cancelling sync")
				syncHandler.Cancel()
				req.Reply(ctx, nil, nil)

			case *resumeSyncReq:
				s.log.Info("Resuming sync")
				// Cancel previous run, if any, just in case.
				syncHandler.CancelAndWait()
				syncHandler.launch(s)
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

			default:
				s.log.Error("Received unknown request")
			}

		case err, ok := <-syncHandler.OnSyncFinishedCH():
			{
				if !ok {
					continue
				}

				if err != nil {
					s.log.WithError(err).Error("Sync failed")
					continue
				}

				s.log.Info("Sync complete, starting API event stream")
				s.eventProvider.Resume()
			}

		case update, ok := <-syncHandler.updater.ch:
			if !ok {
				continue
			}
			s.onSyncUpdate(ctx, update)

		case e, ok := <-s.userSubscriber.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(user proton.User) error {
				return s.onUserEvent(user)
			})
		case e, ok := <-s.addressSubscriber.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(events []proton.AddressEvent) error {
				return s.onAddressEvent(ctx, events)
			})
		case e, ok := <-s.labelSubscriber.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(events []proton.LabelEvent) error {
				return s.onLabelEvent(ctx, events)
			})
		case e, ok := <-s.messageSubscriber.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(events []proton.MessageEvent) error {
				return s.onMessageEvent(ctx, events)
			})
		case e, ok := <-s.refreshSubscriber.OnEventCh():
			if !ok {
				continue
			}
			e.Consume(func(_ proton.RefreshFlag) error {
				return s.onRefreshEvent(ctx, syncHandler)
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

func (s *Service) onRefreshEvent(ctx context.Context, handler *syncHandler) error {
	s.log.Debug("handling refresh event")

	if err := s.identityState.Write(func(identity *useridentity.State) error {
		return identity.OnRefreshEvent(ctx)
	}); err != nil {
		s.log.WithError(err).Error("Failed to apply refresh event to identity state")
		return err
	}

	handler.CancelAndWait()

	if err := s.removeConnectorsFromServer(ctx, s.connectors, true); err != nil {
		return err
	}

	if err := s.syncStateProvider.ClearSyncStatus(); err != nil {
		return fmt.Errorf("failed to clear sync status:%w", err)
	}

	if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
		return err
	}

	handler.launch(s)

	return nil
}

func (s *Service) onUserEvent(user proton.User) error {
	s.log.Debug("handling user event")
	return s.identityState.Write(func(identity *useridentity.State) error {
		identity.OnUserEvent(user)
		return nil
	})
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

func (s *Service) onSyncUpdate(ctx context.Context, syncUpdate syncUpdate) {
	c, ok := s.connectors[syncUpdate.addrID]
	if !ok {
		s.log.Warningf("Received syncUpdate for unknown addr (%v), connector may have been removed", syncUpdate.addrID)
		syncUpdate.update.Done(fmt.Errorf("undeliverable"))
		return
	}

	c.publishUpdate(ctx, syncUpdate.update)
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

func (s *Service) setAddressMode(ctx context.Context, handler *syncHandler, mode usertypes.AddressMode) error {
	if s.addressMode == mode {
		return nil
	}

	s.addressMode = mode
	if mode == usertypes.AddressModeSplit {
		s.log.Info("Setting Split Address Mode")
	} else {
		s.log.Info("Setting Combined Address Mode")
	}

	handler.CancelAndWait()

	if err := s.removeConnectorsFromServer(ctx, s.connectors, true); err != nil {
		return err
	}

	if err := s.syncStateProvider.ClearSyncStatus(); err != nil {
		return fmt.Errorf("failed to clear sync status:%w", err)
	}

	if err := s.rebuildConnectors(); err != nil {
		return fmt.Errorf("failed to rebuild connectors: %w", err)
	}

	if err := s.addConnectorsToServer(ctx, s.connectors); err != nil {
		return err
	}

	handler.launch(s)

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
