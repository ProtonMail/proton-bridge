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

package pmapi

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// Header types.
const (
	MessageHeader       = "-----BEGIN PGP MESSAGE-----"
	MessageTail         = "-----END PGP MESSAGE-----"
	MessageHeaderLegacy = "---BEGIN ENCRYPTED MESSAGE---"
	MessageTailLegacy   = "---END ENCRYPTED MESSAGE---"
	RandomKeyHeader     = "---BEGIN ENCRYPTED RANDOM KEY---"
	RandomKeyTail       = "---END ENCRYPTED RANDOM KEY---"
)

// Sort types.
const (
	SortByTo      = "To"
	SortByFrom    = "From"
	SortBySubject = "Subject"
	SortBySize    = "Size"
	SortByTime    = "Time"
	SortByID      = "ID"
	SortDesc      = true
	SortAsc       = false
)

// Message actions.
const (
	ActionReply    = 0
	ActionReplyAll = 1
	ActionForward  = 2
)

// Message flag definitions.
const (
	FlagReceived   = int64(1)
	FlagSent       = int64(2)
	FlagInternal   = int64(4)
	FlagE2E        = int64(8)
	FlagAuto       = int64(16)
	FlagReplied    = int64(32)
	FlagRepliedAll = int64(64)
	FlagForwarded  = int64(128)

	FlagAutoreplied = int64(256)
	FlagImported    = int64(512)
	FlagOpened      = int64(1024)
	FlagReceiptSent = int64(2048)
)

// Draft flags.
const (
	FlagReceiptRequest = 1 << 16
	FlagPublicKey      = 1 << 17
	FlagSign           = 1 << 18
)

// Spam flags.
const (
	FlagSpfFail        = 1 << 24
	FlagDkimFail       = 1 << 25
	FlagDmarcFail      = 1 << 26
	FlagHamManual      = 1 << 27
	FlagSpamAuto       = 1 << 28
	FlagSpamManual     = 1 << 29
	FlagPhishingAuto   = 1 << 30
	FlagPhishingManual = 1 << 31
)

// Message flag masks.
const (
	FlagMaskGeneral = 4095
	FlagMaskDraft   = FlagReceiptRequest * 7
	FlagMaskSpam    = FlagSpfFail * 255
	FlagMask        = FlagMaskGeneral | FlagMaskDraft | FlagMaskSpam
)

// INTERNAL, AUTO are immutable. E2E is immutable except for drafts on send.
const (
	FlagMaskAdd = 4067 + (16777216 * 168)
)

// Content types.
const (
	ContentTypeMultipartMixed     = "multipart/mixed"
	ContentTypeMultipartEncrypted = "multipart/encrypted"
	ContentTypePlainText          = "text/plain"
	ContentTypeHTML               = "text/html"
)

// LabelsOperation is the operation to apply to labels.
type LabelsOperation int

const (
	KeepLabels    LabelsOperation = iota // KeepLabels Do nothing.
	ReplaceLabels                        // ReplaceLabels Replace current labels with new ones.
	AddLabels                            // AddLabels Add new labels to current ones.
	RemoveLabels                         // RemoveLabels Remove specified labels from current ones.
)

// Due to API limitations, we shouldn't make requests with more than 100 message IDs at a time.
const messageIDPageSize = 100

// ConversationIDDomain is used as a placeholder for conversation reference headers to improve compatibility with various clients.
const ConversationIDDomain = `protonmail.conversationid`

// InternalIDDomain is used as a placeholder for reference/message ID headers to improve compatibility with various clients.
const InternalIDDomain = `protonmail.internalid`

// RxInternalReferenceFormat is compiled regexp which describes the match for
// a message ID used in reference headers.
var RxInternalReferenceFormat = regexp.MustCompile(`(?U)<(.+)@` + regexp.QuoteMeta(InternalIDDomain) + `>`) //nolint:gochecknoglobals

// Message structure.
type Message struct {
	ID             string `json:",omitempty"`
	Order          int64  `json:",omitempty"`
	ConversationID string `json:",omitempty"` // only filter
	Subject        string
	Unread         Boolean
	Flags          int64
	Sender         *mail.Address
	ReplyTo        *mail.Address   `json:",omitempty"`
	ReplyTos       []*mail.Address `json:",omitempty"`
	ToList         []*mail.Address
	CCList         []*mail.Address
	BCCList        []*mail.Address
	Time           int64 // Unix time
	NumAttachments int
	ExpirationTime int64 // Unix time
	SpamScore      int
	AddressID      string
	Body           string `json:",omitempty"`
	Attachments    []*Attachment
	LabelIDs       []string
	ExternalID     string
	Header         mail.Header
	MIMEType       string
}

// NewMessage initializes a new message.
func NewMessage() *Message {
	return &Message{
		ToList:      []*mail.Address{},
		CCList:      []*mail.Address{},
		BCCList:     []*mail.Address{},
		Attachments: []*Attachment{},
		LabelIDs:    []string{},
	}
}

// Define a new type to prevent MarshalJSON/UnmarshalJSON infinite loops.
type message Message

type rawMessage struct {
	message

	Header string `json:",omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	var raw rawMessage
	raw.message = message(*m)

	b := &bytes.Buffer{}
	_ = http.Header(m.Header).Write(b)
	raw.Header = b.String()

	return json.Marshal(&raw)
}

func (m *Message) UnmarshalJSON(b []byte) error {
	var raw rawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	*m = Message(raw.message)

	if raw.Header != "" && raw.Header != "(No Header)" {
		msg, err := mail.ReadMessage(strings.NewReader(raw.Header + "\r\n\r\n"))
		if err != nil {
			logrus.WithField("rawHeader", raw.Header).Trace("Failed to parse header")
			return fmt.Errorf("failed to parse header of message %v: %v", m.ID, err.Error())
		}
		m.Header = msg.Header
	} else {
		m.Header = make(mail.Header)
	}

	return nil
}

// IsDraft returns whether the message should be considered to be a draft.
// A draft is complicated. It might have pmapi.DraftLabel but it might not.
// The real API definition of IsDraft is that it is neither sent nor received -- we should use that here.
func (m *Message) IsDraft() bool {
	return (m.Flags & (FlagReceived | FlagSent)) == 0
}

// HasLabelID returns whether the message has the `labelID`.
func (m *Message) HasLabelID(labelID string) bool {
	for _, l := range m.LabelIDs {
		if l == labelID {
			return true
		}
	}
	return false
}

func (m *Message) IsEncrypted() bool {
	return strings.HasPrefix(m.Header.Get("Content-Type"), "multipart/encrypted") || m.IsBodyEncrypted()
}

func (m *Message) IsBodyEncrypted() bool {
	trimmedBody := strings.TrimSpace(m.Body)
	return strings.HasPrefix(trimmedBody, MessageHeader) &&
		strings.HasSuffix(trimmedBody, MessageTail)
}

func (m *Message) IsLegacyMessage() bool {
	return strings.Contains(m.Body, RandomKeyHeader) &&
		strings.Contains(m.Body, RandomKeyTail) &&
		strings.Contains(m.Body, MessageHeaderLegacy) &&
		strings.Contains(m.Body, MessageTailLegacy) &&
		strings.Contains(m.Body, MessageHeader) &&
		strings.Contains(m.Body, MessageTail)
}

func (m *Message) Decrypt(kr *crypto.KeyRing) ([]byte, error) {
	if m.IsLegacyMessage() {
		return m.decryptLegacy(kr)
	}

	if !m.IsBodyEncrypted() {
		return []byte(m.Body), nil
	}

	armored := strings.TrimSpace(m.Body)

	body, err := decrypt(kr, armored)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type Signature struct {
	Hash string
	Data []byte
}

func (m *Message) ExtractSignatures(kr *crypto.KeyRing) ([]Signature, error) {
	var entities openpgp.EntityList

	for _, key := range kr.GetKeys() {
		entities = append(entities, key.GetEntity())
	}

	p, err := armor.Decode(strings.NewReader(m.Body))
	if err != nil {
		return nil, err
	}

	msg, err := openpgp.ReadMessage(p.Body, entities, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err := io.ReadAll(msg.UnverifiedBody); err != nil {
		return nil, err
	}

	if !msg.IsSigned {
		return nil, nil
	}

	signatures := make([]Signature, 0, len(msg.UnverifiedSignatures))

	for _, signature := range msg.UnverifiedSignatures {
		buf := new(bytes.Buffer)

		if err := signature.Serialize(buf); err != nil {
			return nil, err
		}

		signatures = append(signatures, Signature{
			Hash: signature.Hash.String(),
			Data: buf.Bytes(),
		})
	}

	return signatures, nil
}

func (m *Message) decryptLegacy(kr *crypto.KeyRing) (dec []byte, err error) {
	randomKeyStart := strings.Index(m.Body, RandomKeyHeader) + len(RandomKeyHeader)
	randomKeyEnd := strings.Index(m.Body, RandomKeyTail)
	randomKey := m.Body[randomKeyStart:randomKeyEnd]

	signedKey, err := decrypt(kr, strings.TrimSpace(randomKey))
	if err != nil {
		return
	}
	bytesKey, err := decodeBase64UTF8(string(signedKey))
	if err != nil {
		return
	}

	messageStart := strings.Index(m.Body, MessageHeaderLegacy) + len(MessageHeaderLegacy)
	messageEnd := strings.Index(m.Body, MessageTailLegacy)
	message := m.Body[messageStart:messageEnd]
	bytesMessage, err := decodeBase64UTF8(message)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(bytesKey)
	if err != nil {
		return
	}

	prefix := make([]byte, block.BlockSize()+2)
	bytesMessageReader := bytes.NewReader(bytesMessage)

	_, err = io.ReadFull(bytesMessageReader, prefix)
	if err != nil {
		return
	}
	s := packet.NewOCFBDecrypter(block, prefix, packet.OCFBResync)
	if s == nil {
		err = errors.New("pmapi: incorrect key for legacy decryption")
		return
	}

	reader := cipher.StreamReader{S: s, R: bytesMessageReader}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)
	plaintextBytes := buf.Bytes()

	plaintext := ""
	for i := 0; i < len(plaintextBytes); i++ {
		plaintext += string(plaintextBytes[i])
	}
	bytesPlaintext, err := decodeBase64UTF8(plaintext)
	if err != nil {
		return
	}

	return bytesPlaintext, nil
}

func decodeBase64UTF8(input string) (output []byte, err error) {
	input = strings.TrimSpace(input)
	decodedMessage, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return
	}
	utf8DecodedMessage := []rune(string(decodedMessage))
	output = make([]byte, len(utf8DecodedMessage))
	for i := 0; i < len(utf8DecodedMessage); i++ {
		output[i] = byte(int(utf8DecodedMessage[i]))
	}
	return
}

func (m *Message) Encrypt(encrypter, signer *crypto.KeyRing) (err error) {
	if m.IsBodyEncrypted() {
		err = errors.New("pmapi: trying to encrypt an already encrypted message")
		return
	}

	m.Body, err = encrypt(encrypter, m.Body, signer)
	return
}

func (m *Message) Has(flag int64) bool {
	return (m.Flags & flag) == flag
}

func (m *Message) Recipients() []*mail.Address {
	var recipients []*mail.Address
	recipients = append(recipients, m.ToList...)
	recipients = append(recipients, m.CCList...)
	recipients = append(recipients, m.BCCList...)
	return recipients
}

// MessagesCount contains message counts for one label.
type MessagesCount struct {
	LabelID string
	Total   int
	Unread  int
}

// MessagesFilter contains fields to filter messages.
type MessagesFilter struct {
	Page           int
	PageSize       int
	Limit          int
	LabelID        string
	Sort           string // Time by default (Time, To, From, Subject, Size).
	Desc           *bool
	Begin          int64 // Unix time.
	End            int64 // Unix time.
	BeginID        string
	EndID          string
	Keyword        string
	To             string
	From           string
	Subject        string
	ConversationID string
	AddressID      string
	ID             []string
	Attachments    *bool
	Unread         *bool
	ExternalID     string // MIME Message-Id (only valid for messages).
	AutoWildcard   *bool
}

func (filter *MessagesFilter) urlValues() url.Values { //nolint:funlen
	v := url.Values{}

	if filter.Page != 0 {
		v.Set("Page", strconv.Itoa(filter.Page))
	}
	if filter.PageSize != 0 {
		v.Set("PageSize", strconv.Itoa(filter.PageSize))
	}
	if filter.Limit != 0 {
		v.Set("Limit", strconv.Itoa(filter.Limit))
	}
	if filter.LabelID != "" {
		v.Set("LabelID", filter.LabelID)
	}
	if filter.Sort != "" {
		v.Set("Sort", filter.Sort)
	}
	if filter.Desc != nil {
		if *filter.Desc {
			v.Set("Desc", "1")
		} else {
			v.Set("Desc", "0")
		}
	}
	if filter.Begin != 0 {
		v.Set("Begin", strconv.Itoa(int(filter.Begin)))
	}
	if filter.End != 0 {
		v.Set("End", strconv.Itoa(int(filter.End)))
	}
	if filter.BeginID != "" {
		v.Set("BeginID", filter.BeginID)
	}
	if filter.EndID != "" {
		v.Set("EndID", filter.EndID)
	}
	if filter.Keyword != "" {
		v.Set("Keyword", filter.Keyword)
	}
	if filter.To != "" {
		v.Set("To", filter.To)
	}
	if filter.From != "" {
		v.Set("From", filter.From)
	}
	if filter.Subject != "" {
		v.Set("Subject", filter.Subject)
	}
	if filter.ConversationID != "" {
		v.Set("ConversationID", filter.ConversationID)
	}
	if filter.AddressID != "" {
		v.Set("AddressID", filter.AddressID)
	}
	if len(filter.ID) > 0 {
		for _, id := range filter.ID {
			v.Add("ID[]", id)
		}
	}
	if filter.Attachments != nil {
		if *filter.Attachments {
			v.Set("Attachments", "1")
		} else {
			v.Set("Attachments", "0")
		}
	}
	if filter.Unread != nil {
		if *filter.Unread {
			v.Set("Unread", "1")
		} else {
			v.Set("Unread", "0")
		}
	}
	if filter.ExternalID != "" {
		v.Set("ExternalID", filter.ExternalID)
	}
	if filter.AutoWildcard != nil {
		if *filter.AutoWildcard {
			v.Set("AutoWildcard", "1")
		} else {
			v.Set("AutoWildcard", "0")
		}
	}

	return v
}

// ListMessages gets message metadata.
func (c *client) ListMessages(ctx context.Context, filter *MessagesFilter) ([]*Message, int, error) {
	var res struct {
		Messages []*Message
		Total    int
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetQueryParamsFromValues(filter.urlValues()).
			SetResult(&res).
			Get("/mail/v4/messages")
	}); err != nil {
		return nil, 0, err
	}

	return res.Messages, res.Total, nil
}

// CountMessages counts messages by label.
func (c *client) CountMessages(ctx context.Context, addressID string) (counts []*MessagesCount, err error) {
	var res struct {
		Counts []*MessagesCount
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		if addressID != "" {
			r = r.SetQueryParam("AddressID", addressID)
		}
		return r.SetResult(&res).Get("/mail/v4/messages/count")
	}); err != nil {
		return nil, err
	}

	return res.Counts, nil
}

// GetMessage retrieves a message.
func (c *client) GetMessage(ctx context.Context, messageID string) (msg *Message, err error) {
	var res struct {
		Message *Message
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/mail/v4/messages/" + messageID)
	}); err != nil {
		return nil, err
	}

	return res.Message, nil
}

type MessagesActionReq struct {
	IDs []string
}

func (c *client) MarkMessagesRead(ctx context.Context, messageIDs []string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := MessagesActionReq{IDs: messageIDs}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/read")
		}); err != nil {
			return err
		}

		return nil
	})
}

func (c *client) MarkMessagesUnread(ctx context.Context, messageIDs []string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := MessagesActionReq{IDs: messageIDs}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/unread")
		}); err != nil {
			return err
		}

		return nil
	})
}

func (c *client) DeleteMessages(ctx context.Context, messageIDs []string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := MessagesActionReq{IDs: messageIDs}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/delete")
		}); err != nil {
			return err
		}

		return nil
	})
}

func (c *client) UndeleteMessages(ctx context.Context, messageIDs []string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := MessagesActionReq{IDs: messageIDs}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/undelete")
		}); err != nil {
			return err
		}

		return nil
	})
}

type LabelMessagesReq struct {
	LabelID string
	IDs     []string
}

// LabelMessages labels the given message IDs with the given label.
// The requests are performed paged; this can eventually be done in parallel.
func (c *client) LabelMessages(ctx context.Context, messageIDs []string, labelID string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := LabelMessagesReq{
			LabelID: labelID,
			IDs:     messageIDs,
		}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/label")
		}); err != nil {
			return err
		}

		return nil
	})
}

// UnlabelMessages removes the given label from the given message IDs.
// The requests are performed paged; this can eventually be done in parallel.
func (c *client) UnlabelMessages(ctx context.Context, messageIDs []string, labelID string) error {
	return doPaged(messageIDs, defaultPageSize, func(messageIDs []string) (err error) {
		req := LabelMessagesReq{
			LabelID: labelID,
			IDs:     messageIDs,
		}

		if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
			return r.SetBody(req).Put("/mail/v4/messages/unlabel")
		}); err != nil {
			return err
		}

		return nil
	})
}

func (c *client) EmptyFolder(ctx context.Context, labelID, addressID string) error {
	if labelID == "" {
		return errors.New("labelID parameter is empty string")
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		if addressID != "" {
			r.SetQueryParam("AddressID", addressID)
		}

		return r.SetQueryParam("LabelID", labelID).Delete("/mail/v4/messages/empty")
	}); err != nil {
		return err
	}

	return nil
}

// ComputeMessageFlagsByLabels returns flags based on labels.
func ComputeMessageFlagsByLabels(labels []string) (flag int64) {
	for _, labelID := range labels {
		switch labelID {
		case SentLabel:
			flag = (flag | FlagSent)
		case ArchiveLabel, InboxLabel:
			flag = (flag | FlagReceived)
		}
	}

	// NOTE: if the labels are custom only
	if flag == 0 {
		flag = FlagReceived
	}

	return flag
}
