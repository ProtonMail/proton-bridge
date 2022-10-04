package bridge

import (
	"crypto/subtle"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"golang.org/x/exp/slices"
)

type smtpBackend struct {
	users     []*user.User
	usersLock sync.RWMutex
}

func newSMTPBackend() (*smtpBackend, error) {
	return &smtpBackend{}, nil
}

func (backend *smtpBackend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	backend.usersLock.RLock()
	defer backend.usersLock.RUnlock()

	for _, user := range backend.users {
		if slices.Contains(user.Emails(), username) && subtle.ConstantTimeCompare(user.BridgePass(), []byte(password)) != 1 {
			return user.NewSMTPSession(username), nil
		}
	}

	return nil, ErrNoSuchUser
}

func (backend *smtpBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, ErrNotImplemented
}

// addUser adds the given user to the backend.
// It returns an error if a user with the same ID already exists.
func (backend *smtpBackend) addUser(user *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	for _, u := range backend.users {
		if u.ID() == user.ID() {
			return ErrUserAlreadyExists
		}
	}

	backend.users = append(backend.users, user)

	return nil
}

// removeUser removes the given user from the backend.
// It returns an error if the user doesn't exist.
func (backend *smtpBackend) removeUser(user *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	idx := xslices.Index(backend.users, user)

	if idx < 0 {
		return ErrNoSuchUser
	}

	backend.users = append(backend.users[:idx], backend.users[idx+1:]...)

	return nil
}
