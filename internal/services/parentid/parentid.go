// Copyright (c) 2026 Proton AG
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

// Package parentid resolves a Proton message ParentID from RFC 5322
// In-Reply-To / References headers. Used by both the SMTP send path and the
// IMAP draft-creation path so they share a single source of truth for thread
// linkage and the "don't pick a draft as parent" rule (GODT-2935).
package parentid

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
)

// Client is the narrow API surface the parent-lookup needs. *proton.Client
// satisfies it directly; callers that hold a wider interface should ensure
// it exposes the same method.
type Client interface {
	GetMessageMetadata(ctx context.Context, filter proton.MessageFilter) ([]proton.MessageMetadata, error)
}

// Find returns the ParentID of the message the given references chain
// targets, plus a list of draft IDs encountered as internal references
// (callers in the send path use these to clean up post-send; the draft
// creation path should ignore them).
//
// Internal references (those at the InternalIDDomain placeholder) are tried
// first. Drafts found there are recorded in draftsToDelete but never picked
// as the parent. If no internal reference resolves to a non-draft parent,
// the last external reference is searched; the result must be a sent or
// received message (drafts are rejected per GODT-2935).
func Find(
	ctx context.Context,
	client Client,
	authAddrID string,
	addrMode usertypes.AddressMode,
	references []string,
) (parentID string, draftsToDelete []string, err error) {
	var (
		internal []string
		external []string
	)

	for _, ref := range references {
		if strings.Contains(ref, message.InternalIDDomain) {
			internal = append(internal, strings.TrimSuffix(ref, "@"+message.InternalIDDomain))
		} else {
			external = append(external, ref)
		}
	}

	for _, internal := range internal {
		var addrID string

		if addrMode == usertypes.AddressModeSplit {
			addrID = authAddrID
		}

		metadata, err := client.GetMessageMetadata(ctx, proton.MessageFilter{
			ID:        []string{internal},
			AddressID: addrID,
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to get message metadata: %w", err)
		}

		for _, metadata := range metadata {
			if !metadata.IsDraft() {
				parentID = metadata.ID
			} else {
				// We need to record this ID to delete later after the message has been sent successfully. This is
				// required for Apple Mail to correctly delete a draft when a draft is created in Apple Mail, then
				// edited on the web, edited again in Apple Mail and then Send from Apple Mail. If we don't
				// delete the referenced draft it is never deleted from the drafts folder.
				draftsToDelete = append(draftsToDelete, metadata.ID)
			}
		}
	}

	if parentID == "" && len(external) > 0 {
		var addrID string

		if addrMode == usertypes.AddressModeSplit {
			addrID = authAddrID
		}

		metadata, err := client.GetMessageMetadata(ctx, proton.MessageFilter{
			ExternalID: external[len(external)-1],
			AddressID:  addrID,
		})
		if err != nil {
			return "", nil, fmt.Errorf("failed to get message metadata: %w", err)
		}

		switch len(metadata) {
		case 1:
			// found exactly one parent
			// We can only reference messages that have been sent or received. If this message is a draft
			// it needs to be ignored.
			if metadata[0].Flags.Has(proton.MessageFlagSent) || metadata[0].Flags.Has(proton.MessageFlagReceived) {
				parentID = metadata[0].ID
			}
		case 0:
			// found no parents
		default:
			// found multiple parents, search through metadata to try to find a singular parent that
			// was sent by this account.
			for _, metadata := range metadata {
				if metadata.Flags.Has(proton.MessageFlagSent) {
					parentID = metadata.ID
					break
				}
			}
		}
	}

	return parentID, draftsToDelete, nil
}
