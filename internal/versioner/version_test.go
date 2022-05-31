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

package versioner

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/sum"
	tests "github.com/ProtonMail/proton-bridge/v2/test"
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

	kr := createSignedFiles(t, tempDir,
		"f1.txt",
		"f2.png",
		"f3.dat",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.tgz"),
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

	kr := createSignedFiles(t, tempDir,
		"f1.txt",
		"f2.png",
		"f3.bad",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.tgz"),
	)

	badKeyRing := tests.MakeKeyRing(t)
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

	kr := createSignedFiles(t, tempDir,
		"f1.txt",
		"f2.png",
		"f3.dat",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.bad"),
	)

	badKeyRing := tests.MakeKeyRing(t)
	signFile(t, filepath.Join(tempDir, "sub", "f5.bad"), badKeyRing)

	assert.Error(t, version.VerifyFiles(kr))
}

func createSignedFiles(t *testing.T, root string, paths ...string) *crypto.KeyRing {
	kr := tests.MakeKeyRing(t)

	for _, path := range paths {
		makeFile(t, filepath.Join(root, path))
	}

	sum, err := sum.RecursiveSum(root, "")
	require.NoError(t, err)

	sumFile, err := os.Create(filepath.Join(root, sumFile))
	require.NoError(t, err)

	_, err = sumFile.Write(sum)
	require.NoError(t, err)

	signFile(t, sumFile.Name(), kr)

	return kr
}

func makeFile(t *testing.T, path string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))

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
	require.NoError(t, ioutil.WriteFile(path+".sig", sig.GetBinary(), 0o700))
}
