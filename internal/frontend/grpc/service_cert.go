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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package grpc

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) IsTLSCertificateInstalled(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	defer async.HandlePanic(s.panicHandler)

	s.log.Info("IsTLSCertificateInstalled")

	cert, _ := s.bridge.GetBridgeTLSCert()

	return &wrapperspb.BoolValue{Value: certs.NewInstaller().IsCertInstalled(cert)}, nil
}

func (s *Service) InstallTLSCertificate(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.log.Info("InstallTLSCertificate")

	go func() {
		defer async.HandlePanic(s.panicHandler)
		cert, _ := s.bridge.GetBridgeTLSCert()

		err := certs.NewInstaller().InstallCert(cert)
		switch {
		case err == nil:
			_ = s.SendEvent(NewCertInstallSuccessEvent())
		case errors.Is(err, certs.ErrUserCanceledCertificateInstall):
			_ = s.SendEvent(NewCertInstallCanceledEvent())
		default:
			_ = s.SendEvent(NewCertInstallFailedEvent())
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *Service) ExportTLSCertificates(_ context.Context, folderPath *wrapperspb.StringValue) (*emptypb.Empty, error) {
	s.log.WithField("folderPath", folderPath).Info("ExportTLSCertificates")

	go func() {
		defer async.HandlePanic(s.panicHandler)

		cert, key := s.bridge.GetBridgeTLSCert()

		if err := os.WriteFile(filepath.Join(folderPath.Value, "cert.pem"), cert, 0o600); err != nil {
			_ = s.SendEvent(NewGenericErrorEvent(ErrorCode_TLS_CERT_EXPORT_ERROR))
		}

		if err := os.WriteFile(filepath.Join(folderPath.Value, "key.pem"), key, 0o600); err != nil {
			_ = s.SendEvent(NewGenericErrorEvent(ErrorCode_TLS_KEY_EXPORT_ERROR))
		}
	}()

	return &emptypb.Empty{}, nil
}
