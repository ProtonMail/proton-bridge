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

// Package context allows integration tests to be written in a fluent, english-like way.
package context

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
	"github.com/ProtonMail/proton-bridge/v2/test/mocks"
	"github.com/sirupsen/logrus"
)

type server interface {
	ListenAndServe()
	Close()
}

// TestContext manages a bridge test (mocked API, bridge instance, IMAP/SMTP servers, setup steps).
type TestContext struct {
	// Base setup for the whole bridge (core & imap & smtp).
	t            *bddT
	cache        *fakeCache
	locations    *fakeLocations
	settings     *fakeSettings
	listener     listener.Listener
	userAgent    *useragent.UserAgent
	testAccounts *accounts.TestAccounts

	// pmapiController is used to control real or fake pmapi clients.
	// The clients are created by the clientManager.
	pmapiController PMAPIController
	clientManager   pmapi.Manager

	// Core related variables.
	bridge    *bridge.Bridge
	users     *users.Users
	credStore users.CredentialsStorer
	lastError error

	// IMAP related variables.
	imapAddr           string
	imapServer         server
	imapClients        map[string]*mocks.IMAPClient
	imapLastResponses  map[string]*mocks.IMAPResponse
	imapResponseLocker sync.Locker

	// SMTP related variables.
	smtpAddr           string
	smtpServer         server
	smtpClients        map[string]*mocks.SMTPClient
	smtpLastResponses  map[string]*mocks.SMTPResponse
	smtpResponseLocker sync.Locker

	// Store releated variables.
	bddMessageIDsToAPIIDs map[string]string

	// These are the cleanup steps executed when Cleanup() is called.
	cleanupSteps []*Cleaner

	// logger allows logging of test labels to be handled differently (silenced/diverted/whatever).
	logger logrus.FieldLogger
}

// New returns a new test TestContext.
func New() *TestContext {
	setLogrusVerbosityFromEnv()

	listener := listener.New()
	pmapiController, clientManager := newPMAPIController(listener)

	ctx := &TestContext{
		t:                     &bddT{},
		cache:                 newFakeCache(),
		locations:             newFakeLocations(),
		settings:              newFakeSettings(),
		listener:              listener,
		userAgent:             useragent.New(),
		pmapiController:       pmapiController,
		clientManager:         clientManager,
		testAccounts:          newTestAccounts(),
		credStore:             newFakeCredStore(),
		imapClients:           make(map[string]*mocks.IMAPClient),
		imapLastResponses:     make(map[string]*mocks.IMAPResponse),
		imapResponseLocker:    &sync.Mutex{},
		smtpClients:           make(map[string]*mocks.SMTPClient),
		smtpLastResponses:     make(map[string]*mocks.SMTPResponse),
		smtpResponseLocker:    &sync.Mutex{},
		bddMessageIDsToAPIIDs: make(map[string]string),
		logger:                logrus.StandardLogger().WithField("ctx", "scenario"),
	}

	ctx.logger.Info("New context")
	ctx.addCleanup(func() { ctx.logger.Info("Context end") }, "End of context")

	// Ensure that the config is cleaned up after the test is over.
	ctx.addCleanupChecked(ctx.locations.Clear, "Cleaning bridge config data")

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

// GetClientManager returns client manager being used for testing.
func (ctx *TestContext) GetClientManager() pmapi.Manager {
	return ctx.clientManager
}

// GetUserAgent returns the current user agent.
func (ctx *TestContext) GetUserAgent() string {
	return ctx.userAgent.String()
}

// GetTestingT returns testing.T compatible struct.
func (ctx *TestContext) GetTestingT() *bddT { //nolint:revive
	return ctx.t
}

// GetTestingError returns error if test failed by using custom TestingT.
func (ctx *TestContext) GetTestingError() error {
	return ctx.t.getErrors()
}

// SetLastError sets the last error that occurred while executing an action.
func (ctx *TestContext) SetLastError(err error) {
	ctx.lastError = err
}

// GetLastError returns the last error that occurred while executing an action.
func (ctx *TestContext) GetLastError() error {
	return ctx.lastError
}

func (ctx *TestContext) MessagePreparationStarted(username string) {
	ctx.pmapiController.LockEvents(username)
}

func (ctx *TestContext) MessagePreparationFinished(username string) {
	ctx.pmapiController.UnlockEvents(username)
}

func (ctx *TestContext) CredentialsFailsOnWrite(shouldFail bool) {
	ctx.credStore.(*fakeCredStore).failOnWrite = shouldFail                                              //nolint:forcetypeassert
	ctx.addCleanup(func() { ctx.credStore.(*fakeCredStore).failOnWrite = false }, "credentials-cleanup") //nolint:forcetypeassert
}
