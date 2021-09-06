// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package tests

import (
	"bytes"
	"fmt"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-mbox"
)

func TransferSetupFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^there are EML files$`, thereAreEMLFiles)
	s.Step(`^there is EML file "([^"]*)"$`, thereIsEMLFile)
	s.Step(`^there is MBOX file "([^"]*)" with messages$`, thereIsMBOXFileWithMessages)
	s.Step(`^there is MBOX file "([^"]*)"$`, thereIsMBOXFile)
	s.Step(`^there are IMAP mailboxes$`, thereAreIMAPMailboxes)
	s.Step(`^there are IMAP messages$`, thereAreIMAPMessages)
	s.Step(`^there is IMAP message in mailbox "([^"]*)" with seq (\d+), uid (\d+), time "([^"]*)" and subject "([^"]*)"$`, thereIsIMAPMessage)
	s.Step(`^there is skip encrypted messages set to "([^"]*)"$`, thereIsSkipEncryptedMessagesSetTo)
}

func thereAreEMLFiles(messages *godog.Table) error {
	head := messages.Rows[0].Cells
	for _, row := range messages.Rows[1:] {
		fileName := ""
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "file":
				fileName = cell.Value
			case "from", "to", "subject", "time", "body":
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		body := getBodyFromDataRow(head, row)
		if err := createFile(fileName, body); err != nil {
			return err
		}
	}
	return nil
}

func thereIsEMLFile(fileName string, message *godog.DocString) error {
	return createFile(fileName, message.Content)
}

func thereIsMBOXFileWithMessages(fileName string, messages *godog.Table) error {
	mboxBuffer := &bytes.Buffer{}
	mboxWriter := mbox.NewWriter(mboxBuffer)

	head := messages.Rows[0].Cells
	for _, row := range messages.Rows[1:] {
		from := ""
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "from":
				from = cell.Value
			case "to", "subject", "time", "body":
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		body := getBodyFromDataRow(head, row)

		messageWriter, err := mboxWriter.CreateMessage(from, time.Now())
		if err != nil {
			return err
		}
		_, err = messageWriter.Write([]byte(body))
		if err != nil {
			return err
		}
	}

	return createFile(fileName, mboxBuffer.String())
}

func thereIsMBOXFile(fileName string, messages *godog.DocString) error {
	return createFile(fileName, messages.Content)
}

func thereAreIMAPMailboxes(mailboxes *godog.Table) error {
	imapServer := ctx.GetTransferRemoteIMAPServer()
	head := mailboxes.Rows[0].Cells
	for _, row := range mailboxes.Rows[1:] {
		mailboxName := ""
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "name":
				mailboxName = cell.Value
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}
		imapServer.AddMailbox(mailboxName)
	}
	return nil
}

func thereAreIMAPMessages(messages *godog.Table) (err error) {
	imapServer := ctx.GetTransferRemoteIMAPServer()
	head := messages.Rows[0].Cells
	for _, row := range messages.Rows[1:] {
		mailboxName := ""
		date := time.Now()
		subject := ""
		seqNum := 0
		uid := 0
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "mailbox":
				mailboxName = cell.Value
			case "uid":
				uid, err = strconv.Atoi(cell.Value)
				if err != nil {
					return internalError(err, "failed to parse uid")
				}
			case "seqnum":
				seqNum, err = strconv.Atoi(cell.Value)
				if err != nil {
					return internalError(err, "failed to parse seqnum")
				}
			case "time":
				date, err = time.Parse(timeFormat, cell.Value)
				if err != nil {
					return internalError(err, "failed to parse time")
				}
			case "subject":
				subject = cell.Value
			case "from", "to", "body":
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		body := getBodyFromDataRow(head, row)
		imapMessage, err := getIMAPMessage(seqNum, uid, date, subject, body)
		if err != nil {
			return err
		}
		imapServer.AddMessage(mailboxName, imapMessage)
	}
	return nil
}

func thereIsIMAPMessage(mailboxName string, seqNum, uid int, dateValue, subject string, message *godog.DocString) error {
	imapServer := ctx.GetTransferRemoteIMAPServer()

	date, err := time.Parse(timeFormat, dateValue)
	if err != nil {
		return internalError(err, "failed to parse time")
	}

	imapMessage, err := getIMAPMessage(seqNum, uid, date, subject, message.Content)
	if err != nil {
		return err
	}
	imapServer.AddMessage(mailboxName, imapMessage)

	return nil
}

func getBodyFromDataRow(head []*messages.PickleTableCell, row *messages.PickleTableRow) string {
	body := "hello"
	headers := textproto.MIMEHeader{}
	headers.Set("Received", "by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000")
	for n, cell := range row.Cells {
		switch head[n].Value {
		case "from":
			headers.Set("from", cell.Value)
		case "to":
			headers.Set("to", cell.Value)
		case "subject":
			headers.Set("subject", cell.Value)
		case "time":
			date, err := time.Parse(timeFormat, cell.Value)
			if err != nil {
				panic(err)
			}
			headers.Set("date", date.Format(time.RFC1123))
		case "body":
			body = cell.Value
		}
	}

	buffer := &bytes.Buffer{}
	_ = message.WriteHeader(buffer, headers)
	return buffer.String() + body + "\n\n"
}

func getIMAPMessage(seqNum, uid int, date time.Time, subject, body string) (*imap.Message, error) {
	reader := bytes.NewBufferString(body)
	bodyStructure, err := message.NewBodyStructure(reader)
	if err != nil {
		return nil, internalError(err, "failed to parse body structure")
	}
	imapBodyStructure, err := bodyStructure.IMAPBodyStructure([]int{})
	if err != nil {
		return nil, internalError(err, "failed to parse body structure")
	}
	bodySection, _ := imap.ParseBodySectionName("BODY[]")

	return &imap.Message{
		SeqNum: uint32(seqNum),
		Uid:    uint32(uid),
		Size:   uint32(len(body)),
		Envelope: &imap.Envelope{
			Date:    date,
			Subject: subject,
		},
		BodyStructure: imapBodyStructure,
		Body: map[*imap.BodySectionName]imap.Literal{
			bodySection: bytes.NewBufferString(body),
		},
	}, nil
}

func createFile(fileName, body string) error {
	root := ctx.GetTransferLocalRootForImport()
	filePath := filepath.Join(root, fileName)

	dirPath := filepath.Dir(filePath)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return internalError(err, "failed to create dir")
	}

	f, err := os.Create(filePath)
	if err != nil {
		return internalError(err, "failed to create file")
	}
	defer f.Close() //nolint

	_, err = f.WriteString(body)
	return internalError(err, "failed to write to file")
}

func thereIsSkipEncryptedMessagesSetTo(value string) error {
	switch value {
	case "true":
		ctx.SetTransferSkipEncryptedMessages(true)
	case "false":
		ctx.SetTransferSkipEncryptedMessages(false)
	default:
		return fmt.Errorf("expected either true or false, was %v", value)
	}
	return nil
}
