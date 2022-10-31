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

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative bridge.proto

import (
	"context"
	cryptotls "crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	serverConfigFileName   = "grpcServerConfig.json"
	serverTokenMetadataKey = "server-token"
)

// Service is the RPC service struct.
type Service struct { // nolint:structcheck
	UnimplementedBridgeServer
	grpcServer        *grpc.Server //  the gGRPC server
	listener          net.Listener
	eventStreamCh     chan *StreamEvent
	eventStreamDoneCh chan struct{}
	eventQueue        []*StreamEvent
	eventQueueMutex   sync.Mutex

	panicHandler       types.PanicHandler
	eventListener      listener.Listener
	updater            types.Updater
	updateCheckMutex   sync.Mutex
	bridge             types.Bridger
	restarter          types.Restarter
	showOnStartup      bool
	authClient         pmapi.Client
	auth               *pmapi.Auth
	password           []byte
	newVersionInfo     updater.VersionInfo
	log                *logrus.Entry
	initializing       sync.WaitGroup
	initializationDone sync.Once
	firstTimeAutostart sync.Once
	locations          *locations.Locations
	token              string
	pemCert            string
}

// NewService returns a new instance of the service.
func NewService(
	showOnStartup bool,
	panicHandler types.PanicHandler,
	eventListener listener.Listener,
	updater types.Updater,
	restarter types.Restarter,
	locations *locations.Locations,
) *Service {
	s := Service{
		UnimplementedBridgeServer: UnimplementedBridgeServer{},
		panicHandler:              panicHandler,
		eventListener:             eventListener,
		updater:                   updater,
		restarter:                 restarter,
		showOnStartup:             showOnStartup,

		log:                logrus.WithField("pkg", "grpc"),
		initializing:       sync.WaitGroup{},
		initializationDone: sync.Once{},
		firstTimeAutostart: sync.Once{},
		locations:          locations,
		token:              uuid.NewString(),
	}

	// Initializing.Done is only called sync.Once. Please keep the increment
	// set to 1
	s.initializing.Add(1)

	go func() {
		defer s.panicHandler.HandlePanic()
		s.watchEvents()
	}()

	return &s
}

func (s *Service) startGRPCServer() {
	tlsConfig, pemCert, err := s.generateTLSConfig()
	if err != nil {
		s.log.WithError(err).Panic("Could not generate gRPC TLS config")
	}

	s.pemCert = string(pemCert)

	s.grpcServer = grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.UnaryInterceptor(s.validateUnaryServerToken),
		grpc.StreamInterceptor(s.validateStreamServerToken),
	)

	RegisterBridgeServer(s.grpcServer, s)

	s.listener, err = net.Listen("tcp", "127.0.0.1:0") // Port 0 means that the port is randomly picked by the system.
	if err != nil {
		s.log.WithError(err).Panic("Could not create gRPC listener")
	}

	if path, err := s.saveGRPCServerConfigFile(); err != nil {
		s.log.WithError(err).WithField("path", path).Panic("Could not write gRPC service config file")
	} else {
		s.log.WithField("path", path).Debug("Successfully saved gRPC service config file")
	}

	s.log.Info("gRPC server listening at ", s.listener.Addr())
}

func (s *Service) initAutostart() {
	// GODT-1507 Windows: autostart needs to be created after Qt is initialized.
	// GODT-1206: if preferences file says it should be on enable it here.

	// TO-DO GODT-1681 Autostart needs to be properly implement for gRPC approach.

	s.firstTimeAutostart.Do(func() {
		shouldAutostartBeOn := s.bridge.GetBool(settings.AutostartKey)
		if s.bridge.IsFirstStart() || shouldAutostartBeOn {
			if err := s.bridge.EnableAutostart(); err != nil {
				s.log.WithField("prefs", shouldAutostartBeOn).WithError(err).Error("Failed to enable first autostart")
			}
			return
		}
	})
}

func (s *Service) Loop(b types.Bridger) error {
	s.bridge = b
	s.initAutostart()
	s.startGRPCServer()

	defer func() {
		s.bridge.SetBool(settings.FirstStartGUIKey, false)
	}()

	if s.bridge.HasError(bridge.ErrLocalCacheUnavailable) {
		_ = s.SendEvent(NewCacheErrorEvent(CacheErrorType_CACHE_UNAVAILABLE_ERROR))
	}

	err := s.grpcServer.Serve(s.listener)
	if err != nil {
		s.log.WithError(err).Error("error serving RPC")
		return err
	}
	return nil
}

func (s *Service) NotifyManualUpdate(version updater.VersionInfo, canInstall bool) {
	if canInstall {
		_ = s.SendEvent(NewUpdateManualReadyEvent(version.Version.String()))
	} else {
		_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
	}
}

func (s *Service) SetVersion(update updater.VersionInfo) {
	s.newVersionInfo = update
	_ = s.SendEvent(NewUpdateVersionChangedEvent())
}

func (s *Service) NotifySilentUpdateInstalled() {
	_ = s.SendEvent(NewUpdateSilentRestartNeededEvent())
}

func (s *Service) NotifySilentUpdateError(err error) {
	s.log.WithError(err).Error("In app update failed, asking for manual.")
	_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_SILENT_ERROR))
}

func (s *Service) WaitUntilFrontendIsReady() {
	s.initializing.Wait()
}

func (s *Service) watchEvents() { // nolint:funlen
	errorCh := s.eventListener.ProvideChannel(events.ErrorEvent)
	credentialsErrorCh := s.eventListener.ProvideChannel(events.CredentialsErrorEvent)
	noActiveKeyForRecipientCh := s.eventListener.ProvideChannel(events.NoActiveKeyForRecipientEvent)
	internetConnChangedCh := s.eventListener.ProvideChannel(events.InternetConnChangedEvent)
	secondInstanceCh := s.eventListener.ProvideChannel(events.SecondInstanceEvent)
	restartBridgeCh := s.eventListener.ProvideChannel(events.RestartBridgeEvent)
	addressChangedCh := s.eventListener.ProvideChannel(events.AddressChangedEvent)
	addressChangedLogoutCh := s.eventListener.ProvideChannel(events.AddressChangedLogoutEvent)
	logoutCh := s.eventListener.ProvideChannel(events.LogoutEvent)
	updateApplicationCh := s.eventListener.ProvideChannel(events.UpgradeApplicationEvent)
	userChangedCh := s.eventListener.ProvideChannel(events.UserRefreshEvent)
	certIssue := s.eventListener.ProvideChannel(events.TLSCertIssue)

	// we forward events to the GUI/frontend via the gRPC event stream.
	for {
		select {
		case errorDetails := <-errorCh:
			if strings.Contains(errorDetails, "IMAP failed") {
				_ = s.SendEvent(NewMailSettingsErrorEvent(MailSettingsErrorType_IMAP_PORT_ISSUE))
			}
			if strings.Contains(errorDetails, "SMTP failed") {
				_ = s.SendEvent(NewMailSettingsErrorEvent(MailSettingsErrorType_SMTP_PORT_ISSUE))
			}
		case reason := <-credentialsErrorCh:
			if reason == keychain.ErrMacKeychainRebuild.Error() {
				_ = s.SendEvent(NewKeychainRebuildKeychainEvent())
				continue
			}
			_ = s.SendEvent(NewKeychainHasNoKeychainEvent())
		case email := <-noActiveKeyForRecipientCh:
			_ = s.SendEvent(NewMailNoActiveKeyForRecipientEvent(email))
		case stat := <-internetConnChangedCh:
			if stat == events.InternetOff {
				_ = s.SendEvent(NewInternetStatusEvent(false))
			}
			if stat == events.InternetOn {
				_ = s.SendEvent(NewInternetStatusEvent(true))
			}

		case <-secondInstanceCh:
			_ = s.SendEvent(NewShowMainWindowEvent())
		case <-restartBridgeCh:
			_, _ = s.Restart(
				metadata.AppendToOutgoingContext(context.Background(), serverTokenMetadataKey, s.token),
				&emptypb.Empty{},
			)
		case address := <-addressChangedCh:
			_ = s.SendEvent(NewMailAddressChangeEvent(address))
		case address := <-addressChangedLogoutCh:
			_ = s.SendEvent(NewMailAddressChangeLogoutEvent(address))
		case userID := <-logoutCh:
			if s.bridge == nil {
				logrus.Error("Received a logout event but bridge is not yet instantiated.")
				break
			}
			user, err := s.bridge.GetUserInfo(userID)
			if err != nil {
				return
			}
			_ = s.SendEvent(NewUserDisconnectedEvent(user.Username))
		case <-updateApplicationCh:
			s.updateForce()
		case userID := <-userChangedCh:
			_ = s.SendEvent(NewUserChangedEvent(userID))
		case <-certIssue:
			_ = s.SendEvent(NewMailApiCertIssue())
		}
	}
}

func (s *Service) loginAbort() {
	s.loginClean()
}

func (s *Service) loginClean() {
	s.auth = nil
	s.authClient = nil
	for i := range s.password {
		s.password[i] = '\x00'
	}
	s.password = s.password[0:0]
}

func (s *Service) finishLogin() {
	defer s.loginClean()

	if len(s.password) == 0 || s.auth == nil || s.authClient == nil {
		s.log.
			WithField("hasPass", len(s.password) != 0).
			WithField("hasAuth", s.auth != nil).
			WithField("hasClient", s.authClient != nil).
			Error("Finish login: authentication incomplete")

		_ = s.SendEvent(NewLoginError(LoginErrorType_TWO_PASSWORDS_ABORT, "Missing authentication, try again."))
		return
	}

	done := make(chan string)
	s.eventListener.Add(events.UserChangeDone, done)
	defer s.eventListener.Remove(events.UserChangeDone, done)

	userID, err := s.bridge.FinishLogin(s.authClient, s.auth, s.password)

	if err != nil && err != users.ErrUserAlreadyConnected {
		s.log.WithError(err).Errorf("Finish login failed")
		_ = s.SendEvent(NewLoginError(LoginErrorType_TWO_PASSWORDS_ABORT, err.Error()))
		return
	}

	// The user changed should be triggered by FinishLogin, but it is not
	// guaranteed when this is going to happen. Therefor we should wait
	// until we receive the signal from userChanged function.
	s.waitForUserChangeDone(done, userID)

	s.log.WithField("userID", userID).Debug("Login finished")
	_ = s.SendEvent(NewLoginFinishedEvent(userID))

	if err == users.ErrUserAlreadyConnected {
		s.log.WithError(err).Error("User already logged in")
		_ = s.SendEvent(NewLoginAlreadyLoggedInEvent(userID))
	}
}

func (s *Service) waitForUserChangeDone(done <-chan string, userID string) {
	for {
		select {
		case changedID := <-done:
			if changedID == userID {
				return
			}
		case <-time.After(2 * time.Second):
			s.log.WithField("ID", userID).Warning("Login finished but user not added within 2 seconds")
			return
		}
	}
}

func (s *Service) triggerReset() {
	defer func() {
		_ = s.SendEvent(NewResetFinishedEvent())
	}()
	s.bridge.FactoryReset()
}

func (s *Service) checkUpdate() {
	version, err := s.updater.Check()
	if err != nil {
		s.log.WithError(err).Error("An error occurred while checking for updates")
		s.SetVersion(updater.VersionInfo{})
		return
	}
	s.SetVersion(version)
}

func (s *Service) updateForce() {
	s.updateCheckMutex.Lock()
	defer s.updateCheckMutex.Unlock()
	s.checkUpdate()
	_ = s.SendEvent(NewUpdateForceEvent(s.newVersionInfo.Version.String()))
}

func (s *Service) checkUpdateAndNotify(isReqFromUser bool) {
	s.updateCheckMutex.Lock()
	defer func() {
		s.updateCheckMutex.Unlock()
		_ = s.SendEvent(NewUpdateCheckFinishedEvent())
	}()

	s.checkUpdate()
	version := s.newVersionInfo
	if (version.Version == nil) || (version.Version.String() == "") {
		if isReqFromUser {
			_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		}
		return
	}
	if !s.updater.IsUpdateApplicable(s.newVersionInfo) {
		s.log.Info("No need to update")
		if isReqFromUser {
			_ = s.SendEvent(NewUpdateIsLatestVersionEvent())
		}
	} else if isReqFromUser {
		s.NotifyManualUpdate(s.newVersionInfo, s.updater.CanInstall(s.newVersionInfo))
	}
}

func (s *Service) installUpdate() {
	s.updateCheckMutex.Lock()
	defer s.updateCheckMutex.Unlock()

	if !s.updater.CanInstall(s.newVersionInfo) {
		s.log.Warning("Skipping update installation, current version too old")
		_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		return
	}

	if err := s.updater.InstallUpdate(s.newVersionInfo); err != nil {
		if errors.Cause(err) == updater.ErrDownloadVerify {
			s.log.WithError(err).Warning("Skipping update installation due to temporary error")
		} else {
			s.log.WithError(err).Error("The update couldn't be installed")
			_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		}
		return
	}

	_ = s.SendEvent(NewUpdateSilentRestartNeededEvent())
}

func (s *Service) generateTLSConfig() (tlsConfig *cryptotls.Config, pemCert []byte, err error) {
	pemCert, pemKey, err := tls.NewPEMKeyPair()
	if err != nil {
		return nil, nil, errors.New("Could not get TLS config")
	}

	tlsConfig, err = tls.GetConfigFromPEMKeyPair(pemCert, pemKey)
	if err != nil {
		return nil, nil, errors.New("Could not get TLS config")
	}

	tlsConfig.ClientAuth = cryptotls.NoClientCert // skip client auth if the certificate allow it.

	return tlsConfig, pemCert, nil
}

func (s *Service) saveGRPCServerConfigFile() (string, error) {
	address, ok := s.listener.Addr().(*net.TCPAddr)
	if !ok {
		return "", fmt.Errorf("could not retrieve gRPC service listener address")
	}

	sc := config{
		Port:  address.Port,
		Cert:  s.pemCert,
		Token: s.token,
	}

	settingsPath, err := s.locations.ProvideSettingsPath()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(settingsPath, serverConfigFileName)

	return configPath, sc.save(configPath)
}

// validateServerToken verify that the server token provided by the client is valid.
func (s *Service) validateServerToken(ctx context.Context) error {
	values, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing server token")
	}

	token := values.Get(serverTokenMetadataKey)
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "missing server token")
	}

	if len(token) > 1 {
		return status.Error(codes.Unauthenticated, "more than one server token was provided")
	}

	if token[0] != s.token {
		return status.Error(codes.Unauthenticated, "invalid server token")
	}

	return nil
}

// validateUnaryServerToken check the server token for every unary gRPC call.
func (s *Service) validateUnaryServerToken(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	if err := s.validateServerToken(ctx); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// validateStreamServerToken check the server token for every gRPC stream request.
func (s *Service) validateStreamServerToken(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	logEntry := s.log.WithField("FullMethod", info.FullMethod)

	if err := s.validateServerToken(ss.Context()); err != nil {
		logEntry.WithError(err).Error("Stream validator failed")
		return err
	}

	return handler(srv, ss)
}
