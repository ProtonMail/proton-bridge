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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"

	"github.com/stretchr/testify/require"
)

const testAttachmentCleartext = `cc,
dille.
`

// Attachment cleartext encrypted with testPrivateKeyRing.
const testKeyPacket = `wcBMA0fcZ7XLgmf2AQf/cHhfDRM9zlIuBi+h2W6DKjbbyIHMkgF6ER3JEvn/tSruUH8KTGt0N7Z+a80FFMCuXn1Y1I/nW7MVrNhGuJZAF4OymD8ugvuoAMIQX0eCYEpPXzRIWJBZg82AuowmFMsv8Dgvq4bTZq4cttI3CZcxKUNXuAearmNpmgplUKWj5USmRXK4iGB3VFGjidXkxbElrP4fD5A/rfEZ5aJgCsegqcXxX3MEjWXi9pFzgd/9phOvl1ZFm9U9hNoVAW3QsgmVeihnKaDZUyf2Qsigij21QKAUxw9U3y89eTUIqZAcmIgqeDujA3RWBgJwjtY/lOyhEmkf3AWKzehvf1xtJmCWDg==`
const testDataPacket = `0ksB6S4f4l8C1NB8yzmd/jNi0xqEZsyTDLdTP+N4Qxh3NZjla+yGRvC9rGmoUL7XVyowsG/GKTf2LXF/5E5FkX/3WMYwIv1n11ExyAE=`

var testAttachment = &Attachment{
	ID:         "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw==",
	Name:       "croutonmail.txt",
	Size:       77,
	MIMEType:   "text/plain",
	KeyPackets: testKeyPacket,

	Header: textproto.MIMEHeader{
		"Content-Description": {"You'll never believe what's in this text file"},
		"X-Mailer":            {"Microsoft Outlook 15.0", "Microsoft Live Mail 42.0"},
	},
	MessageID: "h3CD-DT7rLoAw1vmpcajvIPAl-wwDfXR2MHtWID3wuQURDBKTiGUAwd6E2WBbS44QQKeXImW-axm6X0hAfcVCA==",
}

// Part of GET /mail/messages/{id} response from server.
const testAttachmentJSON = `{
    "ID": "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw==",
    "Name": "croutonmail.txt",
    "Size": 77,
    "MIMEType": "text/plain",
    "KeyPackets": "` + testKeyPacket + `",
    "Headers": {
        "content-description": "You'll never believe what's in this text file",
        "x-mailer": [
            "Microsoft Outlook 15.0",
            "Microsoft Live Mail 42.0"
        ]
    }
}
`

// POST /mail/attachment/ response from server.
const testCreatedAttachmentBody = `{
    "Code": 1000,
    "Attachment": {"ID": "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw=="}
}`

func TestAttachment_UnmarshalJSON(t *testing.T) {
	r := require.New(t)
	att := new(Attachment)
	err := json.Unmarshal([]byte(testAttachmentJSON), att)
	r.NoError(err)

	att.MessageID = testAttachment.MessageID // This isn't in the server response

	r.Equal(testAttachment, att)
}

func TestClient_CreateAttachment(t *testing.T) {
	r := require.New(t)
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(checkMethodAndPath(req, "POST", "/mail/v4/attachments"))

		contentType, params, err := pmmime.ParseMediaType(req.Header.Get("Content-Type"))
		r.NoError(err)
		r.Equal("multipart/form-data", contentType)

		mr := multipart.NewReader(req.Body, params["boundary"])
		form, err := mr.ReadForm(10 * 1024)
		r.NoError(err)
		defer r.NoError(form.RemoveAll())

		r.Equal(testAttachment.Name, form.Value["Filename"][0])
		r.Equal(testAttachment.MessageID, form.Value["MessageID"][0])
		r.Equal(testAttachment.MIMEType, form.Value["MIMEType"][0])

		dataFile, err := form.File["DataPacket"][0].Open()
		r.NoError(err)
		defer r.NoError(dataFile.Close())

		b, err := ioutil.ReadAll(dataFile)
		r.NoError(err)
		r.Equal(testAttachmentCleartext, string(b))

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, testCreatedAttachmentBody)
	}))
	defer s.Close()

	reader := strings.NewReader(testAttachmentCleartext) // In reality, this thing is encrypted
	created, err := c.CreateAttachment(context.Background(), testAttachment, reader, strings.NewReader(""))
	r.NoError(err)

	r.Equal(testAttachment.ID, created.ID)
}

func TestClient_GetAttachment(t *testing.T) {
	r := require.New(t)
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(checkMethodAndPath(req, "GET", "/mail/v4/attachments/"+testAttachment.ID))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testAttachmentCleartext)
	}))
	defer s.Close()

	att, err := c.GetAttachment(context.Background(), testAttachment.ID)
	r.NoError(err)
	defer att.Close() //nolint:errcheck

	// In reality, r contains encrypted data
	b, err := ioutil.ReadAll(att)
	r.NoError(err)

	r.Equal(testAttachmentCleartext, string(b))
}

func TestAttachmentDecrypt(t *testing.T) {
	r := require.New(t)

	rawKeyPacket, err := base64.StdEncoding.DecodeString(testKeyPacket)
	r.NoError(err)

	rawDataPacket, err := base64.StdEncoding.DecodeString(testDataPacket)
	r.NoError(err)

	decryptAndCheck(r, bytes.NewBuffer(append(rawKeyPacket, rawDataPacket...)))
}

func TestAttachmentEncrypt(t *testing.T) {
	r := require.New(t)

	encryptedReader, err := testAttachment.Encrypt(
		testPublicKeyRing,
		bytes.NewBufferString(testAttachmentCleartext),
	)
	r.NoError(err)

	// The result is always different due to session key. The best way is to
	// test result of encryption by decrypting again acn coparet to cleartext.
	decryptAndCheck(r, encryptedReader)
}

func decryptAndCheck(r *require.Assertions, data io.Reader) {
	// First separate KeyPacket from encrypted data. In our case keypacket
	// has 271 bytes.
	raw, err := ioutil.ReadAll(data)
	r.NoError(err)
	rawKeyPacket := raw[:271]
	rawDataPacket := raw[271:]

	// KeyPacket is retrieve by get GET /mail/messages/{id}
	haveAttachment := &Attachment{
		KeyPackets: base64.StdEncoding.EncodeToString(rawKeyPacket),
	}

	// DataPacket is received from GET /mail/attachments/{id}
	decryptedReader, err := haveAttachment.Decrypt(bytes.NewBuffer(rawDataPacket), testPrivateKeyRing)
	r.NoError(err)

	b, err := ioutil.ReadAll(decryptedReader)
	r.NoError(err)

	r.Equal(testAttachmentCleartext, string(b))
}
