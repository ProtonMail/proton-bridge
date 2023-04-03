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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package versioner

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/sum"
	"github.com/ProtonMail/proton-bridge/v3/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyFiles(t *testing.T) {
	dir := t.TempDir()
	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    dir,
	}

	kr := createSignedFiles(t, dir,
		"f1.txt",
		"f2.png",
		"f3.dat",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.tgz"),
	)

	assert.NoError(t, version.VerifyFiles(kr))
}

func TestVerifyWithBadFile(t *testing.T) {
	dir := t.TempDir()

	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    dir,
	}

	kr := createSignedFiles(t, dir,
		"f1.txt",
		"f2.png",
		"f3.bad",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.tgz"),
	)

	signFile(t, filepath.Join(dir, "f3.bad"), utils.MakeKeyRing(t))

	assert.Error(t, version.VerifyFiles(kr))
}

func TestVerifyWithBadSubFile(t *testing.T) {
	dir := t.TempDir()

	version := &Version{
		version: semver.MustParse("1.2.3"),
		path:    dir,
	}

	kr := createSignedFiles(t, dir,
		"f1.txt",
		"f2.png",
		"f3.dat",
		filepath.Join("sub", "f4.tar"),
		filepath.Join("sub", "f5.bad"),
	)

	signFile(t, filepath.Join(dir, "sub", "f5.bad"), utils.MakeKeyRing(t))

	assert.Error(t, version.VerifyFiles(kr))
}

func createSignedFiles(t *testing.T, root string, paths ...string) *crypto.KeyRing {
	kr := utils.MakeKeyRing(t)

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

	require.NoError(t, sumFile.Close())

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
	file, err := os.ReadFile(path)
	require.NoError(t, err)

	sig, err := kr.SignDetached(crypto.NewPlainMessage(file))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path+".sig", sig.GetBinary(), 0o700))
}
