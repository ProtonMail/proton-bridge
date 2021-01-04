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

// +build !nogui

package qtie

import (
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/therecipe/qt/core"
)

func init() {
	ErrorListModel_QRegisterMetaType()
}

// ErrorListModel to sending error details to Qt
type ErrorListModel struct {
	core.QAbstractListModel

	// Qt list model
	_ func()                   `constructor:"init"`
	_ map[int]*core.QByteArray `property:"roles"`
	_ int                      `property:"count"`

	Progress *transfer.Progress
	records  []*transfer.MessageStatus
}

func (e *ErrorListModel) init() {
	e.SetRoles(map[int]*core.QByteArray{
		MailSubject:  qtcommon.NewQByteArrayFromString("mailSubject"),
		MailDate:     qtcommon.NewQByteArrayFromString("mailDate"),
		MailFrom:     qtcommon.NewQByteArrayFromString("mailFrom"),
		InputFolder:  qtcommon.NewQByteArrayFromString("inputFolder"),
		ErrorMessage: qtcommon.NewQByteArrayFromString("errorMessage"),
	})
	// basic QAbstractListModel mehods
	e.ConnectData(e.data)
	e.ConnectRowCount(e.rowCount)
	e.ConnectColumnCount(e.columnCount)
	e.ConnectRoleNames(e.roleNames)
}

func (e *ErrorListModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if !index.IsValid() {
		return core.NewQVariant()
	}

	if index.Row() >= len(e.records) {
		return core.NewQVariant()
	}

	var r = e.records[index.Row()]

	switch role {
	case MailSubject:
		return qtcommon.NewQVariantString(r.Subject)
	case MailDate:
		if r.Time.IsZero() {
			return qtcommon.NewQVariantString("Unavailable")
		}
		return qtcommon.NewQVariantString(r.Time.String())
	case MailFrom:
		return qtcommon.NewQVariantString(r.From)
	case InputFolder:
		return qtcommon.NewQVariantString(r.SourceID)
	case ErrorMessage:
		return qtcommon.NewQVariantString(r.GetErrorMessage())
	default:
		return core.NewQVariant()
	}
}

func (e *ErrorListModel) rowCount(parent *core.QModelIndex) int    { return len(e.records) }
func (e *ErrorListModel) columnCount(parent *core.QModelIndex) int { return 1 }
func (e *ErrorListModel) roleNames() map[int]*core.QByteArray      { return e.Roles() }

func (e *ErrorListModel) load() {
	if e.Progress == nil {
		log.Error("Progress not connected")
		return
	}

	e.BeginResetModel()
	e.records = e.Progress.GetFailedMessages()
	e.EndResetModel()
}
