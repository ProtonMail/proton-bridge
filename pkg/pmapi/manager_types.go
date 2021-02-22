package pmapi

import (
	"context"
	"net/http"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
)

type Manager interface {
	NewClient(string, string, string, time.Time) Client
	NewClientWithRefresh(context.Context, string, string) (Client, *Auth, error)
	NewClientWithLogin(context.Context, string, string) (Client, *Auth, error)

	DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error)
	ReportBug(context.Context, ReportBugReq) error
	SendSimpleMetric(context.Context, string, string, string) error

	SetLogger(resty.Logger)
	SetTransport(http.RoundTripper)
	SetCookieJar(http.CookieJar)
	SetRetryCount(int)
	AddConnectionObserver(ConnectionObserver)
}
