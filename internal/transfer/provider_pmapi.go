// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package transfer

import (
	"sort"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

// PMAPIProvider implements import and export to/from ProtonMail server.
type PMAPIProvider struct {
	clientConfig  *pmapi.ClientConfig
	clientManager ClientManager
	userID        string
	addressID     string
	keyRing       *crypto.KeyRing

	nextImportRequests     map[string]*pmapi.ImportMsgReq // Key is msg transfer ID.
	nextImportRequestsSize int

	timeIt *timeIt
}

// NewPMAPIProvider returns new PMAPIProvider.
func NewPMAPIProvider(config *pmapi.ClientConfig, clientManager ClientManager, userID, addressID string) (*PMAPIProvider, error) {
	provider := &PMAPIProvider{
		clientConfig:  config,
		clientManager: clientManager,
		userID:        userID,
		addressID:     addressID,

		nextImportRequests:     map[string]*pmapi.ImportMsgReq{},
		nextImportRequestsSize: 0,

		timeIt: newTimeIt("pmapi"),
	}

	if addressID != "" {
		keyRing, err := clientManager.GetClient(userID).KeyRingForAddressID(addressID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get key ring")
		}
		provider.keyRing = keyRing
	}

	return provider, nil
}

func (p *PMAPIProvider) client() pmapi.Client {
	return p.clientManager.GetClient(p.userID)
}

// ID returns identifier of current setup of PMAPI provider.
// Identification is unique per user.
func (p *PMAPIProvider) ID() string {
	return p.userID
}

// Mailboxes returns all available labels in ProtonMail account.
func (p *PMAPIProvider) Mailboxes(includeEmpty, includeAllMail bool) ([]Mailbox, error) {
	labels, err := p.client().ListLabels()
	if err != nil {
		return nil, err
	}
	sortedLabels := byFoldersLabels(labels)
	sort.Sort(sortedLabels)

	emptyLabelsMap := map[string]bool{}
	if !includeEmpty {
		messagesCounts, err := p.client().CountMessages(p.addressID)
		if err != nil {
			return nil, err
		}
		for _, messagesCount := range messagesCounts {
			if messagesCount.Total == 0 {
				emptyLabelsMap[messagesCount.LabelID] = true
			}
		}
	}

	mailboxes := []Mailbox{}
	for _, mailbox := range getSystemMailboxes(includeAllMail) {
		if !includeEmpty && emptyLabelsMap[mailbox.ID] {
			continue
		}

		mailboxes = append(mailboxes, mailbox)
	}
	for _, label := range sortedLabels {
		if !includeEmpty && emptyLabelsMap[label.ID] {
			continue
		}

		mailboxes = append(mailboxes, Mailbox{
			ID:          label.ID,
			Name:        label.Name,
			Color:       label.Color,
			IsExclusive: label.Exclusive == 1,
		})
	}
	return mailboxes, nil
}

func getSystemMailboxes(includeAllMail bool) []Mailbox {
	mailboxes := []Mailbox{
		{ID: pmapi.InboxLabel, Name: "Inbox", IsExclusive: true},
		{ID: pmapi.DraftLabel, Name: "Drafts", IsExclusive: true},
		{ID: pmapi.SentLabel, Name: "Sent", IsExclusive: true},
		{ID: pmapi.StarredLabel, Name: "Starred", IsExclusive: true},
		{ID: pmapi.ArchiveLabel, Name: "Archive", IsExclusive: true},
		{ID: pmapi.SpamLabel, Name: "Spam", IsExclusive: true},
		{ID: pmapi.TrashLabel, Name: "Trash", IsExclusive: true},
	}

	if includeAllMail {
		mailboxes = append(mailboxes, Mailbox{
			ID:          pmapi.AllMailLabel,
			Name:        "All Mail",
			IsExclusive: true,
		})
	}

	return mailboxes
}

type byFoldersLabels []*pmapi.Label

func (l byFoldersLabels) Len() int {
	return len(l)
}

func (l byFoldersLabels) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Less sorts first folders, then labels, by user order.
func (l byFoldersLabels) Less(i, j int) bool {
	if l[i].Exclusive == 1 && l[j].Exclusive == 0 {
		return true
	}
	if l[i].Exclusive == 0 && l[j].Exclusive == 1 {
		return false
	}
	return l[i].Order < l[j].Order
}
