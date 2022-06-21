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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
)

type header textproto.MIMEHeader

type rawHeader map[string]json.RawMessage

func (h *header) UnmarshalJSON(b []byte) error {
	if *h == nil {
		*h = make(header)
	}

	raw := make(rawHeader)
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	for k, v := range raw {
		// Most headers are string because they have only one value.
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			textproto.MIMEHeader(*h).Set(k, s)
			continue
		}

		// If it's not a string, it must be an array of strings.
		var a []string
		if err := json.Unmarshal(v, &a); err != nil {
			return fmt.Errorf("pmapi: attachment header field is neither a string nor an array of strings: %v", err)
		}
		for _, vv := range a {
			textproto.MIMEHeader(*h).Add(k, vv)
		}
	}

	return nil
}

const (
	DispositionInline     = "inline"
	DispositionAttachment = "attachment"
)

// Attachment represents a message attachment.
type Attachment struct {
	ID          string `json:",omitempty"`
	MessageID   string `json:",omitempty"` // msg v3 ???
	Name        string `json:",omitempty"`
	Size        int64  `json:",omitempty"`
	MIMEType    string `json:",omitempty"`
	ContentID   string `json:",omitempty"`
	Disposition string
	KeyPackets  string `json:",omitempty"`
	Signature   string `json:",omitempty"`

	Header textproto.MIMEHeader `json:"-"`
}

// Define a new type to prevent MarshalJSON/UnmarshalJSON infinite loops.
type attachment Attachment

type rawAttachment struct {
	attachment

	Header header `json:"Headers,omitempty"`
}

func (a *Attachment) MarshalJSON() ([]byte, error) {
	var raw rawAttachment
	raw.attachment = attachment(*a)

	if a.Header != nil {
		raw.Header = header(a.Header)
	}

	return json.Marshal(&raw)
}

func (a *Attachment) UnmarshalJSON(b []byte) error {
	var raw rawAttachment
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	*a = Attachment(raw.attachment)

	if raw.Header != nil {
		a.Header = textproto.MIMEHeader(raw.Header)
	}

	return nil
}

// Decrypt decrypts this attachment's data from r using the keys from kr.
func (a *Attachment) Decrypt(r io.Reader, kr *crypto.KeyRing) (decrypted io.Reader, err error) {
	keyPackets, err := base64.StdEncoding.DecodeString(a.KeyPackets)
	if err != nil {
		return
	}
	return decryptAttachment(kr, keyPackets, r)
}

// Encrypt encrypts an attachment.
func (a *Attachment) Encrypt(kr *crypto.KeyRing, att io.Reader) (encrypted io.Reader, err error) {
	return encryptAttachment(kr, att, a.Name)
}

func (a *Attachment) DetachedSign(kr *crypto.KeyRing, att io.Reader) (signed io.Reader, err error) {
	return signAttachment(kr, att)
}

// CreateAttachment uploads an attachment. It must be already encrypted and contain a MessageID.
//
// The returned created attachment contains the new attachment ID and its size.
func (c *client) CreateAttachment(ctx context.Context, att *Attachment, attData io.Reader, sigData io.Reader) (*Attachment, error) {
	var res struct {
		Attachment *Attachment
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).
			SetMultipartFormData(map[string]string{
				"Filename":  att.Name,
				"MessageID": att.MessageID,
				"MIMEType":  att.MIMEType,
				"ContentID": att.ContentID,
			}).
			SetMultipartField("DataPacket", "DataPacket.pgp", "application/octet-stream", attData).
			SetMultipartField("Signature", "Signature.pgp", "application/octet-stream", sigData).
			Post("/mail/v4/attachments")
	}); err != nil {
		return nil, err
	}

	return res.Attachment, nil
}

// GetAttachment gets an attachment's content. The returned data is encrypted.
func (c *client) GetAttachment(ctx context.Context, attachmentID string) (att io.ReadCloser, err error) {
	res, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetDoNotParseResponse(true).Get("/mail/v4/attachments/" + attachmentID)
	})
	if err != nil {
		return nil, err
	}

	return res.RawBody(), nil
}
