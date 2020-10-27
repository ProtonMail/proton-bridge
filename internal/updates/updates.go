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
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/pkg/constants"
	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"
)

const (
	sigExtension = ".sig"
)

var (
	Host         = "https://protonmail.com" //nolint[gochecknoglobals]
	DownloadPath = "download"               //nolint[gochecknoglobals]

	// BuildType specifies type of build (e.g. QA or beta).
	BuildType = "" //nolint[gochecknoglobals]
)

var (
	log = logrus.WithField("pkg", "bridgeUtils/updates") //nolint[gochecknoglobals]

	ErrDownloadFailed     = errors.New("error happened during download") //nolint[gochecknoglobals]
	ErrUpdateVerifyFailed = errors.New("cannot verify signature")        //nolint[gochecknoglobals]
)

type Updates struct {
	version             string
	revision            string
	buildTime           string
	releaseNotes        string
	releaseFixedBugs    string
	updateTempDir       string
	landingPagePath     string       // Based on Host/; default landing page for download.
	winInstallerFile    string       // File for initial install or manual reinstall for windows
	macInstallerFile    string       // File for initial install or manual reinstall for mac
	linInstallerFile    string       // File for initial install or manual reinstall for linux
	versionFileBaseName string       // Text file containing information about current file. per goos [_linux,_darwin,_windows].json (have .sig file).
	updateFileBaseName  string       // File for automatic update. per goos [_linux,_darwin,_windows].tgz  (have .sig file).
	linuxFileBaseName   string       // Prefix of linux package names.
	macAppBundleName    string       // Name of Mac app file in the bundle for update procedure.
	cachedNewerVersion  *VersionInfo // To have info about latest version even when the internet connection drops.
}

// NewBridge inits Updates struct for bridge.
func NewBridge(updateTempDir string) *Updates {
	return &Updates{
		version:             constants.Version,
		revision:            constants.Revision,
		buildTime:           constants.BuildTime,
		releaseNotes:        bridge.ReleaseNotes,
		releaseFixedBugs:    bridge.ReleaseFixedBugs,
		updateTempDir:       updateTempDir,
		landingPagePath:     "bridge/download",
		winInstallerFile:    "Bridge-Installer.exe",
		macInstallerFile:    "Bridge-Installer.dmg",
		linInstallerFile:    "Bridge-Installer.sh",
		versionFileBaseName: "current_version",
		updateFileBaseName:  "bridge_upgrade",
		linuxFileBaseName:   "protonmail-bridge",
		macAppBundleName:    "ProtonMail Bridge.app",
	}
}

// NewImportExport inits Updates struct for import-export.
func NewImportExport(updateTempDir string) *Updates {
	return &Updates{
		version:             constants.Version,
		revision:            constants.Revision,
		buildTime:           constants.BuildTime,
		releaseNotes:        importexport.ReleaseNotes,
		releaseFixedBugs:    importexport.ReleaseFixedBugs,
		updateTempDir:       updateTempDir,
		landingPagePath:     "import-export",
		winInstallerFile:    "ie/Import-Export-app-installer.exe",
		macInstallerFile:    "ie/Import-Export-app.dmg",
		linInstallerFile:    "ie/Import-Export-app-installer.sh",
		versionFileBaseName: "current_version_ie",
		updateFileBaseName:  "ie/ie_upgrade",
		linuxFileBaseName:   "ie/protonmail-import-export-app",
		macAppBundleName:    "ProtonMail Import-Export app.app",
	}
}

func (u *Updates) CreateJSONAndSign(deployDir, goos string) error {
	versionInfo := u.getLocalVersion(goos)
	versionInfo.Version = sanitizeVersion(versionInfo.Version)

	versionFileName := filepath.Base(u.versionFileURL(goos))
	versionFilePath := filepath.Join(deployDir, versionFileName)

	txt, err := json.Marshal(versionInfo)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(versionFilePath, txt, 0600); err != nil {
		return err
	}

	if err := singAndVerify(versionFilePath); err != nil {
		return err
	}

	updateFileName := filepath.Base(versionInfo.UpdateFile)
	updateFilePath := filepath.Join(deployDir, updateFileName)
	if err := singAndVerify(updateFilePath); err != nil {
		return err
	}

	return nil
}

func (u *Updates) CheckIsUpToDate() (isUpToDate bool, latestVersion VersionInfo, err error) {
	localVersion := u.GetLocalVersion()
	latestVersion, err = u.getLatestVersion()
	if err != nil {
		return
	}

	localIsOld, err := isFirstVersionNewer(latestVersion.Version, localVersion.Version)
	return !localIsOld, latestVersion, err
}

func (u *Updates) GetDownloadLink() string {
	latestVersion, err := u.getLatestVersion()
	if err != nil || latestVersion.InstallerFile == "" {
		localVersion := u.GetLocalVersion()
		return localVersion.GetDownloadLink()
	}
	return latestVersion.GetDownloadLink()
}

func (u *Updates) GetLocalVersion() VersionInfo {
	return u.getLocalVersion(runtime.GOOS)
}

func (u *Updates) getLocalVersion(goos string) VersionInfo {
	version := u.version
	if BuildType != "" {
		version += " " + BuildType
	}

	versionInfo := VersionInfo{
		Version:          version,
		Revision:         u.revision,
		ReleaseDate:      u.buildTime,
		ReleaseNotes:     u.releaseNotes,
		ReleaseFixedBugs: u.releaseFixedBugs,
		FixedBugs:        strings.Split(u.releaseFixedBugs, "\n"),
		URL:              u.installerFileURL(goos),

		LandingPage:   u.landingPageURL(),
		UpdateFile:    u.updateFileURL(goos),
		InstallerFile: u.installerFileURL(goos),
	}

	if goos == "linux" {
		pkgName := u.linuxFileBaseName
		pkgRel := "1"
		pkgBaseFile := strings.Join([]string{Host, DownloadPath, pkgName}, "/")

		pkgBasePath := DownloadPath + "/" + pkgName // add at least one dir
		pkgBasePath = filepath.Dir(pkgBasePath)     // keep only last dir
		pkgBasePath = Host + "/" + pkgBasePath      // add host in the end to not strip off double slash in URL

		versionInfo.DebFile = pkgBaseFile + "_" + u.version + "-" + pkgRel + "_amd64.deb"
		versionInfo.RpmFile = pkgBaseFile + "-" + u.version + "-" + pkgRel + ".x86_64.rpm"
		versionInfo.PkgFile = strings.Join([]string{pkgBasePath, "PKGBUILD"}, "/")
	}

	return versionInfo
}

func (u *Updates) getLatestVersion() (latestVersion VersionInfo, err error) {
	version, err := downloadToBytes(u.versionFileURL(runtime.GOOS))
	if err != nil {
		if u.cachedNewerVersion != nil {
			return *u.cachedNewerVersion, nil
		}
		return
	}

	signature, err := downloadToBytes(u.signatureFileURL(runtime.GOOS))
	if err != nil {
		if u.cachedNewerVersion != nil {
			return *u.cachedNewerVersion, nil
		}
		return
	}

	if err = verifyBytes(bytes.NewReader(version), bytes.NewReader(signature)); err != nil {
		return
	}

	if err = json.NewDecoder(bytes.NewReader(version)).Decode(&latestVersion); err != nil {
		return
	}
	if localIsOld, _ := isFirstVersionNewer(latestVersion.Version, u.version); localIsOld {
		u.cachedNewerVersion = &latestVersion
	}
	return
}

func (u *Updates) landingPageURL() string {
	return strings.Join([]string{Host, u.landingPagePath}, "/")
}

func (u *Updates) signatureFileURL(goos string) string {
	return u.versionFileURL(goos) + sigExtension
}

func (u *Updates) versionFileURL(goos string) string {
	return strings.Join([]string{Host, DownloadPath, u.versionFileBaseName + "_" + goos + ".json"}, "/")
}

func (u *Updates) installerFileURL(goos string) string {
	installerFile := u.linInstallerFile
	switch goos {
	case "darwin": //nolint[goconst]
		installerFile = u.macInstallerFile
	case "windows": //nolint[goconst]
		installerFile = u.winInstallerFile
	}
	return strings.Join([]string{Host, DownloadPath, installerFile}, "/")
}

func (u *Updates) updateFileURL(goos string) string {
	return strings.Join([]string{Host, DownloadPath, u.updateFileBaseName + "_" + goos + ".tgz"}, "/")
}

func (u *Updates) StartUpgrade(currentStatus chan<- Progress) { // nolint[funlen]
	status := &Progress{channel: currentStatus}
	defer status.Update()

	// Get latest version.
	var verInfo VersionInfo
	status.UpdateDescription(InfoCurrentVersion)
	if verInfo, status.Err = u.getLatestVersion(); status.Err != nil {
		return
	}

	if verInfo.UpdateFile == "" {
		log.Warn("Empty update URL. Update manually.")
		status.Err = ErrDownloadFailed
		return
	}

	// Download.
	status.UpdateDescription(InfoDownloading)
	if status.Err = mkdirAllClear(u.updateTempDir); status.Err != nil {
		return
	}
	var updateTar string
	updateTar, status.Err = downloadWithSignature(
		status,
		verInfo.UpdateFile,
		u.updateTempDir,
	)
	if status.Err != nil {
		return
	}

	// Check signature.
	status.UpdateDescription(InfoVerifying)
	status.Err = verifyFile(updateTar)
	if status.Err != nil {
		log.Warnf("Cannot verify update file %s: %v", updateTar, status.Err)
		status.Err = ErrUpdateVerifyFailed
		return
	}

	// Untar.
	status.UpdateDescription(InfoUnpacking)
	status.Err = untarToDir(updateTar, u.updateTempDir, status)
	if status.Err != nil {
		return
	}

	// Run upgrade (OS specific).
	status.UpdateDescription(InfoUpgrading)
	switch runtime.GOOS {
	case "windows": //nolint[goconst]
		// Cannot use filepath.Base on windows it has different delimiter
		split := strings.Split(u.winInstallerFile, "/")
		installerFile := split[len(split)-1]
		cmd := exec.Command("./" + installerFile) // nolint[gosec]
		cmd.Dir = u.updateTempDir
		status.Err = cmd.Start()
	case "darwin": //nolint[goconst]
		// current path is better then appDir = filepath.Join("/Applications")
		var exePath string
		exePath, status.Err = osext.Executable()
		if status.Err != nil {
			return
		}
		localPath := filepath.Dir(exePath)  // Macos
		localPath = filepath.Dir(localPath) // Contents
		localPath = filepath.Dir(localPath) // .app

		updatePath := filepath.Join(u.updateTempDir, u.macAppBundleName)
		log.WithField("local", localPath).
			WithField("update", updatePath).
			Info("Syncing folders..")
		status.Err = syncFolders(localPath, updatePath)
		if status.Err != nil {
			log.WithField("from", localPath).
				WithField("to", updatePath).
				WithError(status.Err).
				Error("Sync failed.")
			return
		}
		status.UpdateDescription(InfoRestartApp)
		return
	default:
		status.Err = errors.New("upgrade for " + runtime.GOOS + " not implemented")
	}

	status.UpdateDescription(InfoQuitApp)
}
