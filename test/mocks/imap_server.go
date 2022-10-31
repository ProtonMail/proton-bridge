// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package mocks

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapbackend "github.com/emersion/go-imap/backend"
	imapserver "github.com/emersion/go-imap/server"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type IMAPServer struct {
	Username string
	Password string
	Host     string
	Port     string

	mailboxes []string
	messages  map[string][]*imap.Message // Key is mailbox.
	server    *imapserver.Server
}

func NewIMAPServer(username, password, host, port string) *IMAPServer {
	return &IMAPServer{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,

		mailboxes: []string{},
		messages:  map[string][]*imap.Message{},
	}
}

func (s *IMAPServer) AddMailbox(mailboxName string) {
	s.mailboxes = append(s.mailboxes, mailboxName)
	s.messages[strings.ToLower(mailboxName)] = []*imap.Message{}
}

func (s *IMAPServer) AddMessage(mailboxName string, message *imap.Message) {
	mailboxName = strings.ToLower(mailboxName)
	s.messages[mailboxName] = append(s.messages[mailboxName], message)
}

func (s *IMAPServer) Start() {
	server := imapserver.New(&IMAPBackend{server: s})
	server.Addr = net.JoinHostPort(s.Host, s.Port)
	server.AllowInsecureAuth = true
	server.ErrorLog = logrus.WithField("pkg", "imap-server")
	server.Debug = logrus.WithField("pkg", "imap-server").WriterLevel(logrus.DebugLevel)
	server.AutoLogout = 30 * time.Minute

	s.server = server

	go func() {
		err := server.ListenAndServe()
		logrus.WithError(err).Warn("IMAP server stopped")
	}()

	time.Sleep(100 * time.Millisecond)
}

func (s *IMAPServer) Stop() {
	_ = s.server.Close()
}

type IMAPBackend struct {
	server *IMAPServer
}

func (b *IMAPBackend) Login(connInfo *imap.ConnInfo, username, password string) (imapbackend.User, error) {
	if username != b.server.Username || password != b.server.Password {
		return nil, errors.New("invalid credentials")
	}
	return &IMAPUser{
		server:   b.server,
		username: username,
	}, nil
}

type IMAPUser struct {
	server   *IMAPServer
	username string
}

func (u *IMAPUser) Username() string {
	return u.username
}

func (u *IMAPUser) ListMailboxes(subscribed bool) ([]imapbackend.Mailbox, error) {
	mailboxes := []imapbackend.Mailbox{}
	for _, mailboxName := range u.server.mailboxes {
		mailboxes = append(mailboxes, &IMAPMailbox{
			server: u.server,
			name:   mailboxName,
		})
	}
	return mailboxes, nil
}

func (u *IMAPUser) GetMailbox(name string) (imapbackend.Mailbox, error) {
	name = strings.ToLower(name)
	_, ok := u.server.messages[name]
	if !ok {
		return nil, fmt.Errorf("mailbox %s not found", name)
	}
	return &IMAPMailbox{
		server: u.server,
		name:   name,
	}, nil
}

func (u *IMAPUser) CreateMailbox(name string) error {
	return errors.New("not supported: create mailbox")
}

func (u *IMAPUser) DeleteMailbox(name string) error {
	return errors.New("not supported: delete mailbox")
}

func (u *IMAPUser) RenameMailbox(existingName, newName string) error {
	return errors.New("not supported: rename mailbox")
}

func (u *IMAPUser) Logout() error {
	return nil
}

type IMAPMailbox struct {
	server     *IMAPServer
	name       string
	attributes []string
}

func (m *IMAPMailbox) Name() string {
	return m.name
}

func (m *IMAPMailbox) Info() (*imap.MailboxInfo, error) {
	return &imap.MailboxInfo{
		Name:       m.name,
		Attributes: m.attributes,
	}, nil
}

func (m *IMAPMailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	status := imap.NewMailboxStatus(m.name, items)
	status.UidValidity = 1
	status.Messages = uint32(len(m.server.messages[m.name]))
	return status, nil
}

func (m *IMAPMailbox) SetSubscribed(subscribed bool) error {
	return errors.New("not supported: set subscribed")
}

func (m *IMAPMailbox) Check() error {
	return errors.New("not supported: check")
}

func (m *IMAPMailbox) ListMessages(uid bool, seqset *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer func() {
		close(ch)
	}()

	for index, message := range m.server.messages[m.name] {
		seqNum := uint32(index + 1)
		var id uint32
		if uid {
			id = message.Uid
		} else {
			id = seqNum
		}
		if seqset.Contains(id) {
			msg := imap.NewMessage(seqNum, items)
			msg.Envelope = message.Envelope
			msg.BodyStructure = message.BodyStructure
			msg.Body = message.Body
			msg.Size = message.Size
			msg.Uid = message.Uid

			ch <- msg
		}
	}
	return nil
}

func (m *IMAPMailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	return nil, errors.New("not supported: search")
}

func (m *IMAPMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return errors.New("not supported: create")
}

func (m *IMAPMailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	return errors.New("not supported: update flags")
}

func (m *IMAPMailbox) CopyMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	return errors.New("not supported: copy")
}

func (m *IMAPMailbox) Expunge() error {
	return errors.New("not supported: expunge")
}
