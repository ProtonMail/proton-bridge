// Copyright (c) 2024 Proton AG
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

import "github.com/cucumber/godog"

func (s *scenario) steps(ctx *godog.ScenarioContext) {
	// ==== ENVIRONMENT ====
	ctx.Step(`^it succeeds$`, s.itSucceeds)
	ctx.Step(`^it fails$`, s.itFails)
	ctx.Step(`^it fails with error "([^"]*)"$`, s.itFailsWithError)
	ctx.Step(`^the internet is turned off$`, s.internetIsTurnedOff)
	ctx.Step(`^the internet is turned on$`, s.internetIsTurnedOn)
	ctx.Step(`^the user agent is "([^"]*)"$`, s.theUserAgentIs)
	ctx.Step(`^the header in the "([^"]*)" request to "([^"]*)" has "([^"]*)" set to "([^"]*)"$`, s.theHeaderInTheRequestToHasSetTo)
	ctx.Step(`^the header in the "([^"]*)" multipart request to "([^"]*)" has "([^"]*)" set to "([^"]*)"$`, s.theHeaderInTheMultipartRequestToHasSetTo)
	ctx.Step(`^the header in the "([^"]*)" multipart request to "([^"]*)" has file "([^"]*)"$`, s.theHeaderInTheMultipartRequestToHasFile)
	ctx.Step(`^the header in the "([^"]*)" multipart request to "([^"]*)" has no file "([^"]*)"$`, s.theHeaderInTheMultipartRequestToHasNoFile)
	ctx.Step(`^the body in the "([^"]*)" request to "([^"]*)" is:$`, s.theBodyInTheRequestToIs)
	ctx.Step(`^the body in the "([^"]*)" response to "([^"]*)" is:$`, s.theBodyInTheResponseToIs)
	ctx.Step(`^the API requires bridge version at least "([^"]*)"$`, s.theAPIRequiresBridgeVersion)
	ctx.Step(`^the network port (\d+) is busy$`, s.networkPortIsBusy)
	ctx.Step(`^the network port range (\d+)-(\d+) is busy$`, s.networkPortRangeIsBusy)
	ctx.Step(`^bridge IMAP port is (\d+)`, s.bridgeIMAPPortIs)
	ctx.Step(`^bridge SMTP port is (\d+)`, s.bridgeSMTPPortIs)
	ctx.Step(`^the message used "([^"]*)" key for sending$`, s.theMessageUsedKeyForSending)
	ctx.Step(`^the key for address "([^"]*)" was used to import`, s.theKeyForAddressWasUsedToImport)
	ctx.Step(`^the key for address "([^"]*)" was used to create draft`, s.theKeyForAddressWasUsedToCreateDraft)

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
	ctx.Step(`^the user reports a bug with field "([^"]*)" set to "([^"]*)"$`, s.theUserReportsABugWithSingleHeaderChange)
	ctx.Step(`^the user reports a bug with details:$`, s.theUserReportsABugWithDetails)
	ctx.Step(`^the description "([^"]*)" provides the following KB suggestions:$`, s.theDescriptionInputProvidesKnowledgeBaseArticles)
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
	ctx.Step(`^the user logs in with alias address "([^"]*)" and password "([^"]*)"$`, s.userLogsInWithAliasAddressAndPassword)
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

	// ==== ACCOUNT SETTINGS ====
	ctx.Step(`^the account "([^"]*)" has public key attachment "([^"]*)"`, s.accountHasPublicKeyAttachment)
	ctx.Step(`^the account "([^"]*)" has sign external messages "([^"]*)"`, s.accountHasSignExternalMessages)
	ctx.Step(`^the account "([^"]*)" has default draft format "([^"]*)"`, s.accountHasDefaultDraftFormat)
	ctx.Step(`^the account "([^"]*)" has default PGP schema "([^"]*)"`, s.accountHasDefaultPGPSchema)
	ctx.Step(`^the account "([^"]*)" matches the following settings:$`, s.accountMatchesSettings)

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
	ctx.Step(`^IMAP client "([^"]*)" closes$`, s.imapClientCloses)
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
	ctx.Step(`^IMAP client "([^"]*)" eventually sees the following message in "([^"]*)" with this structure:$`, s.imapClientSeesMessageInMailboxWithStructure)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees (\d+) messages in "([^"]*)"$`, s.imapClientEventuallySeesMessagesInMailbox)
	ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as deleted$`, s.imapClientMarksMessageAsDeleted)
	ctx.Step(`^IMAP client "([^"]*)" marks the message with subject "([^"]*)" as deleted$`, s.imapClientMarksTheMessageWithSubjectAsDeleted)
	ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as not deleted$`, s.imapClientMarksMessageAsNotDeleted)
	ctx.Step(`^IMAP client "([^"]*)" marks all messages as deleted$`, s.imapClientMarksAllMessagesAsDeleted)
	ctx.Step(`^IMAP client "([^"]*)" expunges$`, s.imapClientExpunges)
	ctx.Step(`^IMAP client "([^"]*)" marks message (\d+) as "([^"]*)"$`, s.imapClientMarksMessageAsState)
	ctx.Step(`^IMAP client "([^"]*)" marks the message with subject "([^"]*)" as "([^"]*)"$`, s.imapClientMarksTheMessageWithSubjectAsState)
	ctx.Step(`^IMAP client "([^"]*)" marks all messages as "([^"]*)"$`, s.imapClientMarksAllMessagesAsState)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that message at row (\d+) has the flag "([^"]*)"$`, s.imapClientEventuallySeesThatMessageHasTheFlag)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that message at row (\d+) does not have the flag "([^"]*)"$`, s.imapClientSeesThatMessageDoesNotHaveTheFlag)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that the message with subject "([^"]*)" has the flag "([^"]*)"`, s.imapClientEventuallySeesThatTheMessageWithSubjectHasTheFlag)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that the message with subject "([^"]*)" does not have the flag "([^"]*)"`, s.imapClientEventuallySeesThatTheMessageWithSubjectDoesNotHaveTheFlag)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that all the messages have the flag "([^"]*)"`, s.imapClientEventuallySeesThatAllTheMessagesHaveTheFlag)
	ctx.Step(`^IMAP client "([^"]*)" eventually sees that all the messages do not have the flag "([^"]*)"`, s.imapClientEventuallySeesThatAllTheMessagesDoNotHaveTheFlag)
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
	ctx.Step(`^SMTP client "([^"]*)" sends the following EML "([^"]*)" from "([^"]*)" to "([^"]*)"$`, s.smtpClientSendsTheFollowingEmlFromTo)
	ctx.Step(`^SMTP client "([^"]*)" logs out$`, s.smtpClientLogsOut)

	// ==== EXTERNAL ====
	ctx.Step(`^external client deletes all messages`, s.externalClientDeletesAllMessages)
	ctx.Step(`^external client sends the following message from "([^"]*)" to "([^"]*)":$`, s.externalClientSendsTheFollowingMessageFromTo)
	ctx.Step(`^external client fetches the following message with subject "([^"]*)" and sender "([^"]*)" and state "([^"]*)"$`, s.externalClientFetchesTheFollowingMessage)
	ctx.Step(`^external client fetches the following message with subject "([^"]*)" and sender "([^"]*)" and state "([^"]*)" with this structure:$`, s.externalClientSeesMessageWithStructure)

	// ==== TELEMETRY ====
	ctx.Step(`^bridge eventually sends the following heartbeat:$`, s.bridgeEventuallySendsTheFollowingHeartbeat)
	ctx.Step(`^bridge needs to send heartbeat`, s.bridgeNeedsToSendHeartbeat)
	ctx.Step(`^bridge do not need to send heartbeat`, s.bridgeDoNotNeedToSendHeartbeat)
	ctx.Step(`^heartbeat is not whitelisted`, s.heartbeatIsNotwhitelisted)
	ctx.Step(`^config status file exist for user "([^"]*)"$`, s.configStatusFileExistForUser)
	ctx.Step(`^config status is pending for user "([^"]*)"$`, s.configStatusIsPendingForUser)
	ctx.Step(`^config status is pending with failure for user "([^"]*)"$`, s.configStatusIsPendingWithFailureForUser)
	ctx.Step(`^config status succeed for user "([^"]*)"$`, s.configStatusSucceedForUser)
	ctx.Step(`^config status event "([^"]*)" is eventually send (\d+) time`, s.configStatusEventIsEventuallySendXTime)
	ctx.Step(`^config status event "([^"]*)" is not send more than (\d+) time`, s.configStatusEventIsNotSendMoreThanXTime)
	ctx.Step(`^force config status progress to be sent for user"([^"]*)"$`, s.forceConfigStatusProgressToBeSentForUser)

	// ==== CONTACT ====
	ctx.Step(`^user "([^"]*)" has contact "([^"]*)" with name "([^"]*)"$`, s.userHasContactWithName)
	ctx.Step(`^user "([^"]*)" has contacts:$`, s.userHasContacts)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has no message format$`, s.contactOfUserHasNoMessageFormat)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has message format "([^"]*)"$`, s.contactOfUserHasMessageFormat)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has no encryption scheme$`, s.contactOfUserHasNoEncryptionScheme)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has encryption scheme "([^"]*)"$`, s.contactOfUserHasEncryptionScheme)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has no signature$`, s.contactOfUserHasNoSignature)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has signature "([^"]*)"$`, s.contactOfUserHasSignature)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has no encryption$`, s.contactOfUserHasNoEncryption)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has encryption "([^"]*)"$`, s.contactOfUserHasEncryption)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has public key:$`, s.contactOfUserHasPubKey)
	ctx.Step(`^the contact "([^"]*)" of user "([^"]*)" has public key from file "([^"]*)"$`, s.contactOfUserHasPubKeyFromFile)

	// ==== OBSERVABILITY METRICS ====
	ctx.Step(`^the user with username "([^"]*)" sends the following remote notification observability metric "([^"]*)"$`,
		s.userRemoteNotificationMetricTest)
	ctx.Step(`^the user with username "([^"]*)" sends all possible observability heartbeat metrics$`, s.userHeartbeatPermutationsObservability)
	ctx.Step(`^the user with username "([^"]*)" sends all possible user distinction metrics$`, s.userDistinctionMetricsPermutationsObservability)
	ctx.Step(`^the user with username "([^"]*)" sends all possible sync message event failure observability metrics$`, s.syncFailureMessageEventsObservability)
	ctx.Step(`^the user with username "([^"]*)" sends all possible event loop message events observability metrics$`, s.eventLoopFailureMessageEventsObservability)
	ctx.Step(`^the user with username "([^"]*)" sends all possible sync message building failure observability metrics$`, s.syncFailureMessageBuiltObservability)
	ctx.Step(`^the user with username "([^"]*)" sends all possible sync message building success observability metrics$`, s.syncSuccessMessageBuiltObservability)
	// SMTP metrics
	ctx.Step(`^the user with username "([^"]*)" sends all possible SMTP error observability metrics$`, s.SMTPErrorObservabilityMetrics)
	ctx.Step(`^the user with username "([^"]*)" sends SMTP send success observability metric$`, s.SMTPSendSuccessObservabilityMetric)

	// Gluon related metrics
	ctx.Step(`^the user with username "([^"]*)" sends all possible gluon error observability metrics$`, s.testGluonErrorObservabilityMetrics)
}
