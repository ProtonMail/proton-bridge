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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// TransferTo exports messages based on rules to channel.
func (p *EMLProvider) TransferTo(rules transferRules, progress *Progress, ch chan<- Message) {
	log.Info("Started transfer from EML to channel")
	defer log.Info("Finished transfer from EML to channel")

	filePathsPerFolder, err := p.getFilePathsPerFolder(rules)
	if err != nil {
		progress.fatal(err)
		return
	}

	if len(filePathsPerFolder) == 0 {
		return
	}

	// This list is not filtered by time but instead going throgh each file
	// twice or keeping all in memory we will tell rough estimation which
	// will be updated during processing each file.
	for folderName, filePaths := range filePathsPerFolder {
		if progress.shouldStop() {
			break
		}

		progress.updateCount(folderName, uint(len(filePaths)))
	}
	progress.countsFinal()

	for folderName, filePaths := range filePathsPerFolder {
		// No error guaranteed by getFilePathsPerFolder.
		rule, _ := rules.getRuleBySourceMailboxName(folderName)
		log.WithField("rule", rule).Debug("Processing rule")
		p.exportMessages(rule, filePaths, progress, ch)
	}
}

func (p *EMLProvider) getFilePathsPerFolder(rules transferRules) (map[string][]string, error) {
	filePaths, err := getFilePathsWithSuffix(p.root, ".eml")
	if err != nil {
		return nil, err
	}

	filePathsMap := map[string][]string{}
	for _, filePath := range filePaths {
		folder := filepath.Base(filepath.Dir(filepath.Join(p.root, filePath)))
		_, err := rules.getRuleBySourceMailboxName(folder)
		if err != nil {
			log.WithField("msg", filePath).Trace("Message skipped due to folder name")
			continue
		}

		filePathsMap[folder] = append(filePathsMap[folder], filePath)
	}

	return filePathsMap, nil
}

func (p *EMLProvider) exportMessages(rule *Rule, filePaths []string, progress *Progress, ch chan<- Message) {
	for _, filePath := range filePaths {
		if progress.shouldStop() {
			break
		}

		msg, err := p.exportMessage(rule, filePath)

		progress.addMessage(filePath, msg.sourceNames(), msg.targetNames())

		// Read and check time in body only if the rule specifies it
		// to not waste energy.
		if err == nil && rule.HasTimeLimit() {
			msgTime, msgTimeErr := getMessageTime(msg.Body)
			if msgTimeErr != nil {
				err = msgTimeErr
			} else if !rule.isTimeInRange(msgTime) {
				log.WithField("msg", filePath).Debug("Message skipped due to time")
				progress.messageSkipped(filePath)
				continue
			}
		}

		progress.messageExported(filePath, msg.Body, err)
		if err == nil {
			ch <- msg
		}
	}
}

func (p *EMLProvider) exportMessage(rule *Rule, filePath string) (Message, error) {
	fullFilePath := filepath.Clean(filepath.Join(p.root, filePath))
	file, err := os.Open(fullFilePath) //nolint[gosec]
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to open message")
	}
	defer file.Close() //nolint[errcheck]

	body, err := ioutil.ReadAll(file)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to read message")
	}

	return Message{
		ID:      filePath,
		Unread:  false,
		Body:    body,
		Sources: []Mailbox{rule.SourceMailbox},
		Targets: rule.TargetMailboxes,
	}, nil
}
