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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"

	"github.com/ProtonMail/gluon/watcher"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
)

type bridgeEventSubscription struct {
	b *Bridge
}

func (b bridgeEventSubscription) Add(ofType ...events.Event) *watcher.Watcher[events.Event] {
	return b.b.addWatcher(ofType...)
}

func (b bridgeEventSubscription) Remove(watcher *watcher.Watcher[events.Event]) {
	b.b.remWatcher(watcher)
}

type bridgeEventPublisher struct {
	b *Bridge
}

func (b bridgeEventPublisher) PublishEvent(_ context.Context, event events.Event) {
	b.b.publish(event)
}
