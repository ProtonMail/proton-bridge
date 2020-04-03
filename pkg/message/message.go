// Copyright (c) 2020 Proton Technologies AG
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

	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/jhillyerd/enmime"
	log "github.com/sirupsen/logrus"
)

const textPlain = "text/plain"

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

func GetBodyStructure(m *pmapi.Message) (bs *imap.BodyStructure) { //nolint[funlen]
	bs = &imap.BodyStructure{
		MIMEType:    "multipart",
		MIMESubType: "mixed",
		Params:      map[string]string{"boundary": GetBoundary(m)},
	}
	var inlineParts []*imap.BodyStructure
	var attParts []*imap.BodyStructure

	for _, att := range m.Attachments {
		typeParts := strings.SplitN(att.MIMEType, "/", 2)
		if len(typeParts) != 2 {
			continue
		}

		if typeParts[0] == "application" && typeParts[1] == "pgp-encrypted" {
			typeParts[1] = "octet-stream"
		}

		part := &imap.BodyStructure{
			MIMEType:    typeParts[0],
			MIMESubType: typeParts[1],
			Params:      map[string]string{"name": att.Name},
			Encoding:    "base64",
		}

		if strings.Contains(att.Header.Get("Content-Disposition"), "inline") {
			part.Disposition = "inline"
			inlineParts = append(inlineParts, part)
		} else {
			part.Disposition = "attachment"
			attParts = append(attParts, part)
		}
	}

	if len(inlineParts) > 0 {
		// Set to multipart-related for inline attachments.
		relatedPart := &imap.BodyStructure{
			MIMEType:    "multipart",
			MIMESubType: "related",
			Params:      map[string]string{"boundary": GetRelatedBoundary(m)},
		}

		subType := "html"

		if m.MIMEType == textPlain {
			subType = "plain"
		}

		relatedPart.Parts = append(relatedPart.Parts, &imap.BodyStructure{
			MIMEType:    "text",
			MIMESubType: subType,
			Params:      map[string]string{"charset": "utf-8"},
			Encoding:    "quoted-printable",
			Disposition: "inline",
		})

		bs.Parts = append(bs.Parts, relatedPart)
	} else {
		subType := "html"

		if m.MIMEType == textPlain {
			subType = "plain"
		}

		bs.Parts = append(bs.Parts, &imap.BodyStructure{
			MIMEType:    "text",
			MIMESubType: subType,
			Params:      map[string]string{"charset": "utf-8"},
			Encoding:    "quoted-printable",
			Disposition: "inline",
		})
	}

	bs.Parts = append(bs.Parts, attParts...)

	return bs
}

func SeparateInlineAttachments(m *pmapi.Message) (atts, inlines []*pmapi.Attachment) {
	for _, att := range m.Attachments {
		if strings.Contains(att.Header.Get("Content-Disposition"), "inline") {
			inlines = append(inlines, att)
		} else {
			atts = append(atts, att)
		}
	}
	return
}

func GetMIMEBodyStructure(m *pmapi.Message, parsedMsg *enmime.Envelope) (bs *imap.BodyStructure, err error) {
	// We recursively look through the MIME structure.
	root := parsedMsg.Root
	if root == nil {
		return GetBodyStructure(m), nil
	}

	mediaType, params, err := pmmime.ParseMediaType(root.ContentType)
	if err != nil {
		log.Warnf("Cannot parse Content-Type '%v': %v", root.ContentType, err)
		err = nil
		mediaType = textPlain
	}

	typeParts := strings.SplitN(mediaType, "/", 2)

	bs = &imap.BodyStructure{
		MIMEType: typeParts[0],
		Params:   params,
	}

	if len(typeParts) > 1 {
		bs.MIMESubType = typeParts[1]
	}

	bs.Parts = getChildrenParts(root)

	return
}

func getChildrenParts(root *enmime.Part) (parts []*imap.BodyStructure) {
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		mediaType, params, err := pmmime.ParseMediaType(child.ContentType)
		if err != nil {
			log.Warnf("Cannot parse Content-Type '%v': %v", child.ContentType, err)
			mediaType = textPlain
		}
		typeParts := strings.SplitN(mediaType, "/", 2)
		childrenParts := getChildrenParts(child)
		part := &imap.BodyStructure{
			MIMEType:    typeParts[0],
			Params:      params,
			Encoding:    child.Charset,
			Disposition: child.Disposition,
			Parts:       childrenParts,
		}
		if len(typeParts) > 1 {
			part.MIMESubType = typeParts[1]
		}
		parts = append(parts, part)
	}
	return
}
