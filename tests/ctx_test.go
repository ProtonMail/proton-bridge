// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	frontend "github.com/ProtonMail/proton-bridge/v3/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var defaultVersion = semver.MustParse("3.10.0")

type testUser struct {
	name       string      // the test user name
	userID     string      // the user's account ID
	addresses  []*testAddr // the user's addresses
	userPass   string      // the user's account password
	bridgePass string      // the user's bridge password
}

func newTestUser(userID, name, userPass string) *testUser {
	return &testUser{
		userID:   userID,
		name:     name,
		userPass: userPass,
	}
}

func (user *testUser) getName() string {
	return user.name
}

func (user *testUser) getUserID() string {
	return user.userID
}

func (user *testUser) getEmails() []string {
	return xslices.Map(user.addresses, func(addr *testAddr) string {
		return addr.email
	})
}

func (user *testUser) getAddrID(email string) string {
	for _, addr := range user.addresses {
		if addr.email == email {
			return addr.addrID
		}
	}

	panic(fmt.Sprintf("unknown email %q", email))
}

func (user *testUser) addAddress(addrID, email string) {
	user.addresses = append(user.addresses, newTestAddr(addrID, email))
}

func (user *testUser) getUserPass() string {
	return user.userPass
}

func (user *testUser) getBridgePass() string {
	return user.bridgePass
}

func (user *testUser) setBridgePass(pass string) {
	user.bridgePass = pass
}

type testAddr struct {
	addrID string // the remote address ID
	email  string // the test address email
}

func newTestAddr(addrID, email string) *testAddr {
	return &testAddr{
		addrID: addrID,
		email:  email,
	}
}

type testCtx struct {
	// These are the objects supporting the test.
	dir       string
	api       API
	netCtl    *proton.NetCtl
	locator   *locations.Locations
	storeKey  []byte
	version   *semver.Version
	mocks     *bridge.Mocks
	events    *eventCollector
	reporter  *reportRecorder
	heartbeat *heartbeatRecorder

	// bridge holds the bridge app under test.
	bridge *bridge.Bridge
	vault  *vault.Vault

	// service holds the gRPC frontend service under test.
	service   *frontend.Service
	serviceWG sync.WaitGroup

	// client holds the gRPC frontend client under test.
	client        frontend.BridgeClient
	clientConn    *grpc.ClientConn
	clientEventCh *async.QueuedChannel[*frontend.StreamEvent]

	// These maps hold test objects created during the test.
	userByID       map[string]*testUser
	userUUIDByName map[string]string
	addrByID       map[string]*testAddr
	addrUUIDByName map[string]string

	// These are the IMAP and SMTP clients used to connect to bridge.
	imapClients map[string]*imapClient
	smtpClients map[string]*smtpClient

	// calls holds calls made to the API during each step of the test.
	calls     [][]server.Call
	callsLock sync.RWMutex

	// errors holds test-related errors encountered while running test steps.
	errors     [][]error
	errorsLock sync.RWMutex

	// This slice contains the dummy listeners that are intended to block network ports.
	dummyListeners []net.Listener

	imapServerStarted bool
	smtpServerStarted bool

	rt *http.RoundTripper
}

type imapClient struct {
	userID string
	client *client.Client
}

type smtpClient struct {
	userID string
	client *smtp.Client
}

func newTestCtx(tb testing.TB) *testCtx {
	dir := tb.TempDir()

	t := &testCtx{
		dir:       dir,
		api:       newTestAPI(),
		netCtl:    proton.NewNetCtl(),
		locator:   locations.New(bridge.NewTestLocationsProvider(dir), "config-name"),
		storeKey:  []byte("super-secret-store-key"),
		version:   defaultVersion,
		mocks:     bridge.NewMocks(tb, defaultVersion, defaultVersion),
		events:    newEventCollector(),
		reporter:  newReportRecorder(tb),
		heartbeat: newHeartbeatRecorder(tb),

		userByID:       make(map[string]*testUser),
		userUUIDByName: make(map[string]string),
		addrByID:       make(map[string]*testAddr),
		addrUUIDByName: make(map[string]string),

		imapClients: make(map[string]*imapClient),
		smtpClients: make(map[string]*smtpClient),
	}

	t.api.AddCallWatcher(func(call server.Call) {
		t.callsLock.Lock()
		defer t.callsLock.Unlock()

		t.calls[len(t.calls)-1] = append(t.calls[len(t.calls)-1], call)
	})

	return t
}

func (t *testCtx) replace(value string) string {
	// Replace [GOOS] with the current OS.
	value = strings.ReplaceAll(value, "[GOOS]", runtime.GOOS)

	// Replace [domain] with the domain of the test server.
	value = strings.ReplaceAll(value, "[domain]", t.api.GetDomain())

	// Replace [user:NAME] with a unique ID for the user NAME.
	value = regexp.MustCompile(`\[user:(\w+)\]`).ReplaceAllStringFunc(value, func(match string) string {
		name := regexp.MustCompile(`\[user:(\w+)\]`).FindStringSubmatch(match)[1]

		// Create a new user if it doesn't exist yet.
		if _, ok := t.userUUIDByName[name]; !ok {
			val := uuid.NewString()

			if name != strings.ToLower(name) {
				val = "Mixed-Caps-" + val
			}

			t.userUUIDByName[name] = val
		}

		return t.userUUIDByName[name]
	})

	// Replace [alias:EMAIL] with a unique address for the email EMAIL.
	value = regexp.MustCompile(`\[alias:(\w+)\]`).ReplaceAllStringFunc(value, func(match string) string {
		email := regexp.MustCompile(`\[alias:(\w+)\]`).FindStringSubmatch(match)[1]

		// Create a new address if it doesn't exist yet.
		if _, ok := t.addrUUIDByName[email]; !ok {
			t.addrUUIDByName[email] = uuid.NewString()
		}

		return t.addrUUIDByName[email]
	})

	// Replace {toUpper:VALUE} with VALUE in uppercase.
	value = regexp.MustCompile(`\{toUpper:([^}]+)\}`).ReplaceAllStringFunc(value, func(match string) string {
		return strings.ToUpper(regexp.MustCompile(`\{toUpper:([^}].+)\}`).FindStringSubmatch(match)[1])
	})

	return value
}

func (t *testCtx) beforeStep(st *godog.Step) {
	logrus.Debugf("Running step: %s", st.Text)

	t.callsLock.Lock()
	defer t.callsLock.Unlock()

	t.errorsLock.Lock()
	defer t.errorsLock.Unlock()

	t.calls = append(t.calls, nil)
	t.errors = append(t.errors, nil)
}

func (t *testCtx) afterStep(st *godog.Step, status godog.StepResultStatus) {
	logrus.Debugf("Finished step (%v): %s", status, st.Text)
}

func (t *testCtx) addUser(userID, name, userPass string) {
	t.userByID[userID] = newTestUser(userID, name, userPass)
}

func (t *testCtx) getUserByName(name string) *testUser {
	for _, user := range t.userByID {
		if user.name == name {
			return user
		}
	}

	panic(fmt.Sprintf("user %q not found", name))
}

func (t *testCtx) getUserByID(userID string) *testUser {
	return t.userByID[userID]
}

func (t *testCtx) getUserByAddress(email string) *testUser {
	for _, user := range t.userByID {
		for _, addr := range user.addresses {
			if addr.email == email {
				return user
			}
		}
	}

	panic(fmt.Sprintf("unknown email %q", email))
}

func (t *testCtx) getMBoxID(userID string, name string) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var labelID string

	if err := t.withClient(ctx, t.getUserByID(userID).getName(), func(ctx context.Context, client *proton.Client) error {
		labels, err := client.GetLabels(ctx, proton.LabelTypeLabel, proton.LabelTypeFolder, proton.LabelTypeSystem)
		if err != nil {
			panic(err)
		}

		idx := xslices.IndexFunc(labels, func(label proton.Label) bool {
			var labelName string
			switch label.Type {
			case proton.LabelTypeSystem:
				labelName = label.Name
			case proton.LabelTypeFolder:
				labelName = fmt.Sprintf("Folders/%v", label.Name)
			case proton.LabelTypeLabel:
				labelName = fmt.Sprintf("Labels/%v", label.Name)
			case proton.LabelTypeContactGroup:
				labelName = label.Name
			}

			return labelName == name
		})

		if idx < 0 {
			panic(fmt.Errorf("label %q not found", name))
		}

		labelID = labels[idx].ID

		return nil
	}); err != nil {
		panic(err)
	}

	return labelID
}

// getDraftID will return the API ID of draft message with draftIndex, where
// draftIndex is similar to sequential ID i.e. 1 represents the first message
// of draft folder sorted by API creation time.
func (t *testCtx) getDraftID(username string, draftIndex int) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var draftID string

	if err := t.withClient(ctx, username, func(ctx context.Context, client *proton.Client) error {
		messages, err := client.GetMessageMetadata(ctx, proton.MessageFilter{LabelID: proton.DraftsLabel})
		if err != nil {
			return fmt.Errorf("failed to get message metadata: %w", err)
		} else if len(messages) < draftIndex {
			return fmt.Errorf("draft index %d is out of range", draftIndex)
		}

		draftID = messages[draftIndex-1].ID

		return nil
	}); err != nil {
		return "", err
	}

	return draftID, nil
}

func (t *testCtx) getLastCall(method, pathExp string) (server.Call, error) {
	matches, err := t.getAllCalls(method, pathExp)
	if err != nil {
		return server.Call{}, err
	}
	if len(matches) > 0 {
		return matches[len(matches)-1], nil
	}
	return server.Call{}, fmt.Errorf("no call with method %q and path %q was made", method, pathExp)
}

func (t *testCtx) getAllCalls(method, pathExp string) ([]server.Call, error) {
	t.callsLock.RLock()
	defer t.callsLock.RUnlock()

	root, err := url.Parse(t.api.GetHostURL())
	if err != nil {
		return []server.Call{}, err
	}

	if matches := xslices.Filter(xslices.Join(t.calls...), func(call server.Call) bool {
		return call.Method == method && regexp.MustCompile("^"+pathExp+"$").MatchString(strings.TrimPrefix(call.URL.Path, root.Path))
	}); len(matches) > 0 {
		return matches, nil
	}

	return []server.Call{}, fmt.Errorf("no call with method %q and path %q was made", method, pathExp)
}

func (t *testCtx) getLastCallExcludingHTTPOverride(method, pathExp string) (server.Call, error) {
	t.callsLock.RLock()
	defer t.callsLock.RUnlock()

	root, err := url.Parse(t.api.GetHostURL())
	if err != nil {
		return server.Call{}, err
	}

	if matches := xslices.Filter(xslices.Join(t.calls...), func(call server.Call) bool {
		if len(call.RequestHeader.Get("X-HTTP-Method-Override")) != 0 || len(call.RequestHeader.Get("X-Http-Method")) != 0 {
			return false
		}

		return call.Method == method && regexp.MustCompile("^"+pathExp+"$").MatchString(strings.TrimPrefix(call.URL.Path, root.Path))
	}); len(matches) > 0 {
		return matches[len(matches)-1], nil
	}

	return server.Call{}, fmt.Errorf("no call with method %q and path %q was made", method, pathExp)
}

func (t *testCtx) pushError(err error) {
	t.errorsLock.Lock()
	defer t.errorsLock.Unlock()

	t.errors[len(t.errors)-1] = append(t.errors[len(t.errors)-1], err)
}

func (t *testCtx) getLastError() error {
	t.errorsLock.RLock()
	defer t.errorsLock.RUnlock()

	if lastStep := t.errors[len(t.errors)-2]; len(lastStep) > 0 {
		return lastStep[len(lastStep)-1]
	}

	return nil
}

func (t *testCtx) close(ctx context.Context) {
	for _, client := range t.imapClients {
		if err := client.client.Logout(); err != nil {
			logrus.WithError(err).Error("Failed to logout IMAP client")
		}
	}

	for _, client := range t.smtpClients {
		if err := client.client.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close SMTP client")
		}
	}

	if t.service != nil {
		if err := t.closeFrontendService(ctx); err != nil {
			logrus.WithError(err).Error("Failed to close frontend service")
		}
	}

	if t.client != nil {
		if err := t.closeFrontendClient(); err != nil {
			logrus.WithError(err).Error("Failed to close frontend client")
		}
	}

	if t.bridge != nil {
		if err := t.closeBridge(ctx); err != nil {
			logrus.WithError(err).Error("Failed to close bridge")
		}
	}

	for _, listener := range t.dummyListeners {
		if err := listener.Close(); err != nil {
			logrus.WithError(err).Errorf("Failed to close dummy listener %v", listener.Addr())
		}
	}

	t.api.Close()
	t.events.close()
	t.reporter.close()
	t.reporter.assertEmpty()
}
