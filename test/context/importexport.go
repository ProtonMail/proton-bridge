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
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/users"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// GetImportExport returns import-export instance.
func (ctx *TestContext) GetImportExport() *importexport.ImportExport {
	return ctx.importExport
}

// withImportExportInstance creates a import-export instance for use in the test.
// TestContext has this by default once called with env variable TEST_APP=ie.
func (ctx *TestContext) withImportExportInstance() {
	ctx.importExport = newImportExportInstance(ctx.t, ctx.locations, ctx.cache, ctx.credStore, ctx.listener, ctx.clientManager)
	ctx.users = ctx.importExport.Users
}

// newImportExportInstance creates a new import-export instance configured to use the given config/credstore.
func newImportExportInstance(
	t *bddT,
	locations importexport.Locator,
	cache importexport.Cacher,
	credStore users.CredentialsStorer,
	eventListener listener.Listener,
	clientManager pmapi.Manager,
) *importexport.ImportExport {
	panicHandler := &panicHandler{t: t}
	return importexport.New(locations, cache, panicHandler, eventListener, clientManager, credStore)
}
