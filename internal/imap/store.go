// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package imap

import (
	"io"
	"net/mail"
	"net/textproto"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	pkgMsg "github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type storeUserProvider interface {
	UserID() string
	GetSpaceKB() (usedSpace, maxSpace uint32, err error)
	GetMaxUpload() (int64, error)

	GetAddress(addressID string) (storeAddressProvider, error)

	CreateDraft(
		kr *crypto.KeyRing,
		message *pmapi.Message,
		attachmentReaders []io.Reader,
		attachedPublicKey,
		attachedPublicKeyName string,
		parentID string) (*pmapi.Message, []*pmapi.Attachment, error)

	SetChangeNotifier(store.ChangeNotifier)
}

type storeAddressProvider interface {
	AddressString() string
	AddressID() string
	APIAddress() *pmapi.Address

	CreateMailbox(name string) error
	ListMailboxes() []storeMailboxProvider
	GetMailbox(name string) (storeMailboxProvider, error)
}

type storeMailboxProvider interface {
	LabelID() string
	Name() string
	Color() string
	IsSystem() bool
	IsFolder() bool
	IsLabel() bool
	UIDValidity() uint32

	Rename(newName string) error
	Delete() error

	GetAPIIDsFromUIDRange(start, stop uint32) ([]string, error)
	GetAPIIDsFromSequenceRange(start, stop uint32) ([]string, error)
	GetLatestAPIID() (string, error)
	GetNextUID() (uint32, error)
	GetDeletedAPIIDs() ([]string, error)
	GetCounts() (dbTotal, dbUnread, dbUnreadSeqNum uint, err error)
	GetUIDList(apiIDs []string) *uidplus.OrderedSeq
	GetUIDByHeader(header *mail.Header) uint32
	GetDelimiter() string

	GetMessage(apiID string) (storeMessageProvider, error)
	LabelMessages(apiID []string) error
	UnlabelMessages(apiID []string) error
	MarkMessagesRead(apiID []string) error
	MarkMessagesUnread(apiID []string) error
	MarkMessagesStarred(apiID []string) error
	MarkMessagesUnstarred(apiID []string) error
	MarkMessagesDeleted(apiID []string) error
	MarkMessagesUndeleted(apiID []string) error
	ImportMessage(enc []byte, seen bool, labelIDs []string, flags, time int64) (string, error)
	RemoveDeleted(apiIDs []string) error
}

type storeMessageProvider interface {
	ID() string
	UID() (uint32, error)
	SequenceNumber() (uint32, error)
	Message() *pmapi.Message
	IsMarkedDeleted() bool

	GetHeader() ([]byte, error)
	GetRFC822() ([]byte, error)
	GetRFC822Size() (uint32, error)
	GetMIMEHeaderFast() textproto.MIMEHeader
	IsFullHeaderCached() bool
	GetBodyStructure() (*pkgMsg.BodyStructure, error)
}

type storeUserWrap struct {
	*store.Store
}

// newStoreUserWrap wraps store struct into local storeUserWrap to implement local
// interface. The problem is that store returns the store package's Address type, so
// every method that returns an address has to be overridden to fulfill the interface.
// The same is true for other store structs i.e. storeAddress or storeMailbox.
func newStoreUserWrap(store *store.Store) *storeUserWrap {
	return &storeUserWrap{Store: store}
}

func (s *storeUserWrap) GetAddress(addressID string) (storeAddressProvider, error) {
	address, err := s.Store.GetAddress(addressID)
	if err != nil {
		return nil, err
	}
	return newStoreAddressWrap(address), nil //nolint:typecheck missing methods are inherited
}

type storeAddressWrap struct {
	*store.Address
}

func newStoreAddressWrap(address *store.Address) *storeAddressWrap {
	return &storeAddressWrap{Address: address}
}

func (s *storeAddressWrap) ListMailboxes() []storeMailboxProvider {
	mailboxes := []storeMailboxProvider{}
	for _, mailbox := range s.Address.ListMailboxes() {
		mailboxes = append(mailboxes, newStoreMailboxWrap(mailbox)) //nolint:typecheck missing methods are inherited
	}
	return mailboxes
}

func (s *storeAddressWrap) GetMailbox(name string) (storeMailboxProvider, error) {
	mailbox, err := s.Address.GetMailbox(name)
	if err != nil {
		return nil, err
	}
	return newStoreMailboxWrap(mailbox), nil //nolint:typecheck missing methods are inherited
}

type storeMailboxWrap struct {
	*store.Mailbox
}

func newStoreMailboxWrap(mailbox *store.Mailbox) *storeMailboxWrap {
	return &storeMailboxWrap{Mailbox: mailbox}
}

func (s *storeMailboxWrap) GetMessage(apiID string) (storeMessageProvider, error) {
	return s.Mailbox.GetMessage(apiID)
}
