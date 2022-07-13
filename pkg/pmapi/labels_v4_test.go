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
	"fmt"
	"net/http"
	"reflect"
	"testing"

	r "github.com/stretchr/testify/require"
)

const testFoldersBody = `{
    "Labels": [
        {
            "ID": "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==",
            "Name": "CroutonMail is awesome :)",
            "Color": "#7272a7",
            "Display": 0,
            "Order": 1,
            "Type": 3
        },
        {
            "ID": "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==",
            "Name": "Royal sausage",
            "Color": "#cf5858",
            "Display": 1,
            "Order": 2,
            "Type": 3
        }
    ],
    "Code": 1000
}
`

var testFolders = []*Label{
	{ID: "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==", Name: "CroutonMail is awesome :)", Color: "#7272a7", Order: 1, Display: 0, Type: LabelTypeV4Folder},
	{ID: "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==", Name: "Royal sausage", Color: "#cf5858", Order: 2, Display: 1, Type: LabelTypeV4Folder},
}

func TestClient_ListLabelsOnly(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/core/v4/labels?Type=1"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testLabelsBody)
	}))
	defer s.Close()

	labels, err := c.ListLabelsOnly(context.Background())
	r.NoError(t, err)
	r.Equal(t, testLabels, labels)
}

func TestClient_ListFoldersOnly(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/core/v4/labels?Type=3"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testFoldersBody)
	}))
	defer s.Close()

	folders, err := c.ListFoldersOnly(context.Background())
	r.NoError(t, err)
	r.Equal(t, testFolders, folders)
}
func TestClient_CreateLabelV4(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "POST", "/core/v4/labels"))

		body := &bytes.Buffer{}
		_, err := body.ReadFrom(req.Body)
		r.NoError(t, err)

		if bytes.Contains(body.Bytes(), []byte("Order")) {
			t.Fatal("Body contains `Order`: ", body.String())
		}

		var labelReq LabelReq
		err = json.NewDecoder(body).Decode(&labelReq)
		r.NoError(t, err)
		r.Equal(t, testLabelReq.Label, labelReq.Label)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testCreateLabelBody)
	}))
	defer s.Close()

	created, err := c.CreateLabelV4(context.Background(), testLabelReq.Label)
	r.NoError(t, err)

	if !reflect.DeepEqual(created, testLabelCreated) {
		t.Fatalf("Invalid created label: expected %+v, got %+v", testLabelCreated, created)
	}
}

func TestClient_CreateEmptyLabelV4(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		r.Fail(t, "API should not be called")
	}))
	defer s.Close()

	_, err := c.CreateLabelV4(context.Background(), &Label{})
	r.EqualError(t, err, "name is required")
}

func TestClient_UpdateLabelV4(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "PUT", "/core/v4/labels/"+testLabelCreated.ID))

		var labelReq LabelReq
		err := json.NewDecoder(req.Body).Decode(&labelReq)
		r.NoError(t, err)
		r.Equal(t, testLabelCreated, labelReq.Label)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testCreateLabelBody)
	}))
	defer s.Close()

	updated, err := c.UpdateLabelV4(context.Background(), testLabelCreated)
	r.NoError(t, err)

	if !reflect.DeepEqual(updated, testLabelCreated) {
		t.Fatalf("Invalid updated label: expected %+v, got %+v", testLabelCreated, updated)
	}
}

func TestClient_UpdateLabelToEmptyNameV4(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		r.Fail(t, "API should not be called")
	}))
	defer s.Close()

	_, err := c.UpdateLabelV4(context.Background(), &Label{ID: "label"})
	r.EqualError(t, err, "name is required")
}

func TestClient_DeleteLabelV4(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "DELETE", "/core/v4/labels/"+testLabelCreated.ID))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testDeleteLabelBody)
	}))
	defer s.Close()

	err := c.DeleteLabelV4(context.Background(), testLabelCreated.ID)
	r.NoError(t, err)
}
