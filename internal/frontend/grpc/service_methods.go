// Copyright (c) 2024 Proton AG
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
	"errors"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/frontend/theme"
	"github.com/ProtonMail/proton-bridge/v3/internal/kb"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/service"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/pkg/ports"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CheckTokens implements the CheckToken gRPC service call.
func (s *Service) CheckTokens(_ context.Context, clientConfigPath *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("CheckTokens")

	path := clientConfigPath.Value
	logEntry := s.log.WithField("path", path)

	var clientConfig service.Config
	if err := clientConfig.Load(path); err != nil {
		logEntry.WithError(err).Error("Could not read gRPC client config file")

		return nil, err
	}

	logEntry.Info("gRPC client config file was successfully loaded")

	return &wrapperspb.StringValue{Value: clientConfig.Token}, nil
}

func (s *Service) AddLogEntry(_ context.Context, request *AddLogEntryRequest) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
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
func (s *Service) GuiReady(_ context.Context, _ *emptypb.Empty) (*GuiReadyResponse, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("GuiReady")

	s.initializationDone.Do(s.initializing.Done)

	// Splash screen should be displayed only to users who start v3.0.17 or later for the first time after upgrading from v2.
	return &GuiReadyResponse{
		ShowSplashScreen: (!s.bridge.GetFirstStart()) &&
			s.bridge.GetLastVersion().LessThan(semver.MustParse("3.0.0")) &&
			s.bridge.GetCurrentVersion().GreaterThan(semver.MustParse("3.0.16")),
	}, nil
}

// Quit implement the Quit gRPC service call.
func (s *Service) Quit(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("Quit")
	s.quit()
	return &emptypb.Empty{}, nil
}

func (s *Service) quit() {
	// Windows is notably slow at Quitting. We do it in a goroutine to speed things up a bit.
	go func() {
		defer async.HandlePanic(s.panicHandler)

		if s.parentPID >= 0 {
			s.parentPIDDoneCh <- struct{}{}
		}

		var err error
		if s.isStreamingEvents() {
			if err = s.stopEventStream(); err != nil {
				s.log.WithError(err).Error("Quit failed.")
			}
		}

		// The following call is launched as a goroutine, as it will wait for current calls to end, including this one.
		s.grpcServer.GracefulStop() // gRPC does clean up and remove the file socket if used.
	}()
}

// Restart implement the Restart gRPC service call.
func (s *Service) Restart(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("Restart")

	s.restarter.Set(true, false)
	return s.Quit(ctx, empty)
}

func (s *Service) ShowOnStartup(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("ShowOnStartup")

	return wrapperspb.Bool(s.showOnStartup), nil
}

func (s *Service) SetIsAutostartOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.WithField("show", isOn.Value).Debug("SetIsAutostartOn")

	defer func() { _ = s.SendEvent(NewToggleAutostartFinishedEvent()) }()

	if isOn.Value == s.bridge.GetAutostart() {
		s.initAutostart()
		return &emptypb.Empty{}, nil
	}

	s.initAutostart()

	if err := s.bridge.SetAutostart(isOn.Value); err != nil {
		s.log.WithField("makeItEnabled", isOn.Value).WithError(err).Error("Autostart change failed")
		return nil, status.Errorf(codes.Internal, "failed to set autostart: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutostartOn(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("IsAutostartOn")

	return wrapperspb.Bool(s.bridge.GetAutostart()), nil
}

func (s *Service) SetIsBetaEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.WithField("isEnabled", isEnabled.Value).Debug("SetIsBetaEnabled")

	channel := updater.StableChannel
	if isEnabled.Value {
		channel = updater.EarlyChannel
	}

	if err := s.bridge.SetUpdateChannel(channel); err != nil {
		s.log.WithError(err).Error("Failed to set update channel")
		return nil, status.Errorf(codes.Internal, "failed to set update channel: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsBetaEnabled(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("IsBetaEnabled")

	return wrapperspb.Bool(s.bridge.GetUpdateChannel() == updater.EarlyChannel), nil
}

func (s *Service) SetIsAllMailVisible(_ context.Context, isVisible *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.WithField("isVisible", isVisible.Value).Debug("SetIsAllMailVisible")

	if err := s.bridge.SetShowAllMail(isVisible.Value); err != nil {
		s.log.WithError(err).Error("Failed to set show all mail")
		return nil, status.Errorf(codes.Internal, "failed to set show all mail: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAllMailVisible(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("IsAllMailVisible")

	return wrapperspb.Bool(s.bridge.GetShowAllMail()), nil
}

func (s *Service) SetIsTelemetryDisabled(_ context.Context, isDisabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.WithField("isEnabled", isDisabled.Value).Debug("SetIsTelemetryDisabled")

	if err := s.bridge.SetTelemetryDisabled(isDisabled.Value); err != nil {
		s.log.WithError(err).Error("Failed to set telemetry status")
		return nil, status.Errorf(codes.Internal, "failed to set telemetry status: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsTelemetryDisabled(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("IsTelemetryDisabled")

	return wrapperspb.Bool(s.bridge.GetTelemetryDisabled()), nil
}

func (s *Service) GoOs(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("GoOs") // TO-DO We can probably get rid of this and use QSysInfo::product name

	return wrapperspb.String(runtime.GOOS), nil
}

func (s *Service) TriggerReset(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("TriggerReset")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		s.triggerReset()
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) Version(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("Version")

	return wrapperspb.String(s.bridge.GetCurrentVersion().Original()), nil
}

func (s *Service) LogsPath(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("LogsPath")

	path, err := s.bridge.GetLogsPath()
	if err != nil {
		s.log.WithError(err).Error("Cannot determine logs path")
		return nil, err
	}
	return wrapperspb.String(path), nil
}

func (s *Service) LicensePath(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	s.log.Debug("LicensePath")

	return wrapperspb.String(s.bridge.GetLicenseFilePath()), nil
}

func (s *Service) DependencyLicensesLink(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)
	return wrapperspb.String(s.bridge.GetDependencyLicensesLink()), nil
}

func (s *Service) ReleaseNotesPageLink(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.latestLock.RLock()
	defer func() {
		async.HandlePanic(s.panicHandler)
		s.latestLock.RUnlock()
	}()

	return wrapperspb.String(s.latest.ReleaseNotesPage), nil
}

func (s *Service) LandingPageLink(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	s.latestLock.RLock()
	defer func() {
		async.HandlePanic(s.panicHandler)
		s.latestLock.RUnlock()
	}()

	return wrapperspb.String(s.latest.LandingPage), nil
}

func (s *Service) SetColorSchemeName(_ context.Context, name *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("ColorSchemeName", name.Value).Debug("SetColorSchemeName")

	if !theme.IsAvailable(theme.Theme(name.Value)) {
		s.log.WithField("scheme", name.Value).Warn("Color scheme not available")
		return nil, status.Error(codes.NotFound, "Color scheme not available")
	}

	if err := s.bridge.SetColorScheme(name.Value); err != nil {
		s.log.WithError(err).Error("Failed to set color scheme")
		return nil, status.Errorf(codes.Internal, "failed to set color scheme: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) ColorSchemeName(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("ColorSchemeName")

	current := s.bridge.GetColorScheme()
	if !theme.IsAvailable(theme.Theme(current)) {
		current = string(theme.DefaultTheme())
		if err := s.bridge.SetColorScheme(current); err != nil {
			s.log.WithError(err).Error("Failed to set color scheme")
			return nil, status.Errorf(codes.Internal, "failed to set color scheme: %v", err)
		}
	}

	return wrapperspb.String(current), nil
}

func (s *Service) CurrentEmailClient(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("CurrentEmailClient")

	return wrapperspb.String(s.bridge.GetCurrentUserAgent()), nil
}

func (s *Service) ReportBug(_ context.Context, report *ReportBugRequest) (*emptypb.Empty, error) {
	s.log.WithFields(logrus.Fields{
		"osType":      report.OsType,
		"osVersion":   report.OsVersion,
		"title":       report.Title,
		"description": report.Description,
		"address":     report.Address,
		"emailClient": report.EmailClient,
		"includeLogs": report.IncludeLogs,
	}).Debug("ReportBug")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		defer func() { _ = s.SendEvent(NewReportBugFinishedEvent()) }()
		reportReq := bridge.ReportBugReq{
			OSType:      report.OsType,
			OSVersion:   report.OsVersion,
			Title:       report.Title,
			Description: report.Description,
			Username:    report.Address,
			Email:       report.Address,
			EmailClient: report.EmailClient,
			IncludeLogs: report.IncludeLogs,
		}
		if err := s.bridge.ReportBug(context.Background(), &reportReq); err != nil {
			s.log.WithError(err).Error("Failed to report bug")
			_ = s.SendEvent(NewReportBugErrorEvent())
			return
		}

		_ = s.SendEvent(NewReportBugSuccessEvent())
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) ForceLauncher(_ context.Context, launcher *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("launcher", launcher.Value).Debug("ForceLauncher")

	s.restarter.Override(launcher.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) SetMainExecutable(_ context.Context, exe *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("executable", exe.Value).Debug("SetMainExecutable")

	s.restarter.AddFlags("--wait", exe.Value)

	return &emptypb.Empty{}, nil
}

func (s *Service) RequestKnowledgeBaseSuggestions(_ context.Context, userInput *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.Debug("RequestKnowledgeBaseSuggestions")
	go func() {
		defer async.HandlePanic(s.panicHandler)

		articles, err := s.bridge.GetKnowledgeBaseSuggestions(userInput.Value)
		if err != nil {
			s.log.WithError(err).Error("Could not retrieve KB article suggestions")
			articles = kb.ArticleList{}
		}

		_ = s.SendEvent(NewRequestKnowledgeBaseSuggestionsEvent(articles))
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) Login(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login")

	var hvDetails *proton.APIHVDetails
	if login.UseHvDetails != nil && *login.UseHvDetails {
		hvDetails = s.hvDetails
		s.useHvDetails = true
	} else {
		s.useHvDetails = false
	}

	go func() {
		defer async.HandlePanic(s.panicHandler)

		s.twoPasswordAttemptCount = 0
		password, err := base64Decode(login.Password)
		if err != nil {
			s.log.WithError(err).Error("Cannot decode password")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode password"))
			return
		}

		client, auth, err := s.bridge.LoginAuth(context.Background(), login.Username, password, hvDetails)
		if err != nil {
			defer s.loginClean()

			if errors.Is(err, bridge.ErrUserAlreadyLoggedIn) {
				_ = s.SendEvent(NewLoginAlreadyLoggedInEvent(auth.UserID))
			} else if apiErr := new(proton.APIError); errors.As(err, &apiErr) {
				switch apiErr.Code { // nolint:exhaustive
				case proton.PasswordWrong, proton.UsernameInvalid:
					_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, ""))

				case proton.PaidPlanRequired:
					_ = s.SendEvent(NewLoginError(LoginErrorType_FREE_USER, ""))

				case proton.HumanVerificationRequired:
					s.handleHvRequest(apiErr)

				case proton.HumanValidationInvalidToken:
					s.hvDetails = nil
					_ = s.SendEvent(NewLoginError(LoginErrorType_HV_ERROR, err.Error()))

				default:
					_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, err.Error()))
				}
			} else {
				_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, err.Error()))
			}

			return
		}

		s.password = password
		s.authClient = client
		s.auth = auth

		switch {
		case auth.TwoFA.Enabled&proton.HasTOTP != 0:
			_ = s.SendEvent(NewLoginTfaRequestedEvent(login.Username))

		case auth.PasswordMode == proton.TwoPasswordMode:
			_ = s.SendEvent(NewLoginTwoPasswordsRequestedEvent(login.Username))

		default:
			s.finishLogin()
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) Login2FA(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login2FA")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		if s.auth.UID == "" || s.authClient == nil {
			s.log.Errorf("Login 2FA: authentication incomplete %s %p", s.auth.UID, s.authClient)
			_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ABORT, "Missing authentication, try again."))
			s.loginClean()
			return
		}

		twoFA, err := base64Decode(login.Password)
		if err != nil {
			s.log.WithError(err).Error("Cannot decode 2fa code")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode 2fa code"))
			s.loginClean()
			return
		}

		if err := s.authClient.Auth2FA(context.Background(), proton.Auth2FAReq{TwoFactorCode: string(twoFA)}); err != nil {
			if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Code == proton.PasswordWrong {
				s.log.Warn("Login 2FA: retry 2fa")
				_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ERROR, ""))
			} else {
				s.log.WithError(err).Warn("Login 2FA: failed")
				_ = s.SendEvent(NewLoginError(LoginErrorType_TFA_ABORT, err.Error()))
				s.loginClean()
			}

			return
		}

		if s.auth.PasswordMode == proton.TwoPasswordMode {
			_ = s.SendEvent(NewLoginTwoPasswordsRequestedEvent(login.Username))
			return
		}

		s.finishLogin()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) Login2Passwords(_ context.Context, login *LoginRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", login.Username).Debug("Login2Passwords")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		password, err := base64Decode(login.Password)
		if err != nil {
			s.log.WithError(err).Error("Cannot decode mbox password")
			_ = s.SendEvent(NewLoginError(LoginErrorType_USERNAME_PASSWORD_ERROR, "Cannot decode mbox password"))
			s.loginClean()
			return
		}

		s.password = password

		s.finishLogin()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) LoginAbort(_ context.Context, loginAbort *LoginAbortRequest) (*emptypb.Empty, error) {
	s.log.WithField("username", loginAbort.Username).Debug("LoginAbort")

	go func() {
		defer async.HandlePanic(s.panicHandler)
		s.loginAbort()
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckUpdate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("CheckUpdate")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		updateCh, done := s.bridge.GetEvents(
			events.UpdateAvailable{},
			events.UpdateNotAvailable{},
			events.UpdateCheckFailed{},
		)
		defer done()

		s.bridge.CheckForUpdates()

		switch (<-updateCh).(type) {
		case events.UpdateNotAvailable:
			_ = s.SendEvent(NewUpdateIsLatestVersionEvent())

		case events.UpdateAvailable:
			// ... this is handled by the main event loop

		case events.UpdateCheckFailed:
			// ... maybe show an error? but do nothing for now
		}

		_ = s.SendEvent(NewUpdateCheckFinishedEvent())
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) InstallUpdate(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("InstallUpdate")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		safe.RLock(func() {
			s.bridge.InstallUpdate(s.target)
		}, s.targetLock)
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsAutomaticUpdateOn(_ context.Context, isOn *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("isOn", isOn.Value).Debug("SetIsAutomaticUpdateOn")

	if currentlyOn := s.bridge.GetAutoUpdate(); currentlyOn == isOn.Value {
		return &emptypb.Empty{}, nil
	}

	if err := s.bridge.SetAutoUpdate(isOn.Value); err != nil {
		s.log.WithError(err).Error("Failed to set auto update")
		return nil, status.Errorf(codes.Internal, "failed to set auto update: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsAutomaticUpdateOn(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("IsAutomaticUpdateOn")

	return wrapperspb.Bool(s.bridge.GetAutoUpdate()), nil
}

func (s *Service) DiskCachePath(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("DiskCachePath")

	return wrapperspb.String(s.bridge.GetGluonCacheDir()), nil
}

func (s *Service) SetDiskCachePath(_ context.Context, newPath *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("path", newPath.Value).Debug("setDiskCachePath")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		defer func() {
			_ = s.SendEvent(NewDiskCachePathChangeFinishedEvent())
		}()

		path := newPath.Value

		//goland:noinspection GoBoolExpressions
		if (runtime.GOOS == "windows") && (path[0] == '/') {
			path = path[1:]
		}

		if path != s.bridge.GetGluonCacheDir() {
			if err := s.bridge.SetGluonDir(context.Background(), path); err != nil {
				s.log.WithError(err).Error("The local cache location could not be changed.")
				_ = s.SendEvent(NewDiskCacheErrorEvent(DiskCacheErrorType_CANT_MOVE_DISK_CACHE_ERROR))
				return
			}

			_ = s.SendEvent(NewDiskCachePathChangedEvent(s.bridge.GetGluonCacheDir()))
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) SetIsDoHEnabled(_ context.Context, isEnabled *wrapperspb.BoolValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("isEnabled", isEnabled.Value).Debug("SetIsDohEnabled")

	if err := s.bridge.SetProxyAllowed(isEnabled.Value); err != nil {
		s.log.WithError(err).Error("Failed to set DoH")
		return nil, status.Errorf(codes.Internal, "failed to set DoH: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) IsDoHEnabled(_ context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("IsDohEnabled")

	return wrapperspb.Bool(s.bridge.GetProxyAllowed()), nil
}

func (s *Service) MailServerSettings(_ context.Context, _ *emptypb.Empty) (*ImapSmtpSettings, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("ConnectionMode")

	return &ImapSmtpSettings{
		state:         protoimpl.MessageState{},
		sizeCache:     0,
		unknownFields: nil,
		ImapPort:      int32(s.bridge.GetIMAPPort()), //nolint:gosec // disable G115
		SmtpPort:      int32(s.bridge.GetSMTPPort()), //nolint:gosec // disable G115
		UseSSLForImap: s.bridge.GetIMAPSSL(),
		UseSSLForSmtp: s.bridge.GetSMTPSSL(),
	}, nil
}

func (s *Service) SetMailServerSettings(_ context.Context, settings *ImapSmtpSettings) (*emptypb.Empty, error) {
	s.log.
		WithField("ImapPort", settings.ImapPort).
		WithField("SmtpPort", settings.SmtpPort).
		WithField("UseSSUseSSLForIMAP", settings.UseSSLForImap).
		WithField("UseSSLForSMTP", settings.UseSSLForSmtp).
		Debug("SetConnectionMode")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		defer func() { _ = s.SendEvent(NewChangeMailServerSettingsFinishedEvent()) }()

		ctx := context.Background() // async operation, we cannot use the context provided by gRPC.

		if s.bridge.GetIMAPSSL() != settings.UseSSLForImap {
			if err := s.bridge.SetIMAPSSL(ctx, settings.UseSSLForImap); err != nil {
				s.log.WithError(err).Error("Failed to set IMAP SSL")
				_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_CONNECTION_MODE_CHANGE_ERROR))
			}
		}

		if s.bridge.GetSMTPSSL() != settings.UseSSLForSmtp {
			if err := s.bridge.SetSMTPSSL(ctx, settings.UseSSLForSmtp); err != nil {
				s.log.WithError(err).Error("Failed to set SMTP SSL")
				_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_CONNECTION_MODE_CHANGE_ERROR))
			}
		}

		if s.bridge.GetIMAPPort() != int(settings.ImapPort) {
			if err := s.bridge.SetIMAPPort(ctx, int(settings.ImapPort)); err != nil {
				s.log.WithError(err).Error("Failed to set IMAP port")
				_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_PORT_CHANGE_ERROR))
			}
		}

		if s.bridge.GetSMTPPort() != int(settings.SmtpPort) {
			if err := s.bridge.SetSMTPPort(ctx, int(settings.SmtpPort)); err != nil {
				s.log.WithError(err).Error("Failed to set SMTP port")
				_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_PORT_CHANGE_ERROR))
			}
		}

		_ = s.SendEvent(NewMailServerSettingsChangedEvent(s.getMailServerSettings()))
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) Hostname(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("Hostname")

	return wrapperspb.String(constants.Host), nil
}

func (s *Service) IsPortFree(_ context.Context, port *wrapperspb.Int32Value) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("IsPortFree")

	return wrapperspb.Bool(ports.IsPortFree(int(port.Value))), nil
}

func (s *Service) AvailableKeychains(_ context.Context, _ *emptypb.Empty) (*AvailableKeychainsResponse, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("AvailableKeychains")

	return &AvailableKeychainsResponse{Keychains: s.bridge.GetHelpersNames()}, nil
}

func (s *Service) SetCurrentKeychain(ctx context.Context, keychain *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.WithField("keychain", keychain.Value).Debug("SetCurrentKeyChain") // we do not check validity.

	defer func() { _, _ = s.Restart(ctx, &emptypb.Empty{}) }()
	defer func() { _ = s.SendEvent(NewKeychainChangeKeychainFinishedEvent()) }()

	helper, err := s.bridge.GetKeychainApp()
	if err != nil {
		s.log.WithError(err).Error("Failed to get current keychain")
		return nil, status.Errorf(codes.Internal, "failed to get current keychain: %v", err)
	}

	if helper == keychain.Value {
		return &emptypb.Empty{}, nil
	}

	if err := s.bridge.SetKeychainApp(keychain.Value); err != nil {
		s.log.WithError(err).Error("Failed to set keychain")
		return nil, status.Errorf(codes.Internal, "failed to set keychain: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CurrentKeychain(_ context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Debug("CurrentKeychain")

	helper, err := s.bridge.GetKeychainApp()
	if err != nil {
		s.log.WithError(err).Error("Failed to get current keychain")
		return nil, status.Errorf(codes.Internal, "failed to get current keychain: %v", err)
	}

	return wrapperspb.String(helper), nil
}

func (s *Service) TriggerRepair(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Debug("TriggerRepair")
	go func() {
		defer func() {
			async.HandlePanic(s.panicHandler)
			_ = s.SendEvent(NewRepairStartedEvent())
		}()

		s.bridge.Repair()
	}()

	return &emptypb.Empty{}, nil
}

func base64Decode(in []byte) ([]byte, error) {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(in)))

	n, err := base64.StdEncoding.Decode(out, in)
	if err != nil {
		return nil, err
	}

	return out[:n], nil
}

func (s *Service) getMailServerSettings() *ImapSmtpSettings {
	return &ImapSmtpSettings{
		ImapPort:      int32(s.bridge.GetIMAPPort()), //nolint:gosec // disable G115
		SmtpPort:      int32(s.bridge.GetSMTPPort()), //nolint:gosec // disable G115
		UseSSLForImap: s.bridge.GetIMAPSSL(),
		UseSSLForSmtp: s.bridge.GetSMTPSSL(),
	}
}
