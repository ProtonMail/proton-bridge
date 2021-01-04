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
	Account  = int(core.Qt__UserRole) + 1 + iota // 256 + 1 = 257
	Status                                       // 258
	Password                                     // 259
	Aliases                                      // ...
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
	// Transfer rules mbox
	MboxSelectedIndex
	MboxIsActive
	MboxID
	MboxName
	MboxType
	MboxColor
	// Transfer Rules
	RuleTargetLabelColors
	RuleFromDate
	RuleToDate
)

const (
	// This should match enums in GuiIE.qml
	errUnknownError             = 0
	errEventAPILogout           = 1
	errUpdateAPI                = 2
	errUpdateJSON               = 3
	errUserAuth                 = 4
	errQApplication             = 18
	errEmailExportFailed        = 6
	errEmailExportMissing       = 7
	errNothingToImport          = 8
	errEmailImportFailed        = 12
	errDraftImportFailed        = 13
	errDraftLabelFailed         = 14
	errEncryptMessageAttachment = 15
	errEncryptMessage           = 16
	errNoInternetWhileImport    = 17
	errUnlockUser               = 5
	errSourceMessageNotSelected = 19
	errCannotParseMail          = 5000
	errWrongLoginOrPassword     = 5001
	errWrongServerPathOrPort    = 5002
	errWrongAuthMethod          = 5003
	errIMAPFetchFailed          = 5004
	errLocalSourceLoadFailed    = 1000
	errPMLoadFailed             = 1001
	errRemoteSourceLoadFailed   = 1002
	errLoadAccountList          = 1005
	errExit                     = 1006
	errRetry                    = 1007
	errAsk                      = 1008
	errImportFailed             = 1009
	errCreateLabelFailed        = 1010
	errCreateFolderFailed       = 1011
	errUpdateLabelFailed        = 1012
	errUpdateFolderFailed       = 1013
	errFillFolderName           = 1014
	errSelectFolderColor        = 1015
	errNoInternet               = 1016
)
