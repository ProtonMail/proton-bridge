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

	"github.com/go-resty/resty/v2"
)

type MailSettings struct {
	DisplayName        string
	Signature          string `json:",omitempty"`
	Theme              string `json:",omitempty"`
	AutoSaveContacts   int
	AutoWildcardSearch int
	ComposerMode       int
	MessageButtons     int
	ShowImages         int
	ShowMoved          int
	ViewMode           int
	ViewLayout         int
	SwipeLeft          int
	SwipeRight         int
	AlsoArchive        int
	Hotkeys            int
	PMSignature        int
	ImageProxy         int
	TLS                int
	RightToLeft        int
	AttachPublicKey    int
	Sign               int
	PGPScheme          PackageFlag
	PromptPin          int
	Autocrypt          int
	NumMessagePerPage  int
	DraftMIMEType      string
	ReceiveMIMEType    string
	ShowMIMEType       string

	// Undocumented -- there's only `null` in example:
	// AutoResponder string
}

// GetMailSettings gets contact details specified by contact ID.
func (c *client) GetMailSettings(ctx context.Context) (settings MailSettings, err error) {
	var res struct {
		MailSettings MailSettings
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/mail/v4/settings")
	}); err != nil {
		return MailSettings{}, err
	}

	return res.MailSettings, nil
}
