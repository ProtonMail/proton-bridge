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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/url"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

// sendWithKey sends the message with the given address key.
func sendWithKey( //nolint:funlen
	ctx context.Context,
	client *liteapi.Client,
	authAddrID string,
	addrMode vault.AddressMode,
	settings liteapi.MailSettings,
	userKR *crypto.KeyRing,
	addrKR *crypto.KeyRing,
	emails []string,
	from string,
	to []string,
	message message.Message,
) (liteapi.Message, error) {
	parentID, err := getParentID(ctx, client, authAddrID, addrMode, message.References)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to get parent ID: %w", err)
	}

	var decBody string

	switch message.MIMEType { //nolint:exhaustive
	case rfc822.TextHTML:
		decBody = string(message.RichBody)

	case rfc822.TextPlain:
		decBody = string(message.PlainBody)

	default:
		return liteapi.Message{}, fmt.Errorf("unsupported MIME type: %v", message.MIMEType)
	}

	encBody, err := addrKR.Encrypt(crypto.NewPlainMessageFromString(decBody), nil)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to encrypt message body: %w", err)
	}

	armBody, err := encBody.GetArmored()
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to get armored message body: %w", err)
	}

	draft, err := createDraft(ctx, client, emails, from, to, parentID, message.InReplyTo, liteapi.DraftTemplate{
		Subject:  message.Subject,
		Body:     armBody,
		MIMEType: message.MIMEType,

		Sender:  message.Sender,
		ToList:  message.ToList,
		CCList:  message.CCList,
		BCCList: message.BCCList,

		ExternalID: message.ExternalID,
	})
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to create attachments: %w", err)
	}

	attKeys, err := createAttachments(ctx, client, addrKR, draft.ID, message.Attachments)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to create attachments: %w", err)
	}

	recipients, err := getRecipients(ctx, client, userKR, settings, draft)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to get recipients: %w", err)
	}

	req, err := createSendReq(addrKR, message.MIMEBody, message.RichBody, message.PlainBody, recipients, attKeys)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to create packages: %w", err)
	}

	res, err := client.SendDraft(ctx, draft.ID, req)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to send draft: %w", err)
	}

	return res, nil
}

func getParentID( //nolint:funlen
	ctx context.Context,
	client *liteapi.Client,
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
		filter := url.Values{
			"ID": {internal},
		}

		if addrMode == vault.SplitMode {
			filter["AddressID"] = []string{authAddrID}
		}

		metadata, err := client.GetMessageMetadata(ctx, filter)
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
		filter := url.Values{
			"ExternalID": {external[len(external)-1]},
		}

		if addrMode == vault.SplitMode {
			filter["AddressID"] = []string{authAddrID}
		}

		metadata, err := client.GetMessageMetadata(ctx, filter)
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
	client *liteapi.Client,
	emails []string,
	from string,
	to []string,
	parentID string,
	replyToID string,
	template liteapi.DraftTemplate,
) (liteapi.Message, error) {
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
		return liteapi.Message{}, fmt.Errorf("address %q is not owned by user", template.Sender.Address)
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

	var action liteapi.CreateDraftAction

	if len(replyToID) > 0 {
		action = liteapi.ReplyAction
	} else {
		action = liteapi.ForwardAction
	}

	return client.CreateDraft(ctx, liteapi.CreateDraftReq{
		Message:  template,
		ParentID: parentID,
		Action:   action,
	})
}

func createAttachments(
	ctx context.Context,
	client *liteapi.Client,
	addrKR *crypto.KeyRing,
	draftID string,
	attachments []message.Attachment,
) (map[string]*crypto.SessionKey, error) {
	type attKey struct {
		attID string
		key   *crypto.SessionKey
	}

	keys, err := parallel.MapContext(ctx, runtime.NumCPU(), attachments, func(ctx context.Context, att message.Attachment) (attKey, error) {
		sig, err := addrKR.SignDetached(crypto.NewPlainMessage(att.Data))
		if err != nil {
			return attKey{}, fmt.Errorf("failed to sign attachment: %w", err)
		}

		encData, err := addrKR.EncryptAttachment(crypto.NewPlainMessage(att.Data), att.Name)
		if err != nil {
			return attKey{}, fmt.Errorf("failed to encrypt attachment: %w", err)
		}

		attachment, err := client.UploadAttachment(ctx, liteapi.CreateAttachmentReq{
			Filename:    att.Name,
			MessageID:   draftID,
			MIMEType:    rfc822.MIMEType(att.MIMEType),
			Disposition: liteapi.Disposition(att.Disposition),
			ContentID:   att.ContentID,
			KeyPackets:  encData.KeyPacket,
			DataPacket:  encData.DataPacket,
			Signature:   sig.GetBinary(),
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
	client *liteapi.Client,
	userKR *crypto.KeyRing,
	settings liteapi.MailSettings,
	draft liteapi.Message,
) (recipients, error) {
	addresses := xslices.Map(xslices.Join(draft.ToList, draft.CCList, draft.BCCList), func(addr *mail.Address) string {
		return addr.Address
	})

	prefs, err := parallel.MapContext(ctx, runtime.NumCPU(), addresses, func(ctx context.Context, recipient string) (liteapi.SendPreferences, error) {
		pubKeys, recType, err := client.GetPublicKeys(ctx, recipient)
		if err != nil {
			return liteapi.SendPreferences{}, fmt.Errorf("failed to get public keys: %w", err)
		}

		contactSettings, err := getContactSettings(ctx, client, userKR, recipient)
		if err != nil {
			return liteapi.SendPreferences{}, fmt.Errorf("failed to get contact settings: %w", err)
		}

		return buildSendPrefs(contactSettings, settings, pubKeys, draft.MIMEType, recType == liteapi.RecipientTypeInternal)
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
	client *liteapi.Client,
	userKR *crypto.KeyRing,
	recipient string,
) (liteapi.ContactSettings, error) {
	contacts, err := client.GetAllContactEmails(ctx, recipient)
	if err != nil {
		return liteapi.ContactSettings{}, fmt.Errorf("failed to get contact data: %w", err)
	}

	idx := xslices.IndexFunc(contacts, func(contact liteapi.ContactEmail) bool {
		return contact.Email == recipient
	})

	if idx < 0 {
		return liteapi.ContactSettings{}, nil
	}

	contact, err := client.GetContact(ctx, contacts[idx].ContactID)
	if err != nil {
		return liteapi.ContactSettings{}, fmt.Errorf("failed to get contact: %w", err)
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
