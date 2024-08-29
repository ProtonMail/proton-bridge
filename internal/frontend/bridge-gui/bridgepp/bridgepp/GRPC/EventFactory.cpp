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


#include "EventFactory.h"


namespace bridgepp {


namespace {


//****************************************************************************************************************************************************
/// \return a new SPStreamEvent
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent newStreamEvent() {
    return std::make_shared<grpc::StreamEvent>();
}


//****************************************************************************************************************************************************
/// \param[in] appEvent The app event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapAppEvent(grpc::AppEvent *appEvent) {
    auto event = newStreamEvent();
    event->set_allocated_app(appEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] loginEvent The login event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapLoginEvent(grpc::LoginEvent *loginEvent) {
    auto event = newStreamEvent();
    event->set_allocated_login(loginEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] updateEvent The app event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapUpdateEvent(grpc::UpdateEvent *updateEvent) {
    auto event = newStreamEvent();
    event->set_allocated_update(updateEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] cacheEvent The cache event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapCacheEvent(grpc::DiskCacheEvent *cacheEvent) {
    auto event = newStreamEvent();
    event->set_allocated_cache(cacheEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] mailSettingsEvent The mail settings event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapMailServerSettingsEvent(grpc::MailServerSettingsEvent *mailServerSettingsEvent) {
    auto event = newStreamEvent();
    event->set_allocated_mailserversettings(mailServerSettingsEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] keychainEvent The keychain event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapKeychainEvent(grpc::KeychainEvent *keychainEvent) {
    auto event = newStreamEvent();
    event->set_allocated_keychain(keychainEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] mailEvent The mail event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapMailEvent(grpc::MailEvent *mailEvent) {
    auto event = newStreamEvent();
    event->set_allocated_mail(mailEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] userEvent The user event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapUserEvent(grpc::UserEvent *userEvent) {
    auto event = newStreamEvent();
    event->set_allocated_user(userEvent);
    return event;
}


//****************************************************************************************************************************************************
/// \param[in] genericErrorEvent The generic error event.
/// \return The stream event.
//****************************************************************************************************************************************************
bridgepp::SPStreamEvent wrapGenericErrorEvent(grpc::GenericErrorEvent *genericErrorEvent) {
    auto event = newStreamEvent();
    event->set_allocated_genericerror(genericErrorEvent);
    return event;
}


} // namespace

//****************************************************************************************************************************************************
/// \param[in] connected The internet status.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newInternetStatusEvent(bool connected) {
    auto *internetStatusEvent = new grpc::InternetStatusEvent();
    internetStatusEvent->set_connected(connected);
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_internetstatus(internetStatusEvent);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newToggleAutostartFinishedEvent() {
    auto *event = new grpc::ToggleAutostartFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_toggleautostartfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newResetFinishedEvent() {
    auto event = new grpc::ResetFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_resetfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugFinishedEvent() {
    auto event = new grpc::ReportBugFinishedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugfinished(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugSuccessEvent() {
    auto event = new grpc::ReportBugSuccessEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugsuccess(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugErrorEvent() {
    auto event = new grpc::ReportBugErrorEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugerror(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newReportBugFallbackEvent() {
    auto event = new grpc::ReportBugFallbackEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_reportbugfallback(event);
    return wrapAppEvent(appEvent);
}



//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newCertificateInstallSuccessEvent() {
    auto event = new grpc::CertificateInstallSuccessEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_certificateinstallsuccess(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newCertificateInstallCanceledEvent() {
    auto event = new grpc::CertificateInstallCanceledEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_certificateinstallcanceled(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newCertificateInstallFailedEvent() {
    auto event = new grpc::CertificateInstallFailedEvent;
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_certificateinstallfailed(event);
    return wrapAppEvent(appEvent);
}


//****************************************************************************************************************************************************
/// \param[in] suggestions the suggestions
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newKnowledgeBaseSuggestionsEvent(QList<KnowledgeBaseSuggestion> const& suggestions) {
    auto event = new grpc::KnowledgeBaseSuggestionsEvent;
    for (KnowledgeBaseSuggestion const &suggestion: suggestions) {
            grpc::KnowledgeBaseSuggestion *s = event->add_suggestions();
            s->set_url(suggestion.url.toStdString());
            s->set_title(suggestion.title.toStdString());
    }
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_knowledgebasesuggestions(event);
    return wrapAppEvent(appEvent);
}

//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newShowMainWindowEvent() {
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
SPStreamEvent newLoginError(grpc::LoginErrorType error, QString const &message) {
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
SPStreamEvent newLoginTfaRequestedEvent(QString const &username) {
    auto event = new ::grpc::LoginTfaRequestedEvent;
    event->set_username(username.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_tfarequested(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginHvRequestedEvent() {
        auto event = new ::grpc::LoginHvRequestedEvent;
        event->set_hvurl("https://verify.proton.me/?methods=captcha&token=SOME_RANDOM_TOKEN");
        auto loginEvent = new grpc::LoginEvent;
        loginEvent->set_allocated_hvrequested(event);
        return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginTwoPasswordsRequestedEvent(QString const &username) {
    auto event = new ::grpc::LoginTwoPasswordsRequestedEvent;
    event->set_username(username.toStdString());
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_twopasswordrequested(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] wasSignedOut Was the user signed-out.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginFinishedEvent(QString const &userID, bool wasSignedOut) {
    auto event = new ::grpc::LoginFinishedEvent;
    event->set_userid(userID.toStdString());
    event->set_wassignedout(wasSignedOut);
    auto loginEvent = new grpc::LoginEvent;
    loginEvent->set_allocated_finished(event);
    return wrapLoginEvent(loginEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newLoginAlreadyLoggedInEvent(QString const &userID) {
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
SPStreamEvent newUpdateErrorEvent(grpc::UpdateErrorType errorType) {
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
SPStreamEvent newUpdateManualReadyEvent(QString const &version) {
    auto event = new grpc::UpdateManualReadyEvent;
    event->set_version(version.toStdString());
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_manualready(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return the event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateManualRestartNeededEvent() {
    auto event = new grpc::UpdateManualRestartNeededEvent;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_manualrestartneeded(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \param[in] version The version.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateForceEvent(QString const &version) {
    auto event = new grpc::UpdateForceEvent;
    event->set_version(version.toStdString());
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_force(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return the event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateSilentRestartNeededEvent() {
    auto event = new grpc::UpdateSilentRestartNeeded;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_silentrestartneeded(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateIsLatestVersionEvent() {
    auto event = new grpc::UpdateIsLatestVersion;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_islatestversion(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateCheckFinishedEvent() {
    auto event = new grpc::UpdateCheckFinished;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_checkfinished(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUpdateVersionChangedEvent() {
    auto event = new grpc::UpdateVersionChanged;
    auto updateEvent = new grpc::UpdateEvent;
    updateEvent->set_allocated_versionchanged(event);
    return wrapUpdateEvent(updateEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorType The error type.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newDiskCacheErrorEvent(grpc::DiskCacheErrorType errorType) {
    auto event = new grpc::DiskCacheErrorEvent;
    event->set_type(errorType);
    auto cacheEvent = new grpc::DiskCacheEvent;
    cacheEvent->set_allocated_error(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \param[in] path The path of the cache.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newDiskCachePathChangedEvent(QString const &path) {
    auto event = new grpc::DiskCachePathChangedEvent;
    event->set_path(path.toStdString());
    auto cacheEvent = new grpc::DiskCacheEvent;
    cacheEvent->set_allocated_pathchanged(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newDiskCachePathChangeFinishedEvent() {
    auto event = new grpc::DiskCachePathChangeFinishedEvent;
    auto cacheEvent = new grpc::DiskCacheEvent;
    cacheEvent->set_allocated_pathchangefinished(event);
    return wrapCacheEvent(cacheEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorType The error type.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newMailServerSettingsErrorEvent(grpc::MailServerSettingsErrorType errorType) {
    auto event = new grpc::MailServerSettingsErrorEvent;
    event->set_type(errorType);
    auto mailServerSettingsEvent = new grpc::MailServerSettingsEvent;
    mailServerSettingsEvent->set_allocated_error(event);
    return wrapMailServerSettingsEvent(mailServerSettingsEvent);
}


//****************************************************************************************************************************************************
/// \param[in] settings The settings.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newMailServerSettingsChanged(grpc::ImapSmtpSettings const &settings) {
    auto event = new grpc::MailServerSettingsChangedEvent;
    event->set_allocated_settings(new grpc::ImapSmtpSettings(settings));
    auto mailServerSettingsEvent = new grpc::MailServerSettingsEvent;
    mailServerSettingsEvent->set_allocated_mailserversettingschanged(event);
    return wrapMailServerSettingsEvent(mailServerSettingsEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newChangeMailServerSettingsFinished() {
    auto event = new grpc::ChangeMailServerSettingsFinishedEvent;
    auto mailServerSettingsEvent = new grpc::MailServerSettingsEvent;
    mailServerSettingsEvent->set_allocated_changemailserversettingsfinished(event);
    return wrapMailServerSettingsEvent(mailServerSettingsEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newChangeKeychainFinishedEvent() {
    auto event = new grpc::ChangeKeychainFinishedEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_changekeychainfinished(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newHasNoKeychainEvent() {
    auto event = new grpc::HasNoKeychainEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_hasnokeychain(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newRebuildKeychainEvent() {
    auto event = new grpc::RebuildKeychainEvent;
    auto keychainEvent = new grpc::KeychainEvent;
    keychainEvent->set_allocated_rebuildkeychain(event);
    return wrapKeychainEvent(keychainEvent);
}


//****************************************************************************************************************************************************
/// \param[in] address The address.
/// /// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newAddressChangedEvent(QString const &address) {
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
SPStreamEvent newAddressChangedLogoutEvent(QString const &address) {
    auto event = new grpc::AddressChangedLogoutEvent;
    event->set_address(address.toStdString());
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_addresschangedlogout(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newApiCertIssueEvent() {
    auto event = new grpc::ApiCertIssueEvent;
    auto mailEvent = new grpc::MailEvent;
    mailEvent->set_allocated_apicertissue(event);
    return wrapMailEvent(mailEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newToggleSplitModeFinishedEvent(QString const &userID) {
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
SPStreamEvent newUserDisconnectedEvent(QString const &username) {
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
SPStreamEvent newUserChangedEvent(QString const &userID) {
    auto event = new grpc::UserChangedEvent;
    event->set_userid(userID.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_userchanged(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] errorMessage The errorMessage
//****************************************************************************************************************************************************
SPStreamEvent newUserBadEvent(QString const &userID, QString const &errorMessage) {
    auto event = new grpc::UserBadEvent;
    event->set_userid(userID.toStdString());
    event->set_errormessage(errorMessage.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_userbadevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] usedBytes The number of used bytes.
//****************************************************************************************************************************************************
SPStreamEvent newUsedBytesChangedEvent(QString const &userID, qint64 usedBytes) {
    auto event = new grpc::UsedBytesChangedEvent;
    event->set_userid(userID.toStdString());
    event->set_usedbytes(usedBytes);
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_usedbyteschangedevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] username The username that was provided for the failed IMAP login attempt.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newIMAPLoginFailedEvent(QString const &username) {
    auto event = new grpc::ImapLoginFailedEvent;
    event->set_username(username.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_imaploginfailedevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newSyncStartedEvent(QString const &userID) {
    auto event = new grpc::SyncStartedEvent;
    event->set_userid(userID.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_syncstartedevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newSyncFinishedEvent(QString const &userID) {
    auto event = new grpc::SyncFinishedEvent;
    event->set_userid(userID.toStdString());
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_syncfinishedevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] progress The progress ratio.
/// \param[in] elapsedMs The elapsed time in milliseconds.
/// \param[in] remainingMs The remaining time in milliseconds.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newSyncProgressEvent(QString const &userID, double progress, qint64 elapsedMs, qint64 remainingMs) {
    auto event = new grpc::SyncProgressEvent;
    event->set_userid(userID.toStdString());
    event->set_progress(progress);
    event->set_elapsedms(elapsedMs);
    event->set_remainingms(remainingMs);
    auto userEvent = new grpc::UserEvent;
    userEvent->set_allocated_syncprogressevent(event);
    return wrapUserEvent(userEvent);
}


//****************************************************************************************************************************************************
/// \param[in] errorCode The error errorCode.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newGenericErrorEvent(grpc::ErrorCode errorCode) {
    auto event = new grpc::GenericErrorEvent;
    event->set_code(errorCode);
    return wrapGenericErrorEvent(event);
}


//****************************************************************************************************************************************************
/// \param[in] userID The user ID that received the notification.
/// \param[in] title The title of the notification.
/// \param[in] subtitle The subtitle of the notification.
/// \param[in] body The body of the notification.
/// \return The event.
//****************************************************************************************************************************************************
SPStreamEvent newUserNotificationEvent(QString const &userID, QString const title, QString const subtitle, QString const body) {
    auto event = new grpc::UserNotificationEvent;
    event->set_userid(userID.toStdString());
    event->set_body(body.toStdString());
    event->set_subtitle(subtitle.toStdString());
    event->set_title(title.toStdString());
    auto appEvent = new grpc::AppEvent;
    appEvent->set_allocated_usernotification(event);
    return wrapAppEvent(appEvent);
}


} // namespace bridgepp
