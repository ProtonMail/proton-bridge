// Copyright (c) 2025 Proton AG
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

package syncservice

import (
	"bytes"
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/bradenaw/juniper/xmaps"
)

type StateProvider interface {
	AddFailedMessageID(context.Context, ...string) error
	RemFailedMessageID(context.Context, ...string) error
	GetSyncStatus(context.Context) (Status, error)
	ClearSyncStatus(context.Context) error
	SetHasLabels(context.Context, bool) error
	SetHasMessages(context.Context, bool) error
	SetLastMessageID(context.Context, string, int64) error
	SetMessageCount(context.Context, int64) error
}

type Status struct {
	HasLabels           bool
	HasMessages         bool
	HasMessageCount     bool
	FailedMessages      xmaps.Set[string]
	LastSyncedMessageID string
	NumSyncedMessages   int64
	TotalMessageCount   int64
}

func DefaultStatus() Status {
	return Status{
		FailedMessages: make(map[string]struct{}),
	}
}

func (s Status) IsComplete() bool {
	return s.HasLabels && s.HasMessages
}

func (s Status) InProgress() bool {
	return s.HasLabels || s.HasMessageCount
}

// Regulator is an abstraction for the sync service, since it regulates the number of concurrent sync activities.
type Regulator interface {
	Sync(ctx context.Context, stage *Job) error
}

type BuildResult struct {
	AddressID string
	MessageID string
	Update    *imap.MessageCreated
}

type MessageBuilder interface {
	WithKeys(f func(*crypto.KeyRing, map[string]*crypto.KeyRing) error) error
	BuildMessage(apiLabels map[string]proton.Label, full proton.FullMessage, addrKR *crypto.KeyRing, buffer *bytes.Buffer) (BuildResult, error)
}

type UpdateApplier interface {
	ApplySyncUpdates(ctx context.Context, updates []BuildResult) error
	SyncLabels(ctx context.Context, labels map[string]proton.Label) error
}

type Reporter interface {
	OnStart(ctx context.Context)
	OnFinished(ctx context.Context)
	OnError(ctx context.Context, err error)
	OnProgress(ctx context.Context, delta int64)
	InitializeProgressCounter(ctx context.Context, current int64, total int64)
}
