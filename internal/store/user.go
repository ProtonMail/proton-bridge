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

package store

import "context"

// UserID returns user ID.
func (store *Store) UserID() string {
	return store.user.ID()
}

// GetSpace returns used and total space in bytes.
func (store *Store) GetSpace() (usedSpace, maxSpace uint, err error) {
	apiUser, err := store.client().CurrentUser(context.TODO())
	if err != nil {
		return 0, 0, err
	}
	return uint(apiUser.UsedSpace), uint(apiUser.MaxSpace), nil
}

// GetMaxUpload returns max size of message + all attachments in bytes.
func (store *Store) GetMaxUpload() (int64, error) {
	apiUser, err := store.client().CurrentUser(context.TODO())
	if err != nil {
		return 0, err
	}
	return apiUser.MaxUpload, nil
}
