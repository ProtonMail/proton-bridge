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
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"
)

var testImportReqs = []*ImportMsgReq{
	{
		AddressID: "QMJs2dzTx7uqpH5PNgIzjULywU4gO9uMBhEMVFOAVJOoUml54gC0CCHtW9qYwzH-zYbZwMv3MFYncPjW1Usq7Q==",
		Body:      []byte("Hello World!"),
		Unread:    0,
		Flags:     FlagReceived | FlagImported,
		LabelIDs:  []string{ArchiveLabel},
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

func TestClient_Import(t *testing.T) { // nolint[funlen]
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/import"))

		contentType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Error("Expected no error while parsing request content type, got:", err)
		}
		if contentType != "multipart/form-data" {
			t.Errorf("Invalid request content type: expected %v but got %v", "multipart/form-data", contentType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])

		// First part is metadata.
		p, err := mr.NextPart()
		if err != nil {
			t.Error("Expected no error while reading first part of request body, got:", err)
		}

		contentDisp, params, err := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
		if err != nil {
			t.Error("Expected no error while parsing part content disposition, got:", err)
		}
		if contentDisp != "form-data" {
			t.Errorf("Invalid part content disposition: expected %v but got %v", "form-data", contentType)
		}
		if params["name"] != "Metadata" {
			t.Errorf("Invalid part name: expected %v but got %v", "Metadata", params["name"])
		}

		metadata := map[string]*ImportMsgReq{}
		if err := json.NewDecoder(p).Decode(&metadata); err != nil {
			t.Error("Expected no error while parsing metadata json, got:", err)
		}

		if len(metadata) != 1 {
			t.Errorf("Expected metadata to contain exactly one item, got %v", metadata)
		}

		req := metadata["0"]
		if metadata["0"] == nil {
			t.Errorf("Expected metadata to contain one item indexed by 0, got %v", metadata)
		}

		// No Body in metadata.
		expected := *testImportReqs[0]
		expected.Body = nil
		if !reflect.DeepEqual(&expected, req) {
			t.Errorf("Invalid message metadata: expected %v, got %v", &expected, req)
		}

		// Second part is message body.
		p, err = mr.NextPart()
		if err != nil {
			t.Error("Expected no error while reading second part of request body, got:", err)
		}

		contentDisp, params, err = mime.ParseMediaType(p.Header.Get("Content-Disposition"))
		if err != nil {
			t.Error("Expected no error while parsing part content disposition, got:", err)
		}
		if contentDisp != "form-data" {
			t.Errorf("Invalid part content disposition: expected %v but got %v", "form-data", contentType)
		}
		if params["name"] != "0" {
			t.Errorf("Invalid part name: expected %v but got %v", "0", params["name"])
		}

		b, err := ioutil.ReadAll(p)
		if err != nil {
			t.Error("Expected no error while reading second part body, got:", err)
		}

		if string(b) != string(testImportReqs[0].Body) {
			t.Errorf("Invalid message body: expected %v but got %v", string(testImportReqs[0].Body), string(b))
		}

		// No more parts.
		_, err = mr.NextPart()
		if err != io.EOF {
			t.Error("Expected no more parts but error was not EOF, got:", err)
		}

		fmt.Fprint(w, testImportBody)
	}))
	defer s.Close()

	imported, err := c.Import(testImportReqs)
	if err != nil {
		t.Fatal("Expected no error while importing, got:", err)
	}

	if len(imported) != 1 {
		t.Fatalf("Expected exactly one imported message, got %v", len(imported))
	}

	if !reflect.DeepEqual(testImportRes, imported[0]) {
		t.Errorf("Invalid response for imported message: expected %+v but got %+v", testImportRes, imported[0])
	}
}
