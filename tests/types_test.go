package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Message struct {
	Subject string `bdd:"subject"`

	From string `bdd:"sender"`
	To   string `bdd:"recipient"`
	CC   string `bdd:"cc"`
	BCC  string `bdd:"bcc"`

	Unread bool `bdd:"unread"`
}

func newMessageFromRow(header, row *messages.PickleTableRow) Message {
	var msg Message

	if err := unmarshalRow(header, row, &msg); err != nil {
		panic(err)
	}

	return msg
}

func matchMessages(have []Message, want *godog.Table) error {
	if want := parseMessages(want); !cmp.Equal(want, have, cmpopts.SortSlices(func(a, b Message) bool { return a.Subject < b.Subject })) {
		return fmt.Errorf("want: %v, have: %v", want, have)
	}

	return nil
}

func parseMessages(table *godog.Table) []Message {
	header := table.Rows[0]

	return xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) Message {
		return newMessageFromRow(header, row)
	})
}

type Mailbox struct {
	Name   string `bdd:"name"`
	Total  int    `bdd:"total"`
	Unread int    `bdd:"unread"`
}

func newMailboxFromRow(header, row *messages.PickleTableRow) Mailbox {
	var mbox Mailbox

	if err := unmarshalRow(header, row, &mbox); err != nil {
		panic(err)
	}

	return mbox
}

func matchMailboxes(have []Mailbox, want *godog.Table) error {
	if want := parseMailboxes(want); !cmp.Equal(want, have, cmpopts.SortSlices(func(a, b Mailbox) bool { return a.Name < b.Name })) {
		return fmt.Errorf("want: %v, have: %v", want, have)
	}

	return nil
}

func parseMailboxes(table *godog.Table) []Mailbox {
	header := table.Rows[0]

	return xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) Mailbox {
		return newMailboxFromRow(header, row)
	})
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

func getCellValue(header, row *messages.PickleTableRow, name string) (string, bool) {
	for idx, cell := range header.Cells {
		if cell.Value == name {
			return row.Cells[idx].Value, true
		}
	}

	return "", false
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
