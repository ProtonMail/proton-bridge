// Copyright (c) 2024 Proton AG
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
	"encoding/json"
	"fmt"
	"mime"
	"net/mail"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	GmailService "github.com/ProtonMail/proton-bridge/v3/tests/utils/gmail"
	"github.com/cucumber/godog"
)

func (s *scenario) externalClientSendsTheFollowingMessageFromTo(from, to string, message *godog.DocString) error {
	return GmailService.ExternalSendEmail(from, to, message)
}

func (s *scenario) externalClientFetchesTheFollowingMessage(subject, sender, state string) error {
	err := eventually(func() error {
		_, err := GmailService.FetchMessageBySubjectAndSender(subject, sender, state)
		return err
	})
	return err
}

func (s *scenario) externalClientSeesMessageWithStructure(subject, sender, state string, message *godog.DocString) error {
	err := eventually(func() error {
		gmailMessage, err := GmailService.FetchMessageBySubjectAndSender(subject, sender, state)
		if err != nil {
			return err
		}

		var msgStruct MessageStruct
		if err := json.Unmarshal([]byte(message.Content), &msgStruct); err != nil {
			return err
		}

		parsedMessage, err := GmailService.GetRawMessage(gmailMessage)
		if err != nil {
			return err
		}

		var structs []MessageStruct
		messageStruct := parseGmail(parsedMessage)
		structs = append(structs, messageStruct)

		return matchStructureRecursive(structs, msgStruct)
	})
	return err
}

func (s *scenario) externalClientDeletesAllMessages() {
	GmailService.DeleteAllMessages()
}

func parseGmail(rawMsg string) MessageStruct {
	msg, err := mail.ReadMessage(strings.NewReader(rawMsg))
	if err != nil {
		panic(err)
	}

	var dec mime.WordDecoder
	decodedSubject, err := dec.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		decodedSubject = msg.Header.Get("Subject")
	}

	parser, err := parser.New(strings.NewReader(rawMsg))
	if err != nil {
		panic(fmt.Errorf("parser error: %e", err))
	}

	m, err := message.ParseWithParser(parser, true)
	if err != nil {
		panic(fmt.Errorf("parser with parser: %e", err))
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

	// There might be an issue with the dates if we end up using them
	return MessageStruct{
		From:    msg.Header.Get("From"),
		To:      msg.Header.Get("To"),
		CC:      msg.Header.Get("CC"),
		BCC:     msg.Header.Get("BCC"),
		Subject: decodedSubject,
		Date:    msg.Header.Get("Date"),
		Content: parseMessageSection([]byte(strings.TrimSpace(rawMsg)), strings.TrimSpace(body)),
	}
}
