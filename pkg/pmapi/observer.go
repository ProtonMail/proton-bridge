// Copyright (c) 2022 Proton Technologies AG
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

type ConnectionObserver interface {
	OnDown()
	OnUp()
}

type observer struct {
	onDown, onUp func()
}

// NewConnectionObserver is a helper function to create a new connection observer from two callbacks.
// It doesn't need to be used; anything which implements the ConnectionObserver interface can be an observer.
func NewConnectionObserver(onDown, onUp func()) ConnectionObserver {
	return &observer{
		onDown: onDown,
		onUp:   onUp,
	}
}

func (o observer) OnDown() { o.onDown() }

func (o observer) OnUp() { o.onUp() }
