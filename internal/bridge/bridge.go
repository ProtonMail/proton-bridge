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
	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
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
	api         *liteapi.Manager
	cookieJar   *cookies.Jar
	proxyDialer ProxyDialer
	identifier  Identifier

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
	focusService *focus.FocusService

	// autostarter is the bridge's autostarter.
	autostarter Autostarter

	// locator is the bridge's locator.
	locator Locator

	// errors contains errors encountered during startup.
	errors []error
}

// New creates a new bridge.
func New(
	apiURL string, // the URL of the API to use
	locator Locator, // the locator to provide paths to store data
	vault *vault.Vault, // the bridge's encrypted data store
	identifier Identifier, // the identifier to keep track of the user agent
	tlsReporter TLSReporter, // the TLS reporter to report TLS errors
	proxyDialer ProxyDialer, // the DoH dialer
	autostarter Autostarter, // the autostarter to manage autostart settings
	updater Updater, // the updater to fetch and install updates
	curVersion *semver.Version, // the current version of the bridge
) (*Bridge, error) {
	if vault.GetProxyAllowed() {
		proxyDialer.AllowProxy()
	} else {
		proxyDialer.DisallowProxy()
	}

	cookieJar, err := cookies.NewCookieJar(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	api := liteapi.New(
		liteapi.WithHostURL(apiURL),
		liteapi.WithAppVersion(constants.AppVersion),
		liteapi.WithCookieJar(cookieJar),
		liteapi.WithTransport(&http.Transport{DialTLSContext: proxyDialer.DialTLSContext}),
	)

	tlsConfig, err := loadTLSConfig(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	// TODO: Handle case that the gluon directory is missing!
	gluonDir, err := getGluonDir(vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gluon directory: %w", err)
	}

	imapServer, err := newIMAPServer(gluonDir, curVersion, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP server: %w", err)
	}

	smtpBackend, err := newSMTPBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP backend: %w", err)
	}

	smtpServer, err := newSMTPServer(smtpBackend, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP server: %w", err)
	}

	focusService, err := focus.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create focus service: %w", err)
	}

	bridge := &Bridge{
		vault: vault,
		users: make(map[string]*user.User),

		api:         api,
		cookieJar:   cookieJar,
		proxyDialer: proxyDialer,
		identifier:  identifier,

		tlsConfig:   tlsConfig,
		imapServer:  imapServer,
		smtpServer:  smtpServer,
		smtpBackend: smtpBackend,

		updater:       updater,
		curVersion:    curVersion,
		updateCheckCh: make(chan struct{}, 1),

		focusService: focusService,
		autostarter:  autostarter,
		locator:      locator,
	}

	api.AddStatusObserver(func(status liteapi.Status) {
		switch {
		case status == liteapi.StatusUp:
			go bridge.onStatusUp()

		case status == liteapi.StatusDown:
			go bridge.onStatusDown()
		}
	})

	api.AddErrorHandler(liteapi.AppVersionBadCode, func() {
		bridge.publish(events.UpdateForced{})
	})

	api.AddPreRequestHook(func(_ *resty.Client, req *resty.Request) error {
		req.SetHeader("User-Agent", bridge.identifier.GetUserAgent())
		return nil
	})

	go func() {
		for range tlsReporter.GetTLSIssueCh() {
			bridge.publish(events.TLSIssue{})
		}
	}()

	go func() {
		for range focusService.GetRaiseCh() {
			bridge.publish(events.Raise{})
		}
	}()

	go func() {
		for event := range imapServer.AddWatcher() {
			bridge.handleIMAPEvent(event)
		}
	}()

	if err := bridge.loadUsers(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load connected users: %w", err)
	}

	if err := bridge.serveIMAP(); err != nil {
		bridge.PushError(ErrServeIMAP)
	}

	if err := bridge.serveSMTP(); err != nil {
		bridge.PushError(ErrServeSMTP)
	}

	if err := bridge.watchForUpdates(); err != nil {
		bridge.PushError(ErrWatchUpdates)
	}

	return bridge, nil
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
		if err := user.Close(ctx); err != nil {
			logrus.WithError(err).Error("Failed to close user")
		}
	}

	// Persist the cookies.
	if err := bridge.cookieJar.PersistCookies(); err != nil {
		logrus.WithError(err).Error("Failed to persist cookies")
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

	for _, userID := range bridge.vault.GetUserIDs() {
		if _, ok := bridge.users[userID]; !ok {
			if vaultUser, err := bridge.vault.GetUser(userID); err != nil {
				logrus.WithError(err).Error("Failed to get user from vault")
			} else if err := bridge.loadUser(context.Background(), vaultUser); err != nil {
				logrus.WithError(err).Error("Failed to load user")
			}
		}
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
