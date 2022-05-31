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

package smtp

import (
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/confirmer"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	goSMTPBackend "github.com/emersion/go-smtp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type panicHandler interface {
	HandlePanic()
}

type settingsProvider interface {
	GetBool(string) bool
}

type smtpBackend struct {
	panicHandler  panicHandler
	eventListener listener.Listener
	settings      settingsProvider
	bridge        bridger
	confirmer     *confirmer.Confirmer
	sendRecorder  *sendRecorder
}

// NewSMTPBackend returns struct implementing go-smtp/backend interface.
func NewSMTPBackend(
	panicHandler panicHandler,
	eventListener listener.Listener,
	settings settingsProvider,
	bridge *bridge.Bridge,
) *smtpBackend { //nolint:revive
	return newSMTPBackend(panicHandler, eventListener, settings, newBridgeWrap(bridge))
}

func newSMTPBackend(
	panicHandler panicHandler,
	eventListener listener.Listener,
	settings settingsProvider,
	bridge bridger,
) *smtpBackend {
	return &smtpBackend{
		panicHandler:  panicHandler,
		eventListener: eventListener,
		settings:      settings,
		bridge:        bridge,
		confirmer:     confirmer.New(),
		sendRecorder:  newSendRecorder(),
	}
}

// Login authenticates a user.
func (sb *smtpBackend) Login(_ *goSMTPBackend.ConnectionState, username, password string) (goSMTPBackend.Session, error) {
	// Called from go-smtp in goroutines - we need to handle panics for each function.
	defer sb.panicHandler.HandlePanic()

	if sb.bridge.HasError(bridge.ErrLocalCacheUnavailable) {
		return nil, users.ErrLoggedOutUser
	}

	username = strings.ToLower(username)

	user, err := sb.bridge.GetUser(username)
	if err != nil {
		log.Warn("Cannot get user: ", err)
		return nil, err
	}
	if err := user.CheckBridgeLogin(password); err != nil {
		log.WithError(err).Error("Could not check bridge password")
		// Apple Mail sometimes generates a lot of requests very quickly. It's good practice
		// to have a timeout after bad logins so that we can slow those requests down a little bit.
		time.Sleep(10 * time.Second)
		return nil, err
	}
	// Client can log in only using address so we can properly close all SMTP connections.
	addressID, err := user.GetAddressID(username)
	if err != nil {
		log.Error("Cannot get addressID: ", err)
		return nil, err
	}
	// AddressID is only for split mode--it has to be empty for combined mode.
	if user.IsCombinedAddressMode() {
		addressID = ""
	}
	return newSMTPUser(sb.panicHandler, sb.eventListener, sb, user, username, addressID)
}

func (sb *smtpBackend) AnonymousLogin(_ *goSMTPBackend.ConnectionState) (goSMTPBackend.Session, error) {
	// Called from go-smtp in goroutines - we need to handle panics for each function.
	defer sb.panicHandler.HandlePanic()

	return nil, errors.New("anonymous login not supported")
}

func (sb *smtpBackend) shouldReportOutgoingNoEnc() bool {
	return sb.settings.GetBool(settings.ReportOutgoingNoEncKey)
}

func (sb *smtpBackend) ConfirmNoEncryption(messageID string, shouldSend bool) {
	if err := sb.confirmer.SetResult(messageID, shouldSend); err != nil {
		logrus.WithError(err).Error("Failed to set confirmation value")
	}
}
