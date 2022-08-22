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


#include "GRPCService.h"
#include "MainWindow.h"
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/GRPC/EventFactory.h>


using namespace grpc;
using namespace google::protobuf;
using namespace bridgepp;

namespace
{


QString const defaultKeychain = "defaultKeychain"; ///< The default keychain.


}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCService::connectProxySignals()
{
    qtProxy_.connectSignals();
}


//****************************************************************************************************************************************************
/// \return true iff the service is streaming events.
//****************************************************************************************************************************************************
bool GRPCService::isStreaming() const
{
    QMutexLocker locker(&eventStreamMutex_);
    return isStreaming_;
}


//****************************************************************************************************************************************************
/// \param[in] request the request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::AddLogEntry(ServerContext *, AddLogEntryRequest const *request, Empty *)
{
    app().bridgeGUILog().addEntry(logLevelFromGRPC(request->level()), QString::fromStdString(request->message()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::GuiReady(ServerContext *, Empty const *, Empty *)
{
    app().log().debug(__FUNCTION__);
    app().mainWindow().settingsTab().setGUIReady(true);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Quit(ServerContext *, Empty const *, Empty *)
{
    // We do not actually quit.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Restart(ServerContext *, Empty const *, Empty *)
{
    // we do not actually restart.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ShowOnStartup(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().showOnStartup());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ShowSplashScreen(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().showSplashScreen());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsFirstGuiStart(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isFirstGUIStart());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsAutostartOn(ServerContext *, BoolValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    app().mainWindow().settingsTab().setIsAutostartOn(request->value());
    qtProxy_.sendDelayedEvent(newToggleAutostartFinishedEvent());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsAutostartOn(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isAutostartOn());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsBetaEnabled(ServerContext *, BoolValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsBetaEnabled(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsBetaEnabled(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isBetaEnabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::GoOs(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().os().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::TriggerReset(ServerContext *, Empty const *, Empty *)
{
    app().log().debug(__FUNCTION__);
    app().log().info("Bridge GUI requested a reset");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status GRPCService::Version(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().bridgeVersion().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LogsPath(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().logsPath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LicensePath(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().licensePath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ReleaseNotesPageLink(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().releaseNotesPageLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::DependencyLicensesLink(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().dependencyLicenseLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LandingPageLink(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().landingPageLink().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetColorSchemeName(ServerContext *, StringValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setColorSchemeName(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ColorSchemeName(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().colorSchemeName().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::CurrentEmailClient(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().currentEmailClient().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ForceLauncher(ServerContext *, StringValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    app().log().info(QString("ForceLauncher: %1").arg(QString::fromStdString(request->value())));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request
//****************************************************************************************************************************************************
Status GRPCService::ReportBug(ServerContext *, ReportBugRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    SettingsTab &tab = app().mainWindow().settingsTab();
    qtProxy_.reportBug(QString::fromStdString(request->ostype()), QString::fromStdString(request->osversion()),
        QString::fromStdString(request->emailclient()), QString::fromStdString(request->address()), QString::fromStdString(request->description()),
        request->includelogs());
    qtProxy_.sendDelayedEvent(tab.nextBugReportWillSucceed() ? newReportBugSuccessEvent() : newReportBugErrorEvent());
    qtProxy_.sendDelayedEvent(newReportBugFinishedEvent());

    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login(ServerContext *, LoginRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    UsersTab &usersTab = app().mainWindow().usersTab();
    loginUsername_ = QString::fromStdString(request->username());
    if (usersTab.nextUserUsernamePasswordError())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::USERNAME_PASSWORD_ERROR, "Username/password error."));
        return Status::OK;
    }
    if (usersTab.nextUserFreeUserError())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::FREE_USER, "Free user error."));
        return Status::OK;
    }
    if (usersTab.nextUserTFARequired())
    {
        qtProxy_.sendDelayedEvent(newLoginTfaRequestedEvent(loginUsername_));
        return Status::OK;
    }
    if (usersTab.nextUserTwoPasswordsRequired())
    {
        qtProxy_.sendDelayedEvent(newLoginTwoPasswordsRequestedEvent());
        return Status::OK;
    }

    SPUser const user = randomUser();
    QString const userID = user->id();
    user->setUsername(QString::fromStdString(request->username()));
    usersTab.userTable().append(user);

    if (usersTab.nextUserAlreadyLoggedIn())
        qtProxy_.sendDelayedEvent(newLoginAlreadyLoggedInEvent(userID));
    qtProxy_.sendDelayedEvent(newLoginFinishedEvent(userID));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login2FA(ServerContext *, LoginRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    UsersTab &usersTab = app().mainWindow().usersTab();
    if (usersTab.nextUserTFAError())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TFA_ERROR, "2FA Error."));
        return Status::OK;
    }
    if (usersTab.nextUserTFAAbort())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TFA_ABORT, "2FA Abort."));
        return Status::OK;
    }
    if (usersTab.nextUserTwoPasswordsRequired())
    {
        qtProxy_.sendDelayedEvent(newLoginTwoPasswordsRequestedEvent());
        return Status::OK;
    }

    SPUser const user = randomUser();
    QString const userID = user->id();
    user->setUsername(QString::fromStdString(request->username()));
    usersTab.userTable().append(user);

    if (usersTab.nextUserAlreadyLoggedIn())
        qtProxy_.sendDelayedEvent(newLoginAlreadyLoggedInEvent(userID));
    qtProxy_.sendDelayedEvent(newLoginFinishedEvent(userID));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Login2Passwords(ServerContext *, LoginRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    UsersTab &usersTab = app().mainWindow().usersTab();

    if (usersTab.nextUserTwoPasswordsError())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TWO_PASSWORDS_ERROR, "Two Passwords error."));
        return Status::OK;
    }

    if (usersTab.nextUserTwoPasswordsAbort())
    {
        qtProxy_.sendDelayedEvent(newLoginError(LoginErrorType::TWO_PASSWORDS_ABORT, "Two Passwords abort."));
        return Status::OK;
    }

    SPUser const user = randomUser();
    QString const userID = user->id();
    user->setUsername(QString::fromStdString(request->username()));
    usersTab.userTable().append(user);

    if (usersTab.nextUserAlreadyLoggedIn())
        qtProxy_.sendDelayedEvent(newLoginAlreadyLoggedInEvent(userID));
    qtProxy_.sendDelayedEvent(newLoginFinishedEvent(userID));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LoginAbort(ServerContext *, LoginAbortRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    loginUsername_ = QString();
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::CheckUpdate(ServerContext *, Empty const *, Empty *)
{
    /// \todo simulate update availability.
    app().log().debug(__FUNCTION__);
    app().log().info("Check for updates");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::InstallUpdate(ServerContext *, Empty const *, Empty *)
{
    /// Simulate update availability.
    app().log().debug(__FUNCTION__);
    app().log().info("Install update");
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsAutomaticUpdateOn(ServerContext *, BoolValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsAutomaticUpdateOn(request->value());
    return Status::OK;
}

//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsAutomaticUpdateOn(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isAutomaticUpdateOn());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::IsCacheOnDiskEnabled(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isCacheOnDiskEnabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::DiskCachePath(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().diskCachePath().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ChangeLocalCache(ServerContext *, ChangeLocalCacheRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    SettingsTab &tab = app().mainWindow().settingsTab();
    QString const path = QString::fromStdString(request->diskcachepath());

    // we mimic the behaviour of Bridge
    if (!tab.nextCacheChangeWillSucceed())
        qtProxy_.sendDelayedEvent(newCacheErrorEvent(grpc::CacheErrorType(tab.cacheError())));
    else
        qtProxy_.sendDelayedEvent(newCacheLocationChangeSuccessEvent());
    qtProxy_.sendDelayedEvent(newDiskCachePathChanged(path));
    qtProxy_.sendDelayedEvent(newIsCacheOnDiskEnabledChanged(request->enablediskcache()));
    qtProxy_.sendDelayedEvent(newChangeLocalCacheFinishedEvent());

    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetIsDoHEnabled(ServerContext *, BoolValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setIsDoHEnabled(request->value());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsDoHEnabled(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isDoHEnabled());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetUseSslForSmtp(ServerContext *, BoolValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setUseSSLForSMTP(request->value());
    qtProxy_.sendDelayedEvent(newUseSslForSmtpFinishedEvent());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::UseSslForSmtp(ServerContext *, Empty const *, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().useSSLForSMTP());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::Hostname(ServerContext *, Empty const *, StringValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().hostname().toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ImapPort(ServerContext *, Empty const *, Int32Value *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().imapPort());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SmtpPort(ServerContext *, Empty const *, Int32Value *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().smtpPort());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::ChangePorts(ServerContext *, ChangePortsRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.changePorts(request->imapport(), request->smtpport());
    qtProxy_.sendDelayedEvent(newChangePortsFinishedEvent());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \param[out] response The response.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::IsPortFree(ServerContext *, Int32Value const *request, BoolValue *response)
{
    app().log().debug(__FUNCTION__);
    response->set_value(app().mainWindow().settingsTab().isPortFree());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::AvailableKeychains(ServerContext *, Empty const *, AvailableKeychainsResponse *response)
{
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
Status GRPCService::SetCurrentKeychain(ServerContext *, StringValue const *request, Empty *)
{
    /// \todo Implement keychains configuration.
    app().log().debug(__FUNCTION__);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::CurrentKeychain(ServerContext *, Empty const *, StringValue *response)
{
    /// \todo Implement keychains configuration.
    app().log().debug(__FUNCTION__);
    response->set_value(defaultKeychain.toStdString());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[out] response The response.
/// \return The status for the call
//****************************************************************************************************************************************************
Status GRPCService::GetUserList(ServerContext *, Empty const *, UserListResponse *response)
{
    app().log().debug(__FUNCTION__);
    response->clear_users();

    QList<SPUser> userList = app().mainWindow().usersTab().userTable().users();
    RepeatedPtrField<grpc::User> *users = response->mutable_users();
    for (SPUser const &user: userList)
    {
        if (!user)
            continue;
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
Status GRPCService::GetUser(ServerContext *, StringValue const *request, grpc::User *response)
{
    app().log().debug(__FUNCTION__);
    QString userID = QString::fromStdString(request->value());
    SPUser user = app().mainWindow().usersTab().userWithID(userID);
    if (!user)
        return Status(NOT_FOUND, QString("user not found %1").arg(userID).toStdString());

    userToGRPC(*user, *response);
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::SetUserSplitMode(ServerContext *, UserSplitModeRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.setUserSplitMode(QString::fromStdString(request->userid()), request->active());
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::LogoutUser(ServerContext *, StringValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.logoutUser(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::RemoveUser(ServerContext *, StringValue const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.removeUser(QString::fromStdString(request->value()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request.
//****************************************************************************************************************************************************
Status GRPCService::ConfigureUserAppleMail(ServerContext *, ConfigureAppleMailRequest const *request, Empty *)
{
    app().log().debug(__FUNCTION__);
    qtProxy_.configureUserAppleMail(QString::fromStdString(request->userid()), QString::fromStdString(request->address()));
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] request The request
/// \param[in] writer The writer
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::RunEventStream(ServerContext *, EventStreamRequest const *request, ServerWriter<StreamEvent> *writer)
{
    app().log().debug(__FUNCTION__);
    {
        QMutexLocker locker(&eventStreamMutex_);
        if (isStreaming_)
            return { grpc::ALREADY_EXISTS, "the service is already streaming" };
        isStreaming_ = true;
        qtProxy_.setIsStreaming(true);
        qtProxy_.setClientPlatform(QString::fromStdString(request->clientplatform()));
        eventStreamShouldStop_ = false;
    }

    while (true)
    {
        QMutexLocker locker(&eventStreamMutex_);
        if (eventStreamShouldStop_)
        {
            qtProxy_.setIsStreaming(false);
            qtProxy_.setClientPlatform(QString());
            isStreaming_ = false;
            return Status::OK;
        }


        if (eventQueue_.isEmpty())
        {
            locker.unlock();
            QThread::msleep(100);
            continue;
        }
        SPStreamEvent const event = eventQueue_.front();
        eventQueue_.pop_front();
        locker.unlock();

        if (writer->Write(*event))
            app().log().debug(QString("event sent: %1").arg(QString::fromStdString(event->ShortDebugString())));
        else
            app().log().error(QString("Could not send event: %1").arg(QString::fromStdString(event->ShortDebugString())));
    }
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
Status GRPCService::StopEventStream(ServerContext *, Empty const *, Empty *)
{
    app().log().debug(__FUNCTION__);
    QMutexLocker mutex(&eventStreamMutex_);
    if (!isStreaming_)
        return Status(NOT_FOUND, "The service is not streaming");
    eventStreamShouldStop_ = true;
    return Status::OK;
}


//****************************************************************************************************************************************************
/// \param[in] event The event
/// \return true if the event was queued, and false if the server in not streaming.
//****************************************************************************************************************************************************
bool GRPCService::sendEvent(SPStreamEvent const &event)
{
    QMutexLocker mutexLocker(&eventStreamMutex_);
    if (isStreaming_)
        eventQueue_.push_back(event);
    return isStreaming_;
}




