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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

const testLabelsBody = `{
    "Labels": [
        {
            "ID": "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==",
            "Name": "CroutonMail is awesome :)",
            "Color": "#7272a7",
            "Display": 0,
            "Order": 1,
            "Type": 1
        },
        {
            "ID": "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==",
            "Name": "Royal sausage",
            "Color": "#cf5858",
            "Display": 1,
            "Order": 2,
            "Type": 1
        }
    ],
    "Code": 1000
}
`

var testLabels = []*Label{
	{ID: "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==", Name: "CroutonMail is awesome :)", Color: "#7272a7", Order: 1, Display: 0, Type: LabelTypeMailbox},
	{ID: "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==", Name: "Royal sausage", Color: "#cf5858", Order: 2, Display: 1, Type: LabelTypeMailbox},
}

var testLabelReq = LabelReq{&Label{
	Name:    "sava",
	Color:   "#c26cc7",
	Display: 1,
}}

const testCreateLabelBody = `{
    "Label": {
        "ID": "otkpEZzG--8dMXvwyLXLQWB72hhBhNGzINjH14rUDfywvOyeN01cDxDrS3Koifxf6asA7Xcwtldm0r_MCmWiAQ==",
        "Name": "sava",
        "Color": "#c26cc7",
        "Display": 1,
        "Order": 3,
        "Type": 1
    },
    "Code": 1000
}
`

var testLabelCreated = &Label{
	ID:      "otkpEZzG--8dMXvwyLXLQWB72hhBhNGzINjH14rUDfywvOyeN01cDxDrS3Koifxf6asA7Xcwtldm0r_MCmWiAQ==",
	Name:    "sava",
	Color:   "#c26cc7",
	Order:   3,
	Display: 1,
	Type:    LabelTypeMailbox,
}

const testDeleteLabelBody = `{
    "Code": 1000
}
`

func TestClient_ListLabels(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/labels?1"))

		fmt.Fprint(w, testLabelsBody)
	}))
	defer s.Close()

	labels, err := c.ListLabels()
	if err != nil {
		t.Fatal("Expected no error while listing labels, got:", err)
	}

	if !reflect.DeepEqual(labels, testLabels) {
		for i, l := range testLabels {
			t.Errorf("expected %d: %#v\n", i, l)
		}
		for i, l := range labels {
			t.Errorf("got %d: %#v\n", i, l)
		}
		t.Fatalf("Not same")
	}
}

func TestClient_CreateLabel(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/labels"))

		body := &bytes.Buffer{}
		_, err := body.ReadFrom(r.Body)
		Ok(t, err)

		if bytes.Contains(body.Bytes(), []byte("Order")) {
			t.Fatal("Body contains `Order`: ", body.String())
		}

		var labelReq LabelReq
		if err := json.NewDecoder(body).Decode(&labelReq); err != nil {
			t.Error("Expecting no error while reading request body, got:", err)
		}
		if !reflect.DeepEqual(testLabelReq.Label, labelReq.Label) {
			t.Errorf("Invalid label request: expected %+v but got %+v", testLabelReq.Label, labelReq.Label)
		}

		fmt.Fprint(w, testCreateLabelBody)
	}))
	defer s.Close()

	created, err := c.CreateLabel(testLabelReq.Label)
	if err != nil {
		t.Fatal("Expected no error while creating label, got:", err)
	}

	if !reflect.DeepEqual(created, testLabelCreated) {
		t.Fatalf("Invalid created label: expected %+v, got %+v", testLabelCreated, created)
	}
}

func TestClient_UpdateLabel(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "PUT", "/labels/"+testLabelCreated.ID))

		var labelReq LabelReq
		if err := json.NewDecoder(r.Body).Decode(&labelReq); err != nil {
			t.Error("Expecting no error while reading request body, got:", err)
		}
		if !reflect.DeepEqual(testLabelCreated, labelReq.Label) {
			t.Errorf("Invalid label request: expected %+v but got %+v", testLabelCreated, labelReq.Label)
		}

		fmt.Fprint(w, testCreateLabelBody)
	}))
	defer s.Close()

	updated, err := c.UpdateLabel(testLabelCreated)
	if err != nil {
		t.Fatal("Expected no error while updating label, got:", err)
	}

	if !reflect.DeepEqual(updated, testLabelCreated) {
		t.Fatalf("Invalid updated label: expected %+v, got %+v", testLabelCreated, updated)
	}
}

func TestClient_DeleteLabel(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "DELETE", "/labels/"+testLabelCreated.ID))

		fmt.Fprint(w, testDeleteLabelBody)
	}))
	defer s.Close()

	err := c.DeleteLabel(testLabelCreated.ID)
	if err != nil {
		t.Fatal("Expected no error while deleting label, got:", err)
	}
}
