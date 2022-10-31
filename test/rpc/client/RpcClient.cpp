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


#pragma clang diagnostic push
#pragma ide diagnostic ignored "OCUnusedGlobalDeclarationInspection"


#include "RpcClient.h"
#include <QtTest>


using namespace google::protobuf;
using namespace grpc;


namespace
{


// \todo Decide where to store this certificate.
std::string cert = R"(-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIJAMUQK0VGexMsMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0yMjA2MTQxNjUyNTVaFw0yMjA3MTQxNjUyNTVaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAL6T1JQ0jptq512PBLASpCLFB0px7KIzEml0oMUCkVgUF+2cayrvdBXJZnaO
SG+/JPnHDcQ/ecgqkh2Ii6a2x2kWA5KqWiV+bSHp0drXyUGJfM85muLsnrhYwJ83
HHtweoUVebRZvHn66KjaH8nBJ+YVWyYbSUhJezcg6nBSEtkW+I/XUHu4S2C7FUc5
DXPO3yWWZuZ22OZz70DY3uYE/9COuilotuKdj7XgeKDyKIvRXjPFyqGxwnnp6bXC
vWvrQdcxy0wM+vZxew3QtA/Ag9uKJU9owP6noauXw95l49lEVIA5KXVNtdaldVht
MO/QoelLZC7h79PK22zbii3x930CAwEAAaM6MDgwFAYDVR0RBA0wC4IJbG9jYWxo
b3N0MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATANBgkqhkiG9w0B
AQsFAAOCAQEAW/9PE8dcAN+0C3K96Xd6Y3qOOtQhRw+WlZXhtiqMtlJfTjvuGKs9
58xuKcTvU5oobxLv+i5+4gpqLjUZZ9FBnYXZIACNVzq4PEXf+YdzcA+y6RS/rqT4
dUjsuYrScAmdXK03Duw3HWYrTp8gsJzIaYGTltUrOn0E4k/TsZb/tZ6z+oH7Fi+p
wdsI6Ut6Zwm3Z7WLn5DDk8KvFjHjZkdsCb82SFSAUVrzWo5EtbLIY/7y3A5rGp9D
t0AVpuGPo5Vn+MW1WA9HT8lhjz0v5wKGMOBi3VYW+Yx8FWHDpacvbZwVM0MjMSAd
M7SXYbNDiLF4LwPLsunoLsW133Ky7s99MA==
-----END CERTIFICATE-----)";

Empty empty; // A protobuf empty message reused for the sake of simplicity
}

//**********************************************************************************************************************
/// \brief Help class that manage a client context. Provided context cannot be re-used, the underlying context is freed
/// and re-allocated at every call to get() (or at destruction).
//**********************************************************************************************************************
class Ctx
{
public: // member functions
    Ctx()
    {} ///< Default constructor.
    Ctx(Ctx const &) = delete; ///< Disabled copy-constructor.
    Ctx(Ctx &&) = delete; ///< Disabled assignment copy-constructor.
    ~Ctx() = default; ///< Destructor.
    Ctx &operator=(Ctx const &) = delete; ///< Disabled assignment operator.
    Ctx &operator=(Ctx &&) = delete; ///< Disabled move assignment operator.
    ClientContext *get()
    {
        ctx_ = std::make_unique<ClientContext>();
        return ctx_.get();
    } ///< Release the previous context, if any and allocate a new one
private: // data members
    std::unique_ptr<ClientContext> ctx_{nullptr};
};

//**********************************************************************************************************************
//
//**********************************************************************************************************************
RpcClient::RpcClient()
    : QObject(nullptr)
{
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::initTestCase()
{
    SslCredentialsOptions opts;
    opts.pem_root_certs += cert;

    channel_ = CreateChannel("localhost:9292", grpc::SslCredentials(opts));
    QVERIFY2(channel_, "Channel creation failed.");

    stub_ = Bridge::NewStub(channel_);
    QVERIFY2(stub_, "Stub creation failed.");

    QVERIFY2(channel_->WaitForConnected(gpr_time_add(gpr_now(GPR_CLOCK_REALTIME),
        gpr_time_from_seconds(10, GPR_TIMESPAN))), "Connection to the RPC server failed.");

    QVERIFY2(channel_ && stub_ && (channel_->GetState(true) == GRPC_CHANNEL_READY), "connection check failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testGetCursorPos()
{
    Ctx ctx;
    PointResponse resp;
    Status s = stub_->GetCursorPos(ctx.get(), empty, &resp);
    QVERIFY2(s.ok(), "GetCursorPos failed.");
    QVERIFY2(resp.x() == 100, "Invalid x value");
    QVERIFY2(resp.y() == 200, "Invalid y value");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testGuiReady()
{
    Ctx ctx;
    QVERIFY2(stub_->GuiReady(ctx.get(), empty, &empty).ok(), "GuiReady failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testQuit()
{
    Ctx ctx;
    QVERIFY2(stub_->Quit(ctx.get(), empty, &empty).ok(), "Quit failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testRestart()
{
    Ctx ctx;
    QVERIFY2(stub_->Restart(ctx.get(), empty, &empty).ok(), "Restart failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testShowOnStartup()
{
    Ctx ctx;
    BoolValue showOnStartup;
    QVERIFY2(stub_->ShowOnStartup(ctx.get(), empty, &showOnStartup).ok(), "ShowOnStartup failed.");
    bool const newValue = !showOnStartup.value();

    showOnStartup.set_value(newValue);
    QVERIFY2(stub_->SetShowOnStartup(ctx.get(), showOnStartup, &empty).ok(), "SetShowOnStartup failed.");

    QVERIFY2(stub_->ShowOnStartup(ctx.get(), empty, &showOnStartup).ok(), "ShowOnStartup failed.");
    QVERIFY2(showOnStartup.value() == newValue, "ShowOnStartup failed readback.");
}

//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testShowSplashScreen()
{
    Ctx ctx;
    BoolValue showSplashScreen;
    QVERIFY2(stub_->ShowSplashScreen(ctx.get(), empty, &showSplashScreen).ok(), "ShowSplashScreen failed.");
    bool const newValue = !showSplashScreen.value();

    showSplashScreen.set_value(newValue);
    QVERIFY2(stub_->SetShowSplashScreen(ctx.get(), showSplashScreen, &empty).ok(), "SetShowSplashScreen failed.");

    QVERIFY2(stub_->ShowOnStartup(ctx.get(), empty, &showSplashScreen).ok(), "ShowSplashScreen failed.");
    QVERIFY2(showSplashScreen.value() == newValue, "ShowSplashScreen failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testDockIconVisible()
{
    Ctx ctx;
    BoolValue dockIconVisible;
    QVERIFY2(stub_->DockIconVisible(ctx.get(), empty, &dockIconVisible).ok(), "DockIconVisible failed.");
    bool const newValue = !dockIconVisible.value();

    dockIconVisible.set_value(newValue);
    QVERIFY2(stub_->SetDockIconVisible(ctx.get(), dockIconVisible, &empty).ok(), "SetShowSplashScreen failed.");

    QVERIFY2(stub_->DockIconVisible(ctx.get(), empty, &dockIconVisible).ok(), "DockIconVisible failed.");
    QVERIFY2(dockIconVisible.value() == newValue, "DockIconVisible failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsFirstGuiStart()
{
    Ctx ctx;
    BoolValue isFirst;
    QVERIFY2(stub_->IsFirstGuiStart(ctx.get(), empty, &isFirst).ok(), "IsFirstGuiStart failed.");
    bool const newValue = !isFirst.value();

    isFirst.set_value(newValue);
    QVERIFY2(stub_->SetIsFirstGuiStart(ctx.get(), isFirst, &empty).ok(), "SetIsFirstGuiStart failed.");

    QVERIFY2(stub_->IsFirstGuiStart(ctx.get(), empty, &isFirst).ok(), "IsFirstGuiStart failed.");
    QVERIFY2(isFirst.value() == newValue, "IsFirstGuiStart failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsAutostartOn()
{
    Ctx ctx;
    BoolValue isOn;
    QVERIFY2(stub_->IsAutostartOn(ctx.get(), empty, &isOn).ok(), "IsAutostartOn failed.");
    bool const newValue = !isOn.value();

    isOn.set_value(newValue);
    QVERIFY2(stub_->SetIsAutostartOn(ctx.get(), isOn, &empty).ok(), "SetIsAutostartOn failed.");

    QVERIFY2(stub_->IsAutostartOn(ctx.get(), empty, &isOn).ok(), "IsAutostartOn failed.");
    QVERIFY2(isOn.value() == newValue, "IsAutostartOn failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsBetaEnabled()
{
    Ctx ctx;
    BoolValue isOn;
    QVERIFY2(stub_->IsBetaEnabled(ctx.get(), empty, &isOn).ok(), "IsBetaEnabled failed.");
    bool const newValue = !isOn.value();

    isOn.set_value(newValue);
    QVERIFY2(stub_->SetIsBetaEnabled(ctx.get(), isOn, &empty).ok(), "SetIsBetaEnabled failed.");

    QVERIFY2(stub_->IsBetaEnabled(ctx.get(), empty, &isOn).ok(), "IsBetaEnabled failed.");
    QVERIFY2(isOn.value() == newValue, "IsBetaEnabled failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testGoOs()
{
    Ctx ctx;
    StringValue goos;
    QVERIFY2(stub_->GoOs(ctx.get(), empty, &goos).ok(), "GoOs failed.");
    QVERIFY2(goos.value().length() > 0, "Invalid GoOs value.");
}

//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testTriggerReset()
{
    Ctx ctx;
    QVERIFY2(stub_->TriggerReset(ctx.get(), empty, &empty).ok(), "TriggerReset failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testVersion()
{
    Ctx ctx;
    StringValue version;
    QVERIFY2(stub_->Version(ctx.get(), empty, &version).ok(), "Version failed.");
    QVERIFY2(version.value().length() > 0, "Invalid version number.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLogPath()
{
    Ctx ctx;
    StringValue logPath;
    QVERIFY2(stub_->LogPath(ctx.get(), empty, &logPath).ok(), "LogPath failed.");
    QVERIFY2(logPath.value().length() > 0, "Invalid LogPath.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLicensePath()
{
    Ctx ctx;
    StringValue licensePath;
    QVERIFY2(stub_->LicensePath(ctx.get(), empty, &licensePath).ok(), "LicensePath failed.");
    QVERIFY2(licensePath.value().length() > 0, "Invalid LicensePath.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testReleaseNotesLink()
{
    Ctx ctx;
    StringValue link;
    QVERIFY2(stub_->ReleaseNotesLink(ctx.get(), empty, &link).ok(), "ReleaseNotesLink failed.");
    QVERIFY2(link.value().length() > 0, "Invalid ReleaseNotesLink.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLandingPageLink()
{
    Ctx ctx;
    StringValue link;
    QVERIFY2(stub_->LandingPageLink(ctx.get(), empty, &link).ok(), "LandingPageLink failed.");
    QVERIFY2(link.value().length() > 0, "Invalid LandingPageLink.");
}


//**********************************************************************************************************************
// 
//**********************************************************************************************************************
void RpcClient::testColorScheme()
{
    Ctx ctx;
    StringValue name;
    string const schemeName = "dummyColors";
    name.set_value(schemeName);
    QVERIFY2(stub_->SetColorSchemeName(ctx.get(), name, &empty).ok(), "SetColorSchemeName failed.");

    QVERIFY2(stub_->ColorSchemeName(ctx.get(), empty, &name).ok(), "ColorSchemeName failed.");
    QVERIFY2(name.value() == schemeName, "ColorSchemeName failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testCurrentEmailClient()
{
    Ctx ctx;
    StringValue name;
    string const clientName = "dummyClient";
    name.set_value(clientName);
    QVERIFY2(stub_->SetCurrentEmailClient(ctx.get(), name, &empty).ok(), "SeturrentEmailClient failed.");

    QVERIFY2(stub_->CurrentEmailClient(ctx.get(), empty, &name).ok(), "CurrentEmailClient failed.");
    QVERIFY2(name.value() == clientName, "CurrentEmailClient failed readback.");
}


//**********************************************************************************************************************
// 
//**********************************************************************************************************************
void RpcClient::testReportBug()
{
    Ctx ctx;
    ReportBugRequest report;
    report.set_description("dummy description");
    report.set_address("dummy@proton.me");
    report.set_emailclient("dummyClient");
    report.set_includelogs(true);
    QVERIFY2(stub_->ReportBug(ctx.get(), report, &empty).ok(), "ReportBug failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLogin()
{
    Ctx ctx;
    LoginRequest login;
    login.set_username("dummyuser");
    login.set_password("dummypassword");
    QVERIFY2(stub_->Login(ctx.get(), login, &empty).ok(), "Login failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLogin2FA()
{
    Ctx ctx;
    LoginRequest login;
    login.set_username("dummyuser");
    login.set_password("dummypassword");
    QVERIFY2(stub_->Login2FA(ctx.get(), login, &empty).ok(), "Login2FA failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLogin2Passwords()
{
    Ctx ctx;
    LoginRequest login;
    login.set_username("dummyuser");
    login.set_password("dummypassword");
    QVERIFY2(stub_->Login2Passwords(ctx.get(), login, &empty).ok(), "Login2Passwords failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testLoginAbort()
{
    Ctx ctx;
    LoginAbortRequest loginAbort;
    loginAbort.set_username("dummyuser");
    QVERIFY2(stub_->LoginAbort(ctx.get(), loginAbort, &empty).ok(), "loginAbort failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testCheckUpdate()
{
    Ctx ctx;
    QVERIFY2(stub_->CheckUpdate(ctx.get(), empty, &empty).ok(), "CheckUpdate failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testInstallUpdate()
{
    Ctx ctx;
    QVERIFY2(stub_->InstallUpdate(ctx.get(), empty, &empty).ok(), "InstallUpdate failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsAutomaticUpdateOn()
{
    Ctx ctx;
    BoolValue isOn;
    QVERIFY2(stub_->IsAutomaticUpdateOn(ctx.get(), empty, &isOn).ok(), "IsAutomaticUpdateOn failed.");
    bool newValue = !isOn.value();

    isOn.set_value(newValue);
    QVERIFY2(stub_->SetIsAutomaticUpdateOn(ctx.get(), isOn, &empty).ok(), "SetIsAutomaticUpdateOn failed.");

    QVERIFY2(stub_->IsAutomaticUpdateOn(ctx.get(), empty, &isOn).ok(), "IsAutomaticUpdateOn failed.");
    QVERIFY2(isOn.value() == newValue, "IsAutomaticUpdateOn failed readback.");
}


//**********************************************************************************************************************
// 
//**********************************************************************************************************************
void RpcClient::testIsCacheOnDiskEnabled()
{
    Ctx ctx;
    BoolValue isEnabled;
    QVERIFY2(stub_->IsCacheOnDiskEnabled(ctx.get(), empty, &isEnabled).ok(), "IsCacheOnDiskEnabled failed.");
    bool const newValue = !isEnabled.value();

    isEnabled.set_value(newValue);
    QVERIFY2(stub_->SetIsCacheOnDiskEnabled(ctx.get(), isEnabled, &empty).ok(), "SetIsCacheOnDiskEnabled failed.");

    QVERIFY2(stub_->IsCacheOnDiskEnabled(ctx.get(), empty, &isEnabled).ok(), "IsCacheOnDiskEnabled failed.");
    QVERIFY2(isEnabled.value() == newValue, "IsCacheOnDiskEnabled failed readback.");
}


//**********************************************************************************************************************
// 
//**********************************************************************************************************************
void RpcClient::testDiskCachePath()
{
    Ctx ctx;
    string const dummyPath = "/dummy/path";
    StringValue path;
    path.set_value(dummyPath);
    QVERIFY2(stub_->SetDiskCachePath(ctx.get(), path, &empty).ok(), "SetDiskCachePath failed.");

    QVERIFY2(stub_->DiskCachePath(ctx.get(), empty, &path).ok(), "DiskCachePath failed.");
    QVERIFY2(path.value() == dummyPath, "DiskCachePath failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testChangeLocalCache()
{
    Ctx ctx;
    BoolValue isEnabled;
    QVERIFY2(stub_->IsCacheOnDiskEnabled(ctx.get(), empty, &isEnabled).ok(), "IsCacheOnDiskEnabled failed.");
    bool const newEnabled = !isEnabled.value();

    string const dummyPath = "/another/dummy/path";
    ChangeLocalCacheRequest request;
    request.set_enablediskcache(newEnabled);
    request.set_diskcachepath(dummyPath);
    QVERIFY2(stub_->ChangeLocalCache(ctx.get(), request, &empty).ok(), "ChangeLocalCache failed.");

    QVERIFY2(stub_->IsCacheOnDiskEnabled(ctx.get(), empty, &isEnabled).ok(), "IsCacheOnDiskEnabled failed.");
    QVERIFY2(isEnabled.value() == newEnabled, "IsCacheOnDiskEnabled readback failed.");

    StringValue path;
    QVERIFY2(stub_->DiskCachePath(ctx.get(), empty, &path).ok(), "DiskCachePath failed.");
    QVERIFY2(path.value() == dummyPath, "DiskCachePath failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsDohEnabled()
{
    Ctx ctx;
    BoolValue isEnabled;
    QVERIFY2(stub_->IsDoHEnabled(ctx.get(), empty, &isEnabled).ok(), "IsDoHEnabled failed.");
    bool const newValue = !isEnabled.value();

    isEnabled.set_value(newValue);
    QVERIFY2(stub_->SetIsDoHEnabled(ctx.get(), isEnabled, &empty).ok(), "SetIsDoHEnabled failed.");

    QVERIFY2(stub_->IsDoHEnabled(ctx.get(), empty, &isEnabled).ok(), "IsDoHEnabled failed.");
    QVERIFY2(isEnabled.value() == newValue, "IsDoHEnabled failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testUseSslForSmtp()
{
    Ctx ctx;
    BoolValue useSsl;
    QVERIFY2(stub_->UseSslForSmtp(ctx.get(), empty, &useSsl).ok(), "UseSslForSmtp failed.");
    bool const newValue = !useSsl.value();

    useSsl.set_value(newValue);
    QVERIFY2(stub_->SetUseSslForSmtp(ctx.get(), useSsl, &empty).ok(), "SetUseSslForSmtp failed.");

    QVERIFY2(stub_->UseSslForSmtp(ctx.get(), empty, &useSsl).ok(), "UseSslForSmtp failed.");
    QVERIFY2(useSsl.value() == newValue, "UseSslForSmtp failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testHostname()
{
    Ctx ctx;
    StringValue hostname;
    QVERIFY2(stub_->Hostname(ctx.get(), empty, &hostname).ok(), "Hostname failed.");
    QVERIFY2(hostname.value().length() > 0, "Hostname failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testImapPort()
{
    Ctx ctx;
    Int32Value port;
    QVERIFY2(stub_->ImapPort(ctx.get(), empty, &port).ok(), "ImapPort failed.");
    qint16 const newValue = qint16(port.value()) + 1;

    port.set_value(qint32(newValue));
    QVERIFY2(stub_->SetImapPort(ctx.get(), port, &empty).ok(), "SetImapPort failed.");

    QVERIFY2(stub_->ImapPort(ctx.get(), empty, &port).ok(), "ImapPort failed.");
    QVERIFY2(qint16(port.value()) == newValue, "ImapPort failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testSmtpPort()
{
    Ctx ctx;
    Int32Value port;
    QVERIFY2(stub_->SmtpPort(ctx.get(), empty, &port).ok(), "SmtpPort failed.");
    qint16 const newValue = qint16(port.value()) + 1;

    port.set_value(qint32(newValue));
    QVERIFY2(stub_->SetSmtpPort(ctx.get(), port, &empty).ok(), "SetSmtpPort failed.");

    QVERIFY2(stub_->SmtpPort(ctx.get(), empty, &port).ok(), "SmtpPort failed.");
    QVERIFY2(qint16(port.value()) == newValue, "SmtpPort failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testChangePorts()
{
    Ctx ctx;
    ChangePortsRequest request;
    const qint32 imapPort = 2143, smtpPort = 2025;
    request.set_imapport(imapPort);
    request.set_smtpport(smtpPort);
    QVERIFY2(stub_->ChangePorts(ctx.get(), request, &empty).ok(), "");

    Int32Value port;
    QVERIFY2(stub_->ImapPort(ctx.get(), empty, &port).ok(), "ImapPort failed.");
    QVERIFY2(port.value() == imapPort, "ImapPort failed readback.");

    QVERIFY2(stub_->SmtpPort(ctx.get(), empty, &port).ok(), "SmtpPort failed.");
    QVERIFY2(port.value() == smtpPort, "SmtpPort failed readback.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testIsPortFree()
{
    Ctx ctx;
    Int32Value port;
    port.set_value(143);
    BoolValue isFree;

    QVERIFY2(stub_->IsPortFree(ctx.get(), port, &isFree).ok(), "IsPortFree failed.");
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void RpcClient::testKeychain()
{
    Ctx ctx;
    AvailableKeychainsResponse resp;
    QVERIFY2(stub_->AvailableKeychains(ctx.get(), empty, &resp).ok(), "AvailableKeychains failed.");
    QVERIFY2(resp.keychains().size() > 0, "AvailableKeychains returned an invalid result.");

    string const newKeychain = resp.keychains().at(resp.keychains_size() - 1);

    StringValue keychain;
    keychain.set_value(newKeychain);
    QVERIFY2(stub_->SetCurrentKeychain(ctx.get(), keychain, &empty).ok(), "SetCurrentKeychain failed.");

    QVERIFY2(stub_->CurrentKeychain(ctx.get(), empty, &keychain).ok(), "CurrentKeychain failed.");
    QVERIFY2(newKeychain == keychain.value(), "CurrentKeychain readback failed.");
}


//**********************************************************************************************************************
// 
//**********************************************************************************************************************
void RpcClient::testUser()
{
    Ctx ctx;
    UserListResponse users;
    QVERIFY2(stub_->GetUserList(ctx.get(), empty, &users).ok(), "GetUserList failed.");
    QVERIFY2(users.users_size() > 0, "GetUserList returned an invalid value.");

    std::string const userID = users.users(0).id();

    UserSplitModeRequest splitModeRequest;
    splitModeRequest.set_userid(userID);
    splitModeRequest.set_active(true);
    QVERIFY2(stub_->SetUserSplitMode(ctx.get(), splitModeRequest, &empty).ok(), "SetUserSplitMode failed");

    ConfigureAppleMailRequest appleMailRequest;
    appleMailRequest.set_userid(userID);
    appleMailRequest.set_address("dummy@proton.ch");
    QVERIFY2(stub_->ConfigureUserAppleMail(ctx.get(), appleMailRequest, &empty).ok(), "ConfigureUserAppleMail failed.");

    StringValue stringValue;
    stringValue.set_value(userID);
    QVERIFY2(stub_->LogoutUser(ctx.get(), stringValue, &empty).ok(), "LogoutUser failed.");

    QVERIFY2(stub_->RemoveUser(ctx.get(), stringValue, &empty).ok(), "RemoveUser failed.");
}

void checkAppEvents(ClientReader<StreamEvent>& reader)
{
    QList<AppEvent::EventCase> expected = {
        AppEvent::kInternetStatus,
        AppEvent::kAutostartFinished,
        AppEvent::kResetFinished,
        AppEvent::kReportBugFinished,
        AppEvent::kReportBugSuccess,
        AppEvent::kReportBugError,
        AppEvent::kShowMainWindow,
    };
    StreamEvent event;
    while (reader.Read(&event)) {
        QVERIFY2(event.event_case() == StreamEvent::kApp, "Received invalid event while waiting for app event.");
        AppEvent const& appEvent = event.app();
        AppEvent::EventCase const eventCase = appEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected app event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected app event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for app events.");
}


void checkLoginEvents(ClientReader<StreamEvent>& reader)
{
    QList<LoginEvent::EventCase> expected = {
        LoginEvent::kError,
        LoginEvent::kTfaRequested,
        LoginEvent::kTwoPasswordRequested,
        LoginEvent::kFinished,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kLogin, "Received invalid event while waiting for login event.");
        LoginEvent const& loginEvent = event.login();
        LoginEvent::EventCase const eventCase = loginEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected login event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected login event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for login events.");
}


void checkUpdateEvents(ClientReader<StreamEvent>& reader)
{
    QList<UpdateEvent::EventCase> expected = {
        UpdateEvent::kError,
        UpdateEvent::kManualReady,
        UpdateEvent::kManualRestartNeeded,
        UpdateEvent::kForce,
        UpdateEvent::kSilentRestartNeeded,
        UpdateEvent::kIsLatestVersion,
        UpdateEvent::kCheckFinished
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kUpdate, "Received invalid event while waiting for update event.");
        UpdateEvent const& updateEvent = event.update();
        UpdateEvent::EventCase const eventCase = updateEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected update event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected update event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for update events.");
}


void checkCacheEvents(ClientReader<StreamEvent>& reader)
{
    QList<CacheEvent::EventCase> expected = {
        CacheEvent::kError,
        CacheEvent::kLocationChangedSuccess,
        CacheEvent::kChangeLocalCacheFinished,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kCache, "Received invalid event while waiting for cache event.");
        CacheEvent const& cacheEvent = event.cache();
        CacheEvent::EventCase const eventCase = cacheEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected cache event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected cache event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for cache events.");
}


void checkMailsSettingsEvents(ClientReader<StreamEvent>& reader)
{
    QList<MailSettingsEvent::EventCase> expected = {
        MailSettingsEvent::kError,
        MailSettingsEvent::kUseSslForSmtpFinished,
        MailSettingsEvent::kChangePortsFinished,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kMailSettings, "Received invalid event while waiting for mail settings event.");
        MailSettingsEvent const& mailSettingsEvent = event.mailsettings();
        MailSettingsEvent::EventCase const eventCase = mailSettingsEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected mail settings event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected mail settings event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for mail settings events.");
}


void checkKeychainEvents(ClientReader<StreamEvent>& reader)
{
    QList<KeychainEvent::EventCase> expected = {
        KeychainEvent::kChangeKeychainFinished,
        KeychainEvent::kHasNoKeychain,
        KeychainEvent::kRebuildKeychain,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kKeychain, "Received invalid event while waiting for keychain event.");
        KeychainEvent const& keychainEvent = event.keychain();
        KeychainEvent::EventCase const eventCase = keychainEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected keychain event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected keychain event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for keychain events.");
}


void checkMailEvents(ClientReader<StreamEvent>& reader)
{
    QList<MailEvent::EventCase> expected = {
        MailEvent::kNoActiveKeyForRecipientEvent,
        MailEvent::kAddressChanged,
        MailEvent::kAddressChangedLogout,
        MailEvent::kApiCertIssue,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kMail, "Received invalid event while waiting for mail event.");
        MailEvent const& mailEvent = event.mail();
        MailEvent::EventCase const eventCase = mailEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected mail event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected mail event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for mail events.");
}


void checkUserEvents(ClientReader<StreamEvent>& reader)
{
    QList<UserEvent::EventCase> expected = {
        UserEvent::kToggleSplitModeFinished,
        UserEvent::kUserDisconnected,
        UserEvent::kUserChanged,
    };

    StreamEvent event;
    while (reader.Read(&event))
    {
        QVERIFY2(event.event_case() == StreamEvent::kUser, "Received invalid event while waiting for user event.");
        UserEvent const& userEvent = event.user();
        UserEvent::EventCase const eventCase = userEvent.event_case();
        QVERIFY2(expected.size() > 0, "Empty expected user event list.");
        QVERIFY2(eventCase == expected.front(), "Unexpected user event received.");
        expected.pop_front();
        if (expected.isEmpty())
            return;
    }
    QFAIL("Stream ended while waiting for user events.");
}


void RpcClient::testStream()
{
    Ctx ctx;
    std::unique_ptr<ClientReader<StreamEvent>> reader = stub_->GetEvents(ctx.get(), empty);
    QVERIFY2(reader, "Could not instanciate event stream reader");

    checkAppEvents(*reader);
    checkLoginEvents(*reader);
    checkUpdateEvents(*reader);
    checkCacheEvents(*reader);
    checkMailsSettingsEvents(*reader);
    checkKeychainEvents(*reader);
    checkMailEvents(*reader);
    checkUserEvents(*reader);
}


#pragma clang diagnostic pop
