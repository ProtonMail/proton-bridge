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

//go:build build_qt
// +build build_qt

package qt

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/therecipe/qt/core"
)

func init() {
	QMLUser_QRegisterMetaType()
	QMLUserModel_QRegisterMetaType()
}

// QMLUserModel stores list of of users
type QMLUserModel struct {
	core.QAbstractListModel

	_ map[int]*core.QByteArray     `property:"roles"`
	_ int                          `property:"count"`
	_ func()                       `constructor:"init"`
	_ func(row int) *core.QVariant `slot:"get"`

	userIDs  []string
	userByID map[string]*QMLUser
	access   sync.RWMutex
	f        *FrontendQt
}

func (um *QMLUserModel) init() {
	um.access.Lock()
	defer um.access.Unlock()
	um.SetCount(0)
	um.ConnectRowCount(um.rowCount)
	um.ConnectRoleNames(um.roleNames)
	um.ConnectData(um.data)
	um.ConnectGet(um.get)
	um.ConnectCount(func() int {
		um.access.RLock()
		defer um.access.RUnlock()
		return len(um.userIDs)
	})
	um.userIDs = []string{}
	um.userByID = map[string]*QMLUser{}
}

func (um *QMLUserModel) roleNames() map[int]*core.QByteArray {
	return map[int]*core.QByteArray{
		int(core.Qt__DisplayRole): core.NewQByteArray2("user", -1),
	}
}

func (um *QMLUserModel) data(index *core.QModelIndex, property int) *core.QVariant {
	if !index.IsValid() {
		um.f.log.WithField("size", len(um.userIDs)).Info("Trying to get user by invalid index")
		return core.NewQVariant()
	}
	return um.get(index.Row())
}

func (um *QMLUserModel) get(index int) *core.QVariant {
	um.access.Lock()
	defer um.access.Unlock()
	if index < 0 || index >= len(um.userIDs) {
		um.f.log.WithField("index", index).WithField("size", len(um.userIDs)).Info("Trying to get user by wrong index")
		return core.NewQVariant()
	}

	u, err := um.getUserByID(um.userIDs[index])
	if err != nil {
		um.f.log.WithError(err).Error("Cannot get user from backend")
		return core.NewQVariant()
	}
	return u.ToVariant()
}

func (um *QMLUserModel) getUserByID(userID string) (*QMLUser, error) {
	u, ok := um.userByID[userID]
	if ok {
		return u, nil
	}

	user, err := um.f.bridge.GetUser(userID)
	if err != nil {
		return nil, err
	}

	u = newQMLUserFromBacked(um, user)
	um.userByID[userID] = u
	return u, nil
}

func (um *QMLUserModel) rowCount(*core.QModelIndex) int {
	um.access.RLock()
	defer um.access.RUnlock()
	return len(um.userIDs)
}

func (um *QMLUserModel) setCount() {
	um.SetCount(len(um.userIDs))
}

func (um *QMLUserModel) addUser(userID string) {
	um.BeginInsertRows(core.NewQModelIndex(), len(um.userIDs), len(um.userIDs))
	um.access.Lock()
	if um.indexByIDNotSafe(userID) < 0 {
		um.userIDs = append(um.userIDs, userID)
	}
	um.access.Unlock()
	um.EndInsertRows()
	um.setCount()
}

func (um *QMLUserModel) removeUser(row int) {
	um.BeginRemoveRows(core.NewQModelIndex(), row, row)
	um.access.Lock()
	id := um.userIDs[row]
	um.userIDs = append(um.userIDs[:row], um.userIDs[row+1:]...)
	delete(um.userByID, id)
	um.access.Unlock()
	um.EndRemoveRows()
	um.setCount()
}

func (um *QMLUserModel) clear() {
	um.BeginResetModel()
	um.access.Lock()
	um.userIDs = []string{}
	um.userByID = map[string]*QMLUser{}
	um.SetCount(0)
	um.access.Unlock()
	um.EndResetModel()
}

func (um *QMLUserModel) load() {
	um.clear()

	for _, user := range um.f.bridge.GetUsers() {
		um.addUser(user.ID())

		// We need mark that all existing users already saw setup
		// guide. This it is OK to construct QML here because it is in main thread.
		u, err := um.getUserByID(user.ID())
		if err != nil {
			um.f.log.WithError(err).Error("Cannot get QMLUser while loading users")
		}
		u.SetSetupGuideSeen(true)
	}

	// If there are no active accounts.
	if um.Count() == 0 {
		um.f.log.Info("No active accounts")
	}
}

func (um *QMLUserModel) userChanged(userID string) {
	defer um.f.eventListener.Emit(events.UserChangeDone, userID)

	index := um.indexByIDNotSafe(userID)
	user, err := um.f.bridge.GetUser(userID)

	if user == nil || err != nil {
		if index >= 0 { // delete existing user
			um.removeUser(index)
		}
		// if not exiting do nothing
		return
	}

	if index < 0 { // add non-existing user
		um.addUser(userID)
		return
	}

	// update exiting user
	um.userByID[userID].update(user)
}

func (um *QMLUserModel) indexByIDNotSafe(wantID string) int {
	for i, id := range um.userIDs {
		if id == wantID {
			return i
		}
	}
	return -1
}

func (um *QMLUserModel) indexByID(id string) int {
	um.access.RLock()
	defer um.access.RUnlock()

	return um.indexByIDNotSafe(id)
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
	_ func()                  `slot:"remove"`
	_ func(address string)    `slot:"configureAppleMail"`

	ID string
}

func newQMLUserFromBacked(um *QMLUserModel, user types.User) *QMLUser {
	qu := NewQMLUser(um)
	qu.ID = user.ID()

	qu.update(user)

	qu.ConnectToggleSplitMode(func(activateSplitMode bool) {
		go func() {
			defer um.f.panicHandler.HandlePanic()
			defer qu.ToggleSplitModeFinished()
			if activateSplitMode == user.IsCombinedAddressMode() {
				user.SwitchAddressMode()
			}
			qu.SetSplitMode(!user.IsCombinedAddressMode())
		}()
	})

	qu.ConnectLogout(func() {
		qu.SetLoggedIn(false)
		go func() {
			defer um.f.panicHandler.HandlePanic()
			user.Logout()
		}()
	})

	qu.ConnectRemove(func() {
		go func() {
			defer um.f.panicHandler.HandlePanic()

			// TODO: remove preferences
			if err := um.f.bridge.DeleteUser(qu.ID, false); err != nil {
				um.f.log.WithError(err).Error("Failed to remove user")
				// TODO: notification
			}
		}()
	})

	qu.ConnectConfigureAppleMail(func(address string) {
		go func() {
			defer um.f.panicHandler.HandlePanic()
			um.f.configureAppleMail(qu.ID, address)
		}()
	})

	return qu
}

func (qu *QMLUser) update(user types.User) {
	username := user.Username()
	qu.SetAvatarText(getInitials(username))
	qu.SetUsername(username)
	qu.SetLoggedIn(user.IsConnected())
	qu.SetSplitMode(!user.IsCombinedAddressMode())
	qu.SetSetupGuideSeen(false)
	qu.SetUsedBytes(float32(user.UsedBytes()))
	qu.SetTotalBytes(float32(user.TotalBytes()))
	qu.SetPassword(user.GetBridgePassword())
	qu.SetAddresses(user.GetAddresses())
}
