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
	"bytes"
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

var errWasNotUpdated = errors.New("message was not updated")

func (api *FakePMAPI) GetMessage(apiID string) (*pmapi.Message, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/messages/"+apiID, nil); err != nil {
		return nil, err
	}
	for _, message := range api.messages {
		if message.ID == apiID {
			return message, nil
		}
	}
	return nil, fmt.Errorf("message %s not found", apiID)
}

// ListMessages does not implement following filters:
//  * Sort (it sorts by ID only), but Desc works
//  * Keyword
//  * To
//  * Subject
//  * ID
//  * Attachments
//  * AutoWildcard
func (api *FakePMAPI) ListMessages(filter *pmapi.MessagesFilter) ([]*pmapi.Message, int, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/messages", filter); err != nil {
		return nil, 0, err
	}
	pageSize := filter.PageSize
	if pageSize > 150 {
		pageSize = 150
	}

	messages := []*pmapi.Message{}
	messageCount := 0

	skipByIDBegin := true
	skipByIDEnd := false
	skipByPaging := pageSize * filter.Page

	for idx := 0; idx < len(api.messages); idx++ {
		var message *pmapi.Message
		if filter.Desc == nil || !*filter.Desc {
			message = api.messages[idx]
			if filter.BeginID == "" || message.ID == filter.BeginID {
				skipByIDBegin = false
			}
		} else {
			message = api.messages[len(api.messages)-1-idx]
			if filter.EndID == "" || message.ID == filter.EndID {
				skipByIDBegin = false
			}
		}
		if skipByIDBegin || skipByIDEnd {
			continue
		}
		if filter.Desc == nil || !*filter.Desc {
			if message.ID == filter.EndID {
				skipByIDEnd = true
			}
		} else {
			if message.ID == filter.BeginID {
				skipByIDEnd = true
			}
		}
		if !isMessageMatchingFilter(filter, message) {
			continue
		}
		messageCount++

		if skipByPaging > 0 {
			skipByPaging--
			continue
		}
		if len(messages) == pageSize || (filter.Limit != 0 && len(messages) == filter.Limit) {
			continue
		}
		messages = append(messages, copyFilteredMessage(message))
	}

	return messages, messageCount, nil
}

func isMessageMatchingFilter(filter *pmapi.MessagesFilter, message *pmapi.Message) bool {
	if filter.ExternalID != "" && filter.ExternalID != message.ExternalID {
		return false
	}
	if filter.ConversationID != "" && filter.ConversationID != message.ConversationID {
		return false
	}
	if filter.AddressID != "" && filter.AddressID != message.AddressID {
		return false
	}
	if filter.From != "" && filter.From != message.Sender.Address {
		return false
	}
	if filter.LabelID != "" && !hasItem(message.LabelIDs, filter.LabelID) {
		return false
	}
	if filter.Begin != 0 && filter.Begin > message.Time {
		return false
	}
	if filter.End != 0 && filter.End < message.Time {
		return false
	}
	if filter.Unread != nil {
		wantUnread := 0
		if *filter.Unread {
			wantUnread = 1
		}
		if message.Unread != wantUnread {
			return false
		}
	}
	return true
}

func copyFilteredMessage(message *pmapi.Message) *pmapi.Message {
	filteredMessage := &pmapi.Message{}
	*filteredMessage = *message
	filteredMessage.Body = ""
	filteredMessage.Header = nil
	return filteredMessage
}

func (api *FakePMAPI) CreateDraft(message *pmapi.Message, parentID string, action int) (*pmapi.Message, error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/messages", &pmapi.DraftReq{
		Message:              message,
		ParentID:             parentID,
		Action:               action,
		AttachmentKeyPackets: []string{},
	}); err != nil {
		return nil, err
	}
	if parentID != "" {
		if _, err := api.GetMessage(parentID); err != nil {
			return nil, err
		}
	}
	if message.Subject == "" {
		message.Subject = "(No Subject)"
	}
	message.LabelIDs = append(message.LabelIDs, pmapi.DraftLabel)
	message.LabelIDs = append(message.LabelIDs, pmapi.AllMailLabel)
	message.ID = api.controller.messageIDGenerator.next("")
	api.addMessage(message)
	return message, nil
}

func (api *FakePMAPI) SendMessage(messageID string, sendMessageRequest *pmapi.SendMessageReq) (sent, parent *pmapi.Message, err error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/messages/"+messageID, sendMessageRequest); err != nil {
		return nil, nil, err
	}
	message, err := api.GetMessage(messageID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "draft does not exist")
	}
	message.Time = time.Now().Unix()
	message.LabelIDs = append(message.LabelIDs, pmapi.SentLabel)
	api.addEventMessage(pmapi.EventUpdate, message)
	return message, nil, nil
}

func (api *FakePMAPI) Import(importMessageRequests []*pmapi.ImportMsgReq) ([]*pmapi.ImportMsgRes, error) {
	if err := api.checkAndRecordCall(POST, "/import", importMessageRequests); err != nil {
		return nil, err
	}
	msgRes := []*pmapi.ImportMsgRes{}
	for _, msgReq := range importMessageRequests {
		message, err := api.generateMessageFromImportRequest(msgReq)
		if err != nil {
			msgRes = append(msgRes, &pmapi.ImportMsgRes{
				Error: err,
			})
			continue
		}
		msgRes = append(msgRes, &pmapi.ImportMsgRes{
			Error:     nil,
			MessageID: message.ID,
		})
		api.addMessage(message)
	}
	return msgRes, nil
}

func (api *FakePMAPI) generateMessageFromImportRequest(msgReq *pmapi.ImportMsgReq) (*pmapi.Message, error) {
	m, _, _, _, err := message.Parse(bytes.NewReader(msgReq.Body), "", "")
	if err != nil {
		return nil, err
	}

	messageID := api.controller.messageIDGenerator.next("")

	return &pmapi.Message{
		ID:        messageID,
		AddressID: msgReq.AddressID,
		Sender:    m.Sender,
		ToList:    m.ToList,
		Subject:   m.Subject,
		Unread:    msgReq.Unread,
		LabelIDs:  append(msgReq.LabelIDs, pmapi.AllMailLabel),
		Body:      m.Body,
		Header:    m.Header,
		Flags:     msgReq.Flags,
		Time:      msgReq.Time,
	}, nil
}

func (api *FakePMAPI) addMessage(message *pmapi.Message) {
	api.messages = append(api.messages, message)
	api.addEventMessage(pmapi.EventCreate, message)
}

func (api *FakePMAPI) DeleteMessages(apiIDs []string) error {
	err := api.deleteMessages(PUT, "/mail/v4/messages/delete", &pmapi.MessagesActionReq{
		IDs: apiIDs,
	}, func(message *pmapi.Message) bool {
		return hasItem(apiIDs, message.ID)
	})
	if err != nil {
		return err
	}

	if len(apiIDs) == 0 {
		return errBadRequest
	}

	return nil
}

func (api *FakePMAPI) EmptyFolder(labelID string, addressID string) error {
	err := api.deleteMessages(DELETE, "/mail/v4/messages/empty?LabelID="+labelID+"&AddressID="+addressID, nil, func(message *pmapi.Message) bool {
		return hasItem(message.LabelIDs, labelID) && message.AddressID == addressID
	})
	if err != nil {
		return err
	}

	if labelID == "" {
		return errBadRequest
	}

	return nil
}

func (api *FakePMAPI) deleteMessages(method method, path string, request interface{}, shouldBeDeleted func(*pmapi.Message) bool) error {
	if err := api.checkAndRecordCall(method, path, request); err != nil {
		return err
	}
	newMessages := []*pmapi.Message{}
	for _, message := range api.messages {
		if shouldBeDeleted(message) {
			if hasItem(message.LabelIDs, pmapi.TrashLabel) {
				api.addEventMessage(pmapi.EventDelete, message)
				continue
			}
			message.LabelIDs = []string{pmapi.TrashLabel, pmapi.AllMailLabel}
			api.addEventMessage(pmapi.EventUpdate, message)
		}
		newMessages = append(newMessages, message)
	}
	api.messages = newMessages
	return nil
}

func (api *FakePMAPI) LabelMessages(apiIDs []string, labelID string) error {
	return api.updateMessages(PUT, "/mail/v4/messages/label", &pmapi.LabelMessagesReq{
		IDs:     apiIDs,
		LabelID: labelID,
	}, apiIDs, func(message *pmapi.Message) error {
		if labelID == "" {
			return errBadRequest
		}
		if labelID == pmapi.TrashLabel {
			message.LabelIDs = []string{pmapi.TrashLabel, pmapi.AllMailLabel}
			return nil
		}
		if api.isLabelFolder(labelID) {
			labelIDs := []string{}
			for _, existingLabelID := range message.LabelIDs {
				if !api.isLabelFolder(existingLabelID) {
					labelIDs = append(labelIDs, existingLabelID)
				}
			}
			message.LabelIDs = labelIDs
		}
		message.LabelIDs = addItem(message.LabelIDs, labelID)
		return nil
	})
}

func (api *FakePMAPI) UnlabelMessages(apiIDs []string, labelID string) error {
	return api.updateMessages(PUT, "/mail/v4/messages/unlabel", &pmapi.LabelMessagesReq{
		IDs:     apiIDs,
		LabelID: labelID,
	}, apiIDs, func(message *pmapi.Message) error {
		if labelID == "" {
			return errBadRequest
		}
		// All Mail and Sent cannot be unlabeled, but API will not throw error.
		if labelID == pmapi.AllMailLabel || labelID == pmapi.SentLabel {
			return errWasNotUpdated
		}

		message.LabelIDs = removeItem(message.LabelIDs, labelID)
		return nil
	})
}

func (api *FakePMAPI) MarkMessagesRead(apiIDs []string) error {
	return api.updateMessages(PUT, "/mail/v4/messages/read", &pmapi.MessagesActionReq{
		IDs: apiIDs,
	}, apiIDs, func(message *pmapi.Message) error {
		if message.Unread == 0 {
			return errWasNotUpdated
		}
		message.Unread = 0
		return nil
	})
}

func (api *FakePMAPI) MarkMessagesUnread(apiIDs []string) error {
	err := api.updateMessages(PUT, "/mail/v4/messages/unread", &pmapi.MessagesActionReq{
		IDs: apiIDs,
	}, apiIDs, func(message *pmapi.Message) error {
		if message.Unread == 1 {
			return errWasNotUpdated
		}
		message.Unread = 1
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (api *FakePMAPI) updateMessages(method method, path string, request interface{}, apiIDs []string, updateCallback func(*pmapi.Message) error) error { //nolint[unparam]
	if err := api.checkAndRecordCall(method, path, request); err != nil {
		return err
	}
	// API will return error if you send request for no apiIDs
	if len(apiIDs) == 0 {
		return errBadRequest
	}
	for _, message := range api.messages {
		if hasItem(apiIDs, message.ID) {
			err := updateCallback(message)
			if err != nil {
				if err == errWasNotUpdated {
					continue
				} else {
					return err
				}
			}
			api.addEventMessage(pmapi.EventUpdate, message)
		}
	}
	return nil
}
