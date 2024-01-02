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

package imapservice

import (
	"bytes"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
)

type SyncMessageBuilder struct {
	state *rwIdentity
}

func NewSyncMessageBuilder(rw *rwIdentity) *SyncMessageBuilder {
	return &SyncMessageBuilder{state: rw}
}

func (s SyncMessageBuilder) WithKeys(f func(*crypto.KeyRing, map[string]*crypto.KeyRing) error) error {
	return s.state.WithAddrKRs(f)
}

func (s SyncMessageBuilder) BuildMessage(
	apiLabels map[string]proton.Label,
	full proton.FullMessage,
	addrKR *crypto.KeyRing,
	buffer *bytes.Buffer,
) (syncservice.BuildResult, error) {
	buffer.Grow(full.Size)

	if err := message.DecryptAndBuildRFC822Into(addrKR, full.Message, full.AttData, defaultMessageJobOpts(), buffer); err != nil {
		return syncservice.BuildResult{}, err
	}

	update, err := newMessageCreatedUpdate(apiLabels, full.MessageMetadata, buffer.Bytes())
	if err != nil {
		return syncservice.BuildResult{}, err
	}

	return syncservice.BuildResult{
		AddressID: full.Message.AddressID,
		MessageID: full.Message.ID,
		Update:    update,
	}, nil
}
