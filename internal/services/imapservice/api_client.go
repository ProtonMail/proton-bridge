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
	"context"
	"io"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

type APIClient interface {
	CreateLabel(ctx context.Context, req proton.CreateLabelReq) (proton.Label, error)
	GetLabel(ctx context.Context, labelID string, labelTypes ...proton.LabelType) (proton.Label, error)
	UpdateLabel(ctx context.Context, labelID string, req proton.UpdateLabelReq) (proton.Label, error)
	DeleteLabel(ctx context.Context, labelID string) error
	LabelMessages(ctx context.Context, messageIDs []string, labelID string) error
	UnlabelMessages(ctx context.Context, messageIDs []string, labelID string) error
	GetLabels(ctx context.Context, labelTypes ...proton.LabelType) ([]proton.Label, error)

	GetGroupedMessageCount(ctx context.Context) ([]proton.MessageGroupCount, error)
	GetMessage(ctx context.Context, messageID string) (proton.Message, error)
	GetMessageMetadataPage(ctx context.Context, page, pageSize int, filter proton.MessageFilter) ([]proton.MessageMetadata, error)
	GetAllMessageIDs(ctx context.Context, afterID string) ([]string, error)
	CreateDraft(ctx context.Context, addrKR *crypto.KeyRing, req proton.CreateDraftReq) (proton.Message, error)
	UploadAttachment(ctx context.Context, addrKR *crypto.KeyRing, req proton.CreateAttachmentReq) (proton.Attachment, error)
	ImportMessages(ctx context.Context, addrKR *crypto.KeyRing, workers, buffer int, req ...proton.ImportReq) (proton.ImportResStream, error)
	GetFullMessage(ctx context.Context, messageID string, scheduler proton.Scheduler, storageProvider proton.AttachmentAllocator) (proton.FullMessage, error)
	GetAttachmentInto(ctx context.Context, attachmentID string, reader io.ReaderFrom) error
	GetAttachment(ctx context.Context, attachmentID string) ([]byte, error)
	DeleteMessage(ctx context.Context, messageIDs ...string) error
	MarkMessagesRead(ctx context.Context, messageIDs ...string) error
	MarkMessagesUnread(ctx context.Context, messageIDs ...string) error
	MarkMessagesForwarded(ctx context.Context, messageIDs ...string) error
	MarkMessagesUnForwarded(ctx context.Context, messageIDs ...string) error
}
