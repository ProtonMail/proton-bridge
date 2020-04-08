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

package pmapi

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"
)

// Import errors.
const (
	ImportMessageTooLarge = 36022
)

// ImportReq is an import request.
type ImportReq struct {
	// A list of messages that will be imported.
	Messages []*ImportMsgReq
}

// WriteTo writes the import request to a multipart writer.
func (req *ImportReq) WriteTo(w *multipart.Writer) (err error) {
	// Create Metadata field.
	mw, err := w.CreateFormField("Metadata")
	if err != nil {
		return
	}

	// Build metadata.
	metadata := map[string]*ImportMsgReq{}
	for i, msg := range req.Messages {
		name := strconv.Itoa(i)
		metadata[name] = msg
	}

	// Write metadata.
	if err = json.NewEncoder(mw).Encode(metadata); err != nil {
		return
	}

	// Write messages.
	for i, msg := range req.Messages {
		name := strconv.Itoa(i)

		var fw io.Writer
		if fw, err = w.CreateFormFile(name, name+".eml"); err != nil {
			return err
		}

		if _, err = fw.Write(msg.Body); err != nil {
			return
		}
	}

	return err
}

// ImportMsgReq is a request to import a message. All fields are optional except AddressID and Body.
type ImportMsgReq struct {
	// The address where the message will be imported.
	AddressID string
	// The full MIME message.
	Body []byte `json:"-"`

	// 0: read, 1: unread.
	Unread int
	// 1 if the message has been replied.
	IsReplied int
	// 1 if the message has been replied to all.
	IsRepliedAll int
	// 1 if the message has been forwarded.
	IsForwarded int
	// The time when the message was received as a Unix time.
	Time int64
	// The type of the imported message.
	Flags int64
	// The labels to apply to the imported message. Must contain at least one system label.
	LabelIDs []string
}

// ImportRes is a response to an import request.
type ImportRes struct {
	Res

	Responses []struct {
		Name     string
		Response struct {
			Res
			MessageID string
		}
	}
}

// ImportMsgRes is a response to a single message import request.
type ImportMsgRes struct {
	// The error encountered while importing the message, if any.
	Error error
	// The newly created message ID.
	MessageID string
}

// Import imports messages to the user's account.
func (c *Client) Import(reqs []*ImportMsgReq) (resps []*ImportMsgRes, err error) {
	importReq := &ImportReq{Messages: reqs}

	req, w, err := NewMultipartRequest("POST", "/import")
	if err != nil {
		return
	}

	// We will write the request as long as it is sent to the API.
	var importRes ImportRes
	done := make(chan error, 1)
	go (func() {
		done <- c.DoJSON(req, &importRes)
	})()

	// Write the request.
	if err = importReq.WriteTo(w.Writer); err != nil {
		return
	}
	_ = w.Close()

	if err = <-done; err != nil {
		return
	}
	if err = importRes.Err(); err != nil {
		return
	}

	resps = make([]*ImportMsgRes, len(importRes.Responses))
	for i, r := range importRes.Responses {
		resps[i] = &ImportMsgRes{
			Error:     r.Response.Err(),
			MessageID: r.Response.MessageID,
		}
	}

	return resps, err
}
