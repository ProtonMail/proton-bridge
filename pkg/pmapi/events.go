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

package pmapi

import (
	"context"
	"encoding/json"
	"net/mail"

	"github.com/go-resty/resty/v2"
)

// Event represents changes since the last check.
type Event struct {
	// The current event ID.
	EventID string
	// If set to one, all cached data must be fetched again.
	Refresh int
	// If set to one, fetch more events.
	More Boolean
	// Changes applied to messages.
	Messages []*EventMessage
	// Counts of messages per labels.
	MessageCounts []*MessagesCount
	// Changes applied to labels.
	Labels []*EventLabel
	// Current user status.
	User *User
	// Changes to addresses.
	Addresses []*EventAddress
	// Messages to show to the user.
	Notices []string

	// Update of used user space
	UsedSpace *int64
}

// EventAction is the action that created a change.
type EventAction int

const (
	EventDelete      EventAction = iota // EventDelete Item has been deleted.
	EventCreate                         // EventCreate Item has been created.
	EventUpdate                         // EventUpdate Item has been updated.
	EventUpdateFlags                    // EventUpdateFlags For messages: flags have been updated.
)

// Flags for event refresh.
const (
	EventRefreshMail    = 1
	EventRefreshContact = 2
	EventRefreshAll     = 255
)

// maxNumberOfMergedEvents limits how many events are merged into one. It means
// when GetEvent is called and event returns there is more events, it will
// automatically fetch next one and merge it up to this number of events.
const maxNumberOfMergedEvents = 50

// EventItem is an item that has changed.
type EventItem struct {
	ID     string
	Action EventAction
}

// EventMessage is a message that has changed.
type EventMessage struct {
	EventItem

	// If the message has been created, the new message.
	Created *Message `json:"-"`
	// If the message has been updated, the updated fields.
	Updated *EventMessageUpdated `json:"-"`
}

// eventMessage defines a new type to prevent MarshalJSON/UnmarshalJSON infinite loops.
type eventMessage EventMessage

type rawEventMessage struct {
	eventMessage

	// This will be parsed depending on the action.
	Message json.RawMessage `json:",omitempty"`
}

func (em *EventMessage) UnmarshalJSON(b []byte) (err error) {
	var raw rawEventMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*em = EventMessage(raw.eventMessage)

	switch em.Action {
	case EventCreate:
		em.Created = &Message{ID: raw.ID}
		return json.Unmarshal(raw.Message, em.Created)
	case EventUpdate, EventUpdateFlags:
		em.Updated = &EventMessageUpdated{ID: raw.ID}
		return json.Unmarshal(raw.Message, em.Updated)
	case EventDelete:
		return nil
	}
	return nil
}

func (em *EventMessage) MarshalJSON() ([]byte, error) {
	var raw rawEventMessage
	raw.eventMessage = eventMessage(*em)

	var err error
	switch em.Action {
	case EventCreate:
		raw.Message, err = json.Marshal(em.Created)
	case EventUpdate, EventUpdateFlags:
		raw.Message, err = json.Marshal(em.Updated)
	case EventDelete:
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(raw)
}

// EventMessageUpdated contains changed fields for an updated message.
type EventMessageUpdated struct {
	ID string

	Subject *string
	Unread  *Boolean
	Flags   *int64
	Sender  *mail.Address
	ToList  *[]*mail.Address
	CCList  *[]*mail.Address
	BCCList *[]*mail.Address
	Time    int64

	// Fields only present for EventUpdateFlags.
	LabelIDs        []string
	LabelIDsAdded   []string
	LabelIDsRemoved []string
}

// EventLabel is a label that has changed.
type EventLabel struct {
	EventItem
	Label *Label
}

// EventAddress is an address that has changed.
type EventAddress struct {
	EventItem
	Address *Address
}

// GetEvent returns a summary of events that occurred since last. To get the latest event,
// provide an empty last value. The latest event is always empty.
func (c *client) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	return c.getEvent(ctx, eventID, 1)
}

func (c *client) getEvent(ctx context.Context, eventID string, numberOfMergedEvents int) (*Event, error) {
	if eventID == "" {
		eventID = "latest"
	}

	var event *Event

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&event).Get("/events/" + eventID)
	}); err != nil {
		return nil, err
	}

	// API notifies about used space two ways:
	//    - by `event.User.UsedSpace`
	//    - by `event.UsedSpace`
	//
	// Because event merging is implemented for User object we copy the
	// value from event.UsedSpace to event.User.UsedSpace and continue with
	// user.
	if event.UsedSpace != nil {
		if event.User == nil {
			event.User = &User{UsedSpace: event.UsedSpace}
		} else {
			event.User.UsedSpace = event.UsedSpace
		}
	}

	if event.More && numberOfMergedEvents < maxNumberOfMergedEvents {
		nextEvent, err := c.getEvent(ctx, event.EventID, numberOfMergedEvents+1)
		if err != nil {
			return nil, err
		}
		event = mergeEvents(event, nextEvent)
	}

	return event, nil
}

// mergeEvents combines an old events and a new events object.
// This is not as simple as just blindly joining the two because some things should only be taken from the new events.
func mergeEvents(eventsOld *Event, eventsNew *Event) (mergedEvents *Event) {
	return &Event{
		EventID:       eventsNew.EventID,
		Refresh:       eventsOld.Refresh | eventsNew.Refresh,
		More:          eventsNew.More,
		Messages:      append(eventsOld.Messages, eventsNew.Messages...),
		MessageCounts: append(eventsOld.MessageCounts, eventsNew.MessageCounts...),
		Labels:        append(eventsOld.Labels, eventsNew.Labels...),
		User:          mergeUserEvents(eventsOld.User, eventsNew.User),
		Addresses:     append(eventsOld.Addresses, eventsNew.Addresses...),
		Notices:       append(eventsOld.Notices, eventsNew.Notices...),
	}
}

func mergeUserEvents(userOld, userNew *User) *User {
	if userNew == nil {
		return userOld
	}

	if userOld != nil {
		if userNew.MaxSpace == nil {
			userNew.MaxSpace = userOld.MaxSpace
		}
		if userNew.UsedSpace == nil {
			userNew.UsedSpace = userOld.UsedSpace
		}
	}

	return userNew
}
