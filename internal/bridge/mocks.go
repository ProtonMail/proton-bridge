package bridge

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge/mocks"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/golang/mock/gomock"
)

type Mocks struct {
	ProxyDialer *mocks.MockProxyDialer
	TLSReporter *mocks.MockTLSReporter
	TLSIssueCh  chan struct{}

	Updater     *TestUpdater
	Autostarter *mocks.MockAutostarter
}

func NewMocks(tb testing.TB, dialer *TestDialer, version, minAuto *semver.Version) *Mocks {
	ctl := gomock.NewController(tb)

	mocks := &Mocks{
		ProxyDialer: mocks.NewMockProxyDialer(ctl),
		TLSReporter: mocks.NewMockTLSReporter(ctl),
		TLSIssueCh:  make(chan struct{}),

		Updater:     NewTestUpdater(version, minAuto),
		Autostarter: mocks.NewMockAutostarter(ctl),
	}

	// When using the proxy dialer, we want to use the test dialer.
	mocks.ProxyDialer.EXPECT().DialTLSContext(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialTLSContext(ctx, network, address)
	}).AnyTimes()

	// When getting the TLS issue channel, we want to return the test channel.
	mocks.TLSReporter.EXPECT().GetTLSIssueCh().Return(mocks.TLSIssueCh).AnyTimes()

	return mocks
}

type TestDialer struct {
	canDial bool
}

func NewTestDialer() *TestDialer {
	return &TestDialer{
		canDial: true,
	}
}

func (d *TestDialer) DialTLSContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	if !d.canDial {
		return nil, errors.New("cannot dial")
	}

	return (&tls.Dialer{Config: &tls.Config{InsecureSkipVerify: true}}).DialContext(ctx, network, address)
}

func (d *TestDialer) SetCanDial(canDial bool) {
	d.canDial = canDial
}

type TestLocationsProvider struct {
	config, cache string
}

func NewTestLocationsProvider(tb testing.TB) *TestLocationsProvider {
	return &TestLocationsProvider{
		config: tb.TempDir(),
		cache:  tb.TempDir(),
	}
}

func (provider *TestLocationsProvider) UserConfig() string {
	return provider.config
}

func (provider *TestLocationsProvider) UserCache() string {
	return provider.cache
}

type TestUpdater struct {
	latest updater.VersionInfo
}

func NewTestUpdater(version, minAuto *semver.Version) *TestUpdater {
	return &TestUpdater{
		latest: updater.VersionInfo{
			Version: version,
			MinAuto: minAuto,

			RolloutProportion: 1.0,
		},
	}
}

func (testUpdater *TestUpdater) SetLatestVersion(version, minAuto *semver.Version) {
	testUpdater.latest = updater.VersionInfo{
		Version: version,
		MinAuto: minAuto,

		RolloutProportion: 1.0,
	}
}

func (updater *TestUpdater) GetVersionInfo(downloader updater.Downloader, channel updater.Channel) (updater.VersionInfo, error) {
	return updater.latest, nil
}

func (updater *TestUpdater) InstallUpdate(downloader updater.Downloader, update updater.VersionInfo) error {
	return nil
}
