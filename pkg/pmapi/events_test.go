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
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestClient_GetEvent(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/events/latest"))

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, testEventBody)
	}))
	defer s.Close()

	event, err := c.GetEvent(context.Background(), "")
	r.NoError(t, err)
	r.Equal(t, testEvent, event)
}

func TestClient_GetEvent_withID(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/events/"+testEvent.EventID))

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, testEventBody)
	}))
	defer s.Close()

	event, err := c.GetEvent(context.Background(), testEvent.EventID)
	r.NoError(t, err)
	r.Equal(t, testEvent, event)
}

// We first call GetEvent with id of eventID1, which returns More=1 so we fetch with id eventID2.
func TestClient_GetEvent_mergeEvents(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch req.URL.RequestURI() {
		case "/events/eventID1":
			r.NoError(t, checkMethodAndPath(req, "GET", "/events/eventID1"))
			fmt.Fprint(w, testEventBodyMore1)
		case "/events/eventID2":
			r.NoError(t, checkMethodAndPath(req, "GET", "/events/eventID2"))
			fmt.Fprint(w, testEventBodyMore2)
		default:
			t.Fail()
		}
	}))
	defer s.Close()

	event, err := c.GetEvent(context.Background(), "eventID1")
	r.NoError(t, err)
	r.Equal(t, testEventMerged, event)
}

func TestClient_GetEvent_mergeMaxNumberOfEvents(t *testing.T) {
	numberOfCalls := 0

	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		numberOfCalls++

		re := regexp.MustCompile(`/eventID([0-9]+)`)
		eventIDString := re.FindStringSubmatch(req.URL.RequestURI())[1]
		eventID, err := strconv.Atoi(eventIDString)
		r.NoError(t, err)

		if numberOfCalls > maxNumberOfMergedEvents*2 {
			r.Fail(t, "Too many calls!")
		}

		body := strings.ReplaceAll(testEventBodyMore1, "eventID2", "eventID"+strconv.Itoa(eventID+1))
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, body)
	}))
	defer s.Close()

	event, err := c.GetEvent(context.Background(), "eventID1")
	r.NoError(t, err)
	r.Equal(t, maxNumberOfMergedEvents, numberOfCalls)
	r.True(t, bool(event.More))
}

var (
	testEventMessageUpdateUnread = Boolean(false)

	testEvent = &Event{
		EventID: "eventID1",
		Refresh: 0,
		Messages: []*EventMessage{
			{
				EventItem: EventItem{ID: "hdI7aIgUO1hFplCIcJHB0jShRVsAzS0AB75wGCaiNVeIHXLmaUnt4eJ8l7c7L6uk4g0ZdXhGWG5gfh6HHgAZnw==", Action: EventCreate},
				Created: &Message{
					ID:      "hdI7aIgUO1hFplCIcJHB0jShRVsAzS0AB75wGCaiNVeIHXLmaUnt4eJ8l7c7L6uk4g0ZdXhGWG5gfh6HHgAZnw==",
					Header:  make(mail.Header),
					Subject: "Hey there",
				},
			},
			{
				EventItem: EventItem{ID: "bSFLAimPSfGz2Kj0aV3l3AyXsof_Vf7sfrrMJ8ifgGJe-f2NG2eLaEGXLytjMhq9wnLMtkoZpO2uBXM4nOVa5g==", Action: EventUpdateFlags},
				Updated: &EventMessageUpdated{
					ID:              "bSFLAimPSfGz2Kj0aV3l3AyXsof_Vf7sfrrMJ8ifgGJe-f2NG2eLaEGXLytjMhq9wnLMtkoZpO2uBXM4nOVa5g==",
					Unread:          &testEventMessageUpdateUnread,
					Time:            1472391377,
					LabelIDsAdded:   []string{ArchiveLabel},
					LabelIDsRemoved: []string{InboxLabel},
				},
			},
			{
				EventItem: EventItem{ID: "XRBMBYnSkaEJWtqFACp2kjlNc-7GjzX3SnPcOtWK4PyLG11Nhsg0uxPYjTXoClQfB-EHVDl9gE3w2PVuj93jBg==", Action: EventDelete},
			},
		},
		MessageCounts: []*MessagesCount{
			{
				LabelID: "0",
				Total:   19,
				Unread:  2,
			},
			{
				LabelID: "6",
				Total:   1,
				Unread:  0,
			},
		},
		Notices: []string{"Server will be down in 2min because of a NSA attack"},
	}

	testEventMerged = &Event{
		EventID: "eventID3",
		Refresh: 1,
		Messages: []*EventMessage{
			{
				EventItem: EventItem{ID: "msgID1", Action: EventCreate},
				Created: &Message{
					ID:      "id",
					Header:  make(mail.Header),
					Subject: "Hey there",
				},
			},
			{
				EventItem: EventItem{ID: "msgID2", Action: EventCreate},
				Created: &Message{
					ID:      "id",
					Header:  make(mail.Header),
					Subject: "Hey there again",
				},
			},
		},
		MessageCounts: []*MessagesCount{
			{
				LabelID: "label1",
				Total:   19,
				Unread:  2,
			},
			{
				LabelID: "label2",
				Total:   1,
				Unread:  0,
			},
			{
				LabelID: "label2",
				Total:   2,
				Unread:  1,
			},
			{
				LabelID: "label3",
				Total:   1,
				Unread:  0,
			},
		},
		Notices: []string{"Server will be down in 2min because of a NSA attack", "Just kidding lol"},
		Labels: []*EventLabel{
			{
				EventItem: EventItem{
					ID:     "labelID1",
					Action: 1,
				},
				Label: &Label{
					ID:   "id",
					Name: "Event Label 1",
				},
			},
			{
				EventItem: EventItem{
					ID:     "labelID2",
					Action: 1,
				},
				Label: &Label{
					ID:   "id",
					Name: "Event Label 2",
				},
			},
		},
		User: &User{
			ID:        "userID1",
			Name:      "user",
			UsedSpace: &usedSpace,
			MaxSpace:  &maxSpace,
		},
		Addresses: []*EventAddress{
			{
				EventItem: EventItem{
					ID:     "addressID1",
					Action: 2,
				},
				Address: &Address{
					ID:          "id",
					DisplayName: "address 1",
				},
			},
			{
				EventItem: EventItem{
					ID:     "addressID2",
					Action: 2,
				},
				Address: &Address{
					ID:          "id",
					DisplayName: "address 2",
				},
			},
		},
	}
)

const (
	testEventBody = `{
    "EventID": "eventID1",
    "Refresh": 0,
    "Messages": [
        {
            "ID": "hdI7aIgUO1hFplCIcJHB0jShRVsAzS0AB75wGCaiNVeIHXLmaUnt4eJ8l7c7L6uk4g0ZdXhGWG5gfh6HHgAZnw==",
            "Action": 1,
            "Message": {
                "ID": "hdI7aIgUO1hFplCIcJHB0jShRVsAzS0AB75wGCaiNVeIHXLmaUnt4eJ8l7c7L6uk4g0ZdXhGWG5gfh6HHgAZnw==",
                "Subject": "Hey there"
            }
        },
        {
            "ID": "bSFLAimPSfGz2Kj0aV3l3AyXsof_Vf7sfrrMJ8ifgGJe-f2NG2eLaEGXLytjMhq9wnLMtkoZpO2uBXM4nOVa5g==",
            "Action": 3,
            "Message": {
                "ConversationID": "2oX3EILYRuZ0IRBVlzMg1oV5eazQL67sFIHlcR8bjickPn7K4id4sJZuAB6n0pdtI3hRIVsjCpgWfRm8c_x3IQ==",
                "Unread": 0,
                "Time": 1472391377,
                "Location": 6,
                "LabelIDsAdded": [
                    "6"
                ],
                "LabelIDsRemoved": [
                    "0"
                ]
            }
        },
        {
            "ID": "XRBMBYnSkaEJWtqFACp2kjlNc-7GjzX3SnPcOtWK4PyLG11Nhsg0uxPYjTXoClQfB-EHVDl9gE3w2PVuj93jBg==",
            "Action": 0
        }
    ],
    "Conversations": [
        {
            "ID": "2oX3EILYRuZ0IRBVlzMg1oV5eazQL67sFIHlcR8bjickPn7K4id4sJZuAB6n0pdtI3hRIVsjCpgWfRm8c_x3IQ==",
            "Action": 1,
            "Conversation": {
                "ID": "2oX3EILYRuZ0IRBVlzMg1oV5eazQL67sFIHlcR8bjickPn7K4id4sJZuAB6n0pdtI3hRIVsjCpgWfRm8c_x3IQ==",
                "Order": 1616,
                "Subject": "Hey there",
                "Senders": [
                    {
                        "Address": "apple@protonmail.com",
                        "Name": "apple@protonmail.com"
                    }
                ],
                "Recipients": [
                    {
                        "Address": "apple@protonmail.com",
                        "Name": "apple@protonmail.com"
                    }
                ],
                "NumMessages": 1,
                "NumUnread": 1,
                "NumAttachments": 0,
                "ExpirationTime": 0,
                "TotalSize": 636,
                "AddressID": "QMJs2dzTx7uqpH5PNgIzjULywU4gO9uMBhEMVFOAVJOoUml54gC0CCHtW9qYwzH-zYbZwMv3MFYncPjW1Usq7Q==",
                "LabelIDs": [
                    "0"
                ],
                "Labels": [
                    {
                        "Count": 1,
                        "NumMessages": 1,
                        "NumUnread": 1,
                        "ID": "0"
                    }
                ]
            }
        }
    ],
    "Total": {
        "Locations": [
            {
                "Location": 0,
                "Count": 19
            },
            {
                "Location": 1,
                "Count": 16
            },
            {
                "Location": 2,
                "Count": 16
            },
            {
                "Location": 3,
                "Count": 17
            },
            {
                "Location": 6,
                "Count": 1
            }
        ],
        "Labels": [
            {
                "LabelID": "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==",
                "Count": 2
            },
            {
                "LabelID": "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==",
                "Count": 2
            }
        ],
        "Starred": 3
    },
    "Unread": {
        "Locations": [
            {
                "Location": 0,
                "Count": 2
            },
            {
                "Location": 1,
                "Count": 0
            },
            {
                "Location": 2,
                "Count": 0
            },
            {
                "Location": 3,
                "Count": 0
            },
            {
                "Location": 6,
                "Count": 0
            }
        ],
        "Labels": [
            {
                "LabelID": "LLz8ysmVxwr4dF6mWpClePT0SpSWOEvzTdq17RydSl4ndMckvY1K63HeXDzn03BJQwKYvgf-eWT8Qfd9WVuIEQ==",
                "Count": 0
            },
            {
                "LabelID": "BvbqbySUPo9uWW_eR8tLA13NUsQMz3P4Zhw4UnpvrKqURnrHlE6L2Au0nplHfHlVXFgGz4L4hJ9-BYllOL-L5g==",
                "Count": 0
            }
        ],
        "Starred": 0
    },
    "MessageCounts": [
        {
            "LabelID": "0",
            "Total": 19,
            "Unread": 2
        },
        {
            "LabelID": "6",
            "Total": 1,
            "Unread": 0
        }
    ],
    "ConversationCounts": [
        {
            "LabelID": "0",
            "Total": 19,
            "Unread": 2
        },
        {
            "LabelID": "6",
            "Total": 1,
            "Unread": 0
        }
    ],
    "Notices": ["Server will be down in 2min because of a NSA attack"],
    "Code": 1000
}
`

	testEventBodyMore1 = `{
    "EventID": "eventID2",
		"More": 1,
    "Refresh": 1,
    "Messages": [
        {
            "ID": "msgID1",
            "Action": 1,
            "Message": {
                "ID": "id",
                "Subject": "Hey there"
            }
        }
    ],
    "MessageCounts": [
        {
            "LabelID": "label1",
            "Total": 19,
            "Unread": 2
        },
        {
            "LabelID": "label2",
            "Total": 1,
            "Unread": 0
        }
    ],
		"Labels": [
        {
            "ID":"labelID1",
            "Action":1,
            "Label":{
                "ID":"id",
                "Name":"Event Label 1"
            }
        }
    ],
		"User": {
        "ID": "userID1",
        "Name": "user",
        "UsedSpace": 444,
        "MaxSpace": 12345678
    },
		"Addresses": [
        {
            "ID": "addressID1",
            "Action": 2,
            "Address": {
                "ID": "id",
                "DisplayName": "address 1"
            }
        }
    ],
    "UsedSpace": 12345,
    "Notices": ["Server will be down in 2min because of a NSA attack"]
}
`

	testEventBodyMore2 = `{
    "EventID": "eventID3",
    "Refresh": 0,
    "Messages": [
        {
            "ID": "msgID2",
            "Action": 1,
            "Message": {
                "ID": "id",
                "Subject": "Hey there again"
            }
        }
    ],
    "MessageCounts": [
        {
            "LabelID": "label2",
            "Total": 2,
            "Unread": 1
        },
        {
            "LabelID": "label3",
            "Total": 1,
            "Unread": 0
        }
    ],
		"Labels": [
        {
            "ID":"labelID2",
            "Action":1,
            "Label":{
                "ID":"id",
                "Name":"Event Label 2"
            }
        }
    ],
		"User": {
        "ID": "userID1",
        "Name": "user",
        "UsedSpace": 23456
    },
		"Addresses": [
        {
            "ID": "addressID2",
            "Action": 2,
            "Address": {
                "ID": "id",
                "DisplayName": "address 2"
            }
        }
    ],
    "Notices": ["Just kidding lol"]
}
`
)
