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

package logging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogging_SessionID(t *testing.T) {
	now := time.Now()
	sessionID := NewSessionID()
	sessionTime := sessionID.toTime()
	require.False(t, sessionTime.IsZero())
	require.WithinRange(t, sessionTime, now.Add(-1*time.Millisecond), now.Add(1*time.Millisecond))

	fromString := NewSessionIDFromString("")
	require.True(t, len(fromString) > 0)
	fromString = NewSessionIDFromString(string(sessionID))
	require.True(t, len(fromString) > 0)
	require.Equal(t, sessionID, fromString)
}
