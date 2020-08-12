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
	"github.com/pkg/errors"
)

const (
	TypeEML  = "EML"
	TypeMBOX = "MBOX"
)

func (f *FrontendQt) LoadStructureForExport(addressOrID string) {
	errCode := errUnknownError
	var err error
	defer func() {
		if err != nil {
			f.showError(errCode, errors.Wrap(err, "failed to load structure for "+addressOrID))
			f.Qml.ExportStructureLoadFinished(false)
		} else {
			f.Qml.ExportStructureLoadFinished(true)
		}
	}()

	if f.transfer, err = f.ie.GetEMLExporter(addressOrID, ""); err != nil {
		// The only error can be problem to load PM user and address.
		errCode = errPMLoadFailed
		return
	}

	// Export has only one option to set time limits--by global time range.
	// In case user changes file or because of some bug global time is saved
	// to all rules, let's clear it, because there is no way to show it in
	// GUI and user would be confused and see it does not work at all.
	for _, rule := range f.transfer.GetRules() {
		isActive := rule.Active
		f.transfer.SetRule(rule.SourceMailbox, rule.TargetMailboxes, 0, 0)
		if !isActive {
			f.transfer.UnsetRule(rule.SourceMailbox)
		}
	}

	f.TransferRules.setTransfer(f.transfer)
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
}
