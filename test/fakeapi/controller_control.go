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

package fakeapi

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

var systemLabelNameToID = map[string]string{ //nolint[gochecknoglobals]
	"INBOX":    pmapi.InboxLabel,
	"Trash":    pmapi.TrashLabel,
	"Spam":     pmapi.SpamLabel,
	"All Mail": pmapi.AllMailLabel,
	"Archive":  pmapi.ArchiveLabel,
	"Sent":     pmapi.SentLabel,
	"Drafts":   pmapi.DraftLabel,
}

func (cntrl *Controller) TurnInternetConnectionOff() {
	cntrl.log.Warn("Turning OFF internet")
	cntrl.noInternetConnection = true
}

func (cntrl *Controller) TurnInternetConnectionOn() {
	cntrl.log.Warn("Turning ON internet")
	cntrl.noInternetConnection = false
}

func (cntrl *Controller) AddUser(user *pmapi.User, addresses *pmapi.AddressList, password string, twoFAEnabled bool) error {
	cntrl.usersByUsername[user.Name] = &fakeUser{
		user:     user,
		password: password,
		has2FA:   twoFAEnabled,
	}
	cntrl.addressesByUsername[user.Name] = addresses
	return nil
}

func (cntrl *Controller) AddUserLabel(username string, label *pmapi.Label) error {
	if _, ok := cntrl.labelsByUsername[username]; !ok {
		cntrl.labelsByUsername[username] = []*pmapi.Label{}
	}

	labelName := getLabelNameWithoutPrefix(label.Name)
	for _, existingLabel := range cntrl.labelsByUsername[username] {
		if existingLabel.Name == labelName {
			return fmt.Errorf("folder or label %s already exists", label.Name)
		}
	}

	label.Exclusive = getLabelExclusive(label.Name)
	prefix := "label"
	if label.Exclusive == 1 {
		prefix = "folder"
	}
	label.ID = cntrl.labelIDGenerator.next(prefix)
	label.Name = labelName
	cntrl.labelsByUsername[username] = append(cntrl.labelsByUsername[username], label)
	cntrl.resetUsers()
	return nil
}

func (cntrl *Controller) GetLabelIDs(username string, labelNames []string) ([]string, error) {
	labelIDs := []string{}
	for _, labelName := range labelNames {
		labelID, err := cntrl.getLabelID(username, labelName)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, labelID)
	}
	return labelIDs, nil
}

func (cntrl *Controller) getLabelID(username, labelName string) (string, error) {
	if labelID, ok := systemLabelNameToID[labelName]; ok {
		return labelID, nil
	}
	labelName = getLabelNameWithoutPrefix(labelName)
	for _, label := range cntrl.labelsByUsername[username] {
		if label.Name == labelName {
			return label.ID, nil
		}
	}
	return "", fmt.Errorf("label %s:%s does not exist", username, labelName)
}

func getLabelNameWithoutPrefix(name string) string {
	if strings.HasPrefix(name, "Folders/") {
		return strings.TrimPrefix(name, "Folders/")
	}
	if strings.HasPrefix(name, "Labels/") {
		return strings.TrimPrefix(name, "Labels/")
	}
	return name
}

func getLabelExclusive(name string) int {
	if strings.HasPrefix(name, "Folders/") {
		return 1
	}
	return 0
}

func (cntrl *Controller) AddUserMessage(username string, message *pmapi.Message) error {
	if _, ok := cntrl.messagesByUsername[username]; !ok {
		cntrl.messagesByUsername[username] = []*pmapi.Message{}
	}
	message.ID = cntrl.messageIDGenerator.next("")
	message.LabelIDs = append(message.LabelIDs, pmapi.AllMailLabel)
	cntrl.messagesByUsername[username] = append(cntrl.messagesByUsername[username], message)
	cntrl.resetUsers()
	return nil
}

func (cntrl *Controller) resetUsers() {
	for _, fakeAPI := range cntrl.fakeAPIs {
		_ = fakeAPI.setUser(fakeAPI.username)
	}
}

func (cntrl *Controller) GetMessageID(username, messageIndex string) string {
	return messageIndex
}
