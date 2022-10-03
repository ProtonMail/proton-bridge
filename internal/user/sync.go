package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"gitlab.protontech.ch/go/liteapi"
)

const chunkSize = 1 << 20

func (user *User) syncLabels(ctx context.Context, addrIDs ...string) error {
	// Sync the system folders.
	system, err := user.client.GetLabels(ctx, liteapi.LabelTypeSystem)
	if err != nil {
		return err
	}

	for _, label := range xslices.Filter(system, func(label liteapi.Label) bool { return wantLabelID(label.ID) }) {
		for _, addrID := range addrIDs {
			user.updateCh[addrID].Enqueue(newSystemMailboxCreatedUpdate(imap.LabelID(label.ID), label.Name))
		}
	}

	// Create Folders/Labels mailboxes with a random ID and with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, addrID := range addrIDs {
			user.updateCh[addrID].Enqueue(newPlaceHolderMailboxCreatedUpdate(prefix))
		}
	}

	// Sync the API folders.
	folders, err := user.client.GetLabels(ctx, liteapi.LabelTypeFolder)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		for _, addrID := range addrIDs {
			user.updateCh[addrID].Enqueue(newMailboxCreatedUpdate(imap.LabelID(folder.ID), []string{folderPrefix, folder.Path}))
		}
	}

	// Sync the API labels.
	labels, err := user.client.GetLabels(ctx, liteapi.LabelTypeLabel)
	if err != nil {
		return err
	}

	for _, label := range labels {
		for _, addrID := range addrIDs {
			user.updateCh[addrID].Enqueue(newMailboxCreatedUpdate(imap.LabelID(label.ID), []string{labelPrefix, label.Path}))
		}
	}

	return nil
}

func (user *User) syncMessages(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Determine which messages to sync.
	// TODO: This needs to be done better using the new API route to retrieve just the message IDs.
	metadata, err := user.client.GetAllMessageMetadata(ctx, nil)
	if err != nil {
		return err
	}

	// If in split mode, we need to send each message to a different IMAP connector.
	isSplitMode := user.vault.AddressMode() == vault.SplitMode

	// Collect the build requests -- we need:
	// - the message ID to build,
	// - the keyring to decrypt the message,
	// - and the address to send the message to (for split mode).
	requests := xslices.Map(metadata, func(metadata liteapi.MessageMetadata) request {
		var addressID string

		if isSplitMode {
			addressID = metadata.AddressID
		} else {
			addressID = user.apiAddrs.primary()
		}

		return request{
			messageID: metadata.ID,
			addressID: addressID,
			addrKR:    user.addrKRs[metadata.AddressID],
		}
	})

	// Create the flushers, one per update channel.
	flushers := make(map[string]*flusher)

	for addrID, updateCh := range user.updateCh {
		flusher := newFlusher(user.ID(), updateCh, user.eventCh, len(requests), chunkSize)
		defer flusher.flush()

		flushers[addrID] = flusher
	}

	// Build the messages and send them to the correct flusher.
	if err := user.builder.Process(ctx, requests, func(req request, res *imap.MessageCreated, err error) error {
		if err != nil {
			return fmt.Errorf("failed to build message %s: %w", req.messageID, err)
		}

		flushers[req.addressID].push(res)

		return nil
	}); err != nil {
		return fmt.Errorf("failed to build messages: %w", err)
	}

	return nil
}

func (user *User) syncWait() {
	for _, updateCh := range user.updateCh {
		waiter := imap.NewNoop()
		defer waiter.Wait()

		updateCh.Enqueue(waiter)
	}
}

func newSystemMailboxCreatedUpdate(labelID imap.LabelID, labelName string) *imap.MailboxCreated {
	if strings.EqualFold(labelName, imap.Inbox) {
		labelName = imap.Inbox
	}

	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           []string{labelName},
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(imap.AttrNoInferiors),
	})
}

func newPlaceHolderMailboxCreatedUpdate(labelName string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             imap.LabelID(uuid.NewString()),
		Name:           []string{labelName},
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(imap.AttrNoSelect),
	})
}

func newMailboxCreatedUpdate(labelID imap.LabelID, labelName []string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           labelName,
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(),
	})
}

func wantLabelID(labelID string) bool {
	switch labelID {
	case liteapi.AllDraftsLabel, liteapi.AllSentLabel, liteapi.OutboxLabel:
		return false

	default:
		return true
	}
}
