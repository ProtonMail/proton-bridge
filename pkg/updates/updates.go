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

	"github.com/ProtonMail/proton-bridge/pkg/logs"
	"github.com/kardianos/osext"
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
	log = logs.GetLogEntry("bridgeUtils/updates") //nolint[gochecknoglobals]

	installFileSuffix = map[string]string{ //nolint[gochecknoglobals]
		"darwin":  ".dmg",
		"windows": ".exe",
		"linux":   ".sh",
	}

	ErrDownloadFailed     = errors.New("error happened during download") //nolint[gochecknoglobals]
	ErrUpdateVerifyFailed = errors.New("cannot verify signature")        //nolint[gochecknoglobals]
)

type Updates struct {
	appName               string
	version               string
	revision              string
	buildTime             string
	releaseNotes          string
	releaseFixedBugs      string
	updateTempDir         string
	landingPagePath       string       // Based on Host/; default landing page for download.
	installerFileBaseName string       // File for initial install or manual reinstall. per goos [exe, dmg, sh].
	versionFileBaseName   string       // Text file containing information about current file. per goos [_linux,_darwin,_windows].json (have .sig file).
	updateFileBaseName    string       // File for automatic update. per goos [_linux,_darwin,_windows].tgz  (have .sig file).
	macAppBundleName      string       // For update procedure.
	cachedNewerVersion    *VersionInfo // To have info about latest version even when the internet connection drops.
}

// New inits Updates struct.
// `appName` should be in camelCase format for file names. For installer files is converted to CamelCase.
func New(appName, version, revision, buildTime, releaseNotes, releaseFixedBugs, updateTempDir string) *Updates {
	return &Updates{
		appName:               appName,
		version:               version,
		revision:              revision,
		buildTime:             buildTime,
		releaseNotes:          releaseNotes,
		releaseFixedBugs:      releaseFixedBugs,
		updateTempDir:         updateTempDir,
		landingPagePath:       appName + "/download",
		installerFileBaseName: strings.Title(appName) + "-Installer",
		versionFileBaseName:   "current_version",
		updateFileBaseName:    appName + "_upgrade",
		macAppBundleName:      "ProtonMail " + strings.Title(appName) + ".app", // For update procedure.
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

	if err = ioutil.WriteFile(versionFilePath, txt, 0644); err != nil {
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

func (u *Updates) CheckIsBridgeUpToDate() (isUpToDate bool, latestVersion VersionInfo, err error) {
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
		pkgName := "protonmail-" + u.appName
		pkgRel := "1"
		pkgBase := strings.Join([]string{Host, DownloadPath, pkgName}, "/")

		versionInfo.DebFile = pkgBase + "_" + u.version + "-" + pkgRel + "_amd64.deb"
		versionInfo.RpmFile = pkgBase + "-" + u.version + "-" + pkgRel + ".x86_64.rpm"
		versionInfo.PkgFile = strings.Join([]string{Host, DownloadPath, "PKGBUILD"}, "/")
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
	return strings.Join([]string{Host, DownloadPath, u.installerFileBaseName + installFileSuffix[goos]}, "/")
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
		cmd := exec.Command("./" + u.installerFileBaseName) // nolint[gosec]
		cmd.Dir = u.updateTempDir
		status.Err = cmd.Start()
	case "darwin":
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
		log.Warn("localPath ", localPath)
		log.Warn("updatePath ", updatePath)
		status.Err = syncFolders(localPath, updatePath)
		if status.Err != nil {
			return
		}
		status.UpdateDescription(InfoRestartApp)
		return
	default:
		status.Err = errors.New("upgrade for " + runtime.GOOS + " not implemented")
	}

	status.UpdateDescription(InfoQuitApp)
}
