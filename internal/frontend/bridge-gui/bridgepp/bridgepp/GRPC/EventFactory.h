// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
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


#ifndef BRIDGE_PP_EVENT_FACTORY_H
#define BRIDGE_PP_EVENT_FACTORY_H


#include "bridge.grpc.pb.h"
#include "GRPCUtils.h"
#include <bridgepp/GRPC/GRPCClient.h>


namespace bridgepp {


// App events
SPStreamEvent newInternetStatusEvent(bool connected); ///< Create a new InternetStatusEvent event.
SPStreamEvent newToggleAutostartFinishedEvent(); ///< Create a new ToggleAutostartFinishedEvent event.
SPStreamEvent newResetFinishedEvent(); ///< Create a new ResetFinishedEvent event.
SPStreamEvent newReportBugFinishedEvent(); ///< Create a new ReportBugFinishedEvent event.
SPStreamEvent newReportBugSuccessEvent(); ///< Create a new ReportBugSuccessEvent event.
SPStreamEvent newReportBugErrorEvent(); ///< Create a new ReportBugErrorEvent event.
SPStreamEvent newReportBugFallbackEvent(); ///< Create a new ReportBugFallbackEvent event.
SPStreamEvent newCertificateInstallSuccessEvent(); ///< Create a new CertificateInstallSuccessEvent event.
SPStreamEvent newCertificateInstallCanceledEvent(); ///< Create a new CertificateInstallCanceledEvent event.
SPStreamEvent newCertificateInstallFailedEvent(); ///< Create anew CertificateInstallFailedEvent event.
SPStreamEvent newShowMainWindowEvent(); ///< Create a new ShowMainWindowEvent event.
SPStreamEvent newKnowledgeBaseSuggestionsEvent(QList<KnowledgeBaseSuggestion> const& suggestions); ///< Create a new KnowledgeBaseSuggestions event.

// Login events
SPStreamEvent newLoginError(grpc::LoginErrorType error, QString const &message); ///< Create a new LoginError event.
SPStreamEvent newLoginTfaRequestedEvent(QString const &username); ///< Create a new LoginTfaRequestedEvent event.
SPStreamEvent newLoginTwoPasswordsRequestedEvent(QString const &username); ///< Create a new LoginTwoPasswordsRequestedEvent event.
SPStreamEvent newLoginFinishedEvent(QString const &userID, bool wasSignedOut); ///< Create a new LoginFinishedEvent event.
SPStreamEvent newLoginAlreadyLoggedInEvent(QString const &userID); ///< Create a new LoginAlreadyLoggedInEvent event.
SPStreamEvent newLoginHvRequestedEvent(); ///< Create a new LoginHvRequestedEvent

// Update related events
SPStreamEvent newUpdateErrorEvent(grpc::UpdateErrorType errorType); ///< Create a new UpdateErrorEvent event.
SPStreamEvent newUpdateManualReadyEvent(QString const &version); ///< Create a new UpdateManualReadyEvent event.
SPStreamEvent newUpdateManualRestartNeededEvent(); ///< Create a new UpdateManualRestartNeededEvent event.
SPStreamEvent newUpdateForceEvent(QString const &version); ///< Create a new UpdateForceEvent event.
SPStreamEvent newUpdateSilentRestartNeededEvent(); ///< Create a new UpdateSilentRestartNeeded event.
SPStreamEvent newUpdateIsLatestVersionEvent(); ///< Create a new UpdateIsLatestVersion event.
SPStreamEvent newUpdateCheckFinishedEvent(); ///< Create a new UpdateCheckFinished event.
SPStreamEvent newUpdateVersionChangedEvent(); ///< Create a new updateVersionChanged event.

// Cache on disk related events
SPStreamEvent newDiskCacheErrorEvent(grpc::DiskCacheErrorType errorType); ///< Create a new DiskCacheErrorEvent event.
SPStreamEvent newDiskCachePathChangedEvent(QString const &path); ///< Create a new DiskCachePathChanged event.
SPStreamEvent newDiskCachePathChangeFinishedEvent(); ///< Create a new DiskCachePathChangeFinishedEvent event.

// Mail settings related events
SPStreamEvent newMailServerSettingsErrorEvent(grpc::MailServerSettingsErrorType errorType); ///< Create a new MailSettingsErrorEvent event.
SPStreamEvent newMailServerSettingsChanged(grpc::ImapSmtpSettings const &settings); ///< Create a new ConnectionModeChanged event.
SPStreamEvent newChangeMailServerSettingsFinished(); ///< Create a new ChangeMailServerSettingsFinished event.

// keychain related events
SPStreamEvent newChangeKeychainFinishedEvent(); ///< Create a new ChangeKeychainFinishedEvent event.
SPStreamEvent newHasNoKeychainEvent(); ///< Create a new HasNoKeychainEvent event.
SPStreamEvent newRebuildKeychainEvent(); ///< Create a new RebuildKeychainEvent event.

// Mail related events
SPStreamEvent newAddressChangedEvent(QString const &address); ///< Create a new AddressChangedEvent event.
SPStreamEvent newAddressChangedLogoutEvent(QString const &address); ///< Create a new AddressChangedLogoutEvent event.
SPStreamEvent newApiCertIssueEvent(); ///< Create a new ApiCertIssueEvent event.

// User list related event
SPStreamEvent newToggleSplitModeFinishedEvent(QString const &userID); ///< Create a new ToggleSplitModeFinishedEvent event.
SPStreamEvent newUserDisconnectedEvent(QString const &username); ///< Create a new UserDisconnectedEvent event.
SPStreamEvent newUserChangedEvent(QString const &userID); ///< Create a new UserChangedEvent event.
SPStreamEvent newUserBadEvent(QString const &userID, QString const& errorMessage); ///< Create a new UserBadEvent event.
SPStreamEvent newUsedBytesChangedEvent(QString const &userID, qint64 usedBytes);  ///< Create a new UsedBytesChangedEvent event.
SPStreamEvent newIMAPLoginFailedEvent(QString const &username); ///< Create a new ImapLoginFailedEvent event.
SPStreamEvent newSyncStartedEvent(QString const &userID); ///< Create a new SyncStarted event.
SPStreamEvent newSyncFinishedEvent(QString const &userID); ///< Create a new SyncFinished event.
SPStreamEvent newSyncProgressEvent(QString const &userID, double progress, qint64 elapsedMs, qint64 remainingMs); ///< Create a new SyncFinished event.

// Generic error event
SPStreamEvent newGenericErrorEvent(grpc::ErrorCode errorCode); ///< Create a new GenericErrrorEvent event.

// User notification event
SPStreamEvent newUserNotificationEvent(QString const &userID, QString const title, QString const subtitle, QString const body);

} // namespace bridgepp


#endif //BRIDGE_PP_EVENT_FACTORY_H
