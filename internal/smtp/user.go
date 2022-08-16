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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// NOTE: Comments in this file refer to a specification in a document called
// "Proton Mail Encryption logic". It will be referred to via abbreviation PMEL.

package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	pkgMsg "github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	goSMTPBackend "github.com/emersion/go-smtp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type smtpUser struct {
	panicHandler  panicHandler
	eventListener listener.Listener
	backend       *smtpBackend
	user          bridgeUser
	storeUser     storeUserProvider
	username      string
	addressID     string

	returnPath string
	to         []string
}

// newSMTPUser returns struct implementing go-smtp/session interface.
func newSMTPUser(
	panicHandler panicHandler,
	eventListener listener.Listener,
	smtpBackend *smtpBackend,
	user bridgeUser,
	username string,
	addressID string,
) (goSMTPBackend.Session, error) {
	storeUser := user.GetStore()
	if storeUser == nil {
		return nil, errors.New("user database is not initialized")
	}

	return &smtpUser{
		panicHandler:  panicHandler,
		eventListener: eventListener,
		backend:       smtpBackend,
		user:          user,
		storeUser:     storeUser,
		username:      username,
		addressID:     addressID,
	}, nil
}

// This method should eventually no longer be necessary. Everything should go via store.
func (su *smtpUser) client() pmapi.Client {
	return su.user.GetClient()
}

// Send sends an email from the given address to the given addresses with the given body.
func (su *smtpUser) getSendPreferences(
	recipient, messageMIMEType string,
	mailSettings pmapi.MailSettings,
) (preferences SendPreferences, err error) {
	b := &sendPreferencesBuilder{}

	// 1. contact vcard data
	vCardData, err := su.getContactVCardData(recipient)
	if err != nil {
		return
	}

	// 2. api key data
	apiKeys, isInternal, err := su.getAPIKeyData(recipient)
	if err != nil {
		return
	}

	// 1 + 2 -> 3. advanced PGP settings
	if err = b.setPGPSettings(vCardData, apiKeys, isInternal); err != nil {
		return
	}

	// 4. mail settings
	// Passed in from su.client().GetMailSettings()

	// 3 + 4 -> 5. encryption preferences
	b.setEncryptionPreferences(mailSettings)

	// 6. composer preferences -- in our case, this comes from the MIME type of the message.

	// 5 + 6 -> 7. send preferences
	b.setMIMEPreferences(messageMIMEType)

	return b.build(), nil
}

func (su *smtpUser) getContactVCardData(recipient string) (meta *ContactMetadata, err error) {
	emails, err := su.client().GetContactEmailByEmail(context.TODO(), recipient, 0, 1000)
	if err != nil {
		return
	}

	for _, email := range emails {
		if email.Defaults == 1 {
			// NOTE: Can we still ignore this?
			continue
		}

		var contact pmapi.Contact
		if contact, err = su.client().GetContactByID(context.TODO(), email.ContactID); err != nil {
			return
		}

		var cards []pmapi.Card
		if cards, err = su.client().DecryptAndVerifyCards(contact.Cards); err != nil {
			return
		}

		return GetContactMetadataFromVCards(cards, recipient)
	}

	return
}

func (su *smtpUser) getAPIKeyData(recipient string) (apiKeys []pmapi.PublicKey, isInternal bool, err error) {
	return su.client().GetPublicKeysForEmail(context.TODO(), recipient)
}

// Discard currently processed message.
func (su *smtpUser) Reset() {
	log.Trace("Resetting the session")
	su.returnPath = ""
	su.to = []string{}
}

// Set return path for currently processed message.
func (su *smtpUser) Mail(returnPath string, opts goSMTPBackend.MailOptions) error {
	log.WithField("returnPath", returnPath).WithField("opts", opts).Trace("Setting mail from")

	// REQUIRETLS and SMTPUTF8 have to be announced to be used by client.
	// Bridge does not use those extensions so this should not happen.
	if opts.RequireTLS {
		return errors.New("REQUIRETLS extension is not supported")
	}
	if opts.UTF8 {
		return errors.New("SMTPUTF8 extension is not supported")
	}

	if opts.Auth != nil && *opts.Auth != "" && *opts.Auth != su.username {
		return errors.New("changing identity is not supported")
	}

	if returnPath != "" {
		addr := su.client().Addresses().ByEmail(returnPath)
		if addr == nil {
			return errors.New("backend: invalid return path: not owned by user")
		}
	}

	su.returnPath = returnPath
	return nil
}

// Add recipient for currently processed message.
func (su *smtpUser) Rcpt(to string) error {
	log.WithField("to", to).Trace("Adding recipient")
	if to != "" {
		su.to = append(su.to, to)
	}
	return nil
}

// Set currently processed message contents and send it.
func (su *smtpUser) Data(r io.Reader) error {
	log.Trace("Sending the message")
	if su.returnPath == "" {
		return errors.New("missing return path")
	}
	if len(su.to) == 0 {
		return errors.New("missing recipient")
	}
	return su.Send(su.returnPath, su.to, r)
}

// Send sends an email from the given address to the given addresses with the given body.
func (su *smtpUser) Send(returnPath string, to []string, messageReader io.Reader) (err error) { //nolint:funlen,gocyclo
	// Called from go-smtp in goroutines - we need to handle panics for each function.
	defer su.panicHandler.HandlePanic()

	b := new(bytes.Buffer)

	messageReader = io.TeeReader(messageReader, b)

	mailSettings, err := su.client().GetMailSettings(context.TODO())
	if err != nil {
		return err
	}

	returnPathAddr := su.client().Addresses().ByEmail(returnPath)
	if returnPathAddr == nil {
		err = errors.New("backend: invalid return path: not owned by user")
		return
	}

	parser, err := parser.New(messageReader)
	if err != nil {
		err = errors.Wrap(err, "failed to create new parser")
		return
	}
	message, plainBody, attReaders, err := pkgMsg.ParserWithParser(parser)
	if err != nil {
		log.WithError(err).Error("Failed to parse message")
		return
	}
	richBody := message.Body

	externalID := message.Header.Get("Message-Id")
	externalID = strings.Trim(externalID, "<>")

	draftID, parentID := su.handleReferencesHeader(message)

	if err = su.handleSenderAndRecipients(message, returnPathAddr, returnPath, to); err != nil {
		return err
	}

	addr := su.client().Addresses().ByEmail(message.Sender.Address)
	if addr == nil {
		err = errors.New("backend: invalid email address: not owned by user")
		return
	}

	message.Sender.Address = pmapi.ConstructAddress(message.Sender.Address, addr.Email)

	kr, err := su.client().KeyRingForAddressID(addr.ID)
	if err != nil {
		return
	}

	var attachedPublicKey string
	var attachedPublicKeyName string
	if mailSettings.AttachPublicKey > 0 {
		firstKey, err := kr.GetKey(0)
		if err != nil {
			return err
		}

		attachedPublicKey, err = firstKey.GetArmoredPublicKey()
		if err != nil {
			return err
		}

		attachedPublicKeyName = fmt.Sprintf("publickey - %v - %v", kr.GetIdentities()[0].Name, firstKey.GetFingerprint()[:8])
	}

	if attachedPublicKey != "" {
		pkgMsg.AttachPublicKey(parser, attachedPublicKey, attachedPublicKeyName)
	}

	mimeBody, err := pkgMsg.BuildMIMEBody(parser)
	if err != nil {
		log.WithError(err).Error("Failed to build message")
		return
	}

	message.AddressID = addr.ID

	// Apple Mail Message-Id has to be stored to avoid recovered message after each send.
	// Before it was done only for Apple Mail, but it should work for any client. Also, the client
	// is set up from IMAP and no one can be sure that the same client is used for SMTP as well.
	// Also, user can use more than one client which could break the condition as well.
	// If there is any problem, condition to Apple Mail only should be returned.
	// Note: for that, we would need to refactor a little bit and pass the last client name from
	// the IMAP through the bridge user.
	message.ExternalID = externalID

	// If Outlook does not get a response quickly, it will try to send the message again, leading
	// to sending the same message multiple times. In case we detect the same message is in the
	// sending queue, we wait a minute to finish the first request. If the message is still being
	// sent after the timeout, we return an error back to the client. The UX is not the best,
	// but it's better than sending the message many times. If the message was sent, we simply return
	// nil to indicate it's OK.
	sendRecorderMessageHash := su.backend.sendRecorder.getMessageHash(message)
	isSending, wasSent := su.backend.sendRecorder.isSendingOrSent(su.client(), sendRecorderMessageHash)

	startTime := time.Now()
	for isSending && time.Since(startTime) < 90*time.Second {
		log.Warn("Message is still in send queue, waiting for a bit")
		time.Sleep(15 * time.Second)
		isSending, wasSent = su.backend.sendRecorder.isSendingOrSent(su.client(), sendRecorderMessageHash)
	}
	if isSending {
		log.Warn("Message is still in send queue, returning error to prevent client from adding it to the sent folder prematurely")
		return errors.New("original message is still being sent")
	}
	if wasSent {
		log.Warn("Message was already sent")
		return nil
	}

	su.backend.sendRecorder.addMessage(sendRecorderMessageHash)
	message, atts, err := su.storeUser.CreateDraft(kr, message, attReaders, attachedPublicKey, attachedPublicKeyName, parentID)
	if err != nil {
		su.backend.sendRecorder.removeMessage(sendRecorderMessageHash)
		log.WithError(err).Error("Draft could not be created")
		return err
	}
	su.backend.sendRecorder.setMessageID(sendRecorderMessageHash, message.ID)
	log.WithField("messageID", message.ID).Debug("Draft was created successfully")

	// We always have to create a new draft even if there already is one,
	// because clients don't necessarily save the draft before sending, which
	// can lead to sending the wrong message. Also clients do not necessarily
	// delete the old draft.
	if draftID != "" {
		if err := su.client().DeleteMessages(context.TODO(), []string{draftID}); err != nil {
			log.WithError(err).WithField("draftID", draftID).Warn("Original draft cannot be deleted")
		}
	}

	atts = append(atts, message.Attachments...)
	// Decrypt attachment keys, because we will need to re-encrypt them with the recipients' public keys.
	attkeys := make(map[string]*crypto.SessionKey)

	for _, att := range atts {
		var keyPackets []byte
		if keyPackets, err = base64.StdEncoding.DecodeString(att.KeyPackets); err != nil {
			return errors.Wrap(err, "decoding attachment key packets")
		}
		if attkeys[att.ID], err = kr.DecryptSessionKey(keyPackets); err != nil {
			return errors.Wrap(err, "decrypting attachment session key")
		}
	}

	req := pmapi.NewSendMessageReq(kr, mimeBody, plainBody, richBody, attkeys)
	containsUnencryptedRecipients := false

	for _, recipient := range message.Recipients() {
		email := recipient.Address
		if !looksLikeEmail(email) {
			return errors.New(`"` + email + `" is not a valid recipient.`)
		}

		sendPreferences, err := su.getSendPreferences(email, message.MIMEType, mailSettings)
		if !sendPreferences.Encrypt {
			containsUnencryptedRecipients = true
		}
		if err != nil {
			return err
		}

		var signature pmapi.SignatureFlag
		if sendPreferences.Sign {
			signature = pmapi.SignatureDetached
		} else {
			signature = pmapi.SignatureNone
		}

		if err := req.AddRecipient(email, sendPreferences.Scheme, sendPreferences.PublicKey, signature, sendPreferences.MIMEType, sendPreferences.Encrypt); err != nil {
			return errors.Wrap(err, "failed to add recipient")
		}
	}

	if containsUnencryptedRecipients {
		dec := new(mime.WordDecoder)
		subject, err := dec.DecodeHeader(message.Header.Get("Subject"))
		if err != nil {
			return errors.New("error decoding subject message " + message.Header.Get("Subject"))
		}
		if !su.continueSendingUnencryptedMail(subject) {
			if err := su.client().DeleteMessages(context.TODO(), []string{message.ID}); err != nil {
				log.WithError(err).Warn("Failed to delete canceled messages")
			}
			return errors.New("sending was canceled by user")
		}
	}

	req.PreparePackages()

	dumpMessageData(b.Bytes(), message.Subject)

	return su.storeUser.SendMessage(message.ID, req)
}

func (su *smtpUser) handleReferencesHeader(m *pmapi.Message) (draftID, parentID string) {
	// Remove the internal IDs from the references header before sending to avoid confusion.
	references := m.Header.Get("References")
	newReferences := []string{}
	for _, reference := range strings.Fields(references) {
		if !strings.Contains(reference, "@"+pmapi.InternalIDDomain) {
			newReferences = append(newReferences, reference)
		} else { // internalid is the parentID.
			idMatch := pmapi.RxInternalReferenceFormat.FindStringSubmatch(reference)
			if len(idMatch) == 2 {
				lastID := idMatch[1]
				filter := &pmapi.MessagesFilter{ID: []string{lastID}}
				if su.addressID != "" {
					filter.AddressID = su.addressID
				}
				metadata, _, _ := su.client().ListMessages(context.TODO(), filter)
				for _, m := range metadata {
					if m.IsDraft() {
						draftID = m.ID
					} else {
						parentID = m.ID
					}
				}
			}
		}
	}

	m.Header["References"] = newReferences

	if parentID == "" && len(newReferences) > 0 {
		externalID := strings.Trim(newReferences[len(newReferences)-1], "<>")
		filter := &pmapi.MessagesFilter{ExternalID: externalID}
		if su.addressID != "" {
			filter.AddressID = su.addressID
		}
		metadata, _, _ := su.client().ListMessages(context.TODO(), filter)
		// There can be two or messages with the same external ID and then we cannot
		// be sure which message should be parent. Better to not choose any.
		if len(metadata) == 1 {
			parentID = metadata[0].ID
		}
	}

	return draftID, parentID
}

func (su *smtpUser) handleSenderAndRecipients(m *pmapi.Message, returnPathAddr *pmapi.Address, returnPath string, to []string) (err error) {
	returnPath = pmapi.ConstructAddress(returnPath, returnPathAddr.Email)

	// Check sender.
	if m.Sender == nil {
		m.Sender = &mail.Address{Address: returnPath}
	} else if m.Sender.Address == "" {
		m.Sender.Address = returnPath
	}

	// Check recipients.
	if len(to) == 0 {
		err = errors.New("backend: no recipient specified")
		return
	}

	// Sanitize ToList because some clients add *Sender* in the *ToList* when only Bcc is filled.
	i := 0
	for _, keep := range m.ToList {
		keepThis := false
		for _, addr := range to {
			if addr == keep.Address {
				keepThis = true
				break
			}
		}
		if keepThis {
			m.ToList[i] = keep
			i++
		}
	}
	m.ToList = m.ToList[:i]

	// Build a map of recipients visible to all.
	// Bcc should be empty when sending a message.
	var recipients []*mail.Address
	recipients = append(recipients, m.ToList...)
	recipients = append(recipients, m.CCList...)
	recipients = append(recipients, m.BCCList...)

	rm := map[string]bool{}
	for _, r := range recipients {
		rm[r.Address] = true
	}

	for _, r := range to {
		if !rm[r] {
			// Recipient is not known, add it to Bcc.
			m.BCCList = append(m.BCCList, &mail.Address{Address: r})
		}
	}

	return nil
}

func (su *smtpUser) continueSendingUnencryptedMail(subject string) bool {
	if !su.backend.shouldReportOutgoingNoEnc() {
		return true
	}

	// GUI should always respond in 10 seconds, but let's have safety timeout
	// in case GUI will not respond properly. If GUI didn't respond, we cannot
	// be sure if user even saw the notice: better to not send the e-mail.
	req := su.backend.confirmer.NewRequest(15 * time.Second)

	su.eventListener.Emit(events.OutgoingNoEncEvent, req.ID()+":"+subject)

	res, err := req.Result()
	if err != nil {
		logrus.WithError(err).Error("Failed to determine whether to send unencrypted, assuming no")
		return false
	}

	return res
}

// Logout is called when this User will no longer be used.
func (su *smtpUser) Logout() error {
	log.Debug("SMTP client logged out user ", su.addressID)
	return nil
}
