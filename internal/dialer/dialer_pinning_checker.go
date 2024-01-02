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

package dialer

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
)

// ErrTLSMismatch indicates that no TLS fingerprint match could be found.
var ErrTLSMismatch = errors.New("no TLS fingerprint match found")

type TLSPinChecker struct {
	trustedPins []string
}

func NewTLSPinChecker(trustedPins []string) *TLSPinChecker {
	return &TLSPinChecker{
		trustedPins: trustedPins,
	}
}

func certFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf(`pin-sha256=%q`, algo.HashBase64SHA256(string(cert.RawSubjectPublicKeyInfo)))
}
