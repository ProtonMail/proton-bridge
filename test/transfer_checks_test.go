// Copyright (c) 2020 Proton Technologies AG
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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
	"github.com/emersion/go-mbox"
	"github.com/emersion/go-message"
	"github.com/pkg/errors"
	a "github.com/stretchr/testify/assert"
)

func TransferChecksFeatureContext(s *godog.Suite) {
	s.Step(`^progress result is "([^"]*)"$`, progressFinishedWith)
	s.Step(`^transfer exported (\d+) messages$`, transferExportedNumberOfMessages)
	s.Step(`^transfer imported (\d+) messages$`, transferImportedNumberOfMessages)
	s.Step(`^transfer failed for (\d+) messages$`, transferFailedForNumberOfMessages)
	s.Step(`^transfer exported messages$`, transferExportedMessages)
	s.Step(`^exported messages match the original ones$`, exportedMessagesMatchTheOriginalOnes)
}

func progressFinishedWith(wantResponse string) error {
	progress := ctx.GetTransferProgress()
	// Wait till transport is finished.
	updateCh := progress.GetUpdateChannel()
	if updateCh != nil {
		for range updateCh {
		}
	}

	err := progress.GetFatalError()
	if wantResponse == "OK" {
		a.NoError(ctx.GetTestingT(), err)
	} else {
		a.EqualError(ctx.GetTestingT(), err, wantResponse)
	}
	return ctx.GetTestingError()
}

func transferExportedNumberOfMessages(wantCount int) error {
	progress := ctx.GetTransferProgress()
	_, _, exported, _, _ := progress.GetCounts() //nolint[dogsled]
	a.Equal(ctx.GetTestingT(), uint(wantCount), exported)
	return ctx.GetTestingError()
}

func transferImportedNumberOfMessages(wantCount int) error {
	progress := ctx.GetTransferProgress()
	_, imported, _, _, _ := progress.GetCounts() //nolint[dogsled]
	a.Equal(ctx.GetTestingT(), uint(wantCount), imported)
	return ctx.GetTestingError()
}

func transferFailedForNumberOfMessages(wantCount int) error {
	progress := ctx.GetTransferProgress()
	failedMessages := progress.GetFailedMessages()
	a.Equal(ctx.GetTestingT(), wantCount, len(failedMessages), "failed messages: %v", failedMessages)
	return ctx.GetTestingError()
}

func transferExportedMessages(messages *gherkin.DataTable) error {
	expectedMessages := map[string][]MessageAttributes{}

	head := messages.Rows[0].Cells
	for _, row := range messages.Rows[1:] {
		folder := ""
		msg := MessageAttributes{}

		for n, cell := range row.Cells {
			switch head[n].Value {
			case "folder":
				folder = cell.Value
			case "subject":
				msg.subject = cell.Value
			case "from":
				msg.from = cell.Value
			case "to":
				msg.to = []string{cell.Value}
			case "time":
				date, err := time.Parse(timeFormat, cell.Value)
				if err != nil {
					return internalError(err, "failed to parse time")
				}
				msg.date = date.Unix()
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		expectedMessages[folder] = append(expectedMessages[folder], msg)
		sort.Sort(BySubject(expectedMessages[folder]))
	}

	exportRoot := ctx.GetTransferLocalRootForExport()
	exportedMessages, err := readMessages(exportRoot)
	if err != nil {
		return errors.Wrap(err, "scanning exported messages")
	}

	a.Equal(ctx.GetTestingT(), expectedMessages, exportedMessages)
	return ctx.GetTestingError()
}

func exportedMessagesMatchTheOriginalOnes() error {
	importRoot := ctx.GetTransferLocalRootForImport()
	exportRoot := ctx.GetTransferLocalRootForExport()

	importMessages, err := readMessages(importRoot)
	if err != nil {
		return errors.Wrap(err, "scanning messages for import")
	}
	exportMessages, err := readMessages(exportRoot)
	if err != nil {
		return errors.Wrap(err, "scanning exported messages")
	}
	delete(exportMessages, "All Mail") // Ignore All Mail.

	a.Equal(ctx.GetTestingT(), importMessages, exportMessages)
	return ctx.GetTestingError()
}

func readMessages(root string) (map[string][]MessageAttributes, error) {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	messagesPerLabel := map[string][]MessageAttributes{}
	for _, file := range files {
		if !file.IsDir() {
			fileReader, err := os.Open(filepath.Join(root, file.Name()))
			if err != nil {
				return nil, errors.Wrap(err, "opening file")
			}

			if filepath.Ext(file.Name()) == ".eml" {
				label := filepath.Base(root)
				msg, err := readMessageAttributes(fileReader)
				if err != nil {
					return nil, err
				}
				messagesPerLabel[label] = append(messagesPerLabel[label], msg)
				sort.Sort(BySubject(messagesPerLabel[label]))
			} else if filepath.Ext(file.Name()) == ".mbox" {
				label := strings.TrimSuffix(file.Name(), ".mbox")
				mboxReader := mbox.NewReader(fileReader)
				for {
					msgReader, err := mboxReader.NextMessage()
					if err == io.EOF {
						break
					} else if err != nil {
						return nil, errors.Wrap(err, "reading next message")
					}
					msg, err := readMessageAttributes(msgReader)
					if err != nil {
						return nil, err
					}
					messagesPerLabel[label] = append(messagesPerLabel[label], msg)
				}
				sort.Sort(BySubject(messagesPerLabel[label]))
			}
		} else {
			subfolderRoot := filepath.Join(root, file.Name())
			subfolderMessagesPerLabel, err := readMessages(subfolderRoot)
			if err != nil {
				return nil, err
			}
			for key, value := range subfolderMessagesPerLabel {
				messagesPerLabel[key] = append(messagesPerLabel[key], value...)
				sort.Sort(BySubject(messagesPerLabel[key]))
			}
		}
	}
	return messagesPerLabel, nil
}

type MessageAttributes struct {
	subject string
	from    string
	to      []string
	date    int64
}

func readMessageAttributes(fileReader io.Reader) (MessageAttributes, error) {
	entity, err := message.Read(fileReader)
	if err != nil {
		return MessageAttributes{}, errors.Wrap(err, "reading file")
	}
	date, err := parseTime(entity.Header.Get("date"))
	if err != nil {
		return MessageAttributes{}, errors.Wrap(err, "parsing date")
	}
	from, err := parseAddress(entity.Header.Get("from"))
	if err != nil {
		return MessageAttributes{}, errors.Wrap(err, "parsing from")
	}
	to, err := parseAddresses(entity.Header.Get("to"))
	if err != nil {
		return MessageAttributes{}, errors.Wrap(err, "parsing to")
	}
	return MessageAttributes{
		subject: entity.Header.Get("subject"),
		from:    from,
		to:      to,
		date:    date.Unix(),
	}, nil
}

func parseTime(input string) (time.Time, error) {
	for _, format := range []string{time.RFC1123, time.RFC1123Z} {
		t, err := time.Parse(format, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("Unrecognized time format")
}

func parseAddresses(input string) ([]string, error) {
	addresses, err := rfc5322.ParseAddressList(input)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, address := range addresses {
		result = append(result, address.Address)
	}
	return result, nil
}

func parseAddress(input string) (string, error) {
	address, err := rfc5322.ParseAddressList(input)
	if err != nil {
		return "", err
	}
	return address[0].Address, nil
}

// BySubject implements sort.Interface based on the subject field.
type BySubject []MessageAttributes

func (a BySubject) Len() int           { return len(a) }
func (a BySubject) Less(i, j int) bool { return a[i].subject < a[j].subject }
func (a BySubject) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
