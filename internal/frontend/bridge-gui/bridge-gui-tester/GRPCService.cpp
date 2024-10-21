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


#include "GRPCService.h"
#include "MainWindow.h"
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/GRPC/EventFactory.h>
#include <bridgepp/GRPC/GRPCConfig.h>


using namespace grpc;
using namespace google::protobuf;
using namespace bridgepp;

namespace {

QString const defaultKeychain = "defaultKeychain"; ///< The default keychain.
QString const HV_ERROR_TEMPLATE = "failed to create new API client: 422 POST https://mail-api.proton.me/auth/v4: CAPTCHA validation failed (Code=12087, Status=422)";

}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCService::connectProxySignals() const {
    qtProxy_.connectSignals();
}


//****************************************************************************************************************************************************
/// \return true iff the service is streaming events.
//****************************************************************************************************************************************************
bool GRPCService::isStreaming() const {
    QMutexLocker locker(&eventStreamMutex_);
    return isStreaming_;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \param[out] response The response.
//****************************************************************************************************************************************************
Status GRPCService::CheckTokens(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::StringValue *response) {
    Log &log = app().log();
    log.debug(__FUNCTION__);
    GRPCConfig config;
    QString error;
    if (!config.load(QString::fromStdString(request->value()), &error)) {
        QString const err = "Could not load gRPC client config";
        log.error(err);
        return grpc::Status(StatusCode::UNAUTHENTICATED, err.toStdString());
    }

    response->set_value(config.token.toStdString());
    return grpc::Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request the request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::AddLogEntry(ServerContext *, AddLogEntryRequest const *request, Empty *) {
    app().bridgeGUILog().addEntry(logLevelFromGRPC(request->level()), QString::fromStdString(request->message()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::GuiReady(ServerContext *, Empty const *, GuiReadyResponse *response) {
    app().log().debug(__FUNCTION__);
    app().mainWindow().settingsTab().setGUIReady(true);
    response->set_showsplashscreen(app().mainWindow().settingsTab().showSplashScreen());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Quit(ServerContext *, Empty const *, Empty *) {
    // We do not actually quit.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Restart(ServerContext *, Empty const *, Empty *) {
    // we do not actually restart.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ShowOnStartup(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().showOnStartup());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsAutostartOn(ServerContext *, BoolValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    app().mainWindow().settingsTab().setIsAutostartOn(request->value());
    qtProxy_.sendDelayedEvent(newToggleAutostartFinishedEvent());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsAutostartOn(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isAutostartOn());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsBetaEnabled(ServerContext *, BoolValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsBetaEnabled(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsBetaEnabled(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isBetaEnabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsAllMailVisible(ServerContext *, BoolValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsAllMailVisible(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsAllMailVisible(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isAllMailVisible());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status GRPCService::SetIsTelemetryDisabled(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsTelemetryDisabledReceived(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status GRPCService::IsTelemetryDisabled(::grpc::ServerContext *, ::google::protobuf::Empty const *, ::google::protobuf::BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isTelemetryDisabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::GoOs(ServerContext *, Empty const*, StringValue *response) {
    response->set_value(app().mainWindow().settingsTab().os().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::TriggerReset(ServerContext *, Empty const *, Empty *) {
    app().log().debug(__FUNCTION__);
    app().log().info("Bridge GUI requested a reset");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status GRPCService::Version(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().bridgeVersion().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LogsPath(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().logsPath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LicensePath(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().licensePath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ReleaseNotesPageLink(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().releaseNotesPageLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::DependencyLicensesLink(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().dependencyLicenseLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LandingPageLink(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().landingPageLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetColorSchemeName(ServerContext *, StringValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setColorSchemeName(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ColorSchemeName(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().colorSchemeName().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::CurrentEmailClient(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().currentEmailClient().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ForceLauncher(ServerContext *, StringValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    app().log().info(QString("ForceLauncher: %1").arg(QString::fromStdString(request->value())));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetMainExecutable(ServerContext *, StringValue const *request, Empty *) {
    resetHv();
    app().log().debug(__FUNCTION__);
    app().log().info(QString("SetMainExecutable: %1").arg(QString::fromStdString(request->value())));
    return Status::OK;
}

//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status GRPCService::RequestKnowledgeBaseSuggestions(ServerContext*, StringValue const* request, Empty*) {
    QString const userInput = QString::fromUtf8(request->value());
    app().log().info(QString("RequestKnowledgeBaseSuggestions: %1").arg(userInput.left(10) + "..."));
    qtProxy_.requestKnowledgeBaseSuggestionsReceived(userInput);

    QList<bridgepp::KnowledgeBaseSuggestion> suggestions;
    for (qsizetype i = 1; i <= 3; ++i) {
        suggestions.push_back({
            .url = QString("https://proton.me/support/bridge#%1").arg(i),
            .title = QString("Suggested link %1").arg(i),
        });
    }
    qtProxy_.sendDelayedEvent(newKnowledgeBaseSuggestionsEvent(app().mainWindow().knowledgeBaseTab().getSuggestions()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request
//****************************************************************************************************************************************************
Status GRPCService::ReportBug(ServerContext *, ReportBugRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    EventsTab const&eventsTab = app().mainWindow().eventsTab();
    qtProxy_.reportBug(QString::fromStdString(request->ostype()), QString::fromStdString(request->osversion()),
        QString::fromStdString(request->emailclient()), QString::fromStdString(request->address()), QString::fromStdString(request->description()),
        request->includelogs());
    SPStreamEvent event;
    switch (eventsTab.nextBugReportResult()) {
    case EventsTab::BugReportResult::Success:
        event = newReportBugSuccessEvent();
        break;
    case EventsTab::BugReportResult::Error:
        event = newReportBugErrorEvent();
        break;
    case EventsTab::BugReportResult::DataSharingError:
        event = newReportBugFallbackEvent();
        break;
    }
    qtProxy_.sendDelayedEvent(event);
    qtProxy_.sendDelayedEvent(newReportBugFinishedEvent());

    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login(ServerContext *, LoginRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    UsersTab &usersTab = app().mainWindow().usersTab();
    loginUsername_ = QString::fromStdString(request->username());

    SPUser const &user = usersTab.userTable().userWithUsernameOrEmail(QString::fromStdString(request->username()));
    if (user) {
        qtProxy_.sendDelayedEvent(newLoginAlreadyLoggedInEvent(user->id()));
        return Status::OK;
    }

    if (usersTab.nextUserHvRequired() && !hvWasRequested_ && previousHvUsername_ != QString::fromStdString(request->username())) {
        hvWasRequested_ = true;
        previousHvUsername_ = QString::fromStdString(request->username());
        qtProxy_.sendDelayedEvent(newLoginHvRequestedEvent());
        return Status::OK;
    } else {
        hvWasRequested_ = false;
        previousHvUsername_ = "";
    }
    if (usersTab.nextUserHvError()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::HV_ERROR, HV_ERROR_TEMPLATE));
        return Status::OK;
    }
    if (usersTab.nextUserUsernamePasswordError()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::USERNAME_PASSWORD_ERROR, usersTab.usernamePasswordErrorMessage()));
        return Status::OK;
    }
    if (usersTab.nextUserFreeUserError()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::FREE_USER, "Free user error."));
        return Status::OK;
    }
    if (usersTab.nextUserTFARequired()) {
        qtProxy_.sendDelayedEvent(newLoginTfaRequestedEvent(loginUsername_));
        return Status::OK;
    }
    if (usersTab.nextUserTwoPasswordsRequired()) {
        qtProxy_.sendDelayedEvent(newLoginTwoPasswordsRequestedEvent(loginUsername_));
        return Status::OK;
    }

    this->finishLogin();
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login2FA(ServerContext *, LoginRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    UsersTab const &usersTab = app().mainWindow().usersTab();
    if (usersTab.nextUserTFAError()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TFA_ERROR, "2FA Error."));
        return Status::OK;
    }
    if (usersTab.nextUserTFAAbort()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TFA_ABORT, "2FA Abort."));
        return Status::OK;
    }
    if (usersTab.nextUserTwoPasswordsRequired()) {
        qtProxy_.sendDelayedEvent(newLoginTwoPasswordsRequestedEvent(loginUsername_));
        return Status::OK;
    }

    this->finishLogin();
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login2Passwords(ServerContext *, LoginRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    UsersTab const &usersTab = app().mainWindow().usersTab();

    if (usersTab.nextUserTwoPasswordsError()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TWO_PASSWORDS_ERROR, "Two Passwords error."));
        return Status::OK;
    }

    if (usersTab.nextUserTwoPasswordsAbort()) {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TWO_PASSWORDS_ABORT, "Two Passwords abort."));
        return Status::OK;
    }

    this->finishLogin();
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LoginAbort(ServerContext *, LoginAbortRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    this->resetHv();
    loginUsername_ = QString();
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::CheckUpdate(ServerContext *, Empty const *, Empty *) {
    /// \todo simulate update availability.
    app().log().debug(__FUNCTION__);
    app().log().info("Check for updates");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::InstallUpdate(ServerContext *, Empty const *, Empty *) {
    /// Simulate update availability.
    app().log().debug(__FUNCTION__);
    app().log().info("Install update");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsAutomaticUpdateOn(ServerContext *, BoolValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsAutomaticUpdateOn(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsAutomaticUpdateOn(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isAutomaticUpdateOn());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::DiskCachePath(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().diskCachePath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] path The path.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetDiskCachePath(ServerContext *, StringValue const *path, Empty *) {
    app().log().debug(__FUNCTION__);

    EventsTab const &eventsTab = app().mainWindow().eventsTab();
    QString const qPath = QString::fromStdString(path->value());

    // we mimic the behaviour of Bridge
    if (!eventsTab.nextCacheChangeWillSucceed()) {
        qtProxy_.sendDelayedEvent(newDiskCacheErrorEvent(static_cast<DiskCacheErrorType>(CANT_MOVE_DISK_CACHE_ERROR)));
    } else {
        qtProxy_.setDiskCachePath(qPath);
        qtProxy_.sendDelayedEvent(newDiskCachePathChangedEvent(qPath));
    }
    qtProxy_.sendDelayedEvent(newDiskCachePathChangeFinishedEvent());

    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsDoHEnabled(ServerContext *, BoolValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsDoHEnabled(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsDoHEnabled(ServerContext *, Empty const *, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isDoHEnabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] settings The IMAP/SMTP settings.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetMailServerSettings(ServerContext *, ImapSmtpSettings const *settings, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setMailServerSettings(settings->imapport(), settings->smtpport(), settings->usesslforimap(), settings->usesslforsmtp());
    qtProxy_.sendDelayedEvent(newMailServerSettingsChanged(*settings));
    qtProxy_.sendDelayedEvent(newChangeMailServerSettingsFinished());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] outSettings The settings
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::MailServerSettings(ServerContext *, Empty const *, ImapSmtpSettings *outSettings) {
    app().log().debug(__FUNCTION__);
    SettingsTab const &tab = app().mainWindow().settingsTab();
    outSettings->set_imapport(tab.imapPort());
    outSettings->set_smtpport(tab.smtpPort());
    outSettings->set_usesslforimap(tab.useSSLForIMAP());
    outSettings->set_usesslforsmtp(tab.useSSLForSMTP());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Hostname(ServerContext *, Empty const *, StringValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().hostname().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsPortFree(ServerContext *, Int32Value const *request, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().eventsTab().isPortFree());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::AvailableKeychains(ServerContext *, Empty const *, AvailableKeychainsResponse *response) {
    /// \todo Implement keychains configuration.
    app().log().debug(__FUNCTION__);
    response->clear_keychains();
    response->add_keychains(defaultKeychain.toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::SetCurrentKeychain(ServerContext *, StringValue const *request, Empty *) {
    /// \todo Implement keychains configuration.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::CurrentKeychain(ServerContext *, Empty const *, StringValue *response) {
    /// \todo Implement keychains configuration.
    app().log().debug(__FUNCTION__);
    response->set_value(defaultKeychain.toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::GetUserList(ServerContext *, Empty const *, UserListResponse *response) {
    app().log().debug(__FUNCTION__);
    response->clear_users();

    QList<SPUser> userList = app().mainWindow().usersTab().userTable().users();
    RepeatedPtrField<grpc::User> *users = response->mutable_users();
    for (SPUser const &user: userList) {
        if (!user) {
            continue;
        }
        users->Add();
        grpc::User &grpcUser = (*users)[users->size() - 1];
        userToGRPC(*user, grpcUser);
    }

    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::GetUser(ServerContext *, StringValue const *request, grpc::User *response) {
    app().log().debug(__FUNCTION__);
    QString const userID = QString::fromStdString(request->value());
    SPUser const user = app().mainWindow().usersTab().userWithID(userID);
    if (!user) {
        return Status(NOT_FOUND, QString("user not found %1").arg(userID).toStdString());
    }

    userToGRPC(*user, *response);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetUserSplitMode(ServerContext *, UserSplitModeRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.setUserSplitMode(QString::fromStdString(request->userid()), request->active());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SendBadEventUserFeedback(ServerContext *, UserBadEventFeedbackRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.sendBadEventUserFeedback(QString::fromStdString(request->userid()), request->doresync());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LogoutUser(ServerContext *, StringValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.logoutUser(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::RemoveUser(ServerContext *, StringValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.removeUser(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
//****************************************************************************************************************************************************
Status GRPCService::ConfigureUserAppleMail(ServerContext *, ConfigureAppleMailRequest const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    qtProxy_.configureUserAppleMail(QString::fromStdString(request->userid()), QString::fromStdString(request->address()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ExportTLSCertificates(ServerContext *, StringValue const *request, Empty *) {
    app().log().debug(__FUNCTION__);
    SettingsTab const &tab = app().mainWindow().settingsTab();
    if (!tab.nextTLSCertExportWillSucceed()) {
        qtProxy_.sendDelayedEvent(newGenericErrorEvent(TLS_CERT_EXPORT_ERROR));
    }
    if (!tab.nextTLSKeyExportWillSucceed()) {
        qtProxy_.sendDelayedEvent(newGenericErrorEvent(TLS_KEY_EXPORT_ERROR));
    }
    qtProxy_.exportTLSCertificates(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The reponse.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsTLSCertificateInstalled(ServerContext *, const Empty *request, BoolValue *response) {
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isTLSCertificateInstalled());
    return Status::OK;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Status GRPCService::InstallTLSCertificate(ServerContext *, Empty const *, Empty *) {
    app().log().debug(__FUNCTION__);
    SPStreamEvent event;
    qtProxy_.installTLSCertificate();
    switch (app().mainWindow().settingsTab().nextTLSCertInstallResult()) {
    case SettingsTab::TLSCertInstallResult::Success:
        event = newCertificateInstallSuccessEvent();
        break;
    case SettingsTab::TLSCertInstallResult::Canceled:
        event = newCertificateInstallCanceledEvent();
        break;
    default:
        event = newCertificateInstallFailedEvent();
        break;
    }
    qtProxy_.sendDelayedEvent(event);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
//****************************************************************************************************************************************************
Status GRPCService::ExternalLinkClicked(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) {
    app().log().debug(QString("%1 - URL = %2").arg(__FUNCTION__, QString::fromStdString(request->value())));
    return Status::OK;
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Status GRPCService::ReportBugClicked(::grpc::ServerContext *, ::google::protobuf::Empty const *, ::google::protobuf::Empty *) {
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
//****************************************************************************************************************************************************
Status GRPCService::AutoconfigClicked(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) {
    app().log().debug(QString("%1 - Client = %2").arg(__FUNCTION__, QString::fromStdString(request->value())));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request
/// \param[in] writer The writer
//****************************************************************************************************************************************************
Status GRPCService::RunEventStream(ServerContext *ctx, EventStreamRequest const *request, ServerWriter<StreamEvent> *writer) {
    app().log().debug(__FUNCTION__);
    {
        QMutexLocker locker(&eventStreamMutex_);
        if (isStreaming_) {
            return { grpc::ALREADY_EXISTS, "the service is already streaming" };
        }
        isStreaming_ = true;
        qtProxy_.setIsStreaming(true);
        qtProxy_.setClientPlatform(QString::fromStdString(request->clientplatform()));
        eventStreamShouldStop_ = false;
    }

    while (true) {
        QMutexLocker locker(&eventStreamMutex_);
        if (eventStreamShouldStop_ || ctx->IsCancelled()) {
            qtProxy_.setIsStreaming(false);
            qtProxy_.setClientPlatform(QString());
            isStreaming_ = false;
            return Status::OK;
        }

        if (eventQueue_.isEmpty()) {
            locker.unlock();
            QThread::msleep(100);
            continue;
        }

        SPStreamEvent const event = eventQueue_.front();
        eventQueue_.pop_front();
        locker.unlock();

        if (writer->Write(*event)) {
            app().log().debug(QString("event sent: %1").arg(QString::fromStdString(event->ShortDebugString())));
        } else {
            app().log().error(QString("Could not send event: %1").arg(QString::fromStdString(event->ShortDebugString())));
        }
    }
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::StopEventStream(ServerContext *, Empty const *, Empty *) {
    app().log().debug(__FUNCTION__);
    QMutexLocker mutex(&eventStreamMutex_);
    if (!isStreaming_) {
        return Status(NOT_FOUND, "The service is not streaming");
    }
    eventStreamShouldStop_ = true;
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] event The event
/// \return true if the event was queued, and false if the server in not streaming.
//****************************************************************************************************************************************************
bool GRPCService::sendEvent(SPStreamEvent const &event) {
    QMutexLocker mutexLocker(&eventStreamMutex_);
    if (isStreaming_) {
        eventQueue_.push_back(event);
    }
    return isStreaming_;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCService::finishLogin() {
    UsersTab &usersTab = app().mainWindow().usersTab();
    SPUser user = usersTab.userWithUsernameOrEmail(loginUsername_);
    bool const alreadyExist = user.get();
    if (!user) {
        user = randomUser();
        user->setUsername(loginUsername_);
        usersTab.userTable().append(user);
    } else {
        if (user->state() == EUserState::State::Connected) {
            qtProxy_.sendDelayedEvent(newLoginAlreadyLoggedInEvent(user->id()));
        } else {
            user->setState(EUserState::State::Connected);
        }
    }

    qtProxy_.sendDelayedEvent(newLoginFinishedEvent(user->id(), alreadyExist));
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCService::resetHv() {
    hvWasRequested_ = false;
    previousHvUsername_ = "";
}
