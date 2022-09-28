package tests

import (
	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/rfc822"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

type API interface {
	SetMinAppVersion(*semver.Version)

	GetHostURL() string
	AddCallWatcher(func(server.Call), ...string)

	AddUser(username, password, address string) (string, string, error)
	AddAddress(userID, address, password string) (string, error)
	RemoveAddress(userID, addrID string) error
	RevokeUser(userID string) error

	GetLabels(userID string) ([]liteapi.Label, error)
	AddLabel(userID, name string, labelType liteapi.LabelType) (string, error)

	GetMessages(userID string) ([]liteapi.Message, error)
	AddMessage(userID, addrID string, labelIDs []string, sender, recipient, subject, body string, mimeType rfc822.MIMEType, read, starred bool) (string, error)

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
