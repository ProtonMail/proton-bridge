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
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

func init() {
	user.DefaultEventPeriod = 100 * time.Millisecond
	user.DefaultEventJitter = 0
}

type scenario struct {
	t *testCtx
}

func (s *scenario) reset(tb testing.TB) {
	s.t = newTestCtx(tb)
}

func (s *scenario) close(tb testing.TB) {
	require.NoError(tb, s.t.close(context.Background()))
}

func TestFeatures(testingT *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			var s scenario

			ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				s.reset(testingT)
				return ctx, nil
			})

			ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
				s.close(testingT)
				return ctx, nil
			})

			ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
				s.t.beforeStep()
				st.Text = strings.ReplaceAll(st.Text, "[GOOS]", runtime.GOOS)
				return ctx, nil
			})

			// ==== ENVIRONMENT ====
			ctx.Step(`^it succeeds$`, s.itSucceeds)
			ctx.Step(`^it fails$`, s.itFails)
			ctx.Step(`^it fails with error "([^"]*)"$`, s.itFailsWithError)
			ctx.Step(`^the internet is turned off$`, s.internetIsTurnedOff)
			ctx.Step(`^the internet is turned on$`, s.internetIsTurnedOn)
			ctx.Step(`^the user agent is "([^"]*)"$`, s.theUserAgentIs)
			ctx.Step(`^the value of the "([^"]*)" header in the request to "([^"]*)" is "([^"]*)"$`, s.theValueOfTheHeaderInTheRequestToIs)

			// ==== SETUP ====
			ctx.Step(`^there exists an account with username "([^"]*)" and password "([^"]*)"$`, s.thereExistsAnAccountWithUsernameAndPassword)
			ctx.Step(`^the account "([^"]*)" has additional address "([^"]*)"$`, s.theAccountHasAdditionalAddress)
			ctx.Step(`^the account "([^"]*)" no longer has additional address "([^"]*)"$`, s.theAccountNoLongerHasAdditionalAddress)
			ctx.Step(`^the account "([^"]*)" has (\d+) custom folders$`, s.theAccountHasCustomFolders)
			ctx.Step(`^the account "([^"]*)" has (\d+) custom labels$`, s.theAccountHasCustomLabels)
			ctx.Step(`^the account "([^"]*)" has the following custom mailboxes:$`, s.theAccountHasTheFollowingCustomMailboxes)
			ctx.Step(`^the address "([^"]*)" of account "([^"]*)" has the following messages in "([^"]*)":$`, s.theAddressOfAccountHasTheFollowingMessagesInMailbox)
			ctx.Step(`^the address "([^"]*)" of account "([^"]*)" has (\d+) messages in "([^"]*)"$`, s.theAddressOfAccountHasMessagesInMailbox)

			// ==== BRIDGE ====
			ctx.Step(`^bridge starts$`, s.bridgeStarts)
			ctx.Step(`^bridge restarts$`, s.bridgeRestarts)
			ctx.Step(`^bridge stops$`, s.bridgeStops)
			ctx.Step(`^bridge is version "([^"]*)" and the latest available version is "([^"]*)" reachable from "([^"]*)"$`, s.bridgeVersionIsAndTheLatestAvailableVersionIsReachableFrom)
			ctx.Step(`^the API requires bridge version at least "([^"]*)"$`, s.theAPIRequiresBridgeVersion)
			ctx.Step(`^the user has disabled automatic updates$`, s.theUserHasDisabledAutomaticUpdates)
			ctx.Step(`^the user changes the IMAP port to (\d+)$`, s.theUserChangesTheIMAPPortTo)
			ctx.Step(`^the user changes the SMTP port to (\d+)$`, s.theUserChangesTheSMTPPortTo)
			ctx.Step(`^the user sets the address mode of "([^"]*)" to "([^"]*)"$`, s.theUserSetsTheAddressModeOfTo)
			ctx.Step(`^the user changes the gluon path$`, s.theUserChangesTheGluonPath)
			ctx.Step(`^the user deletes the gluon files$`, s.theUserDeletesTheGluonFiles)
			ctx.Step(`^the user reports a bug$`, s.theUserReportsABug)
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

			// ==== USER ====
			ctx.Step(`^the user logs in with username "([^"]*)" and password "([^"]*)"$`, s.userLogsInWithUsernameAndPassword)
			ctx.Step(`^user "([^"]*)" logs out$`, s.userLogsOut)
			ctx.Step(`^user "([^"]*)" is deleted$`, s.userIsDeleted)
			ctx.Step(`^the auth of user "([^"]*)" is revoked$`, s.theAuthOfUserIsRevoked)
			ctx.Step(`^user "([^"]*)" is listed and connected$`, s.userIsListedAndConnected)
			ctx.Step(`^user "([^"]*)" is eventually listed and connected$`, s.userIsEventuallyListedAndConnected)
			ctx.Step(`^user "([^"]*)" is listed but not connected$`, s.userIsListedButNotConnected)
			ctx.Step(`^user "([^"]*)" is not listed$`, s.userIsNotListed)
			ctx.Step(`^user "([^"]*)" finishes syncing$`, s.userFinishesSyncing)

			// ==== IMAP ====
			ctx.Step(`^user "([^"]*)" connects IMAP client "([^"]*)"$`, s.userConnectsIMAPClient)
			ctx.Step(`^user "([^"]*)" connects IMAP client "([^"]*)" on port (\d+)$`, s.userConnectsIMAPClientOnPort)
			ctx.Step(`^user "([^"]*)" connects and authenticates IMAP client "([^"]*)"$`, s.userConnectsAndAuthenticatesIMAPClient)
			ctx.Step(`^user "([^"]*)" connects and authenticates IMAP client "([^"]*)" with address "([^"]*)"$`, s.userConnectsAndAuthenticatesIMAPClientWithAddress)
			ctx.Step(`^IMAP client "([^"]*)" can authenticate$`, s.imapClientCanAuthenticate)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate$`, s.imapClientCannotAuthenticate)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with address "([^"]*)"$`, s.imapClientCannotAuthenticateWithAddress)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with incorrect username$`, s.imapClientCannotAuthenticateWithIncorrectUsername)
			ctx.Step(`^IMAP client "([^"]*)" cannot authenticate with incorrect password$`, s.imapClientCannotAuthenticateWithIncorrectPassword)
			ctx.Step(`^IMAP client "([^"]*)" announces its ID with name "([^"]*)" and version "([^"]*)"$`, s.imapClientAnnouncesItsIDWithNameAndVersion)
			ctx.Step(`^IMAP client "([^"]*)" creates "([^"]*)"$`, s.imapClientCreatesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" deletes "([^"]*)"$`, s.imapClientDeletesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" renames "([^"]*)" to "([^"]*)"$`, s.imapClientRenamesMailboxTo)
			ctx.Step(`^IMAP client "([^"]*)" sees the following mailbox info:$`, s.imapClientSeesTheFollowingMailboxInfo)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees the following mailbox info:$`, s.imapClientEventuallySeesTheFollowingMailboxInfo)
			ctx.Step(`^IMAP client "([^"]*)" sees the following mailbox info for "([^"]*)":$`, s.imapClientSeesTheFollowingMailboxInfoForMailbox)
			ctx.Step(`^IMAP client "([^"]*)" sees the following mailboxes:$`, s.imapClientSeesTheFollowingMailboxes)
			ctx.Step(`^IMAP client "([^"]*)" sees "([^"]*)"$`, s.imapClientSeesMailbox)
			ctx.Step(`^IMAP client "([^"]*)" does not see "([^"]*)"$`, s.imapClientDoesNotSeeMailbox)
			ctx.Step(`^IMAP client "([^"]*)" counts (\d+) mailboxes under "([^"]*)"$`, s.imapClientCountsMailboxesUnder)
			ctx.Step(`^IMAP client "([^"]*)" selects "([^"]*)"$`, s.imapClientSelectsMailbox)
			ctx.Step(`^IMAP client "([^"]*)" copies the message with subject "([^"]*)" from "([^"]*)" to "([^"]*)"$`, s.imapClientCopiesTheMessageWithSubjectFromTo)
			ctx.Step(`^IMAP client "([^"]*)" copies all messages from "([^"]*)" to "([^"]*)"$`, s.imapClientCopiesAllMessagesFromTo)
			ctx.Step(`^IMAP client "([^"]*)" sees the following messages in "([^"]*)":$`, s.imapClientSeesTheFollowingMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees the following messages in "([^"]*)":$`, s.imapClientEventuallySeesTheFollowingMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" sees (\d+) messages in "([^"]*)"$`, s.imapClientSeesMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" eventually sees (\d+) messages in "([^"]*)"$`, s.imapClientEventuallySeesMessagesInMailbox)
			ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as deleted$`, s.imapClientMarksMessageAsDeleted)
			ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as not deleted$`, s.imapClientMarksMessageAsNotDeleted)
			ctx.Step(`^IMAP client "([^"]*)" marks all messages as deleted$`, s.imapClientMarksAllMessagesAsDeleted)
			ctx.Step(`^IMAP client "([^"]*)" sees that message (\d+) has the flag "([^"]*)"$`, s.imapClientSeesThatMessageHasTheFlag)
			ctx.Step(`^IMAP client "([^"]*)" expunges$`, s.imapClientExpunges)

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
			ctx.Step(`^SMTP client "([^"]*)" sends DATA "([^"]*)"$`, s.smtpClientSendsData)
			ctx.Step(`^SMTP client "([^"]*)" sends RSET$`, s.smtpClientSendsReset)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: testingT,
		},
	}

	if suite.Run() != 0 {
		testingT.Fatal("non-zero status returned, failed to run feature tests")
	}
}
