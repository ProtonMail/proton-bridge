// Copyright (c) 2020 Proton Technologies AG
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

// Package context allows integration tests to be written in a fluent, english-like way.
package context

import (
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/test/accounts"
	"github.com/ProtonMail/proton-bridge/test/mocks"
	"github.com/sirupsen/logrus"
)

type server interface {
	ListenAndServe()
	Close()
}

// TestContext manages a bridge test (mocked API, bridge instance, IMAP/SMTP servers, setup steps).
type TestContext struct {
	// Base setup for the whole bridge (core & imap & smtp).
	t               *bddT
	cfg             *fakeConfig
	listener        listener.Listener
	pmapiController PMAPIController // pmapiController is used to create pmapi clients (either real or fake) and control server state.
	testAccounts    *accounts.TestAccounts

	// Bridge core related variables.
	bridge          *bridge.Bridge
	bridgeLastError error
	credStore       bridge.CredentialsStorer
	clientManager   *pmapi.ClientManager

	// IMAP related variables.
	imapAddr          string
	imapServer        server
	imapClients       map[string]*mocks.IMAPClient
	imapLastResponses map[string]*mocks.IMAPResponse

	// SMTP related variables.
	smtpAddr          string
	smtpServer        server
	smtpClients       map[string]*mocks.SMTPClient
	smtpLastResponses map[string]*mocks.SMTPResponse

	// These are the cleanup steps executed when Cleanup() is called.
	cleanupSteps []*Cleaner

	// logger allows logging of test labels to be handled differently (silenced/diverted/whatever).
	logger logrus.FieldLogger
}

// New returns a new test TestContext.
func New() *TestContext {
	setLogrusVerbosityFromEnv()

	cfg := newFakeConfig()

	cm := pmapi.NewClientManager(cfg.GetAPIConfig())

	ctx := &TestContext{
		t:                 &bddT{},
		cfg:               cfg,
		listener:          listener.New(),
		pmapiController:   newPMAPIController(cm),
		clientManager:     cm,
		testAccounts:      newTestAccounts(),
		credStore:         newFakeCredStore(),
		imapClients:       make(map[string]*mocks.IMAPClient),
		imapLastResponses: make(map[string]*mocks.IMAPResponse),
		smtpClients:       make(map[string]*mocks.SMTPClient),
		smtpLastResponses: make(map[string]*mocks.SMTPResponse),
		logger:            logrus.StandardLogger(),
	}

	// Ensure that the config is cleaned up after the test is over.
	ctx.addCleanupChecked(cfg.ClearData, "Cleaning bridge config data")

	// Create bridge instance under test.
	ctx.withBridgeInstance()

	return ctx
}

// Cleanup runs through all cleanup steps.
// This can be a deferred call so that it is run even if the test steps failed the test.
func (ctx *TestContext) Cleanup() *TestContext {
	for _, cleanup := range ctx.cleanupSteps {
		cleanup.Execute()
	}
	return ctx
}

// GetPMAPIController returns API controller, either fake or live.
func (ctx *TestContext) GetPMAPIController() PMAPIController {
	return ctx.pmapiController
}

// GetTestingT returns testing.T compatible struct.
func (ctx *TestContext) GetTestingT() *bddT { //nolint[golint]
	return ctx.t
}

// GetTestingError returns error if test failed by using custom TestingT.
func (ctx *TestContext) GetTestingError() error {
	return ctx.t.getErrors()
}
