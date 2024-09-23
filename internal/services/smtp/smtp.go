// Copyright (c) 2024 Proton AG
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

package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/rfc5322"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// smtpSendMail sends an email from the given address to the given recipients.
func (s *Service) smtpSendMail(ctx context.Context, authID string, from string, to []string, r io.Reader) error {
	fromAddr, err := s.identityState.GetAddr(from)
	if err != nil {
		return ErrInvalidReturnPath
	}

	emails := xslices.Map(s.identityState.AddressesSorted, func(addr proton.Address) string {
		return addr.Email
	})

	// Read the message to send.
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	// If running a QA build, dump to disk.
	if err := debugDumpToDisk(b); err != nil {
		s.log.WithError(err).Warn("Failed to dump message to disk")
	}

	// Compute the hash of the message (to match it against SMTP messages).
	hash, err := sendrecorder.GetMessageHash(b)
	if err != nil {
		return err
	}

	// Check if we already tried to send this message recently.
	s.log.Debug("Checking for duplicate message")
	srID, ok, err := s.recorder.TryInsertWait(ctx, hash, to, time.Now().Add(90*time.Second))
	if err != nil {
		return fmt.Errorf("failed to check send hash: %w", err)
	} else if !ok {
		s.log.Warn("A duplicate message was already sent recently, skipping")
		return nil
	}

	// Create a new message parser from the reader.
	parser, err := parser.New(bytes.NewReader(b))
	if err != nil {
		s.log.Debug("Message failed to send, removing from send recorder")
		s.recorder.RemoveOnFail(hash, srID)
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// If the message contains a sender, use it instead of the one from the return path.
	if sender, ok := getMessageSender(parser); ok {
		from = sender
		fromAddr, err = s.identityState.GetAddr(from)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to get identity for from address %v", sender)
			return ErrInvalidReturnPath
		}
	}

	if !fromAddr.Send || fromAddr.Status != proton.AddressStatusEnabled {
		s.log.Errorf("Cannot send emails from address: %v", fromAddr.Email)
		return &ErrCannotSendFromAddress{address: fromAddr.Email}
	}

	// Load the user's mail settings.
	settings, err := s.client.GetMailSettings(ctx)
	if err != nil {
		s.log.Debug("Message failed to send, removing from send recorder")
		s.recorder.RemoveOnFail(hash, srID)
		return fmt.Errorf("failed to get mail settings: %w", err)
	}

	if err := usertypes.WithAddrKR(s.identityState.User, fromAddr, s.keyPassProvider.KeyPass(), func(userKR, addrKR *crypto.KeyRing) error {
		// Use the first key for encrypting the message.
		addrKR, err := addrKR.FirstKey()
		if err != nil {
			return fmt.Errorf("failed to get first key: %w", err)
		}

		// Ensure that there is always a text/html or text/plain body part. This is required by the API. If none
		// exists and empty text part will be added.
		parser.AttachEmptyTextPartIfNoneExists()

		// If we have to attach the public key, do it now.
		if settings.AttachPublicKey {
			key, err := addrKR.GetKey(0)
			if err != nil {
				return fmt.Errorf("failed to get sending key: %w", err)
			}

			pubKey, err := key.GetArmoredPublicKey()
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}

			parser.AttachPublicKey(pubKey, fmt.Sprintf(
				"publickey - %v - 0x%v",
				addrKR.GetIdentities()[0].Name,
				strings.ToUpper(key.GetFingerprint()[:8]),
			))
		}

		// Parse the message we want to send (after we have attached the public key).
		message, err := message.ParseWithParser(parser, false)
		if err != nil {
			return fmt.Errorf("failed to parse message: %w", err)
		}

		// Send the message using the correct key.
		sent, err := s.sendWithKey(
			ctx,
			authID,
			s.addressMode,
			settings,
			userKR, addrKR,
			emails, from, to,
			message,
		)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		// If the message was successfully sent, we can update the message ID in the record.
		s.log.Debug("Message sent successfully, signaling recorder")
		s.recorder.SignalMessageSent(hash, srID, sent.ID)

		return nil
	}); err != nil {
		s.log.Debug("Message failed to send, removing from send recorder")
		s.recorder.RemoveOnFail(hash, srID)
		return err
	}

	return nil
}

// sendWithKey sends the message with the given address key.
func (s *Service) sendWithKey(
	ctx context.Context,
	authAddrID string,
	addrMode usertypes.AddressMode,
	settings proton.MailSettings,
	userKR, addrKR *crypto.KeyRing,
	emails []string,
	from string,
	to []string,
	message message.Message,
) (proton.Message, error) {
	references := message.References
	if message.InReplyTo != "" {
		references = append(references, message.InReplyTo)
	}
	parentID, draftsToDelete, err := getParentID(ctx, s.client, authAddrID, addrMode, references)
	if err != nil {
		// Sentry event has been removed; should be replaced with observability - BRIDGE-206.
		s.log.WithError(err).Warn("Failed to get parent ID")
	}

	var decBody string

	// nolint:exhaustive
	switch message.MIMEType {
	case rfc822.TextHTML:
		decBody = string(message.RichBody)

	case rfc822.TextPlain:
		decBody = string(message.PlainBody)

	default:
		return proton.Message{}, fmt.Errorf("unsupported MIME type: %v", message.MIMEType)
	}

	draft, err := s.createDraft(ctx, addrKR, emails, from, to, parentID, message.InReplyTo, message.XForward, proton.DraftTemplate{
		Subject:  message.Subject,
		Body:     decBody,
		MIMEType: message.MIMEType,

		Sender:  message.Sender,
		ToList:  message.ToList,
		CCList:  message.CCList,
		BCCList: message.BCCList,

		ExternalID: message.ExternalID,
	})
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create draft: %w", err)
	}

	attKeys, err := s.createAttachments(ctx, s.client, addrKR, draft.ID, message.Attachments)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create attachments: %w", err)
	}

	recipients, err := s.getRecipients(ctx, s.client, userKR, settings, draft)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to get recipients: %w", err)
	}

	req, err := createSendReq(addrKR, message.MIMEBody, message.RichBody, message.PlainBody, recipients, attKeys)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create packages: %w", err)
	}

	res, err := s.client.SendDraft(ctx, draft.ID, req)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to send draft: %w", err)
	}

	// Only delete the drafts, if any, after message was successfully sent.
	if len(draftsToDelete) != 0 {
		if err := s.client.DeleteMessage(ctx, draftsToDelete...); err != nil {
			s.log.WithField("ids", draftsToDelete).WithError(err).Errorf("Failed to delete requested messages from Drafts")
		}
	}

	return res, nil
}

func getParentID(
	ctx context.Context,
	client *proton.Client,
	authAddrID string,
	addrMode usertypes.AddressMode,
	references []string,
) (string, []string, error) {
	var (
		parentID       string
		internal       []string
		external       []string
		draftsToDelete []string
	)

	// Collect all the internal and external references of the message.
	for _, ref := range references {
		if strings.Contains(ref, message.InternalIDDomain) {
			internal = append(internal, strings.TrimSuffix(ref, "@"+message.InternalIDDomain))
		} else {
			external = append(external, ref)
		}
	}

	// Try to find a parent ID in the internal references.
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

	// If no parent was found, try to find it in the last external reference.
	// There can be multiple messages with the same external ID; in this case, we first look if
	// there is a single one sent by this account (with the `MessageFlagSent` flag set), if yes,
	// then pick that, otherwise don't pick any parent.
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

func (s *Service) createDraft(
	ctx context.Context,
	addrKR *crypto.KeyRing,
	emails []string,
	from string,
	to []string,
	parentID string,
	replyToID string,
	xForwardID string,
	template proton.DraftTemplate,
) (proton.Message, error) {
	// Check sender: set the sender if it's missing.
	if template.Sender == nil {
		template.Sender = &mail.Address{Address: from}
	} else if template.Sender.Address == "" {
		template.Sender.Address = from
	}

	// Check that the sending address is owned by the user, and if so, sanitize it.
	if idx := xslices.IndexFunc(emails, func(email string) bool {
		return strings.EqualFold(email, usertypes.SanitizeEmail(template.Sender.Address))
	}); idx < 0 {
		return proton.Message{}, fmt.Errorf("address %q is not owned by user", template.Sender.Address)
	} else { //nolint:revive
		template.Sender.Address = constructEmail(template.Sender.Address, emails[idx])
	}

	// Check ToList: ensure that ToList only contains addresses we actually plan to send to.
	template.ToList = xslices.Filter(template.ToList, func(addr *mail.Address) bool {
		return slices.Contains(to, addr.Address)
	})

	// Check BCCList: any recipients not present in the ToList or CCList are BCC recipients.
	for _, recipient := range to {
		if !slices.Contains(xslices.Map(xslices.Join(template.ToList, template.CCList, template.BCCList), func(addr *mail.Address) string {
			return addr.Address
		}), recipient) {
			template.BCCList = append(template.BCCList, &mail.Address{Address: recipient})
		}
	}

	var action proton.CreateDraftAction

	if len(replyToID) > 0 {
		// Thunderbird fills both ReplyTo and adds an X-Forwarded-Message-Id header when forwarding.
		if replyToID == xForwardID {
			action = proton.ForwardAction
		} else {
			action = proton.ReplyAction
		}
	} else {
		action = proton.ForwardAction
	}

	return s.client.CreateDraft(ctx, addrKR, proton.CreateDraftReq{
		Message:  template,
		ParentID: parentID,
		Action:   action,
	})
}

func (s *Service) createAttachments(
	ctx context.Context,
	client *proton.Client,
	addrKR *crypto.KeyRing,
	draftID string,
	attachments []message.Attachment,
) (map[string]*crypto.SessionKey, error) {
	type attKey struct {
		attID string
		key   *crypto.SessionKey
	}

	keys, err := parallel.MapContext(ctx, runtime.NumCPU(), attachments, func(ctx context.Context, att message.Attachment) (attKey, error) {
		defer async.HandlePanic(s.panicHandler)

		s.log.WithFields(logrus.Fields{
			"name":        logging.Sensitive(att.Name),
			"contentID":   att.ContentID,
			"disposition": att.Disposition,
			"mime-type":   att.MIMEType,
		}).Debug("Uploading attachment")

		switch att.Disposition {
		case proton.InlineDisposition:
			// Some clients use inline disposition but don't set a content ID. Our API doesn't support this.
			// We could generate our own content ID, but for simplicity, we just set the disposition to attachment.
			if att.ContentID == "" {
				att.Disposition = proton.AttachmentDisposition
			}

		case proton.AttachmentDisposition:
			// Nothing to do.

		default:
			// Some clients leave the content disposition empty or use unsupported values.
			// We default to inline disposition if a content ID is set, and to attachment disposition otherwise.
			if att.ContentID != "" {
				att.Disposition = proton.InlineDisposition
			} else {
				att.Disposition = proton.AttachmentDisposition
			}
		}

		// Exclude name from params since this is already provided using Filename.
		delete(att.MIMEParams, "name")
		delete(att.MIMEParams, "filename")

		attachment, err := client.UploadAttachment(ctx, addrKR, proton.CreateAttachmentReq{
			Filename:    att.Name,
			MessageID:   draftID,
			MIMEType:    rfc822.MIMEType(mime.FormatMediaType(att.MIMEType, att.MIMEParams)),
			Disposition: att.Disposition,
			ContentID:   att.ContentID,
			Body:        att.Data,
		})
		if err != nil {
			return attKey{}, fmt.Errorf("failed to upload attachment: %w", err)
		}

		keyPacket, err := base64.StdEncoding.DecodeString(attachment.KeyPackets)
		if err != nil {
			return attKey{}, fmt.Errorf("failed to decode key packets: %w", err)
		}

		key, err := addrKR.DecryptSessionKey(keyPacket)
		if err != nil {
			return attKey{}, fmt.Errorf("failed to decrypt session key: %w", err)
		}

		return attKey{attID: attachment.ID, key: key}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create attachments: %w", err)
	}

	attKeys := make(map[string]*crypto.SessionKey)

	for _, key := range keys {
		attKeys[key.attID] = key.key
	}

	return attKeys, nil
}

func (s *Service) getRecipients(
	ctx context.Context,
	client *proton.Client,
	userKR *crypto.KeyRing,
	settings proton.MailSettings,
	draft proton.Message,
) (recipients, error) {
	addresses := xslices.Map(xslices.Join(draft.ToList, draft.CCList, draft.BCCList), func(addr *mail.Address) string {
		return addr.Address
	})

	prefs, err := parallel.MapContext(ctx, runtime.NumCPU(), addresses, func(ctx context.Context, recipient string) (proton.SendPreferences, error) {
		defer async.HandlePanic(s.panicHandler)

		pubKeys, recType, err := client.GetPublicKeys(ctx, recipient)
		if err != nil {
			return proton.SendPreferences{}, fmt.Errorf("failed to get public key for %v: %w", recipient, err)
		}

		contactSettings, err := getContactSettings(ctx, client, userKR, recipient)
		if err != nil {
			return proton.SendPreferences{}, fmt.Errorf("failed to get contact settings for %v: %w", recipient, err)
		}

		return buildSendPrefs(contactSettings, settings, pubKeys, draft.MIMEType, recType == proton.RecipientTypeInternal)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get send preferences: %w", err)
	}

	recipients := make(recipients)

	for idx, pref := range prefs {
		recipients[addresses[idx]] = pref
	}

	return recipients, nil
}

func getContactSettings(
	ctx context.Context,
	client *proton.Client,
	userKR *crypto.KeyRing,
	recipient string,
) (proton.ContactSettings, error) {
	contacts, err := client.GetAllContactEmails(ctx, recipient)
	if err != nil {
		return proton.ContactSettings{}, fmt.Errorf("failed to get contact data: %w", err)
	}

	idx := xslices.IndexFunc(contacts, func(contact proton.ContactEmail) bool {
		return contact.Email == recipient
	})

	if idx < 0 {
		return proton.ContactSettings{}, nil
	}

	contact, err := client.GetContact(ctx, contacts[idx].ContactID)
	if err != nil {
		return proton.ContactSettings{}, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact.GetSettings(userKR, recipient, proton.CardTypeSigned)
}

func getMessageSender(parser *parser.Parser) (string, bool) {
	address, err := rfc5322.ParseAddressList(parser.Root().Header.Get("From"))
	if err != nil {
		return "", false
	} else if len(address) == 0 {
		return "", false
	}

	return address[0].Address, true
}

func constructEmail(headerEmail string, addressEmail string) string {
	splitAtHeader := strings.Split(headerEmail, "@")
	if len(splitAtHeader) != 2 {
		return addressEmail
	}

	splitPlus := strings.Split(splitAtHeader[0], "+")
	if len(splitPlus) != 2 {
		return addressEmail
	}

	splitAtAddress := strings.Split(addressEmail, "@")
	if len(splitAtAddress) != 2 {
		return addressEmail
	}

	return splitAtAddress[0] + "+" + splitPlus[1] + "@" + splitAtAddress[1]
}
