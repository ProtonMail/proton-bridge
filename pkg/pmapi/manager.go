package pmapi

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type manager struct {
	rc *resty.Client

	isDown    bool
	locker    sync.Locker
	observers []ConnectionObserver
}

func newManager(cfg Config) *manager {
	m := &manager{
		rc:     resty.New(),
		locker: &sync.Mutex{},
	}

	// Set the API host.
	m.rc.SetHostURL(cfg.HostURL)

	// Set static header values.
	m.rc.SetHeader("x-pm-appversion", cfg.AppVersion)

	// Set middleware.
	m.rc.OnAfterResponse(catchAPIError)

	// Configure retry mechanism.
	m.rc.SetRetryMaxWaitTime(time.Minute)
	m.rc.SetRetryAfter(catchRetryAfter)
	m.rc.AddRetryCondition(catchTooManyRequests)
	m.rc.AddRetryCondition(catchNoResponse)
	m.rc.AddRetryCondition(catchProxyAvailable)

	// Determine what happens when requests succeed/fail.
	m.rc.OnAfterResponse(m.handleRequestSuccess)
	m.rc.OnError(m.handleRequestFailure)

	// Set the data type of API errors.
	m.rc.SetError(&Error{})

	return m
}

func New(cfg Config) Manager {
	return newManager(cfg)
}

func (m *manager) SetLogger(logger resty.Logger) {
	m.rc.SetLogger(logger)
	m.rc.SetDebug(true)
}

func (m *manager) SetTransport(transport http.RoundTripper) {
	m.rc.SetTransport(transport)
}

func (m *manager) SetCookieJar(jar http.CookieJar) {
	m.rc.SetCookieJar(jar)
}

func (m *manager) SetRetryCount(count int) {
	m.rc.SetRetryCount(count)
}

func (m *manager) AddConnectionObserver(observer ConnectionObserver) {
	m.observers = append(m.observers, observer)
}

func (m *manager) r(ctx context.Context) *resty.Request {
	return m.rc.R().SetContext(ctx)
}

func (m *manager) handleRequestSuccess(_ *resty.Client, res *resty.Response) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if !m.isDown {
		return nil
	}

	// We successfully got a response; connection must be up.

	m.isDown = false

	for _, observer := range m.observers {
		observer.OnUp()
	}

	return nil
}

func (m *manager) handleRequestFailure(req *resty.Request, err error) {
	m.locker.Lock()
	defer m.locker.Unlock()

	if m.isDown {
		return
	}

	if res, ok := err.(*resty.ResponseError); ok && res.Response.RawResponse != nil {
		return
	}

	// We didn't get any response; connection must be down.

	m.isDown = true

	for _, observer := range m.observers {
		observer.OnDown()
	}

	go m.pingUntilSuccess()
}

func (m *manager) pingUntilSuccess() {
	for m.testPing(context.Background()) != nil {
		time.Sleep(time.Second) // TODO: How long to sleep here?
	}
}
