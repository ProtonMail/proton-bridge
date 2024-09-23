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

// Package bridge implements the Bridge, which acts as the backend to the UI.
package bridge

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	imapEvents "github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/ProtonMail/proton-bridge/v3/internal/identifier"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapsmtpserver"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/notifications"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/unleash"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/bradenaw/juniper/xslices"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var usernameChangeRegex = regexp.MustCompile(`^/Users/([^/]+)/`)

type Bridge struct {
	// vault holds bridge-specific data, such as preferences and known users (authorized or not).
	vault *vault.Vault

	// users holds authorized users.
	users     map[string]*user.User
	usersLock safe.RWMutex

	// api manages user API clients.
	api        *proton.Manager
	proxyCtl   ProxyController
	identifier identifier.Identifier

	// tlsConfig holds the bridge TLS config used by the IMAP and SMTP servers.
	tlsConfig *tls.Config

	// imapServer is the bridge's IMAP server.
	imapEventCh chan imapEvents.Event

	// updater is the bridge's updater.
	updater   Updater
	installCh chan installJob

	// heartbeat is the telemetry heartbeat for metrics.
	heartbeat *heartBeatState

	// curVersion is the current version of the bridge,
	// newVersion is the version that was installed by the updater.
	curVersion     *semver.Version
	newVersion     *semver.Version
	newVersionLock safe.RWMutex

	// keychains is the utils that own usable keychains found in the OS.
	keychains *keychain.List

	// focusService is used to raise the bridge window when needed.
	focusService *focus.Service

	// autostarter is the bridge's autostarter.
	autostarter Autostarter

	// locator is the bridge's locator.
	locator Locator

	// panicHandler
	panicHandler async.PanicHandler

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

	serverManager *imapsmtpserver.Service
	syncService   *syncservice.Service

	// unleashService is responsible for polling the feature flags and caching
	unleashService *unleash.Service

	// observabilityService is responsible for handling calls to the observability system
	observabilityService *observability.Service

	// notificationStore is used for notification deduplication
	notificationStore *notifications.Store
}

var logPkg = logrus.WithField("pkg", "bridge") //nolint:gochecknoglobals

// New creates a new bridge.
func New(
	locator Locator, // the locator to provide paths to store data
	vault *vault.Vault, // the bridge's encrypted data store
	autostarter Autostarter, // the autostarter to manage autostart settings
	updater Updater, // the updater to fetch and install updates
	curVersion *semver.Version, // the current version of the bridge
	keychains *keychain.List, // usable keychains

	apiURL string, // the URL of the API to use
	cookieJar http.CookieJar, // the cookie jar to use
	identifier identifier.Identifier, // the identifier to keep track of the user agent
	tlsReporter TLSReporter, // the TLS reporter to report TLS errors
	roundTripper http.RoundTripper, // the round tripper to use for API requests
	proxyCtl ProxyController, // the DoH controller
	panicHandler async.PanicHandler,
	reporter reporter.Reporter,
	uidValidityGenerator imap.UIDValidityGenerator,
	heartBeatManager telemetry.HeartbeatManager,

	logIMAPClient, logIMAPServer bool, // whether to log IMAP client/server activity
	logSMTP bool, // whether to log SMTP activity
) (*Bridge, <-chan events.Event, error) {
	// api is the user's API manager.
	api := proton.New(newAPIOptions(apiURL, curVersion, cookieJar, roundTripper, panicHandler)...)

	// tasks holds all the bridge's background tasks.
	tasks := async.NewGroup(context.Background(), panicHandler)

	// imapEventCh forwards IMAP events from gluon instances to the bridge for processing.
	imapEventCh := make(chan imapEvents.Event)

	// bridge is the bridge.
	bridge, err := newBridge(
		context.Background(),
		tasks,
		imapEventCh,

		locator,
		vault,
		autostarter,
		updater,
		curVersion,
		keychains,
		panicHandler,
		reporter,

		api,
		identifier,
		proxyCtl,
		uidValidityGenerator,
		heartBeatManager,
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

func newBridge(
	ctx context.Context,
	tasks *async.Group,
	imapEventCh chan imapEvents.Event,

	locator Locator,
	vault *vault.Vault,
	autostarter Autostarter,
	updater Updater,
	curVersion *semver.Version,
	keychains *keychain.List,
	panicHandler async.PanicHandler,
	reporter reporter.Reporter,

	api *proton.Manager,
	identifier identifier.Identifier,
	proxyCtl ProxyController,
	uidValidityGenerator imap.UIDValidityGenerator,
	heartbeatManager telemetry.HeartbeatManager,

	logIMAPClient, logIMAPServer, logSMTP bool,
) (*Bridge, error) {
	tlsConfig, err := loadTLSConfig(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	firstStart := vault.GetFirstStart()
	if err := vault.SetFirstStart(false); err != nil {
		return nil, fmt.Errorf("failed to save first start indicator: %w", err)
	}

	lastVersion := vault.GetLastVersion()
	if err := vault.SetLastVersion(curVersion); err != nil {
		return nil, fmt.Errorf("failed to save last version indicator: %w", err)
	}

	identifier.SetClientString(vault.GetLastUserAgent())

	focusService, err := focus.NewService(locator, curVersion, panicHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to create focus service: %w", err)
	}

	unleashService := unleash.NewBridgeService(ctx, api, locator, panicHandler)

	observabilityService := observability.NewService(ctx, panicHandler)

	bridge := &Bridge{
		vault: vault,

		users:     make(map[string]*user.User),
		usersLock: safe.NewRWMutex(),

		api:        api,
		proxyCtl:   proxyCtl,
		identifier: identifier,

		tlsConfig:   tlsConfig,
		imapEventCh: imapEventCh,

		updater:   updater,
		installCh: make(chan installJob),

		curVersion:     curVersion,
		newVersion:     curVersion,
		newVersionLock: safe.NewRWMutex(),

		keychains: keychains,

		panicHandler: panicHandler,
		reporter:     reporter,

		heartbeat: newHeartBeatState(ctx, panicHandler),

		focusService: focusService,
		autostarter:  autostarter,
		locator:      locator,

		logIMAPClient: logIMAPClient,
		logIMAPServer: logIMAPServer,
		logSMTP:       logSMTP,

		firstStart:  firstStart,
		lastVersion: lastVersion,

		tasks:       tasks,
		syncService: syncservice.NewService(panicHandler, observabilityService),

		unleashService: unleashService,

		observabilityService: observabilityService,

		notificationStore: notifications.NewStore(locator.ProvideNotificationsCachePath),
	}

	bridge.serverManager = imapsmtpserver.NewService(context.Background(),
		&bridgeSMTPSettings{b: bridge},
		&bridgeIMAPSettings{b: bridge},
		&bridgeEventPublisher{b: bridge},
		panicHandler,
		reporter,
		uidValidityGenerator,
		&bridgeIMAPSMTPTelemetry{b: bridge},
	)

	// Check whether username has changed and correct (macOS only)
	bridge.verifyUsernameChange()

	if err := bridge.serverManager.Init(context.Background(), bridge.tasks, &bridgeEventSubscription{b: bridge}); err != nil {
		return nil, err
	}

	if heartbeatManager == nil {
		bridge.heartbeat.init(bridge, bridge)
	} else {
		bridge.heartbeat.init(bridge, heartbeatManager)
	}

	bridge.syncService.Run()

	bridge.unleashService.Run()

	bridge.observabilityService.Run(bridge)

	return bridge, nil
}

func (bridge *Bridge) init(tlsReporter TLSReporter) error {
	// Enable or disable the proxy at startup.
	if bridge.vault.GetProxyAllowed() {
		bridge.proxyCtl.AllowProxy()
	} else {
		bridge.proxyCtl.DisallowProxy()
	}

	// Handle connection up/down events.
	bridge.api.AddStatusObserver(func(status proton.Status) {
		logPkg.Info("API status changed: ", status)

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
		logPkg.Warn("App version is bad")
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
			logrus.WithField("pkg", "gpa/manager").Infof("%v: %v %v", r.Status(), r.Request.Method, r.Request.URL)
		}

		return nil
	})

	// Publish a TLS issue event if a TLS issue is encountered.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, tlsReporter.GetTLSIssueCh(), func(struct{}) {
			logPkg.Warn("TLS issue encountered")
			bridge.publish(events.TLSIssue{})
		})
	})

	// Publish a raise event if the focus service is called.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, bridge.focusService.GetRaiseCh(), func(struct{}) {
			logPkg.Info("Focus service requested raise")
			bridge.publish(events.Raise{})
		})
	})

	// Handle any IMAP events that are forwarded to the bridge from gluon.
	bridge.tasks.Once(func(ctx context.Context) {
		async.RangeContext(ctx, bridge.imapEventCh, func(event imapEvents.Event) {
			logPkg.WithField("event", fmt.Sprintf("%T", event)).Debug("Received IMAP event")
			bridge.handleIMAPEvent(event)
		})
	})

	// Attempt to load users from the vault when triggered.
	bridge.goLoad = bridge.tasks.Trigger(func(ctx context.Context) {
		if err := bridge.loadUsers(ctx); err != nil {
			logPkg.WithError(err).Error("Failed to load users")
			if netErr := new(proton.NetError); !errors.As(err, &netErr) {
				sentry.ReportError(bridge.reporter, "Failed to load users", err)
			}
			return
		}

		bridge.publish(events.AllUsersLoaded{})
	})
	defer bridge.goLoad()

	// Check for updates when triggered.
	bridge.goUpdate = bridge.tasks.PeriodicOrTrigger(constants.UpdateCheckInterval, 0, func(ctx context.Context) {
		logPkg.Info("Checking for updates")

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
	logPkg.Info("Closing bridge")

	// Stop observability service
	bridge.observabilityService.Stop()

	// Stop heart beat before closing users.
	bridge.heartbeat.stop()

	// Close all users.
	safe.Lock(func() {
		for _, user := range bridge.users {
			user.Close()
		}
	}, bridge.usersLock)

	// Close the servers
	if err := bridge.serverManager.CloseServers(ctx); err != nil {
		logPkg.WithError(err).Error("Failed to close servers")
	}

	bridge.syncService.Close()

	// Stop all ongoing tasks.
	bridge.tasks.CancelAndWait()

	// Close the focus service.
	bridge.focusService.Close()

	// Close the unleash service.
	bridge.unleashService.Close()

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

	logPkg.WithField("event", event).Debug("Publishing event")

	for _, watcher := range bridge.watchers {
		if watcher.IsWatching(event) {
			if ok := watcher.Send(event); !ok {
				logPkg.WithField("event", event).Warn("Failed to send event to watcher")
			}
		}
	}
}

func (bridge *Bridge) addWatcher(ofType ...events.Event) *watcher.Watcher[events.Event] {
	bridge.watchersLock.Lock()
	defer bridge.watchersLock.Unlock()

	watcher := watcher.New(bridge.panicHandler, ofType...)

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

func (bridge *Bridge) onStatusUp(_ context.Context) {
	logPkg.Info("Handling API status up")

	bridge.goLoad()
}

func (bridge *Bridge) onStatusDown(ctx context.Context) {
	logPkg.Info("Handling API status down")

	for backoff := time.Second; ; backoff = min(backoff*2, 30*time.Second) {
		select {
		case <-ctx.Done():
			return

		case <-time.After(backoff):
			logPkg.Info("Pinging API")

			if err := bridge.api.Ping(ctx); err != nil {
				logPkg.WithError(err).Warn("Ping failed, API is still unreachable")
			} else {
				return
			}
		}
	}
}

func (bridge *Bridge) Repair() {
	var wg sync.WaitGroup
	userIDs := bridge.GetUserIDs()

	for _, userID := range userIDs {
		logPkg.Info("Initiating repair for userID:", userID)

		userInfo, err := bridge.GetUserInfo(userID)
		if err != nil {
			logPkg.WithError(err).Error("Failed getting user info for repair; ID:", userID)
			continue
		}

		if userInfo.State != Connected {
			logPkg.Info("User is not connected. Repair will be executed on following successful log in.", userID)
			if err := bridge.vault.GetUser(userID, func(user *vault.User) {
				if err := user.SetShouldSync(true); err != nil {
					logPkg.WithError(err).Error("Failed setting vault should sync for user:", userID)
				}
			}); err != nil {
				logPkg.WithError(err).Error("Unable to get user vault when scheduling repair:", userID)
			}
			continue
		}

		bridgeUser, ok := bridge.users[userID]
		if !ok {
			logPkg.Info("UserID does not exist in bridge user map", userID)
			continue
		}

		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			if err = bridgeUser.TriggerRepair(); err != nil {
				logPkg.WithError(err).Error("Failed re-syncing IMAP for userID", userID)
			}
		}(userID)
	}

	wg.Wait()
}

func loadTLSConfig(vault *vault.Vault) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(vault.GetBridgeTLSCert())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}

	return b
}

func (bridge *Bridge) HasAPIConnection() bool {
	return bridge.api.GetStatus() == proton.StatusUp
}

// verifyUsernameChange - works only on macOS
// it attempts to check whether a username change has taken place by comparing the gluon DB path (which is static and provided by bridge)
// to the gluon Cache path - which can be modified by the user and is stored in the vault;
// if a username discrepancy is detected, and the cache folder does not exist with the "old" username
// then we verify whether the gluon cache exists using the "new" username (provided by the DB path in this case)
// if so we modify the cache directory in the user vault.
func (bridge *Bridge) verifyUsernameChange() {
	if runtime.GOOS != "darwin" {
		return
	}

	gluonDBPath, err := bridge.GetGluonDataDir()
	if err != nil {
		logPkg.WithError(err).Error("Failed to get gluon db path")
		return
	}

	gluonCachePath := bridge.GetGluonCacheDir()
	// If the cache folder exists even on another user account or is in `/Users/Shared` we would still be able to access it
	// though it depends on the permissions; this is an edge-case.
	if _, err := os.Stat(gluonCachePath); err == nil {
		return
	}

	newCacheDir := GetUpdatedCachePath(gluonDBPath, gluonCachePath)
	if newCacheDir == "" {
		return
	}

	if _, err := os.Stat(newCacheDir); err == nil {
		logPkg.Info("Username change detected. Trying to restore gluon cache directory")
		if err = bridge.vault.SetGluonDir(newCacheDir); err != nil {
			logPkg.WithError(err).Error("Failed to restore gluon cache directory")
			return
		}
		logPkg.Info("Successfully restored gluon cache directory")
	}
}

func GetUpdatedCachePath(gluonDBPath, gluonCachePath string) string {
	// If gluon cache is moved to an external drive; regex find will fail; as is expected
	cachePathMatches := usernameChangeRegex.FindStringSubmatch(gluonCachePath)
	if cachePathMatches == nil || len(cachePathMatches) < 2 {
		return ""
	}

	cacheUsername := cachePathMatches[1]
	dbPathMatches := usernameChangeRegex.FindStringSubmatch(gluonDBPath)
	if dbPathMatches == nil || len(dbPathMatches) < 2 {
		return ""
	}

	dbUsername := dbPathMatches[1]
	if cacheUsername == dbUsername {
		return ""
	}

	return strings.Replace(gluonCachePath, "/Users/"+cacheUsername+"/", "/Users/"+dbUsername+"/", 1)
}

func (bridge *Bridge) GetFeatureFlagValue(key string) bool {
	return bridge.unleashService.GetFlagValue(key)
}

func (bridge *Bridge) PushObservabilityMetric(metric proton.ObservabilityMetric) {
	bridge.observabilityService.AddMetrics(metric)
}

func (bridge *Bridge) PushDistinctObservabilityMetrics(errType observability.DistinctionErrorTypeEnum, metrics ...proton.ObservabilityMetric) {
	bridge.observabilityService.AddDistinctMetrics(errType, metrics...)
}

func (bridge *Bridge) ModifyObservabilityHeartbeatInterval(duration time.Duration) {
	bridge.observabilityService.ModifyHeartbeatInterval(duration)
}
