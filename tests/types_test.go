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

package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/messages-go/v16"
	"github.com/emersion/go-imap"
	"golang.org/x/exp/slices"
)

type Message struct {
	Subject     string `bdd:"subject"`
	Body        string `bdd:"body"`
	Attachments string `bdd:"attachments"`

	From string `bdd:"from"`
	To   string `bdd:"to"`
	CC   string `bdd:"cc"`
	BCC  string `bdd:"bcc"`

	Unread bool `bdd:"unread"`
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

	if msg.Body != "" {
		b = append(b, "\r\n"+msg.Body+"\r\n"...)
	}

	return b
}

func newMessageFromIMAP(msg *imap.Message) Message {
	section, err := imap.ParseBodySectionName("BODY[]")
	if err != nil {
		panic(err)
	}

	m, err := message.Parse(msg.GetBody(section))
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
		Attachments: strings.Join(xslices.Map(m.Attachments, func(att message.Attachment) string { return att.Name }), ", "),
		Unread:      !slices.Contains(msg.Flags, imap.SeenFlag),
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

	return message
}

func matchMessages(have, want []Message) error {
	slices.SortFunc(have, func(a, b Message) bool {
		return a.Subject < b.Subject
	})

	slices.SortFunc(want, func(a, b Message) bool {
		return a.Subject < b.Subject
	})

	if !IsSub(ToAny(have), ToAny(want)) {
		return fmt.Errorf("missing messages: %v", want)
	}

	return nil
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
		return fmt.Errorf("missing messages: %v", want)
	}

	return nil
}

func eventually(condition func() error) error {
	ch := make(chan error, 1)

	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			return fmt.Errorf("timed out")

		case <-tick:
			tick = nil

			go func() { ch <- condition() }()

		case err := <-ch:
			if err == nil {
				return nil
			}

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
