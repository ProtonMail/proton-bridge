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

package tls

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetOldConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-tls")
	require.NoError(t, err)

	// Create new tls object.
	tls := New(dir)

	// Create new TLS template.
	tlsTemplate, err := NewTLSTemplate()
	require.NoError(t, err)

	// Make the template be an old key.
	tlsTemplate.NotBefore = time.Now().Add(-365 * 24 * time.Hour)
	tlsTemplate.NotAfter = time.Now()

	// Generate the certs from the template.
	require.NoError(t, tls.GenerateCerts(tlsTemplate))

	// Generate the config from the certs -- it's going to expire soon so we don't want to use it.
	_, err = tls.GetConfig()
	require.Equal(t, err, ErrTLSCertExpiresSoon)
}

func TestGetValidConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-tls")
	require.NoError(t, err)

	// Create new tls object.
	tls := New(dir)

	// Create new TLS template.
	tlsTemplate, err := NewTLSTemplate()
	require.NoError(t, err)

	// Make the template be a new key.
	tlsTemplate.NotBefore = time.Now()
	tlsTemplate.NotAfter = time.Now().Add(2 * 365 * 24 * time.Hour)

	// Generate the certs from the template.
	require.NoError(t, tls.GenerateCerts(tlsTemplate))

	// Generate the config from the certs -- it's not going to expire soon so we want to use it.
	config, err := tls.GetConfig()
	require.NoError(t, err)
	require.Equal(t, len(config.Certificates), 1)

	// Check the cert is valid.
	now, notValidAfter := time.Now(), config.Certificates[0].Leaf.NotAfter
	require.False(t, now.After(notValidAfter), "new certificate expected to be valid at %v but have valid until %v", now, notValidAfter)
}

func TestNewConfig(t *testing.T) {
	pemCert, pemKey, err := NewPEMKeyPair()
	require.NoError(t, err)

	_, err = GetConfigFromPEMKeyPair(pemCert, pemKey)
	require.NoError(t, err)
}
