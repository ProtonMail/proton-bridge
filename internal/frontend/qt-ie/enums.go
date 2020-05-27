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
	"github.com/therecipe/qt/core"
)

// Folder Type
const (
	FolderTypeSystem   = ""
	FolderTypeLabel    = "label"
	FolderTypeFolder   = "folder"
	FolderTypeExternal = "external"
)

// Status
const (
	StatusNoInternet          = "noInternet"
	StatusCheckingInternet    = "internetCheck"
	StatusNewVersionAvailable = "oldVersion"
	StatusUpToDate            = "upToDate"
	StatusForceUpdate         = "forceupdate"
)

// Constants for data map
const (
	// Account info
	Account = int(core.Qt__UserRole) + 1<<iota
	Status
	Password
	Aliases
	IsExpanded
	// Folder info
	FolderId
	FolderName
	FolderColor
	FolderType
	FolderEntries
	IsFolderSelected
	FolderFromDate
	FolderToDate
	TargetFolderID
	TargetLabelIDs
	// Error list
	MailSubject
	MailDate
	MailFrom
	InputFolder
	ErrorMessage
)
