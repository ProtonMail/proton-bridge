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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"crypto/tls"

	pkgTLS "github.com/ProtonMail/proton-bridge/v2/internal/config/tls"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

func (b *Bridge) GetTLSConfig() (*tls.Config, error) {
	if !b.tls.HasCerts() {
		if err := b.generateTLSCerts(); err != nil {
			return nil, err
		}
	}

	tlsConfig, err := b.tls.GetConfig()
	if err == nil {
		return tlsConfig, nil
	}

	logrus.WithError(err).Error("Failed to load TLS config, regenerating certificates")

	if err := b.generateTLSCerts(); err != nil {
		return nil, err
	}

	return b.tls.GetConfig()
}

func (b *Bridge) generateTLSCerts() error {
	template, err := pkgTLS.NewTLSTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to generate TLS template")
	}

	if err := b.tls.GenerateCerts(template); err != nil {
		return errors.Wrap(err, "failed to generate TLS certs")
	}

	if err := b.tls.InstallCerts(); err != nil {
		return errors.Wrap(err, "failed to install TLS certs")
	}

	return nil
}
