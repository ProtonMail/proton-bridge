// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"strings"

	"github.com/cucumber/godog"
)

func SMTPActionsAuthFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^SMTP client authenticates "([^"]*)"$`, smtpClientAuthenticates)
	s.Step(`^SMTP client "([^"]*)" authenticates "([^"]*)"$`, smtpClientNamedAuthenticates)
	s.Step(`^SMTP client authenticates "([^"]*)" with address "([^"]*)"$`, smtpClientAuthenticatesWithAddress)
	s.Step(`^SMTP client "([^"]*)" authenticates "([^"]*)" with address "([^"]*)"$`, smtpClientNamedAuthenticatesWithAddress)
	s.Step(`^SMTP client authenticates "([^"]*)" with bad password$`, smtpClientAuthenticatesWithBadPassword)
	s.Step(`^SMTP client authenticates with username "([^"]*)" and password "([^"]*)"$`, smtpClientAuthenticatesWithUsernameAndPassword)
	s.Step(`^SMTP client logs out$`, smtpClientLogsOut)
	s.Step(`^SMTP client sends message$`, smtpClientSendsMessage)
	s.Step(`^SMTP client "([^"]*)" sends message$`, smtpClientNamedSendsMessage)
	s.Step(`^SMTP client sends message with bcc "([^"]*)"$`, smtpClientSendsMessageWithBCC)
	s.Step(`^SMTP client "([^"]*)" sends message with bcc "([^"]*)"$`, smtpClientNamedSendsMessageWithBCC)
	s.Step(`^SMTP client sends "([^"]*)"$`, smtpClientSendsCommand)
	s.Step(`^SMTP client sends$`, smtpClientSendsCommandMultiline)
	s.Step(`^SMTP client "([^"]*)" sends "([^"]*)"$`, smtpClientNamedSendsCommand)
}

func smtpClientAuthenticates(bddUserID string) error {
	return smtpClientNamedAuthenticates("smtp", bddUserID)
}

func smtpClientNamedAuthenticates(clientID, bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetSMTPClient(clientID).Login(account.Address(), account.BridgePassword())
	ctx.SetSMTPLastResponse(clientID, res)
	return nil
}

func smtpClientAuthenticatesWithAddress(bddUserID, bddAddressID string) error {
	return smtpClientNamedAuthenticatesWithAddress("smtp", bddUserID, bddAddressID)
}

func smtpClientNamedAuthenticatesWithAddress(clientID, bddUserID, bddAddressID string) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetSMTPClient(clientID).Login(account.Address(), account.BridgePassword())
	ctx.SetSMTPLastResponse(clientID, res)
	return nil
}

func smtpClientAuthenticatesWithBadPassword(bddUserID string) error {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return godog.ErrPending
	}
	res := ctx.GetSMTPClient("smtp").Login(account.Address(), "you shall not pass!")
	ctx.SetSMTPLastResponse("smtp", res)
	return nil
}

func smtpClientAuthenticatesWithUsernameAndPassword(bddUserID, password string) error {
	res := ctx.GetSMTPClient("smtp").Login(bddUserID, password)
	ctx.SetSMTPLastResponse("smtp", res)
	return nil
}

func smtpClientLogsOut() error {
	res := ctx.GetSMTPClient("smtp").Logout()
	ctx.SetSMTPLastResponse("smtp", res)
	return nil
}

func smtpClientSendsMessage(message *godog.DocString) error {
	return smtpClientNamedSendsMessage("smtp", message)
}

func smtpClientNamedSendsMessage(clientID string, message *godog.DocString) error {
	return smtpClientNamedSendsMessageWithBCC(clientID, "", message)
}

func smtpClientSendsMessageWithBCC(bcc string, message *godog.DocString) error {
	return smtpClientNamedSendsMessageWithBCC("smtp", bcc, message)
}

func smtpClientNamedSendsMessageWithBCC(clientID, bcc string, message *godog.DocString) error {
	res := ctx.GetSMTPClient(clientID).SendMail(strings.NewReader(message.Content), bcc)
	ctx.SetSMTPLastResponse(clientID, res)
	return nil
}

func smtpClientSendsCommand(command string) error {
	return smtpClientNamedSendsCommand("smtp", command)
}

func smtpClientSendsCommandMultiline(command *godog.DocString) error {
	return smtpClientNamedSendsCommand("smtp", command.Content)
}

func smtpClientNamedSendsCommand(clientName, command string) error {
	command = strings.ReplaceAll(command, "\\r", "\r")
	command = strings.ReplaceAll(command, "\\n", "\n")
	res := ctx.GetSMTPClient(clientName).SendCommands(command)
	ctx.SetSMTPLastResponse(clientName, res)
	return nil
}
