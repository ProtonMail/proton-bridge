// Copyright (c) 2022 Proton AG
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


#ifndef BRIDGE_GUI_TESTER_EVENT_FACTORY_H
#define BRIDGE_GUI_TESTER_EVENT_FACTORY_H


#include "bridge.grpc.pb.h"
#include "GRPCUtils.h"


namespace bridgepp
{


// App events
SPStreamEvent newInternetStatusEvent(bool connected); ///< Create a new InternetStatusEvent event.
SPStreamEvent newToggleAutostartFinishedEvent(); ///< Create a new ToggleAutostartFinishedEvent event.
SPStreamEvent newResetFinishedEvent(); ///< Create a new ResetFinishedEvent event.
SPStreamEvent newReportBugFinishedEvent(); ///< Create a new ReportBugFinishedEvent event.
SPStreamEvent newReportBugSuccessEvent(); ///< Create a new ReportBugSuccessEvent event.
SPStreamEvent newReportBugErrorEvent(); ///< Create a new ReportBugErrorEvent event.
SPStreamEvent newShowMainWindowEvent(); ///< Create a new ShowMainWindowEvent event.

// Login events
SPStreamEvent newLoginError(grpc::LoginErrorType error, QString const &message); ///< Create a new LoginError event.
SPStreamEvent newLoginTfaRequestedEvent(QString const &username); ///< Create a new LoginTfaRequestedEvent event.
SPStreamEvent newLoginTwoPasswordsRequestedEvent(); ///< Create a new LoginTwoPasswordsRequestedEvent event.
SPStreamEvent newLoginFinishedEvent(QString const &userID); ///< Create a new LoginFinishedEvent event.
SPStreamEvent newLoginAlreadyLoggedInEvent(QString const &userID); ///< Create a new LoginAlreadyLoggedInEvent event.

// Update related events
SPStreamEvent newUpdateErrorEvent(grpc::UpdateErrorType errorType); ///< Create a new UpdateErrorEvent event.
SPStreamEvent newUpdateManualReadyEvent(QString const &version); ///< Create a new UpdateManualReadyEvent event.
SPStreamEvent newUpdateManualRestartNeededEvent(); ///< Create a new UpdateManualRestartNeededEvent event.
SPStreamEvent newUpdateForceEvent(QString const &version); ///< Create a new UpdateForceEvent event.
SPStreamEvent newUpdateSilentRestartNeeded(); ///< Create a new UpdateSilentRestartNeeded event.
SPStreamEvent newUpdateIsLatestVersion(); ///< Create a new UpdateIsLatestVersion event.
SPStreamEvent newUpdateCheckFinished(); ///< Create a new UpdateCheckFinished event.

// Cache on disk related events
SPStreamEvent newCacheErrorEvent(grpc::CacheErrorType errorType); ///< Create a new CacheErrorEvent event.
SPStreamEvent newCacheLocationChangeSuccessEvent(); ///< Create a new CacheLocationChangeSuccessEvent event.
SPStreamEvent newChangeLocalCacheFinishedEvent(); ///< Create a new ChangeLocalCacheFinishedEvent event.
SPStreamEvent newIsCacheOnDiskEnabledChanged(bool enabled); ///< Create a new IsCacheOnDiskEnabledChanged event.
SPStreamEvent newDiskCachePathChanged(QString const &path); ///< Create a new DiskCachePathChanged event.

// Mail settings related events
SPStreamEvent newMailSettingsErrorEvent(grpc::MailSettingsErrorType errorType); ///< Create a new MailSettingsErrorEvent event.
SPStreamEvent newUseSslForSmtpFinishedEvent(); ///< Create a new UseSslForSmtpFinishedEvent event.
SPStreamEvent newChangePortsFinishedEvent(); ///< Create a new ChangePortsFinishedEvent event.

// keychain related events
SPStreamEvent newChangeKeychainFinishedEvent(); ///< Create a new ChangeKeychainFinishedEvent event.
SPStreamEvent newHasNoKeychainEvent(); ///< Create a new HasNoKeychainEvent event.
SPStreamEvent newRebuildKeychainEvent(); ///< Create a new RebuildKeychainEvent event.

// Mail related events
SPStreamEvent newNoActiveKeyForRecipientEvent(QString const &email); ///< Create a new NoActiveKeyForRecipientEvent event.
SPStreamEvent newAddressChangedEvent(QString const &address); ///< Create a new AddressChangedEvent event.
SPStreamEvent newAddressChangedLogoutEvent(QString const &address); ///< Create a new AddressChangedLogoutEvent event.
SPStreamEvent newApiCertIssueEvent(); ///< Create a new ApiCertIssueEvent event.

// User list related event
SPStreamEvent newToggleSplitModeFinishedEvent(QString const &userID); ///< Create a new ToggleSplitModeFinishedEvent event.
SPStreamEvent newUserDisconnectedEvent(QString const &username); ///< Create a new UserDisconnectedEvent event.
SPStreamEvent newUserChangedEvent(QString const &userID); ///< Create a new UserChangedEvent event.


} // namespace bridgepp


#endif //BRIDGE_GUI_TESTER_EVENT_FACTORY_H
