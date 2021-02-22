package context

import (
	"context"
	"net/http"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/go-resty/resty/v2"
)

func newLivePMAPIManager() pmapi.Manager {
	return pmapi.New(pmapi.DefaultConfig)
}

func newFakePMAPIManager() pmapi.Manager {
	return &fakePMAPIManager{}
}

type fakePMAPIManager struct{}

func (*fakePMAPIManager) NewClient(string, string, string, time.Time) pmapi.Client {
	panic("TODO")
}

func (*fakePMAPIManager) NewClientWithRefresh(context.Context, string, string) (pmapi.Client, *pmapi.Auth, error) {
	panic("TODO")
}

func (*fakePMAPIManager) NewClientWithLogin(context.Context, string, string) (pmapi.Client, *pmapi.Auth, error) {
	panic("TODO")
}

func (*fakePMAPIManager) DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error) {
	panic("TODO")
}

func (*fakePMAPIManager) ReportBug(context.Context, pmapi.ReportBugReq) error {
	panic("TODO")
}

func (*fakePMAPIManager) SendSimpleMetric(context.Context, string, string, string) error {
	panic("TODO")
}

func (*fakePMAPIManager) SetLogger(resty.Logger) {
	panic("TODO")
}

func (*fakePMAPIManager) SetTransport(http.RoundTripper) {
	panic("TODO")
}

func (*fakePMAPIManager) SetCookieJar(http.CookieJar) {
	panic("TODO")
}

func (*fakePMAPIManager) SetRetryCount(int) {
	panic("TODO")
}

func (*fakePMAPIManager) AddConnectionObserver(pmapi.ConnectionObserver) {
	panic("TODO")
}
