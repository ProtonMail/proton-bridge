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


#ifndef BRIDGE_GUI_QML_BACKEND_H
#define BRIDGE_GUI_QML_BACKEND_H


#include "MacOS/DockIcon.h"
#include "BuildConfig.h"
#include "TrayIcon.h"
#include "UserList.h"
#include <bridgepp/BugReportFlow/BugReportFlow.h>
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/GRPC/GRPCUtils.h>
#include <bridgepp/Worker/Overseer.h>
#include <stack>


//****************************************************************************************************************************************************
/// \brief Bridge C++ backend class.
//****************************************************************************************************************************************************
class QMLBackend : public QObject {
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
    UserList const& users() const; ///< Return the list of users
    bool isInternetOn() const; ///< Check if bridge considers internet as on.
    void showMainWindow(QString const &reason); ///< Show the main window.
    void showHelp(QString const &reason); ///< Show the help page.
    void showSettings(QString const &reason); ///< Show the settings page.
    void selectUser(QString const &userID, bool forceShowWindow, QString const &reason); ///< Select the user and display its account details (or login screen).

    // invocable methods can be called from QML. They generally return a value, which slots cannot do.
    Q_INVOKABLE static QString buildYear(); ///< Return the application build year.
    Q_INVOKABLE QPoint getCursorPos() const; ///< Retrieve the cursor position.
    Q_INVOKABLE bool isPortFree(int port) const; ///< Check if a given network port is available.
    Q_INVOKABLE QString nativePath(QUrl const &url) const; ///< Retrieve the native path of a local URL.
    Q_INVOKABLE bool areSameFileOrFolder(QUrl const &lhs, QUrl const &rhs) const; ///< Check if two local URL point to the same file.
    Q_INVOKABLE QString getBugCategory(quint8 categoryId) const; ///< Get a Category name.
    Q_INVOKABLE QVariantList getQuestionSet(quint8 categoryId) const; ///< Retrieve the set of question for a given bug category.
    Q_INVOKABLE void setQuestionAnswer(quint8 questionId, QString const &answer); ///< Feed an answer for a given question.
    Q_INVOKABLE QString getQuestionAnswer(quint8 questionId) const; ///< Get the answer for a given question.
    Q_INVOKABLE QString collectAnswers(quint8 categoryId) const; ///< Collect answer for a given set of questions.
    Q_INVOKABLE void clearAnswers(); ///< Clear all collected answers.
    Q_INVOKABLE bool isTLSCertificateInstalled(); ///< Check if the bridge certificate is installed in the OS keychain.
    Q_INVOKABLE void openExternalLink(QString const & url = QString()); ///< Open a knowledge base article.
    Q_INVOKABLE void requestKnowledgeBaseSuggestions(qint8 categoryID) const; ///< Request knowledgebase article suggestions.

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
    Q_PROPERTY(QString tag READ tag NOTIFY tagChanged)
    Q_PROPERTY(QString hostname READ hostname NOTIFY hostnameChanged)
    Q_PROPERTY(bool isAutostartOn READ isAutostartOn NOTIFY isAutostartOnChanged)
    Q_PROPERTY(bool isBetaEnabled READ isBetaEnabled NOTIFY isBetaEnabledChanged)
    Q_PROPERTY(bool isAllMailVisible READ isAllMailVisible NOTIFY isAllMailVisibleChanged)
    Q_PROPERTY(bool isTelemetryDisabled READ isTelemetryDisabled NOTIFY isTelemetryDisabledChanged)
    Q_PROPERTY(QString colorSchemeName READ colorSchemeName NOTIFY colorSchemeNameChanged)
    Q_PROPERTY(QUrl diskCachePath READ diskCachePath NOTIFY diskCachePathChanged)
    Q_PROPERTY(bool useSSLForIMAP READ useSSLForIMAP WRITE setUseSSLForIMAP NOTIFY useSSLForIMAPChanged)
    Q_PROPERTY(bool useSSLForSMTP READ useSSLForSMTP WRITE setUseSSLForSMTP NOTIFY useSSLForSMTPChanged)
    Q_PROPERTY(int imapPort READ imapPort WRITE setIMAPPort NOTIFY imapPortChanged)
    Q_PROPERTY(int smtpPort READ smtpPort WRITE setSMTPPort NOTIFY smtpPortChanged)
    Q_PROPERTY(bool isDoHEnabled READ isDoHEnabled NOTIFY isDoHEnabledChanged)
    Q_PROPERTY(bool isAutomaticUpdateOn READ isAutomaticUpdateOn NOTIFY isAutomaticUpdateOnChanged)
    Q_PROPERTY(QString currentEmailClient READ currentEmailClient NOTIFY currentEmailClientChanged)
    Q_PROPERTY(QStringList availableKeychain READ availableKeychain NOTIFY availableKeychainChanged)
    Q_PROPERTY(QString currentKeychain READ currentKeychain NOTIFY currentKeychainChanged)
    Q_PROPERTY(QVariantList bugCategories READ bugCategories NOTIFY bugCategoriesChanged)
    Q_PROPERTY(QVariantList bugQuestions READ bugQuestions NOTIFY bugQuestionsChanged)
    Q_PROPERTY(UserList *users MEMBER users_ NOTIFY usersChanged)
    Q_PROPERTY(bool dockIconVisible READ dockIconVisible WRITE setDockIconVisible NOTIFY dockIconVisibleChanged)

    // Qt Property system setters & getters.
    bool showOnStartup() const; ///< Getter for the 'showOnStartup' property.
    void setShowSplashScreen(bool show);  ///< Setter for the 'showSplashScreen' property.
    bool showSplashScreen() const; ///< Getter for the 'showSplashScreen' property.
    QString goos() const; ///< Getter for the 'GOOS' property.
    QUrl logsPath() const; ///< Getter for the 'logsPath' property.
    QUrl licensePath() const; ///< Getter for the 'licensePath' property.
    QUrl releaseNotesLink() const;///< Getter for the 'releaseNotesLink' property.
    QUrl dependencyLicensesLink() const; ///< Getter for the 'dependencyLicenseLink' property.
    QUrl landingPageLink() const; ///< Getter for the 'landingPageLink' property.
    QString appname() const; ///< Getter for the 'appname' property.
    QString vendor() const; ///< Getter for the 'vendor' property.
    QString version() const; ///< Getter for the 'version' property.
    QString tag() const; ///< Getter for the 'tag' property.
    QString hostname() const; ///< Getter for the 'hostname' property.
    bool isAutostartOn() const; ///< Getter for the 'isAutostartOn' property.
    bool isBetaEnabled() const; ///< Getter for the 'isBetaEnabled' property.
    bool isAllMailVisible() const; ///< Getter for the 'isAllMailVisible' property.
    bool isTelemetryDisabled() const; ///< Getter for the 'isTelemetryDisabled' property.
    QString colorSchemeName() const; ///< Getter for the 'colorSchemeName' property.
    QUrl diskCachePath() const; ///< Getter for the 'diskCachePath' property.
    void setUseSSLForIMAP(bool value); ///< Setter for the 'useSSLForIMAP' property.
    bool useSSLForIMAP() const; ///< Getter for the 'useSSLForIMAP' property.
    void setUseSSLForSMTP(bool value); ///< Setter for the 'useSSLForSMTP' property.
    bool useSSLForSMTP() const; ///< Getter for the 'useSSLForSMTP' property.
    void setIMAPPort(int port); ///< Setter for the 'imapPort' property.
    int imapPort() const; ///< Getter for the 'imapPort' property.
    void setSMTPPort(int port); ///< Setter for the 'smtpPort' property.
    int smtpPort() const; ///< Getter for the 'smtpPort' property.
    bool isDoHEnabled() const; ///< Getter for the 'isDoHEnabled' property.
    bool isAutomaticUpdateOn() const; ///< Getter for the 'isAutomaticUpdateOn' property.
    QString currentEmailClient() const; ///< Getter for the 'currentEmail' property.
    QStringList availableKeychain() const; ///< Getter for the 'availableKeychain' property.
    QString currentKeychain() const; ///< Getter for the 'currentKeychain' property.
    QVariantList bugCategories() const; ///< Getter for the 'bugCategories' property.
    QVariantList bugQuestions() const; ///< Getter for the 'bugQuestions' property.
    void setDockIconVisible(bool visible); ///< Setter for the 'dockIconVisible' property.
    bool dockIconVisible() const;; ///< Getter for the 'dockIconVisible' property.

signals: // Signal used by the Qt property system. Many of them are unused but required to avoid warning from the QML engine.
    void showSplashScreenChanged(bool value); ///<Signal for the change of the 'showSplashScreen' property.
    void showOnStartupChanged(bool value); ///<Signal for the change of the 'showOnStartup' property.
    void goosChanged(QString const &value); ///<Signal for the change of the 'GOOS' property.
    void diskCachePathChanged(QUrl const &url); ///<Signal for the change of the 'diskCachePath' property.
    void imapPortChanged(int port); ///<Signal for the change of the 'imapPort' property.
    void smtpPortChanged(int port); ///<Signal for the change of the 'smtpPort' property.
    void useSSLForSMTPChanged(bool value); ///<Signal for the change of the 'useSSLForSMTP' property.
    void useSSLForIMAPChanged(bool value); ///<Signal for the change of the 'useSSLForIMAP' property.
    void isAutomaticUpdateOnChanged(bool value); ///<Signal for the change of the 'isAutomaticUpdateOn' property.
    void isBetaEnabledChanged(bool value); ///<Signal for the change of the 'isBetaEnabled' property.
    void isAllMailVisibleChanged(bool value); ///<Signal for the change of the 'isAllMailVisible' property.
    void isTelemetryDisabledChanged(bool isDisabled); ///<Signal for the change of the 'isTelemetryDisabled' property.
    void colorSchemeNameChanged(QString const &scheme); ///<Signal for the change of the 'colorSchemeName' property.
    void isDoHEnabledChanged(bool value); ///<Signal for the change of the 'isDoHEnabled' property.
    void logsPathChanged(QUrl const &path); ///<Signal for the change of the 'logsPath' property.
    void licensePathChanged(QUrl const &path); ///<Signal for the change of the 'licensePath' property.
    void releaseNotesLinkChanged(QUrl const &link); ///<Signal for the change of the 'releaseNotesLink' property.
    void dependencyLicensesLinkChanged(QUrl const &link); ///<Signal for the change of the 'dependencyLicensesLink' property.
    void landingPageLinkChanged(QUrl const &link); ///<Signal for the change of the 'landingPageLink' property.
    void appnameChanged(QString const &appname); ///<Signal for the change of the 'appname' property.
    void vendorChanged(QString const &vendor); ///<Signal for the change of the 'vendor' property.
    void versionChanged(QString const &version); ///<Signal for the change of the 'version' property.
    void tagChanged(QString const &tag); ///<Signal for the change of the 'tag' property.
    void currentEmailClientChanged(QString const &email); ///<Signal for the change of the 'currentEmailClient' property.
    void currentKeychainChanged(QString const &keychain); ///<Signal for the change of the 'currentKeychain' property.
    void bugCategoriesChanged(QVariantList const &bugCategories); ///<Signal for the change of the 'bugCategories' property.
    void bugQuestionsChanged(QVariantList const &bugQuestions); ///<Signal for the change of the 'bugQuestions' property.
    void availableKeychainChanged(QStringList const &keychains); ///<Signal for the change of the 'availableKeychain' property.
    void hostnameChanged(QString const &hostname); ///<Signal for the change of the 'hostname' property.
    void isAutostartOnChanged(bool value); ///<Signal for the change of the 'isAutostartOn' property.
    void usersChanged(UserList *users); ///<Signal for the change of the 'users' property.
    void dockIconVisibleChanged(bool value); ///<Signal for the change of the 'dockIconVisible' property.
    void receivedUserNotification(bridgepp::UserNotification const& notification); ///< Signal to display the userNotification modal


public slots: // slot for signals received from QML -> To be forwarded to Bridge via RPC Client calls.
    void toggleAutostart(bool active); ///< Slot for the autostart toggle.
    void toggleBeta(bool active); ///< Slot for the beta toggle.
    void changeIsAllMailVisible(bool isVisible); ///< Slot for the changing of 'All Mail' visibility.
    void toggleIsTelemetryDisabled(bool isDisabled); ///< Slot for toggling telemetry on/off.
    void changeColorScheme(QString const &scheme); ///< Slot for the change of the theme.
    void setDiskCachePath(QUrl const &path) const; ///< Slot for the change of the disk cache path.
    void login(QString const &username, QString const &password) const; ///< Slot for the login button (initial login).
    void loginHv(QString const &username, QString const &password) const; ///< Slot for the login button (after HV challenge completed).
    void login2FA(QString const &username, QString const &code) const; ///< Slot for the login button (2FA login).
    void login2Password(QString const &username, QString const &password) const; ///< Slot for the login button (mailbox password login).
    void loginAbort(QString const &username) const; ///< Slot for the login abort procedure.
    void toggleDoH(bool active); ///, Slot for the DoH toggle.
    void toggleAutomaticUpdate(bool makeItActive); ///< Slot for the automatic update toggle
    void updateCurrentMailClient(); ///< Slot for the change of the current mail client.
    void changeKeychain(QString const &keychain); ///< Slot for the change of keychain.
    void guiReady(); ///< Slot for the GUI ready signal.
    void quit() const; ///< Slot for the quit signal.
    void restart() const; ///< Slot for the restart signal.
    void forceLauncher(QString launcher) const; ///< Slot for the change of the launcher.
    void checkUpdates() const; ///< Slot for the update check.
    void installUpdate() const; ///< Slot for the update install.
    void triggerReset() const; ///< Slot for the triggering of reset.
    void reportBug(QString const &category, QString const &description, QString const &address, QString const &emailClient, bool includeLogs) const; ///< Slot for the bug report.
    void installTLSCertificate(); ///< Installs the Bridge TLS certificate in the Keychain.
    void exportTLSCertificates() const; ///< Slot for the export of the TLS certificates.
    void onResetFinished(); ///< Slot for the reset finish signal.
    void onVersionChanged(); ///< Slot for the version change signal.
    void setMailServerSettings(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP) const; ///< Forwards a connection mode change request from QML to gRPC
    void sendBadEventUserFeedback(QString const &userID, bool doResync); ///< Slot the providing user feedback for a bad event.
    void notifyReportBugClicked() const; ///< Slot for the ReportBugClicked gRPC event.
    void notifyAutoconfigClicked(QString const &client) const; ///< Slot for gAutoconfigClicked gRPC event.
    void notifyExternalLinkClicked(QString const &article) const; ///< Slot for KBArticleClicked gRPC event.
    void triggerRepair() const; ///< Slot for the triggering of the bridge repair function i.e. 'resync'.
    void userNotificationDismissed(); ///< Slot to pop the notification from the stack and display the rest.

public slots: // slots for functions that need to be processed locally.
    void setNormalTrayIcon(); ///< Set the tray icon to normal.
    void setErrorTrayIcon(QString const& stateString, QString const &statusIcon); ///< Set the tray icon to 'error' state.
    void setWarnTrayIcon(QString const& stateString, QString const &statusIcon); ///< Set the tray icon to 'warn' state.
    void setUpdateTrayIcon(QString const& stateString, QString const &statusIcon); ///< Set the tray icon to 'update' state.

public slots: // slot for signals received from gRPC that need transformation instead of simple forwarding
    void internetStatusChanged(bool isOn); ///< Check if bridge considers internet as on.
    void onMailServerSettingsChanged(int imapPort, int smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Slot for the ConnectionModeChanged gRPC event.
    void onGenericError(bridgepp::ErrorInfo const &info); ///< Slot for generic errors received from the gRPC service.
    void onLoginFinished(QString const &userID, bool wasSignedOut); ///< Slot for LoginFinished gRPC event.
    void onLoginAlreadyLoggedIn(QString const &userID); ///< Slot for the LoginAlreadyLoggedIn gRPC event.
    void onUserBadEvent(QString const& userID, QString const& errorMessage); ///< Slot for the userBadEvent gRPC event.
    void onIMAPLoginFailed(QString const& username); ///< Slot the the imapLoginFailed event.
    void processUserNotification(bridgepp::UserNotification const& notification); ///< Slot for the userNotificationReceived gRCP event.

signals: // Signals received from the Go backend, to be forwarded to QML
    void toggleAutostartFinished(); ///< Signal for the 'toggleAutostartFinished' gRPC stream event.
    void cantMoveDiskCache(); ///< Signal for the 'cantMoveDiskCache' gRPC stream event.
    void diskCachePathChangeFinished(); ///< Signal for the 'diskCachePathChangeFinished' gRPC stream event.
    void loginUsernamePasswordError(QString const &errorMsg); ///< Signal for the 'loginUsernamePasswordError' gRPC stream event.
    void loginFreeUserError(); ///< Signal for the 'loginFreeUserError' gRPC stream event.
    void loginConnectionError(QString const &errorMsg); ///< Signal for the 'loginConnectionError' gRPC stream event.
    void login2FARequested(QString const &username); ///< Signal for the 'login2FARequested' gRPC stream event.
    void login2FAError(QString const &errorMsg); ///< Signal for the 'login2FAError' gRPC stream event.
    void login2FAErrorAbort(QString const &errorMsg); ///< Signal for the 'login2FAErrorAbort' gRPC stream event.
    void login2PasswordRequested(QString const &username); ///< Signal for the 'login2PasswordRequested' gRPC stream event.
    void login2PasswordError(QString const &errorMsg); ///< Signal for the 'login2PasswordError' gRPC stream event.
    void login2PasswordErrorAbort(QString const &errorMsg); ///< Signal for the 'login2PasswordErrorAbort' gRPC stream event.
    void loginFinished(int index, bool wasSignedOut); ///< Signal for the 'loginFinished' gRPC stream event.
    void loginAlreadyLoggedIn(int index); ///< Signal for the 'loginAlreadyLoggedIn' gRPC stream event.
    void loginHvRequested(QString const &hvUrl); ///< Signal for the 'loginHvRequested' gRPC stream event.
    void loginHvError(QString const &errorMsg); ///< Signal for the 'loginHvError' gRPC stream event.
    void updateManualReady(QString const &version); ///< Signal for the 'updateManualReady' gRPC stream event.
    void updateManualRestartNeeded(); ///< Signal for the 'updateManualRestartNeeded' gRPC stream event.
    void updateManualError(); ///< Signal for the 'updateManualError' gRPC stream event.
    void updateForce(QString const &version); ///< Signal for the 'updateForce' gRPC stream event.
    void updateForceError(); ///< Signal for the 'updateForceError' gRPC stream event.
    void updateSilentRestartNeeded(); ///< Signal for the 'updateSilentRestartNeeded' gRPC stream event.
    void updateSilentError(); ///< Signal for the 'updateSilentError' gRPC stream event.
    void updateIsLatestVersion(); ///< Signal for the 'updateIsLatestVersion' gRPC stream event.
    void checkUpdatesFinished(); ///< Signal for the 'checkUpdatesFinished' gRPC stream event.
    void imapPortStartupError(); ///< Signal for the 'imapPortStartupError' gRPC stream event.
    void smtpPortStartupError(); ///< Signal for the 'smtpPortStartupError' gRPC stream event.
    void imapPortChangeError(); ///< Signal for the 'imapPortChangeError' gRPC stream event.
    void smtpPortChangeError(); ///< Signal for the 'smtpPortChangeError' gRPC stream event.
    void imapConnectionModeChangeError(); ///< Signal for the 'imapConnectionModeChangeError' gRPC stream event.
    void smtpConnectionModeChangeError(); ///< Signal for the 'smtpConnectionModeChangeError' gRPC stream event.
    void changeMailServerSettingsFinished(); ///< Signal for the 'changeMailServerSettingsFinished' gRPC stream event.
    void changeKeychainFinished(); ///< Signal for the 'changeKeychainFinished' gRPC stream event.
    void notifyHasNoKeychain(); ///< Signal for the 'notifyHasNoKeychain' gRPC stream event.
    void notifyRebuildKeychain(); ///< Signal for the 'notifyRebuildKeychain' gRPC stream event.
    void addressChanged(QString const &address); ///< Signal for the 'addressChanged' gRPC stream event.
    void addressChangedLogout(QString const &address); ///< Signal for the 'addressChangedLogout' gRPC stream event.
    void apiCertIssue(); ///< Signal for the 'apiCertIssue' gRPC stream event.
    void userDisconnected(QString const &username); ///< Signal for the 'userDisconnected' gRPC stream event.
    void userBadEvent(QString const &userID, QString const &description); ///< Signal for the 'userBadEvent' gRPC stream event.
    void internetOff(); ///< Signal for the 'internetOff' gRPC stream event.
    void internetOn(); ///< Signal for the 'internetOn' gRPC stream event.
    void resetFinished(); ///< Signal for the 'resetFinished' gRPC stream event.
    void reportBugFinished(); ///< Signal for the 'reportBugFinished' gRPC stream event.
    void bugReportSendSuccess(); ///< Signal for the 'bugReportSendSuccess' gRPC stream event.
    void bugReportSendFallback(); ///< Signal for the 'bugReportSendFallback' gRPC stream event.
    void bugReportSendError(); ///< Signal for the 'bugReportSendError' gRPC stream event.
    void certificateInstallSuccess(); ///< Signal for the 'certificateInstallSuccess' gRPC stream event.
    void certificateInstallCanceled(); ///< Signal for the 'certificateInstallCanceled' gRPC stream event.
    void certificateInstallFailed(); /// Signal for the 'certificateInstallFailed' gRPC stream event.
    void showMainWindow(); ///< Signal for the 'showMainWindow' gRPC stream event.
    void hideMainWindow(); ///< Signal for the 'hideMainWindow' gRPC stream event.
    void showHelp(); ///< Signal for the 'showHelp' event (from the context menu).
    void showSettings(); ///< Signal for the 'showSettings' event (from the context menu).
    void selectUser(QString const& userID, bool forceShowWindow); ///< Signal emitted in order to selected a user with a given ID in the list.
    void genericError(QString const &title, QString const &description); ///< Signal for the 'genericError' gRPC stream event.
    void imapLoginWhileSignedOut(QString const& username); ///< Signal for the notification of IMAP login attempt on a signed out account.
    void receivedKnowledgeBaseSuggestions(QList<bridgepp::KnowledgeBaseSuggestion> const& suggestions); ///< Signal for the reception of knowledge base article suggestions.
    void repairStarted(); ///< Signal for the 'repairStarted' gRPC stream event.
    void allUsersLoaded(); ///< Signal for the 'allUsersLoaded' gRPC stream event

    // This signal is emitted when an exception is intercepted is calls triggered by QML. QML engine would intercept the exception otherwise.
    void fatalError(bridgepp::Exception const& e) const; ///< Signal emitted when an fatal error occurs.

private: // member functions
    void retrieveUserList(); ///< Retrieve the list of users via gRPC.
    void connectGrpcEvents(); ///< Connect gRPC that need to be forwarded to QML via backend signals
    void displayBadEventDialog(QString const& userID); ///< Displays the bad event dialog for a user.

private: // data members
    UserList *users_ { nullptr }; ///< The user list. Owned by backend.
    std::unique_ptr<bridgepp::Overseer> eventStreamOverseer_; ///< The event stream overseer.
    bool showSplashScreen_ { false }; ///< The cached version of show splash screen. Retrieved on startup from bridge, and potentially modified locally.
    QString goos_; ///< The cached version of the GOOS variable.
    QUrl logsPath_; ///< The logs path. Retrieved from bridge on startup.
    QUrl licensePath_; ///< The license path. Retrieved from bridge on startup.
    int imapPort_ { 0 }; ///< The cached value for the IMAP port.
    int smtpPort_ { 0 }; ///< The cached value for the SMTP port.
    bool useSSLForIMAP_ { false }; ///< The cached value for useSSLForIMAP.
    bool useSSLForSMTP_ { false }; ///< The cached value for useSSLForSMTP.
    bool isInternetOn_ { true }; ///< Does bridge consider internet as on?
    QList<QString> badEventDisplayQueue_; ///< THe queue for displaying 'bad event feedback request dialog'.
    std::unique_ptr<TrayIcon> trayIcon_; ///< The tray icon for the application.
    bridgepp::BugReportFlow reportFlow_;  ///< The bug report flow.
    std::stack<bridgepp::UserNotification> userNotificationStack_; ///< The stack which holds all of the active notifications that the user needs to acknowledge.
    friend class AppController;
};


#endif // BRIDGE_GUI_QML_BACKEND_H
