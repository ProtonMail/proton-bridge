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

package listener

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const minEventReceiveTime = 100 * time.Millisecond

func Example() {
	eventListener := New()

	ch := make(chan string)
	eventListener.Add("eventname", ch)
	for eventdata := range ch {
		fmt.Println(eventdata + " world")
	}

	eventListener.Emit("eventname", "hello")
}

func TestAddAndEmitSameEvent(t *testing.T) {
	listener, channel := newListener()

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")
}

func TestAddAndEmitDifferentEvent(t *testing.T) {
	listener, channel := newListener()

	listener.Emit("other", "hello!")
	checkChannelNotEmitted(t, channel)
}

func TestAddAndRemove(t *testing.T) {
	listener := New()

	channel := make(chan string)
	listener.Add("event", channel)
	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")

	listener.Remove("event", channel)
	listener.Emit("event", "hello!")

	checkChannelNotEmitted(t, channel)
}

func TestNoLimit(t *testing.T) {
	listener, channel := newListener()

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")
}

func TestLimit(t *testing.T) {
	listener, channel := newListener()
	listener.SetLimit("event", 1*time.Second)

	channel2 := make(chan string)
	listener.Add("event", channel2)

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")
	checkChannelEmitted(t, channel2, "hello!")

	listener.Emit("event", "hello!")
	checkChannelNotEmitted(t, channel)
	checkChannelNotEmitted(t, channel2)

	time.Sleep(1 * time.Second)

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")
	checkChannelEmitted(t, channel2, "hello!")
}

func TestLimitDifferentData(t *testing.T) {
	listener, channel := newListener()
	listener.SetLimit("event", 1*time.Second)

	listener.Emit("event", "hello!")
	checkChannelEmitted(t, channel, "hello!")

	listener.Emit("event", "hello?")
	checkChannelEmitted(t, channel, "hello?")
}

func TestReEmit(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	listener := New()
	listener.Emit("event", "hello?")

	listener.SetBuffer("event")
	listener.SetBuffer("other")

	listener.Emit("event", "hello1")
	listener.Emit("event", "hello2")
	listener.Emit("other", "hello!")
	listener.Emit("event", "hello3")
	listener.Emit("other", "hello!")

	eventCH := make(chan string, 3)
	listener.Add("event", eventCH)

	otherCH := make(chan string)
	listener.Add("other", otherCH)

	listener.RetryEmit("event")
	listener.RetryEmit("other")
	time.Sleep(time.Millisecond)

	receivedEvents := map[string]int{}
	for i := 0; i < 5; i++ {
		select {
		case res := <-eventCH:
			receivedEvents[res]++
		case res := <-otherCH:
			receivedEvents[res+":other"]++
		case <-time.After(minEventReceiveTime):
			t.Fatalf("Channel not emitted %d times", i+1)
		}
	}
	expectedEvents := map[string]int{"hello1": 1, "hello2": 1, "hello3": 1, "hello!:other": 2}
	require.Equal(t, expectedEvents, receivedEvents)
}

func newListener() (Listener, chan string) {
	listener := New()

	channel := make(chan string)
	listener.Add("event", channel)

	return listener, channel
}

func checkChannelEmitted(t testing.TB, channel chan string, expectedData string) {
	select {
	case res := <-channel:
		require.Equal(t, expectedData, res)
	case <-time.After(minEventReceiveTime):
		t.Fatalf("Channel not emitted with expected data: %s", expectedData)
	}
}

func checkChannelNotEmitted(t testing.TB, channel chan string) {
	select {
	case res := <-channel:
		t.Fatalf("Channel emitted with a unexpected response: %s", res)
	case <-time.After(minEventReceiveTime):
	}
}
