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
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"

	"github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-textwrapper"
	openpgperrors "golang.org/x/crypto/openpgp/errors"
)

// Builder for converting PM message to RFC822. Builder will directly write
// changes to message when fetching or building message.
type Builder struct {
	cl  pmapi.Client
	msg *pmapi.Message

	EncryptedToHTML bool
	succDcrpt       bool
}

// NewBuilder initiated with client and message meta info.
func NewBuilder(client pmapi.Client, message *pmapi.Message) *Builder {
	return &Builder{cl: client, msg: message, EncryptedToHTML: true, succDcrpt: false}
}

// fetchMessage will update original PM message if successful
func (bld *Builder) fetchMessage() (err error) {
	if bld.msg.Body != "" {
		return nil
	}

	complete, err := bld.cl.GetMessage(bld.msg.ID)
	if err != nil {
		return
	}

	*bld.msg = *complete

	return
}

func (bld *Builder) writeMessageBody(w io.Writer) error {
	if err := bld.fetchMessage(); err != nil {
		return err
	}

	err := bld.WriteBody(w)
	if err != nil {
		_, _ = io.WriteString(w, "\r\n")
		if bld.EncryptedToHTML {
			_ = CustomMessage(bld.msg, err, true)
		}
		_, err = io.WriteString(w, bld.msg.Body)
		_, _ = io.WriteString(w, "\r\n")
	}

	return err
}

func (bld *Builder) writeAttachmentBody(w io.Writer, att *pmapi.Attachment) error {
	// Retrieve encrypted attachment
	r, err := bld.cl.GetAttachment(att.ID)
	if err != nil {
		return err
	}
	defer r.Close() //nolint[errcheck]

	if err := bld.WriteAttachmentBody(w, att, r); err != nil {
		// Returning an error here makes e-mail clients like Thunderbird behave
		// badly, trying to retrieve the message again and again
		log.Warnln("Cannot write attachment body:", err)
	}
	return nil
}

func (bld *Builder) writeRelatedPart(p io.Writer, inlines []*pmapi.Attachment) error {
	related := multipart.NewWriter(p)

	_ = related.SetBoundary(GetRelatedBoundary(bld.msg))

	buf := &bytes.Buffer{}
	if err := bld.writeMessageBody(buf); err != nil {
		return err
	}

	// Write the body part
	h := GetBodyHeader(bld.msg)

	var err error
	if p, err = related.CreatePart(h); err != nil {
		return err
	}

	_, _ = buf.WriteTo(p)

	for _, inline := range inlines {
		buf = &bytes.Buffer{}
		if err = bld.writeAttachmentBody(buf, inline); err != nil {
			return err
		}

		h := GetAttachmentHeader(inline)
		if p, err = related.CreatePart(h); err != nil {
			return err
		}
		_, _ = buf.WriteTo(p)
	}

	_ = related.Close()
	return nil
}

// BuildMessage converts PM message to body structure (not RFC3501) and bytes
// of RC822 message. If successful the original PM message will contain decrypted body.
func (bld *Builder) BuildMessage() (structure *BodyStructure, message []byte, err error) { //nolint[funlen]
	if err = bld.fetchMessage(); err != nil {
		return nil, nil, err
	}

	bodyBuf := &bytes.Buffer{}

	mainHeader := GetHeader(bld.msg)
	mainHeader.Set("Content-Type", "multipart/mixed; boundary="+GetBoundary(bld.msg))
	if err = WriteHeader(bodyBuf, mainHeader); err != nil {
		return nil, nil, err
	}
	_, _ = io.WriteString(bodyBuf, "\r\n")

	// NOTE: Do we really need extra encapsulation? i.e. Bridge-IMAP message is always multipart/mixed

	if bld.msg.MIMEType == pmapi.ContentTypeMultipartMixed {
		_, _ = io.WriteString(bodyBuf, "\r\n--"+GetBoundary(bld.msg)+"\r\n")
		if err = bld.writeMessageBody(bodyBuf); err != nil {
			return nil, nil, err
		}
		_, _ = io.WriteString(bodyBuf, "\r\n--"+GetBoundary(bld.msg)+"--\r\n")
	} else {
		mw := multipart.NewWriter(bodyBuf)
		_ = mw.SetBoundary(GetBoundary(bld.msg))

		var partWriter io.Writer
		atts, inlines := SeparateInlineAttachments(bld.msg)

		if len(inlines) > 0 {
			relatedHeader := GetRelatedHeader(bld.msg)
			if partWriter, err = mw.CreatePart(relatedHeader); err != nil {
				return nil, nil, err
			}
			_ = bld.writeRelatedPart(partWriter, inlines)
		} else {
			buf := &bytes.Buffer{}
			if err = bld.writeMessageBody(buf); err != nil {
				return nil, nil, err
			}

			// Write the body part
			bodyHeader := GetBodyHeader(bld.msg)
			if partWriter, err = mw.CreatePart(bodyHeader); err != nil {
				return nil, nil, err
			}

			_, _ = buf.WriteTo(partWriter)
		}

		// Write the attachments parts
		for _, att := range atts {
			buf := &bytes.Buffer{}
			if err = bld.writeAttachmentBody(buf, att); err != nil {
				return nil, nil, err
			}

			attachmentHeader := GetAttachmentHeader(att)
			if partWriter, err = mw.CreatePart(attachmentHeader); err != nil {
				return nil, nil, err
			}

			_, _ = buf.WriteTo(partWriter)
		}

		_ = mw.Close()
	}

	// wee need to copy buffer before building body structure
	message = bodyBuf.Bytes()
	structure, err = NewBodyStructure(bodyBuf)
	return structure, message, err
}

// SuccessfullyDecrypted is true when message was fetched and decrypted successfully
func (bld *Builder) SuccessfullyDecrypted() bool { return bld.succDcrpt }

// WriteBody decrypts PM message and writes main body section. The external PGP
// message is written as is (including attachments)
func (bld *Builder) WriteBody(w io.Writer) error {
	kr, err := bld.cl.KeyRingForAddressID(bld.msg.AddressID)
	if err != nil {
		return err
	}
	// decrypt body
	if err := bld.msg.Decrypt(kr); err != nil && err != openpgperrors.ErrSignatureExpired {
		return err
	}
	bld.succDcrpt = true
	if bld.msg.MIMEType != pmapi.ContentTypeMultipartMixed {
		// transfer encoding
		qp := quotedprintable.NewWriter(w)
		if _, err := io.WriteString(qp, bld.msg.Body); err != nil {
			return err
		}
		return qp.Close()
	}
	_, err = io.WriteString(w, bld.msg.Body)
	return err
}

// WriteAttachmentBody decrypts and writes the attachments
func (bld *Builder) WriteAttachmentBody(w io.Writer, att *pmapi.Attachment, attReader io.Reader) (err error) {
	kr, err := bld.cl.KeyRingForAddressID(bld.msg.AddressID)
	if err != nil {
		return err
	}
	// Decrypt it
	var dr io.Reader
	dr, err = att.Decrypt(attReader, kr)
	if err == openpgperrors.ErrKeyIncorrect {
		// Do not fail if attachment is encrypted with a different key
		dr = attReader
		err = nil
		att.Name += ".gpg"
		att.MIMEType = "application/pgp-encrypted"
	} else if err != nil && err != openpgperrors.ErrSignatureExpired {
		err = fmt.Errorf("cannot decrypt attachment: %v", err)
		return err
	}

	// transfer encoding
	ww := textwrapper.NewRFC822(w)
	bw := base64.NewEncoder(base64.StdEncoding, ww)

	var n int64
	if n, err = io.Copy(bw, dr); err != nil {
		err = fmt.Errorf("cannot write attachment: %v (wrote %v bytes)", err, n)
	}

	_ = bw.Close()
	return err
}

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
	bodyHeader.Set("Content-Disposition", "inline")
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
		h := GetAttachmentHeader(att)
		p, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}
		// Create line wrapper writer.
		ww := textwrapper.NewRFC822(p)

		// Create base64 writer.
		bw := base64.NewEncoder(base64.StdEncoding, ww)

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		// Create encrypted writer.
		pgpMessage, err := kr.Encrypt(crypto.NewPlainMessage(data), nil)
		if err != nil {
			return nil, err
		}
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
