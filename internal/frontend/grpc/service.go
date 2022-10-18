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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/restarter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
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

	panicHandler   *crash.Handler
	restarter      *restarter.Restarter
	bridge         *bridge.Bridge
	eventCh        <-chan events.Event
	newVersionInfo updater.VersionInfo

	authClient *liteapi.Client
	auth       liteapi.Auth
	password   []byte

	log                *logrus.Entry
	initializing       sync.WaitGroup
	initializationDone sync.Once
	firstTimeAutostart sync.Once

	showOnStartup bool
}

// NewService returns a new instance of the service.
func NewService(
	panicHandler *crash.Handler,
	restarter *restarter.Restarter,
	locations *locations.Locations,
	bridge *bridge.Bridge,
	eventCh <-chan events.Event,
	showOnStartup bool,
) (*Service, error) {
	tlsConfig, certPEM, err := newTLSConfig()
	if err != nil {
		logrus.WithError(err).Panic("Could not generate gRPC TLS config")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0") // Port should be provided by the OS.
	if err != nil {
		logrus.WithError(err).Panic("Could not create gRPC listener")
	}

	token := uuid.NewString()

	if path, err := saveGRPCServerConfigFile(locations, listener, token, certPEM); err != nil {
		logrus.WithError(err).WithField("path", path).Panic("Could not write gRPC service config file")
	} else {
		logrus.WithField("path", path).Info("Successfully saved gRPC service config file")
	}

	s := &Service{
		grpcServer: grpc.NewServer(
			grpc.Creds(credentials.NewTLS(tlsConfig)),
			grpc.UnaryInterceptor(newUnaryTokenValidator(token)),
			grpc.StreamInterceptor(newStreamTokenValidator(token)),
		),
		listener: listener,

		panicHandler: panicHandler,
		restarter:    restarter,
		bridge:       bridge,
		eventCh:      eventCh,

		log:                logrus.WithField("pkg", "grpc"),
		initializing:       sync.WaitGroup{},
		initializationDone: sync.Once{},
		firstTimeAutostart: sync.Once{},

		showOnStartup: showOnStartup,
	}

	// Initializing.Done is only called sync.Once. Please keep the increment set to 1
	s.initializing.Add(1)

	// Initialize the autostart.
	s.initAutostart()

	// Register the gRPC service implementation.
	RegisterBridgeServer(s.grpcServer, s)

	s.log.Info("gRPC server listening on ", s.listener.Addr())

	return s, nil
}

// GODT-1507 Windows: autostart needs to be created after Qt is initialized.
// GODT-1206: if preferences file says it should be on enable it here.
// TO-DO GODT-1681 Autostart needs to be properly implement for gRPC approach.
func (s *Service) initAutostart() {
	s.firstTimeAutostart.Do(func() {
		shouldAutostartBeOn := s.bridge.GetAutostart()
		if s.bridge.GetFirstStart() || shouldAutostartBeOn {
			if err := s.bridge.SetAutostart(true); err != nil {
				s.log.WithField("prefs", shouldAutostartBeOn).WithError(err).Error("Failed to enable first autostart")
			}
			return
		}
	})
}

func (s *Service) Loop() error {
	defer func() {
		_ = s.bridge.SetFirstStartGUI(false)
	}()

	go func() {
		defer s.panicHandler.HandlePanic()
		s.watchEvents()
	}()

	s.log.Info("Starting gRPC server")

	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.log.WithError(err).Error("Failed to serve gRPC")
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

func (s *Service) watchEvents() { //nolint:funlen
	// GODT-1949 Better error events.
	for _, err := range s.bridge.GetErrors() {
		switch {
		case errors.Is(err, bridge.ErrVaultCorrupt):
			_ = s.SendEvent(NewKeychainHasNoKeychainEvent())

		case errors.Is(err, bridge.ErrVaultInsecure):
			_ = s.SendEvent(NewKeychainHasNoKeychainEvent())

		case errors.Is(err, bridge.ErrServeIMAP):
			_ = s.SendEvent(NewMailSettingsErrorEvent(MailSettingsErrorType_IMAP_PORT_ISSUE))

		case errors.Is(err, bridge.ErrServeSMTP):
			_ = s.SendEvent(NewMailSettingsErrorEvent(MailSettingsErrorType_SMTP_PORT_ISSUE))
		}
	}

	for event := range s.eventCh {
		switch event := event.(type) {
		case events.ConnStatusUp:
			_ = s.SendEvent(NewInternetStatusEvent(true))

		case events.ConnStatusDown:
			_ = s.SendEvent(NewInternetStatusEvent(false))

		case events.Raise:
			_ = s.SendEvent(NewShowMainWindowEvent())

		case events.UserAddressCreated:
			_ = s.SendEvent(NewMailAddressChangeEvent(event.Email))

		case events.UserAddressUpdated:
			_ = s.SendEvent(NewMailAddressChangeEvent(event.Email))

		case events.UserAddressDeleted:
			_ = s.SendEvent(NewMailAddressChangeLogoutEvent(event.Email))

		case events.UserChanged:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoaded:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoggedIn:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoggedOut:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserDeleted:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserDeauth:
			if user, err := s.bridge.GetUserInfo(event.UserID); err != nil {
				s.log.WithError(err).Error("Failed to get user info")
			} else {
				_ = s.SendEvent(NewUserDisconnectedEvent(user.Username))
			}

		case events.TLSIssue:
			_ = s.SendEvent(NewMailApiCertIssue())

		case events.UpdateForced:
			panic("TODO")
		}
	}
}

func (s *Service) loginAbort() {
	s.loginClean()
}

func (s *Service) loginClean() {
	s.auth = liteapi.Auth{}
	s.authClient = nil
	for i := range s.password {
		s.password[i] = '\x00'
	}
	s.password = s.password[0:0]
}

func (s *Service) finishLogin() {
	defer s.loginClean()

	if len(s.password) == 0 || s.auth.UID == "" || s.authClient == nil {
		s.log.
			WithField("hasPass", len(s.password) != 0).
			WithField("hasAuth", s.auth.UID != "").
			WithField("hasClient", s.authClient != nil).
			Error("Finish login: authentication incomplete")

		_ = s.SendEvent(NewLoginError(LoginErrorType_TWO_PASSWORDS_ABORT, "Missing authentication, try again."))
		return
	}

	eventCh, done := s.bridge.GetEvents(events.UserLoggedIn{})
	defer done()

	userID, err := s.bridge.LoginUser(context.Background(), s.authClient, s.auth, s.password)
	if err != nil {
		s.log.WithError(err).Errorf("Finish login failed")
		_ = s.SendEvent(NewLoginError(LoginErrorType_TWO_PASSWORDS_ABORT, err.Error()))
		return
	}

	s.waitForUserChangeDone(eventCh, userID)

	s.log.WithField("userID", userID).Debug("Login finished")

	_ = s.SendEvent(NewLoginFinishedEvent(userID))
}

func (s *Service) waitForUserChangeDone(eventCh <-chan events.Event, userID string) {
	for {
		select {
		case event := <-eventCh:
			if login, ok := event.(events.UserLoggedIn); ok && login.UserID == userID {
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

	s.bridge.FactoryReset(context.Background())
}

func newTLSConfig() (*tls.Config, []byte, error) {
	template, err := certs.NewTLSTemplate()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS template: %w", err)
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate cert: %w", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load cert: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	}, certPEM, nil
}

func saveGRPCServerConfigFile(locations *locations.Locations, listener net.Listener, token string, certPEM []byte) (string, error) {
	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return "", fmt.Errorf("could not retrieve gRPC service listener address")
	}

	sc := config{
		Port:  address.Port,
		Cert:  string(certPEM),
		Token: token,
	}

	settingsPath, err := locations.ProvideSettingsPath()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(settingsPath, serverConfigFileName)

	return configPath, sc.save(configPath)
}

// validateServerToken verify that the server token provided by the client is valid.
func validateServerToken(ctx context.Context, wantToken string) error {
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

	if token[0] != wantToken {
		return status.Error(codes.Unauthenticated, "invalid server token")
	}

	return nil
}

// newUnaryTokenValidator checks the server token for every unary gRPC call.
func newUnaryTokenValidator(wantToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validateServerToken(ctx, wantToken); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// newStreamTokenValidator checks the server token for every gRPC stream request.
func newStreamTokenValidator(wantToken string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := validateServerToken(stream.Context(), wantToken); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}
