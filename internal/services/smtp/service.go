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
	"io"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/cpc"
	"github.com/sirupsen/logrus"
)

// UserInterface is just wrapper to avoid recursive go module imports. To be removed when the identity service is ready.
type UserInterface interface {
	ID() string
	CheckAuth(string, []byte) (string, error)
	WithSMTPData(context.Context, func(context.Context, map[string]proton.Address, proton.User, *vault.User) error) error
	ReportSMTPAuthSuccess(context.Context)
	ReportSMTPAuthFailed(username string)
}

type Service struct {
	panicHandler async.PanicHandler
	cpc          *cpc.CPC
	user         UserInterface
	client       *proton.Client
	recorder     *sendrecorder.SendRecorder
	log          *logrus.Entry
	reporter     reporter.Reporter
}

func NewService(
	user UserInterface,
	client *proton.Client,
	recorder *sendrecorder.SendRecorder,
	handler async.PanicHandler,
	reporter reporter.Reporter,
) *Service {
	return &Service{
		panicHandler: handler,
		user:         user,
		cpc:          cpc.NewCPC(),
		recorder:     recorder,
		log: logrus.WithFields(logrus.Fields{
			"user":    user.ID(),
			"service": "smtp",
		}),
		reporter: reporter,
		client:   client,
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

func (s *Service) Start(group *async.Group) {
	s.log.Debug("Starting service")
	group.Once(func(ctx context.Context) {
		logging.DoAnnotated(ctx, func(ctx context.Context) {
			s.run(ctx)
		}, logging.Labels{
			"user":    s.user.ID(),
			"service": "smtp",
		})
	})
}

func (s *Service) UserID() string {
	return s.user.ID()
}

func (s *Service) run(ctx context.Context) {
	s.log.Debug("Starting service main loop")
	defer s.log.Debug("Exiting service main loop")
	defer s.cpc.Close()

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

			default:
				s.log.Error("Received unknown request")
			}
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
