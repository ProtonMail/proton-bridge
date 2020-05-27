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

package qtie

import (
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/therecipe/qt/core"
)

// ErrorDetail stores information about email and error
type ErrorDetail struct {
	MailSubject, MailDate, MailFrom, InputFolder, ErrorMessage string
}

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

	Details []*ErrorDetail
}

func (s *ErrorListModel) init() {
	s.SetRoles(map[int]*core.QByteArray{
		MailSubject:  qtcommon.NewQByteArrayFromString("mailSubject"),
		MailDate:     qtcommon.NewQByteArrayFromString("mailDate"),
		MailFrom:     qtcommon.NewQByteArrayFromString("mailFrom"),
		InputFolder:  qtcommon.NewQByteArrayFromString("inputFolder"),
		ErrorMessage: qtcommon.NewQByteArrayFromString("errorMessage"),
	})
	// basic QAbstractListModel mehods
	s.ConnectData(s.data)
	s.ConnectRowCount(s.rowCount)
	s.ConnectColumnCount(s.columnCount)
	s.ConnectRoleNames(s.roleNames)
}

func (s *ErrorListModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if !index.IsValid() {
		return core.NewQVariant()
	}

	if index.Row() >= len(s.Details) {
		return core.NewQVariant()
	}

	var p = s.Details[index.Row()]

	switch role {
	case MailSubject:
		return qtcommon.NewQVariantString(p.MailSubject)
	case MailDate:
		return qtcommon.NewQVariantString(p.MailDate)
	case MailFrom:
		return qtcommon.NewQVariantString(p.MailFrom)
	case InputFolder:
		return qtcommon.NewQVariantString(p.InputFolder)
	case ErrorMessage:
		return qtcommon.NewQVariantString(p.ErrorMessage)
	default:
		return core.NewQVariant()
	}
}

func (s *ErrorListModel) rowCount(parent *core.QModelIndex) int    { return len(s.Details) }
func (s *ErrorListModel) columnCount(parent *core.QModelIndex) int { return 1 }
func (s *ErrorListModel) roleNames() map[int]*core.QByteArray      { return s.Roles() }

// Add more errors to list
func (s *ErrorListModel) Add(more []*ErrorDetail) {
	s.BeginInsertRows(core.NewQModelIndex(), len(s.Details), len(s.Details))
	s.Details = append(s.Details, more...)
	s.SetCount(len(s.Details))
	s.EndInsertRows()
}

// Clear removes all items in model
func (s *ErrorListModel) Clear() {
	s.BeginRemoveRows(core.NewQModelIndex(), 0, len(s.Details))
	s.Details = s.Details[0:0]
	s.SetCount(len(s.Details))
	s.EndRemoveRows()
}

func (s *ErrorListModel) load(importLogFileName string) {
	/*
		err := backend.LoopDetailsInFile(importLogFileName, func(d *backend.MessageDetails) {
			if d.MessageID != "" { // imported ok
				return
			}
			ed := &ErrorDetail{
				MailSubject:  d.Subject,
				MailDate:     d.Time,
				MailFrom:     d.From,
				InputFolder:  d.Folder,
				ErrorMessage: d.Error,
			}
			s.Add([]*ErrorDetail{ed})
		})
		if err != nil {
			log.Errorf("load import report from %q: %v", importLogFileName, err)
		}
	*/
}
