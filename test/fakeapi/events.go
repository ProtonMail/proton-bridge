// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package fakeapi

import (
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (api *FakePMAPI) GetEvent(eventID string) (*pmapi.Event, error) {
	if err := api.checkAndRecordCall(GET, "/events/"+eventID, nil); err != nil {
		return nil, err
	}
	// Request for empty ID returns the latest event.
	if eventID == "" {
		if len(api.events) == 0 {
			return &pmapi.Event{EventID: ""}, nil
		}
		return api.events[len(api.events)-1], nil
	}
	// Otherwise it tries to find specific ID and return all next events merged into one.
	var foundEvent *pmapi.Event
	mergedEvent := &pmapi.Event{}
	for _, event := range api.events {
		if event.EventID == eventID {
			foundEvent = event
			continue
		}
		if foundEvent != nil {
			mergedEvent.EventID = event.EventID
			mergedEvent.Refresh |= event.Refresh
			mergedEvent.Messages = append(mergedEvent.Messages, event.Messages...)
			mergedEvent.MessageCounts = append(mergedEvent.MessageCounts, event.MessageCounts...)
			mergedEvent.Labels = append(mergedEvent.Labels, event.Labels...)
			mergedEvent.Notices = append(mergedEvent.Notices, event.Notices...)
			mergedEvent.User = event.User
		}
	}
	// If there isn't next event, return the same one.
	if mergedEvent.EventID == "" {
		return foundEvent, nil
	}
	return mergedEvent, nil
}

func (api *FakePMAPI) addEventLabel(action pmapi.EventAction, label *pmapi.Label) {
	api.addEvent(&pmapi.Event{
		EventID: api.eventIDGenerator.next("event"),
		Labels: []*pmapi.EventLabel{{
			EventItem: pmapi.EventItem{
				ID:     label.ID,
				Action: action,
			},
			Label: label,
		}},
	})
}

func (api *FakePMAPI) addEventMessage(action pmapi.EventAction, message *pmapi.Message) {
	created := message
	updated := &pmapi.EventMessageUpdated{
		ID:       message.ID,
		Subject:  &message.Subject,
		Unread:   &message.Unread,
		Flags:    &message.Flags,
		Sender:   message.Sender,
		ToList:   &message.ToList,
		CCList:   &message.CCList,
		BCCList:  &message.BCCList,
		Time:     message.Time,
		LabelIDs: message.LabelIDs,
	}
	if action == pmapi.EventCreate {
		updated = nil
	} else {
		created = nil
	}
	api.addEvent(&pmapi.Event{
		EventID: api.eventIDGenerator.next("event"),
		Messages: []*pmapi.EventMessage{{
			EventItem: pmapi.EventItem{
				ID:     message.ID,
				Action: action,
			},
			Created: created,
			Updated: updated,
		}},
		MessageCounts: api.getAllCounts(),
	})
}

func (api *FakePMAPI) addEvent(event *pmapi.Event) {
	api.events = append(api.events, event)
}
