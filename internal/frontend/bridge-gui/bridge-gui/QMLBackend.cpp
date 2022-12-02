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


#include "QMLBackend.h"
#include "EventStreamWorker.h"
#include "Version.h"
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/Worker/Overseer.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QMLBackend::QMLBackend()
    : QObject()
{
}


//****************************************************************************************************************************************************
/// \param[in] serviceConfig
//****************************************************************************************************************************************************
void QMLBackend::init(GRPCConfig const &serviceConfig)
{
    users_ = new UserList(this);

    Log& log = app().log();
    log.info(QString("Connecting to gRPC service"));
    app().grpc().setLog(&log);
    this->connectGrpcEvents();

    QString error;
    if (app().grpc().connectToServer(serviceConfig, app().bridgeMonitor(), error))
        app().log().info("Connected to backend via gRPC service.");
    else
        throw Exception(QString("Cannot connectToServer to go backend via gRPC: %1").arg(error));

    QString bridgeVer;
    app().grpc().version(bridgeVer);
    if (bridgeVer != PROJECT_VER)
        throw Exception(QString("Version Mismatched from Bridge (%1) and Bridge-GUI (%2)").arg(bridgeVer, PROJECT_VER));

    eventStreamOverseer_ = std::make_unique<Overseer>(new EventStreamReader(nullptr), nullptr);
    eventStreamOverseer_->startWorker(true);

    connect(&app().log(), &Log::entryAdded, [&](Log::Level level, QString const& message) {
        app().grpc().addLogEntry(level, "frontend/bridge-gui", message);
    });

    // Grab from bridge the value that will not change during the execution of this app (or that will only change locally
    app().grpc().showSplashScreen(showSplashScreen_);
    app().grpc().goos(goos_);
    app().grpc().logsPath(logsPath_);
    app().grpc().licensePath(licensePath_);
    bool sslForIMAP = false, sslForSMTP = false;
    int imapPort = 0, smtpPort = 0;
    app().grpc().mailServerSettings(imapPort, smtpPort, sslForIMAP, sslForSMTP);
    this->setIMAPPort(imapPort);
    this->setSMTPPort(smtpPort);
    this->setUseSSLForIMAP(sslForIMAP);
    this->setUseSSLForSMTP(sslForSMTP);
    this->retrieveUserList();
}


//****************************************************************************************************************************************************
/// \param timeoutMs The timeout after which the function should return false if the event stream reader is not finished. if -1 one, the function
/// never times out.
/// \return false if and only if the timeout delay was reached.
//****************************************************************************************************************************************************
bool QMLBackend::waitForEventStreamReaderToFinish(qint32 timeoutMs)
{
    return eventStreamOverseer_->wait(timeoutMs);
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
    connect(client, &GRPCClient::resetFinished, this, &QMLBackend::onResetFinished);
    connect(client, &GRPCClient::reportBugFinished, this, &QMLBackend::reportBugFinished);
    connect(client, &GRPCClient::reportBugSuccess, this, &QMLBackend::bugReportSendSuccess);
    connect(client, &GRPCClient::reportBugError, this, &QMLBackend::bugReportSendError);
    connect(client, &GRPCClient::showMainWindow, this, &QMLBackend::showMainWindow);

    // cache events
    connect(client, &GRPCClient::diskCacheUnavailable, this, &QMLBackend::diskCacheUnavailable);
    connect(client, &GRPCClient::cantMoveDiskCache, this, &QMLBackend::cantMoveDiskCache);
    connect(client, &GRPCClient::diskFull, this, &QMLBackend::diskFull);
    connect(client, &GRPCClient::diskCachePathChanged, this, &QMLBackend::diskCachePathChanged);
    connect(client, &GRPCClient::diskCachePathChangeFinished, this, &QMLBackend::diskCachePathChangeFinished);

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
    connect(client, &GRPCClient::loginFinished, this, &QMLBackend::onLoginFinished);
    connect(client, &GRPCClient::loginAlreadyLoggedIn, this, &QMLBackend::onLoginAlreadyLoggedIn);

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
    connect(client, &GRPCClient::updateVersionChanged, this, &QMLBackend::onVersionChanged);

    // mail settings events
    connect(client, &GRPCClient::imapPortStartupError, this, &QMLBackend::imapPortStartupError);
    connect(client, &GRPCClient::smtpPortStartupError, this, &QMLBackend::smtpPortStartupError);
    connect(client, &GRPCClient::imapPortChangeError, this, &QMLBackend::imapPortChangeError);
    connect(client, &GRPCClient::smtpPortChangeError, this, &QMLBackend::smtpPortChangeError);
    connect(client, &GRPCClient::imapConnectionModeChangeError, this, &QMLBackend::imapConnectionModeChangeError);
    connect(client, &GRPCClient::smtpConnectionModeChangeError, this, &QMLBackend::smtpConnectionModeChangeError);
    connect(client, &GRPCClient::mailServerSettingsChanged, this, &QMLBackend::onMailServerSettingsChanged);
    connect(client, &GRPCClient::changeMailServerSettingsFinished, this, &QMLBackend::changeMailServerSettingsFinished);

    // keychain events
    connect(client, &GRPCClient::changeKeychainFinished, this, &QMLBackend::changeKeychainFinished);
    connect(client, &GRPCClient::hasNoKeychain, this, &QMLBackend::notifyHasNoKeychain);
    connect(client, &GRPCClient::rebuildKeychain, this, &QMLBackend::notifyRebuildKeychain);

    // mail events
    connect(client, &GRPCClient::noActiveKeyForRecipient, this, &QMLBackend::noActiveKeyForRecipient);
    connect(client, &GRPCClient::addressChanged, this, &QMLBackend::addressChanged);
    connect(client, &GRPCClient::addressChangedLogout, this, &QMLBackend::addressChangedLogout);
    connect(client, &GRPCClient::apiCertIssue, this, &QMLBackend::apiCertIssue);

    // generic error events
    connect(client, &GRPCClient::genericError, this, &QMLBackend::onGenericError);

    // user events
    connect(client, &GRPCClient::userDisconnected, this, &QMLBackend::userDisconnected);
    users_->connectGRPCEvents();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::retrieveUserList()
{
    QList<SPUser> newUsers;
    app().grpc().getUserList(newUsers);

    // As we want to use shared pointers here, we do not want to use the Qt ownership system, so we set parent to nil.
    // But: From https://doc.qt.io/qt-5/qtqml-cppintegration-data.html:
    // " When data is transferred from C++ to QML, the ownership of the data always remains with C++. The exception to this rule
    // is when a QObject is returned from an explicit C++ method call: in this case, the QML engine assumes ownership of the object. "
    // This is the case here, so we explicitly indicate that the object is owned by C++.
    for (SPUser const& user: newUsers)

    for (qsizetype i = 0; i < newUsers.size(); ++i)
    {
        SPUser newUser = newUsers[i];
        SPUser existingUser = users_->getUserWithID(newUser->id());
        if (!existingUser)
        {
            // The user is new. We indicate to QML that it is managed by the C++ backend.
            QQmlEngine::setObjectOwnership(user.get(), QQmlEngine::CppOwnership);
            continue;
        }

        // The user is already listed. QML code may have a pointer because of an ongoing process (for instance in the SetupGuide),
        // As a consequence we do not want to replace this existing user, but we want to update it.
        existingUser->update(*newUser);
        newUsers[i] = existingUser;
    }

    users_->reset(newUsers);
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
    app().grpc().isPortFree(port, isFree);
    return isFree;
}


//****************************************************************************************************************************************************
/// \return true the native local file path of the given URL.
//****************************************************************************************************************************************************
QString QMLBackend::nativePath(QUrl const &url)
{
    return QDir::toNativeSeparators(url.toLocalFile());
}


//****************************************************************************************************************************************************
/// \return true iff the two URL point to the same local file or folder.
//****************************************************************************************************************************************************
bool QMLBackend::areSameFileOrFolder(QUrl const &lhs, QUrl const &rhs)
{
    return QFileInfo(lhs.toLocalFile()) == QFileInfo(rhs.toLocalFile());
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::guiReady()
{
    app().grpc().guiReady();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::quit()
{
    app().grpc().quit();
    qApp->exit(0);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::restart()
{
    app().grpc().restart();
}


//****************************************************************************************************************************************************
/// \param[in] launcher The path to the launcher.
//****************************************************************************************************************************************************
void QMLBackend::forceLauncher(QString launcher)
{
    app().grpc().forceLauncher(launcher);
}


//****************************************************************************************************************************************************
/// \param[in] active Should we activate autostart.
//****************************************************************************************************************************************************
void QMLBackend::toggleAutostart(bool active)
{
    app().grpc().setIsAutostartOn(active);
    emit isAutostartOnChanged(this->isAutostartOn());
}


//****************************************************************************************************************************************************
/// \param[in] active The new state for the beta enabled property.
//****************************************************************************************************************************************************
void QMLBackend::toggleBeta(bool active)
{
    app().grpc().setIsBetaEnabled(active);
    emit isBetaEnabledChanged(this->isBetaEnabled());
}

//****************************************************************************************************************************************************
/// \param[in] active The new state for the All Mail visibility property.
//****************************************************************************************************************************************************
void QMLBackend::changeIsAllMailVisible(bool isVisible)
{
    app().grpc().setIsAllMailVisible(isVisible);
    emit isAllMailVisibleChanged(this->isAllMailVisible());
}


//****************************************************************************************************************************************************
/// \param[in] scheme the scheme name
//****************************************************************************************************************************************************
void QMLBackend::changeColorScheme(QString const &scheme)
{
    app().grpc().setColorSchemeName(scheme);
    emit colorSchemeNameChanged(this->colorSchemeName());
}

//****************************************************************************************************************************************************
/// \param[in] path The path of the disk cache.
//****************************************************************************************************************************************************
void QMLBackend::setDiskCachePath(QUrl const &path) const
{
    app().grpc().setDiskCachePath(path);
}


//****************************************************************************************************************************************************
/// \return The IMAP port.
//****************************************************************************************************************************************************
int QMLBackend::imapPort() const
{
    return imapPort_;
}


//****************************************************************************************************************************************************
/// \param[in] port The IMAP port.
//****************************************************************************************************************************************************
void QMLBackend::setIMAPPort(int port)
{
    if (port == imapPort_)
        return;
    imapPort_ = port;
    emit imapPortChanged(port);
}


//****************************************************************************************************************************************************
/// \return The SMTP port.
//****************************************************************************************************************************************************
int QMLBackend::smtpPort() const
{
    return smtpPort_;
}


//****************************************************************************************************************************************************
/// \param[in] port The SMTP port.
//****************************************************************************************************************************************************
void QMLBackend::setSMTPPort(int port)
{
    if (port == smtpPort_)
        return;
    smtpPort_ = port;
    emit smtpPortChanged(port);
}


//****************************************************************************************************************************************************
/// \return The value for the 'Use SSL for IMAP' property.
//****************************************************************************************************************************************************
bool QMLBackend::useSSLForIMAP() const
{
    return useSSLForIMAP_;
}


//****************************************************************************************************************************************************
/// \param[in] value The value for the 'Use SSL for IMAP' property.
//****************************************************************************************************************************************************
void QMLBackend::setUseSSLForIMAP(bool value)
{
    if (value == useSSLForIMAP_)
        return;
    useSSLForIMAP_ = value;
    emit useSSLForIMAPChanged(value);
}


//****************************************************************************************************************************************************
/// \return The value for the 'Use SSL for SMTP' property.
//****************************************************************************************************************************************************
bool QMLBackend::useSSLForSMTP() const
{
    return useSSLForSMTP_;
}


//****************************************************************************************************************************************************
/// \param[in] value The value for the 'Use SSL for SMTP' property.
//****************************************************************************************************************************************************
void QMLBackend::setUseSSLForSMTP(bool value)
{
    if (value == useSSLForSMTP_)
        return;
    useSSLForSMTP_ = value;
    emit useSSLForSMTPChanged(value);
}


//****************************************************************************************************************************************************
/// \param[in] imapPort The IMAP port.
/// \param[in] smtpPort The SMTP port.
/// \param[in] useSSLForIMAP The value for the 'Use SSL for IMAP' property
/// \param[in] useSSLForSMTP The value for the 'Use SSL for SMTP' property
//****************************************************************************************************************************************************
void QMLBackend::setMailServerSettings(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP)
{
    app().grpc().setMailServerSettings(imapPort, smtpPort, useSSLForIMAP, useSSLForSMTP);
}


//****************************************************************************************************************************************************
/// \param[in] userID the userID.
/// \param[in] wasSignedOut Was the user signed-out.
//****************************************************************************************************************************************************
void QMLBackend::onLoginFinished(QString const &userID, bool wasSignedOut)
{
    this->retrieveUserList();
    qint32 const index = users_->rowOfUserID(userID);
    emit loginFinished(index, wasSignedOut);
}


//****************************************************************************************************************************************************
/// \param[in] userID the userID.
//****************************************************************************************************************************************************
void QMLBackend::onLoginAlreadyLoggedIn(QString const &userID)
{
    this->retrieveUserList();
    qint32 const index = users_->rowOfUserID(userID);
    emit loginAlreadyLoggedIn(index);
}

//****************************************************************************************************************************************************
/// \param[in] useSSLForIMAP The value for the 'Use SSL for IMAP' property
/// \param[in] useSSLForSMTP The value for the 'Use SSL for SMTP' property
//****************************************************************************************************************************************************
void QMLBackend::onMailServerSettingsChanged(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP)
{
    this->setIMAPPort(imapPort);
    this->setSMTPPort(smtpPort);
    this->setUseSSLForIMAP(useSSLForIMAP);
    this->setUseSSLForSMTP(useSSLForSMTP);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::onGenericError(ErrorInfo const &info)
{
    emit genericError(info.title, info.description);
}


//****************************************************************************************************************************************************
/// \param[in] active Should DoH be active.
//****************************************************************************************************************************************************
void QMLBackend::toggleDoH(bool active)
{
    if (app().grpc().setIsDoHEnabled(active).ok())
        emit isDoHEnabledChanged(active);
}


//****************************************************************************************************************************************************
/// \param[in] keychain The new keychain.
//****************************************************************************************************************************************************
void QMLBackend::changeKeychain(QString const &keychain)
{
    if (app().grpc().setCurrentKeychain(keychain).ok())
        emit currentKeychainChanged(keychain);
}


//****************************************************************************************************************************************************
/// \param[in] active Should automatic update be turned on.
//****************************************************************************************************************************************************
void QMLBackend::toggleAutomaticUpdate(bool active)
{
    if (app().grpc().setIsAutomaticUpdateOn(active).ok())
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
    app().grpc().installUpdate();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::triggerReset()
{
    app().grpc().triggerReset();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::onResetFinished()
{
    emit resetFinished();
    this->restart();
}


//****************************************************************************************************************************************************
// onVersionChanged update dynamic link related to version
//****************************************************************************************************************************************************
void QMLBackend::onVersionChanged()
{
    emit releaseNotesLinkChanged(releaseNotesLink());
    emit landingPageLinkChanged(landingPageLink());
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void QMLBackend::exportTLSCertificates()
{
    QString const folderPath = QFileDialog::getExistingDirectory(nullptr, QObject::tr("Select directory"),
        QStandardPaths::writableLocation(QStandardPaths::HomeLocation));
    if (!folderPath.isEmpty())
        app().grpc().exportTLSCertificates(folderPath);
}
