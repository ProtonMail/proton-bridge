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
    void init(); ///< Initialize the backend.
    bool waitForEventStreamReaderToFinish(qint32 timeoutMs); ///< Wait for the event stream reader to finish.

    // invokable methods can be called from QML. They generally return a value, which slots cannot do.
    Q_INVOKABLE static QPoint getCursorPos();                                                                                                         //   _ func() *core.QPoint             `slot:"getCursorPos"`
    Q_INVOKABLE static bool isPortFree(int port);                                                                                                     //   _ func(port int) bool             `slot:"isPortFree"`

public: // Qt/QML properties. Note that the NOTIFY-er signal is required even for read-only properties (QML warning otherwise)
    Q_PROPERTY(bool showOnStartup READ showOnStartup NOTIFY showOnStartupChanged)                                                                     //    _ bool        `property:showOnStartup`
    Q_PROPERTY(bool showSplashScreen READ showSplashScreen WRITE setShowSplashScreen NOTIFY showSplashScreenChanged)                                  //    _ bool        `property:showSplashScreen`
    Q_PROPERTY(QString goos READ goos NOTIFY goosChanged)                                                                                             //    _ string      `property:"goos"`
    Q_PROPERTY(QUrl logsPath READ logsPath NOTIFY logsPathChanged)                                                                                    //    _ core.QUrl   `property:"logsPath"`
    Q_PROPERTY(QUrl licensePath READ licensePath NOTIFY licensePathChanged)                                                                           //    _ core.QUrl   `property:"licensePath"`
    Q_PROPERTY(QUrl releaseNotesLink READ releaseNotesLink NOTIFY releaseNotesLinkChanged)                                  //    _ core.QUrl   `property:"releaseNotesLink"`
    Q_PROPERTY(QUrl dependencyLicensesLink READ dependencyLicensesLink NOTIFY dependencyLicensesLinkChanged)                                          //    _ core.QUrl   `property:"dependencyLicensesLink"`
    Q_PROPERTY(QUrl landingPageLink READ landingPageLink NOTIFY landingPageLinkChanged)                                      //    _ core.QUrl   `property:"landingPageLink"`
    Q_PROPERTY(QString version READ version NOTIFY versionChanged)                                                                                    //    _ string      `property:"version"`
    Q_PROPERTY(QString hostname READ hostname NOTIFY hostnameChanged)                                                                                 //    _ string      `property:"hostname"`
    Q_PROPERTY(bool isAutostartOn READ isAutostartOn NOTIFY isAutostartOnChanged)                                                                     //    _ bool        `property:"isAutostartOn"`
    Q_PROPERTY(bool isBetaEnabled READ isBetaEnabled NOTIFY isBetaEnabledChanged)                                                                     //    _ bool        `property:"isBetaEnabled"`
    Q_PROPERTY(QString colorSchemeName READ colorSchemeName NOTIFY colorSchemeNameChanged)                                                            //    _ string      `property:"colorSchemeName"`
    Q_PROPERTY(bool isDiskCacheEnabled READ isDiskCacheEnabled NOTIFY isDiskCacheEnabledChanged)                                                      //    _ bool        `property:"isDiskCacheEnabled"`
    Q_PROPERTY(QUrl diskCachePath READ diskCachePath NOTIFY diskCachePathChanged)                                                                     //    _ core.QUrl   `property:"diskCachePath"`
    Q_PROPERTY(bool useSSLforSMTP READ useSSLForSMTP NOTIFY useSSLforSMTPChanged)                                                                     //    _ bool        `property:"useSSLforSMTP"`
    Q_PROPERTY(int portIMAP READ portIMAP NOTIFY portIMAPChanged)                                                                                     //    _ int         `property:"portIMAP"`
    Q_PROPERTY(int portSMTP READ portSMTP NOTIFY portSMTPChanged)                                                                                     //    _ int         `property:"portSMTP"`
    Q_PROPERTY(bool isDoHEnabled READ isDoHEnabled NOTIFY isDoHEnabledChanged)                                                                        //    _ bool        `property:"isDoHEnabled"`
    Q_PROPERTY(bool isFirstGUIStart READ isFirstGUIStart)                                                                                             //    _ bool        `property:"isFirstGUIStart"`
    Q_PROPERTY(bool isAutomaticUpdateOn READ isAutomaticUpdateOn NOTIFY isAutomaticUpdateOnChanged)                                                   //    _ bool        `property:"isAutomaticUpdateOn"`
    Q_PROPERTY(QString currentEmailClient READ currentEmailClient NOTIFY currentEmailClientChanged)                                                   //    _ string      `property:"currentEmailClient"`
    Q_PROPERTY(QStringList availableKeychain READ availableKeychain NOTIFY availableKeychainChanged)                                                  //    _ []string    `property:"availableKeychain"`
    Q_PROPERTY(QString currentKeychain READ currentKeychain NOTIFY currentKeychainChanged)                                                            //    _ string      `property:"currentKeychain"`
    Q_PROPERTY(UserList* users MEMBER users_ NOTIFY usersChanged)
    Q_PROPERTY(bool dockIconVisible READ dockIconVisible WRITE setDockIconVisible NOTIFY dockIconVisibleChanged)                                      //    _ bool        `property:dockIconVisible`

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
    QString version() const { QString version; app().grpc().version(version); return version; }
    QString hostname() const { QString hostname; app().grpc().hostname(hostname); return hostname; }
    bool isAutostartOn() const { bool v; app().grpc().isAutostartOn(v); return v; };
    bool isBetaEnabled() const { bool v; app().grpc().isBetaEnabled(v); return v; }
    QString colorSchemeName() const { QString name; app().grpc().colorSchemeName(name); return name; }
    bool isDiskCacheEnabled() const { bool enabled; app().grpc().isCacheOnDiskEnabled(enabled); return enabled;}
    QUrl diskCachePath() const { QUrl path; app().grpc().diskCachePath(path); return path; }
    bool useSSLForSMTP() const{ bool useSSL; app().grpc().useSSLForSMTP(useSSL); return useSSL; }
    int portIMAP() const { int port; app().grpc().portIMAP(port); return port; }
    int portSMTP() const { int port; app().grpc().portSMTP(port); return port; }
    bool isDoHEnabled() const { bool isEnabled; app().grpc().isDoHEnabled(isEnabled); return isEnabled;}
    bool isFirstGUIStart() const { bool v; app().grpc().isFirstGUIStart(v); return v; };
    bool isAutomaticUpdateOn() const { bool isOn = false; app().grpc().isAutomaticUpdateOn(isOn); return isOn; }
    QString currentEmailClient() { QString client; app().grpc().currentEmailClient(client); return client;}
    QStringList availableKeychain() const { QStringList keychains; app().grpc().availableKeychains(keychains); return keychains; }
    QString currentKeychain() const { QString keychain; app().grpc().currentKeychain(keychain); return keychain; }
    bool dockIconVisible() const { return getDockIconVisibleState(); };
    void setDockIconVisible(bool visible) { setDockIconVisibleState(visible); emit dockIconVisibleChanged(visible); }

signals: // Signal used by the Qt property system. Many of them are unused but required to avoir warning from the QML engine.
    void showSplashScreenChanged(bool value);
    void showOnStartupChanged(bool value);
    void goosChanged(QString const &value);
    void isDiskCacheEnabledChanged(bool value);
    void diskCachePathChanged(QUrl const &url);
    void useSSLforSMTPChanged(bool value);
    void isAutomaticUpdateOnChanged(bool value);
    void isBetaEnabledChanged(bool value);
    void colorSchemeNameChanged(QString const &scheme);
    void isDoHEnabledChanged(bool value);
    void logsPathChanged(QUrl const &path);
    void licensePathChanged(QUrl const &path);
    void releaseNotesLinkChanged(QUrl const &link);
    void dependencyLicensesLinkChanged(QUrl const &link);
    void landingPageLinkChanged(QUrl const &link);
    void versionChanged(QString const &version);
    void currentEmailClientChanged(QString const &email);
    void currentKeychainChanged(QString const &keychain);
    void availableKeychainChanged(QStringList const &keychains);
    void hostnameChanged(QString const &hostname);
    void isAutostartOnChanged(bool value);
    void portIMAPChanged(int port);
    void portSMTPChanged(int port);
    void usersChanged(UserList* users);
    void dockIconVisibleChanged(bool value);

public slots: // slot for signals received from QML -> To be forwarded to Bridge via RPC Client calls.
    void toggleAutostart(bool active);                                                                                                                //    _ func(makeItActive bool)                                             `slot:"toggleAutostart"`
    void toggleBeta(bool active);                                                                                                                     //    _ func(makeItActive bool)                                             `slot:"toggleBeta"`
    void changeColorScheme(QString const &scheme);                                                                                                    //    _ func(string)                                                        `slot:"changeColorScheme"`
    void changeLocalCache(bool enable, QUrl const& path) { app().grpc().changeLocalCache(enable, path); }      //    _ func(enableDiskCache bool, diskCachePath core.QUrl)                 `slot:"changeLocalCache"`
    void login(QString const& username, QString const& password) { app().grpc().login(username, password);}               //    _ func(username, password string)                                     `slot:"login"`
    void login2FA(QString const& username, QString const& code) { app().grpc().login2FA(username, code);}              //    _ func(username, code string)                                         `slot:"login2FA"`
    void login2Password(QString const& username, QString const& password) { app().grpc().login2Passwords(username, password);}                                                                                                                          //    _ func(username, password string)                                     `slot:"login2Password"`
    void loginAbort(QString const& username){ app().grpc().loginAbort(username);}                                    //    _ func(username string)                                               `slot:"loginAbort"`
    void toggleUseSSLforSMTP(bool makeItActive);                                                                                                      //    _ func(makeItActive bool)                                             `slot:"toggleUseSSLforSMTP"`
    void changePorts(int imapPort, int smtpPort);                                                                                                     //    _ func(imapPort, smtpPort int)                                        `slot:"changePorts"`
    void toggleDoH(bool active);                                                                                                                      //    _ func(makeItActive bool)                                             `slot:"toggleDoH"`
    void toggleAutomaticUpdate(bool makeItActive);                                                                                                    //    _ func(makeItActive bool)                                             `slot:"toggleAutomaticUpdate"`
    void updateCurrentMailClient() { emit currentEmailClientChanged(currentEmailClient()); }                                                          //    _ func() `slot:"updateCurrentMailClient"`
    void changeKeychain(QString const &keychain);                                                                                                     //    _ func(keychain string)                                               `slot:"changeKeychain"`
    void guiReady();                                                                                                                                  //    _ func()                                                              `slot:"guiReady"`
    void quit();                                                                                                                                      //    _ func()                                                              `slot:"quit"`
    void restart();                                                                                                                                   //    _ func()                                                              `slot:"restart"`
    void forceLauncher(QString launcher);                                                                                                             //    _ func()                                                              `slot:"forceLauncher"`
    void checkUpdates();                                                                                                                              //    _ func()                                                              `slot:"checkUpdates"`
    void installUpdate();                                                                                                                             //    _ func()                                                              `slot:"installUpdate"`
    void triggerReset();                                                                                                                              //    _ func()                                                              `slot:"triggerReset"`
    void reportBug(QString const &description, QString const& address, QString const &emailClient, bool includeLogs) {
        app().grpc().reportBug(description, address, emailClient, includeLogs); }                                                  //    _ func(description, address, emailClient string, includeLogs bool)    `slot:"reportBug"`
    void onResetFinished();

signals: // Signals received from the Go backend, to be forwarded to QML
    void toggleAutostartFinished();                                                                                                                   //    _ func()                  `signal:"toggleAutostartFinished"`
    void cacheUnavailable();                                                                                                                          //    _ func()                  `signal:"cacheUnavailable"`
    void cacheCantMove();                                                                                                                             //    _ func()                  `signal:"cacheCantMove"`
    void cacheLocationChangeSuccess();                                                                                                                //    _ func()                  `signal:"cacheLocationChangeSuccess"`
    void diskFull();                                                                                                                                  //    _ func()                  `signal:"diskFull"`
    void changeLocalCacheFinished();                                                                                                                  //    _ func()                  `signal:"changeLocalCacheFinished"`
    void loginUsernamePasswordError(QString const &errorMsg);                                                                                         //    _ func(errorMsg string)   `signal:"loginUsernamePasswordError"`
    void loginFreeUserError();                                                                                                                        //    _ func()                  `signal:"loginFreeUserError"`
    void loginConnectionError(QString const &errorMsg);                                                                                               //    _ func(errorMsg string)   `signal:"loginConnectionError"`
    void login2FARequested(QString const &username);                                                                                                  //    _ func(username string)   `signal:"login2FARequested"`
    void login2FAError(QString const& errorMsg);                                                                                                      //    _ func(errorMsg string)   `signal:"login2FAError"`
    void login2FAErrorAbort(QString const& errorMsg);                                                                                                 //    _ func(errorMsg string)   `signal:"login2FAErrorAbort"`
    void login2PasswordRequested();                                                                                                                   //    _ func()                  `signal:"login2PasswordRequested"`
    void login2PasswordError(QString const& errorMsg);                                                                                                //    _ func(errorMsg string)   `signal:"login2PasswordError"`
    void login2PasswordErrorAbort(QString const& errorMsg);                                                                                           //    _ func(errorMsg string)   `signal:"login2PasswordErrorAbort"`
    void loginFinished(int index);                                                                                                                    //    _ func(index int)         `signal:"loginFinished"`
    void loginAlreadyLoggedIn(int index);                                                                                                             //    _ func(index int)         `signal:"loginAlreadyLoggedIn"`
    void updateManualReady(QString const& version);                                                                                                   //    _ func(version string)    `signal:"updateManualReady"`
    void updateManualRestartNeeded();                                                                                                                 //    _ func()                  `signal:"updateManualRestartNeeded"`
    void updateManualError();                                                                                                                         //    _ func()                  `signal:"updateManualError"`
    void updateForce(QString const& version);                                                                                                         //    _ func(version string)    `signal:"updateForce"`
    void updateForceError();                                                                                                                          //    _ func()                  `signal:"updateForceError"`
    void updateSilentRestartNeeded();                                                                                                                 //    _ func()                  `signal:"updateSilentRestartNeeded"`
    void updateSilentError();                                                                                                                         //    _ func()                  `signal:"updateSilentError"`
    void updateIsLatestVersion();                                                                                                                     //    _ func()                  `signal:"updateIsLatestVersion"`
    void checkUpdatesFinished();                                                                                                                      //    _ func()                  `signal:"checkUpdatesFinished"`
    void toggleUseSSLFinished();                                                                                                                      //    _ func()                  `signal:"toggleUseSSLFinished"`
    void changePortFinished();                                                                                                                        //    _ func()                  `signal:"changePortFinished"`
    void portIssueIMAP();                                                                                                                             //    _ func()                  `signal:"portIssueIMAP"`
    void portIssueSMTP();                                                                                                                             //    _ func()                  `signal:"portIssueSMTP"`
    void changeKeychainFinished();                                                                                                                    //    _ func()                  `signal:"changeKeychainFinished"`
    void notifyHasNoKeychain();                                                                                                                       //    _ func()                  `signal:"notifyHasNoKeychain"`
    void notifyRebuildKeychain();                                                                                                                     //    _ func()                  `signal:"notifyRebuildKeychain"`
    void noActiveKeyForRecipient(QString const& email);                                                                                               //    _ func(email string)      `signal:noActiveKeyForRecipient`
    void addressChanged(QString const& address);                                                                                                      //    _ func(address string)    `signal:addressChanged`
    void addressChangedLogout(QString const& address);                                                                                                //    _ func(address string)    `signal:addressChangedLogout`
    void apiCertIssue();                                                                                                                              //    _ func()                  `signal:apiCertIssue`
    void userDisconnected(QString const& username);                                                                                                   //    _ func(username string)   `signal:userDisconnected`
    void internetOff();                                                                                                                               //    _ func()                  `signal:"internetOff"`
    void internetOn();                                                                                                                                //    _ func()                  `signal:"internetOn"`
    void resetFinished();                                                                                                                             //    _ func()                  `signal:"resetFinished"`
    void reportBugFinished();                                                                                                                         //    _ func()                  `signal:"reportBugFinished"`
    void bugReportSendSuccess();                                                                                                                      //    _ func()                  `signal:"bugReportSendSuccess"`
    void bugReportSendError();                                                                                                                        //    _ func()                  `signal:"bugReportSendError"`
    void showMainWindow();                                                                                                                            //    _ func()                  `signal:showMainWindow`

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

    friend class AppController;
};


#endif // BRIDGE_GUI_QML_BACKEND_H
