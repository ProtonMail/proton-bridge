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


#ifndef BRIDGE_PP_RPC_CLIENT_H
#define BRIDGE_PP_RPC_CLIENT_H


#include "../User/User.h"
#include "../Log/Log.h"
#include "GRPCConfig.h"
#include "GRPCErrors.h"
#include "bridge.grpc.pb.h"
#include "grpc++/grpc++.h"


namespace bridgepp {


typedef grpc::Status (grpc::Bridge::Stub::*SimpleMethod)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*BoolSetter)(grpc::ClientContext *, const google::protobuf::BoolValue &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*BoolGetter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::BoolValue *);
typedef grpc::Status (grpc::Bridge::Stub::*Int32Setter)(grpc::ClientContext *, const google::protobuf::Int32Value &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*Int32Getter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::Int32Value *);
typedef grpc::Status (grpc::Bridge::Stub::*StringGetter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::StringValue *);
typedef grpc::Status (grpc::Bridge::Stub::*StringSetter)(grpc::ClientContext *, const google::protobuf::StringValue &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*StringParamMethod)(grpc::ClientContext *, const google::protobuf::StringValue &, google::protobuf::Empty *);
typedef std::unique_ptr<grpc::ClientContext> UPClientContext;


//****************************************************************************************************************************************************
/// \brief A struct for knowledge base suggestion.
//****************************************************************************************************************************************************
struct KnowledgeBaseSuggestion {
    //  The following lines make the type transmissible to QML (but not instanciable there)
    Q_GADGET
    Q_PROPERTY(QString url MEMBER url)
    Q_PROPERTY(QString title MEMBER title)
public:
    QString url; ///< The URL of the knowledge base article
    QString title; ///< The title of the knowledge base article.
};


//****************************************************************************************************************************************************
/// \brief gRPC client class. This class encapsulate the gRPC service, abstracting all data type conversions.
//****************************************************************************************************************************************************
class GRPCClient : public QObject {
Q_OBJECT
public: // static member functions
    static void removeServiceConfigFile(QString const &configDir); ///< Delete the service config file.
    static GRPCConfig waitAndRetrieveServiceConfig(QString const &sessionID, QString const &configDir, qint64 timeoutMs,
        class ProcessMonitor *serverProcess); ///< Wait and retrieve the service configuration.

public: // member functions.
    GRPCClient() = default; ///< Default constructor.
    GRPCClient(GRPCClient const &) = delete; ///< Disabled copy-constructor.
    GRPCClient(GRPCClient &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCClient() override = default; ///< Destructor.
    GRPCClient &operator=(GRPCClient const &) = delete; ///< Disabled assignment operator.
    GRPCClient &operator=(GRPCClient &&) = delete; ///< Disabled move assignment operator.
    void setLog(Log *log); ///< Set the log for the client.
    void connectToServer(QString const &sessionID, QString const &configDir, GRPCConfig const &config, class ProcessMonitor *serverProcess); ///< Establish connection to the gRPC server.
    bool isConnected() const; ///< Check whether the gRPC client is connected to the server.

    grpc::Status checkTokens(QString const &clientConfigPath, QString &outReturnedClientToken); ///< Performs a token check.
    grpc::Status addLogEntry(Log::Level level, QString const &package, QString const &message); ///< Performs the "AddLogEntry" gRPC call.
    grpc::Status guiReady(bool &outShowSplashScreen); ///< performs the "GuiReady" gRPC call.
    grpc::Status isAutostartOn(bool &outIsOn); ///< Performs the "isAutostartOn" gRPC call.
    grpc::Status setIsAutostartOn(bool on); ///< Performs the "setIsAutostartOn" gRPC call.
    grpc::Status isBetaEnabled(bool &outEnabled); ///< Performs the "isBetaEnabled" gRPC call.
    grpc::Status setIsBetaEnabled(bool enabled); ///< Performs the 'setIsBetaEnabled' gRPC call.
    grpc::Status isAllMailVisible(bool &outIsVisible); ///< Performs the "isAllMailVisible" gRPC call.
    grpc::Status setIsAllMailVisible(bool isVisible); ///< Performs the 'setIsAllMailVisible' gRPC call.
    grpc::Status isTelemetryDisabled(bool &outIsDisabled); ///< Performs the 'setIsTelemetryDisabled' gRPC call.
    grpc::Status setIsTelemetryDisabled(bool isDisabled); ///< Performs the 'isTelemetryDisabled' gRPC call.
    grpc::Status colorSchemeName(QString &outName); ///< Performs the "colorSchemeName' gRPC call.
    grpc::Status setColorSchemeName(QString const &name); ///< Performs the "setColorSchemeName' gRPC call.
    grpc::Status currentEmailClient(QString &outName); ///< Performs the 'currentEmailClient' gRPC call.
    grpc::Status reportBug(QString const &category, QString const &description, QString const &address, QString const &emailClient, bool includeLogs); ///< Performs the 'ReportBug' gRPC call.
    grpc::Status quit(); ///< Perform the "Quit" gRPC call.
    grpc::Status restart(); ///< Performs the Restart gRPC call.
    grpc::Status triggerReset(); ///< Performs the triggerReset gRPC call.
    grpc::Status forceLauncher(QString const &launcher); ///< Performs the 'ForceLauncher' call.
    grpc::Status setMainExecutable(QString const &exe); ///< Performs the 'SetMainExecutable' call.
    grpc::Status isPortFree(qint32 port, bool &outFree); ///< Performs the 'IsPortFree' call.
    grpc::Status showOnStartup(bool &outValue); ///< Performs the 'ShowOnStartup' call.
    grpc::Status goos(QString &outGoos); ///< Performs the 'GoOs' call.
    grpc::Status logsPath(QUrl &outPath); ///< Performs the 'LogsPath' call.
    grpc::Status licensePath(QUrl &outPath); ///< Performs the 'LicensePath' call.
    grpc::Status dependencyLicensesLink(QUrl &outUrl); ///< Performs the 'DependencyLicensesLink' call.
    grpc::Status version(QString &outVersion); ///< Performs the 'Version' call.
    grpc::Status releaseNotesPageLink(QUrl &outUrl); ///< Performs the 'releaseNotesPageLink' call.
    grpc::Status landingPageLink(QUrl &outUrl); ///< Performs the 'landingPageLink' call.
    grpc::Status hostname(QString &outHostname); ///< Performs the 'Hostname' call.
    grpc::Status requestKnowledgeBaseSuggestions(QString const &input); ///< Performs the 'RequestKnowledgeBaseSuggestions' call.
    grpc::Status triggerRepair(); ///< Performs the triggerRepair gRPC call.

signals: // app related signals
    void internetStatus(bool isOn);
    void toggleAutostartFinished();
    void resetFinished();
    void reportBugFinished();
    void reportBugSuccess();
    void reportBugError();
    void reportBugFallback();
    void certificateInstallSuccess();
    void certificateInstallCanceled();
    void certificateInstallFailed();
    void showMainWindow();
    void knowledgeBasSuggestionsReceived(QList<KnowledgeBaseSuggestion> const& suggestions);
    void repairStarted();
    void allUsersLoaded();


public: // cache related calls
    grpc::Status diskCachePath(QUrl &outPath); ///< Performs the 'diskCachePath' call.
    grpc::Status setDiskCachePath(QUrl const &path); ///< Performs the 'setDiskCachePath' call

signals:
    void cantMoveDiskCache();
    void diskCachePathChanged(QUrl const &path);
    void diskCachePathChangeFinished();


public: // mail settings related calls
    grpc::Status mailServerSettings(qint32 &outIMAPPort, qint32 &outSMTPPort, bool &outUseSSLForIMAP, bool &outUseSSLForSMTP); ///< Performs the 'MailServerSettings' gRPC call.
    grpc::Status setMailServerSettings(qint32 imapPort, qint32 smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Performs the 'SetMailServerSettings' gRPC call.
    grpc::Status isDoHEnabled(bool &outEnabled); ///< Performs the 'isDoHEnabled' gRPC call.
    grpc::Status setIsDoHEnabled(bool enabled); ///< Performs the 'setIsDoHEnabled' gRPC call.

signals:
    void imapPortStartupError();
    void smtpPortStartupError();
    void imapPortChangeError();
    void smtpPortChangeError();
    void imapConnectionModeChangeError();
    void smtpConnectionModeChangeError();
    void mailServerSettingsChanged(qint32 imapPort, qint32 smtpPort, bool useSSLForIMAP, bool useSSLForSMTP);
    void changeMailServerSettingsFinished();

public: // login related calls
    grpc::Status login(QString const &username, QString const &password); ///< Performs the 'login' call.
    grpc::Status login2FA(QString const &username, QString const &code); ///< Performs the 'login2FA' call.
    grpc::Status login2Passwords(QString const &username, QString const &password); ///< Performs the 'login2Passwords' call.
    grpc::Status loginAbort(QString const &username); ///< Performs the 'loginAbort' call.
    grpc::Status loginHv(QString const &username, QString const &password); ///< Performs the 'login' call with additional useHv flag

signals:
    void loginUsernamePasswordError(QString const &errMsg);
    void loginFreeUserError();
    void loginConnectionError(QString const &errMsg);
    void login2FARequested(QString const &username);
    void login2FAError(QString const &errMsg);
    void login2FAErrorAbort(QString const &errMsg);
    void login2PasswordRequested(QString const &username);
    void login2PasswordError(QString const &errMsg);
    void login2PasswordErrorAbort(QString const &errMsg);
    void loginFinished(QString const &userID, bool wasSignedOut);
    void loginAlreadyLoggedIn(QString const &userID);
    void loginHvRequested(QString const &hvUrl);
    void loginHvError(QString const &errMsg);

public: // Update related calls
    grpc::Status checkUpdate();
    grpc::Status installUpdate();
    grpc::Status setIsAutomaticUpdateOn(bool on);
    grpc::Status isAutomaticUpdateOn(bool &isOn);

signals:
    void updateManualError();
    void updateForceError();
    void updateSilentError();
    void updateManualReady(QString const &version);
    void updateManualRestartNeeded();
    void updateForce(QString const &version);
    void updateSilentRestartNeeded();
    void updateIsLatestVersion();
    void checkUpdatesFinished();
    void updateVersionChanged();

public: // user related calls
    grpc::Status getUserList(QList<SPUser> &outUsers);
    grpc::Status getUser(QString const &userID, SPUser &outUser);
    grpc::Status logoutUser(QString const &userID); ///< Performs the 'logoutUser' call.
    grpc::Status removeUser(QString const &userID); ///< Performs the 'removeUser' call.
    grpc::Status configureAppleMail(QString const &userID, QString const &address); ///< Performs the 'configureAppleMail' call.
    grpc::Status setUserSplitMode(QString const &userID, bool active); ///< Performs the 'SetUserSplitMode' call.
    grpc::Status sendBadEventUserFeedback(QString const& userID, bool doResync); ///< Performs the 'SendBadEventUserFeedback' call.

signals:
    void toggleSplitModeFinished(QString const &userID);
    void userDisconnected(QString const &username);
    void userChanged(QString const &userID);
    void userBadEvent(QString const &userID, QString const& errorMessage);
    void usedBytesChanged(QString const &userID, qint64 usedBytes);
    void imapLoginFailed(QString const& username);
    void syncStarted(QString const &userID);
    void syncFinished(QString const &userID);
    void syncProgress(QString const &userID, double progress, qint64 elapsedMs, qint64 remainingMs);

public: // telemetry related calls
    grpc::Status reportBugClicked();  ///< Performs the 'reportBugClicked' call.
    grpc::Status autoconfigClicked(QString const &userID); ///< Performs the 'AutoconfigClicked' call.
    grpc::Status externalLinkClicked(QString const &userID); ///< Performs the 'KBArticleClicked' call.

public: // keychain related calls
    grpc::Status availableKeychains(QStringList &outKeychains);
    grpc::Status currentKeychain(QString &outKeychain);
    grpc::Status setCurrentKeychain(QString const &keychain);

public: // cert related calls
    grpc::Status isTLSCertificateInstalled(bool &outIsInstalled); ///< Perform the 'IsTLSCertificateInstalled' gRPC call.
    grpc::Status installTLSCertificate(); ///< Perform the 'InstallTLSCertificate' gRPC call.
    grpc::Status exportTLSCertificates(QString const &folderPath); ///< Performs the 'ExportTLSCertificates' gRPC call.

signals:
    void changeKeychainFinished();
    void hasNoKeychain();
    void rebuildKeychain();
    void certIsReady();

signals: // mail related events
    void addressChanged(QString const &address);
    void addressChangedLogout(QString const &address);
    void apiCertIssue();

signals: // errors events
    void genericError(ErrorInfo info);

public:
    bool isEventStreamActive() const; ///< Check if the event stream is active.
    grpc::Status runEventStreamReader(); ///< Retrieve and signal the events in the event stream.
    grpc::Status stopEventStreamReader(); ///< Stop the event stream.

private:
    void log(Log::Level level, QString const &message); ///< Log an event.
    void logTrace(QString const &message); ///< Log a trace event.
    void logDebug(QString const &message); ///< Log a debug event.
    void logError(QString const &message); ///< Log an error event.
    void logInfo(QString const &message); ///< Log an info event.

    grpc::Status logGRPCCallStatus(grpc::Status const &status, QString const &callName, QList<grpc::StatusCode> allowedErrors = {}); ///< Log the status of a gRPC code.
    grpc::Status simpleMethod(SimpleMethod method); ///< perform a gRPC call to a bool setter.
    grpc::Status setBool(BoolSetter setter, bool value); ///< perform a gRPC call to a bool setter.
    grpc::Status getBool(BoolGetter getter, bool &outValue); ///< perform a gRPC call to a bool getter.
    grpc::Status setInt32(Int32Setter setter, int value); ///< perform a gRPC call to an int setter.
    grpc::Status getInt32(Int32Getter getter, int &outValue); ///< perform a gRPC call to an int getter.
    grpc::Status setString(StringSetter getter, QString const &value); ///< Perform a gRPC call to a string setter.
    grpc::Status getString(StringGetter getter, QString &outValue); ///< Perform a gRPC call to a string getter.
    grpc::Status getURLForLocalFile(StringGetter getter, QUrl &outValue); ///< Perform a gRPC call to a string getter, with resulted converted to QUrl for a local file path.
    grpc::Status getURL(StringGetter getter, QUrl &outValue); ///< Perform a gRPC call to a string getter, with resulted converted to QUrl.
    grpc::Status methodWithStringParam(StringParamMethod method, QString const &str); ///< Perform a gRPC call that takes a string as a parameter and returns an Empty.
    SPUser parseGRPCUser(grpc::User const &grpcUser); ///< Parse a gRPC user struct and return a User.

    void processAppEvent(grpc::AppEvent const &event); ///< Process an 'App' event.
    void processLoginEvent(grpc::LoginEvent const &event); ///< Process a 'Login' event.
    void processUpdateEvent(grpc::UpdateEvent const &event); ///< Process an 'Update' event.
    void processCacheEvent(grpc::DiskCacheEvent const &event); ///< Process a 'Cache' event.
    void processMailServerSettingsEvent(grpc::MailServerSettingsEvent const &event); ///< Process a 'MailSettings' event.
    void processKeychainEvent(grpc::KeychainEvent const &event); ///< Process a 'Keychain' event.
    void processMailEvent(grpc::MailEvent const &event); ///< Process a 'Mail' event.
    void processUserEvent(grpc::UserEvent const &event); ///< Process a 'User' event.
    void processGenericErrorEvent(grpc::GenericErrorEvent const &event); ///< Process an 'GenericError' event.
    UPClientContext clientContext() const; ///< Returns a client context with the server token set in metadata.

private: // data members.
    Log *log_ { nullptr }; ///< The log for the GRPC client.
    std::string serverToken_; ///< The token to for communications with the gRPC server
    std::shared_ptr<grpc::Channel> channel_ { nullptr }; ///< The gRPC channel.
    std::shared_ptr<grpc::Bridge::Stub> stub_ { nullptr }; ///< The gRPC stub (a.k.a. client).
    mutable QMutex eventStreamMutex_; ///< The event stream mutex.
    UPClientContext eventStreamContext_; /// the client context for the gRPC event stream. Access protected by  eventStreamMutex_.
};


}


#endif // BRIDGE_PP_RPC_CLIENT_H
