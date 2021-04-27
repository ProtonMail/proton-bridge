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

package liveapi

import (
	"fmt"

	messageUtils "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

func (ctl *Controller) AddUserMessage(username string, message *pmapi.Message) (string, error) {
	if message.NumAttachments != 0 {
		return "", errors.New("add user messages with attachments is not implemented for live")
	}

	client, ok := ctl.pmapiByUsername[username]
	if !ok {
		return "", fmt.Errorf("user %s does not exist", username)
	}

	if message.Flags == 0 {
		message.Flags = pmapi.ComputeMessageFlagsByLabels(message.LabelIDs)
	}

	kr, err := client.KeyRingForAddressID(message.AddressID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get keyring while adding user message")
	}

	body, err := messageUtils.BuildEncrypted(message, nil, kr)
	if err != nil {
		return "", errors.Wrap(err, "failed to build message")
	}

	req := &pmapi.ImportMsgReq{
		AddressID: message.AddressID,
		Body:      body,
		Unread:    message.Unread,
		Time:      message.Time,
		Flags:     message.Flags,
		LabelIDs:  message.LabelIDs,
	}

	results, err := client.Import([]*pmapi.ImportMsgReq{req})
	if err != nil {
		return "", errors.Wrap(err, "failed to make an import")
	}
	result := results[0]
	if result.Error != nil {
		return "", errors.Wrap(result.Error, "failed to import message")
	}
	ctl.messageIDsByUsername[username] = append(ctl.messageIDsByUsername[username], result.MessageID)

	return result.MessageID, nil
}

func (ctl *Controller) GetMessages(username, labelID string) ([]*pmapi.Message, error) {
	client, ok := ctl.pmapiByUsername[username]
	if !ok {
		return nil, fmt.Errorf("user %s does not exist", username)
	}

	page := 0
	messages := []*pmapi.Message{}

	for {
		// ListMessages returns empty result, not error, asking for page out of range.
		pageMessages, _, err := client.ListMessages(&pmapi.MessagesFilter{
			Page:     page,
			PageSize: 150,
			LabelID:  labelID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list messages")
		}
		messages = append(messages, pageMessages...)
		if len(pageMessages) < 150 {
			break
		}
	}

	return messages, nil
}
