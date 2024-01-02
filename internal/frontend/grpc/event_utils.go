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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package grpc

import "github.com/bradenaw/juniper/xslices"

// isInternetStatus returns true iff the event is InternetStatus.
func (x *StreamEvent) isInternetStatus() bool {
	appEvent := x.GetApp()

	return (appEvent != nil) && (appEvent.GetInternetStatus() != nil)
}

// filterOutInternetStatusEvents return a copy of the events list where all internet connection events have been removed.
func filterOutInternetStatusEvents(events []*StreamEvent) []*StreamEvent {
	return xslices.Filter(events, func(event *StreamEvent) bool { return !event.isInternetStatus() })
}
