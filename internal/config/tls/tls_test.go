// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tls

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTLSKeyRenewal(t *testing.T) {
	// Remove keys.
	configPath := "/tmp"
	certPath := filepath.Join(configPath, "cert.pem")
	keyPath := filepath.Join(configPath, "key.pem")
	_ = os.Remove(certPath)
	_ = os.Remove(keyPath)

	dir, err := ioutil.TempDir("", "test-tls")
	require.NoError(t, err)

	tls := New(dir)

	// Put old key there.
	tlsTemplate.NotBefore = time.Now().Add(-365 * 24 * time.Hour)
	tlsTemplate.NotAfter = time.Now()
	cert, err := tls.GenerateConfig()
	require.Equal(t, err, ErrTLSCertExpireSoon)
	require.Equal(t, len(cert.Certificates), 1)
	time.Sleep(time.Second)
	now, notValidAfter := time.Now(), cert.Certificates[0].Leaf.NotAfter
	require.True(t, now.After(notValidAfter), "old certificate expected to not be valid at %v but have valid until %v", now, notValidAfter)

	// Renew key.
	tlsTemplate.NotBefore = time.Now()
	tlsTemplate.NotAfter = time.Now().Add(2 * 365 * 24 * time.Hour)
	cert, err = tls.GetConfig()
	if runtime.GOOS != "darwin" { // Darwin is not supported.
		require.NoError(t, err)
	}
	require.Equal(t, len(cert.Certificates), 1)
	now, notValidAfter = time.Now(), cert.Certificates[0].Leaf.NotAfter
	require.False(t, now.After(notValidAfter), "new certificate expected to be valid at %v but have valid until %v", now, notValidAfter)
}
