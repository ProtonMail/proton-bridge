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

// Package bridge implements the Bridge, which acts as the backend to the UI.
package bridge

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Bridge struct {
	// vault holds bridge-specific data, such as preferences and known users (authorized or not).
	vault *vault.Vault

	// users holds authorized users.
	users     map[string]*user.User
	usersLock safe.RWMutex

	// api manages user API clients.
	api        *proton.Manager
	proxyCtl   ProxyController
	identifier Identifier

	// tlsConfig holds the bridge TLS config used by the IMAP and SMTP servers.
	tlsConfig *tls.Config

	// imapServer is the bridge's IMAP server.
	imapServer   *gluon.Server
	imapListener net.Listener
	imapEventCh  chan imapEvents.Event

	// smtpServer is the bridge's SMTP server.
	smtpServer   *smtp.Server
	smtpListener net.Listener

	// updater is the bridge's updater.
	updater   Updater
	installCh chan installJob

	// curVersion is the current version of the bridge,
	// newVersion is the version that was installed by the updater.
	curVersion     *semver.Version
	newVersion     *semver.Version
	newVersionLock safe.RWMutex

	// focusService is used to raise the bridge window when needed.
	focusService *focus.Service

	// autostarter is the bridge's autostarter.
	autostarter Autostarter

	// locator is the bridge's locator.
	locator Locator

	// crashHandler
	crashHandler async.PanicHandler

	// reporter
	reporter reporter.Reporter

	// watchers holds all registered event watchers.
	watchers     []*watcher.Watcher[events.Event]
	watchersLock sync.RWMutex

	// errors contains errors encountered during startup.
	errors []error

	// These control the bridge's IMAP and SMTP logging behaviour.
	logIMAPClient bool
	logIMAPServer bool
	logSMTP       bool

	// These two variables keep track of the startup values for the two settings of the same name.
	// They are updated in the vault on startup so that we're sure they're updated in case of kill/crash,
	// but we need to keep their initial value for the current instance of bridge.
	firstStart  bool
	lastVersion *semver.Version

	// tasks manages the bridge's goroutines.
	tasks *async.Group

	// goLoad triggers a load of disconnected users from the vault.
	goLoad func()

	// goUpdate triggers a check/install of updates.
	goUpdate func()

	uidValidityGenerator imap.UIDValidityGenerator
}

// New creates a new bridge.
func New( //nolint:funlen
	locator Locator, // the locator to provide paths to store data
	vault *vault.Vault, // the bridge's encrypted data store
	autostarter Autostarter, // the autostarter to manage autostart settings
	updater Updater, // the updater to fetch and install updates
	curVersion *semver.Version, // the current version of the bridge

	apiURL string, // the URL of the API to use
	cookieJar http.CookieJar, // the cookie jar to use
	identifier Identifier, // the identifier to keep track of the user agent
	tlsReporter TLSReporter, // the TLS reporter to report TLS errors
	roundTripper http.RoundTripper, // the round tripper to use for API requests
	proxyCtl ProxyController, // the DoH controller
	crashHandler async.PanicHandler,
	reporter reporter.Reporter,
	uidValidityGenerator imap.UIDValidityGenerator,

	logIMAPClient, logIMAPServer bool, // whether to log IMAP client/server activity
	logSMTP bool, // whether to log SMTP activity
) (*Bridge, <-chan events.Event, error) {
	// api is the user's API manager.
	api := proton.New(newAPIOptions(apiURL, curVersion, cookieJar, roundTripper)...)

	// tasks holds all the bridge's background tasks.
	tasks := async.NewGroup(context.Background(), crashHandler)

	// imapEventCh forwards IMAP events from gluon instances to the bridge for processing.
	imapEventCh := make(chan imapEvents.Event)

	// bridge is the bridge.
	bridge, err := newBridge(
		tasks,
		imapEventCh,

		locator,
		vault,
		autostarter,
		updater,
		curVersion,
		crashHandler,
		reporter,

		api,
		identifier,
		proxyCtl,
		uidValidityGenerator,
		logIMAPClient, logIMAPServer, logSMTP,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bridge: %w", err)
	}

	// Get an event channel for all events (individual events can be subscribed to later).
	eventCh, _ := bridge.GetEvents()

	// Initialize all of bridge's background tasks and operations.
	if err := bridge.init(tlsReporter); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize bridge: %w", err)
	}

	return bridge, eventCh, nil
}

// nolint:funlen
func newBridge(
	tasks *async.Group,
	imapEventCh chan imapEvents.Event,

	locator Locator,
	vault *vault.Vault,
	autostarter Autostarter,
	updater Updater,
	curVersion *semver.Version,
	crashHandler async.PanicHandler,
	reporter reporter.Reporter,

	api *proton.Manager,
	identifier Identifier,
	proxyCtl ProxyController,
	uidValidityGenerator imap.UIDValidityGenerator,

	logIMAPClient, logIMAPServer, logSMTP bool,
) (*Bridge, error) {
	tlsConfig, err := loadTLSConfig(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	gluonCacheDir, err := getGluonDir(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon directory: %w", err)
	}

	gluonDataDir, err := locator.ProvideGluonDataPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon Database directory: %w", err)
	}

	firstStart := vault.GetFirstStart()
	if err := vault.SetFirstStart(false); err != nil {
		return nil, fmt.Errorf("failed to save first start indicator: %w", err)
	}

	lastVersion := vault.GetLastVersion()
	if err := vault.SetLastVersion(curVersion); err != nil {
		return nil, fmt.Errorf("failed to save last version indicator: %w", err)
	}

	imapServer, err := newIMAPServer(
		gluonCacheDir,
		gluonDataDir,
		curVersion,
		tlsConfig,
		reporter,
		logIMAPClient,
		logIMAPServer,
		imapEventCh,
		tasks,
		uidValidityGenerator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP server: %w", err)
	}

	focusService, err := focus.NewService(curVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create focus service: %w", err)
	}

	bridge := &Bridge{
		vault: vault,

		users:     make(map[string]*user.User),
		usersLock: safe.NewRWMutex(),

		api:        api,
		proxyCtl:   proxyCtl,
		identifier: identifier,

		tlsConfig:   tlsConfig,
		imapServer:  imapServer,
		imapEventCh: imapEventCh,

		updater:   updater,
		installCh: make(chan installJob),

		curVersion:     curVersion,
		newVersion:     curVersion,
		newVersionLock: safe.NewRWMutex(),

		crashHandler: crashHandler,
		reporter:     reporter,

		focusService: focusService,
		autostarter:  autostarter,
		locator:      locator,

		logIMAPClient: logIMAPClient,
		logIMAPServer: logIMAPServer,
		logSMTP:       logSMTP,

		firstStart:  firstStart,
		lastVersion: lastVersion,

		tasks: tasks,

		uidValidityGenerator: uidValidityGenerator,
	}

	bridge.smtpServer = newSMTPServer(bridge, tlsConfig, logSMTP)

	return bridge, nil
}

// nolint:funlen
func (bridge *Bridge) init(tlsReporter TLSReporter) error {
	// Enable or disable the proxy at startup.
	if bridge.vault.GetProxyAllowed() {
		bridge.proxyCtl.AllowProxy()
	} else {
		bridge.proxyCtl.DisallowProxy()
	}

	// Handle connection up/down events.
	bridge.api.AddStatusObserver(func(status proton.Status) {
		logrus.Info("API status changed: ", status)

		switch {
		case status == proton.StatusUp:
			bridge.publish(events.ConnStatusUp{})
			bridge.tasks.Once(bridge.onStatusUp)

		case status == proton.StatusDown:
			bridge.publish(events.ConnStatusDown{})
			bridge.tasks.Once(bridge.onStatusDown)
		}
	})

	// If any call returns a bad version code, we need to update.
	bridge.api.AddErrorHandler(proton.AppVersionBadCode, func() {
		logrus.Warn("App version is bad")
		bridge.publish(events.UpdateForced{})
	})

	// Ensure all outgoing headers have the correct user agent.
	bridge.api.AddPreRequestHook(func(_ *resty.Client, req *resty.Request) error {
		req.SetHeader("User-Agent", bridge.identifier.GetUserAgent())
		return nil
	})

	// Log all manager API requests (client requests are logged separately).
	bridge.api.AddPostRequestHook(func(_ *resty.Client, r *resty.Response) error {
		if _, ok := proton.ClientIDFromContext(r.Request.Context()); !ok {
			logrus.Infof("[MANAGER] %v: %v %v", r.Status(), r.Request.Method, r.Request.URL)
		}

		return nil
	})

	// Publish a TLS issue event if a TLS issue is encountered.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, tlsReporter.GetTLSIssueCh(), func(struct{}) {
			logrus.Warn("TLS issue encountered")
			bridge.publish(events.TLSIssue{})
		})
	})

	// Publish a raise event if the focus service is called.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, bridge.focusService.GetRaiseCh(), func(struct{}) {
			logrus.Info("Focus service requested raise")
			bridge.publish(events.Raise{})
		})
	})

	// Handle any IMAP events that are forwarded to the bridge from gluon.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, bridge.imapEventCh, func(event imapEvents.Event) {
			logrus.WithField("event", fmt.Sprintf("%T", event)).Debug("Received IMAP event")
			bridge.handleIMAPEvent(event)
		})
	})

	// Attempt to lazy load users when triggered. Only load once.
	var loaded bool
	bridge.goLoad = bridge.tasks.Trigger(func(ctx context.Context) {
		if loaded {
			logrus.Debug("All users are already loaded, skipping")
			return
		}

		logrus.Info("Loading users")

		if err := bridge.loadUsers(ctx); err != nil {
			logrus.WithError(err).Error("Failed to load users")
			return
		}

		bridge.publish(events.AllUsersLoaded{})

		// Once all users have been loaded, start the bridge's IMAP and SMTP servers.
		if err := bridge.serveIMAP(); err != nil {
			logrus.WithError(err).Error("Failed to start IMAP server")
		}

		if err := bridge.serveSMTP(); err != nil {
			logrus.WithError(err).Error("Failed to start SMTP server")
		}

		loaded = true
	})
	defer bridge.goLoad()

	// Check for updates when triggered.
	bridge.goUpdate = bridge.tasks.PeriodicOrTrigger(constants.UpdateCheckInterval, 0, func(ctx context.Context) {
		logrus.Info("Checking for updates")

		version, err := bridge.updater.GetVersionInfo(ctx, bridge.api, bridge.vault.GetUpdateChannel())
		if err != nil {
			bridge.publish(events.UpdateCheckFailed{Error: err})
		} else {
			bridge.handleUpdate(version)
		}
	})
	defer bridge.goUpdate()

	// Install updates when available.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, bridge.installCh, func(job installJob) {
			bridge.installUpdate(ctx, job)
		})
	})

	return nil
}

// GetEvents returns a channel of events of the given type.
// If no types are supplied, all events are returned.
func (bridge *Bridge) GetEvents(ofType ...events.Event) (<-chan events.Event, context.CancelFunc) {
	watcher := bridge.addWatcher(ofType...)

	return watcher.GetChannel(), func() { bridge.remWatcher(watcher) }
}

func (bridge *Bridge) PushError(err error) {
	bridge.errors = append(bridge.errors, err)
}

func (bridge *Bridge) GetErrors() []error {
	return bridge.errors
}

func (bridge *Bridge) Close(ctx context.Context) {
	logrus.Info("Closing bridge")

	// Close the IMAP server.
	if err := bridge.closeIMAP(ctx); err != nil {
		logrus.WithError(err).Error("Failed to close IMAP server")
	}

	// Close the SMTP server.
	if err := bridge.closeSMTP(); err != nil {
		logrus.WithError(err).Error("Failed to close SMTP server")
	}

	// Close all users.
	safe.RLock(func() {
		for _, user := range bridge.users {
			user.Close()
		}
	}, bridge.usersLock)

	// Stop all ongoing tasks.
	bridge.tasks.CancelAndWait()

	// Close the focus service.
	bridge.focusService.Close()

	// Close the watchers.
	bridge.watchersLock.Lock()
	defer bridge.watchersLock.Unlock()

	for _, watcher := range bridge.watchers {
		watcher.Close()
	}

	bridge.watchers = nil
}

func (bridge *Bridge) publish(event events.Event) {
	bridge.watchersLock.RLock()
	defer bridge.watchersLock.RUnlock()

	logrus.WithField("event", event).Debug("Publishing event")

	for _, watcher := range bridge.watchers {
		if watcher.IsWatching(event) {
			if ok := watcher.Send(event); !ok {
				logrus.WithField("event", event).Warn("Failed to send event to watcher")
			}
		}
	}
}

func (bridge *Bridge) addWatcher(ofType ...events.Event) *watcher.Watcher[events.Event] {
	bridge.watchersLock.Lock()
	defer bridge.watchersLock.Unlock()

	watcher := watcher.New(ofType...)

	bridge.watchers = append(bridge.watchers, watcher)

	return watcher
}

func (bridge *Bridge) remWatcher(watcher *watcher.Watcher[events.Event]) {
	bridge.watchersLock.Lock()
	defer bridge.watchersLock.Unlock()

	idx := xslices.Index(bridge.watchers, watcher)

	if idx < 0 {
		return
	}

	bridge.watchers = append(bridge.watchers[:idx], bridge.watchers[idx+1:]...)

	watcher.Close()
}

func (bridge *Bridge) onStatusUp(ctx context.Context) {
	logrus.Info("Handling API status up")

	safe.RLock(func() {
		for _, user := range bridge.users {
			user.OnStatusUp(ctx)
		}
	}, bridge.usersLock)

	bridge.goLoad()
}

func (bridge *Bridge) onStatusDown(ctx context.Context) {
	logrus.Info("Handling API status down")

	safe.RLock(func() {
		for _, user := range bridge.users {
			user.OnStatusDown(ctx)
		}
	}, bridge.usersLock)

	for backoff := time.Second; ; backoff = min(backoff*2, 30*time.Second) {
		select {
		case <-ctx.Done():
			return

		case <-time.After(backoff):
			logrus.Info("Pinging API")

			if err := bridge.api.Ping(ctx); err != nil {
				logrus.WithError(err).Warn("Ping failed, API is still unreachable")
			} else {
				return
			}
		}
	}
}

func loadTLSConfig(vault *vault.Vault) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(vault.GetBridgeTLSCert(), vault.GetBridgeTLSKey())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func newListener(port int, useTLS bool, tlsConfig *tls.Config) (net.Listener, error) {
	if useTLS {
		tlsListener, err := tls.Listen("tcp", fmt.Sprintf("%v:%v", constants.Host, port), tlsConfig)
		if err != nil {
			return nil, err
		}

		return tlsListener, nil
	}

	netListener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", constants.Host, port))
	if err != nil {
		return nil, err
	}

	return netListener, nil
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}

	return b
}
