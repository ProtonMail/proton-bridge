// Copyright (c) 2023 Proton AG
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
	"net/smtp"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	frontend "github.com/ProtonMail/proton-bridge/v3/internal/frontend/grpc"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/bradenaw/juniper/xslices"
	"github.com/cucumber/godog"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
)

var defaultVersion = semver.MustParse("3.0.6")

type testCtx struct {
	// These are the objects supporting the test.
	dir      string
	api      API
	netCtl   *proton.NetCtl
	locator  *locations.Locations
	storeKey []byte
	version  *semver.Version
	mocks    *bridge.Mocks
	events   *eventCollector
	reporter *reportRecorder

	// bridge holds the bridge app under test.
	bridge *bridge.Bridge

	// service holds the gRPC frontend service under test.
	service   *frontend.Service
	serviceWG sync.WaitGroup

	// client holds the gRPC frontend client under test.
	client        frontend.BridgeClient
	clientConn    *grpc.ClientConn
	clientEventCh *queue.QueuedChannel[*frontend.StreamEvent]

	// These maps hold expected userIDByName, their primary addresses and bridge passwords.
	userUUIDByName     map[string]string
	addrUUIDByName     map[string]string
	userIDByName       map[string]string
	userAddrByEmail    map[string]map[string]string
	userPassByID       map[string]string
	userBridgePassByID map[string][]byte

	// These are the IMAP and SMTP clients used to connect to bridge.
	imapClients map[string]*imapClient
	smtpClients map[string]*smtpClient

	// calls holds calls made to the API during each step of the test.
	calls     [][]server.Call
	callsLock sync.RWMutex

	// errors holds test-related errors encountered while running test steps.
	errors     [][]error
	errorsLock sync.RWMutex
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
		dir:      dir,
		api:      newTestAPI(),
		netCtl:   proton.NewNetCtl(),
		locator:  locations.New(bridge.NewTestLocationsProvider(dir), "config-name"),
		storeKey: []byte("super-secret-store-key"),
		version:  defaultVersion,
		mocks:    bridge.NewMocks(tb, defaultVersion, defaultVersion),
		events:   newEventCollector(),
		reporter: newReportRecorder(tb),

		userUUIDByName:     make(map[string]string),
		addrUUIDByName:     make(map[string]string),
		userIDByName:       make(map[string]string),
		userAddrByEmail:    make(map[string]map[string]string),
		userPassByID:       make(map[string]string),
		userBridgePassByID: make(map[string][]byte),

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
			t.userUUIDByName[name] = uuid.NewString()
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

func (t *testCtx) getName(wantUserID string) string {
	for name, userID := range t.userIDByName {
		if userID == wantUserID {
			return name
		}
	}

	panic(fmt.Sprintf("unknown user ID %q", wantUserID))
}

func (t *testCtx) getUserID(username string) string {
	return t.userIDByName[username]
}

func (t *testCtx) setUserID(username, userID string) {
	t.userIDByName[username] = userID
}

func (t *testCtx) getUserAddrID(userID, email string) string {
	return t.userAddrByEmail[userID][email]
}

func (t *testCtx) getUserAddrs(userID string) []string {
	return maps.Keys(t.userAddrByEmail[userID])
}

func (t *testCtx) setUserAddr(userID, addrID, email string) {
	if _, ok := t.userAddrByEmail[userID]; !ok {
		t.userAddrByEmail[userID] = make(map[string]string)
	}

	t.userAddrByEmail[userID][email] = addrID
}

func (t *testCtx) unsetUserAddr(userID, wantAddrID string) {
	for email, addrID := range t.userAddrByEmail[userID] {
		if addrID == wantAddrID {
			delete(t.userAddrByEmail[userID], email)
		}
	}
}

func (t *testCtx) getUserPass(userID string) string {
	return t.userPassByID[userID]
}

func (t *testCtx) setUserPass(userID, pass string) {
	t.userPassByID[userID] = pass
}

func (t *testCtx) getUserBridgePass(userID string) string {
	return string(t.userBridgePassByID[userID])
}

func (t *testCtx) setUserBridgePass(userID string, pass []byte) {
	t.userBridgePassByID[userID] = pass
}

func (t *testCtx) getMBoxID(userID string, name string) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var labelID string

	if err := t.withClient(ctx, t.getName(userID), func(ctx context.Context, client *proton.Client) error {
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
	t.callsLock.RLock()
	defer t.callsLock.RUnlock()

	root, err := url.Parse(t.api.GetHostURL())
	if err != nil {
		return server.Call{}, err
	}

	if matches := xslices.Filter(xslices.Join(t.calls...), func(call server.Call) bool {
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

	t.api.Close()
	t.events.close()
	t.reporter.close()
	t.reporter.assertEmpty()
}
