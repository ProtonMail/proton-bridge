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

// Package importexport provides core functionality of Import-Export app.
package importexport

import (
	"bytes"

	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/internal/users"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "importexport") //nolint[gochecknoglobals]
)

type ImportExport struct {
	*users.Users

	config        Configer
	panicHandler  users.PanicHandler
	clientManager users.ClientManager
}

func New(
	config Configer,
	panicHandler users.PanicHandler,
	eventListener listener.Listener,
	clientManager users.ClientManager,
	credStorer users.CredentialsStorer,
) *ImportExport {
	u := users.New(config, panicHandler, eventListener, clientManager, credStorer, &storeFactory{}, false)
	return &ImportExport{
		Users: u,

		config:        config,
		panicHandler:  panicHandler,
		clientManager: clientManager,
	}
}

// ReportBug reports a new bug from the user.
func (ie *ImportExport) ReportBug(osType, osVersion, description, accountName, address, emailClient string) error {
	c := ie.clientManager.GetAnonymousClient()
	defer c.Logout()

	title := "[Import-Export] Bug"
	report := pmapi.ReportReq{
		OS:          osType,
		OSVersion:   osVersion,
		Browser:     emailClient,
		Title:       title,
		Description: description,
		Username:    accountName,
		Email:       address,
	}

	if err := c.Report(report); err != nil {
		log.Error("Reporting bug failed: ", err)
		return err
	}

	log.Info("Bug successfully reported")

	return nil
}

// ReportFile submits import report file
func (ie *ImportExport) ReportFile(osType, osVersion, accountName, address string, logdata []byte) error {
	c := ie.clientManager.GetAnonymousClient()
	defer c.Logout()

	title := "[Import-Export] report file"
	description := "An Import-Export report from the user swam down the river."

	report := pmapi.ReportReq{
		OS:          osType,
		OSVersion:   osVersion,
		Description: description,
		Title:       title,
		Username:    accountName,
		Email:       address,
	}

	report.AddAttachment("log", "report.log", bytes.NewReader(logdata))

	if err := c.Report(report); err != nil {
		log.Error("Sending report failed: ", err)
		return err
	}

	log.Info("Report successfully sent")

	return nil
}

// GetLocalImporter returns transferrer from local EML or MBOX structure to ProtonMail account.
func (ie *ImportExport) GetLocalImporter(address, path string) (*transfer.Transfer, error) {
	source := transfer.NewLocalProvider(path)
	target, err := ie.getPMAPIProvider(address)
	if err != nil {
		return nil, err
	}
	return transfer.New(ie.panicHandler, newImportMetricsManager(ie), ie.config.GetLogDir(), ie.config.GetTransferDir(), source, target)
}

// GetRemoteImporter returns transferrer from remote IMAP to ProtonMail account.
func (ie *ImportExport) GetRemoteImporter(address, username, password, host, port string) (*transfer.Transfer, error) {
	source, err := transfer.NewIMAPProvider(username, password, host, port)
	if err != nil {
		return nil, err
	}
	target, err := ie.getPMAPIProvider(address)
	if err != nil {
		return nil, err
	}
	return transfer.New(ie.panicHandler, newImportMetricsManager(ie), ie.config.GetLogDir(), ie.config.GetTransferDir(), source, target)
}

// GetEMLExporter returns transferrer from ProtonMail account to local EML structure.
func (ie *ImportExport) GetEMLExporter(address, path string) (*transfer.Transfer, error) {
	source, err := ie.getPMAPIProvider(address)
	if err != nil {
		return nil, err
	}
	target := transfer.NewEMLProvider(path)
	return transfer.New(ie.panicHandler, newExportMetricsManager(ie), ie.config.GetLogDir(), ie.config.GetTransferDir(), source, target)
}

// GetMBOXExporter returns transferrer from ProtonMail account to local MBOX structure.
func (ie *ImportExport) GetMBOXExporter(address, path string) (*transfer.Transfer, error) {
	source, err := ie.getPMAPIProvider(address)
	if err != nil {
		return nil, err
	}
	target := transfer.NewMBOXProvider(path)
	return transfer.New(ie.panicHandler, newExportMetricsManager(ie), ie.config.GetLogDir(), ie.config.GetTransferDir(), source, target)
}

func (ie *ImportExport) getPMAPIProvider(address string) (*transfer.PMAPIProvider, error) {
	user, err := ie.Users.GetUser(address)
	if err != nil {
		return nil, err
	}

	addressID, err := user.GetAddressID(address)
	if err != nil {
		log.WithError(err).Info("Address does not exist, using all addresses")
	}

	return transfer.NewPMAPIProvider(ie.clientManager, user.ID(), addressID)
}

// SetCurrentOS TODO
func (ie *ImportExport) SetCurrentOS(os string) {}
