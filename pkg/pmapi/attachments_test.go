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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testAttachment = &Attachment{
	ID:         "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw==",
	Name:       "croutonmail.txt",
	Size:       77,
	MIMEType:   "text/plain",
	KeyPackets: "wcBMA0fcZ7XLgmf2AQgAiRsOlnm1kSB4/lr7tYe6pBsRGn10GqwUhrwU5PMKOHdCgnO12jO3y3CzP0Yl/jGhAYja9wLDqH8X0sk3tY32u4Sb1Qe5IuzggAiCa4dwOJj5gEFMTHMzjIMPHR7A70XqUxMhmILye8V4KRm/j4c1sxbzA1rM3lYBumQuB5l/ck0Kgt4ZqxHVXHK5Q1l65FHhSXRj8qnunasHa30TYNzP8nmBA8BinnJxpiQ7FGc2umnUhgkFtjm5ixu9vyjr9ukwDTbwAXXfmY+o7tK7kqIXJcmTL6k2UeC6Mz1AagQtRCRtU+bv/3zGojq/trZo9lom3naIeQYa36Ketmcpj2Qwjg==",
	Header: textproto.MIMEHeader{
		"Content-Description": {"You'll never believe what's in this text file"},
		"X-Mailer":            {"Microsoft Outlook 15.0", "Microsoft Live Mail 42.0"},
	},
	MessageID: "h3CD-DT7rLoAw1vmpcajvIPAl-wwDfXR2MHtWID3wuQURDBKTiGUAwd6E2WBbS44QQKeXImW-axm6X0hAfcVCA==",
}

const testAttachmentJSON = `{
    "ID": "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw==",
    "Name": "croutonmail.txt",
    "Size": 77,
    "MIMEType": "text/plain",
    "KeyPackets": "wcBMA0fcZ7XLgmf2AQgAiRsOlnm1kSB4/lr7tYe6pBsRGn10GqwUhrwU5PMKOHdCgnO12jO3y3CzP0Yl/jGhAYja9wLDqH8X0sk3tY32u4Sb1Qe5IuzggAiCa4dwOJj5gEFMTHMzjIMPHR7A70XqUxMhmILye8V4KRm/j4c1sxbzA1rM3lYBumQuB5l/ck0Kgt4ZqxHVXHK5Q1l65FHhSXRj8qnunasHa30TYNzP8nmBA8BinnJxpiQ7FGc2umnUhgkFtjm5ixu9vyjr9ukwDTbwAXXfmY+o7tK7kqIXJcmTL6k2UeC6Mz1AagQtRCRtU+bv/3zGojq/trZo9lom3naIeQYa36Ketmcpj2Qwjg==",
    "Headers": {
        "content-description": "You'll never believe what's in this text file",
        "x-mailer": [
            "Microsoft Outlook 15.0",
            "Microsoft Live Mail 42.0"
        ]
    }
}
`

const testAttachmentCleartext = `cc,
dille.
`

const testAttachmentEncrypted = `wcBMA0fcZ7XLgmf2AQf/cHhfDRM9zlIuBi+h2W6DKjbbyIHMkgF6ER3JEvn/tSruUH8KTGt0N7Z+a80FFMCuXn1Y1I/nW7MVrNhGuJZAF4OymD8ugvuoAMIQX0eCYEpPXzRIWJBZg82AuowmFMsv8Dgvq4bTZq4cttI3CZcxKUNXuAearmNpmgplUKWj5USmRXK4iGB3VFGjidXkxbElrP4fD5A/rfEZ5aJgCsegqcXxX3MEjWXi9pFzgd/9phOvl1ZFm9U9hNoVAW3QsgmVeihnKaDZUyf2Qsigij21QKAUxw9U3y89eTUIqZAcmIgqeDujA3RWBgJwjtY/lOyhEmkf3AWKzehvf1xtJmCWDtJLAekuH+JfAtTQfMs5nf4zYtMahGbMkwy3Uz/jeEMYdzWY5WvshkbwvaxpqFC+11cqMLBvxik39i1xf+RORZF/91jGMCL9Z9dRMcgB`

const testCreateAttachmentBody = `{
    "Code": 1000,
    "Attachment": {"ID": "y6uKIlc2HdoHPAwPSrvf7dXoZNMYvBgxshYUN67cY5DJjL2O8NYewuvGHcYvCfd8LpEoAI_GdymO0Jr0mHlsEw=="}
}`

const testDeleteAttachmentBody = `{
    "Code": 1000
}`

func TestAttachment_UnmarshalJSON(t *testing.T) {
	att := new(Attachment)
	if err := json.Unmarshal([]byte(testAttachmentJSON), att); err != nil {
		t.Fatal("Expected no error while unmarshaling JSON, got:", err)
	}

	att.MessageID = testAttachment.MessageID // This isn't in the JSON object

	if !reflect.DeepEqual(testAttachment, att) {
		t.Errorf("Invalid attachment: expected %+v but got %+v", testAttachment, att)
	}
}

func TestClient_CreateAttachment(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/attachments"))

		contentType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Error("Expected no error while parsing request content type, got:", err)
		}
		if contentType != "multipart/form-data" {
			t.Errorf("Invalid request content type: expected %v but got %v", "multipart/form-data", contentType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		form, err := mr.ReadForm(10 * 1024)
		if err != nil {
			t.Error("Expected no error while parsing request form, got:", err)
		}
		defer Ok(t, form.RemoveAll())

		if form.Value["Filename"][0] != testAttachment.Name {
			t.Errorf("Invalid attachment filename: expected %v but got %v", testAttachment.Name, form.Value["Filename"][0])
		}
		if form.Value["MessageID"][0] != testAttachment.MessageID {
			t.Errorf("Invalid attachment message id: expected %v but got %v", testAttachment.MessageID, form.Value["MessageID"][0])
		}
		if form.Value["MIMEType"][0] != testAttachment.MIMEType {
			t.Errorf("Invalid attachment message id: expected %v but got %v", testAttachment.MIMEType, form.Value["MIMEType"][0])
		}

		dataFile, err := form.File["DataPacket"][0].Open()
		if err != nil {
			t.Error("Expected no error while opening packets file, got:", err)
		}
		defer Ok(t, dataFile.Close())

		b, err := ioutil.ReadAll(dataFile)
		if err != nil {
			t.Error("Expected no error while reading packets file, got:", err)
		}
		if string(b) != testAttachmentCleartext {
			t.Errorf("Invalid attachment packets: expected %v but got %v", testAttachment.KeyPackets, string(b))
		}

		fmt.Fprint(w, testCreateAttachmentBody)
	}))
	defer s.Close()

	r := strings.NewReader(testAttachmentCleartext) // In reality, this thing is encrypted
	created, err := c.CreateAttachment(testAttachment, r, strings.NewReader(""))
	if err != nil {
		t.Fatal("Expected no error while creating attachment, got:", err)
	}

	if created.ID != testAttachment.ID {
		t.Errorf("Invalid attachment id: expected %v but got %v", testAttachment.ID, created.ID)
	}
}

func TestClient_DeleteAttachment(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "DELETE", "/attachments/"+testAttachment.ID))

		b := &bytes.Buffer{}
		if n, _ := b.ReadFrom(r.Body); n != 0 {
			t.Fatal("expected no body but have: ", b.String())
		}

		fmt.Fprint(w, testDeleteAttachmentBody)
	}))
	defer s.Close()

	err := c.DeleteAttachment(testAttachment.ID)
	if err != nil {
		t.Fatal("Expected no error while deleting attachment, got:", err)
	}
}

func TestClient_GetAttachment(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/attachments/"+testAttachment.ID))

		fmt.Fprint(w, testAttachmentCleartext)
	}))
	defer s.Close()

	r, err := c.GetAttachment(testAttachment.ID)
	if err != nil {
		t.Fatal("Expected no error while getting attachment, got:", err)
	}
	defer r.Close() //nolint[errcheck]

	// In reality, r contains encrypted data
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal("Expected no error while reading attachment, got:", err)
	}

	if string(b) != testAttachmentCleartext {
		t.Errorf("Invalid attachment data: expected %q but got %q", testAttachmentCleartext, string(b))
	}
}

func TestAttachment_Encrypt(t *testing.T) {
	data := bytes.NewBufferString(testAttachmentCleartext)
	r, err := testAttachment.Encrypt(testPublicKeyRing, data)
	assert.Nil(t, err)
	b, err := ioutil.ReadAll(r)
	assert.Nil(t, err)

	// Result is always different, so the best way is to test it by decrypting again.
	// Another test for decrypting will help us to be sure it's working.
	dataEnc := bytes.NewBuffer(b)
	decryptAndCheck(t, dataEnc)
}

func TestAttachment_Decrypt(t *testing.T) {
	dataBytes, _ := base64.StdEncoding.DecodeString(testAttachmentEncrypted)
	dataReader := bytes.NewBuffer(dataBytes)
	decryptAndCheck(t, dataReader)
}

func decryptAndCheck(t *testing.T, data io.Reader) {
	r, err := testAttachment.Decrypt(data, testPrivateKeyRing)
	assert.Nil(t, err)
	b, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, testAttachmentCleartext, string(b))
}
