// Copyright (c) 2023 Proton AG
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
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type scenario struct {
	t *testCtx
}

// reset resets the test context for a new scenario.
func (s *scenario) reset(tb testing.TB) {
	s.t = newTestCtx(tb)
}

// replace replaces the placeholders in the scenario with the values from the test context.
func (s *scenario) replace(sc *godog.Scenario) {
	for _, step := range sc.Steps {
		step.Text = s.t.replace(step.Text)

		if arg := step.Argument; arg != nil {
			if table := arg.DataTable; table != nil {
				for _, row := range table.Rows {
					for _, cell := range row.Cells {
						cell.Value = s.t.replace(cell.Value)
					}
				}
			}

			if doc := arg.DocString; doc != nil {
				doc.Content = s.t.replace(doc.Content)
			}
		}
	}
}

// close closes the test context.
func (s *scenario) close(_ testing.TB) {
	s.t.close(context.Background())
}

func TestFeatures(testingT *testing.T) {
	var s scenario

	suite := godog.TestSuite{
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.BeforeSuite(func() {
				// Global setup.
			})

			ctx.AfterSuite(func() {
				// Global teardown.
			})
		},

		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				s.reset(testingT)
				s.replace(sc)
				return ctx, nil
			})

			ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.close(testingT)
				return ctx, nil
			})

			ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
				s.t.beforeStep(st)
				return ctx, nil
			})

			ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, _ error) (context.Context, error) {
				s.t.afterStep(st, status)
				return ctx, nil
			})

			// ==== ENVIRONMENT ====
			ctx.Step(`^it succeeds$`, s.itSucceeds)
			ctx.Step(`^it fails$`, s.itFails)
			ctx.Step(`^it fails with error "([^"]*)"$`, s.itFailsWithError)
			ctx.Step(`^the internet is turned off$`, s.internetIsTurnedOff)
			ctx.Step(`^the internet is turned on$`, s.internetIsTurnedOn)
			ctx.Step(`^the user agent is "([^"]*)"$`, s.theUserAgentIs)
			ctx.Step(`^the header in the "([^"]*)" request to "([^"]*)" has "([^"]*)" set to "([^"]*)"$`, s.theHeaderInTheRequestToHasSetTo)
			ctx.Step(`^the body in the "([^"]*)" request to "([^"]*)" is:$`, s.theBodyInTheRequestToIs)
			ctx.Step(`^the body in the "([^"]*)" response to "([^"]*)" is:$`, s.theBodyInTheResponseToIs)
			ctx.Step(`^the API requires bridge version at least "([^"]*)"$`, s.theAPIRequiresBridgeVersion)
			ctx.Step(`^the network port (\d+) is busy$`, s.networkPortIsBusy)
			ctx.Step(`^the network port range (\d+)-(\d+) is busy$`, s.networkPortRangeIsBusy)
			ctx.Step(`^bridge IMAP port is (\d+)`, s.bridgeIMAPPortIs)
			ctx.Step(`^bridge SMTP port is (\d+)`, s.bridgeSMTPPortIs)
			// ==== SETUP ====
			ctx.Step(`^there exists an account with username "([^"]*)" and password "([^"]*)"$`, s.thereExistsAnAccountWithUsernameAndPassword)
			ctx.Step(`^there exists a disabled account with username "([^"]*)" and password "([^"]*)"$`, s.thereExistsAnAccountWithUsernameAndPasswordWithDisablePrimary)
			ctx.Step(`^the account "([^"]*)" has additional address "([^"]*)"$`, s.theAccountHasAdditionalAddress)
			ctx.Step(`^the account "([^"]*)" has additional disabled address "([^"]*)"$`, s.theAccountHasAdditionalDisabledAddress)
			ctx.Step(`^the account "([^"]*)" has additional address "([^"]*)" without keys$`, s.theAccountHasAdditionalAddressWithoutKeys)
			ctx.Step(`^the account "([^"]*)" no longer has additional address "([^"]*)"$`, s.theAccountNoLongerHasAdditionalAddress)
			ctx.Step(`^the account "([^"]*)" has (\d+) custom folders$`, s.theAccountHasCustomFolders)
			ctx.Step(`^the account "([^"]*)" has (\d+) custom labels$`, s.theAccountHasCustomLabels)
			ctx.Step(`^the account "([^"]*)" has the following custom mailboxes:$`, s.theAccountHasTheFollowingCustomMailboxes)
			ctx.Step(`^the address "([^"]*)" of account "([^"]*)" has the following messages in "([^"]*)":$`, s.theAddressOfAccountHasTheFollowingMessagesInMailbox)
			ctx.Step(`^the address "([^"]*)" of account "([^"]*)" has (\d+) messages in "([^"]*)"$`, s.theAddressOfAccountHasMessagesInMailbox)
			ctx.Step(`^the following fields were changed in draft (\d+) for address "([^"]*)" of account "([^"]*)":$`, s.theFollowingFieldsWereChangedInDraftForAddressOfAccount)
			ctx.Step(`^draft (\d+) for address "([^"]*)" of account "([^"]*)" was moved to trash$`, s.drafAtIndexWasMovedToTrashForAddressOfAccount)

			// === REPORTER ===
			ctx.Step(`^test skips reporter checks$`, s.skipReporterChecks)

			// ==== BRIDGE ====
			ctx.Step(`^bridge starts$`, s.bridgeStarts)
			ctx.Step(`^bridge restarts$`, s.bridgeRestarts)
			ctx.Step(`^bridge stops$`, s.bridgeStops)
			ctx.Step(`^bridge is version "([^"]*)" and the latest available version is "([^"]*)" reachable from "([^"]*)"$`, s.bridgeVersionIsAndTheLatestAvailableVersionIsReachableFrom)
			ctx.Step(`^the user has disabled automatic updates$`, s.theUserHasDisabledAutomaticUpdates)
			ctx.Step(`^the user has disabled automatic start`, s.theUserHasDisabledAutomaticStart)
			ctx.Step(`^the user has enabled alternative routing`, s.theUserHasEnabledAlternativeRouting)
			ctx.Step(`^the user set IMAP mode to SSL`, s.theUserSetIMAPModeToSSL)
			ctx.Step(`^the user set SMTP mode to SSL`, s.theUserSetSMTPModeToSSL)
			ctx.Step(`^the user changes the IMAP port to (\d+)$`, s.theUserChangesTheIMAPPortTo)
			ctx.Step(`^the user changes the SMTP port to (\d+)$`, s.theUserChangesTheSMTPPortTo)
			ctx.Step(`^the user sets the address mode of user "([^"]*)" to "([^"]*)"$`, s.theUserSetsTheAddressModeOfUserTo)
			ctx.Step(`^the user changes the default keychain application`, s.theUserChangesTheDefaultKeychainApplication)
			ctx.Step(`^the user changes the gluon path$`, s.theUserChangesTheGluonPath)
			ctx.Step(`^the user deletes the gluon files$`, s.theUserDeletesTheGluonFiles)
			ctx.Step(`^the user deletes the gluon cache$`, s.theUserDeletesTheGluonCache)
			ctx.Step(`^the user reports a bug$`, s.theUserReportsABug)
			ctx.Step(`^the user hides All Mail$`, s.theUserHidesAllMail)
			ctx.Step(`^the user shows All Mail$`, s.theUserShowsAllMail)
			ctx.Step(`^the user disables telemetry in bridge settings$`, s.theUserDisablesTelemetryInBridgeSettings)
			ctx.Step(`^the user enables telemetry in bridge settings$`, s.theUserEnablesTelemetryInBridgeSettings)
			ctx.Step(`^bridge sends a connection up event$`, s.bridgeSendsAConnectionUpEvent)
			ctx.Step(`^bridge sends a connection down event$`, s.bridgeSendsAConnectionDownEvent)
			ctx.Step(`^bridge sends a deauth event for user "([^"]*)"$`, s.bridgeSendsADeauthEventForUser)
			ctx.Step(`^bridge sends an address created event for user "([^"]*)"$`, s.bridgeSendsAnAddressCreatedEventForUser)
			ctx.Step(`^bridge sends an address deleted event for user "([^"]*)"$`, s.bridgeSendsAnAddressDeletedEventForUser)
			ctx.Step(`^bridge sends sync started and finished events for user "([^"]*)"$`, s.bridgeSendsSyncStartedAndFinishedEventsForUser)
			ctx.Step(`^bridge sends an update available event for version "([^"]*)"$`, s.bridgeSendsAnUpdateAvailableEventForVersion)
			ctx.Step(`^bridge sends a manual update event for version "([^"]*)"$`, s.bridgeSendsAManualUpdateEventForVersion)
			ctx.Step(`^bridge sends an update installed event for version "([^"]*)"$`, s.bridgeSendsAnUpdateInstalledEventForVersion)
			ctx.Step(`^bridge sends an update not available event$`, s.bridgeSendsAnUpdateNotAvailableEvent)
			ctx.Step(`^bridge sends a forced update event$`, s.bridgeSendsAForcedUpdateEvent)
			ctx.Step(`^bridge reports a message with "([^"]*)"$`, s.bridgeReportsMessage)
			ctx.Step(`^bridge telemetry feature is enabled$`, s.bridgeTelemetryFeatureEnabled)
			ctx.Step(`^bridge telemetry feature is disabled$`, s.bridgeTelemetryFeatureDisabled)

			// ==== FRONTEND ====
			ctx.Step(`^frontend sees that bridge is version "([^"]*)"$`, s.frontendSeesThatBridgeIsVersion)

			// ==== USER ====
			ctx.Step(`^the user logs in with username "([^"]*)" and password "([^"]*)"$`, s.userLogsInWithUsernameAndPassword)
			ctx.Step(`^user "([^"]*)" logs out$`, s.userLogsOut)
			ctx.Step(`^user "([^"]*)" is deleted$`, s.userIsDeleted)
			ctx.Step(`^the auth of user "([^"]*)" is revoked$`, s.theAuthOfUserIsRevoked)
			ctx.Step(`^user "([^"]*)" is eventually listed and connected$`, s.userIsEventuallyListedAndConnected)
			ctx.Step(`^user "([^"]*)" is listed but not connected$`, s.userIsListedButNotConnected)
			ctx.Step(`^user "([^"]*)" is not listed$`, s.userIsNotListed)
			ctx.Step(`^user "([^"]*)" finishes syncing$`, s.userFinishesSyncing)
			ctx.Step(`^user "([^"]*)" has telemetry set to (\d+)$`, s.userHasTelemetrySetTo)
			ctx.Step(`^the bridge password of user "([^"]*)" is changed to "([^"]*)"`, s.bridgePasswordOfUserIsChangedTo)
			ctx.Step(`^the bridge password of user "([^"]*)" is equal to "([^"]*)"`, s.bridgePasswordOfUserIsEqualTo)

			// ==== IMAP ====
			ctx.Step(`^user "([^"]*)" connects IMAP client "([^"]*)"$`, s.userConnectsIMAPClient)
			ctx.Step(`^user "([^"]*)" connects IMAP client "([^"]*)" on port (\d+)$`, s.userConnectsIMAPClientOnPort)
			ctx.Step(`^user "([^"]*)" connects and authenticates IMAP client "([^"]*)"$`, s.userConnectsAndAuthenticatesIMAPClient)
			ctx.Step(`^user "([^"]*)" connects and authenticates IMAP client "([^"]*)" with address "([^"]*)"$`, s.userConnectsAndAuthenticatesIMAPClientWithAddress)
			ctx.Step(`^user "([^"]*)" connects and can not authenticate IMAP client "([^"]*)" with address "([^"]*)"$`, s.userConnectsAndCanNotAuthenticateIMAPClientWithAddress)
			ctx.Step(`^IMAP client "([^"]*)" can authenticate$`, s.imapClientCanAuthenticate)
			ctx.Step(`^IMAP client "([^"]*)" can authenticate with address "([^"]*)"$`, s.imapClientCanAuthenticateWithAddress)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate$`, s.imapClientCannotAuthenticate)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with address "([^"]*)"$`, s.imapClientCannotAuthenticateWithAddress)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with incorrect username$`, s.imapClientCannotAuthenticateWithIncorrectUsername)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with incorrect password$`, s.imapClientCannotAuthenticateWithIncorrectPassword)
			ctx.Step(`^IMAP client "([^"]*)" announces its ID with name "([^"]*)" and version "([^"]*)"$`, s.imapClientAnnouncesItsIDWithNameAndVersion)
			ctx.Step(`^IMAP client "([^"]*)" creates "([^"]*)"$`, s.imapClientCreatesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" deletes "([^"]*)"$`, s.imapClientDeletesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" renames "([^"]*)" to "([^"]*)"$`, s.imapClientRenamesMailboxTo)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees the following mailbox info:$`, s.imapClientEventuallySeesTheFollowingMailboxInfo)
			ctx.Step(`^IMAP client "([^"]*)" sees the following mailbox info for "([^"]*)":$`, s.imapClientSeesTheFollowingMailboxInfoForMailbox)
			ctx.Step(`^IMAP client "([^"]*)" sees "([^"]*)"$`, s.imapClientSeesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" does not see "([^"]*)"$`, s.imapClientDoesNotSeeMailbox)
			ctx.Step(`^IMAP client "([^"]*)" counts (\d+) mailboxes under "([^"]*)"$`, s.imapClientCountsMailboxesUnder)
			ctx.Step(`^IMAP client "([^"]*)" selects "([^"]*)"$`, s.imapClientSelectsMailbox)
			ctx.Step(`^IMAP client "([^"]*)" copies the message with subject "([^"]*)" from "([^"]*)" to "([^"]*)"$`, s.imapClientCopiesTheMessageWithSubjectFromTo)
			ctx.Step(`^IMAP client "([^"]*)" copies all messages from "([^"]*)" to "([^"]*)"$`, s.imapClientCopiesAllMessagesFromTo)
			ctx.Step(`^IMAP client "([^"]*)" moves the message with subject "([^"]*)" from "([^"]*)" to "([^"]*)"$`, s.imapClientMovesTheMessageWithSubjectFromTo)
			ctx.Step(`^IMAP client "([^"]*)" moves all messages from "([^"]*)" to "([^"]*)"$`, s.imapClientMovesAllMessagesFromTo)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees the following messages in "([^"]*)":$`, s.imapClientEventuallySeesTheFollowingMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees (\d+) messages in "([^"]*)"$`, s.imapClientEventuallySeesMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as deleted$`, s.imapClientMarksMessageAsDeleted)
			ctx.Step(`^IMAP client "([^"]*)" marks the message with subject "([^"]*)" as deleted$`, s.imapClientMarksTheMessageWithSubjectAsDeleted)
			ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as not deleted$`, s.imapClientMarksMessageAsNotDeleted)
			ctx.Step(`^IMAP client "([^"]*)" marks all messages as deleted$`, s.imapClientMarksAllMessagesAsDeleted)
			ctx.Step(`^IMAP client "([^"]*)" sees that message (\d+) has the flag "([^"]*)"$`, s.imapClientSeesThatMessageHasTheFlag)
			ctx.Step(`^IMAP client "([^"]*)" expunges$`, s.imapClientExpunges)
			ctx.Step(`^IMAP client "([^"]*)" appends the following message to "([^"]*)":$`, s.imapClientAppendsTheFollowingMessageToMailbox)
			ctx.Step(`^IMAP client "([^"]*)" appends the following messages to "([^"]*)":$`, s.imapClientAppendsTheFollowingMessagesToMailbox)
			ctx.Step(`^IMAP client "([^"]*)" appends "([^"]*)" to "([^"]*)"$`, s.imapClientAppendsToMailbox)
			ctx.Step(`^IMAP clients "([^"]*)" and "([^"]*)" move message with subject "([^"]*)" of "([^"]*)" to "([^"]*)" by ([^"]*) ([^"]*) ([^"]*)`, s.imapClientsMoveMessageWithSubjectUserFromToByOrderedOperations)
			ctx.Step(`^IMAP client "([^"]*)" sees header "([^"]*)" in message with subject "([^"]*)" in "([^"]*)"$`, s.imapClientSeesHeaderInMessageWithSubject)
			ctx.Step(`^IMAP client "([^"]*)" does not see header "([^"]*)" in message with subject "([^"]*)" in "([^"]*)"$`, s.imapClientDoesNotSeeHeaderInMessageWithSubject)

			// ==== SMTP ====
			ctx.Step(`^user "([^"]*)" connects SMTP client "([^"]*)"$`, s.userConnectsSMTPClient)
			ctx.Step(`^user "([^"]*)" connects SMTP client "([^"]*)" on port (\d+)$`, s.userConnectsSMTPClientOnPort)
			ctx.Step(`^user "([^"]*)" connects and authenticates SMTP client "([^"]*)"$`, s.userConnectsAndAuthenticatesSMTPClient)
			ctx.Step(`^user "([^"]*)" connects and authenticates SMTP client "([^"]*)" with address "([^"]*)"$`, s.userConnectsAndAuthenticatesSMTPClientWithAddress)
			ctx.Step(`^SMTP client "([^"]*)" can authenticate$`, s.smtpClientCanAuthenticate)
			ctx.Step(`^SMTP client "([^"]*)" cannot authenticate$`, s.smtpClientCannotAuthenticate)
			ctx.Step(`^SMTP client "([^"]*)" cannot authenticate with incorrect username$`, s.smtpClientCannotAuthenticateWithIncorrectUsername)
			ctx.Step(`^SMTP client "([^"]*)" cannot authenticate with incorrect password$`, s.smtpClientCannotAuthenticateWithIncorrectPassword)
			ctx.Step(`^SMTP client "([^"]*)" sends MAIL FROM "([^"]*)"$`, s.smtpClientSendsMailFrom)
			ctx.Step(`^SMTP client "([^"]*)" sends RCPT TO "([^"]*)"$`, s.smtpClientSendsRcptTo)
			ctx.Step(`^SMTP client "([^"]*)" sends DATA:$`, s.smtpClientSendsData)
			ctx.Step(`^SMTP client "([^"]*)" sends RSET$`, s.smtpClientSendsReset)
			ctx.Step(`^SMTP client "([^"]*)" sends the following message from "([^"]*)" to "([^"]*)":$`, s.smtpClientSendsTheFollowingMessageFromTo)
			ctx.Step(`^SMTP client "([^"]*)" logs out$`, s.smtpClientLogsOut)

			// ==== TELEMETRY ====
			ctx.Step(`^bridge eventually sends the following heartbeat:$`, s.bridgeEventuallySendsTheFollowingHeartbeat)
			ctx.Step(`^bridge needs to send heartbeat`, s.bridgeNeedsToSendHeartbeat)
			ctx.Step(`^bridge do not need to send heartbeat`, s.bridgeDoNotNeedToSendHeartbeat)
			ctx.Step(`^heartbeat is not whitelisted`, s.heartbeatIsNotwhitelisted)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    getFeaturePaths(),
			TestingT: testingT,
		},
	}

	if suite.Run() != 0 {
		testingT.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func getFeaturePaths() []string {
	var paths []string

	if features := os.Getenv("FEATURES"); features != "" {
		paths = strings.Split(features, " ")
	} else {
		paths = []string{"features"}
	}

	return paths
}
