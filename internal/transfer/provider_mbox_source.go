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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/emersion/go-mbox"
	"github.com/pkg/errors"
)

// TransferTo exports messages based on rules to channel.
func (p *MBOXProvider) TransferTo(rules transferRules, progress *Progress, ch chan<- Message) {
	log.Info("Started transfer from MBOX to channel")
	defer log.Info("Finished transfer from MBOX to channel")

	filePathsPerFolder, err := p.getFilePathsPerFolder(rules)
	if err != nil {
		progress.fatal(err)
		return
	}

	for folderName, filePaths := range filePathsPerFolder {
		// No error guaranteed by getFilePathsPerFolder.
		rule, _ := rules.getRuleBySourceMailboxName(folderName)
		for _, filePath := range filePaths {
			if progress.shouldStop() {
				break
			}
			p.updateCount(rule, progress, filePath)
		}
	}

	for folderName, filePaths := range filePathsPerFolder {
		// No error guaranteed by getFilePathsPerFolder.
		rule, _ := rules.getRuleBySourceMailboxName(folderName)
		log.WithField("rule", rule).Debug("Processing rule")
		for _, filePath := range filePaths {
			if progress.shouldStop() {
				break
			}
			p.transferTo(rule, progress, ch, filePath)
		}
	}
}

func (p *MBOXProvider) getFilePathsPerFolder(rules transferRules) (map[string][]string, error) {
	filePaths, err := getFilePathsWithSuffix(p.root, ".mbox")
	if err != nil {
		return nil, err
	}

	filePathsMap := map[string][]string{}
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		folder := strings.TrimSuffix(fileName, ".mbox")
		_, err := rules.getRuleBySourceMailboxName(folder)
		if err != nil {
			log.WithField("msg", filePath).Trace("Mailbox skipped due to folder name")
			continue
		}

		filePathsMap[folder] = append(filePathsMap[folder], filePath)
	}
	return filePathsMap, nil
}

func (p *MBOXProvider) updateCount(rule *Rule, progress *Progress, filePath string) {
	mboxReader := p.openMbox(progress, filePath)
	if mboxReader == nil {
		return
	}

	count := 0
	for {
		_, err := mboxReader.NextMessage()
		if err != nil {
			break
		}
		count++
	}
	progress.updateCount(rule.SourceMailbox.Name, uint(count))
}

func (p *MBOXProvider) transferTo(rule *Rule, progress *Progress, ch chan<- Message, filePath string) {
	mboxReader := p.openMbox(progress, filePath)
	if mboxReader == nil {
		return
	}

	index := 0
	count := 0
	for {
		if progress.shouldStop() {
			break
		}

		index++
		id := fmt.Sprintf("%s:%d", filePath, index)

		msgReader, err := mboxReader.NextMessage()
		if err == io.EOF {
			break
		} else if err != nil {
			progress.fatal(err)
			break
		}

		msg, err := p.exportMessage(rule, id, msgReader)

		// Read and check time in body only if the rule specifies it
		// to not waste energy.
		if err == nil && rule.HasTimeLimit() {
			msgTime, msgTimeErr := getMessageTime(msg.Body)
			if msgTimeErr != nil {
				err = msgTimeErr
			} else if !rule.isTimeInRange(msgTime) {
				log.WithField("msg", id).Debug("Message skipped due to time")
				continue
			}
		}

		// Counting only messages filtered by time to update count to correct total.
		count++

		// addMessage is called after time check to not report message
		// which should not be exported but any error from reading body
		// or parsing time is reported as an error.
		progress.addMessage(id, rule)
		progress.messageExported(id, msg.Body, err)
		if err == nil {
			ch <- msg
		}
	}
	progress.updateCount(rule.SourceMailbox.Name, uint(count))
}

func (p *MBOXProvider) exportMessage(rule *Rule, id string, msgReader io.Reader) (Message, error) {
	body, err := ioutil.ReadAll(msgReader)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to read message")
	}

	return Message{
		ID:      id,
		Unread:  false,
		Body:    body,
		Source:  rule.SourceMailbox,
		Targets: rule.TargetMailboxes,
	}, nil
}

func (p *MBOXProvider) openMbox(progress *Progress, mboxPath string) *mbox.Reader {
	mboxPath = filepath.Join(p.root, mboxPath)
	mboxFile, err := os.Open(mboxPath) //nolint[gosec]
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		progress.fatal(err)
		return nil
	}
	return mbox.NewReader(mboxFile)
}
