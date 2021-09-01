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

package fakeapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/test/accounts"
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

func (ctl *Controller) TurnInternetConnectionOff() {
	ctl.log.Warn("Turning OFF internet")
	ctl.noInternetConnection = true
	for _, observer := range ctl.clientManager.connectionObservers {
		observer.OnDown()
	}
}

func (ctl *Controller) TurnInternetConnectionOn() {
	ctl.log.Warn("Turning ON internet")
	ctl.noInternetConnection = false
	for _, observer := range ctl.clientManager.connectionObservers {
		observer.OnUp()
	}
}

func (ctl *Controller) ReorderAddresses(user *pmapi.User, addressIDs []string) error {
	api := ctl.getFakeAPIForUser(user.ID)
	if api == nil {
		return errors.New("no such user")
	}

	return api.ReorderAddresses(context.Background(), addressIDs)
}

func (ctl *Controller) AddUser(account *accounts.TestAccount) error {
	ctl.usersByUsername[account.User().Name] = &fakeUser{
		user:     account.User(),
		password: account.Password(),
		has2FA:   account.IsTwoFAEnabled(),
	}
	ctl.addressesByUsername[account.User().Name] = account.Addresses()
	ctl.createSession(account.User().Name, true)

	return nil
}

func (ctl *Controller) AddUserLabel(username string, label *pmapi.Label) error {
	if _, ok := ctl.labelsByUsername[username]; !ok {
		ctl.labelsByUsername[username] = []*pmapi.Label{}
	}

	labelName := getLabelNameWithoutPrefix(label.Name)
	for _, existingLabel := range ctl.labelsByUsername[username] {
		if existingLabel.Name == labelName {
			return fmt.Errorf("folder or label %s already exists", label.Name)
		}
	}

	label.Exclusive = getLabelExclusive(label.Name)
	prefix := "label"
	if label.Exclusive {
		prefix = "folder"
	}
	label.ID = ctl.labelIDGenerator.next(prefix)
	label.Name = labelName
	if label.Path == "" {
		label.Path = label.Name
	}
	ctl.labelsByUsername[username] = append(ctl.labelsByUsername[username], label)
	ctl.resetUsers()
	return nil
}

func (ctl *Controller) GetLabelIDs(username string, labelNames []string) ([]string, error) {
	labelIDs := []string{}
	for _, labelName := range labelNames {
		labelID, err := ctl.getLabelID(username, labelName)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, labelID)
	}
	return labelIDs, nil
}

func (ctl *Controller) getLabelID(username, labelName string) (string, error) {
	if labelID, ok := systemLabelNameToID[labelName]; ok {
		return labelID, nil
	}
	labelName = getLabelNameWithoutPrefix(labelName)
	for _, label := range ctl.labelsByUsername[username] {
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

func getLabelExclusive(name string) pmapi.Boolean {
	return pmapi.Boolean(strings.HasPrefix(name, "Folders/"))
}

func (ctl *Controller) AddUserMessage(username string, message *pmapi.Message) (string, error) {
	if _, ok := ctl.messagesByUsername[username]; !ok {
		ctl.messagesByUsername[username] = []*pmapi.Message{}
	}
	message.ID = ctl.messageIDGenerator.next("")
	message.LabelIDs = append(message.LabelIDs, pmapi.AllMailLabel)

	for iAtt := 0; iAtt < message.NumAttachments; iAtt++ {
		message.Attachments = append(message.Attachments, newTestAttachment(iAtt, message.ID))
	}

	ctl.messagesByUsername[username] = append(ctl.messagesByUsername[username], message)
	ctl.resetUsers()
	return message.ID, nil
}

func (ctl *Controller) getFakeAPIForUser(userID string) *FakePMAPI {
	for _, fakeAPI := range ctl.fakeAPIs {
		if fakeAPI.userID == userID {
			return fakeAPI
		}
	}
	return nil
}

func (ctl *Controller) resetUsers() {
	for _, fakeAPI := range ctl.fakeAPIs {
		_ = fakeAPI.setUser(fakeAPI.username)
	}
}

func (ctl *Controller) GetMessages(username, labelID string) ([]*pmapi.Message, error) {
	messages := []*pmapi.Message{}
	for _, fakeAPI := range ctl.fakeAPIs {
		if fakeAPI.username == username {
			for _, message := range fakeAPI.messages {
				if labelID == "" || message.HasLabelID(labelID) {
					messages = append(messages, message)
				}
			}
		}
	}
	return messages, nil
}

func (ctl *Controller) GetAuthClient(username string) pmapi.Client {
	for uid, session := range ctl.sessionsByUID {
		if session.username == username {
			return ctl.clientManager.NewClient(uid, session.acc, session.ref, time.Now())
		}
	}

	ctl.log.WithField("username", username).Fatal("Cannot get authenticated client.")

	return nil
}
