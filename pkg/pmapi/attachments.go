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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
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

// Attachment represents a message attachment.
type Attachment struct {
	ID         string `json:",omitempty"`
	MessageID  string `json:",omitempty"` // msg v3 ???
	Name       string `json:",omitempty"`
	Size       int64  `json:",omitempty"`
	MIMEType   string `json:",omitempty"`
	ContentID  string `json:",omitempty"`
	KeyPackets string `json:",omitempty"`
	Signature  string `json:",omitempty"`

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
func (a *Attachment) Decrypt(r io.Reader, kr *pmcrypto.KeyRing) (decrypted io.Reader, err error) {
	keyPackets, err := base64.StdEncoding.DecodeString(a.KeyPackets)
	if err != nil {
		return
	}
	return decryptAttachment(kr, keyPackets, r)
}

// Encrypt encrypts an attachment.
func (a *Attachment) Encrypt(kr *pmcrypto.KeyRing, att io.Reader) (encrypted io.Reader, err error) {
	return encryptAttachment(kr, att, a.Name)
}

func (a *Attachment) DetachedSign(kr *pmcrypto.KeyRing, att io.Reader) (signed io.Reader, err error) {
	return signAttachment(kr, att)
}

type CreateAttachmentRes struct {
	Res

	Attachment *Attachment
}

func writeAttachment(w *multipart.Writer, att *Attachment, r io.Reader, sig io.Reader) (err error) {
	// Create metadata fields.
	if err = w.WriteField("Filename", att.Name); err != nil {
		return
	}
	if err = w.WriteField("MessageID", att.MessageID); err != nil {
		return
	}
	if err = w.WriteField("MIMEType", att.MIMEType); err != nil {
		return
	}

	if err = w.WriteField("ContentID", att.ContentID); err != nil {
		return
	}

	// And send attachment data.
	ff, err := w.CreateFormFile("DataPacket", "DataPacket.pgp")
	if err != nil {
		return
	}
	if _, err = io.Copy(ff, r); err != nil {
		return
	}

	// And send attachment data.
	sigff, err := w.CreateFormFile("Signature", "Signature.pgp")
	if err != nil {
		return
	}

	if _, err = io.Copy(sigff, sig); err != nil {
		return
	}

	return err
}

// CreateAttachment uploads an attachment. It must be already encrypted and contain a MessageID.
//
// The returned created attachment contains the new attachment ID and its size.
func (c *Client) CreateAttachment(att *Attachment, r io.Reader, sig io.Reader) (created *Attachment, err error) {
	req, w, err := c.NewMultipartRequest("POST", "/attachments")
	if err != nil {
		return
	}

	// We will write the request as long as it is sent to the API.
	var res CreateAttachmentRes
	done := make(chan error, 1)
	go (func() {
		done <- c.DoJSON(req, &res)
	})()

	if err = writeAttachment(w.Writer, att, r, sig); err != nil {
		return
	}
	_ = w.Close()

	if err = <-done; err != nil {
		return
	}
	if err = res.Err(); err != nil {
		return
	}

	created = res.Attachment
	return
}

type UpdateAttachmentSignatureReq struct {
	Signature string
}

func (c *Client) UpdateAttachmentSignature(attachmentID, signature string) (err error) {
	updateReq := &UpdateAttachmentSignatureReq{signature}
	req, err := c.NewJSONRequest("PUT", "/attachments/"+attachmentID+"/signature", updateReq)
	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	return
}

// DeleteAttachment removes an attachment. message is the message ID, att is the attachment ID.
func (c *Client) DeleteAttachment(attID string) (err error) {
	req, err := c.NewRequest("DELETE", "/attachments/"+attID, nil)
	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
}

// GetAttachment gets an attachment's content. The returned data is encrypted.
func (c *Client) GetAttachment(id string) (att io.ReadCloser, err error) {
	if id == "" {
		err = errors.New("pmapi: cannot get an attachment with an empty id")
		return
	}

	req, err := c.NewRequest("GET", "/attachments/"+id, nil)
	if err != nil {
		return
	}

	res, err := c.Do(req, true)
	if err != nil {
		return
	}

	att = res.Body
	return
}
