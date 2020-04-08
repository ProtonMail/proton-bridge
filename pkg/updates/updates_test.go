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

package updates

import (
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const testServerPort = "8999"

var testUpdateDir string //nolint[gochecknoglobals]

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	var err error
	testUpdateDir, err = ioutil.TempDir("", "upgrade")
	if err != nil {
		panic(err)
	}

	Host = "http://localhost:" + testServerPort
	go startServer()
}

func shutdown() {
	_ = os.RemoveAll(testUpdateDir)
}

func startServer() {
	http.HandleFunc("/download/current_version_linux.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/current_version_linux.json")
	})
	http.HandleFunc("/download/current_version_linux.json.sig", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/current_version_linux.json.sig")
	})
	http.HandleFunc("/download/current_version_darwin.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/current_version_linux.json")
	})
	http.HandleFunc("/download/current_version_darwin.json.sig", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/current_version_linux.json.sig")
	})
	panic(http.ListenAndServe(":"+testServerPort, nil))
}

func TestCheckBridgeIsUpToDate(t *testing.T) {
	updates := newTestUpdates("1.1.6")
	isUpToDate, _, err := updates.CheckIsBridgeUpToDate()
	require.NoError(t, err)
	require.True(t, isUpToDate, "Bridge should be up to date")
}

func TestCheckBridgeIsNotUpToDate(t *testing.T) {
	updates := newTestUpdates("1.1.5")
	isUpToDate, _, err := updates.CheckIsBridgeUpToDate()
	require.NoError(t, err)
	require.True(t, !isUpToDate, "Bridge should not be up to date")
}

func TestGetLocalVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test because local version for windows is currently not supported by tests.")
	}
	updates := newTestUpdates("1")
	expectedVersion := VersionInfo{
		Version:          "1",
		Revision:         "rev123",
		ReleaseDate:      "42",
		ReleaseNotes:     "• new feature",
		ReleaseFixedBugs: "• fixed foo",
		FixedBugs:        []string{"• fixed foo"},
		URL:              Host + "/" + DownloadPath + "/Bridge-Installer.sh",

		LandingPage:   Host + "/bridge/download",
		UpdateFile:    Host + "/" + DownloadPath + "/bridge_upgrade_linux.tgz",
		InstallerFile: Host + "/" + DownloadPath + "/Bridge-Installer.sh",

		DebFile: Host + "/" + DownloadPath + "/protonmail-bridge_1-1_amd64.deb",
		RpmFile: Host + "/" + DownloadPath + "/protonmail-bridge-1-1.x86_64.rpm",
		PkgFile: Host + "/" + DownloadPath + "/PKGBUILD",
	}
	if runtime.GOOS == "darwin" {
		expectedVersion.URL = Host + "/" + DownloadPath + "/Bridge-Installer.dmg"
		expectedVersion.UpdateFile = Host + "/" + DownloadPath + "/bridge_upgrade_darwin.tgz"
		expectedVersion.InstallerFile = expectedVersion.URL
		expectedVersion.DebFile = ""
		expectedVersion.RpmFile = ""
		expectedVersion.PkgFile = ""
	}
	version := updates.GetLocalVersion()
	require.Equal(t, expectedVersion, version)
}

func TestGetLatestVersion(t *testing.T) {
	updates := newTestUpdates("1")
	expectedVersion := VersionInfo{
		Version:          "1.1.6",
		Revision:         "",
		ReleaseDate:      "10 Jul 19 11:02 +0200",
		ReleaseNotes:     "• Necessary updates reflecting API changes\n• Report wrongly formated messages\n",
		ReleaseFixedBugs: "• Fixed verification for contacts signed by older or missing key\n• Outlook always shows attachment icon\n",
		FixedBugs: []string{
			"• Fixed verification for contacts signed by older or missing key",
			"• Outlook always shows attachment icon",
			"",
		},
		URL: "https://protonmail.com/download/Bridge-Installer.sh",

		LandingPage:   "https://protonmail.com/bridge/download",
		UpdateFile:    "https://protonmail.com/download/bridge_upgrade_linux.tgz",
		InstallerFile: "https://protonmail.com/download/Bridge-Installer.sh",

		DebFile: "https://protonmail.com/download/protonmail-bridge_1.1.6-1_amd64.deb",
		RpmFile: "https://protonmail.com/download/protonmail-bridge-1.1.6-1.x86_64.rpm",
		PkgFile: "https://protonmail.com/download/PKGBUILD",
	}
	version, err := updates.getLatestVersion()
	require.NoError(t, err)
	require.Equal(t, expectedVersion, version)
}

func TestStartUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	if runtime.GOOS != "windows" {
		t.Skip("skipping test because only upgrading on windows is currently supported by tests.")
	}

	updates := newTestUpdates("1")
	progress := make(chan Progress, 1)
	done := make(chan error)

	go func() {
		for current := range progress {
			log.Infof("progress descr: %d processed %f err %v", current.Description, current.Processed, current.Err)
			if current.Err != nil {
				done <- current.Err
				break
			}
		}
		done <- nil
	}()

	updates.StartUpgrade(progress)
	close(progress)
	require.NoError(t, <-done)
}

func newTestUpdates(version string) *Updates {
	return New("bridge", version, "rev123", "42", "• new feature", "• fixed foo", testUpdateDir)
}
