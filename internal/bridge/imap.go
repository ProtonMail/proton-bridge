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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/async"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/store/fallback_v0"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) restartIMAP(ctx context.Context) error {
	return bridge.serverManager.RestartIMAP(ctx)
}

// addIMAPUser connects the given user to gluon.
func (bridge *Bridge) addIMAPUser(ctx context.Context, user *user.User) error {
	return bridge.serverManager.AddIMAPUser(ctx, user)
}

// removeIMAPUser disconnects the given user from gluon, optionally also removing its files.
func (bridge *Bridge) removeIMAPUser(ctx context.Context, user *user.User, withData bool) error {
	return bridge.serverManager.RemoveIMAPUser(ctx, user, withData)
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

func ApplyGluonCachePathSuffix(basePath string) string {
	return filepath.Join(basePath, "backend", "store")
}

func ApplyGluonConfigPathSuffix(basePath string) string {
	return filepath.Join(basePath, "backend", "db")
}

func newIMAPServer(
	gluonCacheDir, gluonConfigDir string,
	version *semver.Version,
	tlsConfig *tls.Config,
	reporter reporter.Reporter,
	logClient, logServer bool,
	eventCh chan<- imapEvents.Event,
	tasks *async.Group,
	uidValidityGenerator imap.UIDValidityGenerator,
	panicHandler async.PanicHandler,
) (*gluon.Server, error) {
	gluonCacheDir = ApplyGluonCachePathSuffix(gluonCacheDir)
	gluonConfigDir = ApplyGluonConfigPathSuffix(gluonConfigDir)

	logrus.WithFields(logrus.Fields{
		"gluonStore": gluonCacheDir,
		"gluonDB":    gluonConfigDir,
		"version":    version,
		"logClient":  logClient,
		"logServer":  logServer,
	}).Info("Creating IMAP server")

	if logClient || logServer {
		log := logrus.WithField("protocol", "IMAP")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	var imapClientLog io.Writer

	if logClient {
		imapClientLog = logging.NewIMAPLogger()
	} else {
		imapClientLog = io.Discard
	}

	var imapServerLog io.Writer

	if logServer {
		imapServerLog = logging.NewIMAPLogger()
	} else {
		imapServerLog = io.Discard
	}

	imapServer, err := gluon.New(
		gluon.WithTLS(tlsConfig),
		gluon.WithDataDir(gluonCacheDir),
		gluon.WithDatabaseDir(gluonConfigDir),
		gluon.WithStoreBuilder(new(storeBuilder)),
		gluon.WithLogger(imapClientLog, imapServerLog),
		getGluonVersionInfo(version),
		gluon.WithReporter(reporter),
		gluon.WithUIDValidityGenerator(uidValidityGenerator),
		gluon.WithPanicHandler(panicHandler),
	)
	if err != nil {
		return nil, err
	}

	tasks.Once(func(ctx context.Context) {
		async.ForwardContext(ctx, eventCh, imapServer.AddWatcher())
	})

	tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, imapServer.GetErrorCh(), func(err error) {
			logrus.WithError(err).Error("IMAP server error")
		})
	})

	return imapServer, nil
}

func getGluonVersionInfo(version *semver.Version) gluon.Option {
	return gluon.WithVersionInfo(
		int(version.Major()),
		int(version.Minor()),
		int(version.Patch()),
		constants.FullAppName,
		"TODO",
		"TODO",
	)
}

type storeBuilder struct{}

func (*storeBuilder) New(path, userID string, passphrase []byte) (store.Store, error) {
	return store.NewOnDiskStore(
		filepath.Join(path, userID),
		passphrase,
		store.WithFallback(fallback_v0.NewOnDiskStoreV0WithCompressor(&fallback_v0.GZipCompressor{})),
	)
}

func (*storeBuilder) Delete(path, userID string) error {
	return os.RemoveAll(filepath.Join(path, userID))
}
