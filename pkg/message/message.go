// Copyright (c) 2021 Proton Technologies AG
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

package message

import (
	"crypto/sha512"
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/sirupsen/logrus"
)

const (
	rfc822Message = "message/rfc822"
)

var log = logrus.WithField("pkg", "pkg/message") //nolint[gochecknoglobals]

func GetBoundary(m *pmapi.Message) string {
	// The boundary needs to be deterministic because messages are not supposed to
	// change.
	return fmt.Sprintf("%x", sha512.Sum512_256([]byte(m.ID)))
}

func GetRelatedBoundary(m *pmapi.Message) string {
	// The boundary needs to be deterministic because messages are not supposed to
	// change.
	return fmt.Sprintf("%x", sha512.Sum512_256([]byte(m.ID+m.ID)))
}

func SeparateInlineAttachments(m *pmapi.Message) (atts, inlines []*pmapi.Attachment) {
	for _, att := range m.Attachments {
		if strings.Contains(att.Header.Get("Content-Disposition"), pmapi.DispositionInline) {
			inlines = append(inlines, att)
		} else {
			atts = append(atts, att)
		}
	}
	return
}
