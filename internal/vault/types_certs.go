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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import "github.com/ProtonMail/proton-bridge/v3/internal/certs"

type Certs struct {
	Bridge Cert

	// If non-empty, the path to the PEM-encoded certificate file.
	CustomCertPath string
	CustomKeyPath  string
}

type Cert struct {
	Cert, Key []byte
}

func newDefaultCerts() Certs {
	return Certs{
		Bridge: newTLSCert(),
	}
}

func newTLSCert() Cert {
	template, err := certs.NewTLSTemplate()
	if err != nil {
		panic(err)
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		panic(err)
	}

	return Cert{
		Cert: certPEM,
		Key:  keyPEM,
	}
}
