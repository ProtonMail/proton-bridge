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

package syncservice

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/sirupsen/logrus"
)

type ApplyRequest struct {
	childJob
	messages []BuildResult
}

type ApplyStageInput = StageInputConsumer[ApplyRequest]

// ApplyStage applies the sync updates and waits for their completion before proceeding with the next batch. This is
// the final stage in the sync pipeline.
type ApplyStage struct {
	input ApplyStageInput
	log   *logrus.Entry
}

func NewApplyStage(input ApplyStageInput) *ApplyStage {
	return &ApplyStage{input: input, log: logrus.WithField("sync-stage", "apply")}
}

func (a *ApplyStage) Run(group *async.Group) {
	group.Once(func(ctx context.Context) {
		logging.DoAnnotated(
			ctx,
			func(ctx context.Context) {
				a.run(ctx)
			},
			logging.Labels{"sync-stage": "apply"},
		)
	})
}

func (a *ApplyStage) run(ctx context.Context) {
	for {
		req, err := a.input.Consume(ctx)
		if err != nil {
			if !(errors.Is(err, ErrNoMoreInput) || errors.Is(err, context.Canceled)) {
				a.log.WithError(err).Error("Exiting state with error")
			}

			return
		}

		if req.checkCancelled() {
			continue
		}

		if len(req.messages) == 0 {
			req.onFinished(req.getContext())
			continue
		}

		if err := req.job.updateApplier.ApplySyncUpdates(req.getContext(), req.messages); err != nil {
			a.log.WithError(err).Error("Failed to apply sync updates")
			req.job.onError(err)
			continue
		}

		req.onFinished(req.getContext())
	}
}
