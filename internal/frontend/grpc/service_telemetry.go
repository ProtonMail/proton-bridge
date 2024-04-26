// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package grpc

import (
	"context"

	"github.com/ProtonMail/gluon/async"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) ReportBugClicked(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.bridge.ReportBugClicked()
	return &emptypb.Empty{}, nil
}

func (s *Service) AutoconfigClicked(_ context.Context, client *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.bridge.AutoconfigUsed(client.Value)
	return &emptypb.Empty{}, nil
}

func (s *Service) ExternalLinkClicked(_ context.Context, article *wrapperspb.StringValue) (*emptypb.Empty, error) {
	defer async.HandlePanic(s.panicHandler)
	s.bridge.ExternalLinkClicked(article.Value)
	return &emptypb.Empty{}, nil
}
