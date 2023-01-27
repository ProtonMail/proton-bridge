// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
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

	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) GetUserList(ctx context.Context, _ *emptypb.Empty) (*UserListResponse, error) {
	s.log.Debug("GetUserList")

	userIDs := s.bridge.GetUserIDs()
	userList := make([]*User, len(userIDs))

	for idx, userID := range userIDs {
		user, err := s.bridge.GetUserInfo(userID)
		if err != nil {
			return nil, err
		}

		userList[idx] = grpcUserFromInfo(user)
	}

	// If there are no active accounts.
	if len(userList) == 0 {
		s.log.Debug("No active accounts")
	}

	return &UserListResponse{Users: userList}, nil
}

func (s *Service) GetUser(ctx context.Context, userID *wrapperspb.StringValue) (*User, error) {
	s.log.WithField("userID", userID).Debug("GetUser")

	user, err := s.bridge.GetUserInfo(userID.Value)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	return grpcUserFromInfo(user), nil
}

func (s *Service) SetUserSplitMode(ctx context.Context, splitMode *UserSplitModeRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", splitMode.UserID).WithField("Active", splitMode.Active).Debug("SetUserSplitMode")

	user, err := s.bridge.GetUserInfo(splitMode.UserID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", splitMode.UserID)
	}

	go func() {
		defer s.panicHandler.HandlePanic()
		defer func() { _ = s.SendEvent(NewUserToggleSplitModeFinishedEvent(splitMode.UserID)) }()

		var targetMode vault.AddressMode

		if splitMode.Active {
			targetMode = vault.SplitMode
		} else {
			targetMode = vault.CombinedMode
		}

		if user.AddressMode == targetMode {
			return
		}

		if err := s.bridge.SetAddressMode(context.Background(), user.UserID, targetMode); err != nil {
			s.log.WithError(err).Error("Failed to set address mode")
		}

		s.log.WithField("userID", user.UserID).WithField("mode", targetMode).Info("Address mode changed")
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) LogoutUser(ctx context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Debug("LogoutUser")

	if _, err := s.bridge.GetUserInfo(userID.Value); err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	go func() {
		defer s.panicHandler.HandlePanic()

		if err := s.bridge.LogoutUser(context.Background(), userID.Value); err != nil {
			s.log.WithError(err).Error("Failed to log user out")
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveUser(ctx context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Debug("RemoveUser")

	go func() {
		defer s.panicHandler.HandlePanic()

		// remove preferences
		if err := s.bridge.DeleteUser(context.Background(), userID.Value); err != nil {
			s.log.WithError(err).Error("Failed to remove user")
			// notification
		}
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) ConfigureUserAppleMail(ctx context.Context, request *ConfigureAppleMailRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", request.UserID).WithField("Address", request.Address).Debug("ConfigureUserAppleMail")

	sslWasEnabled := s.bridge.GetSMTPSSL()

	if err := s.bridge.ConfigureAppleMail(request.UserID, request.Address); err != nil {
		s.log.WithField("userID", request.UserID).Error("Cannot configure AppleMail for user")
		return nil, status.Error(codes.Internal, "Apple Mail config failed")
	}

	if s.bridge.GetSMTPSSL() != sslWasEnabled {
		// we've changed SMTP SSL settings. This will happen if SSL was off and macOS >= Catalina. We must inform gRPC clients.
		_ = s.SendEvent(NewMailServerSettingsChangedEvent(s.getMailServerSettings()))
	}

	return &emptypb.Empty{}, nil
}
