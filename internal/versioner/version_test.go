// Copyright (c) 2020 Proton Technologies AG
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

package versioner

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyFiles(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "verify-test")
	require.NoError(t, err)

	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    tempDir,
	}

	kr := createSignedFiles(t,
		filepath.Join(tempDir, "f1.txt"),
		filepath.Join(tempDir, "f2.png"),
		filepath.Join(tempDir, "f3.dat"),
		filepath.Join(tempDir, "sub", "f4.tar"),
		filepath.Join(tempDir, "sub", "f5.tgz"),
	)

	assert.NoError(t, version.VerifyFiles(kr))
}

func TestVerifyWithBadFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "verify-test")
	require.NoError(t, err)

	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    tempDir,
	}

	kr := createSignedFiles(t,
		filepath.Join(tempDir, "f1.txt"),
		filepath.Join(tempDir, "f2.png"),
		filepath.Join(tempDir, "f3.bad"),
		filepath.Join(tempDir, "sub", "f4.tar"),
		filepath.Join(tempDir, "sub", "f5.tgz"),
	)

	badKeyRing := makeKeyRing(t)
	signFile(t, filepath.Join(tempDir, "f3.bad"), badKeyRing)

	assert.Error(t, version.VerifyFiles(kr))
}

func TestVerifyWithBadSubFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "verify-test")
	require.NoError(t, err)

	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    tempDir,
	}

	kr := createSignedFiles(t,
		filepath.Join(tempDir, "f1.txt"),
		filepath.Join(tempDir, "f2.png"),
		filepath.Join(tempDir, "f3.dat"),
		filepath.Join(tempDir, "sub", "f4.tar"),
		filepath.Join(tempDir, "sub", "f5.bad"),
	)

	badKeyRing := makeKeyRing(t)
	signFile(t, filepath.Join(tempDir, "sub", "f5.bad"), badKeyRing)

	assert.Error(t, version.VerifyFiles(kr))
}

func createSignedFiles(t *testing.T, paths ...string) *crypto.KeyRing {
	kr := makeKeyRing(t)

	for _, path := range paths {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
		makeFile(t, path)
		signFile(t, path, kr)
	}

	return kr
}

func makeKeyRing(t *testing.T) *crypto.KeyRing {
	key, err := crypto.GenerateKey("name", "email", "rsa", 2048)
	require.NoError(t, err)

	kr, err := crypto.NewKeyRing(key)
	require.NoError(t, err)

	return kr
}

func makeFile(t *testing.T, path string) {
	f, err := os.Create(path)
	require.NoError(t, err)

	data := make([]byte, 64)
	_, err = rand.Read(data)
	require.NoError(t, err)

	_, err = f.Write(data)
	require.NoError(t, err)

	require.NoError(t, f.Close())
}

func signFile(t *testing.T, path string, kr *crypto.KeyRing) {
	file, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	sig, err := kr.SignDetached(crypto.NewPlainMessage(file))
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(path+".sig", sig.GetBinary(), 0700))
}
