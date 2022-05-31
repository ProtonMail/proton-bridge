// Copyright (c) 2022 Proton AG
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

package fakeapi

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

var errWasNotUpdated = errors.New("message was not updated")

func (api *FakePMAPI) GetMessage(_ context.Context, apiID string) (*pmapi.Message, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/messages/"+apiID, nil); err != nil {
		return nil, err
	}
	if msg := api.getMessage(apiID); msg != nil {
		return msg, nil
	}
	return nil, pmapi.ErrUnprocessableEntity{OriginalError: fmt.Errorf("message %s not found", apiID)}
}

// ListMessages does not implement following filters:
//  * Sort (it sorts by ID only), but Desc works
//  * Keyword
//  * To
//  * Subject
//  * ID
//  * Attachments
//  * AutoWildcard
func (api *FakePMAPI) ListMessages(_ context.Context, filter *pmapi.MessagesFilter) ([]*pmapi.Message, int, error) {
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
		if bool(message.Unread) != *filter.Unread {
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

func (api *FakePMAPI) CreateDraft(ctx context.Context, message *pmapi.Message, parentID string, action int) (*pmapi.Message, error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/messages", &pmapi.DraftReq{
		Message:              message,
		ParentID:             parentID,
		Action:               action,
		AttachmentKeyPackets: []string{},
	}); err != nil {
		return nil, err
	}
	if parentID != "" {
		if _, err := api.GetMessage(ctx, parentID); err != nil {
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

func (api *FakePMAPI) SendMessage(ctx context.Context, messageID string, sendMessageRequest *pmapi.SendMessageReq) (sent, parent *pmapi.Message, err error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/messages/"+messageID, sendMessageRequest); err != nil {
		return nil, nil, err
	}
	message := api.getMessage(messageID)
	if message == nil {
		return nil, nil, errors.Wrap(err, "draft does not exist")
	}
	message.Time = time.Now().Unix()
	message.LabelIDs = append(message.LabelIDs, pmapi.SentLabel)
	message.Flags |= pmapi.FlagSent
	api.addEventMessage(pmapi.EventUpdate, message)
	return message, nil, nil
}

func (api *FakePMAPI) Import(_ context.Context, importMessageRequests pmapi.ImportMsgReqs) ([]*pmapi.ImportMsgRes, error) {
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
	m, _, _, _, err := message.Parse(bytes.NewReader(msgReq.Message)) //nolint:dogsled
	if err != nil {
		return nil, err
	}

	existingMsg := api.findMessage(m)
	if existingMsg != nil {
		for _, newLabelID := range api.generateLabelIDsFromImportRequest(msgReq) {
			if !existingMsg.HasLabelID(newLabelID) {
				existingMsg.LabelIDs = append(existingMsg.LabelIDs, newLabelID)
			}
		}
		return existingMsg, nil
	}

	messageID := api.controller.messageIDGenerator.next("")
	return &pmapi.Message{
		ID:         messageID,
		ExternalID: m.ExternalID,
		AddressID:  msgReq.Metadata.AddressID,
		Sender:     m.Sender,
		ToList:     m.ToList,
		Subject:    m.Subject,
		Unread:     msgReq.Metadata.Unread,
		LabelIDs:   api.generateLabelIDsFromImportRequest(msgReq),
		Body:       m.Body,
		Header:     m.Header,
		Flags:      msgReq.Metadata.Flags,
		Time:       msgReq.Metadata.Time,
	}, nil
}

// generateLabelIDsFromImportRequest simulates API where Sent and INBOX is the same
// mailbox but the message is shown in one or other based on the flags instead.
func (api *FakePMAPI) generateLabelIDsFromImportRequest(msgReq *pmapi.ImportMsgReq) []string {
	isInSentOrInbox := false
	labelIDs := []string{pmapi.AllMailLabel}
	for _, labelID := range msgReq.Metadata.LabelIDs {
		if labelID == pmapi.InboxLabel || labelID == pmapi.SentLabel {
			isInSentOrInbox = true
		} else {
			labelIDs = append(labelIDs, labelID)
		}
	}
	if isInSentOrInbox && (msgReq.Metadata.Flags&pmapi.FlagSent) != 0 {
		labelIDs = append(labelIDs, pmapi.SentLabel)
	}
	if isInSentOrInbox && (msgReq.Metadata.Flags&pmapi.FlagReceived) != 0 {
		labelIDs = append(labelIDs, pmapi.InboxLabel)
	}
	return labelIDs
}

func (api *FakePMAPI) findMessage(newMsg *pmapi.Message) *pmapi.Message {
	if newMsg.ExternalID == "" {
		return nil
	}
	for _, msg := range api.messages {
		// API surely has better algorithm, but this one is enough for us for now.
		if !msg.IsDraft() &&
			msg.Subject == newMsg.Subject &&
			msg.ExternalID == newMsg.ExternalID {
			return msg
		}
	}
	return nil
}

func (api *FakePMAPI) getMessage(msgID string) *pmapi.Message {
	for _, msg := range api.messages {
		if msg.ID == msgID {
			return msg
		}
	}
	return nil
}

func (api *FakePMAPI) addMessage(message *pmapi.Message) {
	if api.findMessage(message) != nil {
		return
	}
	api.messages = append(api.messages, message)
	api.addEventMessage(pmapi.EventCreate, message)
}

func (api *FakePMAPI) DeleteMessages(_ context.Context, apiIDs []string) error {
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

func (api *FakePMAPI) EmptyFolder(_ context.Context, labelID string, addressID string) error {
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
			if hasItem(message.LabelIDs, pmapi.TrashLabel) ||
				hasItem(message.LabelIDs, pmapi.SpamLabel) {
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

func (api *FakePMAPI) LabelMessages(_ context.Context, apiIDs []string, labelID string) error {
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

func (api *FakePMAPI) UnlabelMessages(_ context.Context, apiIDs []string, labelID string) error {
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

func (api *FakePMAPI) MarkMessagesRead(_ context.Context, apiIDs []string) error {
	return api.updateMessages(PUT, "/mail/v4/messages/read", &pmapi.MessagesActionReq{
		IDs: apiIDs,
	}, apiIDs, func(message *pmapi.Message) error {
		if !message.Unread {
			return errWasNotUpdated
		}
		message.Unread = false
		return nil
	})
}

func (api *FakePMAPI) MarkMessagesUnread(_ context.Context, apiIDs []string) error {
	err := api.updateMessages(PUT, "/mail/v4/messages/unread", &pmapi.MessagesActionReq{
		IDs: apiIDs,
	}, apiIDs, func(message *pmapi.Message) error {
		if message.Unread {
			return errWasNotUpdated
		}
		message.Unread = true
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (api *FakePMAPI) updateMessages(method method, path string, request interface{}, apiIDs []string, updateCallback func(*pmapi.Message) error) error { //nolint:unparam
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
