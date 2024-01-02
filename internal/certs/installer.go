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

package certs

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	ErrUserCanceledCertificateInstall = errors.New("the user cancelled the authorization dialog")
)

type Installer struct {
	log *logrus.Entry
}

func NewInstaller() *Installer {
	return &Installer{
		log: logrus.WithField("pkg", "certs"),
	}
}

func (installer *Installer) OSSupportCertInstall() bool {
	return osSupportCertInstall()
}

func (installer *Installer) InstallCert(certPEM []byte) error {
	installer.log.Info("Installing the Bridge TLS certificate in the OS keychain")

	if err := installCert(certPEM); err != nil {
		installer.log.WithError(err).Error("The Bridge TLS certificate could not be installed in the OS keychain")
		return err
	}

	installer.log.Info("The Bridge TLS certificate was successfully installed in the OS keychain")
	return nil
}

func (installer *Installer) UninstallCert(certPEM []byte) error {
	installer.log.Info("Uninstalling the Bridge TLS certificate from the OS keychain")

	if err := uninstallCert(certPEM); err != nil {
		installer.log.WithError(err).Error("The Bridge TLS certificate could not be uninstalled from the OS keychain")
		return err
	}

	installer.log.Info("The Bridge TLS certificate was successfully uninstalled from the OS keychain")
	return nil
}

func (installer *Installer) IsCertInstalled(certPEM []byte) bool {
	return isCertInstalled(certPEM)
}

// LogCertInstallStatus reports the current status of the certificate installation in the log.
// If certificate installation is not supported on the platform, this function does nothing.
func (installer *Installer) LogCertInstallStatus(certPEM []byte) {
	if installer.OSSupportCertInstall() {
		if installer.IsCertInstalled(certPEM) {
			installer.log.Info("The Bridge TLS certificate is installed in the OS keychain")
		} else {
			installer.log.Info("The Bridge TLS certificate is not installed in the OS keychain")
		}
	}
}
