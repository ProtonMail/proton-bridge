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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"context"
	"math/rand"
	"time"

	bridgeEvents "github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pollInterval       = 30 * time.Second
	pollIntervalSpread = 5 * time.Second

	// errMaxSentry defines after how many errors in a row to report it to sentry.
	errMaxSentry = 20
)

type eventLoop struct {
	currentEvents  *Events
	currentEventID string
	currentEvent   *pmapi.Event
	pollCh         chan chan struct{}
	stopCh         chan struct{}
	notifyStopCh   chan struct{}
	isRunning      bool // The whole event loop is running.

	pollCounter int
	errCounter  int

	log *logrus.Entry

	store    *Store
	user     BridgeUser
	listener listener.Listener
}

func newEventLoop(currentEvents *Events, store *Store, user BridgeUser, listener listener.Listener) *eventLoop {
	eventLog := log.WithField("userID", user.ID())
	eventLog.Trace("Creating new event loop")

	return &eventLoop{
		currentEvents:  currentEvents,
		currentEventID: currentEvents.getEventID(user.ID()),
		pollCh:         make(chan chan struct{}),
		isRunning:      false,

		log: eventLog,

		store:    store,
		user:     user,
		listener: listener,
	}
}

func (loop *eventLoop) client() pmapi.Client {
	return loop.store.client()
}

func (loop *eventLoop) setFirstEventID() (err error) {
	loop.log.Info("Setting first event ID")

	event, err := loop.client().GetEvent(pmapi.ContextWithoutRetry(context.Background()), "")
	if err != nil {
		loop.log.WithError(err).Error("Could not get latest event ID")
		return
	}

	loop.currentEventID = event.EventID

	if err = loop.currentEvents.setEventID(loop.user.ID(), loop.currentEventID); err != nil {
		loop.log.WithError(err).Error("Could not set latest event ID in user cache")
		return
	}

	return
}

// pollNow starts polling events right away and waits till the events are
// processed so we are sure updates are propagated to the database.
func (loop *eventLoop) pollNow() {
	// When event loop is not running, it would cause infinite wait.
	if !loop.isRunning {
		return
	}

	eventProcessedCh := make(chan struct{})
	loop.pollCh <- eventProcessedCh
	<-eventProcessedCh
	close(eventProcessedCh)
}

func (loop *eventLoop) stop() {
	if loop.isRunning {
		loop.isRunning = false
		close(loop.stopCh)

		select {
		case <-loop.notifyStopCh:
			loop.log.Warn("Event loop was stopped")
		case <-time.After(1 * time.Second):
			loop.log.Warn("Timed out waiting for event loop to stop")
		}
	}
}

func (loop *eventLoop) start() {
	if loop.isRunning {
		return
	}
	defer func() {
		loop.isRunning = false
	}()
	loop.stopCh = make(chan struct{})
	loop.notifyStopCh = make(chan struct{})
	loop.isRunning = true

	events := make(chan *pmapi.Event)
	defer close(events)

	loop.log.WithField("lastEventID", loop.currentEventID).Info("Subscribed to events")
	defer func() {
		loop.log.WithField("lastEventID", loop.currentEventID).Warn("Subscription stopped")
	}()

	go loop.pollNow()

	loop.loop()
}

// loop is the main body of the event loop.
func (loop *eventLoop) loop() {
	t := time.NewTicker(pollInterval - pollIntervalSpread)
	defer t.Stop()

	for {
		var eventProcessedCh chan struct{}
		select {
		case <-loop.stopCh:
			close(loop.notifyStopCh)
			return
		case <-t.C:
			// Randomise periodic calls within range pollInterval ± pollSpread to reduces potential load spikes on API.
			//nolint:gosec    // It is OK to use weaker random number generator here
			time.Sleep(time.Duration(rand.Intn(2*int(pollIntervalSpread.Milliseconds()))) * time.Millisecond)
		case eventProcessedCh = <-loop.pollCh:
			// We don't want to wait here. Polling should happen instantly.
		}

		// Before we fetch the first event, check whether this is the first time we've
		// started the event loop, and if so, trigger a full sync.
		// In case internet connection was not available during start, it will be
		// handled anyway when the connection is back here.
		if loop.isBeforeFirstStart() {
			if eventErr := loop.setFirstEventID(); eventErr != nil {
				loop.log.WithError(eventErr).Warn("Could not set initial event ID")
			}
		}

		// If the sync is not finished then a new sync is triggered.
		if !loop.store.isSyncFinished() {
			loop.store.triggerSync()
		}

		more, err := loop.processNextEvent()
		if eventProcessedCh != nil {
			eventProcessedCh <- struct{}{}
		}
		if err != nil {
			loop.log.WithError(err).Error("Cannot process event, stopping event loop")
			// When event loop stops, the only way to start it again is by login.
			// It should stop only when user is logged out but even if there is other
			// serious error, logout is intended action.
			if errLogout := loop.user.Logout(); errLogout != nil {
				loop.log.
					WithError(errLogout).
					Error("Failed to logout user after loop finished with error")
			}
			return
		}

		if more {
			go loop.pollNow()
		}
	}
}

// isBeforeFirstStart returns whether the initial event ID was already set or not.
func (loop *eventLoop) isBeforeFirstStart() bool {
	return loop.currentEventID == ""
}

// processNextEvent saves only successfully processed `eventID` into cache
// (disk). It will filter out in defer all errors except invalid token error.
// Invalid error will be returned and stop the event loop.
func (loop *eventLoop) processNextEvent() (more bool, err error) { //nolint:funlen
	l := loop.log.
		WithField("currentEventID", loop.currentEventID).
		WithField("pollCounter", loop.pollCounter)

	// We only want to consider invalid tokens as real errors because all other errors might fix themselves eventually
	// (e.g. no internet, ulimit reached etc.)
	defer func() {
		if errors.Cause(err) == pmapi.ErrNoConnection {
			l.Warn("Internet unavailable")
			err = nil
		}

		if err != nil && isFdCloseToULimit() {
			l.Warn("Ulimit reached")
			loop.listener.Emit(bridgeEvents.RestartBridgeEvent, "")
			err = nil
		}

		if errors.Cause(err) == pmapi.ErrUpgradeApplication {
			l.Warn("Need to upgrade application")
			err = nil
		}

		if err == nil {
			loop.errCounter = 0
		}

		// All errors except ErrUnauthorized (which is not possible to recover from) are ignored.
		if err != nil && !pmapi.IsFailedAuth(errors.Cause(err)) && errors.Cause(err) != pmapi.ErrUnauthorized {
			l.WithError(err).WithField("errors", loop.errCounter).Error("Error skipped")
			loop.errCounter++
			if loop.errCounter == errMaxSentry {
				context := map[string]interface{}{
					"EventLoop": map[string]interface{}{
						"EventID": loop.currentEventID,
					},
				}
				if sentryErr := loop.store.sentryReporter.ReportMessageWithContext("Warning: event loop issues: "+err.Error(), context); sentryErr != nil {
					l.WithError(sentryErr).Error("Failed to report error to sentry")
				}
			}
			err = nil
		}
	}()

	l.Trace("Polling next event")
	// Log activity of event loop each 100. poll which means approx. 28
	// lines per day
	if loop.pollCounter%100 == 0 {
		l.Info("Polling next event")
	}
	loop.pollCounter++

	var event *pmapi.Event
	if event, err = loop.client().GetEvent(pmapi.ContextWithoutRetry(context.Background()), loop.currentEventID); err != nil {
		return false, errors.Wrap(err, "failed to get event")
	}

	loop.currentEvent = event

	if event == nil {
		return false, errors.New("received empty event")
	}

	if err = loop.processEvent(event); err != nil {
		return false, errors.Wrap(err, "failed to process event")
	}

	if loop.currentEventID != event.EventID {
		l.WithField("newID", event.EventID).Info("New event processed")
		// In case new event ID cannot be saved to cache, we update it in event loop
		// anyway and continue processing new events to prevent the loop from repeatedly
		// processing the same event.
		// This allows the event loop to continue to function (unless the cache was broken
		// and bridge stopped, in which case it will start from the old event ID anyway).
		loop.currentEventID = event.EventID
		if err = loop.currentEvents.setEventID(loop.user.ID(), event.EventID); err != nil {
			return false, errors.Wrap(err, "failed to save event ID to cache")
		}
	}

	return bool(event.More), err
}

func (loop *eventLoop) processEvent(event *pmapi.Event) (err error) {
	eventLog := loop.log.WithField("event", event.EventID)
	eventLog.Debug("Processing event")

	if (event.Refresh & pmapi.EventRefreshMail) != 0 {
		eventLog.Info("Processing refresh event")
		loop.store.triggerSync()

		context := map[string]interface{}{
			"EventLoop": map[string]interface{}{
				"EventID": loop.currentEventID,
			},
		}
		if sentryErr := loop.store.sentryReporter.ReportMessageWithContext("Warning: refresh occurred", context); sentryErr != nil {
			loop.log.WithError(sentryErr).Error("Failed to report refresh to sentry")
		}

		return
	}

	if len(event.Addresses) != 0 {
		if err = loop.processAddresses(eventLog, event.Addresses); err != nil {
			return errors.Wrap(err, "failed to process address events")
		}
	}

	if len(event.Labels) != 0 {
		if err = loop.processLabels(eventLog, event.Labels); err != nil {
			return errors.Wrap(err, "failed to process label events")
		}
	}

	if len(event.Messages) != 0 {
		if err = loop.processMessages(eventLog, event.Messages); err != nil {
			return errors.Wrap(err, "failed to process message events")
		}
	}

	if event.User != nil {
		loop.user.UpdateSpace(event.User)
		loop.listener.Emit(bridgeEvents.UserRefreshEvent, loop.user.ID())
	}

	// One would expect that every event would contain MessageCount as part of
	// the event.Messages, but this is apparently not the case.
	// MessageCounts are served on an irregular basis, so we should update and
	// compare the counts only when we receive them.
	if len(event.MessageCounts) != 0 {
		if err = loop.processMessageCounts(eventLog, event.MessageCounts); err != nil {
			return errors.Wrap(err, "failed to process message count events")
		}
	}

	if len(event.Notices) != 0 {
		loop.processNotices(eventLog, event.Notices)
	}

	return err
}

func (loop *eventLoop) processAddresses(log *logrus.Entry, addressEvents []*pmapi.EventAddress) (err error) {
	log.Debug("Processing address change event")

	// Get old addresses for comparisons before updating user.
	oldList := loop.client().Addresses()

	if err = loop.user.UpdateUser(context.Background()); err != nil {
		if logoutErr := loop.user.Logout(); logoutErr != nil {
			log.WithError(logoutErr).Error("Failed to logout user after failed update")
		}
		return errors.Wrap(err, "failed to update user")
	}

	for _, addressEvent := range addressEvents {
		switch addressEvent.Action {
		case pmapi.EventCreate:
			log.WithField("email", addressEvent.Address.Email).Info("Address was created")
			loop.listener.Emit(bridgeEvents.AddressChangedEvent, loop.user.GetPrimaryAddress())

		case pmapi.EventUpdate:
			oldAddress := oldList.ByID(addressEvent.ID)
			if oldAddress == nil {
				log.Warning("Event refers to an address that isn't present")
				continue
			}

			email := oldAddress.Email
			log.WithField("email", email).Info("Address was updated")
			if addressEvent.Address.Receive != oldAddress.Receive {
				loop.listener.Emit(bridgeEvents.AddressChangedLogoutEvent, email)
			}

		case pmapi.EventDelete:
			oldAddress := oldList.ByID(addressEvent.ID)
			if oldAddress == nil {
				log.Warning("Event refers to an address that isn't present")
				continue
			}

			email := oldAddress.Email
			log.WithField("email", email).Info("Address was deleted")
			loop.user.CloseConnection(email)
			loop.listener.Emit(bridgeEvents.AddressChangedLogoutEvent, email)
		case pmapi.EventUpdateFlags:
			log.Error("EventUpdateFlags for address event is uknown operation")
		}
	}

	if err = loop.store.createOrUpdateAddressInfo(loop.client().Addresses()); err != nil {
		return errors.Wrap(err, "failed to update address IDs in store")
	}

	if err = loop.store.createOrDeleteAddressesEvent(); err != nil {
		return errors.Wrap(err, "failed to create/delete store addresses")
	}

	return nil
}

func (loop *eventLoop) processLabels(eventLog *logrus.Entry, labels []*pmapi.EventLabel) error {
	eventLog.Debug("Processing label change event")

	for _, eventLabel := range labels {
		label := eventLabel.Label
		switch eventLabel.Action {
		case pmapi.EventCreate, pmapi.EventUpdate:
			if err := loop.store.createOrUpdateMailboxEvent(label); err != nil {
				return errors.Wrap(err, "failed to create or update label")
			}
		case pmapi.EventDelete:
			if err := loop.store.deleteMailboxEvent(eventLabel.ID); err != nil {
				return errors.Wrap(err, "failed to delete label")
			}
		case pmapi.EventUpdateFlags:
			log.Error("EventUpdateFlags for label event is uknown operation")
		}
	}

	return nil
}

func (loop *eventLoop) processMessages(eventLog *logrus.Entry, messages []*pmapi.EventMessage) (err error) { //nolint:funlen
	eventLog.Debug("Processing message change event")

	for _, message := range messages {
		msgLog := eventLog.WithField("msgID", message.ID)

		switch message.Action {
		case pmapi.EventCreate:
			msgLog.Debug("Processing EventCreate for message")

			if message.Created == nil {
				msgLog.Error("Got EventCreate with nil message")
				continue
			}

			if err = loop.store.createOrUpdateMessageEvent(message.Created); err != nil {
				return errors.Wrap(err, "failed to put message into DB")
			}

		case pmapi.EventUpdate, pmapi.EventUpdateFlags:
			msgLog.Debug("Processing EventUpdate(Flags) for message")

			if message.Updated == nil {
				msgLog.Error("Got EventUpdate(Flags) with nil message")
				continue
			}

			var msg *pmapi.Message

			if msg, err = loop.store.getMessageFromDB(message.ID); err != nil {
				if err != ErrNoSuchAPIID {
					return errors.Wrap(err, "failed to get message from DB for updating")
				}

				msgLog.WithError(err).Warning("Message was not present in DB. Trying fetch...")

				if msg, err = loop.client().GetMessage(context.Background(), message.ID); err != nil {
					if pmapi.IsUnprocessableEntity(err) {
						msgLog.WithError(err).Warn("Skipping message update because message exists neither in local DB nor on API")
						err = nil
						continue
					}

					return errors.Wrap(err, "failed to get message from API for updating")
				}
			}

			updateMessage(msgLog, msg, message.Updated)

			loop.removeLabelFromMessageWait(message.Updated.LabelIDsRemoved)
			if err = loop.store.createOrUpdateMessageEvent(msg); err != nil {
				return errors.Wrap(err, "failed to update message in DB")
			}

		case pmapi.EventDelete:
			msgLog.Debug("Processing EventDelete for message")

			loop.removeMessageWait(message.ID)
			if err = loop.store.deleteMessageEvent(message.ID); err != nil {
				return errors.Wrap(err, "failed to delete message from DB")
			}
		}
	}

	return err
}

// removeMessageWait waits for notifier to be ready to accept delete
// operations for given message. It's no-op if message does not exist.
func (loop *eventLoop) removeMessageWait(msgID string) {
	msg, err := loop.store.getMessageFromDB(msgID)
	if err != nil {
		return
	}
	loop.removeLabelFromMessageWait(msg.LabelIDs)
}

// removeLabelFromMessageWait waits for notifier to be ready to accept
// delete operations for given labels.
func (loop *eventLoop) removeLabelFromMessageWait(labelIDs []string) {
	if len(labelIDs) == 0 || loop.store.notifier == nil {
		return
	}

	for {
		wasWaiting := false
		for _, labelID := range labelIDs {
			canDelete, wait := loop.store.notifier.CanDelete(labelID)
			if !canDelete {
				wasWaiting = true
				wait()
			}
		}
		// If we had to wait for some label, we need to check again
		// all labels in case something changed in the meantime.
		if !wasWaiting {
			return
		}
	}
}

func updateMessage(msgLog *logrus.Entry, message *pmapi.Message, updates *pmapi.EventMessageUpdated) { //nolint:funlen
	msgLog.Debug("Updating message")

	message.Time = updates.Time

	if updates.Subject != nil {
		msgLog.WithField("subject", *updates.Subject).Trace("Updating message value")
		message.Subject = *updates.Subject
	}

	if updates.Sender != nil {
		msgLog.WithField("sender", *updates.Sender).Trace("Updating message value")
		message.Sender = updates.Sender
	}

	if updates.ToList != nil {
		msgLog.WithField("toList", *updates.ToList).Trace("Updating message value")
		message.ToList = *updates.ToList
	}

	if updates.CCList != nil {
		msgLog.WithField("ccList", *updates.CCList).Trace("Updating message value")
		message.CCList = *updates.CCList
	}

	if updates.BCCList != nil {
		msgLog.WithField("bccList", *updates.BCCList).Trace("Updating message value")
		message.BCCList = *updates.BCCList
	}

	if updates.Unread != nil {
		msgLog.WithField("unread", *updates.Unread).Trace("Updating message value")
		message.Unread = *updates.Unread
	}

	if updates.Flags != nil {
		msgLog.WithField("flags", *updates.Flags).Trace("Updating message value")
		message.Flags = *updates.Flags
	}

	if updates.LabelIDs != nil {
		msgLog.WithField("labelIDs", updates.LabelIDs).Trace("Updating message value")
		message.LabelIDs = updates.LabelIDs
	} else {
		for _, added := range updates.LabelIDsAdded {
			if !message.HasLabelID(added) {
				msgLog.WithField("added", added).Trace("Adding label to message")
				message.LabelIDs = append(message.LabelIDs, added)
			}
		}

		labels := []string{}
		for _, l := range message.LabelIDs {
			removeLabel := false
			for _, removed := range updates.LabelIDsRemoved {
				if removed == l {
					removeLabel = true
					break
				}
			}
			if removeLabel {
				msgLog.WithField("label", l).Trace("Removing label from message")
			} else {
				labels = append(labels, l)
			}
		}

		message.LabelIDs = labels
	}
}

func (loop *eventLoop) processMessageCounts(l *logrus.Entry, messageCounts []*pmapi.MessagesCount) error {
	l.WithField("apiCounts", messageCounts).Debug("Processing message count change event")

	isSynced, err := loop.store.isSynced(messageCounts)
	if err != nil {
		return err
	}
	if !isSynced {
		log.Error("The counts between DB and API are not matching")
	}

	return nil
}

func (loop *eventLoop) processNotices(l *logrus.Entry, notices []string) {
	l.Debug("Processing notice change event")

	for _, notice := range notices {
		l.Infof("Notice: %q", notice)
		for _, address := range loop.user.GetStoreAddresses() {
			loop.store.notifyNotice(address, notice)
		}
	}
}
