// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"github.com/bradenaw/juniper/xslices"
	"github.com/vmihailenco/msgpack/v5"
)

type Data_2_4_x struct {
	Settings Settings_2_4_x
	Users    []UserData_2_4_x
}

func (data Data_2_4_x) migrate() Data {
	return Data{
		Settings: data.Settings.migrate(),
		Users:    xslices.Map(data.Users, func(user UserData_2_4_x) UserData { return user.migrate() }),
	}
}

type Settings_2_4_x struct {
	GluonDir string

	IMAPPort int
	SMTPPort int
}

func (settings Settings_2_4_x) migrate() Settings {
	newSettings := newDefaultSettings(settings.GluonDir)

	newSettings.IMAPPort = settings.IMAPPort
	newSettings.SMTPPort = settings.SMTPPort

	return newSettings
}

type UserData_2_4_x struct {
	UserID   string
	Username string

	GluonKey  string
	SplitMode bool
}

func (user UserData_2_4_x) migrate() UserData {
	var mode AddressMode

	if user.SplitMode {
		mode = SplitMode
	} else {
		mode = CombinedMode
	}

	return UserData{
		UserID:      user.UserID,
		Username:    user.Username,
		GluonKey:    []byte(user.GluonKey),
		AddressMode: mode,
	}
}

func upgrade_2_4_x(b []byte) ([]byte, error) {
	var old Data_2_4_x

	if err := msgpack.Unmarshal(b, &old); err != nil {
		return nil, err
	}

	return msgpack.Marshal(old.migrate())
}
