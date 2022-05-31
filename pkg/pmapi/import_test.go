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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"testing"

	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"
	r "github.com/stretchr/testify/require"
)

var testImportReqs = []*ImportMsgReq{
	{
		Metadata: &ImportMetadata{
			AddressID: "QMJs2dzTx7uqpH5PNgIzjULywU4gO9uMBhEMVFOAVJOoUml54gC0CCHtW9qYwzH-zYbZwMv3MFYncPjW1Usq7Q==",
			Unread:    Boolean(false),
			Flags:     FlagReceived | FlagImported,
			LabelIDs:  []string{ArchiveLabel},
		},
		Message: []byte("Hello World!"),
	},
}

const testImportBody = `{
    "Code": 1001,
    "Responses": [{
        "Name": "0",
        "Response": {"Code": 1000, "MessageID": "UKjSNz95KubYjrYmfbv1mbIfGxzY6D64mmHmVpWhkeEau-u0PIS4ru5IFMHgX6WjKpWYKCht3oiOtL5-wZChNg=="}
    }]
}`

var testImportRes = &ImportMsgRes{
	Error:     nil,
	MessageID: "UKjSNz95KubYjrYmfbv1mbIfGxzY6D64mmHmVpWhkeEau-u0PIS4ru5IFMHgX6WjKpWYKCht3oiOtL5-wZChNg==",
}

func TestClient_Import(t *testing.T) { //nolint:funlen
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "POST", "/mail/v4/messages/import"))

		contentType, params, err := pmmime.ParseMediaType(req.Header.Get("Content-Type"))
		r.NoError(t, err)
		r.Equal(t, "multipart/form-data", contentType)

		mr := multipart.NewReader(req.Body, params["boundary"])

		// First part is message body.
		p, err := mr.NextPart()
		r.NoError(t, err)

		contentDisp, params, err := pmmime.ParseMediaType(p.Header.Get("Content-Disposition"))
		r.NoError(t, err)
		r.Equal(t, "form-data", contentDisp)
		r.Equal(t, "0", params["name"])

		b, err := ioutil.ReadAll(p)
		r.NoError(t, err)
		r.Equal(t, string(testImportReqs[0].Message), string(b))

		// Second part is metadata.
		p, err = mr.NextPart()
		r.NoError(t, err)

		contentDisp, params, err = pmmime.ParseMediaType(p.Header.Get("Content-Disposition"))
		r.NoError(t, err)
		r.Equal(t, "form-data", contentDisp)
		r.Equal(t, "Metadata", params["name"])

		metadata := map[string]*ImportMetadata{}
		err = json.NewDecoder(p).Decode(&metadata)
		r.NoError(t, err)

		r.Equal(t, 1, len(metadata))

		importReq := metadata["0"]
		r.NotNil(t, req)

		expected := *testImportReqs[0].Metadata
		r.Equal(t, &expected, importReq)

		// No more parts.
		_, err = mr.NextPart()
		r.EqualError(t, err, io.EOF.Error())

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, testImportBody)
	}))
	defer s.Close()

	imported, err := c.Import(context.Background(), testImportReqs)
	r.NoError(t, err)
	r.Equal(t, 1, len(imported))
	r.Equal(t, testImportRes, imported[0])
}

func TestClientImportBigSize(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.FailNow(t, "request is not dropped")
	}))
	defer s.Close()

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const size = MaxImportMessageRequestSize + 1
	msg := make([]byte, size)
	for i := 0; i < size; i++ {
		msg[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	importRequest := []*ImportMsgReq{
		{
			Metadata: &ImportMetadata{
				AddressID: "addressID",
				Unread:    Boolean(false),
				Flags:     FlagReceived | FlagImported,
				LabelIDs:  []string{ArchiveLabel},
			},
			Message: msg,
		},
	}

	_, err := c.Import(context.Background(), importRequest)
	r.EqualError(t, err, "request size is too big")
}
