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


#include "EventFactory.h"


namespace bridgepp
{


namespace
{


//****************************************************************************************************************************************************
/// \return a new SPStreamEvent
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent newStreamEvent()
{
    return std::make_shared<grpc::StreamEvent>();
}


//****************************************************************************************************************************************************
/// \param[in] appEvent The app event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapAppEvent(grpc::AppEvent *appEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_app(appEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] loginEvent The login event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapLoginEvent(grpc::LoginEvent *loginEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_login(loginEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] updateEvent The app event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapUpdateEvent(grpc::UpdateEvent *updateEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_update(updateEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] cacheEvent The cache event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapCacheEvent(grpc::CacheEvent *cacheEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_cache(cacheEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] mailSettingsEvent The mail settings event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapMailSettingsEvent(grpc::MailSettingsEvent *mailSettingsEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_mailsettings(mailSettingsEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] keychainEvent The keychain event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapKeychainEvent(grpc::KeychainEvent *keychainEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_keychain(keychainEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] mailEvent The mail event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapMailEvent(grpc::MailEvent *mailEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_mail(mailEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] userEvent The user event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapUserEvent(grpc::UserEvent *userEvent)
{
    auto event = newStreamEvent();
    event->set_allocated_user(userEvent);
    return event;
}


} // namespace

//****************************************************************************************************************************************************
/// \param[in] connected The internet status.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newInternetStatusEvent(bool connected)
{
    auto *internetStatusEvent = new grpc::InternetStatusEvent();
    internetStatusEvent->set_connected(connected);
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_internetstatus(internetStatusEvent);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newToggleAutostartFinishedEvent()
{
    auto *event = new grpc::ToggleAutostartFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_toggleautostartfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newResetFinishedEvent()
{
    auto event = new grpc::ResetFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_resetfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugFinishedEvent()
{
    auto event = new grpc::ReportBugFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugSuccessEvent()
{
    auto event = new grpc::ReportBugSuccessEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugsuccess(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugErrorEvent()
{
    auto event = new grpc::ReportBugErrorEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugerror(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newShowMainWindowEvent()
{
    auto event = new grpc::ShowMainWindowEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_showmainwindow(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \param[in] error The error.
/// \param[in] message The message.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginError(grpc::LoginErrorType error, QString const &message)
{
    auto event = new ::grpc::LoginErrorEvent;
    event->set_type(error);
    event->set_message(message.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_error(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginTfaRequestedEvent(QString const &username)
{
    auto event = new ::grpc::LoginTfaRequestedEvent;
    event->set_username(username.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_tfarequested(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginTwoPasswordsRequestedEvent()
{
    auto event = new ::grpc::LoginTwoPasswordsRequestedEvent;
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_twopasswordrequested(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginFinishedEvent(QString const &userID)
{
    auto event = new ::grpc::LoginFinishedEvent;
    event->set_userid(userID.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_finished(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginAlreadyLoggedInEvent(QString const &userID)
{
    auto event = new ::grpc::LoginFinishedEvent;
    event->set_userid(userID.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_alreadyloggedin(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorType The error type.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateErrorEvent(grpc::UpdateErrorType errorType)
{
    auto event = new grpc::UpdateErrorEvent;
    event->set_type(errorType);
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_error(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \param[in] version The version.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateManualReadyEvent(QString const &version)
{
    auto event = new grpc::UpdateManualReadyEvent;
    event->set_version(version.toStdString());
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_manualready(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return the event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateManualRestartNeededEvent()
{
    auto event = new grpc::UpdateManualRestartNeededEvent;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_manualrestartneeded(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \param[in] version The version.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateForceEvent(QString const &version)
{
    auto event = new grpc::UpdateForceEvent;
    event->set_version(version.toStdString());
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_force(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return the event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateSilentRestartNeeded()
{
    auto event = new grpc::UpdateSilentRestartNeeded;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_silentrestartneeded(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateIsLatestVersion()
{
    auto event = new grpc::UpdateIsLatestVersion;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_islatestversion(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateCheckFinished()
{
    auto event = new grpc::UpdateCheckFinished;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_checkfinished(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorType The error type.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newCacheErrorEvent(grpc::CacheErrorType errorType)
{
    auto event = new grpc::CacheErrorEvent;
    event->set_type(errorType);
    auto cacheEvent = new grpc::CacheEvent;
    cacheEvent->set_allocated_error(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newCacheLocationChangeSuccessEvent()
{
    auto event = new grpc::CacheLocationChangeSuccessEvent;
    auto cacheEvent = new grpc::CacheEvent;
    cacheEvent->set_allocated_locationchangedsuccess(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newChangeLocalCacheFinishedEvent()
{
    auto event = new grpc::ChangeLocalCacheFinishedEvent;
    auto cacheEvent = new grpc::CacheEvent;
    cacheEvent->set_allocated_changelocalcachefinished(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \param[in] enabled The new state of the cache.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newIsCacheOnDiskEnabledChanged(bool enabled)
{
    auto event = new grpc::IsCacheOnDiskEnabledChanged;
    event->set_enabled(enabled);
    auto cacheEvent = new grpc::CacheEvent;
    cacheEvent->set_allocated_iscacheondiskenabledchanged(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \param[in] path The path of the cache.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newDiskCachePathChanged(QString const &path)
{
    auto event = new grpc::DiskCachePathChanged;
    event->set_path(path.toStdString());
    auto cacheEvent = new grpc::CacheEvent;
    cacheEvent->set_allocated_diskcachepathchanged(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorType The error type.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newMailSettingsErrorEvent(grpc::MailSettingsErrorType errorType)
{
    auto event = new grpc::MailSettingsErrorEvent;
    event->set_type(errorType);
    auto mailSettingsEvent = new grpc::MailSettingsEvent;
    mailSettingsEvent->set_allocated_error(event);
    return wrapMailSettingsEvent(mailSettingsEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUseSslForSmtpFinishedEvent()
{
    auto event = new grpc::UseSslForSmtpFinishedEvent;
    auto mailSettingsEvent = new grpc::MailSettingsEvent;
    mailSettingsEvent->set_allocated_usesslforsmtpfinished(event);
    return wrapMailSettingsEvent(mailSettingsEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newChangePortsFinishedEvent()
{
    auto event = new grpc::ChangePortsFinishedEvent;
    auto mailSettingsEvent = new grpc::MailSettingsEvent;
    mailSettingsEvent->set_allocated_changeportsfinished(event);
    return wrapMailSettingsEvent(mailSettingsEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newChangeKeychainFinishedEvent()
{
    auto event = new grpc::ChangeKeychainFinishedEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_changekeychainfinished(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newHasNoKeychainEvent()
{
    auto event = new grpc::HasNoKeychainEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_hasnokeychain(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newRebuildKeychainEvent()
{
    auto event = new grpc::RebuildKeychainEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_rebuildkeychain(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \param[in] email The email.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newNoActiveKeyForRecipientEvent(QString const &email)
{
    auto event = new grpc::NoActiveKeyForRecipientEvent;
    event->set_email(email.toStdString());
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_noactivekeyforrecipientevent(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \param[in] address The address.
/// /// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newAddressChangedEvent(QString const &address)
{
    auto event = new grpc::AddressChangedEvent;
    event->set_address(address.toStdString());
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_addresschanged(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \param[in] address The address.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newAddressChangedLogoutEvent(QString const &address)
{
    auto event = new grpc::AddressChangedLogoutEvent;
    event->set_address(address.toStdString());
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_addresschangedlogout(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newApiCertIssueEvent()
{
    auto event = new grpc::ApiCertIssueEvent;
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_apicertissue(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newToggleSplitModeFinishedEvent(QString const &userID)
{
    auto event = new grpc::ToggleSplitModeFinishedEvent;
    event->set_userid(userID.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_togglesplitmodefinished(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
/// /// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUserDisconnectedEvent(QString const &username)
{
    auto event = new grpc::UserDisconnectedEvent;
    event->set_username(username.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_userdisconnected(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUserChangedEvent(QString const &userID)
{
    auto event = new grpc::UserChangedEvent;
    event->set_userid(userID.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_userchanged(event);
    return wrapUserEvent(userEvent);
}


} // namespace bridgepp
