// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"bytes"
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

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp/packet"
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
	FlagReceived   = 1
	FlagSent       = 2
	FlagInternal   = 4
	FlagE2E        = 8
	FlagAuto       = 16
	FlagReplied    = 32
	FlagRepliedAll = 64
	FlagForwarded  = 128

	FlagAutoreplied = 256
	FlagImported    = 512
	FlagOpened      = 1024
	FlagReceiptSent = 2048
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
	KeepLabels    LabelsOperation = iota // Do nothing.
	ReplaceLabels                        // Replace current labels with new ones.
	AddLabels                            // Add new labels to current ones.
	RemoveLabels                         // Remove specified labels from current ones.
)

const (
	MessageTypeInbox int = iota
	MessageTypeDraft
	MessageTypeSent
	MessageTypeInboxAndSent
)

// Due to API limitations, we shouldn't make requests with more than 100 message IDs at a time.
const messageIDPageSize = 100

// ConversationIDDomain is used as a placeholder for conversation reference headers to improve compatibility with various clients.
const ConversationIDDomain = `protonmail.conversationid`

// InternalIDDomain is used as a placeholder for reference/message ID headers to improve compatibility with various clients.
const InternalIDDomain = `protonmail.internalid`

// RxInternalReferenceFormat is compiled regexp which describes the match for
// a message ID used in reference headers.
var RxInternalReferenceFormat = regexp.MustCompile(`(?U)<(.+)@` + regexp.QuoteMeta(InternalIDDomain) + `>`) //nolint[gochecknoglobals]

// Message structure.
type Message struct {
	ID             string `json:",omitempty"`
	Order          int64  `json:",omitempty"`
	ConversationID string `json:",omitempty"` // only filter
	Subject        string
	Unread         int
	Type           int
	Flags          int64
	Sender         *mail.Address
	ReplyTo        *mail.Address   `json:",omitempty"`
	ReplyTos       []*mail.Address `json:",omitempty"`
	ToList         []*mail.Address
	CCList         []*mail.Address
	BCCList        []*mail.Address
	Time           int64 // Unix time
	Size           int64
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

func (m *Message) Decrypt(kr *crypto.KeyRing) (err error) {
	if m.IsLegacyMessage() {
		return m.DecryptLegacy(kr)
	}

	if !m.IsBodyEncrypted() {
		return
	}

	armored := strings.TrimSpace(m.Body)
	body, err := decrypt(kr, armored)
	if err != nil {
		return
	}

	m.Body = body
	return
}

func (m *Message) DecryptLegacy(kr *crypto.KeyRing) (err error) {
	randomKeyStart := strings.Index(m.Body, RandomKeyHeader) + len(RandomKeyHeader)
	randomKeyEnd := strings.Index(m.Body, RandomKeyTail)
	randomKey := m.Body[randomKeyStart:randomKeyEnd]

	signedKey, err := decrypt(kr, strings.TrimSpace(randomKey))
	if err != nil {
		return
	}
	bytesKey, err := decodeBase64UTF8(signedKey)
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

	m.Body = string(bytesPlaintext)
	return err
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

func (filter *MessagesFilter) urlValues() url.Values { // nolint[funlen]
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

type MessagesListRes struct {
	Res

	Total    int
	Messages []*Message
}

// ListMessages gets message metadata.
func (c *client) ListMessages(filter *MessagesFilter) (msgs []*Message, total int, err error) {
	req, err := c.NewRequest("GET", "/mail/v4/messages", nil)
	if err != nil {
		return
	}

	req.URL.RawQuery = filter.urlValues().Encode()
	var res MessagesListRes
	if err = c.DoJSON(req, &res); err != nil {
		// If the URI was too long and we searched with IDs, we will try again without the API IDs.
		if strings.Contains(err.Error(), "api returned: 414") && len(filter.ID) > 0 {
			filter.ID = []string{}
			return c.ListMessages(filter)
		}
		return
	}

	msgs, total, err = res.Messages, res.Total, res.Err()
	return
}

type MessagesCountsRes struct {
	Res

	Counts []*MessagesCount
}

// CountMessages counts messages by label.
func (c *client) CountMessages(addressID string) (counts []*MessagesCount, err error) {
	reqURL := "/mail/v4/messages/count"
	if addressID != "" {
		reqURL += ("?AddressID=" + addressID)
	}
	req, err := c.NewRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	var res MessagesCountsRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	counts, err = res.Counts, res.Err()
	return
}

type MessageRes struct {
	Res

	Message *Message
}

// GetMessage retrieves a message.
func (c *client) GetMessage(id string) (msg *Message, err error) {
	req, err := c.NewRequest("GET", "/mail/v4/messages/"+id, nil)
	if err != nil {
		return
	}

	var res MessageRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	return res.Message, res.Err()
}

type MessagesActionReq struct {
	IDs []string
}

type MessagesActionRes struct {
	Res

	Responses []struct {
		ID       string
		Response Res
	}
}

func (res MessagesActionRes) Err() error {
	if err := res.Res.Err(); err != nil {
		return err
	}

	for _, msgRes := range res.Responses {
		if err := msgRes.Response.Err(); err != nil {
			return err
		}
	}

	return nil
}

// doMessagesAction performs paged requests to doMessagesActionInner.
// This can eventually be done in parallel though.
func (c *client) doMessagesAction(action string, ids []string) (err error) {
	for len(ids) > messageIDPageSize {
		var requestIDs []string
		requestIDs, ids = ids[:messageIDPageSize], ids[messageIDPageSize:]
		if err = c.doMessagesActionInner(action, requestIDs); err != nil {
			return
		}
	}

	return c.doMessagesActionInner(action, ids)
}

// doMessagesActionInner is the non-paged inner method of doMessagesAction.
// You should not call this directly unless you know what you are doing (it can overload the server).
func (c *client) doMessagesActionInner(action string, ids []string) (err error) {
	actionReq := &MessagesActionReq{IDs: ids}
	req, err := c.NewJSONRequest("PUT", "/mail/v4/messages/"+action, actionReq)
	if err != nil {
		return
	}

	var res MessagesActionRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()

	return
}

func (c *client) MarkMessagesRead(ids []string) error {
	return c.doMessagesAction("read", ids)
}

func (c *client) MarkMessagesUnread(ids []string) error {
	return c.doMessagesAction("unread", ids)
}

func (c *client) DeleteMessages(ids []string) error {
	return c.doMessagesAction("delete", ids)
}

func (c *client) UndeleteMessages(ids []string) error {
	return c.doMessagesAction("undelete", ids)
}

type LabelMessagesReq struct {
	LabelID string
	IDs     []string
}

// LabelMessages labels the given message IDs with the given label.
// The requests are performed paged; this can eventually be done in parallel.
func (c *client) LabelMessages(ids []string, label string) (err error) {
	for len(ids) > messageIDPageSize {
		var requestIDs []string
		requestIDs, ids = ids[:messageIDPageSize], ids[messageIDPageSize:]
		if err = c.labelMessages(requestIDs, label); err != nil {
			return
		}
	}

	return c.labelMessages(ids, label)
}

func (c *client) labelMessages(ids []string, label string) (err error) {
	labelReq := &LabelMessagesReq{LabelID: label, IDs: ids}
	req, err := c.NewJSONRequest("PUT", "/mail/v4/messages/label", labelReq)
	if err != nil {
		return
	}

	var res MessagesActionRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
}

// UnlabelMessages removes the given label from the given message IDs.
// The requests are performed paged; this can eventually be done in parallel.
func (c *client) UnlabelMessages(ids []string, label string) (err error) {
	for len(ids) > messageIDPageSize {
		var requestIDs []string
		requestIDs, ids = ids[:messageIDPageSize], ids[messageIDPageSize:]
		if err = c.unlabelMessages(requestIDs, label); err != nil {
			return
		}
	}

	return c.unlabelMessages(ids, label)
}

func (c *client) unlabelMessages(ids []string, label string) (err error) {
	labelReq := &LabelMessagesReq{LabelID: label, IDs: ids}
	req, err := c.NewJSONRequest("PUT", "/mail/v4/messages/unlabel", labelReq)
	if err != nil {
		return
	}

	var res MessagesActionRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
}

func (c *client) EmptyFolder(labelID, addressID string) (err error) {
	if labelID == "" {
		return errors.New("pmapi: labelID parameter is empty string")
	}
	reqURL := "/mail/v4/messages/empty?LabelID=" + labelID
	if addressID != "" {
		reqURL += ("&AddressID=" + addressID)
	}

	req, err := c.NewRequest("DELETE", reqURL, nil)

	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
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
