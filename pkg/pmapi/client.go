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

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jaytaylor/html2text"
	"github.com/pkg/errors"
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
	ErrConnectionSlow     = errors.New("request canceled because connection speed was too slow")
)

type ErrUnprocessableEntity struct {
	error
}

func (err *ErrUnprocessableEntity) Error() string {
	return err.error.Error()
}

type ErrUnauthorized struct {
	error
}

func (err *ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized access: %+v", err.error.Error())
}

// ClientConfig contains Client configuration.
type ClientConfig struct {
	// The client application name and version.
	AppVersion string

	// The client application user agent in format `client name/client version (os)`, e.g.:
	// (Intel Mac OS X 10_15_3)
	// Mac OS X Mail/13.0 (3608.60.0.2.5) (Intel Mac OS X 10_15_3)
	// Thunderbird/1.5.0 (Ubuntu 18.04.4 LTS)
	// MSOffice 12 (Windows 10 (10.0))
	UserAgent string

	// The client ID.
	ClientID string

	// Timeout is the timeout of the full request. It is passed to http.Client.
	// If it is left unset, it means no timeout is applied.
	Timeout time.Duration

	// FirstReadTimeout specifies the timeout from getting response to the first read of body response.
	// This timeout is applied only when MinBytesPerSecond is used.
	// Default is 5 minutes.
	FirstReadTimeout time.Duration

	// MinBytesPerSecond specifies minimum Bytes per second or the request will be canceled.
	// Zero means no limitation.
	MinBytesPerSecond int64
}

// client is a client of the protonmail API. It implements the Client interface.
type client struct {
	cm *ClientManager
	hc *http.Client

	uid           string
	accessToken   string
	userID        string
	requestLocker sync.Locker
	refreshLocker sync.Locker

	user        *User
	addresses   AddressList
	userKeyRing *crypto.KeyRing
	addrKeyRing map[string]*crypto.KeyRing
	keyRingLock sync.Locker

	log *logrus.Entry
}

// newClient creates a new API client.
func newClient(cm *ClientManager, userID string) *client {
	return &client{
		cm:            cm,
		hc:            getHTTPClient(cm.config, cm.roundTripper, cm.cookieJar),
		userID:        userID,
		requestLocker: &sync.Mutex{},
		refreshLocker: &sync.Mutex{},
		keyRingLock:   &sync.Mutex{},
		addrKeyRing:   make(map[string]*crypto.KeyRing),
		log:           logrus.WithField("pkg", "pmapi").WithField("userID", userID),
	}
}

// getHTTPClient returns a http client configured by the given client config and using the given transport.
func getHTTPClient(cfg *ClientConfig, rt http.RoundTripper, jar http.CookieJar) (hc *http.Client) {
	return &http.Client{
		Transport: rt,
		Jar:       jar,
		Timeout:   cfg.Timeout,
	}
}

func (c *client) IsUnlocked() bool {
	return c.userKeyRing != nil
}

// Unlock unlocks all the user and address keys using the given passphrase, creating user and address keyrings.
// If the keyrings are already present, they are not recreated.
func (c *client) Unlock(passphrase []byte) (err error) {
	c.keyRingLock.Lock()
	defer c.keyRingLock.Unlock()

	return c.unlock(passphrase)
}

// unlock unlocks the user's keys but without locking the keyring lock first.
// Should only be used internally by methods that first lock the lock.
func (c *client) unlock(passphrase []byte) (err error) {
	if _, err = c.CurrentUser(); err != nil {
		return
	}

	if c.userKeyRing == nil {
		if err = c.unlockUser(passphrase); err != nil {
			return errors.Wrap(err, "failed to unlock user")
		}
	}

	for _, address := range c.addresses {
		if c.addrKeyRing[address.ID] == nil {
			if err = c.unlockAddress(passphrase, address); err != nil {
				return errors.Wrap(err, "failed to unlock address")
			}
		}
	}

	return
}

func (c *client) ReloadKeys(passphrase []byte) (err error) {
	c.keyRingLock.Lock()
	defer c.keyRingLock.Unlock()

	c.clearKeys()

	return c.unlock(passphrase)
}

func (c *client) clearKeys() {
	if c.userKeyRing != nil {
		c.userKeyRing.ClearPrivateParams()
		c.userKeyRing = nil
	}

	for id, kr := range c.addrKeyRing {
		if kr != nil {
			kr.ClearPrivateParams()
		}
		delete(c.addrKeyRing, id)
	}
}

func (c *client) CloseConnections() {
	c.hc.CloseIdleConnections()
}

// Do makes an API request. It does not check for HTTP status code errors.
func (c *client) Do(req *http.Request, retryUnauthorized bool) (res *http.Response, err error) {
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
func (c *client) doBuffered(req *http.Request, bodyBuffer []byte, retryUnauthorized bool) (res *http.Response, err error) { // nolint[funlen]
	isAuthReq := strings.Contains(req.URL.Path, "/auth")

	req.Header.Set("User-Agent", c.cm.config.UserAgent)
	req.Header.Set("x-pm-appversion", c.cm.config.AppVersion)

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
		c.log.Tracef("REQBODY '%s'", printBytes(bodyBuffer))
	}

	hasBody := len(bodyBuffer) > 0
	if res, err = c.hc.Do(req); err != nil {
		if res == nil {
			c.log.WithError(err).Error("Cannot get response")
			err = ErrAPINotReachable
		}
		return
	}

	// Cookies are returned only after request was sent.
	c.log.Tracef("REQCOOKIES '%v'", req.Cookies())

	resDate := res.Header.Get("Date")
	if resDate != "" {
		if serverTime, err := http.ParseTime(resDate); err == nil {
			crypto.UpdateTime(serverTime.Unix())
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
func (c *client) DoJSON(req *http.Request, data interface{}) error {
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
func (c *client) doJSONBuffered(req *http.Request, reqBodyBuffer []byte, data interface{}) error { // nolint[funlen]
	req.Header.Set("Accept", "application/vnd.protonmail.v1+json")

	var cancelRequest context.CancelFunc
	if c.cm.config.MinBytesPerSecond > 0 {
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
	if c.cm.config.MinBytesPerSecond == 0 {
		resBody, err = ioutil.ReadAll(res.Body)
	} else {
		resBody, err = c.readAllMinSpeed(res.Body, cancelRequest)
		if err == context.Canceled {
			err = ErrConnectionSlow
		}
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

func (c *client) readAllMinSpeed(data io.Reader, cancelRequest context.CancelFunc) ([]byte, error) {
	firstReadTimeout := c.cm.config.FirstReadTimeout
	if firstReadTimeout == 0 {
		firstReadTimeout = 5 * time.Minute
	}
	timer := time.AfterFunc(firstReadTimeout, func() {
		cancelRequest()
	})

	// speedCheckSeconds controls how often we check the transfer speed.
	const speedCheckSeconds = 3

	var buffer bytes.Buffer
	for {
		_, err := io.CopyN(&buffer, data, c.cm.config.MinBytesPerSecond*speedCheckSeconds)
		timer.Stop()
		timer.Reset(speedCheckSeconds * time.Second)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return ioutil.ReadAll(&buffer)
}

func (c *client) refreshAccessToken() (err error) {
	c.log.Debug("Refreshing token")

	refreshToken := c.cm.GetToken(c.userID)

	if refreshToken == "" {
		c.sendAuth(nil)
		return ErrInvalidToken
	}

	if _, err := c.AuthRefresh(refreshToken); err != nil {
		if err != ErrAPINotReachable {
			c.sendAuth(nil)
		}
		return errors.Wrap(err, "failed to refresh auth")
	}

	return
}

func (c *client) handleStatusUnauthorized(req *http.Request, reqBodyBuffer []byte, res *http.Response, retry bool) (retryRes *http.Response, err error) {
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
