// Copyright (c) 2022 Proton AG
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
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
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

func (h *Handler) HandlePanic() {
	sentry.SkipDuringUnwind()

	r := recover()
	if r == nil {
		return
	}

	for _, action := range h.actions {
		if err := action(r); err != nil {
			logrus.WithError(err).Error("Failed to execute recovery action")
		}
	}
}
