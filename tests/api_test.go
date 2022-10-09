package tests

import (
	"github.com/Masterminds/semver/v3"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

type API interface {
	SetMinAppVersion(*semver.Version)

	GetHostURL() string
	AddCallWatcher(func(server.Call), ...string)

	CreateUser(username, address string, password []byte) (string, string, error)
	CreateAddress(userID, address string, password []byte) (string, error)
	RemoveAddress(userID, addrID string) error
	RevokeUser(userID string) error

	GetLabels(userID string) ([]liteapi.Label, error)
	CreateLabel(userID, name string, labelType liteapi.LabelType) (string, error)

	CreateMessage(userID, addrID string, literal []byte, flags liteapi.MessageFlag, unread, starred bool) (string, error)
	LabelMessage(userID, messageID, labelID string) error
	UnlabelMessage(userID, messageID, labelID string) error

	Close()
}

type fakeAPI struct {
	*server.Server
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		Server: server.NewTLS(),
	}
}
