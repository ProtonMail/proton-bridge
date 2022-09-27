package user

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

const chunkSize = 1 << 20

func (user *User) sync(ctx context.Context) error {
	user.notifyCh <- events.SyncStarted{
		UserID: user.ID(),
	}

	if err := user.syncLabels(ctx); err != nil {
		return fmt.Errorf("failed to sync labels: %w", err)
	}

	if err := user.syncMessages(ctx); err != nil {
		return fmt.Errorf("failed to sync messages: %w", err)
	}

	user.notifyCh <- events.SyncFinished{
		UserID: user.ID(),
	}

	if err := user.vault.SetSync(true); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	return nil
}

func (user *User) syncLabels(ctx context.Context) error {
	// Sync the system folders.
	system, err := user.client.GetLabels(ctx, liteapi.LabelTypeSystem)
	if err != nil {
		return err
	}

	for _, label := range system {
		user.updateCh <- newSystemMailboxCreatedUpdate(imap.LabelID(label.ID), label.Name)
	}

	// Create Folders/Labels mailboxes with a random ID and with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		user.updateCh <- newPlaceHolderMailboxCreatedUpdate(prefix)
	}

	// Sync the API folders.
	folders, err := user.client.GetLabels(ctx, liteapi.LabelTypeFolder)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		user.updateCh <- newMailboxCreatedUpdate(imap.LabelID(folder.ID), []string{folderPrefix, folder.Path})
	}

	// Sync the API labels.
	labels, err := user.client.GetLabels(ctx, liteapi.LabelTypeLabel)
	if err != nil {
		return err
	}

	for _, label := range labels {
		user.updateCh <- newMailboxCreatedUpdate(imap.LabelID(label.ID), []string{labelPrefix, label.Path})
	}

	return nil
}

func (user *User) syncMessages(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	metadata, err := user.client.GetAllMessageMetadata(ctx)
	if err != nil {
		return err
	}

	requests := xslices.Map(metadata, func(metadata liteapi.MessageMetadata) request {
		return request{
			messageID: metadata.ID,
			addrKR:    user.addrKRs[metadata.AddressID],
		}
	})

	flusher := newFlusher(user.ID(), user.updateCh, user.notifyCh, len(metadata), chunkSize)
	defer flusher.flush()

	if err := user.builder.Process(ctx, requests, func(req request, res *imap.MessageCreated, err error) error {
		if err != nil {
			return fmt.Errorf("failed to build message %s: %w", req.messageID, err)
		}

		flusher.push(res)

		return nil
	}); err != nil {
		return fmt.Errorf("failed to build messages: %w", err)
	}

	return nil
}

type flusher struct {
	userID string

	updates      []*imap.MessageCreated
	updateCh     chan<- imap.Update
	notifyCh     chan<- events.Event
	maxChunkSize int
	curChunkSize int

	count int
	total int
	start time.Time

	pushLock sync.Mutex
}

func newFlusher(userID string, updateCh chan<- imap.Update, notifyCh chan<- events.Event, total, maxChunkSize int) *flusher {
	return &flusher{
		userID:       userID,
		updateCh:     updateCh,
		notifyCh:     notifyCh,
		maxChunkSize: maxChunkSize,
		total:        total,
		start:        time.Now(),
	}
}

func (f *flusher) push(update *imap.MessageCreated) {
	f.pushLock.Lock()
	defer f.pushLock.Unlock()

	f.updates = append(f.updates, update)

	if f.curChunkSize += len(update.Literal); f.curChunkSize >= f.maxChunkSize {
		f.flush()
	}
}

func (f *flusher) flush() {
	if len(f.updates) == 0 {
		return
	}

	f.count += len(f.updates)
	f.updateCh <- imap.NewMessagesCreated(f.updates...)
	f.notifyCh <- newSyncProgress(f.userID, f.count, f.total, f.start)
	f.updates = nil
	f.curChunkSize = 0
}

func newSyncProgress(userID string, count, total int, start time.Time) events.SyncProgress {
	return events.SyncProgress{
		UserID:    userID,
		Progress:  float64(count) / float64(total),
		Elapsed:   time.Since(start),
		Remaining: time.Since(start) * time.Duration(total-count) / time.Duration(count),
	}
}

func getMessageCreatedUpdate(message liteapi.Message, literal []byte) (*imap.MessageCreated, error) {
	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return nil, err
	}

	flags := imap.NewFlagSet()

	if !message.Unread {
		flags = flags.Add(imap.FlagSeen)
	}

	if slices.Contains(message.LabelIDs, liteapi.StarredLabel) {
		flags = flags.Add(imap.FlagFlagged)
	}

	imapMessage := imap.Message{
		ID:    imap.MessageID(message.ID),
		Flags: flags,
		Date:  time.Unix(message.Time, 0),
	}

	return &imap.MessageCreated{
		Message:       imapMessage,
		Literal:       literal,
		LabelIDs:      imapLabelIDs(filterLabelIDs(message.LabelIDs)),
		ParsedMessage: parsedMessage,
	}, nil
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

func filterLabelIDs(labelIDs []string) []string {
	var filteredLabelIDs []string

	for _, labelID := range labelIDs {
		switch labelID {
		case liteapi.AllDraftsLabel, liteapi.AllSentLabel, liteapi.OutboxLabel:
			// ... skip ...

		default:
			filteredLabelIDs = append(filteredLabelIDs, labelID)
		}
	}

	return filteredLabelIDs
}
