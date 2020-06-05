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
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-textwrapper"
	openpgperrors "golang.org/x/crypto/openpgp/errors"
)

func WriteBody(w io.Writer, kr *crypto.KeyRing, m *pmapi.Message) error {
	// Decrypt body.
	if err := m.Decrypt(kr); err != nil && err != openpgperrors.ErrSignatureExpired {
		return err
	}
	if m.MIMEType != pmapi.ContentTypeMultipartMixed {
		// Encode it.
		qp := quotedprintable.NewWriter(w)
		if _, err := io.WriteString(qp, m.Body); err != nil {
			return err
		}
		return qp.Close()
	}
	_, err := io.WriteString(w, m.Body)
	return err
}

func WriteAttachmentBody(w io.Writer, kr *crypto.KeyRing, m *pmapi.Message, att *pmapi.Attachment, r io.Reader) (err error) {
	// Decrypt it
	var dr io.Reader
	dr, err = att.Decrypt(r, kr)
	if err == openpgperrors.ErrKeyIncorrect {
		// Do not fail if attachment is encrypted with a different key.
		dr = r
		err = nil
		att.Name += ".gpg"
		att.MIMEType = "application/pgp-encrypted"
	} else if err != nil && err != openpgperrors.ErrSignatureExpired {
		err = fmt.Errorf("cannot decrypt attachment: %v", err)
		return
	}

	// Encode it.
	ww := textwrapper.NewRFC822(w)
	bw := base64.NewEncoder(base64.StdEncoding, ww)

	var n int64
	if n, err = io.Copy(bw, dr); err != nil {
		err = fmt.Errorf("cannot write attachment: %v (wrote %v bytes)", err, n)
	}

	_ = bw.Close()
	return
}
