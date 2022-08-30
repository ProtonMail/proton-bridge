// Copyright (c) 2022 Proton AG
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
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) GetUserList(context.Context, *emptypb.Empty) (*UserListResponse, error) {
	s.log.Info("GetUserList")

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
		s.log.Info("No active accounts")
	}

	return &UserListResponse{Users: userList}, nil
}

func (s *Service) GetUser(_ context.Context, userID *wrapperspb.StringValue) (*User, error) {
	s.log.WithField("userID", userID).Info("GetUser")

	user, err := s.bridge.GetUserInfo(userID.Value)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	return grpcUserFromInfo(user), nil
}

func (s *Service) SetUserSplitMode(_ context.Context, splitMode *UserSplitModeRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", splitMode.UserID).WithField("Active", splitMode.Active).Info("SetUserSplitMode")

	user, err := s.bridge.GetUserInfo(splitMode.UserID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", splitMode.UserID)
	}

	go func() {
		defer s.panicHandler.HandlePanic()
		defer func() { _ = s.SendEvent(NewUserToggleSplitModeFinishedEvent(splitMode.UserID)) }()

		var targetMode users.AddressMode

		if splitMode.Active && user.Mode == users.CombinedMode {
			targetMode = users.SplitMode
		} else if !splitMode.Active && user.Mode == users.SplitMode {
			targetMode = users.CombinedMode
		}

		if err := s.bridge.SetAddressMode(user.ID, targetMode); err != nil {
			logrus.WithError(err).Error("Failed to set address mode")
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) LogoutUser(_ context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Info("LogoutUser")

	if _, err := s.bridge.GetUserInfo(userID.Value); err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	go func() {
		defer s.panicHandler.HandlePanic()

		if err := s.bridge.LogoutUser(userID.Value); err != nil {
			logrus.WithError(err).Error("Failed to log user out")
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveUser(_ context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Info("RemoveUser")

	go func() {
		defer s.panicHandler.HandlePanic()

		// remove preferences
		if err := s.bridge.DeleteUser(userID.Value, false); err != nil {
			s.log.WithError(err).Error("Failed to remove user")
			// notification
		}
	}()
	return &emptypb.Empty{}, nil
}

func (s *Service) ConfigureUserAppleMail(ctx context.Context, request *ConfigureAppleMailRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", request.UserID).WithField("Address", request.Address).Info("ConfigureUserAppleMail")

	restart, err := s.bridge.ConfigureAppleMail(request.UserID, request.Address)
	if err != nil {
		s.log.WithField("userID", request.UserID).Error("Cannot configure AppleMail for user")
		return nil, status.Error(codes.Internal, "Apple Mail config failed")
	}

	// There is delay needed for external window to open.
	if restart {
		s.log.Warn("Detected Catalina or newer with bad SMTP SSL settings, now using SSL, bridge needs to restart")
		time.Sleep(2 * time.Second)
		return s.Restart(ctx, &emptypb.Empty{})
	}

	return &emptypb.Empty{}, nil
}
