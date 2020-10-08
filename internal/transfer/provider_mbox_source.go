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

	filePathsPerFolder, err := p.getFilePathsPerFolder()
	if err != nil {
		progress.fatal(err)
		return
	}

	if len(filePathsPerFolder) == 0 {
		return
	}

	for folderName, filePaths := range filePathsPerFolder {
		log.WithField("folder", folderName).Debug("Estimating folder counts")
		for _, filePath := range filePaths {
			if progress.shouldStop() {
				break
			}
			p.updateCount(progress, filePath)
		}
	}
	progress.countsFinal()

	for folderName, filePaths := range filePathsPerFolder {
		log.WithField("folder", folderName).Debug("Processing folder")
		for _, filePath := range filePaths {
			if progress.shouldStop() {
				break
			}
			p.transferTo(rules, progress, ch, folderName, filePath)
		}
	}
}

func (p *MBOXProvider) getFilePathsPerFolder() (map[string][]string, error) {
	filePaths, err := getFilePathsWithSuffix(p.root, ".mbox")
	if err != nil {
		return nil, err
	}

	filePathsMap := map[string][]string{}
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		folder := strings.TrimSuffix(fileName, ".mbox")
		filePathsMap[folder] = append(filePathsMap[folder], filePath)
	}
	return filePathsMap, nil
}

func (p *MBOXProvider) updateCount(progress *Progress, filePath string) {
	mboxReader := p.openMbox(progress, filePath)
	if mboxReader == nil {
		return
	}

	count := 0
	for {
		_, err := mboxReader.NextMessage()
		if err == io.EOF {
			break
		} else if err != nil {
			progress.fatal(err)
			break
		}
		count++
	}
	progress.updateCount(filePath, uint(count))
}

func (p *MBOXProvider) transferTo(rules transferRules, progress *Progress, ch chan<- Message, folderName, filePath string) {
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

		msg, err := p.exportMessage(rules, folderName, id, msgReader)

		if err == nil && len(msg.Targets) == 0 {
			// Here should be called progress.messageSkipped(id) once we have
			// this feature, and following progress.updateCount can be removed.
			continue
		}

		count++

		// addMessage is called after time check to not report message
		// which should not be exported but any error from reading body
		// or parsing time is reported as an error.
		progress.addMessage(id, msg.sourceNames(), msg.targetNames())
		progress.messageExported(id, msg.Body, err)
		if err == nil {
			ch <- msg
		}
	}
	progress.updateCount(filePath, uint(count))
}

func (p *MBOXProvider) exportMessage(rules transferRules, folderName, id string, msgReader io.Reader) (Message, error) {
	body, err := ioutil.ReadAll(msgReader)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to read message")
	}

	msgRules := p.getMessageRules(rules, folderName, id, body)
	sources := p.getMessageSources(msgRules)
	targets := p.getMessageTargets(msgRules, id, body)
	return Message{
		ID:      id,
		Unread:  false,
		Body:    body,
		Sources: sources,
		Targets: targets,
	}, nil
}

func (p *MBOXProvider) getMessageRules(rules transferRules, folderName, id string, body []byte) []*Rule {
	msgRules := []*Rule{}

	folderRule, err := rules.getRuleBySourceMailboxName(folderName)
	if err != nil {
		log.WithField("msg", id).WithField("source", folderName).Debug("Message skipped due to source")
	} else {
		msgRules = append(msgRules, folderRule)
	}

	gmailLabels, err := getGmailLabelsFromMessage(body)
	if err != nil {
		log.WithError(err).Error("Failed to get gmail labels, ")
	} else {
		for _, label := range gmailLabels {
			rule, err := rules.getRuleBySourceMailboxName(label)
			if err != nil {
				log.WithField("msg", id).WithField("source", label).Debug("Message skipped due to source")
				continue
			}
			msgRules = append(msgRules, rule)
		}
	}

	return msgRules
}

func (p *MBOXProvider) getMessageSources(msgRules []*Rule) []Mailbox {
	sources := []Mailbox{}
	for _, rule := range msgRules {
		sources = append(sources, rule.SourceMailbox)
	}
	return sources
}

func (p *MBOXProvider) getMessageTargets(msgRules []*Rule, id string, body []byte) []Mailbox {
	targets := []Mailbox{}
	haveExclusiveMailbox := false
	for _, rule := range msgRules {
		// Read and check time in body only if the rule specifies it
		// to not waste energy.
		if rule.HasTimeLimit() {
			msgTime, err := getMessageTime(body)
			if err != nil {
				log.WithError(err).Error("Failed to parse time, time check skipped")
			} else if !rule.isTimeInRange(msgTime) {
				log.WithField("msg", id).WithField("source", rule.SourceMailbox.Name).Debug("Message skipped due to time")
				continue
			}
		}
		for _, newTarget := range rule.TargetMailboxes {
			// msgRules is sorted. The first rule is based on the folder name,
			// followed by the order from X-Gmail-Labels. The rule based on
			// the folder name should have priority for exclusive target.
			if newTarget.IsExclusive && haveExclusiveMailbox {
				continue
			}
			found := false
			for _, target := range targets {
				if target.Hash() == newTarget.Hash() {
					found = true
					break
				}
			}
			if found {
				continue
			}
			if newTarget.IsExclusive {
				haveExclusiveMailbox = true
			}
			targets = append(targets, newTarget)
		}
	}
	return targets
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
