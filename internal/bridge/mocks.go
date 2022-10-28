package bridge

import (
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge/mocks"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/golang/mock/gomock"
)

type Mocks struct {
	ProxyCtl    *mocks.MockProxyController
	TLSReporter *mocks.MockTLSReporter
	TLSIssueCh  chan struct{}

	Updater     *TestUpdater
	Autostarter *mocks.MockAutostarter
}

func NewMocks(tb testing.TB, version, minAuto *semver.Version) *Mocks {
	ctl := gomock.NewController(tb)

	mocks := &Mocks{
		ProxyCtl:    mocks.NewMockProxyController(ctl),
		TLSReporter: mocks.NewMockTLSReporter(ctl),
		TLSIssueCh:  make(chan struct{}),

		Updater:     NewTestUpdater(version, minAuto),
		Autostarter: mocks.NewMockAutostarter(ctl),
	}

	// When getting the TLS issue channel, we want to return the test channel.
	mocks.TLSReporter.EXPECT().GetTLSIssueCh().Return(mocks.TLSIssueCh).AnyTimes()

	return mocks
}

func (mocks *Mocks) Close() {
	close(mocks.TLSIssueCh)
}

type TestCookieJar struct {
	cookies map[string][]*http.Cookie
}

func NewTestCookieJar() *TestCookieJar {
	return &TestCookieJar{
		cookies: make(map[string][]*http.Cookie),
	}
}

func (j *TestCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies[u.Host] = cookies
}

func (j *TestCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies[u.Host]
}

type TestLocationsProvider struct {
	config, data, cache string
}

func NewTestLocationsProvider(dir string) *TestLocationsProvider {
	config, err := os.MkdirTemp(dir, "config")
	if err != nil {
		panic(err)
	}

	data, err := os.MkdirTemp(dir, "data")
	if err != nil {
		panic(err)
	}

	cache, err := os.MkdirTemp(dir, "cache")
	if err != nil {
		panic(err)
	}

	return &TestLocationsProvider{
		config: config,
		data:   data,
		cache:  cache,
	}
}

func (provider *TestLocationsProvider) UserConfig() string {
	return provider.config
}

func (provider *TestLocationsProvider) UserData() string {
	return provider.data
}

func (provider *TestLocationsProvider) UserCache() string {
	return provider.cache
}

type TestUpdater struct {
	latest updater.VersionInfo
	lock   sync.RWMutex
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
	testUpdater.lock.Lock()
	defer testUpdater.lock.Unlock()

	testUpdater.latest = updater.VersionInfo{
		Version: version,
		MinAuto: minAuto,

		RolloutProportion: 1.0,
	}
}

func (testUpdater *TestUpdater) GetVersionInfo(downloader updater.Downloader, channel updater.Channel) (updater.VersionInfo, error) {
	testUpdater.lock.RLock()
	defer testUpdater.lock.RUnlock()

	return testUpdater.latest, nil
}

func (testUpdater *TestUpdater) InstallUpdate(downloader updater.Downloader, update updater.VersionInfo) error {
	return nil
}
