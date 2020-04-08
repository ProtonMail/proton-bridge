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

package pmapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/jaytaylor/html2text"
	"github.com/sirupsen/logrus"
)

// Version of the API.
const Version = 3

// API return codes.
const (
	ForceUpgradeBadAPIVersion = 5003
	ForceUpgradeInvalidAPI    = 5004
	ForceUpgradeBadAppVersion = 5005
	APIOffline                = 7001
	ImportMessageTooLong      = 36022
	BansRequests              = 85131
)

// The output errors.
var (
	ErrInvalidToken       = errors.New("refresh token invalid")
	ErrAPINotReachable    = errors.New("cannot reach the server")
	ErrUpgradeApplication = errors.New("application upgrade required")
)

type ErrUnauthorized struct {
	error
}

func (err *ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized access: %+v", err.error.Error())
}

type TokenManager struct {
	tokensLocker sync.Locker
	tokenMap     map[string]string
}

func NewTokenManager() *TokenManager {
	tm := &TokenManager{
		tokensLocker: &sync.Mutex{},
		tokenMap:     map[string]string{},
	}
	return tm
}

func (tm *TokenManager) GetToken(userID string) string {
	tm.tokensLocker.Lock()
	defer tm.tokensLocker.Unlock()

	return tm.tokenMap[userID]
}

func (tm *TokenManager) SetToken(userID, token string) {
	tm.tokensLocker.Lock()
	defer tm.tokensLocker.Unlock()

	tm.tokenMap[userID] = token
}

// ClientConfig contains Client configuration.
type ClientConfig struct {
	// The client application name and version.
	AppVersion string

	// The client ID.
	ClientID string

	TokenManager *TokenManager

	// Transport specifies the mechanism by which individual HTTP requests are made.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	// Timeout specifies the timeout from request to getting response headers to our API.
	// Passed to http.Client, empty means no timeout.
	Timeout time.Duration

	// FirstReadTimeout specifies the timeout from getting response to the first read of body response.
	// This timeout is applied only when MinSpeed is used.
	// Default is 5 minutes.
	FirstReadTimeout time.Duration

	// MinSpeed specifies minimum Bytes per second or the request will be canceled.
	// Zero means no limitation.
	MinSpeed int64
}

// Client to communicate with API.
type Client struct {
	auths chan<- *Auth // Channel that sends Auth responses back to the bridge.

	log    *logrus.Entry
	config *ClientConfig
	client *http.Client
	conrep ConnectionReporter

	uid           string
	accessToken   string
	userID        string // Twice here because Username is not unique.
	requestLocker sync.Locker
	keyLocker     sync.Locker

	tokenManager *TokenManager
	expiresAt    time.Time
	user         *User
	addresses    AddressList
	kr           *pmcrypto.KeyRing
}

// NewClient creates a new API client.
func NewClient(cfg *ClientConfig, userID string) *Client {
	hc := &http.Client{
		Timeout: cfg.Timeout,
	}
	if cfg.Transport != nil {
		cfgTransport, ok := cfg.Transport.(*http.Transport)
		if ok {
			// In future use Clone here.
			// https://go-review.googlesource.com/c/go/+/174597/
			transport := &http.Transport{}
			*transport = *cfgTransport //nolint
			if transport.Proxy == nil {
				transport.Proxy = http.ProxyFromEnvironment
			}
			hc.Transport = transport
		} else {
			hc.Transport = cfg.Transport
		}
	} else if defaultTransport != nil {
		hc.Transport = defaultTransport
	}

	log := logrus.WithFields(logrus.Fields{
		"pkg":    "pmapi",
		"userID": userID,
	})

	return &Client{
		log:           log,
		config:        cfg,
		client:        hc,
		tokenManager:  cfg.TokenManager,
		userID:        userID,
		requestLocker: &sync.Mutex{},
		keyLocker:     &sync.Mutex{},
	}
}

// SetConnectionReporter sets the connection reporter used by the client to report when
// internet connection is lost.
func (c *Client) SetConnectionReporter(conrep ConnectionReporter) {
	c.conrep = conrep
}

// reportLostConnection reports that the internet connection has been lost using the connection reporter.
// If the connection reporter has not been set, this does nothing.
func (c *Client) reportLostConnection() {
	if c.conrep != nil {
		err := c.conrep.NotifyConnectionLost()
		if err != nil {
			logrus.WithError(err).Error("Failed to notify of lost connection")
		}
	}
}

// Do makes an API request. It does not check for HTTP status code errors.
func (c *Client) Do(req *http.Request, retryUnauthorized bool) (res *http.Response, err error) {
	// Copy the request body in case we need to retry it.
	var bodyBuffer []byte
	if req.Body != nil {
		defer req.Body.Close() //nolint[errcheck]
		bodyBuffer, err = ioutil.ReadAll(req.Body)

		if err != nil {
			return nil, err
		}

		r := bytes.NewReader(bodyBuffer)
		req.Body = ioutil.NopCloser(r)
	}

	return c.doBuffered(req, bodyBuffer, retryUnauthorized)
}

// If needed it retries using req and buffered body.
func (c *Client) doBuffered(req *http.Request, bodyBuffer []byte, retryUnauthorized bool) (res *http.Response, err error) { // nolint[funlen]
	isAuthReq := strings.Contains(req.URL.Path, "/auth")

	req.Header.Set("x-pm-appversion", c.config.AppVersion)
	req.Header.Set("x-pm-apiversion", strconv.Itoa(Version))

	if c.uid != "" {
		req.Header.Set("x-pm-uid", c.uid)
	}

	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	c.log.Debugln("Requesting ", req.Method, req.URL.RequestURI())
	if logrus.GetLevel() == logrus.TraceLevel {
		head := ""
		for i, v := range req.Header {
			head += i + ": "
			head += strings.Join(v, "")
			head += "\n"
		}
		c.log.Tracef("REQHEAD \n%s", head)
		c.log.Tracef("REQBODY '%s'", string(bodyBuffer))
	}

	hasBody := len(bodyBuffer) > 0
	if res, err = c.client.Do(req); err != nil {
		if res == nil {
			c.log.WithError(err).Error("Cannot get response")
			err = ErrAPINotReachable
			c.reportLostConnection()
		}
		return
	}

	resDate := res.Header.Get("Date")
	if resDate != "" {
		if serverTime, err := http.ParseTime(resDate); err == nil {
			pmcrypto.GetGopenPGP().UpdateTime(serverTime.Unix())
		}
	}

	if res.StatusCode == http.StatusUnauthorized {
		if hasBody {
			r := bytes.NewReader(bodyBuffer)
			req.Body = ioutil.NopCloser(r)
		}

		if !isAuthReq {
			_, _ = io.Copy(ioutil.Discard, res.Body)
			_ = res.Body.Close()
			return c.handleStatusUnauthorized(req, bodyBuffer, res, retryUnauthorized)
		}
	}

	// Retry induced by HTTP status code>
	retryAfter := 10
	doRetry := res.StatusCode == http.StatusTooManyRequests
	if doRetry {
		if headerAfter, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && headerAfter > 0 {
			retryAfter = headerAfter
		}
		// To avoid spikes when all clients retry at the same time, we add some random wait.
		retryAfter += rand.Intn(10)

		if hasBody {
			r := bytes.NewReader(bodyBuffer)
			req.Body = ioutil.NopCloser(r)
		}

		c.log.Warningf("Retrying %s after %ds induced by http code %d", req.URL.Path, retryAfter, res.StatusCode)
		time.Sleep(time.Duration(retryAfter) * time.Second)
		_, _ = io.Copy(ioutil.Discard, res.Body)
		_ = res.Body.Close()
		return c.doBuffered(req, bodyBuffer, false)
	}

	return res, err
}

// DoJSON performs the request and unmarshals the response as JSON into data.
// If the API returns a non-2xx HTTP status code, the error returned will contain status
// and response as plaintext. API errors must be checked by the caller.
// It is performed buffered, in case we need to retry.
func (c *Client) DoJSON(req *http.Request, data interface{}) error {
	// Copy the request body in case we need to retry it
	var reqBodyBuffer []byte

	if req.Body != nil {
		defer req.Body.Close() //nolint[errcheck]
		var err error
		if reqBodyBuffer, err = ioutil.ReadAll(req.Body); err != nil {
			return err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(reqBodyBuffer))
	}

	return c.doJSONBuffered(req, reqBodyBuffer, data)
}

// doJSONBuffered performs a buffered json request (see DoJSON for more information).
func (c *Client) doJSONBuffered(req *http.Request, reqBodyBuffer []byte, data interface{}) error { // nolint[funlen]
	req.Header.Set("Accept", "application/vnd.protonmail.v1+json")

	var cancelRequest context.CancelFunc
	if c.config.MinSpeed > 0 {
		var ctx context.Context
		ctx, cancelRequest = context.WithCancel(req.Context())
		defer func() {
			cancelRequest()
		}()
		req = req.WithContext(ctx)
	}

	res, err := c.doBuffered(req, reqBodyBuffer, false)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint[errcheck]

	var resBody []byte
	if c.config.MinSpeed == 0 {
		resBody, err = ioutil.ReadAll(res.Body)
	} else {
		resBody, err = c.readAllMinSpeed(res.Body, cancelRequest)
	}

	// The server response may contain data which we want to have in memory
	// for as little time as possible (such as keys). Go is garbage collected,
	// so we are not in charge of when the memory will actually be cleared.
	// We can at least try to rewrite the original data to mitigate this problem.
	defer func() {
		for i := 0; i < len(resBody); i++ {
			resBody[i] = byte(65)
		}
	}()

	if logrus.GetLevel() == logrus.TraceLevel {
		head := ""
		for i, v := range res.Header {
			head += i + ": "
			head += strings.Join(v, "")
			head += "\n"
		}
		c.log.Tracef("RESHEAD \n%s", head)
		c.log.Tracef("RESBODY '%s'", resBody)
	}

	if err != nil {
		return err
	}

	// Retry induced by API code.
	errCode := &Res{}
	if err := json.Unmarshal(resBody, errCode); err == nil {
		if errCode.Code == BansRequests {
			retryAfter := 3
			c.log.Warningf("Retrying %s after %ds induced by API code %d", req.URL.Path, retryAfter, errCode.Code)
			time.Sleep(time.Duration(retryAfter) * time.Second)
			if len(reqBodyBuffer) > 0 {
				req.Body = ioutil.NopCloser(bytes.NewReader(reqBodyBuffer))
			}
			return c.doJSONBuffered(req, reqBodyBuffer, data)
		}
	}

	if err := json.Unmarshal(resBody, data); err != nil {
		// Check to see if this is due to a non 2xx HTTP status code.
		if res.StatusCode != http.StatusOK {
			r := bytes.NewReader(bytes.ReplaceAll(resBody, []byte("\n"), []byte("\\n")))
			plaintext, err := html2text.FromReader(r)
			if err == nil {
				return fmt.Errorf("Error: \n\n" + res.Status + "\n\n" + plaintext)
			}
		}

		if errJS, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("invalid json %v (offset:%d) ", errJS.Error(), errJS.Offset)
		}

		return fmt.Errorf("unmarshal fail: %v ", err)
	}

	// Set StatusCode in case data struct supports that field.
	// It's safe to set StatusCode, server returns Code. StatusCode should be preferred over Code.
	dataValue := reflect.ValueOf(data).Elem()
	statusCodeField := dataValue.FieldByName("StatusCode")
	if statusCodeField.IsValid() && statusCodeField.CanSet() && statusCodeField.Kind() == reflect.Int {
		statusCodeField.SetInt(int64(res.StatusCode))
	}

	if res.StatusCode != http.StatusOK {
		c.log.Warnf("request %s %s NOT OK: %s", req.Method, req.URL.Path, res.Status)
	}

	return nil
}

func (c *Client) readAllMinSpeed(data io.Reader, cancelRequest context.CancelFunc) ([]byte, error) {
	firstReadTimeout := c.config.FirstReadTimeout
	if firstReadTimeout == 0 {
		firstReadTimeout = 5 * time.Minute
	}
	timer := time.AfterFunc(firstReadTimeout, func() {
		cancelRequest()
	})
	var buffer bytes.Buffer
	for {
		_, err := io.CopyN(&buffer, data, c.config.MinSpeed)
		timer.Stop()
		timer.Reset(1 * time.Second)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return ioutil.ReadAll(&buffer)
}

func (c *Client) refreshAccessToken() (err error) {
	c.log.Debug("Refreshing token")
	refreshToken := c.tokenManager.GetToken(c.userID)
	c.log.WithField("token", refreshToken).Info("Current refresh token")
	if refreshToken == "" {
		if c.auths != nil {
			c.auths <- nil
		}
		if c.tokenManager != nil {
			c.tokenManager.SetToken(c.userID, "")
		}
		return ErrInvalidToken
	}

	auth, err := c.AuthRefresh(refreshToken)
	if err != nil {
		c.log.WithError(err).WithField("auths", c.auths).Debug("Token refreshing failed")
		// The refresh failed, so we should log the user out.
		// A nil value in the Auths channel will trigger this.
		if c.auths != nil {
			c.auths <- nil
		}
		if c.tokenManager != nil {
			c.tokenManager.SetToken(c.userID, "")
		}
		return
	}
	c.uid = auth.UID()
	c.accessToken = auth.accessToken
	return err
}

func (c *Client) handleStatusUnauthorized(req *http.Request, reqBodyBuffer []byte, res *http.Response, retry bool) (retryRes *http.Response, err error) {
	c.log.Info("Handling unauthorized status")

	// If this is not a retry, then it is the first time handling status unauthorized,
	// so try again without refreshing the access token.
	if !retry {
		c.log.Debug("Handling unauthorized status by retrying")
		c.requestLocker.Lock()
		defer c.requestLocker.Unlock()

		_, _ = io.Copy(ioutil.Discard, res.Body)
		_ = res.Body.Close()
		return c.doBuffered(req, reqBodyBuffer, true)
	}

	// This is already a retry, so we will try to refresh the access token before trying again.
	if err = c.refreshAccessToken(); err != nil {
		c.log.WithError(err).Warn("Cannot refresh token")
		err = &ErrUnauthorized{err}
		return
	}
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		c.log.WithError(err).Warn("Failed to read out response body")
	}
	_ = res.Body.Close()
	return c.doBuffered(req, reqBodyBuffer, true)
}
