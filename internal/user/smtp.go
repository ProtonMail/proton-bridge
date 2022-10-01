package user

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/slices"
)

type smtpSession struct {
	*User

	// authID holds the ID of the address that the SMTP client authenticated with to send the message.
	authID string

	// fromAddrID is the ID of the current sending address (taken from the return path).
	fromAddrID string

	// to holds all to for the current message.
	to []string
}

func newSMTPSession(user *User, email string) (*smtpSession, error) {
	authID, err := safe.GetSliceErr(user.apiAddrs, func(apiAddrs []liteapi.Address) (string, error) {
		return getAddrID(apiAddrs, email)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get address ID: %w", err)
	}

	return &smtpSession{
		User:   user,
		authID: authID,
	}, nil
}

// Discard currently processed message.
func (session *smtpSession) Reset() {
	logrus.Info("SMTP session reset")

	// Clear the from and to fields.
	session.fromAddrID = ""
	session.to = nil
}

// Free all resources associated with session.
func (session *smtpSession) Logout() error {
	defer session.Reset()

	logrus.Info("SMTP session logout")

	return nil
}

// Set return path for currently processed message.
func (session *smtpSession) Mail(from string, opts smtp.MailOptions) error {
	logrus.Info("SMTP session mail")

	return session.apiAddrs.GetErr(func(apiAddrs []liteapi.Address) error {

		switch {
		case opts.RequireTLS:
			return ErrNotImplemented

		case opts.UTF8:
			return ErrNotImplemented

		case opts.Auth != nil:
			email, err := getAddrEmail(apiAddrs, session.authID)
			if err != nil {
				return fmt.Errorf("invalid auth address: %w", err)
			}

			if *opts.Auth != "" && *opts.Auth != email {
				return ErrNotImplemented
			}
		}

		fromAddrID, err := getAddrID(apiAddrs, from)
		if err != nil {
			return fmt.Errorf("invalid return path: %w", err)
		}

		session.fromAddrID = fromAddrID

		return nil
	})
}

// Add recipient for currently processed message.
func (session *smtpSession) Rcpt(to string) error {
	logrus.Info("SMTP session rcpt")

	if to == "" {
		return ErrInvalidRecipient
	}

	if !slices.Contains(session.to, to) {
		session.to = append(session.to, to)
	}

	return nil
}

// Set currently processed message contents and send it.
func (session *smtpSession) Data(r io.Reader) error {
	logrus.Info("SMTP session data")

	switch {
	case session.fromAddrID == "":
		return ErrInvalidReturnPath

	case len(session.to) == 0:
		return ErrInvalidRecipient
	}

	parser, err := parser.New(r)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// If the message contains a sender, use it instead of the one from the return path.
	session.apiAddrs.Get(func(apiAddrs []liteapi.Address) {
		if sender, ok := getMessageSender(parser); ok {
			for _, addr := range apiAddrs {
				if strings.EqualFold(addr.Email, sanitizeEmail(sender)) {
					session.fromAddrID = addr.ID
				}
			}
		}
	})

	addrKR, ok := session.addrKRs[session.fromAddrID]
	if !ok {
		return ErrMissingAddrKey
	}

	firstAddrKR, err := addrKR.FirstKey()
	if err != nil {
		return fmt.Errorf("failed to get first key: %w", err)
	}

	from, err := safe.GetSliceErr(session.apiAddrs, func(apiAddrs []liteapi.Address) (string, error) {
		email, err := getAddrEmail(apiAddrs, session.fromAddrID)
		if err != nil {
			return "", fmt.Errorf("failed to get address email: %w", err)
		}

		return sanitizeEmail(email), nil
	})
	if err != nil {
		return fmt.Errorf("failed to get address email: %w", err)
	}

	message, err := safe.GetSliceErr(session.apiAddrs, func(apiAddrs []liteapi.Address) (liteapi.Message, error) {
		return safe.GetTypeErr(session.settings, func(settings liteapi.MailSettings) (liteapi.Message, error) {
			return sendWithKey(
				session.client,
				session.authID,
				session.vault.AddressMode(),
				apiAddrs,
				settings,
				session.userKR,
				firstAddrKR,
				parser,
				from,
				session.to,
			)
		})
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logrus.WithField("messageID", message.ID).Info("Message sent")

	return nil
}

// sendWithKey sends the message with the given address key.
func sendWithKey(
	client *liteapi.Client,
	authAddrID string,
	addrMode vault.AddressMode,
	apiAddrs []liteapi.Address,
	settings liteapi.MailSettings,
	userKR, addrKR *crypto.KeyRing,
	parser *parser.Parser,
	from string,
	to []string,
) (liteapi.Message, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if settings.AttachPublicKey == liteapi.AttachPublicKeyEnabled {
		key, err := addrKR.GetKey(0)
		if err != nil {
			return liteapi.Message{}, fmt.Errorf("failed to get user public key: %w", err)
		}

		pubKey, err := key.GetArmoredPublicKey()
		if err != nil {
			return liteapi.Message{}, fmt.Errorf("failed to get user public key: %w", err)
		}

		parser.AttachPublicKey(pubKey, fmt.Sprintf("publickey - %v - %v", addrKR.GetIdentities()[0].Name, key.GetFingerprint()[:8]))
	}

	message, err := message.ParseWithParser(parser)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to parse message: %w", err)
	}

	if err := sanitizeParsedMessage(&message, apiAddrs, from, to); err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to sanitize message: %w", err)
	}

	parentID, err := getParentID(ctx, client, authAddrID, addrMode, message.References)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to get parent ID: %w", err)
	}

	draft, attKeys, err := createDraftWithAttachments(ctx, client, addrKR, message, parentID)
	if err != nil {
		return liteapi.Message{}, fmt.Errorf("failed to create draft: %w", err)
	}

	recipients, err := getRecipients(ctx, client, userKR, settings, message.Recipients(), message.MIMEType)
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

func sanitizeParsedMessage(message *message.Message, apiAddrs []liteapi.Address, from string, to []string) error {
	// Check sender: set the sender in the parsed message if it's missing.
	if message.Sender == nil {
		message.Sender = &mail.Address{Address: from}
	} else if message.Sender.Address == "" {
		message.Sender.Address = from
	}

	// Check that the sending address is owned by the user, and if so, properly capitalize it.
	if idx := xslices.IndexFunc(apiAddrs, func(addr liteapi.Address) bool {
		return strings.EqualFold(addr.Email, sanitizeEmail(message.Sender.Address))
	}); idx < 0 {
		return fmt.Errorf("address %q is not owned by user", message.Sender.Address)
	} else {
		message.Sender.Address = constructEmail(message.Sender.Address, apiAddrs[idx].Email)
	}

	// Check ToList: ensure that ToList only contains addresses we actually plan to send to.
	message.ToList = xslices.Filter(message.ToList, func(addr *mail.Address) bool {
		return slices.Contains(to, addr.Address)
	})

	// Check BCCList: any recipients not present in the ToList or CCList are BCC recipients.
	for _, recipient := range to {
		if !slices.Contains(message.Recipients(), recipient) {
			message.BCCList = append(message.BCCList, &mail.Address{Address: recipient})
		}
	}

	return nil
}

func getParentID(
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

		metadata, err := client.GetAllMessageMetadata(ctx, filter)
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

		metadata, err := client.GetAllMessageMetadata(ctx, filter)
		if err != nil {
			return "", fmt.Errorf("failed to get message metadata: %w", err)
		}

		if len(metadata) == 1 {
			parentID = metadata[0].ID
		}
	}

	return parentID, nil
}

func createDraftWithAttachments(
	ctx context.Context,
	client *liteapi.Client,
	addrKR *crypto.KeyRing,
	message message.Message,
	parentID string,
) (liteapi.Message, map[string]*crypto.SessionKey, error) {
	encBody, err := addrKR.Encrypt(crypto.NewPlainMessageFromString(string(message.RichBody)), nil)
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to encrypt message body: %w", err)
	}

	armBody, err := encBody.GetArmored()
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to armor message body: %w", err)
	}

	draft, err := client.CreateDraft(ctx, liteapi.CreateDraftReq{
		Message: liteapi.DraftTemplate{
			Subject:  message.Subject,
			Sender:   message.Sender,
			ToList:   message.ToList,
			CCList:   message.CCList,
			BCCList:  message.BCCList,
			Body:     armBody,
			MIMEType: message.MIMEType,

			ExternalID: message.ExternalID,
		},

		ParentID: parentID,
	})
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to create draft: %w", err)
	}

	attKeys, err := createAttachments(ctx, client, addrKR, draft.ID, message.Attachments)
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to create attachments: %w", err)
	}

	return draft, attKeys, nil
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
	addresses []string,
	mimeType rfc822.MIMEType,
) (recipients, error) {
	prefs, err := parallel.MapContext(ctx, runtime.NumCPU(), addresses, func(ctx context.Context, address string) (liteapi.SendPreferences, error) {
		return getSendPrefs(ctx, client, userKR, settings, address, mimeType)
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

func getSendPrefs(
	ctx context.Context,
	client *liteapi.Client,
	userKR *crypto.KeyRing,
	settings liteapi.MailSettings,
	recipient string,
	mimeType rfc822.MIMEType,
) (liteapi.SendPreferences, error) {
	pubKeys, recType, err := client.GetPublicKeys(ctx, recipient)
	if err != nil {
		return liteapi.SendPreferences{}, fmt.Errorf("failed to get public keys: %w", err)
	}

	contactSettings, err := getContactSettings(ctx, client, userKR, recipient)
	if err != nil {
		return liteapi.SendPreferences{}, fmt.Errorf("failed to get contact settings: %w", err)
	}

	return buildSendPrefs(contactSettings, settings, pubKeys, mimeType, recType == liteapi.RecipientTypeInternal)
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
