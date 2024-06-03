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

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative bridge.proto

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/hv"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/service"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/bradenaw/juniper/xslices"
	"github.com/elastic/go-sysinfo"
	sysinfotypes "github.com/elastic/go-sysinfo/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

const (
	serverConfigFileName        = "grpcServerConfig.json"
	serverTokenMetadataKey      = "server-token"
	twoPasswordsMaxAttemptCount = 3 // The number of attempts allowed for the mailbox password.
)

// Service is the RPC service struct.
type Service struct { // nolint:structcheck
	UnimplementedBridgeServer

	grpcServer         *grpc.Server //  the gGRPC server
	listener           net.Listener
	eventStreamCh      chan *StreamEvent
	eventStreamChMutex sync.RWMutex
	eventStreamDoneCh  chan struct{}
	eventQueue         []*StreamEvent
	eventQueueMutex    sync.Mutex

	panicHandler async.PanicHandler
	restarter    Restarter
	bridge       *bridge.Bridge
	eventCh      <-chan events.Event
	quitCh       <-chan struct{}

	latest     updater.VersionInfo
	latestLock safe.RWMutex

	target     updater.VersionInfo
	targetLock safe.RWMutex

	authClient              *proton.Client
	auth                    proton.Auth
	password                []byte
	twoPasswordAttemptCount int

	log                *logrus.Entry
	initializing       sync.WaitGroup
	initializationDone sync.Once
	firstTimeAutostart sync.Once
	parentPID          int
	parentPIDDoneCh    chan struct{}
	showOnStartup      bool

	hvDetails    *proton.APIHVDetails
	useHvDetails bool
}

// NewService returns a new instance of the service.
func NewService(
	panicHandler async.PanicHandler,
	restarter Restarter,
	locations service.Locator,
	bridge *bridge.Bridge,
	eventCh <-chan events.Event,
	quitCh <-chan struct{},
	showOnStartup bool,
	parentPID int,
) (*Service, error) {
	tlsConfig, certPEM, err := newTLSConfig()
	if err != nil {
		logrus.WithError(err).Panic("Could not generate gRPC TLS config")
	}

	config := service.Config{
		Cert:  string(certPEM),
		Token: uuid.NewString(),
	}

	var listener net.Listener
	if useFileSocket() {
		var err error
		if config.FileSocketPath, err = computeFileSocketPath(); err != nil {
			logrus.WithError(err).WithError(err).Panic("Could not create gRPC file socket")
		}

		listener, err = net.Listen("unix", config.FileSocketPath)
		if err != nil {
			logrus.WithError(err).Panic("Could not create gRPC file socket listener")
		}
	} else {
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0") // Port should be provided by the OS.
		if err != nil {
			logrus.WithError(err).Panic("Could not create gRPC listener")
		}

		// retrieve the port assigned by the system, so that we can put it in the config file.
		address, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			return nil, fmt.Errorf("could not retrieve gRPC service listener address")
		}
		config.Port = address.Port
	}

	if path, err := service.SaveGRPCServerConfigFile(locations, &config, serverConfigFileName); err != nil {
		logrus.WithError(err).WithField("path", path).Panic("Could not write gRPC service config file")
	} else {
		logrus.WithField("path", path).Info("Successfully saved gRPC service config file")
	}

	s := &Service{
		grpcServer: grpc.NewServer(
			grpc.Creds(credentials.NewTLS(tlsConfig)),
			grpc.UnaryInterceptor(newUnaryTokenValidator(config.Token)),
			grpc.StreamInterceptor(newStreamTokenValidator(config.Token)),
		),
		listener: listener,

		panicHandler: panicHandler,
		restarter:    restarter,
		bridge:       bridge,
		eventCh:      eventCh,
		quitCh:       quitCh,

		latest:     updater.VersionInfo{},
		latestLock: safe.NewRWMutex(),

		target:     updater.VersionInfo{},
		targetLock: safe.NewRWMutex(),

		log:                logrus.WithField("pkg", "grpc"),
		initializing:       sync.WaitGroup{},
		initializationDone: sync.Once{},
		firstTimeAutostart: sync.Once{},

		parentPID:       parentPID,
		parentPIDDoneCh: make(chan struct{}),
		showOnStartup:   showOnStartup,
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
	if s.parentPID < 0 {
		s.log.Info("Not monitoring parent PID")
	} else {
		go func() {
			defer async.HandlePanic(s.panicHandler)
			s.monitorParentPID()
		}()
	}

	go func() {
		defer async.HandlePanic(s.panicHandler)
		s.watchEvents()
	}()

	s.log.WithField("useFileSocket", useFileSocket()).Info("Starting gRPC server")

	doneCh := make(chan struct{})
	defer close(doneCh)

	go func() {
		defer async.HandlePanic(s.panicHandler)

		select {
		case <-s.quitCh:
			s.log.Info("Stopping gRPC server")
			defer s.log.Info("Stopped gRPC server")

			s.grpcServer.Stop()

		case <-doneCh:
			// ...
		}
	}()

	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.log.WithError(err).Error("Failed to serve gRPC")
		return err
	}

	return nil
}

func (s *Service) WaitUntilFrontendIsReady() {
	s.initializing.Wait()
}

// nolint:gocyclo
func (s *Service) watchEvents() {
	// GODT-1949 Better error events.
	for _, err := range s.bridge.GetErrors() {
		switch {
		case errors.Is(err, bridge.ErrVaultCorrupt):
			// _ = s.SendEvent(NewKeychainHasNoKeychainEvent())

		case errors.Is(err, bridge.ErrVaultInsecure):
			_ = s.SendEvent(NewKeychainHasNoKeychainEvent())
		}
	}

	for event := range s.eventCh {
		switch event := event.(type) {
		case events.ConnStatusUp:
			_ = s.SendEvent(NewInternetStatusEvent(true))

		case events.ConnStatusDown:
			_ = s.SendEvent(NewInternetStatusEvent(false))

		case events.IMAPServerError:
			_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_PORT_STARTUP_ERROR))

		case events.SMTPServerError:
			_ = s.SendEvent(NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_PORT_STARTUP_ERROR))

		case events.Raise:
			_ = s.SendEvent(NewShowMainWindowEvent())

		case events.UserAddressCreated:
			_ = s.SendEvent(NewMailAddressChangeEvent(event.Email))

		case events.UserAddressEnabled:
			_ = s.SendEvent(NewMailAddressChangeEvent(event.Email))

		case events.UserAddressDisabled:
			_ = s.SendEvent(NewMailAddressChangeLogoutEvent(event.Email))

		case events.UserAddressDeleted:
			_ = s.SendEvent(NewMailAddressChangeLogoutEvent(event.Email))

		case events.UserChanged:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoadSuccess:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoadFail:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoggedIn:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserLoggedOut:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UserDeleted:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.AddressModeChanged:
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

		case events.UsedSpaceChanged:
			_ = s.SendEvent(NewUsedBytesChangedEvent(event.UserID, event.UsedSpace))

		case events.IMAPLoginFailed:
			_ = s.SendEvent(newIMAPLoginFailedEvent(event.Username))

		case events.UserDeauth:
			// This is the event the GUI cares about.
			_ = s.SendEvent(NewUserChangedEvent(event.UserID))

			// The GUI doesn't care about this event... not sure why we still emit it. GODT-2128.
			if user, err := s.bridge.GetUserInfo(event.UserID); err == nil {
				_ = s.SendEvent(NewUserDisconnectedEvent(user.Username))
			}

		case events.UserBadEvent:
			_ = s.SendEvent(NewUserBadEvent(event.UserID, event.Error.Error()))

		case events.SyncStarted:
			_ = s.SendEvent(NewSyncStartedEvent(event.UserID))

		case events.SyncFinished:
			_ = s.SendEvent(NewSyncFinishedEvent(event.UserID))

		case events.SyncFailed:
			if errors.Is(event.Error, context.Canceled) {
				_ = s.SendEvent(NewSyncFinishedEvent(event.UserID))
			}

		case events.SyncProgress:
			_ = s.SendEvent(NewSyncProgressEvent(event.UserID, event.Progress, event.Elapsed.Milliseconds(), event.Remaining.Milliseconds()))

		case events.UpdateLatest:
			safe.RLock(func() {
				s.latest = event.Version
			}, s.latestLock)

			_ = s.SendEvent(NewUpdateVersionChangedEvent())

		case events.UpdateAvailable:
			switch {
			case !event.Compatible:
				_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))

			case !event.Silent:
				safe.RLock(func() {
					s.target = event.Version
				}, s.targetLock)

				_ = s.SendEvent(NewUpdateManualReadyEvent(event.Version.Version.String()))
			}

		case events.UpdateInstalled:
			if event.Silent {
				_ = s.SendEvent(NewUpdateSilentRestartNeededEvent())
			} else {
				_ = s.SendEvent(NewUpdateManualRestartNeededEvent())
			}

		case events.UpdateFailed:
			if event.Silent {
				_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_SILENT_ERROR))
			} else {
				_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
			}

		case events.UpdateForced:
			var latest string

			if s.latest.Version != nil {
				latest = s.latest.Version.String()
			} else if version, ok := s.checkLatestVersion(); ok {
				latest = version.Version.String()
			} else {
				latest = "unknown"
			}

			_ = s.SendEvent(NewUpdateForceEvent(latest))

		case events.TLSIssue:
			_ = s.SendEvent(NewMailApiCertIssue())

		case events.AllUsersLoaded:
			_ = s.SendEvent(NewAllUsersLoadedEvent())
		}
	}
}

func (s *Service) loginAbort() {
	s.loginClean()
}

func (s *Service) loginClean() {
	s.auth = proton.Auth{}
	s.authClient = nil
	for i := range s.password {
		s.password[i] = '\x00'
	}
	s.password = s.password[0:0]
	s.useHvDetails = false
}

func (s *Service) finishLogin() {
	performCleanup := true
	defer func() {
		if performCleanup {
			s.loginClean()
		}
	}()

	wasSignedOut := s.bridge.HasUser(s.auth.UserID)

	var hvDetails *proton.APIHVDetails
	if s.useHvDetails {
		hvDetails = s.hvDetails
	}

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

	ctx := context.Background()
	userID, err := s.bridge.LoginUser(ctx, s.authClient, s.auth, s.password, hvDetails)
	if err != nil {
		if hv.IsHvRequest(err) {
			s.handleHvRequest(err)
			performCleanup = false
			return
		}

		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Code == proton.HumanValidationInvalidToken {
			s.hvDetails = nil
			_ = s.SendEvent(NewLoginError(LoginErrorType_HV_ERROR, err.Error()))
			return
		}

		s.log.WithError(err).Errorf("Finish login failed")
		s.twoPasswordAttemptCount++
		errType := LoginErrorType_TWO_PASSWORDS_ABORT
		if errors.Is(err, bridge.ErrFailedToUnlock) {
			if s.twoPasswordAttemptCount < twoPasswordsMaxAttemptCount {
				performCleanup = false
				errType = LoginErrorType_TWO_PASSWORDS_ERROR
			} else {
				if deleteErr := s.authClient.AuthDelete(ctx); deleteErr != nil {
					s.log.WithError(deleteErr).Error("Failed to delete auth")
				}
			}
		}

		_ = s.SendEvent(NewLoginError(errType, err.Error()))
		return
	}

	s.waitForUserChangeDone(eventCh, userID)

	s.log.WithField("userID", userID).Debug("Login finished")

	_ = s.SendEvent(NewLoginFinishedEvent(userID, wasSignedOut))
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

func (s *Service) checkLatestVersion() (updater.VersionInfo, bool) {
	updateCh, done := s.bridge.GetEvents(events.UpdateLatest{})
	defer done()

	s.bridge.CheckForUpdates()

	select {
	case event := <-updateCh:
		if latest, ok := event.(events.UpdateLatest); ok {
			return latest.Version, true
		}

	case <-time.After(5 * time.Second):
		// ...
	}

	return updater.VersionInfo{}, false
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

// monitorParentPID check at regular intervals that the parent process is still alive, and if not shuts down the server
// and the applications.
func (s *Service) monitorParentPID() {
	s.log.Infof("Starting to monitor parent PID %v", s.parentPID)
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			if s.parentPID < 0 {
				continue
			}

			processes, err := sysinfo.Processes() // sysinfo.Process(pid) does not seem to work on Windows.
			if err != nil {
				s.log.Debug("Could not retrieve process list")
				continue
			}

			if !xslices.Any(processes, func(p sysinfotypes.Process) bool { return p != nil && p.PID() == s.parentPID }) {
				s.log.Info("Parent process does not exist anymore. Initiating shutdown")
				// quit will write to the parentPIDDoneCh, so we launch a goroutine.
				go func() {
					defer async.HandlePanic(s.panicHandler)

					if err := s.quit(); err != nil {
						logrus.WithError(err).Error("Error on quit")
					}
				}()
			}

		case <-s.parentPIDDoneCh:
			s.log.Infof("Stopping process monitoring for PID %v", s.parentPID)
			return
		}
	}
}

func (s *Service) handleHvRequest(err error) {
	hvDet, hvErr := hv.VerifyAndExtractHvRequest(err)
	if hvErr != nil {
		_ = s.SendEvent(NewLoginError(LoginErrorType_HV_ERROR, hvErr.Error()))
		return
	}

	s.hvDetails = hvDet
	hvChallengeURL := hv.FormatHvURL(hvDet)
	_ = s.SendEvent(NewLoginHvRequestedEvent(hvChallengeURL))
}

// computeFileSocketPath Return an available path for a socket file in the temp folder.
func computeFileSocketPath() (string, error) {
	tempPath := os.TempDir()
	for i := 0; i < 1000; i++ {
		path := filepath.Join(tempPath, fmt.Sprintf("bridge%04d", rand.Intn(10000))) // nolint:gosec
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return path, nil
		}

		if err := os.Remove(path); err != nil {
			logrus.WithField("path", path).WithError(err).Warning("Could not remove existing socket file")
			continue
		}

		return path, nil
	}

	return "", errors.New("unable to find a suitable file socket in user config folder")
}

// useFileSocket return true iff file socket should be used for the gRPC service.
func useFileSocket() bool {
	//goland:noinspection GoBoolExpressions
	return runtime.GOOS != "windows"
}
