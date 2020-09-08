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

package liveapi

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	messageUtils "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

func (ctl *Controller) AddUserMessage(username string, message *pmapi.Message) (string, error) {
	client, ok := ctl.pmapiByUsername[username]
	if !ok {
		return "", fmt.Errorf("user %s does not exist", username)
	}

	if message.Flags == 0 {
		message.Flags = pmapi.ComputeMessageFlagsByLabels(message.LabelIDs)
	}

	body, err := buildMessage(client, message)
	if err != nil {
		return "", errors.Wrap(err, "failed to build message")
	}

	req := &pmapi.ImportMsgReq{
		AddressID: message.AddressID,
		Body:      body.Bytes(),
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

func buildMessage(client pmapi.Client, message *pmapi.Message) (*bytes.Buffer, error) {
	if err := encryptMessage(client, message); err != nil {
		return nil, errors.Wrap(err, "failed to encrypt message")
	}

	body := &bytes.Buffer{}
	if err := buildMessageHeader(message, body); err != nil {
		return nil, errors.Wrap(err, "failed to build message header")
	}
	if err := buildMessageBody(message, body); err != nil {
		return nil, errors.Wrap(err, "failed to build message body")
	}
	return body, nil
}

func encryptMessage(client pmapi.Client, message *pmapi.Message) error {
	kr, err := client.KeyRingForAddressID(message.AddressID)
	if err != nil {
		return errors.Wrap(err, "failed to get keyring for address")
	}

	if err = message.Encrypt(kr, nil); err != nil {
		return errors.Wrap(err, "failed to encrypt message body")
	}
	return nil
}

func buildMessageHeader(message *pmapi.Message, body *bytes.Buffer) error {
	header := messageUtils.GetHeader(message)
	header.Set("Content-Type", "multipart/mixed; boundary="+messageUtils.GetBoundary(message))
	header.Del("Content-Disposition")
	header.Del("Content-Transfer-Encoding")

	if err := http.Header(header).Write(body); err != nil {
		return errors.Wrap(err, "failed to write header")
	}
	_, _ = body.WriteString("\r\n")
	return nil
}

func buildMessageBody(message *pmapi.Message, body *bytes.Buffer) error {
	mw := multipart.NewWriter(body)
	if err := mw.SetBoundary(messageUtils.GetBoundary(message)); err != nil {
		return errors.Wrap(err, "failed to set boundary")
	}

	bodyHeader := messageUtils.GetBodyHeader(message)
	bodyHeader.Set("Content-Transfer-Encoding", "7bit")

	part, err := mw.CreatePart(bodyHeader)
	if err != nil {
		return errors.Wrap(err, "failed to create message body part")
	}
	if _, err := io.WriteString(part, message.Body); err != nil {
		return errors.Wrap(err, "failed to write message body")
	}
	_ = mw.Close()
	return nil
}

func (ctl *Controller) GetMessageID(username, messageIndex string) string {
	idx, err := strconv.Atoi(messageIndex)
	if err != nil {
		panic(fmt.Sprintf("message index %s not found", messageIndex))
	}
	return ctl.messageIDsByUsername[username][idx-1]
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
