package user

import (
	"context"
	"runtime"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/pool"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

var (
	DefaultEventPeriod = 20 * time.Second
	DefaultEventJitter = 20 * time.Second
)

// TODO: Is it bad to store the key pass in the user? Any worse than storing private keys?
type User struct {
	vault   *vault.User
	client  *liteapi.Client
	builder *pool.Pool[request, *imap.MessageCreated]

	apiUser   liteapi.User
	addresses []liteapi.Address
	settings  liteapi.MailSettings

	notifyCh chan events.Event
	updateCh chan imap.Update

	userKR   *crypto.KeyRing
	addrKRs  map[string]*crypto.KeyRing
	imapConn *imapConnector
}

func New(
	ctx context.Context,
	vault *vault.User,
	client *liteapi.Client,
	apiUser liteapi.User,
	apiAddrs []liteapi.Address,
	userKR *crypto.KeyRing,
	addrKRs map[string]*crypto.KeyRing,
) (*User, error) {
	if vault.EventID() == "" {
		eventID, err := client.GetLatestEventID(ctx)
		if err != nil {
			return nil, err
		}

		if err := vault.UpdateEventID(eventID); err != nil {
			return nil, err
		}
	}

	settings, err := client.GetMailSettings(ctx)
	if err != nil {
		return nil, err
	}

	user := &User{
		apiUser:   apiUser,
		addresses: apiAddrs,
		settings:  settings,

		vault:   vault,
		client:  client,
		builder: newBuilder(client, runtime.NumCPU()*runtime.NumCPU(), runtime.NumCPU()*runtime.NumCPU()),

		notifyCh: make(chan events.Event),
		updateCh: make(chan imap.Update),

		userKR:  userKR,
		addrKRs: addrKRs,
	}

	// When we receive an auth object, we update it in the store.
	// This will be used to authorize the user on the next run.
	client.AddAuthHandler(func(auth liteapi.Auth) {
		if err := user.vault.UpdateAuth(auth.UID, auth.RefreshToken); err != nil {
			logrus.WithError(err).Error("Failed to update auth")
		}
	})

	// When we are deauthorized, we send a deauth event to the notify channel.
	// Bridge will catch this and log the user out.
	client.AddDeauthHandler(func() {
		user.notifyCh <- events.UserDeauth{
			UserID: user.ID(),
		}
	})

	// When we receive an API event, we attempt to handle it. If successful, we send the event to the event channel.
	go func() {
		for event := range user.client.NewEventStreamer(DefaultEventPeriod, DefaultEventJitter, vault.EventID()).Subscribe() {
			if err := user.handleAPIEvent(event); err != nil {
				logrus.WithError(err).Error("Failed to handle event")
			} else {
				if err := user.vault.UpdateEventID(event.EventID); err != nil {
					logrus.WithError(err).Error("Failed to update event ID")
				}
			}
		}
	}()

	// TODO: Use a proper sync manager! (if partial sync, pickup from where we last stopped)
	if !vault.HasSync() {
		go user.sync(context.Background())
	}

	return user, nil
}

func (user *User) ID() string {
	return user.apiUser.ID
}

func (user *User) Name() string {
	return user.apiUser.Name
}

func (user *User) Match(query string) bool {
	if query == user.Name() {
		return true
	}

	return slices.Contains(user.Addresses(), query)
}

func (user *User) Addresses() []string {
	return xslices.Map(
		sort(user.addresses, func(a, b liteapi.Address) bool {
			return a.Order < b.Order
		}),
		func(address liteapi.Address) string {
			return address.Email
		},
	)
}

func (user *User) GluonID() string {
	return user.vault.GluonID()
}

func (user *User) GluonKey() []byte {
	return user.vault.GluonKey()
}

func (user *User) BridgePass() string {
	return user.vault.BridgePass()
}

func (user *User) UsedSpace() int {
	return user.apiUser.UsedSpace
}

func (user *User) MaxSpace() int {
	return user.apiUser.MaxSpace
}

// GetNotifyCh returns a channel which notifies of events happening to the user (such as deauth, address change)
func (user *User) GetNotifyCh() <-chan events.Event {
	return user.notifyCh
}

func (user *User) NewGluonConnector(ctx context.Context) (connector.Connector, error) {
	if user.imapConn != nil {
		if err := user.imapConn.Close(ctx); err != nil {
			return nil, err
		}
	}

	user.imapConn = newIMAPConnector(user.client, user.updateCh, user.Addresses(), user.vault.BridgePass())

	return user.imapConn, nil
}

func (user *User) NewSMTPSession(username string) (smtp.Session, error) {
	return newSMTPSession(user.client, username, user.addresses, user.userKR, user.addrKRs, user.settings), nil
}

func (user *User) Logout(ctx context.Context) error {
	return user.client.AuthDelete(ctx)
}

func (user *User) Close(ctx context.Context) error {
	// Close the user's IMAP connectors.
	if user.imapConn != nil {
		if err := user.imapConn.Close(ctx); err != nil {
			return err
		}
	}

	// Close the user's message builder.
	user.builder.Done()

	// Close the user's API client.
	user.client.Close()

	// Close the user's notify channel.
	close(user.notifyCh)

	return nil
}

// sort returns the slice, sorted by the given callback.
func sort[T any](slice []T, less func(a, b T) bool) []T {
	slices.SortFunc(slice, less)

	return slice
}
