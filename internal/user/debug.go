// Copyright (c) 2023 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type DiagnosticMetadata struct {
	MessageIDs       []string
	Metadata         []proton.MessageMetadata
	FailedMessageIDs xmaps.Set[string]
}

type AccountMailboxMap map[string][]MailboxMessage

type MailboxMessage struct {
	ID    string
	Flags imap.FlagSet
}

func (apm DiagnosticMetadata) BuildMailboxToMessageMap(user *User) (map[string]AccountMailboxMap, error) {
	return safe.RLockRetErr(func() (map[string]AccountMailboxMap, error) {
		result := make(map[string]AccountMailboxMap)

		mode := user.GetAddressMode()
		primaryAddrID, err := getPrimaryAddr(user.apiAddrs)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary addr for user: %w", err)
		}

		getAccount := func(addrID string) (AccountMailboxMap, bool) {
			if mode == vault.CombinedMode {
				addrID = primaryAddrID.ID
			}

			addr := user.apiAddrs[addrID]
			if addr.Status != proton.AddressStatusEnabled {
				return nil, false
			}

			v, ok := result[addr.Email]
			if !ok {
				result[addr.Email] = make(AccountMailboxMap)
				v = result[addr.Email]
			}

			return v, true
		}

		for _, metadata := range apm.Metadata {
			for _, label := range metadata.LabelIDs {
				details, ok := user.apiLabels[label]
				if !ok {
					logrus.Warnf("User %v has message with unknown label '%v'", user.Name(), label)
					continue
				}

				if !wantLabel(details) {
					continue
				}

				account, enabled := getAccount(metadata.AddressID)
				if !enabled {
					continue
				}

				var mboxName string
				if details.Type == proton.LabelTypeSystem {
					mboxName = details.Name
				} else {
					mboxName = strings.Join(getMailboxName(details), "/")
				}

				mboxMessage := MailboxMessage{
					ID:    metadata.ID,
					Flags: buildFlagSetFromMessageMetadata(metadata),
				}

				if v, ok := account[mboxName]; ok {
					account[mboxName] = append(v, mboxMessage)
				} else {
					account[mboxName] = []MailboxMessage{mboxMessage}
				}
			}
		}
		return result, nil
	}, user.apiAddrsLock, user.apiLabelsLock)
}

func (user *User) GetDiagnosticMetadata(ctx context.Context) (DiagnosticMetadata, error) {
	failedMessages := xmaps.SetFromSlice(user.vault.SyncStatus().FailedMessageIDs)

	messageIDs, err := user.client.GetMessageIDs(ctx, "")
	if err != nil {
		return DiagnosticMetadata{}, err
	}

	meta := make([]proton.MessageMetadata, 0, len(messageIDs))

	for _, m := range xslices.Chunk(messageIDs, 100) {
		metadata, err := user.client.GetMessageMetadataPage(ctx, 0, len(m), proton.MessageFilter{ID: m})
		if err != nil {
			return DiagnosticMetadata{}, err
		}

		meta = append(meta, metadata...)
	}

	return DiagnosticMetadata{
		MessageIDs:       messageIDs,
		Metadata:         meta,
		FailedMessageIDs: failedMessages,
	}, nil
}
