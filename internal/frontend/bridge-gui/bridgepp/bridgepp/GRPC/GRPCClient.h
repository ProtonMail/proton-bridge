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


#ifndef BRIDGE_PP_RPC_CLIENT_H
#define BRIDGE_PP_RPC_CLIENT_H


#include "../User/User.h"
#include "../Log/Log.h"
#include "bridge.grpc.pb.h"
#include "grpc++/grpc++.h"


namespace bridgepp
{


typedef grpc::Status (grpc::Bridge::Stub::*SimpleMethod)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*BoolSetter)(grpc::ClientContext *, const google::protobuf::BoolValue &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*BoolGetter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::BoolValue *);
typedef grpc::Status (grpc::Bridge::Stub::*Int32Setter)(grpc::ClientContext *, const google::protobuf::Int32Value &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*Int32Getter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::Int32Value *);
typedef grpc::Status (grpc::Bridge::Stub::*StringGetter)(grpc::ClientContext *, const google::protobuf::Empty &, google::protobuf::StringValue *);
typedef grpc::Status (grpc::Bridge::Stub::*StringSetter)(grpc::ClientContext *, const google::protobuf::StringValue &, google::protobuf::Empty *);
typedef grpc::Status (grpc::Bridge::Stub::*StringParamMethod)(grpc::ClientContext *, const google::protobuf::StringValue &, google::protobuf::Empty *);


//****************************************************************************************************************************************************
/// \brief gRPC client class. This class encapsulate the gRPC service, abstracting all data type conversions.
//****************************************************************************************************************************************************
class GRPCClient : public QObject
{
Q_OBJECT
public: // member functions.
    GRPCClient() = default; ///< Default constructor.
    GRPCClient(GRPCClient const &) = delete; ///< Disabled copy-constructor.
    GRPCClient(GRPCClient &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCClient() override = default; ///< Destructor.
    GRPCClient &operator=(GRPCClient const &) = delete; ///< Disabled assignment operator.
    GRPCClient &operator=(GRPCClient &&) = delete; ///< Disabled move assignment operator.
    void setLog(Log *log); ///< Set the log for the client.
    bool connectToServer(QString &outError); ///< Establish connection to the gRPC server.

    grpc::Status addLogEntry(Log::Level level, QString const &package, QString const &message); ///< Performs the "AddLogEntry" gRPC call.
    grpc::Status guiReady(); ///< performs the "GuiReady" gRPC call.
    grpc::Status isFirstGUIStart(bool &outIsFirst); ///< performs the "IsFirstGUIStart" gRPC call.
    grpc::Status isAutostartOn(bool &outIsOn); ///< Performs the "isAutostartOn" gRPC call.
    grpc::Status setIsAutostartOn(bool on); ///< Performs the "setIsAutostartOn" gRPC call.
    grpc::Status isBetaEnabled(bool &outEnabled); ///< Performs the "isBetaEnabled" gRPC call.
    grpc::Status setIsBetaEnabled(bool enabled); ///< Performs the 'setIsBetaEnabled' gRPC call.
    grpc::Status isAllMailVisible(bool &outIsVisible); ///< Performs the "isAllMailVisible" gRPC call.
    grpc::Status setIsAllMailVisible(bool isVisible); ///< Performs the 'setIsAllMailVisible' gRPC call.
    grpc::Status colorSchemeName(QString &outName); ///< Performs the "colorSchemeName' gRPC call.
    grpc::Status setColorSchemeName(QString const &name); ///< Performs the "setColorSchemeName' gRPC call.
    grpc::Status currentEmailClient(QString &outName); ///< Performs the 'currentEmailClient' gRPC call.
    grpc::Status reportBug(QString const &description, QString const &address, QString const &emailClient, bool includeLogs); ///< Performs the 'ReportBug' gRPC call.
    grpc::Status quit(); ///< Perform the "Quit" gRPC call.
    grpc::Status restart(); ///< Performs the Restart gRPC call.
    grpc::Status triggerReset(); ///< Performs the triggerReset gRPC call.
    grpc::Status forceLauncher(QString const &launcher); ///< Performs the 'ForceLauncher' call.
    grpc::Status setMainExecutable(QString const &exe); ///< Performs the 'SetMainExecutable' call.
    grpc::Status isPortFree(qint32 port, bool &outFree); ///< Performs the 'IsPortFree' call.
    grpc::Status showOnStartup(bool &outValue); ///< Performs the 'ShowOnStartup' call.
    grpc::Status showSplashScreen(bool &outValue); ///< Performs the 'ShowSplashScreen' call.
    grpc::Status goos(QString &outGoos); ///< Performs the 'GoOs' call.
    grpc::Status logsPath(QUrl &outPath); ///< Performs the 'LogsPath' call.
    grpc::Status licensePath(QUrl &outPath); ///< Performs the 'LicensePath' call.
    grpc::Status dependencyLicensesLink(QUrl &outUrl); ///< Performs the 'DependencyLicensesLink' call.
    grpc::Status version(QString &outVersion); ///< Performs the 'Version' call.
    grpc::Status releaseNotesPageLink(QUrl &outUrl); ///< Performs the 'releaseNotesPageLink' call.
    grpc::Status landingPageLink(QUrl &outUrl); ///< Performs the 'landingPageLink' call.
    grpc::Status hostname(QString &outHostname); ///< Performs the 'Hostname' call.

signals: // app related signals
    void internetStatus(bool isOn);
    void toggleAutostartFinished();
    void resetFinished();
    void reportBugFinished();
    void reportBugSuccess();
    void reportBugError();
    void showMainWindow();

    // cache related calls
public:
    grpc::Status isCacheOnDiskEnabled(bool &outEnabled); ///< Performs the 'isCacheOnDiskEnabled' call.
    grpc::Status diskCachePath(QUrl &outPath); ///< Performs the 'diskCachePath' call.
    grpc::Status changeLocalCache(bool enabled, QUrl const &path); ///< Performs the 'ChangeLocalCache' call.
signals:
    void isCacheOnDiskEnabledChanged(bool enabled);
    void diskCachePathChanged(QUrl const &outPath);
    void cacheUnavailable();                                                                                            //    _ func()                  `signal:"cacheUnavailable"`
    void cacheCantMove();                                                                                               //    _ func()                  `signal:"cacheCantMove"`
    void cacheLocationChangeSuccess();                                                                                  //    _ func()                  `signal:"cacheLocationChangeSuccess"`
    void diskFull();                                                                                                    //    _ func()                  `signal:"diskFull"`
    void changeLocalCacheFinished();                                                                                    //    _ func()                  `signal:"changeLocalCacheFinished"`


    // mail settings related calls
public:
    grpc::Status useSSLForSMTP(bool &outUseSSL); ///< Performs the 'useSSLForSMTP' gRPC call
    grpc::Status setUseSSLForSMTP(bool useSSL); ///< Performs the 'currentEmailClient' gRPC call.
    grpc::Status portIMAP(int &outPort); ///< Performs the 'portImap' gRPC call.
    grpc::Status portSMTP(int &outPort); ///< Performs the 'portImap' gRPC call.
    grpc::Status changePorts(int portIMAP, int portSMTP); ///< Performs the 'changePorts' gRPC call.
    grpc::Status isDoHEnabled(bool &outEnabled); ///< Performs the 'isDoHEnabled' gRPC call.
    grpc::Status setIsDoHEnabled(bool enabled); ///< Performs the 'setIsDoHEnabled' gRPC call.

signals:
    void portIssueIMAP();
    void portIssueSMTP();
    void toggleUseSSLFinished();
    void changePortFinished();

public: // login related calls
    grpc::Status login(QString const &username, QString const &password); ///< Performs the 'login' call.
    grpc::Status login2FA(QString const &username, QString const &code); ///< Performs the 'login2FA' call.
    grpc::Status login2Passwords(QString const &username, QString const &password); ///< Performs the 'login2Passwords' call.
    grpc::Status loginAbort(QString const &username); ///< Performs the 'loginAbort' call.

signals:
    void loginUsernamePasswordError(QString const &errMsg);                                                             //    _ func(errorMsg string)   `signal:"loginUsernamePasswordError"`
    void loginFreeUserError();                                                                                          //    _ func()                  `signal:"loginFreeUserError"`
    void loginConnectionError(QString const &errMsg);                                                                   //    _ func(errorMsg string)   `signal:"loginConnectionError"`
    void login2FARequested(QString const &userName);                                                                    //    _ func(username string)   `signal:"login2FARequested"`
    void login2FAError(QString const &errMsg);                                                                          //    _ func(errorMsg string)   `signal:"login2FAError"`
    void login2FAErrorAbort(QString const &errMsg);                                                                     //    _ func(errorMsg string)   `signal:"login2FAErrorAbort"`
    void login2PasswordRequested();                                                                                     //    _ func()                  `signal:"login2PasswordRequested"`
    void login2PasswordError(QString const &errMsg);                                                                    //    _ func(errorMsg string)   `signal:"login2PasswordError"`
    void login2PasswordErrorAbort(QString const &errMsg);                                                               //    _ func(errorMsg string)   `signal:"login2PasswordErrorAbort"`
    void loginFinished(QString const &userID);                                                                         //    _ func(index int)         `signal:"loginFinished"`
    void loginAlreadyLoggedIn(QString const &userID);                                                                  //    _ func(index int)         `signal:"loginAlreadyLoggedIn"`

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

signals:
    void toggleSplitModeFinished(QString const &userID);
    void userDisconnected(QString const &username);
    void userChanged(QString const &userID);


public: // keychain related calls
    grpc::Status availableKeychains(QStringList &outKeychains);
    grpc::Status currentKeychain(QString &outKeychain);
    grpc::Status setCurrentKeychain(QString const &keychain);

signals:
    void changeKeychainFinished();
    void hasNoKeychain();
    void rebuildKeychain();
    void certIsReady();

signals: // mail related events
    void noActiveKeyForRecipient(QString const &email);                                                                 //    _ func(email string)      `signal:noActiveKeyForRecipient`
    void addressChanged(QString const &address);                                                                        //    _ func(address string)    `signal:addressChanged`
    void addressChangedLogout(QString const &address);                                                                  //    _ func(address string)    `signal:addressChangedLogout`
    void apiCertIssue();

public:
    bool isEventStreamActive() const; ///< Check if the event stream is active.
    grpc::Status runEventStreamReader(); ///< Retrieve and signal the events in the event stream.
    grpc::Status stopEventStreamReader(); ///< Stop the event stream.

private slots:
    void configFolderChanged();

private:
    void logDebug(QString const &message); ///< Log an event.
    void logError(QString const &message); ///< Log an event.
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

    std::string getServerCertificate(); ///< Wait until server certificates is generated and retrieve it.
    void processAppEvent(grpc::AppEvent const &event); ///< Process an 'App' event.
    void processLoginEvent(grpc::LoginEvent const &event); ///< Process a 'Login' event.
    void processUpdateEvent(grpc::UpdateEvent const &event); ///< Process an 'Update' event.
    void processCacheEvent(grpc::CacheEvent const &event); ///< Process a 'Cache' event.
    void processMailSettingsEvent(grpc::MailSettingsEvent const &event); ///< Process a 'MailSettings' event.
    void processKeychainEvent(grpc::KeychainEvent const &event); ///< Process a 'Keychain' event.
    void processMailEvent(grpc::MailEvent const &event); ///< Process a 'Mail' event.
    void processUserEvent(grpc::UserEvent const &event); ///< Process a 'User' event.

private: // data members.
    Log *log_ { nullptr }; ///< The log for the GRPC client.
    std::shared_ptr<grpc::Channel> channel_ { nullptr }; ///< The gRPC channel.
    std::shared_ptr<grpc::Bridge::Stub> stub_ { nullptr }; ///< The gRPC stub (a.k.a. client).
    mutable QMutex eventStreamMutex_; ///< The event stream mutex.
    std::unique_ptr<grpc::ClientContext> eventStreamContext_; /// the client context for the gRPC event stream. Access protected by  eventStreamMutex_.
};


}


#endif // BRIDGE_PP_RPC_CLIENT_H
