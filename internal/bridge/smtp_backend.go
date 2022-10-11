package bridge

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/emersion/go-smtp"
)

type smtpBackend struct {
	users     map[string]*user.User
	usersLock sync.RWMutex
}

func newSMTPBackend() (*smtpBackend, error) {
	return &smtpBackend{
		users: make(map[string]*user.User),
	}, nil
}

func (backend *smtpBackend) Login(state *smtp.ConnectionState, email, password string) (smtp.Session, error) {
	backend.usersLock.RLock()
	defer backend.usersLock.RUnlock()

	for _, user := range backend.users {
		session, err := user.NewSMTPSession(email, []byte(password))
		if err != nil {
			continue
		}

		return session, nil
	}

	return nil, ErrNoSuchUser
}

func (backend *smtpBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, ErrNotImplemented
}

// addUser adds the given user to the backend.
// It returns an error if a user with the same ID already exists.
func (backend *smtpBackend) addUser(newUser *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	if _, ok := backend.users[newUser.ID()]; ok {
		return ErrUserAlreadyExists
	}

	backend.users[newUser.ID()] = newUser

	return nil
}

// removeUser removes the given user from the backend.
// It returns an error if the user doesn't exist.
func (backend *smtpBackend) removeUser(user *user.User) error {
	backend.usersLock.Lock()
	defer backend.usersLock.Unlock()

	if _, ok := backend.users[user.ID()]; !ok {
		return ErrNoSuchUser
	}

	delete(backend.users, user.ID())

	return nil
}
