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
	"github.com/pkg/errors"

	"github.com/ProtonMail/proton-bridge/internal/transfer"
)

// wrapper for QML
func (f *FrontendQt) setupAndLoadForImport(isFromIMAP bool, sourcePath, sourceEmail, sourcePassword, sourceServer, sourcePort, targetAddress string) {
	errCode := errUnknownError
	var err error
	defer func() {
		if err != nil {
			f.showError(errCode, err)
			f.Qml.ImportStructuresLoadFinished(false)
		} else {
			f.Qml.ImportStructuresLoadFinished(true)
		}
	}()

	if isFromIMAP {
		f.transfer, err = f.ie.GetRemoteImporter(targetAddress, sourceEmail, sourcePassword, sourceServer, sourcePort)
		if err != nil {
			switch {
			case errors.Is(err, &transfer.ErrIMAPConnection{}):
				errCode = errWrongServerPathOrPort
			case errors.Is(err, &transfer.ErrIMAPAuth{}):
				errCode = errWrongLoginOrPassword
			case errors.Is(err, &transfer.ErrIMAPAuthMethod{}):
				errCode = errWrongAuthMethod
			default:
				errCode = errRemoteSourceLoadFailed
			}
			return
		}
	} else {
		f.transfer, err = f.ie.GetLocalImporter(targetAddress, sourcePath)
		if err != nil {
			// The only error can be problem to load PM user and address.
			errCode = errPMLoadFailed
			return
		}
	}

	if err := f.loadStructuresForImport(); err != nil {
		return
	}
}

func (f *FrontendQt) loadStructuresForImport() error {
	f.TransferRules.setTransfer(f.transfer)

	return nil
}

func (f *FrontendQt) StartImport(email string) { // TODO email not needed
	log.Trace("Starting import")

	f.Qml.SetProgressDescription("init") // TODO use const
	f.Qml.SetProgressImported(0)
	f.Qml.SetProgressSkipped(0)
	f.Qml.SetProgressFails(0)
	f.Qml.SetProgress(0.0)
	f.Qml.SetTotal(1)
	f.Qml.SetImportLogFileName("")

	progress := f.transfer.Start()

	f.Qml.SetImportLogFileName(progress.FileReport())
	f.setProgressManager(progress)
}
