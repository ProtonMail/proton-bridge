package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

type buildRes struct {
	messageID string
	addressID string
	update    *imap.MessageCreated
}

func defaultJobOpts() message.JobOptions {
	return message.JobOptions{
		IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
		SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
		AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
		AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
		AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
		AddMessageIDReference:  true, // Whether to include the MessageID in References.
	}
}

func (user *User) buildRFC822(ctx context.Context, metadata liteapi.MessageMetadata) (*buildRes, error) {
	msg, err := user.client.GetMessage(ctx, metadata.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message %s: %w", metadata.ID, err)
	}

	attData, err := user.attPool.ProcessAll(ctx, xslices.Map(msg.Attachments, func(att liteapi.Attachment) string { return att.ID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments for message %s: %w", metadata.ID, err)
	}

	literal, err := message.BuildRFC822(user.addrKRs[msg.AddressID], msg, attData, defaultJobOpts())
	if err != nil {
		return nil, fmt.Errorf("failed to build message %s: %w", metadata.ID, err)
	}

	update, err := newMessageCreatedUpdate(metadata, literal)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP update for message %s: %w", metadata.ID, err)
	}

	return &buildRes{
		messageID: metadata.ID,
		addressID: metadata.AddressID,
		update:    update,
	}, nil
}

func newMessageCreatedUpdate(message liteapi.MessageMetadata, literal []byte) (*imap.MessageCreated, error) {
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
		LabelIDs:      mapTo[string, imap.LabelID](xslices.Filter(message.LabelIDs, wantLabelID)),
		ParsedMessage: parsedMessage,
	}, nil
}
