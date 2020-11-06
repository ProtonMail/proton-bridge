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

package transfer

import (
	"fmt"
	"testing"

	"github.com/emersion/go-imap"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	r "github.com/stretchr/testify/require"
)

func newTestIMAPProvider(t *testing.T, m mocks) *IMAPProvider {
	m.imapClientProvider.EXPECT().State().Return(imap.ConnectedState).AnyTimes()
	m.imapClientProvider.EXPECT().Capability().Return(map[string]bool{
		"AUTH": true,
	}, nil).AnyTimes()

	dialer := func(string) (IMAPClientProvider, error) {
		return m.imapClientProvider, nil
	}
	provider, err := newIMAPProvider(dialer, "user", "pass", "host", "42")
	r.NoError(t, err)
	return provider
}

func TestProviderIMAPLoadMessagesInfo(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	provider := newTestIMAPProvider(t, m)

	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	rule := &Rule{SourceMailbox: Mailbox{Name: "Mailbox"}}
	uidValidity := 1
	count := 2200
	failingIndex := 2100

	m.imapClientProvider.EXPECT().Select(rule.SourceMailbox.Name, gomock.Any()).Return(&imap.MailboxStatus{}, nil).AnyTimes()
	m.imapClientProvider.EXPECT().
		Fetch(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(seqSet *imap.SeqSet, items []imap.FetchItem, ch chan *imap.Message) error {
			defer close(ch)
			for _, seq := range seqSet.Set {
				for i := seq.Start; i <= seq.Stop; i++ {
					if int(i) == failingIndex {
						return errors.New("internal server error")
					}
					ch <- &imap.Message{
						SeqNum: i,
						Uid:    i * 10,
						Size:   i * 100,
					}
				}
			}
			return nil
		}).
		// 2200 messages is split into two batches (2000 and 200),
		// the second one fails and makes 200 calls (one-by-one).
		// Plus two failed requests are repeated `imapRetries` times.
		Times(2 + 200 + (2 * (imapRetries - 1)))

	messageInfo := provider.loadMessagesInfo(rule, &progress, uint32(uidValidity), uint32(count))

	r.Equal(t, count-1, len(messageInfo)) // One message produces internal server error.
	for index := 1; index <= count; index++ {
		uid := index * 10
		key := fmt.Sprintf("%s_%d:%d", rule.SourceMailbox.Name, uidValidity, uid)

		if index == failingIndex {
			r.Empty(t, messageInfo[key])
			continue
		}

		r.Equal(t, imapMessageInfo{
			id:   key,
			uid:  uint32(uid),
			size: uint32(index * 100),
		}, messageInfo[key])
	}
}
