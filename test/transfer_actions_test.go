// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package tests

import (
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/cucumber/godog"
)

func TransferActionsFeatureContext(s *godog.ScenarioContext) {
	s.Step(`^user "([^"]*)" imports local files$`, userImportsLocalFiles)
	s.Step(`^user "([^"]*)" imports local files with rules$`, userImportsLocalFilesWithRules)
	s.Step(`^user "([^"]*)" imports local files to address "([^"]*)"$`, userImportsLocalFilesToAddress)
	s.Step(`^user "([^"]*)" imports local files to address "([^"]*)" with rules$`, userImportsLocalFilesToAddressWithRules)
	s.Step(`^user "([^"]*)" imports remote messages$`, userImportsRemoteMessages)
	s.Step(`^user "([^"]*)" imports remote messages with rules$`, userImportsRemoteMessagesWithRules)
	s.Step(`^user "([^"]*)" imports remote messages to address "([^"]*)"$`, userImportsRemoteMessagesToAddress)
	s.Step(`^user "([^"]*)" imports remote messages to address "([^"]*)" with rules$`, userImportsRemoteMessagesToAddressWithRules)
	s.Step(`^user "([^"]*)" exports to EML files$`, userExportsToEMLFiles)
	s.Step(`^user "([^"]*)" exports to EML files with rules$`, userExportsToEMLFilesWithRules)
	s.Step(`^user "([^"]*)" exports address "([^"]*)" to EML files$`, userExportsAddressToEMLFiles)
	s.Step(`^user "([^"]*)" exports address "([^"]*)" to EML files with rules$`, userExportsAddressToEMLFilesWithRules)
	s.Step(`^user "([^"]*)" exports to MBOX files$`, userExportsToMBOXFiles)
	s.Step(`^user "([^"]*)" exports to MBOX files with rules$`, userExportsToMBOXFilesWithRules)
	s.Step(`^user "([^"]*)" exports address "([^"]*)" to MBOX files$`, userExportsAddressToMBOXFiles)
	s.Step(`^user "([^"]*)" exports address "([^"]*)" to MBOX files with rules$`, userExportsAddressToMBOXFilesWithRules)
}

// Local import.

func userImportsLocalFiles(bddUserID string) error {
	return userImportsLocalFilesToAddressWithRules(bddUserID, "", nil)
}

func userImportsLocalFilesWithRules(bddUserID string, rules *godog.Table) error {
	return userImportsLocalFilesToAddressWithRules(bddUserID, "", rules)
}

func userImportsLocalFilesToAddress(bddUserID, bddAddressID string) error {
	return userImportsLocalFilesToAddressWithRules(bddUserID, bddAddressID, nil)
}

func userImportsLocalFilesToAddressWithRules(bddUserID, bddAddressID string, rules *godog.Table) error {
	return doTransfer(bddUserID, bddAddressID, rules, func(username, address string) (*transfer.Transfer, error) {
		path := ctx.GetTransferLocalRootForImport()
		return ctx.GetImportExport().GetLocalImporter(username, address, path)
	})
}

// Remote import.

func userImportsRemoteMessages(bddUserID string) error {
	return userImportsRemoteMessagesToAddressWithRules(bddUserID, "", nil)
}

func userImportsRemoteMessagesWithRules(bddUserID string, rules *godog.Table) error {
	return userImportsRemoteMessagesToAddressWithRules(bddUserID, "", rules)
}

func userImportsRemoteMessagesToAddress(bddUserID, bddAddressID string) error {
	return userImportsRemoteMessagesToAddressWithRules(bddUserID, bddAddressID, nil)
}

func userImportsRemoteMessagesToAddressWithRules(bddUserID, bddAddressID string, rules *godog.Table) error {
	return doTransfer(bddUserID, bddAddressID, rules, func(username, address string) (*transfer.Transfer, error) {
		imapServer := ctx.GetTransferRemoteIMAPServer()
		return ctx.GetImportExport().GetRemoteImporter(username, address, imapServer.Username, imapServer.Password, imapServer.Host, imapServer.Port)
	})
}

// EML export.

func userExportsToEMLFiles(bddUserID string) error {
	return userExportsAddressToEMLFilesWithRules(bddUserID, "", nil)
}

func userExportsToEMLFilesWithRules(bddUserID string, rules *godog.Table) error {
	return userExportsAddressToEMLFilesWithRules(bddUserID, "", rules)
}

func userExportsAddressToEMLFiles(bddUserID, bddAddressID string) error {
	return userExportsAddressToEMLFilesWithRules(bddUserID, bddAddressID, nil)
}

func userExportsAddressToEMLFilesWithRules(bddUserID, bddAddressID string, rules *godog.Table) error {
	return doTransfer(bddUserID, bddAddressID, rules, func(username, address string) (*transfer.Transfer, error) {
		path := ctx.GetTransferLocalRootForExport()
		return ctx.GetImportExport().GetEMLExporter(username, address, path)
	})
}

// MBOX export.

func userExportsToMBOXFiles(bddUserID string) error {
	return userExportsAddressToMBOXFilesWithRules(bddUserID, "", nil)
}

func userExportsToMBOXFilesWithRules(bddUserID string, rules *godog.Table) error {
	return userExportsAddressToMBOXFilesWithRules(bddUserID, "", rules)
}

func userExportsAddressToMBOXFiles(bddUserID, bddAddressID string) error {
	return userExportsAddressToMBOXFilesWithRules(bddUserID, bddAddressID, nil)
}

func userExportsAddressToMBOXFilesWithRules(bddUserID, bddAddressID string, rules *godog.Table) error {
	return doTransfer(bddUserID, bddAddressID, rules, func(username, address string) (*transfer.Transfer, error) {
		path := ctx.GetTransferLocalRootForExport()
		return ctx.GetImportExport().GetMBOXExporter(username, address, path)
	})
}

// Helpers.

func doTransfer(bddUserID, bddAddressID string, rules *godog.Table, getTransferrer func(string, string) (*transfer.Transfer, error)) error {
	account := ctx.GetTestAccountWithAddress(bddUserID, bddAddressID)
	if account == nil {
		return godog.ErrPending
	}
	transferrer, err := getTransferrer(account.Username(), account.Address())
	if err != nil {
		return internalError(err, "failed to init transfer")
	}
	if err := setRules(transferrer, rules); err != nil {
		return internalError(err, "failed to set rules")
	}
	transferrer.SetSkipEncryptedMessages(ctx.GetTransferSkipEncryptedMessages())
	progress := transferrer.Start()
	ctx.SetTransferProgress(progress)
	return nil
}

func setRules(transferrer *transfer.Transfer, rules *godog.Table) error {
	if rules == nil {
		return nil
	}

	transferrer.ResetRules()

	allSourceMailboxes, err := transferrer.SourceMailboxes()
	if err != nil {
		return internalError(err, "failed to get source mailboxes")
	}
	allTargetMailboxes, err := transferrer.TargetMailboxes()
	if err != nil {
		return internalError(err, "failed to get target mailboxes")
	}

	head := rules.Rows[0].Cells
	for _, row := range rules.Rows[1:] {
		source := ""
		target := ""
		fromTime := int64(0)
		toTime := int64(0)
		for n, cell := range row.Cells {
			switch head[n].Value {
			case "source":
				source = cell.Value
			case "target":
				target = cell.Value
			case "from":
				date, err := time.Parse(timeFormat, cell.Value)
				if err != nil {
					return internalError(err, "failed to parse from time")
				}
				fromTime = date.Unix()
			case "to":
				date, err := time.Parse(timeFormat, cell.Value)
				if err != nil {
					return internalError(err, "failed to parse to time")
				}
				toTime = date.Unix()
			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}
		}

		sourceMailbox, err := getMailboxByName(allSourceMailboxes, source)
		if err != nil {
			return internalError(err, "failed to match source mailboxes")
		}

		// Empty target means the same as source. Useful for exports.
		targetMailboxes := []transfer.Mailbox{}
		if target == "" {
			targetMailboxes = append(targetMailboxes, sourceMailbox)
		} else {
			targetMailbox, err := getMailboxByName(allTargetMailboxes, target)
			if err != nil {
				return internalError(err, "failed to match target mailboxes")
			}
			targetMailboxes = append(targetMailboxes, targetMailbox)
		}

		if err := transferrer.SetRule(sourceMailbox, targetMailboxes, fromTime, toTime); err != nil {
			return internalError(err, "failed to set rule")
		}
	}
	return nil
}

func getMailboxByName(mailboxes []transfer.Mailbox, name string) (transfer.Mailbox, error) {
	for _, mailbox := range mailboxes {
		if mailbox.Name == name {
			return mailbox, nil
		}
	}
	return transfer.Mailbox{}, fmt.Errorf("mailbox %s not found", name)
}
