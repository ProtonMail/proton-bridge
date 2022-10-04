package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/cucumber/messages-go/v16"
	"github.com/emersion/go-imap"
	"golang.org/x/exp/slices"
)

type Message struct {
	Subject string `bdd:"subject"`

	From string `bdd:"from"`
	To   string `bdd:"to"`
	CC   string `bdd:"cc"`
	BCC  string `bdd:"bcc"`

	Unread bool `bdd:"unread"`
}

func newMessageFromIMAP(msg *imap.Message) Message {
	message := Message{
		Subject: msg.Envelope.Subject,
		Unread:  slices.Contains(msg.Flags, imap.SeenFlag),
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

func eventually(condition func() error, waitFor, tick time.Duration) error {
	ch := make(chan error, 1)

	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			return fmt.Errorf("timed out after %v", waitFor)

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

	var res []T

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

			switch field.Type.Kind() {
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
