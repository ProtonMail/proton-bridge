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

package rpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Service is the RPC service struct.
type Service struct {
	UnimplementedBridgeRpcServer
	grpcServer *grpc.Server
	log        *logrus.Entry

	showOnStartup        bool
	showSplashScreen     bool
	dockIconVisible      bool
	isFirstGuiStart      bool
	isAutostartOn        bool
	isBetaEnabled        bool
	colorSchemeName      string
	currentEmailClient   string
	isAutoUpdateOn       bool
	isCacheOnDiskEnabled bool
	diskCachePath        string
	isDohEnabled         bool
	useSSLForSMTP        bool
	hostname             string
	imapPort             uint16
	smtpPort             uint16
	keychains            []string
	currentKeychain      string
	users                []*User
	currentUser          string
}

// NewService returns a new instance of the service.
func NewService(grpcServer *grpc.Server, log *logrus.Entry) *Service {
	service := Service{
		grpcServer:         grpcServer,
		log:                log,
		colorSchemeName:    "aName",
		currentEmailClient: "aClient",
		hostname:           "dummy.proton.me",
		imapPort:           143,
		smtpPort:           25,
		keychains:          []string{uuid.New().String(), uuid.New().String(), uuid.New().String()},
		users: []*User{{
			Id:             uuid.New().String(),
			Username:       "user1",
			AvatarText:     "avatarText1",
			LoggedIn:       true,
			SplitMode:      false,
			SetupGuideSeen: true,
			UsedBytes:      5000000000,
			TotalBytes:     1000000000,
			Password:       "dummyPassword",
			Addresses:      []string{"dummy@proton.me"},
		}, {
			Id:             uuid.New().String(),
			Username:       "user2",
			AvatarText:     "avatarText2",
			LoggedIn:       false,
			SplitMode:      true,
			SetupGuideSeen: false,
			UsedBytes:      4000000000,
			TotalBytes:     2000000000,
			Password:       "dummyPassword2",
			Addresses:      []string{"dummy2@proton.me"},
		}},
	}
	service.currentKeychain = service.keychains[0]
	service.currentUser = service.users[0].Id

	return &service
}

func (s *Service) GetCursorPos(context.Context, *emptypb.Empty) (*PointResponse, error) {
	s.log.Info("GetCursorPos")
	return &PointResponse{X: 100, Y: 200}, nil
}

func (s *Service) GuiReady(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("GuiReady")
	return &emptypb.Empty{}, nil
}
func (s *Service) Quit(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Quit")
	return &emptypb.Empty{}, nil
}

func (s *Service) Restart(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Restart")
	return &emptypb.Empty{}, nil
}
func (s *Service) SetShowOnStartup(_ context.Context, show *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", show.Value).Info("SetShowOnStartup")
	s.showOnStartup = show.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) ShowOnStartup(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("ShowOnStartup")
	return wrapperspb.Bool(s.showOnStartup), nil
}

func (s *Service) SetShowSplashScreen(_ context.Context, show *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", show.Value).Info("SetShowSplashScreen")
	s.showSplashScreen = show.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) ShowSplashScreen(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("ShowSplashScreen")
	return wrapperspb.Bool(s.showSplashScreen), nil
}

func (s *Service) SetDockIconVisible(_ context.Context, visible *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", visible.Value).Info("SetDockIconVisible")
	s.dockIconVisible = visible.Value
	return &emptypb.Empty{}, nil
}
func (s *Service) DockIconVisible(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("DockIconVisible")
	return wrapperspb.Bool(s.dockIconVisible), nil
}
func (s *Service) SetIsFirstGuiStart(_ context.Context, isFirst *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", isFirst.Value).Info("SetIsFirstGuiStart")
	s.isFirstGuiStart = isFirst.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsFirstGuiStart(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsFirstGuiStart")
	return wrapperspb.Bool(s.isFirstGuiStart), nil
}

func (s *Service) SetIsAutostartOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", isOn.Value).Info("SetIsAutostartOn")
	s.isAutostartOn = isOn.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutostartOn(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsAutostartOn")
	return wrapperspb.Bool(s.isAutostartOn), nil
}

func (s *Service) SetIsBetaEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", isEnabled.Value).Info("SetIsBetaEnabled")
	s.isBetaEnabled = isEnabled.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsBetaEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsBetaEnabled")
	return wrapperspb.Bool(s.isBetaEnabled), nil
}

func (s *Service) GoOs(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("GoOs")
	return wrapperspb.String("DummyOsName"), nil
}

func (s *Service) TriggerReset(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("TriggerReset")
	return &emptypb.Empty{}, nil
}

func (s *Service) Version(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("Version")
	return wrapperspb.String("1.0"), nil
}

func (s *Service) LogPath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("LogPath")
	return wrapperspb.String("/path/to/log"), nil
}

func (s *Service) LicensePath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("LicensePath")
	return wrapperspb.String("/path/to/license"), nil
}

func (s *Service) ReleaseNotesLink(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("ReleaseNotesLink")
	return wrapperspb.String("https//proton.me/release/notes.html"), nil
}

func (s *Service) LandingPageLink(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("LandingPageLink")
	return wrapperspb.String("https//proton.me"), nil
}

func (s *Service) SetColorSchemeName(_ context.Context, name *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("ColorSchemeName", name.Value).Info("SetColorSchemeName")
	s.colorSchemeName = name.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) ColorSchemeName(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("ColorSchemeName")
	return wrapperspb.String(s.colorSchemeName), nil
}

func (s *Service) SetCurrentEmailClient(_ context.Context, client *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("CurrentEmailClient", client.Value).Info("SetCurrentEmailClient")
	s.currentEmailClient = client.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) CurrentEmailClient(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("CurrentEmailClient")
	return wrapperspb.String(s.currentEmailClient), nil
}

func (s *Service) ReportBug(_ context.Context, report *ReportBugRequest) (*emptypb.Empty, error) {
	s.log.WithField("description", report.Description).
		WithField("address", report.Address).
		WithField("emailClient", report.EmailClient).
		WithField("includeLogs", report.IncludeLogs).
		Info("ReportBug")
	return &emptypb.Empty{}, nil
}

func (s *Service) Login(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.
		WithField("username", login.Username).
		Info("Login")
	return &emptypb.Empty{}, nil
}

func (s *Service) Login2FA(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.
		WithField("username", login.Username).
		Info("Login2FA")
	return &emptypb.Empty{}, nil
}

func (s *Service) Login2Passwords(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.
		WithField("username", login.Username).
		Info("Login2Passwords")
	return &emptypb.Empty{}, nil
}

func (s *Service) LoginAbort(_ context.Context, loginAbort *LoginAbortRequest) (*emptypb.Empty, error) {
	s.log.
		WithField("username", loginAbort.Username).
		Info("LoginAbort")
	return &emptypb.Empty{}, nil
}

func (s *Service) CheckUpdate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("CheckUpdate")
	return &emptypb.Empty{}, nil
}

func (s *Service) InstallUpdate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("InstallUpdate")
	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsAutomaticUpdateOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isOn", isOn.Value).Info("SetIsAutomaticUpdateOn")
	s.isAutoUpdateOn = isOn.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutomaticUpdateOn(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsAutomaticUpdateOn")
	return wrapperspb.Bool(s.isAutoUpdateOn), nil
}

func (s *Service) SetIsCacheOnDiskEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isOn", isEnabled.Value).Info("SetIsCacheOnDiskEnabled")
	s.isCacheOnDiskEnabled = isEnabled.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsCacheOnDiskEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsCacheOnDiskEnabled")
	return wrapperspb.Bool(s.isCacheOnDiskEnabled), nil
}

func (s *Service) SetDiskCachePath(_ context.Context, path *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("path", path.Value).Info("IsCacheOnDiskEnabled")
	s.diskCachePath = path.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) DiskCachePath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("DiskCachePath")
	return wrapperspb.String(s.diskCachePath), nil
}

func (s *Service) ChangeLocalCache(_ context.Context, change *ChangeLocalCacheRequest) (*emptypb.Empty, error) {
	s.log.
		WithField("enableDiskCache", change.EnableDiskCache).
		WithField("diskCachePath", change.DiskCachePath).
		Info("DiskCachePath")
	s.isCacheOnDiskEnabled = change.EnableDiskCache
	s.diskCachePath = change.DiskCachePath
	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsDoHEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isEnabled", isEnabled.Value).Info("SetIsDohEnabled")
	s.isDohEnabled = isEnabled.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) IsDoHEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsDohEnabled")
	return wrapperspb.Bool(s.isDohEnabled), nil
}
func (s *Service) SetUseSslForSmtp(_ context.Context, useSsl *wrapperspb.BoolValue) (*emptypb.Empty, error) { //nolint:revive,stylecheck
	s.log.WithField("useSsl", useSsl.Value).Info("SetUseSslForSmtp")
	s.useSSLForSMTP = useSsl.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) UseSslForSmtp(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) { //nolint:revive,stylecheck
	s.log.Info("UseSslForSmtp")
	return wrapperspb.Bool(s.useSSLForSMTP), nil
}

func (s *Service) Hostname(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("Hostname")
	return wrapperspb.String(s.hostname), nil
}

func (s *Service) SetImapPort(_ context.Context, port *wrapperspb.Int32Value) (*emptypb.Empty, error) {
	s.log.WithField("port", port.Value).Info("SetImapPort")
	s.imapPort = uint16(port.Value)
	return &emptypb.Empty{}, nil
}

func (s *Service) ImapPort(context.Context, *emptypb.Empty) (*wrapperspb.Int32Value, error) {
	s.log.Info("ImapPort")
	return wrapperspb.Int32(int32(s.imapPort)), nil
}

func (s *Service) SetSmtpPort(_ context.Context, port *wrapperspb.Int32Value) (*emptypb.Empty, error) { //nolint:revive,stylecheck
	s.log.WithField("port", port.Value).Info("SetSmtpPort")
	s.smtpPort = uint16(port.Value)
	return &emptypb.Empty{}, nil
}

func (s *Service) SmtpPort(context.Context, *emptypb.Empty) (*wrapperspb.Int32Value, error) { //nolint:revive,stylecheck
	s.log.Info("SmtpPort")
	return wrapperspb.Int32(int32(s.smtpPort)), nil
}

func (s *Service) ChangePorts(_ context.Context, ports *ChangePortsRequest) (*emptypb.Empty, error) {
	s.log.WithField("imapPort", ports.ImapPort).WithField("smtpPort", ports.SmtpPort).Info("ChangePorts")
	s.imapPort = uint16(ports.ImapPort)
	s.smtpPort = uint16(ports.SmtpPort)
	return &emptypb.Empty{}, nil
}

func (s *Service) IsPortFree(context.Context, *wrapperspb.Int32Value) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsPortFree")
	return wrapperspb.Bool(true), nil
}

func (s *Service) AvailableKeychains(context.Context, *emptypb.Empty) (*AvailableKeychainsResponse, error) {
	s.log.Info("AvailableKeychains")
	return &AvailableKeychainsResponse{Keychains: s.keychains}, nil
}

func (s *Service) SetCurrentKeychain(_ context.Context, keychain *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("keychain", keychain.Value).Info("SetCurrentKeyChain") // we do not check validity.
	s.currentKeychain = keychain.Value
	return &emptypb.Empty{}, nil
}

func (s *Service) CurrentKeychain(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("CurrentKeychain")
	return wrapperspb.String(s.currentKeychain), nil
}

func (s *Service) GetUserList(context.Context, *emptypb.Empty) (*UserListResponse, error) {
	s.log.Info("GetUserList")
	return &UserListResponse{Users: s.users}, nil
}
func (s *Service) GetUser(context.Context, *emptypb.Empty) (*User, error) {
	s.log.Info("GetUser")
	return s.getCurrentUser(), nil
}
func (s *Service) SetUserSplitMode(_ context.Context, splitMode *UserSplitModeRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", splitMode.UserID).WithField("Active", splitMode.Active).Info("SetUserSplitMode")
	user := s.findUser(splitMode.UserID) // we should return an error
	if user != nil {
		user.SplitMode = splitMode.Active
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) LogoutUser(_ context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Info("LogoutUser")
	user := s.findUser(userID.Value)
	if user != nil {
		user.LoggedIn = false
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveUser(_ context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Info("RemoveUser")
	// we actually do nothing
	return &emptypb.Empty{}, nil
}

func (s *Service) ConfigureUserAppleMail(_ context.Context, request *ConfigureAppleMailRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", request.UserID).WithField("Address", request.Address).Info("ConfigureUserAppleMail")
	return &emptypb.Empty{}, nil
}

func (s *Service) GetEvents(_ *emptypb.Empty, server BridgeRpc_GetEventsServer) error { // nolint:funlen
	s.log.Info("Starting Event stream")

	events := []func() *StreamEvent{
		// app
		internetStatusEvent,
		autostartFinishedEvent,
		resetFinishedEvent,
		reportBugFinishedEvent,
		reportBugSuccessEvent,
		reportBugErrorEvent,
		showMainWindowEvent,

		// login
		loginError,
		loginTfaRequestedEvent,
		loginTwoPasswordsRequestedEvent,
		loginFinishedEvent,

		// update
		updateErrorEvent,
		updateManualReadyEvent,
		updateManualRestartNeededEvent,
		updateForceEvent,
		updateSilentRestartNeededEvent,
		updateIsLatestVersionEvent,
		updateCheckFinishedEvent,

		// cache
		cacheErrorEvent,
		cacheLocationChangeSuccessEvent,
		cacheChangeLocalCacheFinishedEvent,

		// mail settings
		mailSettingsErrorEvent,
		mailSettingsUseSslForSmtpFinishedEvent,
		mailSettingsChangePortFinishedEvent,

		// keychain
		keychainChangeKeychainFinishedEvent,
		keychainHasNoKeychainEvent,
		keychainRebuildKeychainEvent,

		// mail
		mailNoActiveKeyForRecipientEvent,
		mailAddressChangeEvent,
		mailAddressChangeLogoutEvent,
		mailApiCertIssue,

		// user
		userToggleSplitModeFinishedEvent,
		userDisconnectedEvent,
		userChangedEvent,
	}

	for _, eventFunc := range events {
		event := eventFunc()
		s.log.WithField("event", event).Info("Sending event")
		if err := server.Send(eventFunc()); err != nil {
			return err
		}
	}

	s.log.Info("Stop Event stream")

	return nil
}

func (s *Service) getCurrentUser() *User {
	return s.findUser(s.currentUser)
}

func (s *Service) findUser(userID string) *User {
	for _, u := range s.users {
		if u.Id == userID {
			return u
		}
	}

	return nil
}

func appEvent(appEvent *AppEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_App{App: appEvent}}
}

func internetStatusEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_InternetStatus{InternetStatus: &InternetStatusEvent{Connected: true}}})
}

func autostartFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_AutostartFinished{AutostartFinished: &AutostartFinishedEvent{}}})
}

func resetFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ResetFinished{ResetFinished: &ResetFinishedEvent{}}})
}

func reportBugFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugFinished{ReportBugFinished: &ReportBugFinishedEvent{}}})
}

func reportBugSuccessEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugSuccess{ReportBugSuccess: &ReportBugSuccessEvent{}}})
}

func reportBugErrorEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugError{ReportBugError: &ReportBugErrorEvent{}}})
}

func showMainWindowEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ShowMainWindow{ShowMainWindow: &ShowMainWindowEvent{}}})
}

func loginEvent(event *LoginEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Login{Login: event}}
}

func loginError() *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_Error{Error: &LoginErrorEvent{Type: LoginErrorType_FREE_USER}}})
}

func loginTfaRequestedEvent() *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_TfaRequested{TfaRequested: &LoginTfaRequestedEvent{Username: "dummy@proton.me"}}})
}

func loginTwoPasswordsRequestedEvent() *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_TwoPasswordRequested{}})
}

func loginFinishedEvent() *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_Finished{Finished: &LoginFinishedEvent{WasAlreadyLoggedIn: true}}})
}

func updateEvent(event *UpdateEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Update{Update: event}}
}

func updateErrorEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_Error{Error: &UpdateErrorEvent{Type: UpdateErrorType_UPDATE_SILENT_ERROR}}})
}

func updateManualReadyEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_ManualReady{ManualReady: &UpdateManualReadyEvent{Version: "1.0"}}})
}

func updateManualRestartNeededEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_ManualRestartNeeded{ManualRestartNeeded: &UpdateManualRestartNeededEvent{}}})
}

func updateForceEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_Force{Force: &UpdateForceEvent{Version: " 2.0"}}})
}

func updateSilentRestartNeededEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_SilentRestartNeeded{SilentRestartNeeded: &UpdateSilentRestartNeeded{}}})
}

func updateIsLatestVersionEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_IsLatestVersion{IsLatestVersion: &UpdateIsLatestVersion{}}})
}

func updateCheckFinishedEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_CheckFinished{CheckFinished: &UpdateCheckFinished{}}})
}

func cacheEvent(event *CacheEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Cache{Cache: event}}
}

func cacheErrorEvent() *StreamEvent {
	return cacheEvent(&CacheEvent{Event: &CacheEvent_Error{Error: &CacheErrorEvent{Type: CacheErrorType_CACHE_UNAVAILABLE_ERROR}}})
}

func cacheLocationChangeSuccessEvent() *StreamEvent {
	return cacheEvent(&CacheEvent{Event: &CacheEvent_LocationChangedSuccess{LocationChangedSuccess: &CacheLocationChangeSuccessEvent{}}})
}

func cacheChangeLocalCacheFinishedEvent() *StreamEvent {
	return cacheEvent(&CacheEvent{Event: &CacheEvent_ChangeLocalCacheFinished{ChangeLocalCacheFinished: &ChangeLocalCacheFinishedEvent{}}})
}

func mailSettingsEvent(event *MailSettingsEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_MailSettings{MailSettings: event}}
}

func mailSettingsErrorEvent() *StreamEvent {
	return mailSettingsEvent(&MailSettingsEvent{Event: &MailSettingsEvent_Error{Error: &MailSettingsErrorEvent{Type: MailSettingsErrorType_IMAP_PORT_ISSUE}}})
}

func mailSettingsUseSslForSmtpFinishedEvent() *StreamEvent { //nolint:revive,stylecheck
	return mailSettingsEvent(&MailSettingsEvent{Event: &MailSettingsEvent_UseSslForSmtpFinished{UseSslForSmtpFinished: &UseSslForSmtpFinishedEvent{}}})
}

func mailSettingsChangePortFinishedEvent() *StreamEvent {
	return mailSettingsEvent(&MailSettingsEvent{Event: &MailSettingsEvent_ChangePortsFinished{ChangePortsFinished: &ChangePortsFinishedEvent{}}})
}

func keychainEvent(event *KeychainEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Keychain{Keychain: event}}
}

func keychainChangeKeychainFinishedEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_ChangeKeychainFinished{ChangeKeychainFinished: &ChangeKeychainFinishedEvent{}}})
}

func keychainHasNoKeychainEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_HasNoKeychain{HasNoKeychain: &HasNoKeychainEvent{}}})
}

func keychainRebuildKeychainEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_RebuildKeychain{RebuildKeychain: &RebuildKeychainEvent{}}})
}

func mailEvent(event *MailEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Mail{Mail: event}}
}

func mailNoActiveKeyForRecipientEvent() *StreamEvent {
	return mailEvent(&MailEvent{Event: &MailEvent_NoActiveKeyForRecipientEvent{NoActiveKeyForRecipientEvent: &NoActiveKeyForRecipientEvent{Email: "dummy@proton.me"}}})
}

func mailAddressChangeEvent() *StreamEvent {
	return mailEvent(&MailEvent{Event: &MailEvent_AddressChanged{AddressChanged: &AddressChangedEvent{Address: "dummy@proton.me"}}})
}

func mailAddressChangeLogoutEvent() *StreamEvent {
	return mailEvent(&MailEvent{Event: &MailEvent_AddressChangedLogout{AddressChangedLogout: &AddressChangedLogoutEvent{Address: "dummy@proton.me"}}})
}

func mailApiCertIssue() *StreamEvent { //nolint:revive,stylecheck
	return mailEvent(&MailEvent{Event: &MailEvent_ApiCertIssue{ApiCertIssue: &ApiCertIssueEvent{}}})
}

func userEvent(event *UserEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_User{User: event}}
}

func userToggleSplitModeFinishedEvent() *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_ToggleSplitModeFinished{ToggleSplitModeFinished: &ToggleSplitModeFinishedEvent{UserID: "userID"}}})
}

func userDisconnectedEvent() *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UserDisconnected{UserDisconnected: &UserDisconnectedEvent{Username: "dummy@proton.me"}}})
}

func userChangedEvent() *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UserChanged{UserChanged: &UserChangedEvent{User: &User{}}}})
}
