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

package grpc

import (
	"context"
	"encoding/base64"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/theme"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/pkg/ports"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CheckTokens implements the CheckToken gRPC service call.
func (s *Service) CheckTokens(ctx context.Context, clientConfigPath *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	s.log.Debug("CheckTokens")

	path := clientConfigPath.Value
	logEntry := s.log.WithField("path", path)

	var clientConfig config
	if err := clientConfig.load(path); err != nil {
		logEntry.WithError(err).Error("Could not read gRPC client config file")

		return nil, err
	}

	logEntry.Info("gRPC client config file was successfully loaded")

	return &wrapperspb.StringValue{Value: clientConfig.Token}, nil
}

func (s *Service) AddLogEntry(ctx context.Context, request *AddLogEntryRequest) (*emptypb.Empty, error) {
	entry := s.log

	if len(request.Package) > 0 {
		entry = entry.WithField("pkg", request.Package)
	}

	level := logrusLevelFromGrpcLevel(request.Level)

	// we do a special case for Panic and Fatal as using logrus.Entry.Log will not panic nor exit respectively.
	if level == logrus.PanicLevel {
		entry.Panic(request.Message)

		return &emptypb.Empty{}, nil
	}

	if level == logrus.FatalLevel {
		entry.Fatal(request.Message)

		return &emptypb.Empty{}, nil
	}

	entry.Log(level, request.Message)

	return &emptypb.Empty{}, nil
}

// GuiReady implement the GuiReady gRPC service call.
func (s *Service) GuiReady(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("GuiReady")

	s.initializationDone.Do(s.initializing.Done)
	return &emptypb.Empty{}, nil
}

// Quit implement the Quit gRPC service call.
func (s *Service) Quit(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("Quit")

	// Windows is notably slow at Quitting. We do it in a goroutine to speed things up a bit.
	go func() {
		var err error
		if s.eventStreamCh != nil {
			if _, err = s.StopEventStream(ctx, empty); err != nil {
				s.log.WithError(err).Error("Quit failed.")
			}
		}

		// The following call is launched as a goroutine, as it will wait for current calls to end, including this one.
		s.grpcServer.GracefulStop()
	}()

	return &emptypb.Empty{}, nil
}

// Restart implement the Restart gRPC service call.
func (s *Service) Restart(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("Restart")

	s.restarter.SetToRestart()
	return s.Quit(ctx, empty)
}

func (s *Service) ShowOnStartup(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("ShowOnStartup")

	return wrapperspb.Bool(s.showOnStartup), nil
}

func (s *Service) ShowSplashScreen(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("ShowSplashScreen")

	if s.bridge.IsFirstStart() {
		return wrapperspb.Bool(false), nil
	}

	ver, err := semver.NewVersion(s.bridge.GetLastVersion())
	if err != nil {
		s.log.WithError(err).WithField("last", s.bridge.GetLastVersion()).Debug("Cannot parse last version")
		return wrapperspb.Bool(false), nil
	}

	// Current splash screen contains update on rebranding. Therefore, it
	// should be shown only if the last used version was less than 2.2.0.
	return wrapperspb.Bool(ver.LessThan(semver.MustParse("2.2.0"))), nil
}

func (s *Service) IsFirstGuiStart(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsFirstGuiStart")

	return wrapperspb.Bool(s.bridge.GetBool(settings.FirstStartGUIKey)), nil
}

func (s *Service) SetIsAutostartOn(ctx context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", isOn.Value).Debug("SetIsAutostartOn")

	defer func() { _ = s.SendEvent(NewToggleAutostartFinishedEvent()) }()

	if isOn.Value == s.bridge.IsAutostartEnabled() {
		s.initAutostart()
		return &emptypb.Empty{}, nil
	}

	var err error
	if isOn.Value {
		err = s.bridge.EnableAutostart()
	} else {
		err = s.bridge.DisableAutostart()
	}

	s.initAutostart()

	if err != nil {
		s.log.WithField("makeItEnabled", isOn.Value).WithError(err).Error("Autostart change failed")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutostartOn(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsAutostartOn")

	return wrapperspb.Bool(s.bridge.IsAutostartEnabled()), nil
}

func (s *Service) SetIsBetaEnabled(ctx context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isEnabled", isEnabled.Value).Debug("SetIsBetaEnabled")

	channel := updater.StableChannel
	if isEnabled.Value {
		channel = updater.EarlyChannel
	}

	s.bridge.SetUpdateChannel(channel)
	s.checkUpdate()

	return &emptypb.Empty{}, nil
}

func (s *Service) IsBetaEnabled(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsBetaEnabled")

	return wrapperspb.Bool(s.bridge.GetUpdateChannel() == updater.EarlyChannel), nil
}

func (s *Service) SetIsAllMailVisible(ctx context.Context, isVisible *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isVisible", isVisible.Value).Debug("SetIsAllMailVisible")

	s.bridge.SetIsAllMailVisible(isVisible.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAllMailVisible(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsAllMailVisible")

	return wrapperspb.Bool(s.bridge.IsAllMailVisible()), nil
}

func (s *Service) GoOs(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("GoOs") // TO-DO We can probably get rid of this and use QSysInfo::product name

	return wrapperspb.String(runtime.GOOS), nil
}

func (s *Service) TriggerReset(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("TriggerReset")

	go func() {
		defer s.panicHandler.HandlePanic()
		s.triggerReset()
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) Version(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("Version")

	return wrapperspb.String(constants.Version), nil
}

func (s *Service) LogsPath(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("LogsPath")

	path, err := s.bridge.ProvideLogsPath()
	if err != nil {
		s.log.WithError(err).Error("Cannot determine logs path")
		return nil, err
	}
	return wrapperspb.String(path), nil
}

func (s *Service) LicensePath(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("LicensePath")

	return wrapperspb.String(s.bridge.GetLicenseFilePath()), nil
}

func (s *Service) DependencyLicensesLink(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(s.bridge.GetDependencyLicensesLink()), nil
}

func (s *Service) ReleaseNotesPageLink(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(s.newVersionInfo.ReleaseNotesPage), nil
}

func (s *Service) LandingPageLink(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(s.newVersionInfo.LandingPage), nil
}

func (s *Service) SetColorSchemeName(ctx context.Context, name *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("ColorSchemeName", name.Value).Debug("SetColorSchemeName")

	if !theme.IsAvailable(theme.Theme(name.Value)) {
		s.log.WithField("scheme", name.Value).Warn("Color scheme not available")
		return nil, status.Error(codes.NotFound, "Color scheme not available")
	}

	s.bridge.Set(settings.ColorScheme, name.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) ColorSchemeName(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("ColorSchemeName")

	current := s.bridge.Get(settings.ColorScheme)
	if !theme.IsAvailable(theme.Theme(current)) {
		current = string(theme.DefaultTheme())
		s.bridge.Set(settings.ColorScheme, current)
	}

	return wrapperspb.String(current), nil
}

func (s *Service) CurrentEmailClient(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("CurrentEmailClient")

	return wrapperspb.String(s.bridge.GetCurrentUserAgent()), nil
}

func (s *Service) ReportBug(ctx context.Context, report *ReportBugRequest) (*emptypb.Empty, error) {
	s.log.WithFields(logrus.Fields{
		"osType":      report.OsType,
		"osVersion":   report.OsVersion,
		"description": report.Description,
		"address":     report.Address,
		"emailClient": report.EmailClient,
		"includeLogs": report.IncludeLogs,
	}).Debug("ReportBug")

	go func() {
		defer func() { _ = s.SendEvent(NewReportBugFinishedEvent()) }()

		if err := s.bridge.ReportBug(
			report.OsType,
			report.OsVersion,
			report.Description,
			report.Address,
			report.Address,
			report.EmailClient,
			report.IncludeLogs,
		); err != nil {
			s.log.WithError(err).Error("Failed to report bug")
			_ = s.SendEvent(NewReportBugErrorEvent())
			return
		}

		_ = s.SendEvent(NewReportBugSuccessEvent())
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) ForceLauncher(ctx context.Context, launcher *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("launcher", launcher.Value).Debug("ForceLauncher")

	go func() {
		defer s.panicHandler.HandlePanic()
		s.restarter.ForceLauncher(launcher.Value)
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) SetMainExecutable(ctx context.Context, exe *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("executable", exe.Value).Debug("SetMainExecutable")

	go func() {
		defer s.panicHandler.HandlePanic()
		s.restarter.SetMainExecutable(exe.Value)
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) Login(ctx context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login")

	go func() {
		defer s.panicHandler.HandlePanic()

		var err error
		s.password, err = base64.StdEncoding.DecodeString(login.Password)
		if err != nil {
			s.log.WithError(err).Error("Cannot decode password")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode password"))
			s.loginClean()
			return
		}

		s.authClient, s.auth, err = s.bridge.Login(login.Username, s.password)
		if err != nil {
			if err == pmapi.ErrPasswordWrong {
				// Remove error message since it is hardcoded in QML.
				_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, ""))
				s.loginClean()
				return
			}
			if err == pmapi.ErrPaidPlanRequired {
				_ = s.SendEvent(NewLoginError(LoginErrorType_FREE_USER, ""))
				s.loginClean()
				return
			}
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, err.Error()))
			s.loginClean()
			return
		}

		if s.auth.HasTwoFactor() {
			_ = s.SendEvent(NewLoginTfaRequestedEvent(login.Username))
			return
		}
		if s.auth.HasMailboxPassword() {
			_ = s.SendEvent(NewLoginTwoPasswordsRequestedEvent())
			return
		}

		s.finishLogin()
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) Login2FA(ctx context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login2FA")

	go func() {
		defer s.panicHandler.HandlePanic()

		if s.auth == nil || s.authClient == nil {
			s.log.Errorf("Login 2FA: authethication incomplete %p %p", s.auth, s.authClient)
			_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ABORT, "Missing authentication, try again."))
			s.loginClean()
			return
		}

		twoFA, err := base64.StdEncoding.DecodeString(login.Password)
		if err != nil {
			s.log.WithError(err).Error("Cannot decode 2fa code")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode 2fa code"))
			s.loginClean()
			return
		}

		err = s.authClient.Auth2FA(context.Background(), string(twoFA))
		if err == pmapi.ErrBad2FACodeTryAgain {
			s.log.Warn("Login 2FA: retry 2fa")
			_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ERROR, ""))
			return
		}

		if err == pmapi.ErrBad2FACode {
			s.log.Warn("Login 2FA: abort 2fa")
			_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ABORT, ""))
			s.loginClean()
			return
		}

		if err != nil {
			s.log.WithError(err).Warn("Login 2FA: failed.")
			_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ABORT, err.Error()))
			s.loginClean()
			return
		}

		if s.auth.HasMailboxPassword() {
			_ = s.SendEvent(NewLoginTwoPasswordsRequestedEvent())
			return
		}

		s.finishLogin()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) Login2Passwords(ctx context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login2Passwords")

	go func() {
		defer s.panicHandler.HandlePanic()

		var err error
		s.password, err = base64.StdEncoding.DecodeString(login.Password)

		if err != nil {
			s.log.WithError(err).Error("Cannot decode mbox password")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode mbox password"))
			s.loginClean()
			return
		}

		s.finishLogin()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) LoginAbort(ctx context.Context, loginAbort *LoginAbortRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", loginAbort.Username).Debug("LoginAbort")

	go func() {
		defer s.panicHandler.HandlePanic()

		s.loginAbort()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckUpdate(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("CheckUpdate")

	go func() {
		defer s.panicHandler.HandlePanic()

		s.checkUpdateAndNotify(true)
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) InstallUpdate(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("InstallUpdate")

	go func() {
		defer s.panicHandler.HandlePanic()

		s.installUpdate()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsAutomaticUpdateOn(ctx context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isOn", isOn.Value).Debug("SetIsAutomaticUpdateOn")

	currentlyOn := s.bridge.GetBool(settings.AutoUpdateKey)
	if currentlyOn == isOn.Value {
		return &emptypb.Empty{}, nil
	}

	s.bridge.SetBool(settings.AutoUpdateKey, isOn.Value)
	go func() {
		defer s.panicHandler.HandlePanic()

		s.checkUpdateAndNotify(false)
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutomaticUpdateOn(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsAutomaticUpdateOn")

	return wrapperspb.Bool(s.bridge.GetBool(settings.AutoUpdateKey)), nil
}

func (s *Service) IsCacheOnDiskEnabled(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsCacheOnDiskEnabled")

	return wrapperspb.Bool(s.bridge.GetBool(settings.CacheEnabledKey)), nil
}

func (s *Service) DiskCachePath(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("DiskCachePath")

	return wrapperspb.String(s.bridge.Get(settings.CacheLocationKey)), nil
}

func (s *Service) ChangeLocalCache(ctx context.Context, change *ChangeLocalCacheRequest) (*emptypb.Empty, error) {
	s.log.WithField("enableDiskCache", change.EnableDiskCache).
		WithField("diskCachePath", change.DiskCachePath).
		Debug("DiskCachePath")

	restart := false
	defer func(willRestart *bool) {
		_ = s.SendEvent(NewCacheChangeLocalCacheFinishedEvent(*willRestart))
		if *willRestart {
			_, _ = s.Restart(ctx, &emptypb.Empty{})
		}
	}(&restart)

	if change.EnableDiskCache != s.bridge.GetBool(settings.CacheEnabledKey) {
		if change.EnableDiskCache {
			if err := s.bridge.EnableCache(); err != nil {
				s.log.WithError(err).Error("Cannot enable disk cache")
			} else {
				restart = true
				_ = s.SendEvent(NewIsCacheOnDiskEnabledChanged(s.bridge.GetBool(settings.CacheEnabledKey)))
			}
		} else {
			if err := s.bridge.DisableCache(); err != nil {
				s.log.WithError(err).Error("Cannot disable disk cache")
			} else {
				restart = true
				_ = s.SendEvent(NewIsCacheOnDiskEnabledChanged(s.bridge.GetBool(settings.CacheEnabledKey)))
			}
		}
	}

	path := change.DiskCachePath
	//goland:noinspection GoBoolExpressions
	if (runtime.GOOS == "windows") && (path[0] == '/') {
		path = path[1:]
	}

	if change.EnableDiskCache && path != s.bridge.Get(settings.CacheLocationKey) {
		if err := s.bridge.MigrateCache(s.bridge.Get(settings.CacheLocationKey), path); err != nil {
			s.log.WithError(err).Error("The local cache location could not be changed.")
			_ = s.SendEvent(NewCacheErrorEvent(CacheErrorType_CACHE_CANT_MOVE_ERROR))
			return &emptypb.Empty{}, nil
		}

		s.bridge.Set(settings.CacheLocationKey, path)
		restart = true
		_ = s.SendEvent(NewDiskCachePathChanged(s.bridge.Get(settings.CacheLocationKey)))
	}

	_ = s.SendEvent(NewCacheLocationChangeSuccessEvent())

	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsDoHEnabled(ctx context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isEnabled", isEnabled.Value).Debug("SetIsDohEnabled")

	s.bridge.SetProxyAllowed(isEnabled.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) IsDoHEnabled(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsDohEnabled")

	return wrapperspb.Bool(s.bridge.GetProxyAllowed()), nil
}

func (s *Service) SetUseSslForSmtp(ctx context.Context, useSsl *wrapperspb.BoolValue) (*emptypb.Empty, error) { //nolint:revive,stylecheck
	s.log.WithField("useSsl", useSsl.Value).Debug("SetUseSslForSmtp")

	if s.bridge.GetBool(settings.SMTPSSLKey) == useSsl.Value {
		return &emptypb.Empty{}, nil
	}

	s.bridge.SetBool(settings.SMTPSSLKey, useSsl.Value)

	defer func() { _, _ = s.Restart(ctx, &emptypb.Empty{}) }()

	return &emptypb.Empty{}, s.SendEvent(NewMailSettingsUseSslForSmtpFinishedEvent())
}

func (s *Service) UseSslForSmtp(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) { //nolint:revive,stylecheck
	s.log.Debug("UseSslForSmtp")

	return wrapperspb.Bool(s.bridge.GetBool(settings.SMTPSSLKey)), nil
}

func (s *Service) Hostname(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("Hostname")

	return wrapperspb.String(bridge.Host), nil
}

func (s *Service) ImapPort(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.Int32Value, error) {
	s.log.Debug("ImapPort")

	return wrapperspb.Int32(int32(s.bridge.GetInt(settings.IMAPPortKey))), nil
}

func (s *Service) SmtpPort(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.Int32Value, error) { //nolint:revive,stylecheck
	s.log.Debug("SmtpPort")

	return wrapperspb.Int32(int32(s.bridge.GetInt(settings.SMTPPortKey))), nil
}

func (s *Service) ChangePorts(ctx context.Context, ports *ChangePortsRequest) (*emptypb.Empty, error) {
	s.log.WithField("imapPort", ports.ImapPort).WithField("smtpPort", ports.SmtpPort).Debug("ChangePorts")

	s.bridge.SetInt(settings.IMAPPortKey, int(ports.ImapPort))
	s.bridge.SetInt(settings.SMTPPortKey, int(ports.SmtpPort))

	defer func() { _, _ = s.Restart(ctx, &emptypb.Empty{}) }()

	return &emptypb.Empty{}, s.SendEvent(NewMailSettingsChangePortFinishedEvent())
}

func (s *Service) IsPortFree(ctx context.Context, port *wrapperspb.Int32Value) (*wrapperspb.BoolValue, error) {
	s.log.Debug("IsPortFree")

	return wrapperspb.Bool(ports.IsPortFree(int(port.Value))), nil
}

func (s *Service) AvailableKeychains(ctx context.Context, _ *emptypb.Empty) (*AvailableKeychainsResponse, error) {
	s.log.Debug("AvailableKeychains")

	keychains := make([]string, 0, len(keychain.Helpers))
	for chain := range keychain.Helpers {
		keychains = append(keychains, chain)
	}

	return &AvailableKeychainsResponse{Keychains: keychains}, nil
}

func (s *Service) SetCurrentKeychain(ctx context.Context, keychain *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("keychain", keychain.Value).Debug("SetCurrentKeyChain") // we do not check validity.

	defer func() { _, _ = s.Restart(ctx, &emptypb.Empty{}) }()
	defer func() { _ = s.SendEvent(NewKeychainChangeKeychainFinishedEvent()) }()

	if s.bridge.GetKeychainApp() == keychain.Value {
		return &emptypb.Empty{}, nil
	}

	s.bridge.SetKeychainApp(keychain.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) CurrentKeychain(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Debug("CurrentKeychain")

	return wrapperspb.String(s.bridge.GetKeychainApp()), nil
}
