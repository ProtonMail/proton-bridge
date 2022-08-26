package tests

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Message struct {
	Sender    string
	Recipient string
	Subject   string
	Unread    bool
}

func matchMessages(have []Message, want *godog.Table) error {
	if want := parseMessages(want); !cmp.Equal(want, have, cmpopts.SortSlices(func(a, b Message) bool { return a.Subject < b.Subject })) {
		return fmt.Errorf("want: %v, have: %v", want, have)
	}

	return nil
}

func parseMessages(table *godog.Table) []Message {
	return xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) Message {
		return Message{
			Sender:    row.Cells[0].Value,
			Recipient: row.Cells[1].Value,
			Subject:   row.Cells[2].Value,
			Unread:    mustParseBool(row.Cells[3].Value),
		}
	})
}

type Mailbox struct {
	Name   string
	Total  int
	Unread int
}

func matchMailboxes(have []Mailbox, want *godog.Table) error {
	if want := parseMailboxes(want); !cmp.Equal(want, have, cmpopts.SortSlices(func(a, b Mailbox) bool { return a.Name < b.Name })) {
		return fmt.Errorf("want: %v, have: %v", want, have)
	}

	return nil
}

func parseMailboxes(table *godog.Table) []Mailbox {
	mustParseInt := func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}

		return i
	}

	return xslices.Map(table.Rows[1:], func(row *messages.PickleTableRow) Mailbox {
		return Mailbox{
			Name:   row.Cells[0].Value,
			Total:  mustParseInt(row.Cells[1].Value),
			Unread: mustParseInt(row.Cells[2].Value),
		}
	})
}

func mustParseBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		panic(err)
	}

	return v
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
