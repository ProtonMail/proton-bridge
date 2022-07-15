// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
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

package pmapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

const (
	MaxImportMessageRequestLength = 10
	MaxImportMessageRequestSize   = 25 * 1024 * 1024 // 25 MB total limit
)

type ImportMsgReq struct {
	Metadata *ImportMetadata // Metadata about the message to import.
	Message  []byte          // The raw RFC822 message.
}

type ImportMsgReqs []*ImportMsgReq

func (reqs ImportMsgReqs) buildMultipartFormData() ([]*resty.MultipartField, error) {
	metadata := make(map[string]*ImportMetadata, len(reqs))
	fields := make([]*resty.MultipartField, 0, len(reqs))

	for i, req := range reqs {
		name := strconv.Itoa(i)

		metadata[name] = req.Metadata

		fields = append(fields, &resty.MultipartField{
			Param:       name,
			FileName:    name + ".eml",
			ContentType: "message/rfc822",
			Reader:      bytes.NewReader(req.Message),
		})
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	fields = append(fields, &resty.MultipartField{
		Param:       "Metadata",
		ContentType: "application/json",
		Reader:      bytes.NewReader(b),
	})

	return fields, nil
}

type ImportMetadata struct {
	AddressID    string
	Unread       Boolean  // 0: read, 1: unread.
	IsReplied    Boolean  // 1 if the message has been replied.
	IsRepliedAll Boolean  // 1 if the message has been replied to all.
	IsForwarded  Boolean  // 1 if the message has been forwarded.
	Time         int64    // The time when the message was received as a Unix time.
	Flags        int64    // The type of the imported message.
	LabelIDs     []string // The labels to apply to the imported message. Must contain at least one system label.
}

type ImportMsgRes struct {
	// The error encountered while importing the message, if any.
	Error error
	// The newly created message ID.
	MessageID string
}

// Import imports messages to the user's account.
func (c *client) Import(ctx context.Context, reqs ImportMsgReqs) ([]*ImportMsgRes, error) {
	if len(reqs) == 0 {
		return nil, errors.New("missing import requests")
	}

	if len(reqs) > MaxImportMessageRequestLength {
		log.
			WithField("count", len(reqs)).
			Warn("Importing too many messages at once.")
		return nil, errors.New("request is too long")
	}

	remainingSize := MaxImportMessageRequestSize
	for _, req := range reqs {
		remainingSize -= len(req.Message)
		if remainingSize < 0 {
			log.
				WithField("count", len(reqs)).
				WithField("size", MaxImportMessageRequestLength-remainingSize).
				Warn("Importing too big message(s)")
			return nil, errors.New("request size is too big")
		}
	}

	fields, err := reqs.buildMultipartFormData()
	if err != nil {
		return nil, err
	}

	var res struct {
		Responses []struct {
			Name     string
			Response struct {
				Error
				MessageID string
			}
		}
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetMultipartFields(fields...).SetResult(&res).Post("/mail/v4/messages/import")
	}); err != nil {
		return nil, err
	}

	resps := make([]*ImportMsgRes, 0, len(res.Responses))

	for _, resp := range res.Responses {
		var err error

		if resp.Response.Code != 1000 {
			err = resp.Response.Error
		}

		resps = append(resps, &ImportMsgRes{
			Error:     err,
			MessageID: resp.Response.MessageID,
		})
	}

	return resps, nil
}
