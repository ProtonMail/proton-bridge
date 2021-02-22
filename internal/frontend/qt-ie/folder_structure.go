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

package qtie

// TODO:
// Proposal for new structure
// It will be a bit more memory but much better performance
// * Rules:
//   * rules []Rule /QAbstracItemModel/
//   * globalFromDate int64
//   * globalToDate int64
//   * globalLabel Mbox
//   * targetPath string
//   * filterEncryptedBodies bool
// * Rule
//   * sourceMbox: Mbox
//   * targetFolders: []Mbox /QAbstracItemModel/ (all available target folders)
//   * targetLabels: []Mbox /QAbstracItemModel/ (all available target labels)
//   * selectedLabelColors: QStringList (need reset context on change) (show label list)
//   * fromDate int64
//   * toDate int64
// * Mbox
//   * IsActive bool (show checkox)
//   * Name string (show name)
//   * Type string (show icon)
//   * Color string (show icon)
//
//  Biggest update: add folder or label for all roles update target models

import (
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/therecipe/qt/core"
)

// FolderStructure model providing container for items (folder info) to QML
//
// QML ListView connects the model from Go and it shows item (entities)
// information.
//
// Copied and edited from `github.com/therecipe/qt/internal/examples/sailfish/listview`
//
// NOTE: When implementing a model it is important to remember that QAbstractItemModel does not store any data itself !!!!
// see https://doc.qt.io/qt-5/model-view-programming.html#designing-a-model
type FolderStructure struct {
	core.QAbstractListModel

	// QtObject Constructor
	_ func() `constructor:"init"`

	// List of item properties
	//
	// All available item properties are inside the map
	_ map[int]*core.QByteArray `property:"roles"`

	// The data storage
	//
	// The slice with all entities. It is not accessed directly but using
	// `data(index,role)`
	entities      []*FolderInfo
	GlobalOptions FolderInfo

	transfer *transfer.Transfer

	// Global Folders/Labels selection flag, use setter from QML
	_ bool `property:"selectedLabels"`
	_ bool `property:"selectedFolders"`
	_ bool `property:"atLeastOneSelected"`

	// Getters (const)
	_ func() int             `slot:"getCount"`
	_ func(index int) string `slot:"getID"`
	_ func(id string) string `slot:"getName"`
	_ func(id string) string `slot:"getType"`
	_ func(id string) string `slot:"getColor"`
	_ func(id string) int64  `slot:"getFrom"`
	_ func(id string) int64  `slot:"getTo"`
	_ func(id string) string `slot:"getTargetLabelIDs"`
	_ func(name string) bool `slot:"hasFolderWithName"`
	_ func() bool            `slot:"hasTarget"`

	// TODO get folders
	// TODO get labels
	// TODO get selected labels
	// TODO get selected folder

	// Setters (emits DataChanged)
	_ func(fileType string, toSelect bool) `slot:"selectType"`
	_ func(id string, toSelect bool)       `slot:"setFolderSelection"`
	_ func(id string, target string)       `slot:"setTargetFolderID"`
	_ func(id string, label string)        `slot:"addTargetLabelID"`
	_ func(id string, label string)        `slot:"removeTargetLabelID"`
	_ func(id string, from, to int64)      `slot:"setFromToDate"`
}

// FolderInfo is the element of model
//
// It contains all data for one structure entry
type FolderInfo struct {
	/*
		FolderId         string
		FolderFullPath   string
		FolderColor      string
		FolderFullName   string
	*/
	mailbox          transfer.Mailbox // TODO how to reference from qml source mailbox to go target mailbox
	FolderType       string
	FolderEntries    int // todo remove
	IsFolderSelected bool
	FromDate         int64  // Unix seconds
	ToDate           int64  // Unix seconds
	TargetFolderID   string // target  ID TODO: this will be hash
	TargetLabelIDs   string // semicolon separated list of label ID same here
}

// Registration of new metatype before creating instance
//
// NOTE: check it is run once per program. write a log
func init() {
	FolderStructure_QRegisterMetaType()
}

// Constructor
//
// Creates the map for item properties and connects the methods
func (s *FolderStructure) init() {
	s.SetRoles(map[int]*core.QByteArray{
		FolderId:         qtcommon.NewQByteArrayFromString("folderId"),
		FolderName:       qtcommon.NewQByteArrayFromString("folderName"),
		FolderColor:      qtcommon.NewQByteArrayFromString("folderColor"),
		FolderType:       qtcommon.NewQByteArrayFromString("folderType"),
		FolderEntries:    qtcommon.NewQByteArrayFromString("folderEntries"),
		IsFolderSelected: qtcommon.NewQByteArrayFromString("isFolderSelected"),
		FolderFromDate:   qtcommon.NewQByteArrayFromString("fromDate"),
		FolderToDate:     qtcommon.NewQByteArrayFromString("toDate"),
		TargetFolderID:   qtcommon.NewQByteArrayFromString("targetFolderID"),
		TargetLabelIDs:   qtcommon.NewQByteArrayFromString("targetLabelIDs"),
	})

	// basic QAbstractListModel mehods
	s.ConnectGetCount(s.getCount)
	s.ConnectRowCount(s.rowCount)
	s.ConnectColumnCount(func(parent *core.QModelIndex) int { return 1 }) // for list  it should be always 1
	s.ConnectData(s.data)
	s.ConnectHeaderData(s.headerData)
	s.ConnectRoleNames(s.roleNames)
	// Editable QAbstractListModel needs: https://doc.qt.io/qt-5/model-view-programming.html#an-editable-model
	s.ConnectSetData(s.setData)
	s.ConnectFlags(s.flags)

	// Custom FolderStructure slots to export

	// Getters (const)
	s.ConnectGetID(func(row int) string { return s.get(row).mailbox.Hash() })
	s.ConnectGetType(func(id string) string { row := s.getRowById(id); return s.get(row).FolderType })
	s.ConnectGetName(func(id string) string { row := s.getRowById(id); return s.get(row).mailbox.Name })
	s.ConnectGetColor(func(id string) string { row := s.getRowById(id); return s.get(row).mailbox.Color })
	s.ConnectGetFrom(func(id string) int64 { row := s.getRowById(id); return s.get(row).FromDate })
	s.ConnectGetTo(func(id string) int64 { row := s.getRowById(id); return s.get(row).ToDate })
	s.ConnectGetTargetLabelIDs(func(id string) string { row := s.getRowById(id); return s.get(row).TargetLabelIDs })
	s.ConnectHasFolderWithName(s.hasFolderWithName)
	s.ConnectHasTarget(s.hasTarget)

	// Setters (emits DataChanged)
	s.ConnectSelectType(s.selectType)
	s.ConnectSetFolderSelection(s.setFolderSelection)
	s.ConnectSetTargetFolderID(s.setTargetFolderID)
	s.ConnectAddTargetLabelID(s.addTargetLabelID)
	s.ConnectRemoveTargetLabelID(s.removeTargetLabelID)
	s.ConnectSetFromToDate(s.setFromToDate)

	s.GlobalOptions = FolderInfo{
		mailbox:        transfer.Mailbox{Name: "="},
		FromDate:       0,
		ToDate:         0,
		TargetFolderID: "",
		TargetLabelIDs: "",
	}
}
