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


#include "Pch.h"
#include "QMLBackend.h"
#include "Exception.h"
#include "GRPC/GRPCClient.h"
#include "Worker/Overseer.h"
#include "EventStreamWorker.h"
#include "Version.h"


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QMLBackend::QMLBackend()
    : QObject()
    , users_(new UserList(this))
{
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::init()
{
    this->connectGrpcEvents();

    QString error;
    if (app().grpc().connectToServer(error))
        app().log().info("Connected to backend via gRPC service.");
    else
        throw Exception(QString("Cannot connectToServer to go backend via gRPC: %1").arg(error));
    QString bridgeVer;
    app().grpc().version(bridgeVer);
    if (bridgeVer != PROJECT_VER)
        throw Exception(QString("Version Mismatched from Bridge (%1) and Bridge-GUI (%2)").arg(bridgeVer).arg(PROJECT_VER));

    eventStreamOverseer_ = std::make_unique<Overseer>(new EventStreamReader(nullptr), nullptr);
    eventStreamOverseer_->startWorker(true);

    // Grab from bridge the value that will not change during the execution of this app (or that will only change locally
    logGRPCCallStatus(app().grpc().showSplashScreen(showSplashScreen_), "showSplashScreen");
    logGRPCCallStatus(app().grpc().goos(goos_), "goos");
    logGRPCCallStatus(app().grpc().logsPath(logsPath_), "logsPath");
    logGRPCCallStatus(app().grpc().licensePath(licensePath_), "licensePath");

    this->retrieveUserList();
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::connectGrpcEvents()
{
    GRPCClient *client = &app().grpc();

    // app events
    connect(client, &GRPCClient::internetStatus, this, [&](bool isOn) { if (isOn) emit internetOn(); else emit internetOff(); });
    connect(client, &GRPCClient::toggleAutostartFinished, this, &QMLBackend::toggleAutostartFinished);
    connect(client, &GRPCClient::resetFinished, this, &QMLBackend::resetFinished);
    connect(client, &GRPCClient::reportBugFinished, this, &QMLBackend::reportBugFinished);
    connect(client, &GRPCClient::reportBugSuccess, this, &QMLBackend::bugReportSendSuccess);
    connect(client, &GRPCClient::reportBugError, this, &QMLBackend::bugReportSendError);

    // cache events
    connect(client, &GRPCClient::isCacheOnDiskEnabledChanged, this, &QMLBackend::isDiskCacheEnabledChanged);
    connect(client, &GRPCClient::diskCachePathChanged, this, &QMLBackend::diskCachePathChanged);
    connect(client, &GRPCClient::cacheUnavailable, this, &QMLBackend::cacheUnavailable);                                                                                            //    _ func()                  `signal:"cacheUnavailable"`
    connect(client, &GRPCClient::cacheCantMove, this, &QMLBackend::cacheCantMove);
    connect(client, &GRPCClient::diskFull, this, &QMLBackend::diskFull);
    connect(client, &GRPCClient::cacheLocationChangeSuccess, this, &QMLBackend::cacheLocationChangeSuccess);
    connect(client, &GRPCClient::changeLocalCacheFinished, this, &QMLBackend::changeLocalCacheFinished);

    // login events
    connect(client, &GRPCClient::loginUsernamePasswordError, this, &QMLBackend::loginUsernamePasswordError);
    connect(client, &GRPCClient::loginFreeUserError, this, &QMLBackend::loginFreeUserError);
    connect(client, &GRPCClient::loginConnectionError, this, &QMLBackend::loginConnectionError);
    connect(client, &GRPCClient::login2FARequested, this, &QMLBackend::login2FARequested);
    connect(client, &GRPCClient::login2FAError, this, &QMLBackend::login2FAError);
    connect(client, &GRPCClient::login2FAErrorAbort, this, &QMLBackend::login2FAErrorAbort);
    connect(client, &GRPCClient::login2PasswordRequested, this, &QMLBackend::login2PasswordRequested);
    connect(client, &GRPCClient::login2PasswordError, this, &QMLBackend::login2PasswordError);
    connect(client, &GRPCClient::login2PasswordErrorAbort, this, &QMLBackend::login2PasswordErrorAbort);
    connect(client, &GRPCClient::loginFinished, this, [&](QString const &userID) {
        qint32 const index = users_->rowOfUserID(userID); emit loginFinished(index); });
    connect(client, &GRPCClient::loginAlreadyLoggedIn, this, [&](QString const &userID) {
        qint32 const index = users_->rowOfUserID(userID); emit loginAlreadyLoggedIn(index); });

    // update events
    connect(client, &GRPCClient::updateManualError, this, &QMLBackend::updateManualError);
    connect(client, &GRPCClient::updateForceError, this, &QMLBackend::updateForceError);
    connect(client, &GRPCClient::updateSilentError, this, &QMLBackend::updateSilentError);
    connect(client, &GRPCClient::updateManualReady, this, &QMLBackend::updateManualReady);
    connect(client, &GRPCClient::updateManualRestartNeeded, this, &QMLBackend::updateManualRestartNeeded);
    connect(client, &GRPCClient::updateForce, this, &QMLBackend::updateForce);
    connect(client, &GRPCClient::updateSilentRestartNeeded, this, &QMLBackend::updateSilentRestartNeeded);
    connect(client, &GRPCClient::updateIsLatestVersion, this, &QMLBackend::updateIsLatestVersion);
    connect(client, &GRPCClient::checkUpdatesFinished, this, &QMLBackend::checkUpdatesFinished);

    // mail settings events
    connect(client, &GRPCClient::portIssueIMAP, this, &QMLBackend::portIssueIMAP);
    connect(client, &GRPCClient::portIssueSMTP, this, &QMLBackend::portIssueSMTP);
    connect(client, &GRPCClient::toggleUseSSLFinished, this, &QMLBackend::toggleUseSSLFinished);
    connect(client, &GRPCClient::changePortFinished, this, &QMLBackend::changePortFinished);

    // keychain events
    connect(client, &GRPCClient::changeKeychainFinished, this, &QMLBackend::changeKeychainFinished);
    connect(client, &GRPCClient::hasNoKeychain, this, &QMLBackend::notifyHasNoKeychain);
    connect(client, &GRPCClient::rebuildKeychain, this, &QMLBackend::notifyRebuildKeychain);

    // mail events
    connect(client, &GRPCClient::noActiveKeyForRecipient, this, &QMLBackend::noActiveKeyForRecipient);
    connect(client, &GRPCClient::addressChanged, this, &QMLBackend::addressChanged);
    connect(client, &GRPCClient::addressChangedLogout, this, &QMLBackend::addressChangedLogout);
    connect(client, &GRPCClient::apiCertIssue, this, &QMLBackend::apiCertIssue);

    // user events
    connect(client, &GRPCClient::userDisconnected, this, &QMLBackend::userDisconnected);
    users_->connectGRPCEvents();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::retrieveUserList()
{
    QList<SPUser> users;
    logGRPCCallStatus(app().grpc().getUserList(users), "getUserList");
    users_->reset(users);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::clearUserList()
{
    users_->reset();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QPoint QMLBackend::getCursorPos()
{
    return QCursor::pos();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
bool QMLBackend::isPortFree(int port)
{
    bool isFree = false;
    logGRPCCallStatus(app().grpc().isPortFree(port, isFree), "isPortFree");
    return isFree;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::guiReady()
{
    logGRPCCallStatus(app().grpc().guiReady(), "guiReady");
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::quit()
{
    logGRPCCallStatus(app().grpc().quit(), "quit");
    qApp->exit(0);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::restart()
{
    logGRPCCallStatus(app().grpc().restart(), "restart");
    app().log().error("RESTART is not implemented"); /// \todo GODT-1671 implement restart.
}


//****************************************************************************************************************************************************
/// \param[in] active Should we activate autostart.
//****************************************************************************************************************************************************
void QMLBackend::toggleAutostart(bool active)
{
    logGRPCCallStatus(app().grpc().setIsAutostartOn(active), "setIsAutostartOn");
    emit isAutostartOnChanged(this->isAutostartOn());
}


//****************************************************************************************************************************************************
/// \param[in] active The new state for the beta enabled property.
//****************************************************************************************************************************************************
void QMLBackend::toggleBeta(bool active)
{
    logGRPCCallStatus(app().grpc().setisBetaEnabled(active), "setIsBetaEnabled");
    emit isBetaEnabledChanged(this->isBetaEnabled());
}


//****************************************************************************************************************************************************
/// \param[in] scheme the scheme name
//****************************************************************************************************************************************************
void QMLBackend::changeColorScheme(QString const &scheme)
{
    logGRPCCallStatus(app().grpc().setColorSchemeName(scheme), "setIsBetaEnabled");
    emit colorSchemeNameChanged(this->colorSchemeName());
}


//****************************************************************************************************************************************************
/// \param[in] makeItActive Should SSL for SMTP be enabled.
//****************************************************************************************************************************************************
void QMLBackend::toggleUseSSLforSMTP(bool makeItActive)
{
    grpc::Status status = app().grpc().setUseSSLForSMTP(makeItActive);
    logGRPCCallStatus(status, "setUseSSLForSMTP");
    if (status.ok())
        emit useSSLforSMTPChanged(makeItActive);
}


//****************************************************************************************************************************************************
/// \param[in] imapPort The IMAP port.
/// \param[in] smtpPort The SMTP port.
//****************************************************************************************************************************************************
void QMLBackend::changePorts(int imapPort, int smtpPort)
{
    grpc::Status status = app().grpc().changePorts(imapPort, smtpPort);
    logGRPCCallStatus(status, "changePorts");
    if (status.ok())
    {
        emit portIMAPChanged(imapPort);
        emit portSMTPChanged(smtpPort);
    }
}


//****************************************************************************************************************************************************
/// \param[in] active Should DoH be active.
//****************************************************************************************************************************************************
void QMLBackend::toggleDoH(bool active)
{
    grpc::Status status = app().grpc().setIsDoHEnabled(active);
    logGRPCCallStatus(status, "toggleDoH");
    if (status.ok())
        emit isDoHEnabledChanged(active);
}


//****************************************************************************************************************************************************
/// \param[in] keychain The new keychain.
//****************************************************************************************************************************************************
void QMLBackend::changeKeychain(QString const &keychain)
{
    grpc::Status status = app().grpc().setCurrentKeychain(keychain);
    logGRPCCallStatus(status, "setCurrentKeychain");
    if (status.ok())
        emit currentKeychainChanged(keychain);
}


//****************************************************************************************************************************************************
/// \param[in] active Should automatic update be turned on.
//****************************************************************************************************************************************************
void QMLBackend::toggleAutomaticUpdate(bool active)
{
    grpc::Status status = app().grpc().setIsAutomaticUpdateOn(active);
    logGRPCCallStatus(status, "toggleAutomaticUpdate");
    if (status.ok())
        emit isAutomaticUpdateOnChanged(active);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::checkUpdates()
{
    app().grpc().checkUpdate();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::installUpdate()
{
    app().log().error(QString("%1() is not implemented.").arg(__FUNCTION__));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::triggerReset()
{
    app().log().error(QString("%1() is not implemented.").arg(__FUNCTION__));
}
