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


#ifndef BRIDGE_GUI_TESTER_GRPC_SERVER_H
#define BRIDGE_GUI_TESTER_GRPC_SERVER_H


#include "GRPCQtProxy.h"
#include <bridgepp/GRPC/bridge.grpc.pb.h>
#include <bridgepp/GRPC/GRPCUtils.h>


//**********************************************************************************************************************
/// \brief gRPC server implementation.
//**********************************************************************************************************************
class GRPCService : public grpc::Bridge::Service
{

public: // member functions.
    GRPCService() = default; ///< Default constructor.
    GRPCService(GRPCService const &) = delete; ///< Disabled copy-constructor.
    GRPCService(GRPCService &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCService() override = default; ///< Destructor.
    GRPCService &operator=(GRPCService const &) = delete; ///< Disabled assignment operator.
    GRPCService &operator=(GRPCService &&) = delete; ///< Disabled move assignment operator.
    void connectProxySignals(); ///< Connect the signals of the Qt Proxy to the GUI components
    bool isStreaming() const; ///< Check if the service is currently streaming events.

    grpc::Status AddLogEntry(::grpc::ServerContext *context, ::grpc::AddLogEntryRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status GuiReady(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Quit(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Restart(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status ShowOnStartup(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status ShowSplashScreen(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status IsFirstGuiStart(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetIsAutostartOn(::grpc::ServerContext *context, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status IsAutostartOn(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetIsBetaEnabled(::grpc::ServerContext *context, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status IsBetaEnabled(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status GoOs(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status TriggerReset(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Version(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status LogsPath(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status LicensePath(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status ReleaseNotesPageLink(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status DependencyLicensesLink(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status LandingPageLink(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status SetColorSchemeName(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status ColorSchemeName(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status CurrentEmailClient(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status ReportBug(::grpc::ServerContext *context, ::grpc::ReportBugRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status ForceLauncher(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Login(::grpc::ServerContext *context, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Login2FA(::grpc::ServerContext *context, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status Login2Passwords(::grpc::ServerContext *context, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status LoginAbort(::grpc::ServerContext *context, ::grpc::LoginAbortRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status CheckUpdate(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status InstallUpdate(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;
    grpc::Status SetIsAutomaticUpdateOn(::grpc::ServerContext *context, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status IsAutomaticUpdateOn(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status IsCacheOnDiskEnabled(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status DiskCachePath(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status ChangeLocalCache(::grpc::ServerContext *context, ::grpc::ChangeLocalCacheRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status SetIsDoHEnabled(::grpc::ServerContext *context, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status IsDoHEnabled(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetUseSslForSmtp(::grpc::ServerContext *context, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status UseSslForSmtp(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status Hostname(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status ImapPort(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Int32Value *response) override;
    grpc::Status SmtpPort(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Int32Value *response) override;
    grpc::Status ChangePorts(::grpc::ServerContext *context, ::grpc::ChangePortsRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status IsPortFree(::grpc::ServerContext *context, ::google::protobuf::Int32Value const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status AvailableKeychains(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::grpc::AvailableKeychainsResponse *response) override;
    grpc::Status SetCurrentKeychain(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status CurrentKeychain(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::StringValue *response) override;
    grpc::Status GetUserList(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::grpc::UserListResponse *response) override;
    grpc::Status GetUser(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::grpc::User *response) override;
    grpc::Status SetUserSplitMode(::grpc::ServerContext *context, ::grpc::UserSplitModeRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status LogoutUser(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status RemoveUser(::grpc::ServerContext *context, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *response) override;
    grpc::Status ConfigureUserAppleMail(::grpc::ServerContext *context, ::grpc::ConfigureAppleMailRequest const *request, ::google::protobuf::Empty *response) override;
    grpc::Status StartEventStream(::grpc::ServerContext *context, ::grpc::EventStreamRequest const *request, ::grpc::ServerWriter<::grpc::StreamEvent> *writer) override;
    grpc::Status StopEventStream(::grpc::ServerContext *context, ::google::protobuf::Empty const *request, ::google::protobuf::Empty *response) override;

    bool sendEvent(bridgepp::SPStreamEvent const &event); ///< Queue an event for sending through the event stream.

private: // data member
    mutable QMutex eventStreamMutex_; ///< Mutex used to access eventQueue_, isStreaming_ and shouldStopStreaming_;
    QList<bridgepp::SPStreamEvent> eventQueue_; ///< The event queue. Acces protected by eventStreamMutex_;
    bool isStreaming_; ///< Is the gRPC stream running. Access protected by eventStreamMutex_;
    bool eventStreamShouldStop_; ///< Should the stream be stopped? Access protected by eventStreamMutex
    QString loginUsername_; ///< The username used for the current login procedure.
    GRPCQtProxy qtProxy_; ///< Qt Proxy used to send signals, as this class is not a QObject.
};


#endif // BRIDGE_GUI_TESTER_GRPC_SERVER_H
