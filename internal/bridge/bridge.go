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
	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

type Bridge struct {
	// vault holds bridge-specific data, such as preferences and known users (authorized or not).
	vault *vault.Vault

	// users holds authorized users.
	users map[string]*user.User

	// api manages user API clients.
	api        *liteapi.Manager
	proxyCtl   ProxyController
	identifier Identifier

	// watchers holds all registered event watchers.
	watchers     []*watcher.Watcher[events.Event]
	watchersLock sync.RWMutex

	// tlsConfig holds the bridge TLS config used by the IMAP and SMTP servers.
	tlsConfig *tls.Config

	// imapServer is the bridge's IMAP server.
	imapServer   *gluon.Server
	imapListener net.Listener

	// smtpServer is the bridge's SMTP server.
	smtpServer  *smtp.Server
	smtpBackend *smtpBackend

	// updater is the bridge's updater.
	updater       Updater
	curVersion    *semver.Version
	updateCheckCh chan struct{}

	// focusService is used to raise the bridge window when needed.
	focusService *focus.Service

	// autostarter is the bridge's autostarter.
	autostarter Autostarter

	// locator is the bridge's locator.
	locator Locator

	// errors contains errors encountered during startup.
	errors []error

	// These control the bridge's IMAP and SMTP logging behaviour.
	logIMAPClient bool
	logIMAPServer bool
	logSMTP       bool

	// stopCh is used to stop ongoing goroutines when the bridge is closed.
	stopCh chan struct{}
}

// New creates a new bridge.
func New(
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

	logIMAPClient, logIMAPServer bool, // whether to log IMAP client/server activity
	logSMTP bool, // whether to log SMTP activity
) (*Bridge, error) {
	api := liteapi.New(
		liteapi.WithHostURL(apiURL),
		liteapi.WithAppVersion(constants.AppVersion),
		liteapi.WithCookieJar(cookieJar),
		liteapi.WithTransport(roundTripper),
	)

	tlsConfig, err := loadTLSConfig(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	gluonDir, err := getGluonDir(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon directory: %w", err)
	}

	smtpBackend, err := newSMTPBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP backend: %w", err)
	}

	imapServer, err := newIMAPServer(gluonDir, curVersion, tlsConfig, logIMAPClient, logIMAPServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP server: %w", err)
	}

	focusService, err := focus.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create focus service: %w", err)
	}

	bridge := newBridge(
		locator,
		vault,
		autostarter,
		updater,
		curVersion,

		api,
		identifier,
		proxyCtl,

		tlsConfig,
		imapServer,
		smtpBackend,
		focusService,
		logIMAPClient,
		logIMAPServer,
		logSMTP,
	)

	if err := bridge.init(tlsReporter); err != nil {
		return nil, fmt.Errorf("failed to initialize bridge: %w", err)
	}

	return bridge, nil
}

func newBridge(
	locator Locator,
	vault *vault.Vault,
	autostarter Autostarter,
	updater Updater,
	curVersion *semver.Version,

	api *liteapi.Manager,
	identifier Identifier,
	proxyCtl ProxyController,

	tlsConfig *tls.Config,
	imapServer *gluon.Server,
	smtpBackend *smtpBackend,
	focusService *focus.Service,
	logIMAPClient, logIMAPServer, logSMTP bool,
) *Bridge {
	return &Bridge{
		vault: vault,
		users: make(map[string]*user.User),

		api:        api,
		proxyCtl:   proxyCtl,
		identifier: identifier,

		tlsConfig:   tlsConfig,
		imapServer:  imapServer,
		smtpServer:  newSMTPServer(smtpBackend, tlsConfig, logSMTP),
		smtpBackend: smtpBackend,

		updater:       updater,
		curVersion:    curVersion,
		updateCheckCh: make(chan struct{}, 1),

		focusService: focusService,
		autostarter:  autostarter,
		locator:      locator,

		logIMAPClient: logIMAPClient,
		logIMAPServer: logIMAPServer,
		logSMTP:       logSMTP,

		stopCh: make(chan struct{}),
	}
}

func (bridge *Bridge) init(tlsReporter TLSReporter) error {
	if bridge.vault.GetProxyAllowed() {
		bridge.proxyCtl.AllowProxy()
	} else {
		bridge.proxyCtl.DisallowProxy()
	}

	bridge.api.AddStatusObserver(func(status liteapi.Status) {
		switch {
		case status == liteapi.StatusUp:
			go bridge.onStatusUp()

		case status == liteapi.StatusDown:
			go bridge.onStatusDown()
		}
	})

	bridge.api.AddErrorHandler(liteapi.AppVersionBadCode, func() {
		bridge.publish(events.UpdateForced{})
	})

	bridge.api.AddPreRequestHook(func(_ *resty.Client, req *resty.Request) error {
		req.SetHeader("User-Agent", bridge.identifier.GetUserAgent())
		return nil
	})

	if err := bridge.loadUsers(); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	go func() {
		for range tlsReporter.GetTLSIssueCh() {
			bridge.publish(events.TLSIssue{})
		}
	}()

	go func() {
		for range bridge.focusService.GetRaiseCh() {
			bridge.publish(events.Raise{})
		}
	}()

	go func() {
		for event := range bridge.imapServer.AddWatcher() {
			bridge.handleIMAPEvent(event)
		}
	}()

	if err := bridge.serveIMAP(); err != nil {
		bridge.PushError(ErrServeIMAP)
	}

	if err := bridge.serveSMTP(); err != nil {
		bridge.PushError(ErrServeSMTP)
	}

	if err := bridge.watchForUpdates(); err != nil {
		bridge.PushError(ErrWatchUpdates)
	}

	return nil
}

// GetEvents returns a channel of events of the given type.
// If no types are supplied, all events are returned.
func (bridge *Bridge) GetEvents(ofType ...events.Event) (<-chan events.Event, func()) {
	newWatcher := bridge.addWatcher(ofType...)

	return newWatcher.GetChannel(), func() { bridge.remWatcher(newWatcher) }
}

func (bridge *Bridge) FactoryReset(ctx context.Context) error {
	panic("TODO")
}

func (bridge *Bridge) PushError(err error) {
	bridge.errors = append(bridge.errors, err)
}

func (bridge *Bridge) GetErrors() []error {
	return bridge.errors
}

func (bridge *Bridge) Close(ctx context.Context) error {
	// Stop ongoing operations such as connectivity checks.
	close(bridge.stopCh)

	// Close the IMAP server.
	if err := bridge.closeIMAP(ctx); err != nil {
		logrus.WithError(err).Error("Failed to close IMAP server")
	}

	// Close the SMTP server.
	if err := bridge.closeSMTP(); err != nil {
		logrus.WithError(err).Error("Failed to close SMTP server")
	}

	// Close all users.
	for _, user := range bridge.users {
		if err := user.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close user")
		}
	}

	// Close the focus service.
	bridge.focusService.Close()

	// Save the last version of bridge that was run.
	if err := bridge.vault.SetLastVersion(bridge.curVersion); err != nil {
		logrus.WithError(err).Error("Failed to save last version")
	}

	return nil
}

func (bridge *Bridge) publish(event events.Event) {
	bridge.watchersLock.RLock()
	defer bridge.watchersLock.RUnlock()

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

	newWatcher := watcher.New(ofType...)

	bridge.watchers = append(bridge.watchers, newWatcher)

	return newWatcher
}

func (bridge *Bridge) remWatcher(oldWatcher *watcher.Watcher[events.Event]) {
	bridge.watchersLock.Lock()
	defer bridge.watchersLock.Unlock()

	bridge.watchers = xslices.Filter(bridge.watchers, func(other *watcher.Watcher[events.Event]) bool {
		return other != oldWatcher
	})
}

func (bridge *Bridge) onStatusUp() {
	bridge.publish(events.ConnStatusUp{})

	if err := bridge.loadUsers(); err != nil {
		logrus.WithError(err).Error("Failed to load users")
	}
}

func (bridge *Bridge) onStatusDown() {
	bridge.publish(events.ConnStatusDown{})

	upCh, done := bridge.GetEvents(events.ConnStatusUp{})
	defer done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backoff := time.Second

	for {
		select {
		case <-upCh:
			return

		case <-bridge.stopCh:
			return

		case <-time.After(backoff):
			if err := bridge.api.Ping(ctx); err != nil {
				logrus.WithError(err).Debug("Failed to ping API")
			}
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func loadTLSConfig(vault *vault.Vault) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(vault.GetBridgeTLSCert(), vault.GetBridgeTLSKey())
	if err != nil {
		return nil, err
	}

	// TODO: Do we have to set MinVersion to tls.VersionTLS12?
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func newListener(port int, useTLS bool, tlsConfig *tls.Config) (net.Listener, error) {
	if useTLS {
		tlsListener, err := tls.Listen("tcp", fmt.Sprintf(":%v", port), tlsConfig)
		if err != nil {
			return nil, err
		}

		return tlsListener, nil
	}

	netListener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	return netListener, nil
}
