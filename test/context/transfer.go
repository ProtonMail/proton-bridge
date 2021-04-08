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

package context

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/test/mocks"
)

// SetTransferProgress sets transfer progress.
func (ctx *TestContext) SetTransferProgress(progress *transfer.Progress) {
	ctx.transferProgress = progress
}

// GetTransferProgress returns transfer progress.
func (ctx *TestContext) GetTransferProgress() *transfer.Progress {
	return ctx.transferProgress
}

// SetTransferSkipEncryptedMessages sets whether encrypted messages will be skipped.
func (ctx *TestContext) SetTransferSkipEncryptedMessages(value bool) {
	ctx.transferSkipEncryptedMessages = value
}

// GetTransferSkipEncryptedMessages gets whether encrypted messages will be skipped.
func (ctx *TestContext) GetTransferSkipEncryptedMessages() bool {
	return ctx.transferSkipEncryptedMessages
}

// GetTransferLocalRootForImport creates temporary root for importing
// if it not exists yet, and returns its path.
func (ctx *TestContext) GetTransferLocalRootForImport() string {
	if ctx.transferLocalRootForImport != "" {
		return ctx.transferLocalRootForImport
	}
	root := ctx.createLocalRoot()
	ctx.transferLocalRootForImport = root
	return root
}

// GetTransferLocalRootForExport creates temporary root for exporting
// if it not exists yet, and returns its path.
func (ctx *TestContext) GetTransferLocalRootForExport() string {
	if ctx.transferLocalRootForExport != "" {
		return ctx.transferLocalRootForExport
	}
	root := ctx.createLocalRoot()
	ctx.transferLocalRootForExport = root
	return root
}

func (ctx *TestContext) createLocalRoot() string {
	root, err := ioutil.TempDir("", "transfer")
	if err != nil {
		panic("failed to create temp transfer root: " + err.Error())
	}

	ctx.addCleanupChecked(func() error {
		return os.RemoveAll(root)
	}, "Cleaning transfer data")

	return root
}

// GetTransferRemoteIMAPServer creates mocked IMAP server if it not created yet, and returns it.
func (ctx *TestContext) GetTransferRemoteIMAPServer() *mocks.IMAPServer {
	if ctx.transferRemoteIMAPServer != nil {
		return ctx.transferRemoteIMAPServer
	}

	port := 21300 + rand.Intn(100) //nolint[gosec] It is OK to use weaker rand generator here
	ctx.transferRemoteIMAPServer = mocks.NewIMAPServer("user", "pass", "127.0.0.1", strconv.Itoa(port))

	ctx.transferRemoteIMAPServer.Start()
	ctx.addCleanupChecked(func() error {
		ctx.transferRemoteIMAPServer.Stop()
		return nil
	}, "Cleaning transfer IMAP server")

	return ctx.transferRemoteIMAPServer
}
