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

// +build build_qt

package qt

import (
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/therecipe/qt/core"
)

// QMLUserModel stores list of of users
type QMLUserModel struct {
	core.QAbstractListModel

	_ map[int]*core.QByteArray     `property:"roles"`
	_ int                          `property:"count"`
	_ func()                       `constructor:"init"`
	_ func(row int) *core.QVariant `slot:"get"`

	users []*QMLUser
}

func (um *QMLUserModel) init() {
	um.SetRoles(map[int]*core.QByteArray{
		int(core.Qt__UserRole + 1): newQByteArrayFromString("object"),
	})
	um.ConnectRowCount(um.rowCount)
	um.ConnectData(um.data)
	um.ConnectGet(um.get)
	um.users = []*QMLUser{}
	um.setCount()
}

func (um *QMLUserModel) data(index *core.QModelIndex, property int) *core.QVariant {
	if !index.IsValid() {
		return core.NewQVariant()
	}
	return um.get(index.Row())
}

func (um *QMLUserModel) get(index int) *core.QVariant {
	if index < 0 || index >= um.rowCount(nil) {
		return core.NewQVariant()
	}
	return um.users[index].ToVariant()
}

func (um *QMLUserModel) rowCount(*core.QModelIndex) int {
	return len(um.users)
}

func (um *QMLUserModel) setCount() {
	um.SetCount(len(um.users))
}

func (um *QMLUserModel) addUser(user *QMLUser) {
	um.BeginInsertRows(core.NewQModelIndex(), um.rowCount(nil), um.rowCount(nil))
	um.users = append(um.users, user)
	um.setCount()
	um.EndInsertRows()
}

func (um *QMLUserModel) removeUser(row int) {
	um.BeginRemoveRows(core.NewQModelIndex(), row, row)
	um.users = append(um.users[:row], um.users[row+1:]...)
	um.setCount()
	um.EndRemoveRows()
}

func (um *QMLUserModel) clear() {
	um.BeginRemoveRows(core.NewQModelIndex(), 0, um.rowCount(nil))
	um.users = []*QMLUser{}
	um.setCount()
	um.EndRemoveRows()
}

func (um *QMLUserModel) indexByID(id string) int {
	for i, qu := range um.users {
		if id == qu.ID {
			return i
		}
	}
	return -1
}

// QMLUser holds data, slots and signals and for user.
type QMLUser struct {
	core.QObject

	_ string   `property:"username"`
	_ string   `property:"avatarText"`
	_ bool     `property:"loggedIn"`
	_ bool     `property:"splitMode"`
	_ bool     `property:"setupGuideSeen"`
	_ float32  `property:"usedBytes"`
	_ float32  `property:"totalBytes"`
	_ string   `property:"password"`
	_ []string `property:"addresses"`

	_ func(makeItActive bool) `slot:"toggleSplitMode"`
	_ func()                  `signal:"toggleSplitModeFinished"`
	_ func()                  `slot:"logout"`
	_ func(address string)    `slot:"configureAppleMail"`

	ID string
}

func (qu *QMLUser) update(user types.User) {
	username := user.Username()
	qu.SetAvatarText(getInitials(username))
	qu.SetUsername(username)
	qu.SetLoggedIn(user.IsConnected())
	qu.SetSplitMode(!user.IsCombinedAddressMode())
	qu.SetSetupGuideSeen(true)
	qu.SetUsedBytes(1.0)      // TODO
	qu.SetTotalBytes(10000.0) // TODO
	qu.SetPassword(user.GetBridgePassword())
	qu.SetAddresses(user.GetAddresses())
}
