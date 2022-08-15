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

    grpc::Status AddLogEntry(::grpc::ServerContext *, ::grpc::AddLogEntryRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status GuiReady(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status Quit(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status Restart(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status ShowOnStartup(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status ShowSplashScreen(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status IsFirstGuiStart(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetIsAutostartOn(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status IsAutostartOn(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetIsBetaEnabled(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status IsBetaEnabled(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status GoOs(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status TriggerReset(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status Version(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status LogsPath(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status LicensePath(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status ReleaseNotesPageLink(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status DependencyLicensesLink(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status LandingPageLink(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status SetColorSchemeName(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status ColorSchemeName(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status CurrentEmailClient(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status ReportBug(::grpc::ServerContext *, ::grpc::ReportBugRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status ForceLauncher(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status Login(::grpc::ServerContext *, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status Login2FA(::grpc::ServerContext *, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status Login2Passwords(::grpc::ServerContext *, ::grpc::LoginRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status LoginAbort(::grpc::ServerContext *, ::grpc::LoginAbortRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status CheckUpdate(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status InstallUpdate(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;
    grpc::Status SetIsAutomaticUpdateOn(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status IsAutomaticUpdateOn(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status IsCacheOnDiskEnabled(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status DiskCachePath(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status ChangeLocalCache(::grpc::ServerContext *, ::grpc::ChangeLocalCacheRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status SetIsDoHEnabled(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status IsDoHEnabled(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status SetUseSslForSmtp(::grpc::ServerContext *, ::google::protobuf::BoolValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status UseSslForSmtp(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::BoolValue *response) override;
    grpc::Status Hostname(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status ImapPort(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Int32Value *response) override;
    grpc::Status SmtpPort(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Int32Value *response) override;
    grpc::Status ChangePorts(::grpc::ServerContext *, ::grpc::ChangePortsRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status IsPortFree(::grpc::ServerContext *, ::google::protobuf::Int32Value const *request, ::google::protobuf::BoolValue *response) override;
    grpc::Status AvailableKeychains(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::grpc::AvailableKeychainsResponse *response) override;
    grpc::Status SetCurrentKeychain(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status CurrentKeychain(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::StringValue *response) override;
    grpc::Status GetUserList(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::grpc::UserListResponse *response) override;
    grpc::Status GetUser(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::grpc::User *response) override;
    grpc::Status SetUserSplitMode(::grpc::ServerContext *, ::grpc::UserSplitModeRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status LogoutUser(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status RemoveUser(::grpc::ServerContext *, ::google::protobuf::StringValue const *request, ::google::protobuf::Empty *) override;
    grpc::Status ConfigureUserAppleMail(::grpc::ServerContext *, ::grpc::ConfigureAppleMailRequest const *request, ::google::protobuf::Empty *) override;
    grpc::Status StartEventStream(::grpc::ServerContext *, ::grpc::EventStreamRequest const *request, ::grpc::ServerWriter<::grpc::StreamEvent> *writer) override;
    grpc::Status StopEventStream(::grpc::ServerContext *, ::google::protobuf::Empty const*, ::google::protobuf::Empty *) override;

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
