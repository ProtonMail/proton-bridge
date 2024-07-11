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

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/kb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RunEventStream implement the gRPC server->Client event stream.
func (s *Service) RunEventStream(request *EventStreamRequest, server Bridge_RunEventStreamServer) error {
	s.log.Debug("Starting Event stream")

	if s.isStreamingEvents() {
		return status.Errorf(codes.AlreadyExists, "the service is already streaming") // TO-DO GODT-1667 decide if we want to kill the existing stream.
	}

	s.bridge.SetCurrentPlatform(request.ClientPlatform)

	s.createEventStreamChannel()
	s.eventStreamDoneCh = make(chan struct{})

	// TO-DO GODT-1667 We should have a safer we to close this channel? What if an event occur while we are closing?
	defer func() {
		close(s.eventStreamCh)
		s.deleteEventStreamChannel()
		close(s.eventStreamDoneCh)
		s.eventStreamDoneCh = nil
	}()

	// if events occurred before streaming started, they've been queued. Now that the stream channel is available
	// we can flush the queued
	go func() {
		defer async.HandlePanic(s.panicHandler)

		s.eventQueueMutex.Lock()
		defer s.eventQueueMutex.Unlock()
		for _, event := range s.eventQueue {
			s.eventStreamCh <- event
		}

		s.eventQueue = nil
	}()

	for {
		select {
		case <-s.eventStreamDoneCh:
			s.log.Debug("Stop Event stream")
			return nil

		case event := <-s.eventStreamCh:
			s.log.WithField("event", event).Debug("Sending event")
			if err := server.Send(event); err != nil {
				s.log.Debug("Stop Event stream")
				return err
			}
		case <-server.Context().Done():
			s.log.Debug("Client closed the stream, exiting")
			s.quit()
			return nil
		}
	}
}

// StopEventStream stops the event stream.
func (s *Service) StopEventStream(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.stopEventStream()
}

func (s *Service) stopEventStream() error {
	s.eventStreamChMutex.RLock()
	defer s.eventStreamChMutex.RUnlock()

	if s.eventStreamCh == nil {
		return status.Errorf(codes.NotFound, "The service is not streaming")
	}

	s.eventStreamDoneCh <- struct{}{}

	return nil
}

// SendEvent sends an event to the via the gRPC event stream.
func (s *Service) SendEvent(event *StreamEvent) error {
	if !s.isStreamingEvents() { // nobody is connected to the event stream, we queue events
		s.queueEvent(event)
		return nil
	}

	s.eventStreamCh <- event

	return nil
}

// StartEventTest sends all the known event via gRPC.
func (s *Service) StartEventTest() error {
	const dummyAddress = "dummy@proton.me"
	events := []*StreamEvent{
		// app
		NewInternetStatusEvent(true),
		NewToggleAutostartFinishedEvent(),
		NewResetFinishedEvent(),
		NewReportBugFinishedEvent(),
		NewReportBugSuccessEvent(),
		NewReportBugErrorEvent(),
		NewShowMainWindowEvent(),
		NewRequestKnowledgeBaseSuggestionsEvent(kb.ArticleList{}),

		// login
		NewLoginError(LoginErrorType_FREE_USER, "error"),
		NewLoginTfaRequestedEvent(dummyAddress),
		NewLoginTwoPasswordsRequestedEvent(dummyAddress),
		NewLoginFinishedEvent("userID", false),
		NewLoginAlreadyLoggedInEvent("userID"),

		// update
		NewUpdateErrorEvent(UpdateErrorType_UPDATE_SILENT_ERROR),
		NewUpdateManualReadyEvent("2.0"),
		NewUpdateManualRestartNeededEvent(),
		NewUpdateForceEvent("2.0"),
		NewUpdateSilentRestartNeededEvent(),
		NewUpdateIsLatestVersionEvent(),
		NewUpdateCheckFinishedEvent(),

		// cache

		NewDiskCacheErrorEvent(DiskCacheErrorType_CANT_MOVE_DISK_CACHE_ERROR),
		NewDiskCachePathChangedEvent("/dummy/path/"),
		NewDiskCachePathChangeFinishedEvent(),

		// mail settings
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_PORT_STARTUP_ERROR),
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_PORT_STARTUP_ERROR),
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_PORT_CHANGE_ERROR),
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_PORT_CHANGE_ERROR),
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_IMAP_CONNECTION_MODE_CHANGE_ERROR),
		NewMailServerSettingsErrorEvent(MailServerSettingsErrorType_SMTP_CONNECTION_MODE_CHANGE_ERROR),
		NewMailServerSettingsChangedEvent(&ImapSmtpSettings{
			ImapPort:      1143,
			SmtpPort:      1025,
			UseSSLForImap: false,
			UseSSLForSmtp: false,
		}),
		NewChangeMailServerSettingsFinishedEvent(),

		// keychain
		NewKeychainChangeKeychainFinishedEvent(),
		NewKeychainHasNoKeychainEvent(),
		NewKeychainRebuildKeychainEvent(),

		// mail
		NewMailAddressChangeEvent(dummyAddress),
		NewMailAddressChangeLogoutEvent(dummyAddress),
		NewMailApiCertIssue(),

		// user
		NewUserToggleSplitModeFinishedEvent("userID"),
		NewUserDisconnectedEvent("username"),
		NewUserChangedEvent("userID"),
		NewUsedBytesChangedEvent("userID", 1000),
	}

	for _, event := range events {
		if err := s.SendEvent(event); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) queueEvent(event *StreamEvent) {
	s.eventQueueMutex.Lock()
	defer s.eventQueueMutex.Unlock()

	if event.isInternetStatus() {
		s.eventQueue = append(filterOutInternetStatusEvents(s.eventQueue), event)
	} else {
		s.eventQueue = append(s.eventQueue, event)
	}
}

func (s *Service) isStreamingEvents() bool {
	s.eventStreamChMutex.RLock()
	defer s.eventStreamChMutex.RUnlock()

	return s.eventStreamCh != nil
}

func (s *Service) createEventStreamChannel() {
	s.eventStreamChMutex.Lock()
	defer s.eventStreamChMutex.Unlock()

	s.eventStreamCh = make(chan *StreamEvent)
}

func (s *Service) deleteEventStreamChannel() {
	s.eventStreamChMutex.Lock()
	defer s.eventStreamChMutex.Unlock()

	s.eventStreamCh = nil
}
