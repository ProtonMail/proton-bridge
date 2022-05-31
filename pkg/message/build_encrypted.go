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

package message

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-message"
	"github.com/emersion/go-textwrapper"
)

// BuildEncrypted is used for importing encrypted message.
func BuildEncrypted(m *pmapi.Message, readers []io.Reader, kr *crypto.KeyRing) ([]byte, error) { //nolint:funlen
	b := &bytes.Buffer{}
	boundary := newBoundary(m.ID).gen()

	// Overwrite content for main header for import.
	// Even if message has just simple body we should upload as multipart/mixed.
	// Each part has encrypted body and header reflects the original header.
	mainHeader := convertGoMessageToTextprotoHeader(getMessageHeader(m, JobOptions{}))
	mainHeader.Set("Content-Type", "multipart/mixed; boundary="+boundary)
	mainHeader.Del("Content-Disposition")
	mainHeader.Del("Content-Transfer-Encoding")
	if err := WriteHeader(b, mainHeader); err != nil {
		return nil, err
	}
	mw := multipart.NewWriter(b)
	if err := mw.SetBoundary(boundary); err != nil {
		return nil, err
	}

	// Write the body part.
	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", m.MIMEType+"; charset=utf-8")
	bodyHeader.Set("Content-Disposition", pmapi.DispositionInline)
	bodyHeader.Set("Content-Transfer-Encoding", "7bit")

	p, err := mw.CreatePart(bodyHeader)
	if err != nil {
		return nil, err
	}
	// First, encrypt the message body.
	if err := m.Encrypt(kr, kr); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(p, m.Body); err != nil {
		return nil, err
	}

	// Write the attachments parts.
	for i := 0; i < len(m.Attachments); i++ {
		att := m.Attachments[i]
		r := readers[i]
		h := getAttachmentHeader(att, false)
		p, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		// Create encrypted writer.
		pgpMessage, err := kr.Encrypt(crypto.NewPlainMessage(data), nil)
		if err != nil {
			return nil, err
		}

		ww := textwrapper.NewRFC822(p)
		bw := base64.NewEncoder(base64.StdEncoding, ww)
		if _, err := bw.Write(pgpMessage.GetBinary()); err != nil {
			return nil, err
		}
		if err := bw.Close(); err != nil {
			return nil, err
		}
	}

	if err := mw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func convertGoMessageToTextprotoHeader(h message.Header) textproto.MIMEHeader {
	out := make(textproto.MIMEHeader)
	hf := h.Fields()
	for hf.Next() {
		// go-message fields are in the reverse order.
		// textproto.MIMEHeader is not ordered except for the values of
		// the same key which are ordered
		key := textproto.CanonicalMIMEHeaderKey(hf.Key())
		out[key] = append([]string{hf.Value()}, out[key]...)
	}
	return out
}

func getAttachmentHeader(att *pmapi.Attachment, buildForIMAP bool) textproto.MIMEHeader {
	mediaType := att.MIMEType
	if mediaType == "application/pgp-encrypted" {
		mediaType = "application/octet-stream"
	}

	transferEncoding := "base64"
	if mediaType == rfc822Message && buildForIMAP {
		transferEncoding = "8bit"
	}

	encodedName := pmmime.EncodeHeader(att.Name)
	disposition := "attachment" //nolint:goconst
	if strings.Contains(att.Header.Get("Content-Disposition"), pmapi.DispositionInline) {
		disposition = pmapi.DispositionInline
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", mime.FormatMediaType(mediaType, map[string]string{"name": encodedName}))
	if transferEncoding != "" {
		h.Set("Content-Transfer-Encoding", transferEncoding)
	}
	h.Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": encodedName}))

	// Forward some original header lines.
	forward := []string{"Content-Id", "Content-Description", "Content-Location"}
	for _, k := range forward {
		v := att.Header.Get(k)
		if v != "" {
			h.Set(k, v)
		}
	}

	return h
}

func WriteHeader(w io.Writer, h textproto.MIMEHeader) (err error) {
	if err = http.Header(h).Write(w); err != nil {
		return
	}
	_, err = io.WriteString(w, "\r\n")
	return
}
