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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) GetUserList(context.Context, *emptypb.Empty) (*UserListResponse, error) {
	s.log.Info("GetUserList")

	users := s.bridge.GetUsers()

	userList := make([]*User, len(users))
	for i, user := range users {
		userList[i] = grpcUserFromBridge(user)
	}

	// If there are no active accounts.
	if len(userList) == 0 {
		s.log.Info("No active accounts")
	}

	return &UserListResponse{Users: userList}, nil
}

func (s *Service) GetUser(_ context.Context, userID *wrapperspb.StringValue) (*User, error) {
	s.log.WithField("userID", userID).Info("GetUser")

	user, err := s.bridge.GetUser(userID.Value)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	return grpcUserFromBridge(user), nil
}

func (s *Service) SetUserSplitMode(_ context.Context, splitMode *UserSplitModeRequest) (*emptypb.Empty, error) {
	s.log.WithField("UserID", splitMode.UserID).WithField("Active", splitMode.Active).Info("SetUserSplitMode")

	user, err := s.bridge.GetUser(splitMode.UserID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", splitMode.UserID)
	}

	go func() {
		defer s.panicHandler.HandlePanic()
		defer func() { _ = s.SendEvent(NewUserToggleSplitModeFinishedEvent(splitMode.UserID)) }()
		if splitMode.Active == user.IsCombinedAddressMode() {
			_ = user.SwitchAddressMode() // check for errors
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) LogoutUser(_ context.Context, userID *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("UserID", userID.Value).Info("LogoutUser")

	user, err := s.bridge.GetUser(userID.Value)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found %v", userID.Value)
	}

	go func() {
		defer s.panicHandler.HandlePanic()
		_ = user.Logout()
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

func (s *Service) ConfigureUserAppleMail(_ context.Context, request *ConfigureAppleMailRequest) (*emptypb.Empty, error) {
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
		s.restart()
	}

	return &emptypb.Empty{}, nil
}
