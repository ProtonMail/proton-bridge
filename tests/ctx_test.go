// Copyright (c) 2022 Proton AG
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
	"reflect"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
	"golang.org/x/exp/maps"
)

var defaultVersion = semver.MustParse("1.0.0")

type testCtx struct {
	// These are the objects supporting the test.
	dir      string
	api      API
	netCtl   *liteapi.NetCtl
	locator  *locations.Locations
	storeKey []byte
	version  *semver.Version
	mocks    *bridge.Mocks
	events   *eventCollector

	// bridge holds the bridge app under test.
	bridge *bridge.Bridge

	// These maps hold expected userIDByName, their primary addresses and bridge passwords.
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
		api:      newFakeAPI(),
		netCtl:   liteapi.NewNetCtl(),
		locator:  locations.New(bridge.NewTestLocationsProvider(dir), "config-name"),
		storeKey: []byte("super-secret-store-key"),
		version:  defaultVersion,
		mocks:    bridge.NewMocks(tb, defaultVersion, defaultVersion),
		events:   newEventCollector(),

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

func (t *testCtx) beforeStep() {
	t.callsLock.Lock()
	defer t.callsLock.Unlock()

	t.errorsLock.Lock()
	defer t.errorsLock.Unlock()

	t.calls = append(t.calls, nil)
	t.errors = append(t.errors, nil)
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

	if err := t.withClient(ctx, t.getName(userID), func(ctx context.Context, client *liteapi.Client) error {
		labels, err := client.GetLabels(ctx, liteapi.LabelTypeLabel, liteapi.LabelTypeFolder, liteapi.LabelTypeSystem)
		if err != nil {
			panic(err)
		}

		idx := xslices.IndexFunc(labels, func(label liteapi.Label) bool {
			return label.Name == name
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

func (t *testCtx) getLastCall(method, pathExp string) (server.Call, error) {
	t.callsLock.RLock()
	defer t.callsLock.RUnlock()

	if matches := xslices.Filter(xslices.Join(t.calls...), func(call server.Call) bool {
		return call.Method == method && regexp.MustCompile("^"+pathExp+"$").MatchString(call.URL.Path)
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

func (t *testCtx) close(ctx context.Context) error {
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

	if t.bridge != nil {
		if err := t.bridge.Close(ctx); err != nil {
			return err
		}
	}

	t.api.Close()

	t.events.close()

	return nil
}

type eventCollector struct {
	events map[reflect.Type]*queue.QueuedChannel[events.Event]
	lock   sync.RWMutex
	wg     sync.WaitGroup
}

func newEventCollector() *eventCollector {
	return &eventCollector{
		events: make(map[reflect.Type]*queue.QueuedChannel[events.Event]),
	}
}

func (c *eventCollector) collectFrom(eventCh <-chan events.Event) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for event := range eventCh {
			c.push(event)
		}
	}()
}

func awaitType[T events.Event](c *eventCollector, ofType T, timeout time.Duration) (T, bool) {
	if event := c.await(ofType, timeout); event == nil {
		return *new(T), false //nolint:gocritic
	} else if event, ok := event.(T); !ok {
		panic(fmt.Errorf("unexpected event type %T", event))
	} else {
		return event, true
	}
}

func (c *eventCollector) await(ofType events.Event, timeout time.Duration) events.Event {
	select {
	case event := <-c.getEventCh(ofType):
		return event

	case <-time.After(timeout):
		return nil
	}
}

func (c *eventCollector) push(event events.Event) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.events[reflect.TypeOf(event)]; !ok {
		c.events[reflect.TypeOf(event)] = queue.NewQueuedChannel[events.Event](0, 0)
	}

	c.events[reflect.TypeOf(event)].Enqueue(event)
}

func (c *eventCollector) getEventCh(ofType events.Event) <-chan events.Event {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.events[reflect.TypeOf(ofType)]; !ok {
		c.events[reflect.TypeOf(ofType)] = queue.NewQueuedChannel[events.Event](0, 0)
	}

	return c.events[reflect.TypeOf(ofType)].GetChannel()
}

func (c *eventCollector) close() {
	c.wg.Wait()

	for _, eventCh := range c.events {
		eventCh.CloseAndDiscardQueued()
	}
}
