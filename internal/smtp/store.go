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

package smtp

import (
	"io"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type storeUserProvider interface {
	CreateDraft(
		kr *crypto.KeyRing,
		message *pmapi.Message,
		attachmentReaders []io.Reader,
		attachedPublicKey,
		attachedPublicKeyName string,
		parentID string) (*pmapi.Message, []*pmapi.Attachment, error)
	SendMessage(messageID string, req *pmapi.SendMessageReq) error
	GetMaxUpload() (int64, error)
}
