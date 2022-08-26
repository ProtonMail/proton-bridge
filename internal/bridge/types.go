package bridge

import (
	"context"
	"net"

	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

type Locator interface {
	ProvideSettingsPath() (string, error)
	ProvideLogsPath() (string, error)
	GetLicenseFilePath() string
	GetDependencyLicensesLink() string
}

type Identifier interface {
	GetUserAgent() string
	HasClient() bool
	SetClient(name, version string)
	SetPlatform(platform string)
}

type TLSReporter interface {
	GetTLSIssueCh() <-chan struct{}
}

type ProxyDialer interface {
	DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error)

	AllowProxy()
	DisallowProxy()
}

type Autostarter interface {
	Enable() error
	Disable() error
}

type Updater interface {
	GetVersionInfo(downloader updater.Downloader, channel updater.Channel) (updater.VersionInfo, error)
	InstallUpdate(downloader updater.Downloader, update updater.VersionInfo) error
}
