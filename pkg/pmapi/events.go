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
	"encoding/json"
	"net/http"
	"net/mail"
)

// Event represents changes since the last check.
type Event struct {
	// The current event ID.
	EventID string
	// If set to one, all cached data must be fetched again.
	Refresh int
	// If set to one, fetch more events.
	More int
	// Changes applied to messages.
	Messages []*EventMessage
	// Counts of messages per labels.
	MessageCounts []*MessagesCount
	// Changes applied to labels.
	Labels []*EventLabel
	// Current user status.
	User User
	// Changes to addresses.
	Addresses []*EventAddress
	// Messages to show to the user.
	Notices []string
}

// EventAction is the action that created a change.
type EventAction int

const (
	EventDelete      EventAction = iota // Item has been deleted.
	EventCreate                         // Item has been created.
	EventUpdate                         // Item has been updated.
	EventUpdateFlags                    // For messages: flags have been updated.
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
	Unread  *int
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

type EventRes struct {
	Res
	*Event
}

type LatestEventRes struct {
	Res
	*Event
}

// GetEvent returns a summary of events that occurred since last. To get the latest event,
// provide an empty last value. The latest event is always empty.
func (c *client) GetEvent(last string) (event *Event, err error) {
	return c.getEvent(last, 1)
}

func (c *client) getEvent(last string, numberOfMergedEvents int) (event *Event, err error) {
	var req *http.Request
	if last == "" {
		req, err = c.NewRequest("GET", "/events/latest", nil)
		if err != nil {
			return
		}

		var res LatestEventRes
		if err = c.DoJSON(req, &res); err != nil {
			return
		}

		event, err = res.Event, res.Err()
	} else {
		req, err = c.NewRequest("GET", "/events/"+last, nil)
		if err != nil {
			return
		}

		var res EventRes
		if err = c.DoJSON(req, &res); err != nil {
			return
		}

		event, err = res.Event, res.Err()
		if err != nil {
			return
		}

		if event.More == 1 && numberOfMergedEvents < maxNumberOfMergedEvents {
			var moreEvents *Event
			if moreEvents, err = c.getEvent(event.EventID, numberOfMergedEvents+1); err != nil {
				return
			}
			event = mergeEvents(event, moreEvents)
		}
	}

	return event, err
}

// mergeEvents combines an old events and a new events object.
// This is not as simple as just blindly joining the two because some things should only be taken from the new events.
func mergeEvents(eventsOld *Event, eventsNew *Event) (mergedEvents *Event) {
	mergedEvents = &Event{
		EventID:       eventsNew.EventID,
		Refresh:       eventsOld.Refresh | eventsNew.Refresh,
		More:          eventsNew.More,
		Messages:      append(eventsOld.Messages, eventsNew.Messages...),
		MessageCounts: append(eventsOld.MessageCounts, eventsNew.MessageCounts...),
		Labels:        append(eventsOld.Labels, eventsNew.Labels...),
		User:          eventsNew.User,
		Addresses:     append(eventsOld.Addresses, eventsNew.Addresses...),
		Notices:       append(eventsOld.Notices, eventsNew.Notices...),
	}

	return
}
