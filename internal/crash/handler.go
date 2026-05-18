// Copyright (c) 2026 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package crash implements a crash handler with configurable recovery actions.
package crash

import (
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/sirupsen/logrus"
)

type RecoveryAction func(interface{}) error

type Handler struct {
	actions []RecoveryAction
}

func NewHandler(actions ...RecoveryAction) *Handler {
	return &Handler{actions: actions}
}

func (h *Handler) AddRecoveryAction(action RecoveryAction) *Handler {
	h.actions = append(h.actions, action)
	return h
}

func (h *Handler) HandlePanic(r interface{}) {
	sentry.SkipDuringUnwind()

	if r == nil {
		return
	}

	for _, action := range h.actions {
		runAction(action, r)
	}
}

// runAction invokes a single recovery action with its own recover guard so
// that a panic in one action (for example, a nil dereference inside a
// notification helper) cannot prevent the remaining actions from running.
func runAction(action RecoveryAction, r interface{}) {
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Error("Recovery action panicked")
		}
	}()

	if action == nil {
		return
	}

	if err := action(r); err != nil {
		logrus.WithError(err).Error("Failed to execute recovery action")
	}
}
