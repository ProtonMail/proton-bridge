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

package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	pmmime "github.com/ProtonMail/proton-bridge/v3/pkg/mime"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/messages-go/v16"
	"github.com/emersion/go-imap"
	"golang.org/x/exp/slices"
)

type Message struct {
	Subject     string `bdd:"subject"`
	Body        string `bdd:"body"`
	MIMEType    string `bdd:"mime-type"`
	Attachments string `bdd:"attachments"`
	MessageID   string `bdd:"message-id"`
	Date        string `bdd:"date"`

	From    string `bdd:"from"`
	To      string `bdd:"to"`
	CC      string `bdd:"cc"`
	BCC     string `bdd:"bcc"`
	ReplyTo string `bdd:"reply-to"`

	Unread  bool `bdd:"unread"`
	Deleted bool `bdd:"deleted"`

	InReplyTo  string `bdd:"in-reply-to"`
	References string `bdd:"references"`
}

type MessageStruct struct {
	From    string `json:"from"`
	To      string `json:"to"`
	CC      string `json:"cc"`
	BCC     string `json:"bcc"`
	Subject string `json:"subject"`
	Date    string `json:"date"`

	Content MessageSection `json:"content"`
}

type MessageSection struct {
	ContentType                string           `json:"content-type"`
	ContentTypeBoundary        string           `json:"content-type-boundary"`
	ContentTypeCharset         string           `json:"content-type-charset"`
	ContentTypeName            string           `json:"content-type-name"`
	ContentDisposition         string           `json:"content-disposition"`
	ContentDispositionFilename string           `json:"content-disposition-filename"`
	Sections                   []MessageSection `json:"sections"`

	TransferEncoding string `json:"transfer-encoding"`
	BodyContains     string `json:"body-contains"`
	BodyIs           string `json:"body-is"`
}

func (msg Message) Build() []byte {
	var b []byte

	if msg.From != "" {
		b = append(b, "From: "+msg.From+"\r\n"...)
	}

	if msg.To != "" {
		b = append(b, "To: "+msg.To+"\r\n"...)
	}

	if msg.CC != "" {
		b = append(b, "Cc: "+msg.CC+"\r\n"...)
	}

	if msg.BCC != "" {
		b = append(b, "Bcc: "+msg.BCC+"\r\n"...)
	}

	if msg.Subject != "" {
		b = append(b, "Subject: "+msg.Subject+"\r\n"...)
	}

	if msg.InReplyTo != "" {
		b = append(b, "In-Reply-To: "+msg.InReplyTo+"\r\n"...)
	}

	if msg.References != "" {
		b = append(b, "References: "+msg.References+"\r\n"...)
	}

	if msg.Date != "" {
		date, err := time.Parse(time.RFC822, msg.Date)
		if err != nil {
			panic(err)
		}
		b = append(b, "Date: "+date.Format(time.RFC822Z)+"\r\n"...)
	}

	b = append(b, "\r\n"+msg.Body+"\r\n"...)

	return b
}

func newMessageFromIMAP(msg *imap.Message) Message {
	section, err := imap.ParseBodySectionName("BODY[]")
	if err != nil {
		panic(err)
	}

	literal, err := io.ReadAll(msg.GetBody(section))
	if err != nil {
		panic(err)
	}

	mimeType, _, err := rfc822.Parse(literal).ContentType()
	if err != nil {
		panic(err)
	}

	m, err := message.Parse(bytes.NewReader(literal))
	if err != nil {
		panic(err)
	}

	var body string

	if m.MIMEType == rfc822.TextPlain {
		body = strings.TrimSpace(string(m.PlainBody))
	} else {
		body = strings.TrimSpace(string(m.RichBody))
	}

	message := Message{
		Subject:     msg.Envelope.Subject,
		Body:        body,
		MIMEType:    string(mimeType),
		Attachments: strings.Join(xslices.Map(m.Attachments, func(att message.Attachment) string { return att.Name }), ", "),
		MessageID:   msg.Envelope.MessageId,
		Unread:      !slices.Contains(msg.Flags, imap.SeenFlag),
		Deleted:     slices.Contains(msg.Flags, imap.DeletedFlag),
		Date:        msg.Envelope.Date.Format(time.RFC822Z),
		InReplyTo:   msg.Envelope.InReplyTo,
		// Go-imap only supports in-reply-to so we have to mimic other client by using it as references.
		References: msg.Envelope.InReplyTo,
	}

	if len(msg.Envelope.From) > 0 {
		message.From = msg.Envelope.From[0].Address()
	}

	if len(msg.Envelope.To) > 0 {
		message.To = msg.Envelope.To[0].Address()
	}

	if len(msg.Envelope.Cc) > 0 {
		message.CC = msg.Envelope.Cc[0].Address()
	}

	if len(msg.Envelope.Bcc) > 0 {
		message.BCC = msg.Envelope.Bcc[0].Address()
	}

	if len(msg.Envelope.ReplyTo) > 0 {
		message.ReplyTo = msg.Envelope.ReplyTo[0].Address()
	}

	return message
}

func newMessageStructFromIMAP(msg *imap.Message) MessageStruct {
	section, err := imap.ParseBodySectionName("BODY[]")
	if err != nil {
		panic(err)
	}

	literal, err := io.ReadAll(msg.GetBody(section))
	if err != nil {
		panic(err)
	}

	parser, err := parser.New(bytes.NewReader(literal))
	if err != nil {
		panic(err)
	}

	m, err := message.ParseWithParser(parser, true)
	if err != nil {
		panic(err)
	}

	var body string
	switch {
	case m.MIMEType == rfc822.TextPlain:
		body = strings.TrimSpace(string(m.PlainBody))
	case m.MIMEType == rfc822.MultipartMixed:
		_, body, _ = strings.Cut(string(m.MIMEBody), "\r\n\r\n")
	default:
		body = strings.TrimSpace(string(m.RichBody))
	}

	message := MessageStruct{
		Subject: msg.Envelope.Subject,
		Date:    msg.Envelope.Date.Format(time.RFC822Z),
		From:    formatAddressList(msg.Envelope.From),
		To:      formatAddressList(msg.Envelope.To),
		CC:      formatAddressList(msg.Envelope.Cc),
		BCC:     formatAddressList(msg.Envelope.Bcc),

		Content: parseMessageSection([]byte(strings.TrimSpace(string(literal))), strings.TrimSpace(body)),
	}
	return message
}

func formatAddressList(list []*imap.Address) string {
	var res string
	for idx, address := range list {
		if address.PersonalName != "" {
			res += address.PersonalName + " <" + address.Address() + ">"
		} else {
			res += address.Address()
		}
		if idx < len(list)-1 {
			res += "; "
		}
	}
	return res
}

func parseMessageSection(literal []byte, body string) MessageSection {
	headers, err := rfc822.Parse(literal).ParseHeader()
	if err != nil {
		panic(err)
	}

	mimeType, boundary, charset, name := parseContentType(headers.Get("Content-Type"))
	disp, filename := parseContentDisposition(headers.Get("Content-Disposition"))

	msgSect := MessageSection{
		ContentType:                mimeType,
		ContentTypeBoundary:        boundary,
		ContentTypeCharset:         charset,
		ContentTypeName:            name,
		ContentDisposition:         disp,
		ContentDispositionFilename: filename,
		TransferEncoding:           headers.Get("content-transfer-encoding"),
		BodyIs:                     body,
	}

	if msgSect.ContentTypeBoundary != "" {
		sections := bytes.Split(literal, []byte("--"+msgSect.ContentTypeBoundary))
		// Remove last element that will be the -- from finale boundary
		sections = sections[:len(sections)-1]
		sections = sections[1:]
		for _, v := range sections {
			str := strings.TrimSpace(string(v))
			_, sectionBody, found := strings.Cut(str, "\r\n\r\n")
			if !found {
				if _, sectionBody, found = strings.Cut(str, "\n\n"); !found {
					sectionBody = str
				}
			}
			msgSect.Sections = append(msgSect.Sections, parseMessageSection([]byte(str), strings.TrimSpace(sectionBody)))
		}
	}
	return msgSect
}

func parseContentType(contentType string) (string, string, string, string) {
	mimeType, params, err := pmmime.ParseMediaType(contentType)
	if err != nil {
		panic(err)
	}
	boundary, ok := params["boundary"]
	if !ok {
		boundary = ""
	}
	charset, ok := params["charset"]
	if !ok {
		charset = ""
	}
	name, ok := params["name"]
	if !ok {
		name = ""
	}
	return mimeType, boundary, charset, name
}

func parseContentDisposition(contentDisp string) (string, string) {
	disp, params, _ := pmmime.ParseMediaType(contentDisp)
	name, ok := params["filename"]
	if !ok {
		name = ""
	}
	return disp, name
}

func matchMessages(have, want []Message) error {
	slices.SortFunc(have, func(a, b Message) bool {
		return a.Subject < b.Subject
	})

	slices.SortFunc(want, func(a, b Message) bool {
		return a.Subject < b.Subject
	})

	if !IsSub(ToAny(have), ToAny(want)) {
		return fmt.Errorf("missing messages: have %#v, want %#v", have, want)
	}

	return nil
}

func matchStructure(have []MessageStruct, want MessageStruct) error {
	mismatches := make([]string, 0)
	for _, msg := range have {
		if want.From != "" && msg.From != want.From {
			mismatches = append(mismatches, "From")
			continue
		}
		if want.To != "" && msg.To != want.To {
			mismatches = append(mismatches, "To")
			continue
		}
		if want.BCC != "" && msg.BCC != want.BCC {
			mismatches = append(mismatches, "BCC")
			continue
		}
		if want.CC != "" && msg.CC != want.CC {
			mismatches = append(mismatches, "CC")
			continue
		}
		if want.Subject != "" && msg.Subject != want.Subject {
			mismatches = append(mismatches, "Subject")
			continue
		}
		if want.Date != "" && want.Date != msg.Date {
			mismatches = append(mismatches, "Date")
			continue
		}

		if ok, mismatch := matchContent(msg.Content, want.Content); !ok {
			mismatches = append(mismatches, "Content: "+mismatch)
			continue
		}
		return nil
	}
	return fmt.Errorf("missing messages: have %#v, want %#v with mismatch list %#v", have, want, mismatches)
}

func matchContent(have MessageSection, want MessageSection) (bool, string) {
	if want.ContentType != "" && want.ContentType != have.ContentType {
		return false, "ContentType"
	}
	if want.ContentTypeBoundary != "" && want.ContentTypeBoundary != have.ContentTypeBoundary {
		return false, "ContentTypeBoundary"
	}
	if want.ContentTypeCharset != "" && want.ContentTypeCharset != have.ContentTypeCharset {
		return false, "ContentTypeCharset"
	}
	if want.ContentTypeName != "" && want.ContentTypeName != have.ContentTypeName {
		return false, "ContentTypeName"
	}
	if want.ContentDisposition != "" && want.ContentDisposition != have.ContentDisposition {
		return false, "ContentDisposition"
	}
	if want.ContentDispositionFilename != "" && want.ContentDispositionFilename != have.ContentDispositionFilename {
		return false, "ContentDispositionFilename"
	}
	if want.TransferEncoding != "" && want.TransferEncoding != have.TransferEncoding {
		return false, "TransferEncoding"
	}
	if want.BodyContains != "" && !strings.Contains(strings.TrimSpace(have.BodyIs), strings.TrimSpace(want.BodyContains)) {
		return false, "BodyContains"
	}
	if want.BodyIs != "" && strings.TrimSpace(have.BodyIs) != strings.TrimSpace(want.BodyIs) {
		return false, "BodyIs"
	}

	for i, section := range want.Sections {
		if ok, mismatch := matchContent(have.Sections[i], section); !ok {
			return false, fmt.Sprintf("section %#v - %#v", i, mismatch)
		}
	}
	return true, ""
}

func matchStructureRecursive(have []MessageStruct, want MessageStruct) error {
	mismatches := make([]string, 0)
	for _, msg := range have {
		if want.From != "" && msg.From != want.From {
			mismatches = append(mismatches, "From")
			continue
		}
		if want.To != "" && msg.To != want.To {
			mismatches = append(mismatches, "To")
			continue
		}
		if want.BCC != "" && msg.BCC != want.BCC {
			mismatches = append(mismatches, "BCC")
			continue
		}
		if want.CC != "" && msg.CC != want.CC {
			mismatches = append(mismatches, "CC")
			continue
		}
		if want.Subject != "" && msg.Subject != want.Subject {
			mismatches = append(mismatches, "Subject")
			continue
		}
		if want.Date != "" && want.Date != msg.Date {
			mismatches = append(mismatches, "Date")
			continue
		}

		if ok, mismatch := matchContentRecursive(msg.Content, want.Content); !ok {
			mismatches = append(mismatches, "Content: "+mismatch)
			continue
		}
		return nil
	}
	return fmt.Errorf("missing messages: have %#v, want %#v with mismatch list %#v", have, want, mismatches)
}

func matchContentRecursive(have MessageSection, want MessageSection) (bool, string) {
	if want.ContentType != "" && !strings.EqualFold(want.ContentType, have.ContentType) {
		return false, "ContentType"
	}
	if want.ContentTypeBoundary != "" && !strings.EqualFold(want.ContentTypeBoundary, have.ContentTypeBoundary) {
		return false, "ContentTypeBoundary"
	}
	if want.ContentTypeCharset != "" && !strings.EqualFold(want.ContentTypeCharset, have.ContentTypeCharset) {
		return false, "ContentTypeCharset"
	}
	if want.ContentTypeName != "" && !strings.EqualFold(want.ContentTypeName, have.ContentTypeName) {
		return false, "ContentTypeName"
	}
	if want.ContentDisposition != "" && !strings.EqualFold(want.ContentDisposition, have.ContentDisposition) {
		return false, "ContentDisposition"
	}
	if want.ContentDispositionFilename != "" && !strings.EqualFold(want.ContentDispositionFilename, have.ContentDispositionFilename) {
		return false, "ContentDispositionFilename"
	}
	if want.TransferEncoding != "" && !strings.EqualFold(want.TransferEncoding, have.TransferEncoding) {
		return false, "TransferEncoding"
	}
	if want.BodyContains != "" && !strings.Contains(strings.TrimSpace(have.BodyContains), strings.TrimSpace(want.BodyContains)) {
		return false, "BodyContains"
	}
	if want.BodyIs != "" && strings.TrimSpace(have.BodyIs) != strings.TrimSpace(want.BodyIs) {
		return false, "BodyIs"
	}

	for _, wantSection := range want.Sections {
		didPass := false
		for _, haveSection := range have.Sections {
			ok, _ := matchContentRecursive(haveSection, wantSection)
			if ok {
				didPass = true
				break
			}
		}
		if !didPass {
			return false, "recursive mismatch found"
		}
	}

	return true, ""
}

type Mailbox struct {
	Name   string `bdd:"name"`
	Total  int    `bdd:"total"`
	Unread int    `bdd:"unread"`
}

func newMailboxFromIMAP(status *imap.MailboxStatus) Mailbox {
	return Mailbox{
		Name:   status.Name,
		Total:  int(status.Messages),
		Unread: int(status.Unseen),
	}
}

func matchMailboxes(have, want []Mailbox) error {
	slices.SortFunc(have, func(a, b Mailbox) bool {
		return a.Name < b.Name
	})

	slices.SortFunc(want, func(a, b Mailbox) bool {
		return a.Name < b.Name
	})

	if !IsSub(want, have) {
		return fmt.Errorf("missing mailboxes: %v", want)
	}

	return nil
}

func eventually(condition func() error) error {
	ch := make(chan error, 1)
	var lastErr error

	var timerDuration = 30 * time.Second
	// Extend to 5min for live API.
	if hostURL := os.Getenv("FEATURE_TEST_HOST_URL"); hostURL != "" {
		timerDuration = 300 * time.Second
	}

	timer := time.NewTimer(timerDuration)
	defer timer.Stop()

	ticker := time.NewTicker(timerDuration / 300)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			return fmt.Errorf("eventually timed out: %w", lastErr)

		case <-tick:
			tick = nil

			go func() { ch <- condition() }()

		case err := <-ch:
			if err == nil {
				return nil
			}

			lastErr = err
			tick = ticker.C
		}
	}
}

func unmarshalTable[T any](table *messages.PickleTable) ([]T, error) {
	if len(table.Rows) == 0 {
		return nil, fmt.Errorf("empty table")
	}

	res := make([]T, 0, len(table.Rows))

	for _, row := range table.Rows[1:] {
		var v T

		if err := unmarshalRow(table.Rows[0], row, &v); err != nil {
			return nil, err
		}

		res = append(res, v)
	}

	return res, nil
}

func unmarshalRow(header, row *messages.PickleTableRow, v any) error {
	typ := reflect.TypeOf(v).Elem()

	for idx := 0; idx < typ.NumField(); idx++ {
		field := typ.Field(idx)

		if tag, ok := field.Tag.Lookup("bdd"); ok {
			cell, ok := getCellValue(header, row, tag)
			if !ok {
				continue
			}

			switch field.Type.Kind() { //nolint:exhaustive
			case reflect.String:
				reflect.ValueOf(v).Elem().Field(idx).SetString(cell)

			case reflect.Int:
				reflect.ValueOf(v).Elem().Field(idx).SetInt(int64(mustParseInt(cell)))

			case reflect.Bool:
				reflect.ValueOf(v).Elem().Field(idx).SetBool(mustParseBool(cell))

			default:
				return fmt.Errorf("unsupported type %q", field.Type.Kind())
			}
		}
	}

	return nil
}

func getCellValue(header, row *messages.PickleTableRow, name string) (string, bool) {
	for idx, cell := range header.Cells {
		if cell.Value == name {
			return row.Cells[idx].Value, true
		}
	}

	return "", false
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func mustParseBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		panic(err)
	}

	return v
}

type Contact struct {
	Name    string `bdd:"name"`
	Email   string `bdd:"email"`
	Format  string `bdd:"format"`
	Scheme  string `bdd:"scheme"`
	Sign    string `bdd:"signature"`
	Encrypt string `bdd:"encryption"`
}

type MailSettings struct {
	DraftMIMEType   rfc822.MIMEType             `bdd:"DraftMIMEType"`
	AttachPublicKey proton.Bool                 `bdd:"AttachPublicKey"`
	Sign            proton.SignExternalMessages `bdd:"Sign"`
	PGPScheme       proton.EncryptionScheme     `bdd:"PGPScheme"`
}
