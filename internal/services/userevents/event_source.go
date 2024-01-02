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

package userevents

import (
	"context"

	"github.com/ProtonMail/go-proton-api"
)

// EventSource represents a type which produces proton events.
type EventSource interface {
	GetLatestEventID(ctx context.Context) (string, error)
	GetEvent(ctx context.Context, id string) ([]proton.Event, bool, error)
}

type NullEventSource struct{}

func (n NullEventSource) GetLatestEventID(_ context.Context) (string, error) {
	return "", nil
}

func (n NullEventSource) GetEvent(_ context.Context, _ string) ([]proton.Event, bool, error) {
	return nil, false, nil
}
