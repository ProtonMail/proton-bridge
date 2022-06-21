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

import "encoding/json"

type Boolean bool

func (boolean *Boolean) UnmarshalJSON(b []byte) error {
	var value int
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	*boolean = Boolean(value == 1)
	return nil
}

func (boolean Boolean) MarshalJSON() ([]byte, error) {
	var value int
	if boolean {
		value = 1
	}
	return json.Marshal(value)
}
