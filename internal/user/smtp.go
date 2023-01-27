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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// sendMail sends an email from the given address to the given recipients.
//
// nolint:funlen
func (user *User) sendMail(authID string, from string, to []string, r io.Reader) error {
	return safe.RLockRet(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if _, err := getAddrID(user.apiAddrs, from); err != nil {
			return ErrInvalidReturnPath
		}

		emails := xslices.Map(maps.Values(user.apiAddrs), func(addr proton.Address) string {
			return addr.Email
		})

		// Read the message to send.
		b, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		// If running a QA build, dump to disk.
		debugDumpToDisk(b)

		// Compute the hash of the message (to match it against SMTP messages).
		hash, err := getMessageHash(b)
		if err != nil {
			return err
		}

		// Check if we already tried to send this message recently.
		if ok, err := user.sendHash.tryInsertWait(ctx, hash, to, time.Now().Add(90*time.Second)); err != nil {
			return fmt.Errorf("failed to check send hash: %w", err)
		} else if !ok {
			user.log.Warn("A duplicate message was already sent recently, skipping")
			return nil
		}

		// If we fail to send this message, we should remove the hash from the send recorder.
		defer user.sendHash.removeOnFail(hash)

		// Create a new message parser from the reader.
		parser, err := parser.New(bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("failed to create parser: %w", err)
		}

		// If the message contains a sender, use it instead of the one from the return path.
		if sender, ok := getMessageSender(parser); ok {
			from = sender
		}

		// Load the user's mail settings.
		settings, err := user.client.GetMailSettings(ctx)
		if err != nil {
			return fmt.Errorf("failed to get mail settings: %w", err)
		}

		addrID, err := getAddrID(user.apiAddrs, from)
		if err != nil {
			return err
		}

		return withAddrKR(user.apiUser, user.apiAddrs[addrID], user.vault.KeyPass(), func(userKR, addrKR *crypto.KeyRing) error {
			// Use the first key for encrypting the message.
			addrKR, err := addrKR.FirstKey()
			if err != nil {
				return fmt.Errorf("failed to get first key: %w", err)
			}

			// If we have to attach the public key, do it now.
			if settings.AttachPublicKey == proton.AttachPublicKeyEnabled {
				key, err := addrKR.GetKey(0)
				if err != nil {
					return fmt.Errorf("failed to get sending key: %w", err)
				}

				pubKey, err := key.GetArmoredPublicKey()
				if err != nil {
					return fmt.Errorf("failed to get public key: %w", err)
				}

				parser.AttachPublicKey(pubKey, fmt.Sprintf("publickey - %v - %v", addrKR.GetIdentities()[0].Name, key.GetFingerprint()[:8]))
			}

			// Parse the message we want to send (after we have attached the public key).
			message, err := message.ParseWithParser(parser)
			if err != nil {
				return fmt.Errorf("failed to parse message: %w", err)
			}

			// Send the message using the correct key.
			sent, err := sendWithKey(
				ctx,
				user.client,
				authID,
				user.vault.AddressMode(),
				settings,
				userKR, addrKR,
				emails, from, to,
				message,
			)
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}

			// If the message was successfully sent, we can update the message ID in the record.
			user.sendHash.addMessageID(hash, sent.ID)

			return nil
		})
	}, user.apiUserLock, user.apiAddrsLock)
}

// sendWithKey sends the message with the given address key.
func sendWithKey( //nolint:funlen
	ctx context.Context,
	client *proton.Client,
	authAddrID string,
	addrMode vault.AddressMode,
	settings proton.MailSettings,
	userKR, addrKR *crypto.KeyRing,
	emails []string,
	from string,
	to []string,
	message message.Message,
) (proton.Message, error) {
	parentID, err := getParentID(ctx, client, authAddrID, addrMode, message.References)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to get parent ID: %w", err)
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

	draft, err := createDraft(ctx, client, addrKR, emails, from, to, parentID, message.InReplyTo, proton.DraftTemplate{
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
		return proton.Message{}, fmt.Errorf("failed to create attachments: %w", err)
	}

	attKeys, err := createAttachments(ctx, client, addrKR, draft.ID, message.Attachments)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create attachments: %w", err)
	}

	recipients, err := getRecipients(ctx, client, userKR, settings, draft)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to get recipients: %w", err)
	}

	req, err := createSendReq(addrKR, message.MIMEBody, message.RichBody, message.PlainBody, recipients, attKeys)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to create packages: %w", err)
	}

	res, err := client.SendDraft(ctx, draft.ID, req)
	if err != nil {
		return proton.Message{}, fmt.Errorf("failed to send draft: %w", err)
	}

	return res, nil
}

func getParentID( //nolint:funlen
	ctx context.Context,
	client *proton.Client,
	authAddrID string,
	addrMode vault.AddressMode,
	references []string,
) (string, error) {
	var (
		parentID string
		internal []string
		external []string
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

		if addrMode == vault.SplitMode {
			addrID = authAddrID
		}

		metadata, err := client.GetMessageMetadata(ctx, proton.MessageFilter{
			ID:        []string{internal},
			AddressID: addrID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get message metadata: %w", err)
		}

		for _, metadata := range metadata {
			if !metadata.IsDraft() {
				parentID = metadata.ID
			} else if err := client.DeleteMessage(ctx, metadata.ID); err != nil {
				return "", fmt.Errorf("failed to delete message: %w", err)
			}
		}
	}

	// If no parent was found, try to find it in the last external reference.
	// There can be multiple messages with the same external ID; in this case, we don't pick any parent.
	if parentID == "" && len(external) > 0 {
		var addrID string

		if addrMode == vault.SplitMode {
			addrID = authAddrID
		}

		metadata, err := client.GetMessageMetadata(ctx, proton.MessageFilter{
			ExternalID: external[len(external)-1],
			AddressID:  addrID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get message metadata: %w", err)
		}

		if len(metadata) == 1 {
			parentID = metadata[0].ID
		}
	}

	return parentID, nil
}

func createDraft(
	ctx context.Context,
	client *proton.Client,
	addrKR *crypto.KeyRing,
	emails []string,
	from string,
	to []string,
	parentID string,
	replyToID string,
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
		return strings.EqualFold(email, sanitizeEmail(template.Sender.Address))
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
		action = proton.ReplyAction
	} else {
		action = proton.ForwardAction
	}

	return client.CreateDraft(ctx, addrKR, proton.CreateDraftReq{
		Message:  template,
		ParentID: parentID,
		Action:   action,
	})
}

// nolint:funlen
func createAttachments(
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
		logrus.WithFields(logrus.Fields{
			"name":        logging.Sensitive(att.Name),
			"contentID":   att.ContentID,
			"disposition": att.Disposition,
			"mime-type":   att.MIMEType,
		}).Debug("Uploading attachment")

		// Some client might have leave empty the content disposition or use unsupported values.
		if att.Disposition != string(proton.InlineDisposition) && att.Disposition != string(proton.AttachmentDisposition) {
			att.Disposition = string(proton.AttachmentDisposition)
		}

		// Some clients use inline disposition but don't set a content ID. Our API doesn't support this.
		// We could generate our own content ID, but for simplicity, we just set the disposition to attachment.
		if att.Disposition == string(proton.InlineDisposition) && att.ContentID == "" {
			att.Disposition = string(proton.AttachmentDisposition)
		}

		attachment, err := client.UploadAttachment(ctx, addrKR, proton.CreateAttachmentReq{
			Filename:    att.Name,
			MessageID:   draftID,
			MIMEType:    rfc822.MIMEType(att.MIMEType),
			Disposition: proton.Disposition(att.Disposition),
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

func getRecipients(
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
		pubKeys, recType, err := client.GetPublicKeys(ctx, recipient)
		if err != nil {
			return proton.SendPreferences{}, fmt.Errorf("failed to get public keys: %w", err)
		}

		contactSettings, err := getContactSettings(ctx, client, userKR, recipient)
		if err != nil {
			return proton.SendPreferences{}, fmt.Errorf("failed to get contact settings: %w", err)
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

	return contact.GetSettings(userKR, recipient)
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

func sanitizeEmail(email string) string {
	splitAt := strings.Split(email, "@")
	if len(splitAt) != 2 {
		return email
	}

	return strings.Split(splitAt[0], "+")[0] + "@" + splitAt[1]
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
