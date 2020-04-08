// Copyright (c) 2020 Proton Technologies AG
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

// +build !nogui

package qt

import (
	"fmt"

	"github.com/therecipe/qt/core"
)

// The element of model.
// It contains all data for one account and its aliases.
type AccountInfo struct {
	core.QObject

	_ string `property:"account"`
	_ string `property:"userID"`
	_ string `property:"status"`
	_ string `property:"hostname"`
	_ string `property:"password"`
	_ string `property:"security"` // Deprecated, not used.
	_ int    `property:"portSMTP"`
	_ int    `property:"portIMAP"`
	_ string `property:"aliases"`
	_ bool   `property:"isExpanded"`
	_ bool   `property:"isCombinedAddressMode"`
}

// Constants for data map.
// enum-like constants in Go.
const (
	Account = int(core.Qt__UserRole) + 1<<iota
	UserID
	Status
	Hostname
	Password
	Security
	PortIMAP
	PortSMTP
	Aliases
	IsExpanded
	IsCombinedAddressMode
)

// Registration of new metatype before creating instance.
// NOTE: check it is run once per program. Write a log.
func init() {
	AccountInfo_QRegisterMetaType()
}

// Model for providing container for items (account information) to QML.
//
// QML ListView connects the model from Go and it shows item (accounts) information.
//
// Copied and edited from `github.com/therecipe/qt/internal/examples/sailfish/listview`.
type AccountsModel struct {
	core.QAbstractListModel

	// QtObject Constructor
	_ func() `constructor:"init"`

	// List of item properties.
	// All available item properties are inside the map.
	_ map[int]*core.QByteArray `property:"roles"`

	// The data storage.
	// The slice with all accounts. It is not accessed directly but using `data(index,role)`.
	_ []*AccountInfo `property:"accounts"`

	// Method for adding account.
	_ func(*AccountInfo) `slot:"addAccount"`

	// Method for retrieving account.
	_ func(row int) *AccountInfo `slot:"get"`

	// Method for login/logout the account.
	_ func(row int) `slot:"toggleIsAvailable"`

	// Method for removing account from list.
	_ func(row int) `slot:"removeAccount"`

	_ int `property:"count"`
}

// init is basically the constructor.
// Creates the map for item properties and connects the methods.
func (s *AccountsModel) init() {
	s.SetRoles(map[int]*core.QByteArray{
		Account:               NewQByteArrayFromString("account"),
		UserID:                NewQByteArrayFromString("userID"),
		Status:                NewQByteArrayFromString("status"),
		Hostname:              NewQByteArrayFromString("hostname"),
		Password:              NewQByteArrayFromString("password"),
		Security:              NewQByteArrayFromString("security"),
		PortIMAP:              NewQByteArrayFromString("portIMAP"),
		PortSMTP:              NewQByteArrayFromString("portSMTP"),
		Aliases:               NewQByteArrayFromString("aliases"),
		IsExpanded:            NewQByteArrayFromString("isExpanded"),
		IsCombinedAddressMode: NewQByteArrayFromString("isCombinedAddressMode"),
	})
	// Basic QAbstractListModel methods.
	s.ConnectData(s.data)
	s.ConnectRowCount(s.rowCount)
	s.ConnectColumnCount(s.columnCount)
	s.ConnectRoleNames(s.roleNames)
	// Custom AccountModel methods.
	s.ConnectGet(s.get)
	s.ConnectAddAccount(s.addAccount)
	s.ConnectToggleIsAvailable(s.toggleIsAvailable)
	s.ConnectRemoveAccount(s.removeAccount)
}

// get is a getter for account info pointer.
func (s *AccountsModel) get(index int) *AccountInfo {
	if index < 0 || index >= len(s.Accounts()) {
		return NewAccountInfo(nil)
	} else {
		return s.Accounts()[index]
	}
}

// data is a getter for account info data.
func (s *AccountsModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if !index.IsValid() {
		return core.NewQVariant()
	}

	if index.Row() >= len(s.Accounts()) {
		return core.NewQVariant()
	}

	var p = s.Accounts()[index.Row()]

	switch role {
	case Account:
		return NewQVariantString(p.Account())
	case UserID:
		return NewQVariantString(p.UserID())
	case Status:
		return NewQVariantString(p.Status())
	case Hostname:
		return NewQVariantString(p.Hostname())
	case Password:
		return NewQVariantString(p.Password())
	case Security:
		return NewQVariantString(p.Security())
	case PortIMAP:
		return NewQVariantInt(p.PortIMAP())
	case PortSMTP:
		return NewQVariantInt(p.PortSMTP())
	case Aliases:
		return NewQVariantString(p.Aliases())
	case IsExpanded:
		return NewQVariantBool(p.IsExpanded())
	case IsCombinedAddressMode:
		return NewQVariantBool(p.IsCombinedAddressMode())
	default:
		return core.NewQVariant()
	}
}

// rowCount returns the dimension of model: number of rows is equivalent to number of items in list.
func (s *AccountsModel) rowCount(parent *core.QModelIndex) int {
	return len(s.Accounts())
}

// columnCount returns the dimension of model: AccountsModel has only one column.
func (s *AccountsModel) columnCount(parent *core.QModelIndex) int {
	return 1
}

// roleNames returns the names of available item properties.
func (s *AccountsModel) roleNames() map[int]*core.QByteArray {
	return s.Roles()
}

// addAccount is connected to the addAccount slot.
func (s *AccountsModel) addAccount(p *AccountInfo) {
	s.BeginInsertRows(core.NewQModelIndex(), len(s.Accounts()), len(s.Accounts()))
	s.SetAccounts(append(s.Accounts(), p))
	s.SetCount(len(s.Accounts()))
	s.EndInsertRows()
}

// Method connected to toggleIsAvailable slot.
func (s *AccountsModel) toggleIsAvailable(row int) {
	var p = s.Accounts()[row]
	currentStatus := p.Status()
	if currentStatus == "active" {
		p.SetStatus("disabled")
	} else if currentStatus == "disabled" {
		p.SetStatus("active")
	} else {
		p.SetStatus("error")
	}
	var pIndex = s.Index(row, 0, core.NewQModelIndex())
	s.DataChanged(pIndex, pIndex, []int{Status})
}

// Method connected to removeAccount slot.
func (s *AccountsModel) removeAccount(row int) {
	s.BeginRemoveRows(core.NewQModelIndex(), row, row)
	s.SetAccounts(append(s.Accounts()[:row], s.Accounts()[row+1:]...))
	s.SetCount(len(s.Accounts()))
	s.EndRemoveRows()
}

// Remove all items in model.
func (s *AccountsModel) Clear() {
	s.BeginRemoveRows(core.NewQModelIndex(), 0, len(s.Accounts()))
	s.SetAccounts(s.Accounts()[0:0])
	s.SetCount(len(s.Accounts()))
	s.EndRemoveRows()
}

// Print the content of account models to console.
func (s *AccountsModel) Dump() {
	fmt.Printf("Dimensions rows %d cols %d\n", s.rowCount(nil), s.columnCount(nil))
	for iAcc := 0; iAcc < s.rowCount(nil); iAcc++ {
		var p = s.Accounts()[iAcc]
		fmt.Printf("  %d. %s\n", iAcc, p.Account())
	}
}
