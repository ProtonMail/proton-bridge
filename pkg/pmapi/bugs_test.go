// Copyright (c) 2021 Proton Technologies AG
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
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

var testBugReportReq = ReportReq{
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

var testBugsCrashReq = ReportReq{
	OS:            runtime.GOOS,
	Client:        "demoapp",
	ClientVersion: "GoPMAPI_1.0.14",
	ClientType:    1,
	Debug:         "main.func·001()\n/Users/sunny/Code/Go/src/scratch/stack.go:21 +0xabruntime.panic(0x80b80, 0x2101fb150)\n/usr/local/Cellar/go/1.2/libexec/src/pkg/runtime/panic.c:248 +0x106\nmain.inner()/Users/sunny/Code/Go/src/scratch/stack.go:27 +0x68\nmain.outer()\n/Users/sunny/Code/Go/src/scratch/stack.go:13 +0x1a\nmain.main()\n/Users/sunny/Code/Go/src/scratch/stack.go:9 +0x1a",
}

const testBugsBody = `{
    "Code": 1000
}
`

const testAttachmentJSONZipped = "PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\b\x00\x00\x00last.log\\Rَ\xaaH\x00}ﯨ\xf8r\x1f\xeeܖED;\xe9\ap\x03\x11\x11\x97\x0e8\x99L\xb0(\xa1\xa0\x16\x85b\x91I\xff\xfbD{\x99\xc9}\xab:K\x9d\xa4\xce\xf9\xe7\t\x00\x00z\xf6\xb4\xf7\x02z\xb7a\xe5\xd8\x04*V̭\x8d\xd1lvE}\xd6\xe3\x80\x1f\xd7nX\x9bI[\xa6\xe1a=\xd4a\xa8M\x97\xd9J\xf1F\xeb\x105U\xbd\xb0`XO\xce\xf1hu\x99q\xc3\xfe{\x11ߨ'-\v\x89Z\xa4\x9c5\xaf\xaf\xbd?>R\xd6\x11E\xf7\x1cX\xf0JpF#L\x9eE+\xbe\xe8\x1d\xee\ued2e\u007f\xde]\u06dd\xedo\x97\x87E\xa0V\xf4/$\xc2\xecK\xed\xa0\xdb&\x829\x12\xe5\x9do\xa0\xe9\x1a\xd2\x19\x1e\xf5`\x95гb\xf8\x89\x81\xb7\xa5G\x18\x95\xf3\x9d9\xe8\x93B\x17!\x1a^\xccr\xbb`\xb2\xb4\xb86\x87\xb4h\x0e\xda\xc6u<+\x9e$̓\x95\xccSo\xea\xa4\xdbH!\xe9g\x8b\xd4\b\xb3hܬ\xa6Wk\x14He\xae\x8aPU\xaa\xc1\xee$\xfbH\xb3\xab.I\f<\x89\x06q\xe3-3-\x99\xcdݽ\xe5v\x99\xedn\xac\xadn\xe8Rp=\xb4nJ\xed\xd5\r\x8d\xde\x06Ζ\xf6\xb3\x01\x94\xcb\xf6\xd4\x19r\xe1\xaa$4+\xeaW\xa6F\xfa0\x97\x9cD\f\x8e\xd7\xd6z\v,G\xf3e2\xd4\xe6V\xba\v\xb6\xd9\xe8\xca*\x16\x95V\xa4J\xfbp\xddmF\x8c\x9a\xc6\xc8Č-\xdb\v\xf6\xf5\xf9\x02*\x15e\x874\xc9\xe7\"\xa3\x1an\xabq}ˊq\x957\xd3\xfd\xa91\x82\xe0Lß\\\x17\x8e\x9e_\xed`\t\xe9~5̕\x03\x9a\f\xddN6\xa2\xc4\x17\xdb\xc9V\x1c~\x9e\xea\xbe\xda-xv\xed\x8b\xe2\xc8ǄS\x95E6\xf2\xc3H\x1d:HPx\xc9\x14\xbfɒ\xff\xea\xb4P\x14\xa3\xe2\xfe\xfd\x1f+z\x80\x903\x81\x98\xf8\x15\xa3\x12\x16\xf8\"0g\xf7~B^\xfd \x040T\xa3\x02\x9c\x10\xc1\xa8F\xa0I#\xf1\xa3\x04\x98\x01\x91\xe2\x12\xdc;\x06gL\xd0g\xc0\xe3\xbd\xf6\xd7}&\xa8轀?\xbfяy`X\xf0\x92\x9f\x05\xf0*A8ρ\xac=K\xff\xf3\xfe\xa6Z\xe1\x1a\x017\xc2\x04\f\x94g\xa9\xf7-\xfb\xebqz\u007fz\u007f\xfa7\x00\x00\xff\xffPK\a\b\xf5\\\v\xe5I\x02\x00\x00\r\x03\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\xf5\\\v\xe5I\x02\x00\x00\r\x03\x00\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00last.logPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x006\x00\x00\x00\u007f\x02\x00\x00\x00\x00" //nolint[misspell]

func TestClient_BugReportWithAttachment(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/reports/bug"))
		Ok(t, isAuthReq(r, testUID, testAccessToken))

		Ok(t, r.ParseMultipartForm(10*1024))

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
			if r.PostFormValue(field) != expected {
				t.Errorf("Field %q has %q but expected %q", field, r.PostFormValue(field), expected)
			}
		}

		attReader, err := r.MultipartForm.File["log"][0].Open()
		Ok(t, err)

		log, err := ioutil.ReadAll(attReader)
		Ok(t, err)

		Equals(t, []byte(testAttachmentJSONZipped), log)

		fmt.Fprint(w, testBugsBody)
	}))
	defer s.Close()
	c.uid = testUID
	c.accessToken = testAccessToken

	rep := testBugReportReq
	rep.AddAttachment("log", "last.log", strings.NewReader(testAttachmentJSON))

	Ok(t, c.Report(rep))
}

func TestClient_BugReport(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/reports/bug"))
		Ok(t, isAuthReq(r, testUID, testAccessToken))

		var bugsReportReq ReportReq
		Ok(t, json.NewDecoder(r.Body).Decode(&bugsReportReq))
		Equals(t, testBugReportReq, bugsReportReq)

		fmt.Fprint(w, testBugsBody)
	}))
	defer s.Close()
	c.uid = testUID
	c.accessToken = testAccessToken

	r := ReportReq{
		OS:          testBugReportReq.OS,
		OSVersion:   testBugReportReq.OSVersion,
		Browser:     testBugReportReq.Browser,
		Title:       testBugReportReq.Title,
		Description: testBugReportReq.Description,
		Username:    testBugReportReq.Username,
		Email:       testBugReportReq.Email,
	}

	Ok(t, c.Report(r))
}

func TestClient_BugsCrash(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/reports/crash"))
		Ok(t, isAuthReq(r, testUID, testAccessToken))

		var bugsCrashReq ReportReq
		Ok(t, json.NewDecoder(r.Body).Decode(&bugsCrashReq))
		Equals(t, testBugsCrashReq, bugsCrashReq)

		fmt.Fprint(w, testBugsBody)
	}))
	defer s.Close()
	c.uid = testUID
	c.accessToken = testAccessToken

	Ok(t, c.ReportCrash(testBugsCrashReq.Debug))
}
