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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package syncservice

import (
	"context"
	"io"

	"github.com/ProtonMail/go-proton-api"
)

type APIClient interface {
	GetGroupedMessageCount(ctx context.Context) ([]proton.MessageGroupCount, error)
	GetLabels(ctx context.Context, labelTypes ...proton.LabelType) ([]proton.Label, error)
	GetMessage(ctx context.Context, messageID string) (proton.Message, error)
	GetMessageMetadataPage(ctx context.Context, page, pageSize int, filter proton.MessageFilter) ([]proton.MessageMetadata, error)
	GetFullMessage(ctx context.Context, messageID string, scheduler proton.Scheduler, storageProvider proton.AttachmentAllocator) (proton.FullMessage, error)
	GetAttachmentInto(ctx context.Context, attachmentID string, reader io.ReaderFrom) error
	GetAttachment(ctx context.Context, attachmentID string) ([]byte, error)
}
