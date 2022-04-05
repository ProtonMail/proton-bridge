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

package store

import "math"

// UserID returns user ID.
func (store *Store) UserID() string {
	return store.user.ID()
}

// GetSpaceKB returns used and total space in kilo bytes (needed for IMAP
// Quota.  Quota is "in units of 1024 octets" (or KB) and PM returns bytes.
func (store *Store) GetSpaceKB() (usedSpace, maxSpace uint32, err error) {
	apiUser, err := store.client().CurrentUser(exposeContextForIMAP())
	if err != nil {
		return 0, 0, err
	}
	if apiUser.UsedSpace != nil {
		usedSpace = store.toKBandLimit(*apiUser.UsedSpace, usedSpaceType)
	}
	if apiUser.MaxSpace != nil {
		maxSpace = store.toKBandLimit(*apiUser.MaxSpace, maxSpaceType)
	}
	return
}

type spaceType string

const (
	usedSpaceType = spaceType("used")
	maxSpaceType  = spaceType("max")
)

func (store *Store) toKBandLimit(n int64, space spaceType) uint32 {
	if n < 0 {
		log.WithField("space", space).Warning("negative number of bytes")
		return uint32(0)
	}
	n /= 1024
	if n > math.MaxUint32 {
		log.WithField("space", space).Warning("too large number of bytes")
		return uint32(math.MaxUint32)
	}
	return uint32(n)
}

// GetMaxUpload returns max size of message + all attachments in bytes.
func (store *Store) GetMaxUpload() (int64, error) {
	apiUser, err := store.client().CurrentUser(exposeContextForIMAP())
	if err != nil {
		return 0, err
	}
	return apiUser.MaxUpload, nil
}
