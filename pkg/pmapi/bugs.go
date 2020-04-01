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
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"runtime"
	"strings"
)

// ClientType is required by API.
const (
	EmailClientType = iota + 1
	VPNClientType
)

type reportAtt struct {
	name, filename string
	body           io.Reader
}

// ReportReq stores data for report.
type ReportReq struct {
	OS                string      `json:",omitempty"`
	OSVersion         string      `json:",omitempty"`
	Browser           string      `json:",omitempty"`
	BrowserVersion    string      `json:",omitempty"`
	BrowserExtensions string      `json:",omitempty"`
	Resolution        string      `json:",omitempty"`
	DisplayMode       string      `json:",omitempty"`
	Client            string      `json:",omitempty"`
	ClientVersion     string      `json:",omitempty"`
	ClientType        int         `json:",omitempty"`
	Title             string      `json:",omitempty"`
	Description       string      `json:",omitempty"`
	Username          string      `json:",omitempty"`
	Email             string      `json:",omitempty"`
	Country           string      `json:",omitempty"`
	ISP               string      `json:",omitempty"`
	Debug             string      `json:",omitempty"`
	Attachments       []reportAtt `json:",omitempty"`
}

// AddAttachment to report.
func (rep *ReportReq) AddAttachment(name, filename string, r io.Reader) {
	rep.Attachments = append(rep.Attachments, reportAtt{name: name, filename: filename, body: r})
}

func writeMultipartReport(w *multipart.Writer, rep *ReportReq) error { // nolint[funlen]
	fieldData := map[string]string{
		"OS":                rep.OS,
		"OSVersion":         rep.OSVersion,
		"Browser":           rep.Browser,
		"BrowserVersion":    rep.BrowserVersion,
		"BrowserExtensions": rep.BrowserExtensions,
		"Resolution":        rep.Resolution,
		"DisplayMode":       rep.DisplayMode,
		"Client":            rep.Client,
		"ClientVersion":     rep.ClientVersion,
		"ClientType":        "1",
		"Title":             rep.Title,
		"Description":       rep.Description,
		"Username":          rep.Username,
		"Email":             rep.Email,
		"Country":           rep.Country,
		"ISP":               rep.ISP,
		"Debug":             rep.Debug,
	}

	for field, data := range fieldData {
		if data == "" {
			continue
		}
		if err := w.WriteField(field, data); err != nil {
			return err
		}
	}

	quoteEscaper := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

	for _, att := range rep.Attachments {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				quoteEscaper.Replace(att.name), quoteEscaper.Replace(att.filename+".zip")))
		h.Set("Content-Type", "application/octet-stream")
		//h.Set("Content-Transfere-Encoding", "base64")
		attWr, err := w.CreatePart(h)
		if err != nil {
			return err
		}

		zipArch := zip.NewWriter(attWr)
		zipWr, err := zipArch.Create(att.filename)
		//b64 := base64.NewEncoder(base64.StdEncoding, zipWr)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipWr, att.body)
		if err != nil {
			return err
		}
		err = zipArch.Close()
		//err = b64.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Report sends request as json or multipart (if has attachment).
func (c *Client) Report(rep ReportReq) (err error) {
	rep.Client = c.cm.GetConfig().ClientID
	rep.ClientVersion = c.cm.GetConfig().AppVersion
	rep.ClientType = EmailClientType

	var req *http.Request
	var w *MultipartWriter
	if len(rep.Attachments) > 0 {
		req, w, err = NewMultipartRequest("POST", "/reports/bug")
	} else {
		req, err = NewJSONRequest("POST", "/reports/bug", rep)
	}
	if err != nil {
		return
	}

	var res Res
	done := make(chan error, 1)
	go func() {
		done <- c.DoJSON(req, &res)
	}()

	if w != nil {
		err = writeMultipartReport(w.Writer, &rep)
		if err != nil {
			c.log.Errorln("report write: ", err)
			return
		}
		err = w.Close()
		if err != nil {
			c.log.Errorln("report close: ", err)
			return
		}
	}

	if err = <-done; err != nil {
		return
	}

	return res.Err()
}

// ReportBug is old. Use Report instead.
func (c *Client) ReportBug(os, osVersion, title, description, username, email string) (err error) {
	return c.ReportBugWithEmailClient(os, osVersion, title, description, username, email, "")
}

// ReportBugWithEmailClient is old. Use Report instead.
func (c *Client) ReportBugWithEmailClient(os, osVersion, title, description, username, email, emailClient string) (err error) {
	bugReq := ReportReq{
		OS:          os,
		OSVersion:   osVersion,
		Browser:     emailClient,
		Title:       title,
		Description: description,
		Username:    username,
		Email:       email,
	}

	return c.Report(bugReq)
}

// ReportCrash is old. Use sentry instead.
func (c *Client) ReportCrash(stacktrace string) (err error) {
	crashReq := ReportReq{
		Client:        c.cm.GetConfig().ClientID,
		ClientVersion: c.cm.GetConfig().AppVersion,
		ClientType:    EmailClientType,
		OS:            runtime.GOOS,
		Debug:         stacktrace,
	}
	req, err := NewJSONRequest("POST", "/reports/crash", crashReq)
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
