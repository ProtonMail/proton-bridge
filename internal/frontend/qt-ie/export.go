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
	"github.com/ProtonMail/proton-bridge/internal/transfer"
)

const (
	TypeEML  = "EML"
	TypeMBOX = "MBOX"
)

func (f *FrontendQt) LoadStructureForExport(addressOrID string) {
	var err error
	defer func() {
		if err != nil {
			f.showError(err)
			f.Qml.ExportStructureLoadFinished(false)
		} else {
			f.Qml.ExportStructureLoadFinished(true)
		}
	}()

	if f.transfer, err = f.ie.GetEMLExporter(addressOrID, ""); err != nil {
		return
	}

	f.PMStructure.Clear()
	sourceMailboxes, err := f.transfer.SourceMailboxes()
	if err != nil {
		return
	}
	for _, mbox := range sourceMailboxes {
		rule := f.transfer.GetRule(mbox)
		f.PMStructure.addEntry(newFolderInfo(mbox, rule))
	}

	f.PMStructure.transfer = f.transfer
}

func (f *FrontendQt) StartExport(rootPath, login, fileType string, attachEncryptedBody bool) {
	var target transfer.TargetProvider
	if fileType == TypeEML {
		target = transfer.NewEMLProvider(rootPath)
	} else if fileType == TypeMBOX {

		target = transfer.NewMBOXProvider(rootPath)
	} else {
		log.Errorln("Wrong file format:", fileType)
		return
	}
	f.transfer.ChangeTarget(target)
	f.transfer.SetSkipEncryptedMessages(!attachEncryptedBody)
	progress := f.transfer.Start()
	f.setProgressManager(progress)

	/*
		TODO
		f.Qml.SetProgress(0.0)
		f.Qml.SetProgressDescription(backend.ProgressInit)
		f.Qml.SetTotal(0)

		settings := backend.ExportSettings{
			FilePath:            fpath,
			Login:               login,
			AttachEncryptedBody: attachEncryptedBody,
			DateBegin:           0,
			DateEnd:             0,
			Labels:              make(map[string]string),
		}

		if fileType == "EML" {
			settings.FileTypeID = backend.EMLFormat
		} else if fileType == "MBOX" {
			settings.FileTypeID = backend.MBOXFormat
		} else {
			log.Errorln("Wrong file format:", fileType)
			return
		}

		username, _, err := backend.ExtractUsername(login)
		if err != nil {
			log.Error("qtfrontend: cannot retrieve username from alias: ", err)
			return
		}

		settings.User, err = backend.ExtractCurrentUser(username)
		if err != nil && !errors.IsCode(err, errors.ErrUnlockUser) {
			return
		}

		for _, entity := range f.PMStructure.entities {
			if entity.IsFolderSelected {
				settings.Labels[entity.FolderName] = entity.FolderId
			}
		}

		settings.DateBegin = f.PMStructure.GlobalOptions.FromDate
		settings.DateEnd = f.PMStructure.GlobalOptions.ToDate

		settings.PM = backend.NewProcessManager()
		f.setHandlers(settings.PM)

		log.Debugln("start export", settings.FilePath)
		go backend.Export(f.panicHandler, settings)
	*/
}
