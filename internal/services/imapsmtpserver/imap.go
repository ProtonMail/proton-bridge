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

package imapsmtpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/async"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/store/fallback_v0"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/files"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/sirupsen/logrus"
)

var logIMAP = logrus.WithField("pkg", "server/imap") //nolint:gochecknoglobals

type IMAPSettingsProvider interface {
	TLSConfig() *tls.Config
	LogClient() bool
	LogServer() bool
	Port() int
	SetPort(int) error
	UseSSL() bool
	CacheDirectory() string
	DataDirectory() (string, error)
	SetCacheDirectory(string) error
	EventPublisher() IMAPEventPublisher
	Version() *semver.Version
}

type IMAPEventPublisher interface {
	PublishIMAPEvent(ctx context.Context, event imapEvents.Event)
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
	eventPublisher IMAPEventPublisher,
	tasks *async.Group,
	uidValidityGenerator imap.UIDValidityGenerator,
	panicHandler async.PanicHandler,
	observabilitySender observability.Sender,
) (*gluon.Server, error) {
	gluonCacheDir = ApplyGluonCachePathSuffix(gluonCacheDir)
	gluonConfigDir = ApplyGluonConfigPathSuffix(gluonConfigDir)

	logIMAP.WithFields(logrus.Fields{
		"gluonStore": gluonCacheDir,
		"gluonDB":    gluonConfigDir,
		"version":    version,
		"logClient":  logClient,
		"logServer":  logServer,
	}).Info("Creating IMAP server")

	if logClient || logServer {
		logIMAP.Warning("================================================")
		logIMAP.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		logIMAP.Warning("================================================")
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
		gluon.WithObservabilitySender(observability.NewAdapter(observabilitySender), int(observability.GluonImapError), int(observability.GluonMessageError), int(observability.GluonOtherError)),
	)
	if err != nil {
		return nil, err
	}

	tasks.Once(func(ctx context.Context) {
		watcher := imapServer.AddWatcher()
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-watcher:
				if !ok {
					return
				}

				eventPublisher.PublishIMAPEvent(ctx, e)
			}
		}
	})

	tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, imapServer.GetErrorCh(), func(err error) {
			logIMAP.WithError(err).Error("IMAP server error")
		})
	})

	return imapServer, nil
}

func getGluonVersionInfo(version *semver.Version) gluon.Option {
	return gluon.WithVersionInfo(
		int(version.Major()), //nolint:gosec // disable G115
		int(version.Minor()), //nolint:gosec // disable G115
		int(version.Patch()), //nolint:gosec // disable G115
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

func moveGluonCacheDir(settings IMAPSettingsProvider, oldGluonDir, newGluonDir string) error {
	logIMAP.WithField("pkg", "service/imap").Infof("gluon cache moving from %s to %s", oldGluonDir, newGluonDir)
	oldCacheDir := ApplyGluonCachePathSuffix(oldGluonDir)
	if err := files.CopyDir(oldCacheDir, ApplyGluonCachePathSuffix(newGluonDir)); err != nil {
		return fmt.Errorf("failed to copy gluon dir: %w", err)
	}

	if err := settings.SetCacheDirectory(newGluonDir); err != nil {
		return fmt.Errorf("failed to set new gluon cache dir: %w", err)
	}

	if err := os.RemoveAll(oldCacheDir); err != nil {
		logIMAP.WithError(err).Error("failed to remove old gluon cache dir")
	}

	return nil
}
