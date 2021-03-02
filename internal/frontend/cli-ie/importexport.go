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

package cliie

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) importLocalMessages(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user, path := f.getUserAndPath(c, false)
	if user == nil || path == "" {
		return
	}

	t, err := f.ie.GetLocalImporter(user.Username(), user.GetPrimaryAddress(), path)
	f.transfer(t, err, false, true)
}

func (f *frontendCLI) importRemoteMessages(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user := f.askUserByIndexOrName(c)
	if user == nil {
		return
	}

	username := f.readStringInAttempts("IMAP username", c.ReadLine, isNotEmpty)
	if username == "" {
		return
	}
	password := f.readStringInAttempts("IMAP password", c.ReadPassword, isNotEmpty)
	if password == "" {
		return
	}
	host := f.readStringInAttempts("IMAP host", c.ReadLine, isNotEmpty)
	if host == "" {
		return
	}
	port := f.readStringInAttempts("IMAP port", c.ReadLine, isNotEmpty)
	if port == "" {
		return
	}

	t, err := f.ie.GetRemoteImporter(user.Username(), user.GetPrimaryAddress(), username, password, host, port)
	f.transfer(t, err, false, true)
}

func (f *frontendCLI) exportMessagesToEML(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user, path := f.getUserAndPath(c, true)
	if user == nil || path == "" {
		return
	}

	t, err := f.ie.GetEMLExporter(user.Username(), user.GetPrimaryAddress(), path)
	f.transfer(t, err, true, false)
}

func (f *frontendCLI) exportMessagesToMBOX(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user, path := f.getUserAndPath(c, true)
	if user == nil || path == "" {
		return
	}

	t, err := f.ie.GetMBOXExporter(user.Username(), user.GetPrimaryAddress(), path)
	f.transfer(t, err, true, false)
}

func (f *frontendCLI) getUserAndPath(c *ishell.Context, createPath bool) (types.User, string) {
	user := f.askUserByIndexOrName(c)
	if user == nil {
		return nil, ""
	}

	path := f.readStringInAttempts("Path of EML and MBOX files", c.ReadLine, isNotEmpty)
	if path == "" {
		return nil, ""
	}

	if createPath {
		_ = os.Mkdir(path, os.ModePerm)
	}

	return user, path
}

func (f *frontendCLI) transfer(t *transfer.Transfer, err error, askSkipEncrypted bool, askGlobalMailbox bool) {
	if err != nil {
		f.printAndLogError("Failed to init transferrer: ", err)
		return
	}

	if askSkipEncrypted {
		skipEncryptedMessages := f.yesNoQuestion("Skip encrypted messages")
		t.SetSkipEncryptedMessages(skipEncryptedMessages)
	}

	if !f.setTransferRules(t) {
		return
	}

	if askGlobalMailbox {
		if err := f.setTransferGlobalMailbox(t); err != nil {
			f.printAndLogError("Failed to create global mailbox: ", err)
			return
		}
	}

	progress := t.Start()
	for range progress.GetUpdateChannel() {
		f.printTransferProgress(progress)
	}
	f.printTransferResult(progress)
}

func (f *frontendCLI) setTransferGlobalMailbox(t *transfer.Transfer) error {
	labelName := fmt.Sprintf("Imported %s", time.Now().Format("Jan-02-2006 15:04"))

	useGlobalLabel := f.yesNoQuestion("Use global label " + labelName)
	if !useGlobalLabel {
		return nil
	}

	globalMailbox, err := t.CreateTargetMailbox(transfer.Mailbox{
		Name:        labelName,
		Color:       pmapi.LabelColors[0],
		IsExclusive: false,
	})
	if err != nil {
		return err
	}

	t.SetGlobalMailbox(&globalMailbox)
	return nil
}

func (f *frontendCLI) setTransferRules(t *transfer.Transfer) bool {
	f.Println("Rules:")
	for _, rule := range t.GetRules() {
		if !rule.Active {
			continue
		}
		targets := strings.Join(rule.TargetMailboxNames(), ", ")
		if rule.HasTimeLimit() {
			f.Printf(" %-30s → %s (%s - %s)\n", rule.SourceMailbox.Name, targets, rule.FromDate(), rule.ToDate())
		} else {
			f.Printf(" %-30s → %s\n", rule.SourceMailbox.Name, targets)
		}
	}

	return f.yesNoQuestion("Proceed")
}

func (f *frontendCLI) printTransferProgress(progress *transfer.Progress) {
	counts := progress.GetCounts()
	if counts.Total != 0 {
		f.Println(fmt.Sprintf(
			"Progress update: %d (%d / %d) / %d, skipped: %d, failed: %d",
			counts.Imported,
			counts.Exported,
			counts.Added,
			counts.Total,
			counts.Skipped,
			counts.Failed,
		))
	}

	if progress.IsPaused() {
		f.Printf("Transfer is paused bacause %s", progress.PauseReason())
		if !f.yesNoQuestion("Continue (y) or stop (n)") {
			progress.Stop()
		}
	}
}

func (f *frontendCLI) printTransferResult(progress *transfer.Progress) {
	err := progress.GetFatalError()
	if err != nil {
		f.Println("Transfer failed: " + err.Error())
		return
	}

	statuses := progress.GetFailedMessages()
	if len(statuses) == 0 {
		f.Println("Transfer finished!")
		return
	}

	f.Println("Transfer finished with errors:")
	for _, messageStatus := range statuses {
		f.Printf(
			" %-17s | %-30s | %-30s\n  %s: %s\n",
			messageStatus.Time.Format("Jan 02 2006 15:04"),
			messageStatus.From,
			messageStatus.Subject,
			messageStatus.SourceID,
			messageStatus.GetErrorMessage(),
		)
	}
}
