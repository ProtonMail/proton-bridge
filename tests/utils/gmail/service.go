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
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/ProtonMail/proton-bridge/v3/tests/utils/gmail/tokenservice"
	"github.com/cucumber/godog"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const DEBUG = false
const GmailUserID = "me"

func getGmailService() *gmail.Service {
	ctx := context.Background()

	gmailClient, err := tokenservice.LoadGmailClient(ctx)
	if err != nil {
		log.Fatalf("unable to retrieve gmail http client: %v", err)
	}

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(gmailClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	return gmailService
}

func ExternalSendEmail(from, to string, message *godog.DocString) error {
	srv := getGmailService()

	var msg gmail.Message

	msgStr := []byte(
		"From: " + from + " \n" +
			"To: " + to + " \n" +
			message.Content)

	msg.Raw = base64.URLEncoding.EncodeToString(msgStr)

	_, err := srv.Users.Messages.Send(GmailUserID, &msg).Do()
	if err != nil {
		return err
	}

	return nil
}

func FetchMessageBySubjectAndSender(subject, sender, state string) (*gmail.Message, error) {
	srv := getGmailService()

	var q string
	switch state {
	case "read":
		q = fmt.Sprintf("(is:read in:inbox OR in:spam) subject:%q from:%q newer:1", subject, sender)
	case "unread":
		q = fmt.Sprintf("(is:unread in:inbox OR in:spam) subject:%q from:%q newer:1", subject, sender)
	default:
		return nil, fmt.Errorf("invalid state argument, must be 'read' or 'unread'")
	}

	if DEBUG {
		fmt.Println("Gmail API Query:", q)
	}

	r, err := srv.Users.Messages.List(GmailUserID).Q(q).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve %s messages with subject: %q and sender: %q with error: %v", state, subject, sender, err)
	}

	if len(r.Messages) == 0 {
		return nil, fmt.Errorf("no %s messages found with subject: %q and sender: %q", state, subject, sender)
	}

	newestMessageID := r.Messages[0].Id
	newestMessage, err := srv.Users.Messages.Get(GmailUserID, newestMessageID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve details of the newest message: %v", err)
	}

	if DEBUG {
		fmt.Println("Email Subject:", getEmailHeader(newestMessage, "Subject"))
		fmt.Println("Email Sender:", getEmailHeader(newestMessage, "From"))
	}

	return newestMessage, nil
}

func getEmailHeader(message *gmail.Message, headerName string) string {
	if message != nil && message.Payload != nil && message.Payload.Headers != nil {
		for _, header := range message.Payload.Headers {
			if header.Name == headerName {
				return header.Value
			}
		}
	}
	return "Header not found"
}

func GetRawMessage(message *gmail.Message) (string, error) {
	srv := getGmailService()
	msg, err := srv.Users.Messages.Get(GmailUserID, message.Id).Format("raw").Do()
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(msg.Raw)

	return string(decodedMsg), err
}

func DeleteAllMessages() {
	srv := getGmailService()

	labels := []string{"INBOX", "SENT", "DRAFT", "SPAM", "TRASH"}

	for _, label := range labels {
		msgs, err := srv.Users.Messages.List(GmailUserID).LabelIds(label).Do()
		if err != nil {
			continue
		}

		for _, m := range msgs.Messages {
			_ = srv.Users.Messages.Delete(GmailUserID, m.Id).Do()
		}
	}
}
