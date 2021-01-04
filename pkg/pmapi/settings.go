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

type UserSettings struct {
	PasswordMode int
	Email        struct {
		Value  string
		Status int
		Notify int
		Reset  int
	}
	Phone struct {
		Value  string
		Status int
		Notify int
		Reset  int
	}
	News        int
	Locale      string
	LogAuth     string
	InvoiceText string
	TOTP        int
	U2FKeys     []struct {
		Label       string
		KeyHandle   string
		Compromised int
	}
}

// GetUserSettings gets general settings.
func (c *client) GetUserSettings() (settings UserSettings, err error) {
	req, err := c.NewRequest("GET", "/settings", nil)

	if err != nil {
		return
	}

	var res struct {
		Res
		UserSettings UserSettings
	}

	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	return res.UserSettings, res.Err()
}

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
func (c *client) GetMailSettings() (settings MailSettings, err error) {
	req, err := c.NewRequest("GET", "/mail/v4/settings", nil)

	if err != nil {
		return
	}

	var res struct {
		Res
		MailSettings MailSettings
	}

	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	return res.MailSettings, res.Err()
}
