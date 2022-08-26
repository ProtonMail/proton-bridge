package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
)

var (
	ErrDownloadVerify = errors.New("failed to download or verify the update")
	ErrInstall        = errors.New("failed to install the update")
)

type Downloader interface {
	DownloadAndVerify(ctx context.Context, kr *crypto.KeyRing, url, sig string) ([]byte, error)
}

type Installer interface {
	InstallUpdate(*semver.Version, io.Reader) error
}

type Updater struct {
	installer Installer
	verifier  *crypto.KeyRing
	product   string
	platform  string
}

func NewUpdater(installer Installer, verifier *crypto.KeyRing, product, platform string) *Updater {
	return &Updater{
		installer: installer,
		verifier:  verifier,
		product:   product,
		platform:  platform,
	}
}

func (u *Updater) GetVersionInfo(downloader Downloader, channel Channel) (VersionInfo, error) {
	b, err := downloader.DownloadAndVerify(
		context.Background(),
		u.verifier,
		u.getVersionFileURL(),
		u.getVersionFileURL()+".sig",
	)
	if err != nil {
		return VersionInfo{}, err
	}

	var versionMap VersionMap

	if err := json.Unmarshal(b, &versionMap); err != nil {
		return VersionInfo{}, err
	}

	version, ok := versionMap[channel]
	if !ok {
		return VersionInfo{}, errors.New("no updates available for this channel")
	}

	return version, nil
}

func (u *Updater) InstallUpdate(downloader Downloader, update VersionInfo) error {
	b, err := downloader.DownloadAndVerify(
		context.Background(),
		u.verifier,
		update.Package,
		update.Package+".sig",
	)
	if err != nil {
		return ErrDownloadVerify
	}

	if err := u.installer.InstallUpdate(update.Version, bytes.NewReader(b)); err != nil {
		return ErrInstall
	}

	return nil
}

// getVersionFileURL returns the URL of the version file.
// For example:
//   - https://protonmail.com/download/bridge/version_linux.json
func (u *Updater) getVersionFileURL() string {
	return fmt.Sprintf("%v/%v/version_%v.json", Host, u.product, u.platform)
}
