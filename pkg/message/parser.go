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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/jaytaylor/html2text"
)

func parseAttachment(filename string, mediaType string, h textproto.MIMEHeader) (att *pmapi.Attachment) {
	if decoded, err := pmmime.DecodeHeader(filename); err == nil {
		filename = decoded
	}
	if filename == "" {
		ext, err := mime.ExtensionsByType(mediaType)
		if err == nil && len(ext) > 0 {
			filename = "attachment" + ext[0]
		}
	}

	att = &pmapi.Attachment{
		Name:     filename,
		MIMEType: mediaType,
		Header:   h,
	}

	headerContentID := strings.Trim(h.Get("Content-Id"), " <>")

	if headerContentID != "" {
		att.ContentID = headerContentID
	}

	return
}

var reEmailComment = regexp.MustCompile("[(][^)]*[)]") //nolint[gochecknoglobals]

// parseAddressComment removes the comments completely even though they should be allowed
// http://tools.wordtothewise.com/rfc/822
// NOTE: This should be supported in go>1.10 but it seems it's not ¯\_(ツ)_/¯
func parseAddressComment(raw string) string {
	return reEmailComment.ReplaceAllString(raw, "")
}

// Some clients incorrectly format messages with embedded attachments to have a format like
// I. text/plain II. attachment III. text/plain
// which we need to convert to a single HTML part with an embedded attachment.
func combineParts(m *pmapi.Message, parts []io.Reader, headers []textproto.MIMEHeader, convertPlainToHTML bool, atts *[]io.Reader) (isHTML bool, err error) { //nolint[funlen]
	isHTML = true
	foundText := false

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		h := headers[i]

		disp, dispParams, _ := pmmime.ParseMediaType(h.Get("Content-Disposition"))

		d := pmmime.DecodeContentEncoding(part, h.Get("Content-Transfer-Encoding"))
		if d == nil {
			log.Warnf("Unsupported Content-Transfer-Encoding '%v'", h.Get("Content-Transfer-Encoding"))
			d = part
		}

		contentType := h.Get("Content-Type")
		if contentType == "" {
			contentType = "text/plain"
		}
		mediaType, params, _ := pmmime.ParseMediaType(contentType)

		if strings.HasPrefix(mediaType, "text/") && mediaType != "text/calendar" && disp != "attachment" {
			// This is text.
			var b []byte
			if b, err = ioutil.ReadAll(d); err != nil {
				continue
			}
			b, err = pmmime.DecodeCharset(b, contentType)
			if err != nil {
				log.Warn("Decode charset error: ", err)
				return false, err
			}
			contents := string(b)
			if strings.Contains(mediaType, "text/plain") && len(contents) > 0 {
				if !convertPlainToHTML {
					isHTML = false
				} else {
					contents = plaintextToHTML(contents)
				}
			} else if strings.Contains(mediaType, "text/html") && len(contents) > 0 {
				contents, err = stripHTML(contents)
				if err != nil {
					return isHTML, err
				}
			}
			m.Body = contents + m.Body
			foundText = true
		} else {
			// This is an attachment.
			filename := dispParams["filename"]
			if filename == "" {
				// Using "name" in Content-Type is discouraged.
				filename = params["name"]
			}
			if filename == "" && mediaType == "text/calendar" {
				filename = "event.ics"
			}

			att := parseAttachment(filename, mediaType, h)

			b := &bytes.Buffer{}
			if d == nil {
				continue
			}
			if _, err = io.Copy(b, d); err != nil {
				continue
			}
			if foundText && att.ContentID == "" && strings.Contains(mediaType, "image") {
				// Treat this as an inline attachment even though it is not marked as one.
				hasher := sha256.New()
				_, _ = hasher.Write([]byte(att.Name + strconv.Itoa(b.Len())))
				bytes := hasher.Sum(nil)
				cid := hex.EncodeToString(bytes) + "@protonmail.com"

				att.ContentID = cid
				embeddedHTML := makeEmbeddedImageHTML(cid, att.Name)
				m.Body = embeddedHTML + m.Body
			}

			m.Attachments = append(m.Attachments, att)
			*atts = append(*atts, b)
		}
	}
	if isHTML {
		m.Body = addOuterHTMLTags(m.Body)
	}
	return isHTML, nil
}

func checkHeaders(headers []textproto.MIMEHeader) bool {
	foundAttachment := false

	for i := 0; i < len(headers); i++ {
		h := headers[i]

		mediaType, _, _ := pmmime.ParseMediaType(h.Get("Content-Type"))

		if !strings.HasPrefix(mediaType, "text/") {
			foundAttachment = true
		} else if foundAttachment {
			// This means that there is a text part after the first attachment,
			// so we will have to convert the body from plain->HTML.
			return true
		}
	}
	return false
}

// ============================== 7bit Filter ==========================
// For every MIME part in the tree that has "8bit" or "binary" content
// transfer encoding: transcode it to "quoted-printable".

type SevenBitFilter struct {
	target pmmime.VisitAcceptor
}

func NewSevenBitFilter(targetAccepter pmmime.VisitAcceptor) *SevenBitFilter {
	return &SevenBitFilter{
		target: targetAccepter,
	}
}

func decodePart(partReader io.Reader, header textproto.MIMEHeader) (decodedPart io.Reader) {
	decodedPart = pmmime.DecodeContentEncoding(partReader, header.Get("Content-Transfer-Encoding"))
	if decodedPart == nil {
		log.Warnf("Unsupported Content-Transfer-Encoding '%v'", header.Get("Content-Transfer-Encoding"))
		decodedPart = partReader
	}
	return
}

func (sd SevenBitFilter) Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSibling bool, isFirst, isLast bool) error {
	cte := strings.ToLower(header.Get("Content-Transfer-Encoding"))
	if isFirst && pmmime.IsLeaf(header) && cte != "quoted-printable" && cte != "base64" && cte != "7bit" {
		decodedPart := decodePart(partReader, header)

		filteredHeader := textproto.MIMEHeader{}
		for k, v := range header {
			filteredHeader[k] = v
		}
		filteredHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		filteredBuffer := &bytes.Buffer{}
		decodedSlice, _ := ioutil.ReadAll(decodedPart)
		w := quotedprintable.NewWriter(filteredBuffer)
		if _, err := w.Write(decodedSlice); err != nil {
			log.Errorf("cannot write quotedprintable from %q: %v", cte, err)
		}
		if err := w.Close(); err != nil {
			log.Errorf("cannot close quotedprintable from %q: %v", cte, err)
		}

		_ = sd.target.Accept(filteredBuffer, filteredHeader, hasPlainSibling, true, isLast)
	} else {
		_ = sd.target.Accept(partReader, header, hasPlainSibling, isFirst, isLast)
	}
	return nil
}

// =================== HTML Only convertor ==================================
// In any part of MIME tree structure, replace standalone text/html with
// multipart/alternative containing both text/html and text/plain.

type HTMLOnlyConvertor struct {
	target pmmime.VisitAcceptor
}

func NewHTMLOnlyConvertor(targetAccepter pmmime.VisitAcceptor) *HTMLOnlyConvertor {
	return &HTMLOnlyConvertor{
		target: targetAccepter,
	}
}

func randomBoundary() string {
	buf := make([]byte, 30)

	// We specifically use `math/rand` here to allow the generator to be seeded for test purposes.
	// The random numbers need not be cryptographically secure; we are simply generating random part boundaries.
	if _, err := rand.Read(buf); err != nil { // nolint[gosec]
		panic(err)
	}

	return fmt.Sprintf("%x", buf)
}

func (hoc HTMLOnlyConvertor) Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSiblings bool, isFirst, isLast bool) error {
	mediaType, _, err := pmmime.ParseMediaType(header.Get("Content-Type"))
	if isFirst && err == nil && mediaType == "text/html" && !hasPlainSiblings {
		multiPartHeaders := make(textproto.MIMEHeader)
		for k, v := range header {
			multiPartHeaders[k] = v
		}
		boundary := randomBoundary()
		multiPartHeaders.Set("Content-Type", "multipart/alternative; boundary=\""+boundary+"\"")
		childCte := header.Get("Content-Transfer-Encoding")

		_ = hoc.target.Accept(partReader, multiPartHeaders, false, true, false)

		partData, _ := ioutil.ReadAll(partReader)

		htmlChildHeaders := make(textproto.MIMEHeader)
		htmlChildHeaders.Set("Content-Transfer-Encoding", childCte)
		htmlChildHeaders.Set("Content-Type", "text/html")
		htmlReader := bytes.NewReader(partData)
		_ = hoc.target.Accept(htmlReader, htmlChildHeaders, false, true, false)

		_ = hoc.target.Accept(partReader, multiPartHeaders, hasPlainSiblings, false, false)

		plainChildHeaders := make(textproto.MIMEHeader)
		plainChildHeaders.Set("Content-Transfer-Encoding", childCte)
		plainChildHeaders.Set("Content-Type", "text/plain")
		unHtmlized, err := html2text.FromReader(bytes.NewReader(partData))
		if err != nil {
			unHtmlized = string(partData)
		}
		plainReader := strings.NewReader(unHtmlized)
		_ = hoc.target.Accept(plainReader, plainChildHeaders, false, true, true)

		_ = hoc.target.Accept(partReader, multiPartHeaders, hasPlainSiblings, false, true)
	} else {
		_ = hoc.target.Accept(partReader, header, hasPlainSiblings, isFirst, isLast)
	}
	return nil
}

// ======= Public Key Attacher ========

type PublicKeyAttacher struct {
	target                pmmime.VisitAcceptor
	attachedPublicKey     string
	attachedPublicKeyName string
	appendToMultipart     bool
	depth                 int
}

func NewPublicKeyAttacher(targetAccepter pmmime.VisitAcceptor, attachedPublicKey, attachedPublicKeyName string) *PublicKeyAttacher {
	return &PublicKeyAttacher{
		target:                targetAccepter,
		attachedPublicKey:     attachedPublicKey,
		attachedPublicKeyName: attachedPublicKeyName,
		appendToMultipart:     false,
		depth:                 0,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func split(input string, sliceLength int) string {
	processed := input
	result := ""
	for len(processed) > 0 {
		cutPoint := min(sliceLength, len(processed))
		part := processed[0:cutPoint]
		result = result + part + "\n"
		processed = processed[cutPoint:]
	}
	return result
}

func createKeyAttachment(publicKey, publicKeyName string) (headers textproto.MIMEHeader, contents io.Reader) {
	attachmentHeaders := make(textproto.MIMEHeader)
	attachmentHeaders.Set("Content-Type", "application/pgp-key; name=\""+publicKeyName+"\"")
	attachmentHeaders.Set("Content-Transfer-Encoding", "base64")
	attachmentHeaders.Set("Content-Disposition", "attachment; filename=\""+publicKeyName+".asc.pgp\"")

	buffer := &bytes.Buffer{}
	w := base64.NewEncoder(base64.StdEncoding, buffer)
	_, _ = w.Write([]byte(publicKey))
	_ = w.Close()

	return attachmentHeaders, strings.NewReader(split(buffer.String(), 73))
}

func (pka *PublicKeyAttacher) Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSiblings bool, isFirst, isLast bool) error {
	if isFirst && !pmmime.IsLeaf(header) {
		pka.depth++
	}
	if isLast && !pmmime.IsLeaf(header) {
		defer func() {
			pka.depth--
		}()
	}
	isRoot := (header.Get("From") != "")

	// NOTE: This should also work for unspecified Content-Type (in which case us-ascii text/plain is assumed)!
	mediaType, _, err := pmmime.ParseMediaType(header.Get("Content-Type"))
	if isRoot && isFirst && err == nil && pka.attachedPublicKey != "" { //nolint[gocritic]
		if strings.HasPrefix(mediaType, "multipart/mixed") {
			pka.appendToMultipart = true
			_ = pka.target.Accept(partReader, header, hasPlainSiblings, isFirst, isLast)
		} else {
			// Create two siblings with attachment in the case toplevel is not multipart/mixed.
			multiPartHeaders := make(textproto.MIMEHeader)
			for k, v := range header {
				multiPartHeaders[k] = v
			}
			boundary := randomBoundary()
			multiPartHeaders.Set("Content-Type", "multipart/mixed; boundary=\""+boundary+"\"")
			multiPartHeaders.Del("Content-Transfer-Encoding")

			_ = pka.target.Accept(partReader, multiPartHeaders, false, true, false)

			originalHeader := make(textproto.MIMEHeader)
			originalHeader.Set("Content-Type", header.Get("Content-Type"))
			if header.Get("Content-Transfer-Encoding") != "" {
				originalHeader.Set("Content-Transfer-Encoding", header.Get("Content-Transfer-Encoding"))
			}

			_ = pka.target.Accept(partReader, originalHeader, false, true, false)
			_ = pka.target.Accept(partReader, multiPartHeaders, hasPlainSiblings, false, false)

			attachmentHeaders, attachmentReader := createKeyAttachment(pka.attachedPublicKey, pka.attachedPublicKeyName)

			_ = pka.target.Accept(attachmentReader, attachmentHeaders, false, true, true)
			_ = pka.target.Accept(partReader, multiPartHeaders, hasPlainSiblings, false, true)
		}
	} else if isLast && pka.depth == 1 && pka.attachedPublicKey != "" {
		_ = pka.target.Accept(partReader, header, hasPlainSiblings, isFirst, false)
		attachmentHeaders, attachmentReader := createKeyAttachment(pka.attachedPublicKey, pka.attachedPublicKeyName)
		_ = pka.target.Accept(attachmentReader, attachmentHeaders, hasPlainSiblings, true, true)
		_ = pka.target.Accept(partReader, header, hasPlainSiblings, isFirst, true)
	} else {
		_ = pka.target.Accept(partReader, header, hasPlainSiblings, isFirst, isLast)
	}
	return nil
}

// ======= Parser ==========

func Parse(r io.Reader, attachedPublicKey, attachedPublicKeyName string) (m *pmapi.Message, mimeBody string, plainContents string, atts []io.Reader, err error) {
	secondReader := new(bytes.Buffer)
	_, _ = secondReader.ReadFrom(r)

	mimeBody = secondReader.String()

	mm, err := mail.ReadMessage(secondReader)
	if err != nil {
		return
	}

	if m, err = parseHeader(mm.Header); err != nil {
		return
	}

	h := textproto.MIMEHeader(m.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)
	if err != nil {
		return
	}

	printAccepter := pmmime.NewMIMEPrinter()

	publicKeyAttacher := NewPublicKeyAttacher(printAccepter, attachedPublicKey, attachedPublicKeyName)
	sevenBitFilter := NewSevenBitFilter(publicKeyAttacher)

	plainTextCollector := pmmime.NewPlainTextCollector(sevenBitFilter)
	htmlOnlyConvertor := NewHTMLOnlyConvertor(plainTextCollector)

	visitor := pmmime.NewMimeVisitor(htmlOnlyConvertor)
	err = pmmime.VisitAll(bytes.NewReader(mmBodyData), h, visitor)
	/*
		err = visitor.VisitAll(h, bytes.NewReader(mmBodyData))
	*/
	if err != nil {
		return
	}

	mimeBody = printAccepter.String()

	plainContents = plainTextCollector.GetPlainText()

	parts, headers, err := pmmime.GetAllChildParts(bytes.NewReader(mmBodyData), h)

	if err != nil {
		return
	}

	convertPlainToHTML := checkHeaders(headers)
	isHTML, err := combineParts(m, parts, headers, convertPlainToHTML, &atts)

	if isHTML {
		m.MIMEType = "text/html"
	} else {
		m.MIMEType = "text/plain"
	}

	return m, mimeBody, plainContents, atts, err
}
