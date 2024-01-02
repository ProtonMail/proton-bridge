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
	"strconv"

	"github.com/bradenaw/juniper/xslices"
	"github.com/vmihailenco/msgpack/v5"
)

type Data_2_3_x struct {
	Settings Settings_2_3_x
	Users    []UserData_2_3_x
}

func (data Data_2_3_x) migrate() Data_2_4_x {
	return Data_2_4_x{
		Settings: data.Settings.migrate(),
		Users:    xslices.Map(data.Users, func(user UserData_2_3_x) UserData_2_4_x { return user.migrate() }),
	}
}

type Settings_2_3_x struct {
	GluonDir string

	IMAPPort string
	SMTPPort string
}

func (settings Settings_2_3_x) migrate() Settings_2_4_x {
	imapPort, err := strconv.Atoi(settings.IMAPPort)
	if err != nil {
		panic(err)
	}

	smtpPort, err := strconv.Atoi(settings.SMTPPort)
	if err != nil {
		panic(err)
	}

	return Settings_2_4_x{
		GluonDir: settings.GluonDir,

		IMAPPort: imapPort,
		SMTPPort: smtpPort,
	}
}

type UserData_2_3_x struct {
	ID   string
	Name string

	GluonKey  []byte
	SplitMode bool
}

func (user UserData_2_3_x) migrate() UserData_2_4_x {
	return UserData_2_4_x{
		UserID:    user.ID,
		Username:  user.Name,
		GluonKey:  string(user.GluonKey),
		SplitMode: user.SplitMode,
	}
}

func upgrade_2_3_x(b []byte) ([]byte, error) {
	var old Data_2_3_x

	if err := msgpack.Unmarshal(b, &old); err != nil {
		return nil, err
	}

	return msgpack.Marshal(old.migrate())
}
