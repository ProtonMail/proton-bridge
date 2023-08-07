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
	"crypto/tls"
	"strings"

	"github.com/Masterminds/semver/v3"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapsmtpserver"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) restartIMAP(ctx context.Context) error {
	return bridge.serverManager.RestartIMAP(ctx)
}

func (bridge *Bridge) handleIMAPEvent(event imapEvents.Event) {
	switch event := event.(type) {
	case imapEvents.UserAdded:
		for labelID, count := range event.Counts {
			logrus.WithFields(logrus.Fields{
				"gluonID": event.UserID,
				"labelID": labelID,
				"count":   count,
			}).Info("Received mailbox message count")
		}

	case imapEvents.IMAPID:
		logrus.WithFields(logrus.Fields{
			"sessionID": event.SessionID,
			"name":      event.IMAPID.Name,
			"version":   event.IMAPID.Version,
		}).Info("Received IMAP ID")

		if event.IMAPID.Name != "" && event.IMAPID.Version != "" {
			bridge.setUserAgent(event.IMAPID.Name, event.IMAPID.Version)
		}

	case imapEvents.LoginFailed:
		logrus.WithFields(logrus.Fields{
			"sessionID": event.SessionID,
			"username":  event.Username,
			"pkg":       "imap",
		}).Error("Incorrect login credentials.")
		bridge.publish(events.IMAPLoginFailed{Username: event.Username})

	case imapEvents.Login:
		if strings.Contains(bridge.GetCurrentUserAgent(), useragent.DefaultUserAgent) {
			bridge.setUserAgent(useragent.UnknownClient, useragent.DefaultVersion)
		}
	}
}

type bridgeIMAPSettings struct {
	b *Bridge
}

func (b *bridgeIMAPSettings) EventPublisher() imapsmtpserver.IMAPEventPublisher {
	return b
}

func (b *bridgeIMAPSettings) TLSConfig() *tls.Config {
	return b.b.tlsConfig
}

func (b *bridgeIMAPSettings) LogClient() bool {
	return b.b.logIMAPClient
}

func (b *bridgeIMAPSettings) LogServer() bool {
	return b.b.logIMAPServer
}

func (b *bridgeIMAPSettings) Port() int {
	return b.b.vault.GetIMAPPort()
}

func (b *bridgeIMAPSettings) SetPort(i int) error {
	return b.b.vault.SetIMAPPort(i)
}

func (b *bridgeIMAPSettings) UseSSL() bool {
	return b.b.vault.GetIMAPSSL()
}

func (b *bridgeIMAPSettings) CacheDirectory() string {
	return b.b.GetGluonCacheDir()
}

func (b *bridgeIMAPSettings) DataDirectory() (string, error) {
	return b.b.GetGluonDataDir()
}

func (b *bridgeIMAPSettings) SetCacheDirectory(s string) error {
	return b.b.vault.SetGluonDir(s)
}

func (b *bridgeIMAPSettings) Version() *semver.Version {
	return b.b.curVersion
}

func (b *bridgeIMAPSettings) PublishIMAPEvent(ctx context.Context, event imapEvents.Event) {
	select {
	case <-ctx.Done():
		return
	case b.b.imapEventCh <- event:
		// do nothing
	}
}
