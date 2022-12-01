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


#ifndef BRIDGE_GUI_QML_BACKEND_H
#define BRIDGE_GUI_QML_BACKEND_H


#include "DockIcon/DockIcon.h"
#include "Version.h"
#include "UserList.h"
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/GRPC/GRPCUtils.h>
#include <bridgepp/Worker/Overseer.h>


//****************************************************************************************************************************************************
/// \brief Bridge C++ backend class.
//****************************************************************************************************************************************************
class QMLBackend: public QObject
{
Q_OBJECT

public: // member functions.
    QMLBackend(); ///< Default constructor.
    QMLBackend(QMLBackend const &) = delete; ///< Disabled copy-constructor.
    QMLBackend(QMLBackend &&) = delete; ///< Disabled assignment copy-constructor.
    ~QMLBackend() override = default; ///< Destructor.
    QMLBackend &operator=(QMLBackend const &) = delete; ///< Disabled assignment operator.
    QMLBackend &operator=(QMLBackend &&) = delete; ///< Disabled move assignment operator.
    void init(GRPCConfig const &serviceConfig); ///< Initialize the backend.
    bool waitForEventStreamReaderToFinish(qint32 timeoutMs); ///< Wait for the event stream reader to finish.

    // invokable methods can be called from QML. They generally return a value, which slots cannot do.
    Q_INVOKABLE static QPoint getCursorPos();
    Q_INVOKABLE static bool isPortFree(int port);
    Q_INVOKABLE static QString nativePath(QUrl const &url);
    Q_INVOKABLE static bool areSameFileOrFolder(QUrl const &lhs, QUrl const &rhs);

public: // Qt/QML properties. Note that the NOTIFY-er signal is required even for read-only properties (QML warning otherwise)
    Q_PROPERTY(bool showOnStartup READ showOnStartup NOTIFY showOnStartupChanged)
    Q_PROPERTY(bool showSplashScreen READ showSplashScreen WRITE setShowSplashScreen NOTIFY showSplashScreenChanged)
    Q_PROPERTY(QString goos READ goos NOTIFY goosChanged)
    Q_PROPERTY(QUrl logsPath READ logsPath NOTIFY logsPathChanged)
    Q_PROPERTY(QUrl licensePath READ licensePath NOTIFY licensePathChanged)
    Q_PROPERTY(QUrl releaseNotesLink READ releaseNotesLink NOTIFY releaseNotesLinkChanged)
    Q_PROPERTY(QUrl dependencyLicensesLink READ dependencyLicensesLink NOTIFY dependencyLicensesLinkChanged)
    Q_PROPERTY(QUrl landingPageLink READ landingPageLink NOTIFY landingPageLinkChanged)
    Q_PROPERTY(QString appname READ appname NOTIFY appnameChanged)
    Q_PROPERTY(QString vendor READ vendor NOTIFY vendorChanged)
    Q_PROPERTY(QString version READ version NOTIFY versionChanged)
    Q_PROPERTY(QString hostname READ hostname NOTIFY hostnameChanged)
    Q_PROPERTY(bool isAutostartOn READ isAutostartOn NOTIFY isAutostartOnChanged)
    Q_PROPERTY(bool isBetaEnabled READ isBetaEnabled NOTIFY isBetaEnabledChanged)
    Q_PROPERTY(bool isAllMailVisible READ isAllMailVisible NOTIFY isAllMailVisibleChanged)
    Q_PROPERTY(QString colorSchemeName READ colorSchemeName NOTIFY colorSchemeNameChanged)
    Q_PROPERTY(QUrl diskCachePath READ diskCachePath NOTIFY diskCachePathChanged)
    Q_PROPERTY(bool useSSLForIMAP READ useSSLForIMAP WRITE setUseSSLForIMAP NOTIFY useSSLForIMAPChanged)
    Q_PROPERTY(bool useSSLForSMTP READ useSSLForSMTP WRITE setUseSSLForSMTP NOTIFY useSSLForSMTPChanged)
    Q_PROPERTY(int imapPort READ imapPort WRITE setIMAPPort NOTIFY imapPortChanged)
    Q_PROPERTY(int smtpPort READ smtpPort WRITE setSMTPPort NOTIFY smtpPortChanged)
    Q_PROPERTY(bool isDoHEnabled READ isDoHEnabled NOTIFY isDoHEnabledChanged)
    Q_PROPERTY(bool isFirstGUIStart READ isFirstGUIStart)
    Q_PROPERTY(bool isAutomaticUpdateOn READ isAutomaticUpdateOn NOTIFY isAutomaticUpdateOnChanged)
    Q_PROPERTY(QString currentEmailClient READ currentEmailClient NOTIFY currentEmailClientChanged)
    Q_PROPERTY(QStringList availableKeychain READ availableKeychain NOTIFY availableKeychainChanged)
    Q_PROPERTY(QString currentKeychain READ currentKeychain NOTIFY currentKeychainChanged)
    Q_PROPERTY(UserList* users MEMBER users_ NOTIFY usersChanged)
    Q_PROPERTY(bool dockIconVisible READ dockIconVisible WRITE setDockIconVisible NOTIFY dockIconVisibleChanged)

    // Qt Property system setters & getters.
    bool showOnStartup() const { bool v = false; app().grpc().showOnStartup(v); return v; };
    bool showSplashScreen() const { return showSplashScreen_; };
    void setShowSplashScreen(bool show) { if (show != showSplashScreen_) { showSplashScreen_ = show; emit showSplashScreenChanged(show); } }
    QString goos() { return goos_; }
    QUrl logsPath() const { return logsPath_; }
    QUrl licensePath() const { return licensePath_; }
    QUrl releaseNotesLink() const { QUrl link; app().grpc().releaseNotesPageLink(link); return link;   }
    QUrl dependencyLicensesLink() const { QUrl link; app().grpc().dependencyLicensesLink(link); return link; }
    QUrl landingPageLink() const { QUrl link; app().grpc().landingPageLink(link); return link;  }
    QString appname() const { return QString(PROJECT_FULL_NAME); }
    QString vendor() const { return QString(PROJECT_VENDOR); }
    QString version() const { QString version; app().grpc().version(version); return version; }
    QString hostname() const { QString hostname; app().grpc().hostname(hostname); return hostname; }
    bool isAutostartOn() const { bool v; app().grpc().isAutostartOn(v); return v; };
    bool isBetaEnabled() const { bool v; app().grpc().isBetaEnabled(v); return v; }
    bool isAllMailVisible() const { bool v; app().grpc().isAllMailVisible(v); return v; }
    QString colorSchemeName() const { QString name; app().grpc().colorSchemeName(name); return name; }
    QUrl diskCachePath() const { QUrl path; app().grpc().diskCachePath(path); return path; }
    bool useSSLForIMAP() const;
    void setUseSSLForIMAP(bool value);
    bool useSSLForSMTP() const;
    void setUseSSLForSMTP(bool value);
    int imapPort() const;
    void setIMAPPort(int port);
    int smtpPort() const;
    void setSMTPPort(int port);
    bool isDoHEnabled() const { bool isEnabled; app().grpc().isDoHEnabled(isEnabled); return isEnabled;}
    bool isFirstGUIStart() const { bool v; app().grpc().isFirstGUIStart(v); return v; };
    bool isAutomaticUpdateOn() const { bool isOn = false; app().grpc().isAutomaticUpdateOn(isOn); return isOn; }
    QString currentEmailClient() { QString client; app().grpc().currentEmailClient(client); return client;}
    QStringList availableKeychain() const { QStringList keychains; app().grpc().availableKeychains(keychains); return keychains; }
    QString currentKeychain() const { QString keychain; app().grpc().currentKeychain(keychain); return keychain; }
    bool dockIconVisible() const { return getDockIconVisibleState(); };
    void setDockIconVisible(bool visible) { setDockIconVisibleState(visible); emit dockIconVisibleChanged(visible); }

signals: // Signal used by the Qt property system. Many of them are unused but required to avoid warning from the QML engine.
    void showSplashScreenChanged(bool value);
    void showOnStartupChanged(bool value);
    void goosChanged(QString const &value);
    void diskCachePathChanged(QUrl const &url);
    void imapPortChanged(int port);
    void smtpPortChanged(int port);
    void useSSLForSMTPChanged(bool value);
    void useSSLForIMAPChanged(bool value);
    void isAutomaticUpdateOnChanged(bool value);
    void isBetaEnabledChanged(bool value);
    void isAllMailVisibleChanged(bool value);
    void colorSchemeNameChanged(QString const &scheme);
    void isDoHEnabledChanged(bool value);
    void logsPathChanged(QUrl const &path);
    void licensePathChanged(QUrl const &path);
    void releaseNotesLinkChanged(QUrl const &link);
    void dependencyLicensesLinkChanged(QUrl const &link);
    void landingPageLinkChanged(QUrl const &link);
    void appnameChanged(QString const &appname);
    void vendorChanged(QString const &vendor);
    void versionChanged(QString const &version);
    void currentEmailClientChanged(QString const &email);
    void currentKeychainChanged(QString const &keychain);
    void availableKeychainChanged(QStringList const &keychains);
    void hostnameChanged(QString const &hostname);
    void isAutostartOnChanged(bool value);
    void usersChanged(UserList* users);
    void dockIconVisibleChanged(bool value);

public slots: // slot for signals received from QML -> To be forwarded to Bridge via RPC Client calls.
    void toggleAutostart(bool active);
    void toggleBeta(bool active);
    void changeIsAllMailVisible(bool isVisible);
    void changeColorScheme(QString const &scheme);
    void setDiskCachePath(QUrl const& path) const;
    void login(QString const& username, QString const& password) { app().grpc().login(username, password);}
    void login2FA(QString const& username, QString const& code) { app().grpc().login2FA(username, code);}
    void login2Password(QString const& username, QString const& password) { app().grpc().login2Passwords(username, password);}
    void loginAbort(QString const& username){ app().grpc().loginAbort(username);}
    void toggleDoH(bool active);
    void toggleAutomaticUpdate(bool makeItActive);
    void updateCurrentMailClient() { emit currentEmailClientChanged(currentEmailClient()); }
    void changeKeychain(QString const &keychain);
    void guiReady();
    void quit();
    void restart();
    void forceLauncher(QString launcher);
    void checkUpdates();
    void installUpdate();
    void triggerReset();
    void reportBug(QString const &description, QString const& address, QString const &emailClient, bool includeLogs) {
        app().grpc().reportBug(description, address, emailClient, includeLogs); }
    void exportTLSCertificates();
    void onResetFinished();
    void onVersionChanged();
    void setMailServerSettings(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Forwards a connection mode change request from QML to gRPC

public slots: // slot for signals received from gRPC that need transformation instead of simple forwarding
    void onMailServerSettingsChanged(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Slot for the ConnectionModeChanged gRPC event.
    void onGenericError(bridgepp::ErrorInfo const& info); ///< Slot for generic errors received from the gRPC service.

signals: // Signals received from the Go backend, to be forwarded to QML
    void toggleAutostartFinished();
    void diskCacheUnavailable();
    void cantMoveDiskCache();
    void diskCachePathChangeFinished();
    void diskFull();
    void loginUsernamePasswordError(QString const &errorMsg);
    void loginFreeUserError();
    void loginConnectionError(QString const &errorMsg);
    void login2FARequested(QString const &username);
    void login2FAError(QString const& errorMsg);
    void login2FAErrorAbort(QString const& errorMsg);
    void login2PasswordRequested();
    void login2PasswordError(QString const& errorMsg);
    void login2PasswordErrorAbort(QString const& errorMsg);
    void loginFinished(int index);
    void loginAlreadyLoggedIn(int index);
    void updateManualReady(QString const& version);
    void updateManualRestartNeeded();
    void updateManualError();
    void updateForce(QString const& version);
    void updateForceError();
    void updateSilentRestartNeeded();
    void updateSilentError();
    void updateIsLatestVersion();
    void checkUpdatesFinished();
    void imapPortStartupError();
    void smtpPortStartupError();
    void imapPortChangeError();
    void smtpPortChangeError();
    void imapConnectionModeChangeError();
    void smtpConnectionModeChangeError();
    void changeMailServerSettingsFinished();
    void changeKeychainFinished();
    void notifyHasNoKeychain();
    void notifyRebuildKeychain();
    void noActiveKeyForRecipient(QString const& email);
    void addressChanged(QString const& address);
    void addressChangedLogout(QString const& address);
    void apiCertIssue();
    void userDisconnected(QString const& username);
    void internetOff();
    void internetOn();
    void resetFinished();
    void reportBugFinished();
    void bugReportSendSuccess();
    void bugReportSendError();
    void showMainWindow();
    void hideMainWindow();
    void genericError(QString const& title, QString const& description);

private: // member functions
    void retrieveUserList(); ///< Retrieve the list of users via gRPC.
    void connectGrpcEvents(); ///< Connect gRPC that need to be forwarded to QML via backend signals

private: // data members
    UserList* users_ { nullptr }; ///< The user list. Owned by backend.
    std::unique_ptr<bridgepp::Overseer> eventStreamOverseer_; ///< The event stream overseer.
    bool showSplashScreen_ { false }; ///< The cached version of show splash screen. Retrieved on startup from bridge, and potentially modified locally.
    QString goos_; ///< The cached version of the GOOS variable.
    QUrl logsPath_; ///< The logs path. Retrieved from bridge on startup.
    QUrl licensePath_; ///< The license path. Retrieved from bridge on startup.
    int imapPort_ { 0 }; ///< The cached value for the IMAP port.
    int smtpPort_ { 0 }; ///< The cached value for the SMTP port.
    bool useSSLForIMAP_ { false }; ///< The cached value for useSSLForIMAP.
    bool useSSLForSMTP_ { false }; ///< The cached value for useSSLForSMTP.

    friend class AppController;
};


#endif // BRIDGE_GUI_QML_BACKEND_H
