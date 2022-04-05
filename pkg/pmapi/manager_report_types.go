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
	"fmt"
	"io"
)

// ClientType is required by API.
const (
	EmailClientType = iota + 1
	VPNClientType
)

type reportAtt struct {
	name, mime string
	body       io.Reader
}

// ReportBugReq stores data for report.
type ReportBugReq struct {
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
func (rep *ReportBugReq) AddAttachment(name, mime string, r io.Reader) {
	rep.Attachments = append(rep.Attachments, reportAtt{name: name, mime: mime, body: r})
}

func (rep *ReportBugReq) GetMultipartFormData() map[string]string {
	return map[string]string{
		"OS":                rep.OS,
		"OSVersion":         rep.OSVersion,
		"Browser":           rep.Browser,
		"BrowserVersion":    rep.BrowserVersion,
		"BrowserExtensions": rep.BrowserExtensions,
		"Resolution":        rep.Resolution,
		"DisplayMode":       rep.DisplayMode,
		"Client":            rep.Client,
		"ClientVersion":     rep.ClientVersion,
		"ClientType":        fmt.Sprintf("%d", rep.ClientType),
		"Title":             rep.Title,
		"Description":       rep.Description,
		"Username":          rep.Username,
		"Email":             rep.Email,
		"Country":           rep.Country,
		"ISP":               rep.ISP,
		"Debug":             rep.Debug,
	}
}
