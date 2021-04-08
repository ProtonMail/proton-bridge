// Copyright (c) 2021 Proton Technologies AG
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
	"time"

	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
)

func getTestMsgBody(subject string) []byte {
	return []byte(fmt.Sprintf(`Subject: %s
From: Bridge Test <bridgetest@pm.test>
To: Bridge Test <bridgetest@protonmail.com>
Content-Type: multipart/mixed; boundary=c672b8d1ef56ed28ab87c3622c5114069bdd3ad7b8f9737498d0c01ecef0967a

--c672b8d1ef56ed28ab87c3622c5114069bdd3ad7b8f9737498d0c01ecef0967a
Content-Disposition: inline
Content-Transfer-Encoding: 7bit
Content-Type: text/plain; charset=utf-8

hello

--c672b8d1ef56ed28ab87c3622c5114069bdd3ad7b8f9737498d0c01ecef0967a--
`, subject))
}

func testTransferTo(t *testing.T, rules transferRules, provider SourceProvider, expectedMessageIDs []string) []Message {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	ch := make(chan Message)
	go func() {
		provider.TransferTo(rules, &progress, ch)
		close(ch)
	}()

	msgs := []Message{}
	gotMessageIDs := []string{}
	for msg := range ch {
		msgs = append(msgs, msg)
		gotMessageIDs = append(gotMessageIDs, msg.ID)
	}
	r.ElementsMatch(t, expectedMessageIDs, gotMessageIDs)

	r.Empty(t, progress.GetFailedMessages())

	return msgs
}

func testTransferFrom(t *testing.T, rules transferRules, provider TargetProvider, messages []Message) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	ch := make(chan Message)
	go func() {
		for _, message := range messages {
			progress.addMessage(message.ID, []string{}, []string{})
			progress.messageExported(message.ID, []byte(""), nil)
			ch <- message
		}
		close(ch)
	}()

	go func() {
		provider.TransferFrom(rules, &progress, ch)
		progress.finish()
	}()

	maxWait := time.Duration(len(messages)*2) * time.Second
	a.Eventually(t, func() bool {
		return progress.updateCh == nil
	}, maxWait, 10*time.Millisecond, "Waiting for imported messages timed out")

	r.Empty(t, progress.GetFailedMessages())
}

func testTransferFromTo(t *testing.T, rules transferRules, source SourceProvider, target TargetProvider, maxWait time.Duration) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	ch := make(chan Message)
	go func() {
		source.TransferTo(rules, &progress, ch)
		close(ch)
	}()
	go func() {
		target.TransferFrom(rules, &progress, ch)
		progress.finish()
	}()

	a.Eventually(t, func() bool {
		return progress.updateCh == nil
	}, maxWait, 10*time.Millisecond, "Waiting for export and import timed out")

	r.Empty(t, progress.GetFailedMessages())
}
