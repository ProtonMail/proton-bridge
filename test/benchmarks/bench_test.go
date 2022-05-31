// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package benchmarks

import (
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/test/context"
	"github.com/ProtonMail/proton-bridge/v2/test/mocks"
)

func benchTestContext() (*context.TestContext, *mocks.IMAPClient) {
	ctx := context.New()

	username := "user"
	account := ctx.GetTestAccount(username)
	if account == nil {
		panic("account " + username + " does not exist")
	}

	_ = ctx.GetPMAPIController().AddUser(account)
	if err := ctx.LoginUser(account.Username(), account.Password(), account.MailboxPassword()); err != nil {
		panic(err)
	}

	imapClient := ctx.GetIMAPClient("client")
	imapClient.Login(account.Address(), account.BridgePassword())

	// waitForSync between bridge and API. There is no way to know precisely
	// from the outside when the bridge is synced. We could wait for first
	// response from any fetch, but we don't know how many messages should be
	// there. Unless we hard code the number of messages.
	// Please, check this time is enough when doing benchmarks and don't forget
	// to exclude this time from total time.
	time.Sleep(10 * time.Second)

	return ctx, imapClient
}

func BenchmarkIMAPFetch(b *testing.B) {
	tc, c := benchTestContext()
	defer tc.Cleanup()

	c.Select("All Mail").AssertOK()

	fetchBench := []struct{ ids, args string }{
		{"1:10", "rfc822.size"},
		{"1:100", "rfc822.size"},
		{"1:1000", "rfc822.size"},
		{"1:*", "rfc822.size"},
	}

	for _, bd := range fetchBench {
		ids, args := bd.ids, bd.args // pin
		b.Run(ids+"-"+args, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.Fetch(ids, args)
			}
		})
	}
}

func BenchmarkCachingFetch(b *testing.B) {
	tc, c := benchTestContext()
	defer tc.Cleanup()

	c.Select("\"All Mail\"").AssertOK()

	ids := "1:100"
	args := "body.peek[]"
	tries := []string{"long", "short"}

	for _, try := range tries {
		b.Run(strings.Join([]string{ids, args, try}, "-"), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.Fetch(ids, args)
			}
		})
	}
}

func BenchmarkIMAPAppleMail(b *testing.B) {
	tc, c := benchTestContext()
	defer tc.Cleanup()

	// assume we have at least 50 messages in INBOX
	idRange := "1:50"
	newUID := "50" // assume that Apple mail don't know about this mail

	// I will use raw send command to completely reproduce the calls
	// (including quotation and case sensitivity)
	b.Run("default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, command := range []string{
				"CAPABILITY",
				"ID (" +
					`"name" "Mac OS X Mail" ` +
					`"version" "11.5 (3445.9.1)" ` +
					`"os" "Mac OS X" ` +
					`"os-version" "10.13.6 (17G3025)" ` +
					`"vendor" "Apple Inc."` +
					")",
				`LIST "" ""`,
				`STATUS INBOX (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`,
				`SELECT INBOX`,
				`FETCH ` + idRange + ` (FLAGS UID)`,
				`FETCH ` + idRange + " " +
					`(` +
					`INTERNALDATE UID RFC822.SIZE FLAGS ` +
					`BODY.PEEK[` +
					`HEADER.FIELDS (` +
					`date subject from to cc message-id in-reply-to references ` +
					`x-priority x-uniform-type-identifier x-universally-unique-identifier ` +
					`list-id list-unsubscribe` +
					`)])`,
				`UID FETCH ` + newUID + ` (BODYSTRUCTURE BODY.PEEK[HEADER])`,
				// if email has attachment it is splitted to several fetches
				//   `UID FETCH 133 (BODY.PEEK[3]<0.5877469> BODY.PEEK[1] BODY.PEEK[2])`,
				//   `UID FETCH 133 BODY.PEEK[3]<5877469.2925661>`,
				// here I will just use section download, which is used by AppleMail
				`UID FETCH ` + newUID + ` BODY.PEEK[1]`,
				// here I will just use partial download, which is used by AppleMail
				`UID FETCH ` + newUID + ` BODY.PEEK[]<0.2000>`,
			} {
				c.SendCommand(command).AssertOK()
			}
		}
	})
}

func BenchmarkIMAPOutlook(b *testing.B) {
	tc, c := benchTestContext()
	defer tc.Cleanup()

	// assume we have at least 50 messages in INBOX
	idRange := "1:50"

	// I will use raw send command to completely reproduce the calls
	// (including quotation and case sensitivity)
	b.Run("default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, command := range []string{
				/*
					"ID ("+
						`"name" "Mac OS X Mail" `+
						`"version" "11.5 (3445.9.1)" `+
						`"os" "Mac OS X" `+
						`"os-version" "10.13.6 (17G3025)" `+
						`"vendor" "Apple Inc."`+
						")",
				*/

				`SELECT "INBOX"`,
				`UID SEARCH ` + idRange + ` SINCE 01-Sep-2019`,
				`UID FETCH 1:* (UID FLAGS)`,
				`UID FETCH ` + idRange + ` (UID FLAGS RFC822.SIZE BODY.PEEK[] INTERNALDATE)`,
			} {
				c.SendCommand(command).AssertOK()
			}
		}
	})
}

func BenchmarkIMAPThunderbird(b *testing.B) {
	tc, c := benchTestContext()
	defer tc.Cleanup()

	// I will use raw send command to completely reproduce the calls
	// (including quotation and case sensitivity)
	b.Run("default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, command := range []string{
				`capability`,
				`ID ("name" "Thunderbird" "version" "68.2.0")`,
				`select "INBOX"`,
				`getquotaroot "INBOX"`,
				`UID fetch 1:* (FLAGS)`,
			} {
				c.SendCommand(command).AssertOK()
			}
		}
	})
}
