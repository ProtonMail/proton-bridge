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
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/theme"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ErrNotImplemented = status.Errorf(codes.Unimplemented, "Not implemented")

// GuiReady implement the GuiReady gRPC service call.
func (s *Service) GuiReady(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("GuiReady")
	// Note nothing to be done. old Qt frontend had a sync.one
	return &emptypb.Empty{}, nil
}

// Quit implement the Quit gRPC service call.
func (s *Service) Quit(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Quit")
	var err error
	if s.eventStreamCh != nil {
		if _, err = s.StopEventStream(ctx, empty); err != nil {
			s.log.WithError(err).Error("Quit failed.")
		}
	}

	// The following call is launched as a goroutine, as it will wait for current calls to end, including this one.
	go func() { s.grpcServer.GracefulStop() }()

	return &emptypb.Empty{}, err
}

// Restart implement the Restart gRPC service call.
func (s *Service) Restart(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("Restart") // TO-DO-GODT-1671  handle restart.

	s.restart()

	return nil, ErrNotImplemented
}

func (s *Service) ShowOnStartup(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("ShowOnStartup")

	return wrapperspb.Bool(s.showOnStartup), nil
}

func (s *Service) ShowSplashScreen(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("ShowSplashScreen")

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

func (s *Service) IsFirstGuiStart(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsFirstGuiStart")

	return wrapperspb.Bool(s.settings.GetBool(settings.FirstStartGUIKey)), nil
}

func (s *Service) SetIsAutostartOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("show", isOn.Value).Info("SetIsAutostartOn")

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

func (s *Service) IsAutostartOn(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsAutostartOn")

	return wrapperspb.Bool(s.bridge.IsAutostartEnabled()), nil
}

func (s *Service) SetIsBetaEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isEnabled", isEnabled.Value).Info("SetIsBetaEnabled")

	channel := updater.StableChannel
	if isEnabled.Value {
		channel = updater.EarlyChannel
	}

	s.bridge.SetUpdateChannel(channel)
	s.checkUpdate()

	return &emptypb.Empty{}, nil
}

func (s *Service) IsBetaEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsBetaEnabled")

	return wrapperspb.Bool(s.bridge.GetUpdateChannel() == updater.EarlyChannel), nil
}

func (s *Service) GoOs(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("GoOs") // TO-DO We can probably get rid of this and use QSysInfo::product name
	return wrapperspb.String(runtime.GOOS), nil
}

func (s *Service) TriggerReset(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("TriggerReset")
	return nil, ErrNotImplemented
}

func (s *Service) Version(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("Version")
	return nil, ErrNotImplemented
}

func (s *Service) LogsPath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("LogsPath")
	path, err := s.locations.ProvideLogsPath()
	if err != nil {
		s.log.WithError(err).Error("Cannot determine logs path")
		return nil, err
	}
	return wrapperspb.String(path), nil
}

func (s *Service) LicensePath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("LicensePath")
	return wrapperspb.String(s.locations.GetLicenseFilePath()), nil
}

func (s *Service) DependencyLicensesLink(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(s.locations.GetDependencyLicensesLink()), nil
}

func (s *Service) SetColorSchemeName(_ context.Context, name *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("ColorSchemeName", name.Value).Info("SetColorSchemeName")

	if !theme.IsAvailable(theme.Theme(name.Value)) {
		s.log.WithField("scheme", name.Value).Warn("Color scheme not available")
		return nil, status.Error(codes.NotFound, "Color scheme not available")
	}

	s.settings.Set(settings.ColorScheme, name.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) ColorSchemeName(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("ColorSchemeName")

	current := s.settings.Get(settings.ColorScheme)
	if !theme.IsAvailable(theme.Theme(current)) {
		current = string(theme.DefaultTheme())
		s.settings.Set(settings.ColorScheme, current)
	}

	return wrapperspb.String(current), nil
}

func (s *Service) CurrentEmailClient(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("CurrentEmailClient")

	return wrapperspb.String(s.userAgent.String()), nil
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
	s.log.WithField("username", login.Username).Info("Login")
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

func (s *Service) Login2FA(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Info("Login2FA")

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

func (s *Service) Login2Passwords(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Info("Login2Passwords")

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

func (s *Service) LoginAbort(_ context.Context, loginAbort *LoginAbortRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", loginAbort.Username).Info("LoginAbort")
	go func() {
		defer s.panicHandler.HandlePanic()

		s.loginAbort()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckUpdate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("CheckUpdate")
	// TO-DO GODT-1670 Implement update check
	return &emptypb.Empty{}, nil
}

func (s *Service) InstallUpdate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("InstallUpdate")
	// TO-DO GODT-1670 Implement update install
	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsAutomaticUpdateOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isOn", isOn.Value).Info("SetIsAutomaticUpdateOn")

	currentlyOn := s.settings.GetBool(settings.AutoUpdateKey)
	if currentlyOn == isOn.Value {
		return &emptypb.Empty{}, nil
	}

	s.settings.SetBool(settings.AutoUpdateKey, isOn.Value)
	s.checkUpdateAndNotify()

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutomaticUpdateOn(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsAutomaticUpdateOn")

	return wrapperspb.Bool(s.settings.GetBool(settings.AutoUpdateKey)), nil
}

func (s *Service) IsCacheOnDiskEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsCacheOnDiskEnabled")
	return wrapperspb.Bool(s.settings.GetBool(settings.CacheEnabledKey)), nil
}

func (s *Service) DiskCachePath(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("DiskCachePath")
	return wrapperspb.String(s.settings.Get(settings.CacheLocationKey)), nil
}

func (s *Service) ChangeLocalCache(_ context.Context, change *ChangeLocalCacheRequest) (*emptypb.Empty, error) {
	s.log.WithField("enableDiskCache", change.EnableDiskCache).
		WithField("diskCachePath", change.DiskCachePath).
		Info("DiskCachePath")

	defer func() { _ = s.SendEvent(NewCacheChangeLocalCacheFinishedEvent()) }()
	defer func() { _ = s.SendEvent(NewIsCacheOnDiskEnabledChanged(s.settings.GetBool(settings.CacheEnabledKey))) }()
	defer func() { _ = s.SendEvent(NewDiskCachePathChanged(s.settings.Get(settings.CacheCompressionKey))) }()

	if change.EnableDiskCache != s.settings.GetBool(settings.CacheEnabledKey) {
		if change.EnableDiskCache {
			if err := s.bridge.EnableCache(); err != nil {
				s.log.WithError(err).Error("Cannot enable disk cache")
			}
		} else {
			if err := s.bridge.DisableCache(); err != nil {
				s.log.WithError(err).Error("Cannot disable disk cache")
			}
		}
	}

	path := change.DiskCachePath
	//goland:noinspection GoBoolExpressions
	if (runtime.GOOS == "windows") && (path[0] == '/') {
		path = path[1:]
	}

	if change.EnableDiskCache && path != s.settings.Get(settings.CacheLocationKey) {
		if err := s.bridge.MigrateCache(s.settings.Get(settings.CacheLocationKey), path); err != nil {
			s.log.WithError(err).Error("The local cache location could not be changed.")
			_ = s.SendEvent(NewCacheErrorEvent(CacheErrorType_CACHE_CANT_MOVE_ERROR))
			return &emptypb.Empty{}, nil
		}
		s.settings.Set(settings.CacheLocationKey, path)
	}

	_ = s.SendEvent(NewCacheLocationChangeSuccessEvent())
	s.restart()

	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsDoHEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	s.log.WithField("isEnabled", isEnabled.Value).Info("SetIsDohEnabled")

	s.bridge.SetProxyAllowed(isEnabled.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) IsDoHEnabled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsDohEnabled")

	return wrapperspb.Bool(s.bridge.GetProxyAllowed()), nil
}

func (s *Service) SetUseSslForSmtp(_ context.Context, useSsl *wrapperspb.BoolValue) (*emptypb.Empty, error) { //nolint:revive,stylecheck
	s.log.WithField("useSsl", useSsl.Value).Info("SetUseSslForSmtp")

	if s.settings.GetBool(settings.SMTPSSLKey) == useSsl.Value {
		return &emptypb.Empty{}, nil
	}

	defer func() { _ = s.SendEvent(NewMailSettingsUseSslForSmtpFinishedEvent()) }()

	s.settings.SetBool(settings.SMTPSSLKey, useSsl.Value)
	s.restart()

	return &emptypb.Empty{}, nil
}

func (s *Service) UseSslForSmtp(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) { //nolint:revive,stylecheck
	s.log.Info("UseSslForSmtp")

	return wrapperspb.Bool(s.settings.GetBool(settings.SMTPSSLKey)), nil
}

func (s *Service) Hostname(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("Hostname")

	return wrapperspb.String(bridge.Host), nil
}

func (s *Service) ImapPort(context.Context, *emptypb.Empty) (*wrapperspb.Int32Value, error) {
	s.log.Info("ImapPort")

	return wrapperspb.Int32(int32(s.settings.GetInt(settings.IMAPPortKey))), nil
}

func (s *Service) SmtpPort(context.Context, *emptypb.Empty) (*wrapperspb.Int32Value, error) { //nolint:revive,stylecheck
	s.log.Info("SmtpPort")

	return wrapperspb.Int32(int32(s.settings.GetInt(settings.SMTPPortKey))), nil
}

func (s *Service) ChangePorts(_ context.Context, ports *ChangePortsRequest) (*emptypb.Empty, error) {
	s.log.WithField("imapPort", ports.ImapPort).WithField("smtpPort", ports.SmtpPort).Info("ChangePorts")

	defer func() { _ = s.SendEvent(NewMailSettingsChangePortFinishedEvent()) }()

	s.settings.SetInt(settings.IMAPPortKey, int(ports.ImapPort))
	s.settings.SetInt(settings.SMTPPortKey, int(ports.SmtpPort))

	s.restart()
	return &emptypb.Empty{}, nil
}

func (s *Service) IsPortFree(_ context.Context, port *wrapperspb.Int32Value) (*wrapperspb.BoolValue, error) {
	s.log.Info("IsPortFree")
	return wrapperspb.Bool(ports.IsPortFree(int(port.Value))), nil
}

func (s *Service) AvailableKeychains(context.Context, *emptypb.Empty) (*AvailableKeychainsResponse, error) {
	s.log.Info("AvailableKeychains")

	keychains := make([]string, 0, len(keychain.Helpers))
	for chain := range keychain.Helpers {
		keychains = append(keychains, chain)
	}

	return &AvailableKeychainsResponse{Keychains: keychains}, nil
}

func (s *Service) SetCurrentKeychain(_ context.Context, keychain *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("keychain", keychain.Value).Info("SetCurrentKeyChain") // we do not check validity.
	defer func() { _ = s.SendEvent(NewKeychainChangeKeychainFinishedEvent()) }()

	if s.bridge.GetKeychainApp() == keychain.Value {
		return &emptypb.Empty{}, nil
	}

	s.bridge.SetKeychainApp(keychain.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) CurrentKeychain(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.log.Info("CurrentKeychain")

	return wrapperspb.String(s.bridge.GetKeychainApp()), nil
}
