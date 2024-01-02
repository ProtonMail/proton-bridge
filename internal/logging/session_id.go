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

package logging

import (
	"fmt"
	"strconv"
	"time"
)

type SessionID string

const (
	timeFormat = "20060102_150405" // time format in Go does not support milliseconds without dot, so we'll process it manually.
)

// NewSessionID creates a sessionID based on the current time.
func NewSessionID() SessionID {
	now := time.Now()
	return SessionID(now.Format(timeFormat) + fmt.Sprintf("%03d", now.Nanosecond()/1000000))
}

// NewSessionIDFromString Return a new sessionID from string. If the str is empty a new time based sessionID is returned, otherwise the string
// is used as the sessionID.
func NewSessionIDFromString(str string) SessionID {
	if (len(str)) > 0 {
		return SessionID(str)
	}

	return NewSessionID()
}

// toTime converts a sessionID to a date/Time, considering the time zone is local.
func (s SessionID) toTime() time.Time {
	if len(s) < 3 {
		return time.Time{}
	}

	t, err := time.ParseInLocation(timeFormat, string(s)[:len(s)-3], time.Local)
	if err != nil {
		return time.Time{}
	}

	var ms int
	if ms, err = strconv.Atoi(string(s)[len(s)-3:]); err != nil {
		return time.Time{}
	}

	return t.Add(time.Duration(ms) * time.Millisecond)
}
