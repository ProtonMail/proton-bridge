// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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


#ifndef BRIDGE_RPC_CLIENT_TEST_RPC_CLIENT_H
#define BRIDGE_RPC_CLIENT_TEST_RPC_CLIENT_H


#include "rpc/bridge_rpc.grpc.pb.h"
#include "grpc++/grpc++.h"
#include <QtCore>


//**********************************************************************************************************************
/// \brief Rpc Client class.
//**********************************************************************************************************************
class RpcClient: public QObject
{
    Q_OBJECT

public: // member functions
    explicit RpcClient(); ///< Default constructor.
    RpcClient(RpcClient const &) = delete; ///< Disabled copy-constructor.
    RpcClient(RpcClient &&) = delete; ///< Disabled assignment copy-constructor.
    ~RpcClient() override = default; ///< Destructor.
    RpcClient &operator=(RpcClient const &) = delete; ///< Disabled assignment operator.
    RpcClient &operator=(RpcClient &&) = delete; ///< Disabled move assignment operator.

private slots:
#pragma clang diagnostic push
#pragma ide diagnostic ignored "OCUnusedGlobalDeclarationInspection"

    void initTestCase(); ///< Check if the connection with the RPC server is established.
    void testGetCursorPos(); ///< Test the GetCursorPos call.
    void testGuiReady(); ///< Test the GuiReady call.
    void testQuit(); ///< Test the Quit call.
    void testRestart(); ///< Test the Restart call.
    void testShowOnStartup(); ///< Test the ShowOnStartup calls.
    void testShowSplashScreen(); ///< Test the ShowSplashScreen calls.
    void testDockIconVisible(); ///< Test the DockIconVisible calls.
    void testIsFirstGuiStart(); ///< Test the IsFirstGuiStart calls.
    void testIsAutostartOn(); ///< Test the IsAutostartOn calls.
    void testIsBetaEnabled(); ///< Test the IsBetaEnabled calls.
    void testGoOs(); ///< Test the GoOs call.
    void testTriggerReset(); ///< Test the TriggerReset call.
    void testVersion(); ///< Test the Version call.
    void testLogPath(); ///< Test the LogPath call.
    void testLicensePath(); ///< Test the LicensePath call.
    void testReleaseNotesLink(); ///< Test the ReleastNotesLink call.
    void testLandingPageLink(); ///< Test the LandingPageLink call.
    void testColorScheme(); ///< Test the ColorScheme calls.
    void testCurrentEmailClient(); ///< Test the CurrentEmailClient calls.
    void testReportBug(); ///< Test the ReportBug call.
    void testLogin(); ///< Test the Login call.
    void testLogin2FA(); ///< Test the Login2FA call.
    void testLogin2Passwords(); ///< Test the Login2Passwords call.
    void testLoginAbort(); ///< Test the LoginAbort call.
    void testCheckUpdate(); ///< Test the CheckUpdate call.
    void testInstallUpdate(); ///< Test the CheckUpdate call.
    void testIsAutomaticUpdateOn(); ///< Test the IsAutomaticUpdateOn calls.
    void testIsCacheOnDiskEnabled(); ///< Test the IsCacheOnDiskEnabled calls.
    void testDiskCachePath(); ///< Test the DiskCachePath calls.
    void testChangeLocalCache(); ///< Test the ChangeLocalPath calls.
    void testIsDohEnabled(); ///< Test the IsDohEnabled calls.
    void testUseSslForSmtp(); ///< Test the UseSslForSmtp calls.
    void testHostname(); ///< Test the Hostname call.
    void testImapPort(); ///< Test the ImapPort calls.
    void testSmtpPort(); ///< Test the SmtpPort calls.
    void testChangePorts(); ///< Test the ChangePorts call.
    void testIsPortFree(); ///< Test the IsPortFree call.
    void testKeychain(); ///< Test the keychains related calls.
    void testUser(); ///< Test the user related calls.
    void testStream(); ///< Test the server to client stream

#pragma clang diagnostic pop

private: // data members
    std::shared_ptr<grpc::Channel> channel_ { nullptr }; ///< The gRPC channel.
    std::shared_ptr<bridgerpc::BridgeRpc::Stub> stub_ { nullptr }; ///< The gRPC stub (a.k.a. client).

};


#endif //BRIDGE_RPC_CLIENT_TEST_RPC_CLIENT_H
