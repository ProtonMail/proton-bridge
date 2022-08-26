package user

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

type smtpSession struct {
	client    *liteapi.Client
	username  string
	addresses []liteapi.Address
	userKR    *crypto.KeyRing
	addrKRs   map[string]*crypto.KeyRing
	settings  liteapi.MailSettings

	from string
	to   map[string]struct{}
}

func newSMTPSession(
	client *liteapi.Client,
	username string,
	addresses []liteapi.Address,
	userKR *crypto.KeyRing,
	addrKRs map[string]*crypto.KeyRing,
	settings liteapi.MailSettings,
) *smtpSession {
	return &smtpSession{
		client:    client,
		username:  username,
		addresses: addresses,
		userKR:    userKR,
		addrKRs:   addrKRs,
		settings:  settings,

		from: "",
		to:   make(map[string]struct{}),
	}
}

// Discard currently processed message.
func (session *smtpSession) Reset() {
	logrus.Info("SMTP session reset")

	// Clear the from and to fields.
	session.from = ""
	session.to = make(map[string]struct{})
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

	if opts.RequireTLS {
		return ErrNotImplemented
	}

	if opts.UTF8 {
		return ErrNotImplemented
	}

	if opts.Auth != nil && *opts.Auth != "" && *opts.Auth != session.username {
		return ErrNotImplemented
	}

	idx := xslices.IndexFunc(session.addresses, func(address liteapi.Address) bool {
		return strings.EqualFold(address.Email, from)
	})

	if idx < 0 {
		return ErrInvalidReturnPath
	}

	session.from = session.addresses[idx].ID

	return nil
}

// Add recipient for currently processed message.
func (session *smtpSession) Rcpt(to string) error {
	logrus.Info("SMTP session rcpt")

	if to == "" {
		return ErrInvalidRecipient
	}

	session.to[to] = struct{}{}

	return nil
}

// Set currently processed message contents and send it.
func (session *smtpSession) Data(r io.Reader) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.Info("SMTP session data")

	if session.from == "" {
		return ErrInvalidReturnPath
	}

	if len(session.to) == 0 {
		return ErrInvalidRecipient
	}

	addrKR, ok := session.addrKRs[session.from]
	if !ok {
		return ErrMissingAddressKey
	}

	addrKR, err := addrKR.FirstKey()
	if err != nil {
		return fmt.Errorf("failed to get first key: %w", err)
	}

	parser, err := parser.New(r)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	if session.settings.AttachPublicKey == liteapi.AttachPublicKeyEnabled {
		key, err := addrKR.GetKey(0)
		if err != nil {
			return fmt.Errorf("failed to get user public key: %w", err)
		}

		pubKey, err := key.GetArmoredPublicKey()
		if err != nil {
			return fmt.Errorf("failed to get user public key: %w", err)
		}

		parser.AttachPublicKey(pubKey, fmt.Sprintf("publickey - %v - %v", addrKR.GetIdentities()[0].Name, key.GetFingerprint()[:8]))
	}

	message, err := message.ParseWithParser(parser)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	draft, attKeys, err := session.createDraft(ctx, addrKR, message)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	recipients, err := session.getRecipients(ctx, message.Recipients(), message.MIMEType)
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	req, err := createSendReq(addrKR, message.MIMEBody, message.RichBody, message.PlainBody, recipients, attKeys)
	if err != nil {
		return fmt.Errorf("failed to create packages: %w", err)
	}

	res, err := session.client.SendDraft(ctx, draft.ID, req)
	if err != nil {
		return fmt.Errorf("failed to send draft: %w", err)
	}

	logrus.WithField("messageID", res.ID).Info("SMTP message sent")

	return nil
}

func (session *smtpSession) createDraft(ctx context.Context, addrKR *crypto.KeyRing, message message.Message) (liteapi.Message, map[string]*crypto.SessionKey, error) {
	encBody, err := addrKR.Encrypt(crypto.NewPlainMessageFromString(string(message.RichBody)), nil)
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to encrypt message body: %w", err)
	}

	armBody, err := encBody.GetArmored()
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to armor message body: %w", err)
	}

	draft, err := session.client.CreateDraft(ctx, liteapi.CreateDraftReq{
		Message: liteapi.DraftTemplate{
			Subject: message.Subject,
			Sender:  message.Sender,
			ToList:  message.ToList,
			CCList:  message.CCList,
			BCCList: message.BCCList,
			Body:    armBody,
		},
		AttachmentKeyPackets: []string{},
	})
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to create draft: %w", err)
	}

	attKeys, err := session.createAttachments(ctx, addrKR, draft.ID, message.Attachments)
	if err != nil {
		return liteapi.Message{}, nil, fmt.Errorf("failed to create attachments: %w", err)
	}

	return draft, attKeys, nil
}

func (session *smtpSession) createAttachments(ctx context.Context, addrKR *crypto.KeyRing, draftID string, attachments []message.Attachment) (map[string]*crypto.SessionKey, error) {
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

		attachment, err := session.client.UploadAttachment(ctx, liteapi.CreateAttachmentReq{
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

func (session *smtpSession) getRecipients(ctx context.Context, addresses []string, mimeType rfc822.MIMEType) (recipients, error) {
	prefs, err := parallel.MapContext(ctx, runtime.NumCPU(), addresses, func(ctx context.Context, address string) (liteapi.SendPreferences, error) {
		return session.getSendPrefs(ctx, address, mimeType)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}

	recipients := make(recipients)

	for idx, pref := range prefs {
		recipients[addresses[idx]] = pref
	}

	return recipients, nil
}

func (session *smtpSession) getSendPrefs(ctx context.Context, recipient string, mimeType rfc822.MIMEType) (liteapi.SendPreferences, error) {
	pubKeys, internal, err := session.client.GetPublicKeys(ctx, recipient)
	if err != nil {
		return liteapi.SendPreferences{}, fmt.Errorf("failed to get public keys: %w", err)
	}

	settings, err := session.getContactSettings(ctx, recipient)
	if err != nil {
		return liteapi.SendPreferences{}, fmt.Errorf("failed to get contact settings: %w", err)
	}

	return buildSendPrefs(settings, session.settings, pubKeys, mimeType, internal)
}

func (session *smtpSession) getContactSettings(ctx context.Context, recipient string) (liteapi.ContactSettings, error) {
	contacts, err := session.client.GetAllContactEmails(ctx, recipient)
	if err != nil {
		return liteapi.ContactSettings{}, fmt.Errorf("failed to get contact data: %w", err)
	}

	idx := xslices.IndexFunc(contacts, func(contact liteapi.ContactEmail) bool {
		return contact.Email == recipient
	})

	if idx < 0 {
		return liteapi.ContactSettings{}, nil
	}

	contact, err := session.client.GetContact(ctx, contacts[idx].ContactID)
	if err != nil {
		return liteapi.ContactSettings{}, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact.GetSettings(session.userKR, recipient)
}
