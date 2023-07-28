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

package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	bridgelogging "github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
)

type Telemetry interface {
	useridentity.Telemetry
	ReportSMTPAuthSuccess(context.Context)
	ReportSMTPAuthFailed(username string)
}

type Service struct {
	userID       string
	panicHandler async.PanicHandler
	cpc          *cpc.CPC
	client       *proton.Client
	recorder     *sendrecorder.SendRecorder
	log          *logrus.Entry
	reporter     reporter.Reporter

	bridgePassProvider useridentity.BridgePassProvider
	keyPassProvider    useridentity.KeyPassProvider
	identityState      *useridentity.State
	telemetry          Telemetry

	eventService      userevents.Subscribable
	refreshSubscriber *userevents.RefreshChanneledSubscriber
	addressSubscriber *userevents.AddressChanneledSubscriber
	userSubscriber    *userevents.UserChanneledSubscriber

	addressMode usertypes.AddressMode
}

func NewService(
	userID string,
	client *proton.Client,
	recorder *sendrecorder.SendRecorder,
	handler async.PanicHandler,
	reporter reporter.Reporter,
	bridgePassProvider useridentity.BridgePassProvider,
	keyPassProvider useridentity.KeyPassProvider,
	telemetry Telemetry,
	eventService userevents.Subscribable,
	mode usertypes.AddressMode,
	identityState *useridentity.State,
) *Service {
	subscriberName := fmt.Sprintf("smpt-%v", userID)

	return &Service{
		panicHandler: handler,
		userID:       userID,
		cpc:          cpc.NewCPC(),
		recorder:     recorder,
		log: logrus.WithFields(logrus.Fields{
			"user":    userID,
			"service": "smtp",
		}),
		reporter: reporter,
		client:   client,

		bridgePassProvider: bridgePassProvider,
		keyPassProvider:    keyPassProvider,
		telemetry:          telemetry,
		identityState:      identityState,
		eventService:       eventService,

		refreshSubscriber: userevents.NewRefreshSubscriber(subscriberName),
		userSubscriber:    userevents.NewUserSubscriber(subscriberName),
		addressSubscriber: userevents.NewAddressSubscriber(subscriberName),

		addressMode: mode,
	}
}

func (s *Service) SendMail(ctx context.Context, authID string, from string, to []string, r io.Reader) error {
	_, err := s.cpc.Send(ctx, &sendMailReq{
		authID: authID,
		from:   from,
		to:     to,
		r:      r,
	})

	return err
}

func (s *Service) SetAddressMode(ctx context.Context, mode usertypes.AddressMode) error {
	_, err := s.cpc.Send(ctx, &setAddressModeReq{mode: mode})

	return err
}

func (s *Service) Resync(ctx context.Context) error {
	_, err := s.cpc.Send(ctx, &resyncReq{})

	return err
}

func (s *Service) checkAuth(ctx context.Context, email string, password []byte) (string, error) {
	return cpc.SendTyped[string](ctx, s.cpc, &checkAuthReq{
		email:    email,
		password: password,
	})
}

func (s *Service) Start(ctx context.Context, group *orderedtasks.OrderedCancelGroup) {
	s.log.Debug("Starting service")
	group.Go(ctx, s.userID, "smtp-service", func(ctx context.Context) {
		logging.DoAnnotated(ctx, func(ctx context.Context) {
			s.run(ctx)
		}, logging.Labels{
			"user":    s.userID,
			"service": "smtp",
		})
	})
}

func (s *Service) UserID() string {
	return s.userID
}

func (s *Service) run(ctx context.Context) {
	s.log.Info("Starting service main loop")
	defer s.log.Info("Exiting service main loop")
	defer s.cpc.Close()

	subscription := userevents.Subscription{
		User:    s.userSubscriber,
		Refresh: s.refreshSubscriber,
		Address: s.addressSubscriber,
	}

	s.eventService.Subscribe(subscription)
	defer s.eventService.Unsubscribe(subscription)

	for {
		select {
		case <-ctx.Done():
			return

		case request, ok := <-s.cpc.ReceiveCh():
			if !ok {
				return
			}

			switch r := request.Value().(type) {
			case *sendMailReq:
				s.log.Debug("Received send mail request")
				err := s.sendMail(ctx, r)
				request.Reply(ctx, nil, err)

			case *setAddressModeReq:
				s.log.Debugf("Set address mode %v", r.mode)
				s.addressMode = r.mode
				request.Reply(ctx, nil, nil)

			case *checkAuthReq:
				s.log.WithField("email", bridgelogging.Sensitive(r.email)).Debug("Checking authentication")
				addrID, err := s.identityState.CheckAuth(r.email, r.password, s.bridgePassProvider, s.telemetry)
				request.Reply(ctx, addrID, err)

			case *resyncReq:
				err := s.identityState.OnRefreshEvent(ctx)
				request.Reply(ctx, nil, err)

			default:
				s.log.Error("Received unknown request")
			}
		case e, ok := <-s.userSubscriber.OnEventCh():
			if !ok {
				continue
			}

			s.log.Debug("Handling user event")
			e.Consume(func(user proton.User) error {
				s.identityState.OnUserEvent(user)
				return nil
			})
		case e, ok := <-s.refreshSubscriber.OnEventCh():
			if !ok {
				continue
			}

			s.log.Debug("Handling refresh event")
			e.Consume(func(_ proton.RefreshFlag) error {
				return s.identityState.OnRefreshEvent(ctx)
			})
		case e, ok := <-s.addressSubscriber.OnEventCh():
			if !ok {
				continue
			}

			s.log.Debug("Handling Address Event")
			e.Consume(func(evt []proton.AddressEvent) error {
				s.identityState.OnAddressEvents(evt)
				return nil
			})
		}
	}
}

type sendMailReq struct {
	authID string
	from   string
	to     []string
	r      io.Reader
}

func (s *Service) sendMail(ctx context.Context, req *sendMailReq) error {
	defer async.HandlePanic(s.panicHandler)
	if err := s.smtpSendMail(ctx, req.authID, req.from, req.to, req.r); err != nil {
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) {
			s.log.WithError(apiErr).WithField("Details", apiErr.DetailsToString()).Error("failed to send message")
		}

		return err
	}

	return nil
}

type setAddressModeReq struct {
	mode usertypes.AddressMode
}

type checkAuthReq struct {
	email    string
	password []byte
}

type resyncReq struct{}
