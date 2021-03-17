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
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-textwrapper"
)

func BuildEncrypted(m *pmapi.Message, readers []io.Reader, kr *crypto.KeyRing) ([]byte, error) { //nolint[funlen]
	b := &bytes.Buffer{}

	// Overwrite content for main header for import.
	// Even if message has just simple body we should upload as multipart/mixed.
	// Each part has encrypted body and header reflects the original header.
	mainHeader := GetHeader(m)
	mainHeader.Set("Content-Type", "multipart/mixed; boundary="+GetBoundary(m))
	mainHeader.Del("Content-Disposition")
	mainHeader.Del("Content-Transfer-Encoding")
	if err := WriteHeader(b, mainHeader); err != nil {
		return nil, err
	}
	mw := multipart.NewWriter(b)
	if err := mw.SetBoundary(GetBoundary(m)); err != nil {
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
		h := GetAttachmentHeader(att, false)
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

func WriteHeader(w io.Writer, h textproto.MIMEHeader) (err error) {
	if err = http.Header(h).Write(w); err != nil {
		return
	}
	_, err = io.WriteString(w, "\r\n")
	return
}
