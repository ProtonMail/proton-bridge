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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	r "github.com/stretchr/testify/require"
)

var testBugReportReq = ReportBugReq{
	OS:            "Mac OSX",
	OSVersion:     "10.11.6",
	Browser:       "AppleMail",
	Client:        "demoapp",
	ClientVersion: "GoPMAPI_1.0.14",
	ClientType:    1,
	Title:         "Big Bug",
	Description:   "Cannot fetch new messages",
	Username:      "Apple",
	Email:         "apple@gmail.com",
}

const testBugsBody = `{
    "Code": 1000
}
`

func TestClient_BugReportWithAttachment(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "POST", "/reports/bug"))
		r.NoError(t, req.ParseMultipartForm(10*1024))

		for field, expected := range map[string]string{
			"OS":            testBugReportReq.OS,
			"OSVersion":     testBugReportReq.OSVersion,
			"Client":        testBugReportReq.Client,
			"ClientVersion": testBugReportReq.ClientVersion,
			"ClientType":    fmt.Sprintf("%d", testBugReportReq.ClientType),
			"Title":         testBugReportReq.Title,
			"Description":   testBugReportReq.Description,
			"Username":      testBugReportReq.Username,
			"Email":         testBugReportReq.Email,
		} {
			r.Equal(t, expected, req.PostFormValue(field))
		}

		attReader, err := req.MultipartForm.File["log"][0].Open()
		r.NoError(t, err)
		_, err = ioutil.ReadAll(attReader)
		r.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testBugsBody)
	}))
	defer s.Close()

	cm := newManager(newTestConfig(s.URL))

	rep := testBugReportReq
	rep.AddAttachment("log", "last.log", strings.NewReader(testAttachmentJSON))

	err := cm.ReportBug(context.Background(), rep)
	r.NoError(t, err)
}
