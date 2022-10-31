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

package tests

import (
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/emersion/go-imap"
)

func IMAPChecksFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^IMAP response is "([^"]*)"$`, imapResponseIs)
	s.Step(`^IMAP response to "([^"]*)" is "([^"]*)"$`, imapResponseNamedIs)
	s.Step(`^IMAP response contains "([^"]*)"$`, imapResponseContains)
	s.Step(`^IMAP response to "([^"]*)" contains "([^"]*)"$`, imapResponseNamedContains)
	s.Step(`^IMAP response has (\d+) message(?:s)?$`, imapResponseHasNumberOfMessages)
	s.Step(`^IMAP response to "([^"]*)" has (\d+) message(?:s)?$`, imapResponseNamedHasNumberOfMessages)
	s.Step(`^IMAP response does(?: not|n't) contain "([^"]*)"$`, imapResponseDoesNotContain)
	s.Step(`^IMAP response to "([^"]*)" does not contain "([^"]*)"$`, imapResponseNamedDoesNotContain)
	s.Step(`^IMAP client receives update marking message seq "([^"]*)" as read within (\d+) seconds$`, imapClientReceivesUpdateMarkingMessageSeqAsReadWithin)
	s.Step(`^IMAP client "([^"]*)" receives update marking message seq "([^"]*)" as read within (\d+) seconds$`, imapClientNamedReceivesUpdateMarkingMessageSeqAsReadWithin)
	s.Step(`^IMAP client receives update marking message seq "([^"]*)" as unread within (\d+) seconds$`, imapClientReceivesUpdateMarkingMessageSeqAsUnreadWithin)
	s.Step(`^IMAP client "([^"]*)" receives update marking message seq "([^"]*)" as unread within (\d+) seconds$`, imapClientNamedReceivesUpdateMarkingMessageSeqAsUnreadWithin)
	s.Step(`^IMAP client "([^"]*)" does not receive update for message seq "([^"]*)" within (\d+) seconds$`, imapClientDoesNotReceiveUpdateForMessageSeqWithin)
	s.Step(`^IMAP client is logged out$`, imapClientIsLoggedOut)
	s.Step(`^IMAP client "([^"]*)" is logged out$`, imapClientNamedIsLoggedOut)
}

func imapResponseIs(expectedResponse string) error {
	return imapResponseNamedIs("imap", expectedResponse)
}

func imapResponseNamedIs(clientID, expectedResponse string) error {
	res := ctx.GetIMAPLastResponse(clientID)
	switch {
	case expectedResponse == "OK":
		res.AssertOK()
	case strings.HasPrefix(expectedResponse, "OK"):
		res.AssertResult(expectedResponse)
	default:
		res.AssertError(expectedResponse)
	}
	return ctx.GetTestingError()
}

func imapResponseContains(expectedResponse string) error {
	return imapResponseNamedContains("imap", expectedResponse)
}

func imapResponseNamedContains(clientID, expectedResponse string) error {
	res := ctx.GetIMAPLastResponse(clientID)
	res.AssertSections(expectedResponse)
	return ctx.GetTestingError()
}

func imapResponseHasNumberOfMessages(expectedCount int) error {
	return imapResponseNamedHasNumberOfMessages("imap", expectedCount)
}

func imapResponseNamedHasNumberOfMessages(clientID string, expectedCount int) error {
	res := ctx.GetIMAPLastResponse(clientID)
	res.AssertSectionsCount(expectedCount)
	return ctx.GetTestingError()
}

func imapResponseDoesNotContain(notExpectedResponse string) error {
	return imapResponseNamedDoesNotContain("imap", notExpectedResponse)
}

func imapResponseNamedDoesNotContain(clientID, notExpectedResponse string) error {
	res := ctx.GetIMAPLastResponse(clientID)
	res.AssertNotSections(notExpectedResponse)
	return ctx.GetTestingError()
}

func imapClientReceivesUpdateMarkingMessageSeqAsReadWithin(messageSeq string, seconds int) error {
	return imapClientNamedReceivesUpdateMarkingMessageSeqAsReadWithin("imap", messageSeq, seconds)
}

func imapClientNamedReceivesUpdateMarkingMessageSeqAsReadWithin(clientID, messageSeq string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageSeq, func(messageUID string) {
		regexps = append(regexps, `FETCH \(FLAGS \(.*\\Seen.*\) UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func imapClientReceivesUpdateMarkingMessageSeqAsUnreadWithin(messageSeq string, seconds int) error {
	return imapClientNamedReceivesUpdateMarkingMessageSeqAsUnreadWithin("imap", messageSeq, seconds)
}

func imapClientNamedReceivesUpdateMarkingMessageSeqAsUnreadWithin(clientID, messageSeq string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageSeq, func(messageUID string) {
		// Golang does not support negative look ahead. Following complex regexp checks \Seen is not there.
		regexps = append(regexps, `FETCH \(FLAGS \(([^S]|S[^e]|Se[^e]|See[^n])*\) UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func imapClientDoesNotReceiveUpdateForMessageSeqWithin(clientID, messageSeq string, seconds int) error {
	regexps := []string{}
	iterateOverSeqSet(messageSeq, func(messageUID string) {
		regexps = append(regexps, `FETCH.*UID `+messageUID)
	})
	ctx.GetIMAPLastResponse(clientID).WaitForNotSections(time.Duration(seconds)*time.Second, regexps...)
	return ctx.GetTestingError()
}

func iterateOverSeqSet(seqSet string, callback func(string)) {
	seq, err := imap.ParseSeqSet(seqSet)
	if err != nil {
		panic(err)
	}
	for _, set := range seq.Set {
		for i := set.Start; i <= set.Stop; i++ {
			callback(strconv.Itoa(int(i)))
		}
	}
}

func imapClientIsLoggedOut() error {
	return imapClientNamedIsLoggedOut("imap")
}

func imapClientNamedIsLoggedOut(clientName string) error {
	res := ctx.GetIMAPClient(clientName).SendCommand("CAPABILITY")
	res.AssertError("read response failed:")
	return ctx.GetTestingError()
}
