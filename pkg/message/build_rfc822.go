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
	"mime"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"
)

func buildRFC822(kr *crypto.KeyRing, msg *pmapi.Message, attData map[string][]byte, opts JobOptions) ([]byte, error) {
	switch {
	case len(msg.Attachments) > 0:
		return buildMultipartRFC822(kr, msg, attData, opts)

	case msg.MIMEType == "multipart/mixed":
		return buildPGPRFC822(kr, msg, opts)

	default:
		return buildSimpleRFC822(kr, msg, opts)
	}
}

func buildSimpleRFC822(kr *crypto.KeyRing, msg *pmapi.Message, opts JobOptions) ([]byte, error) {
	dec, err := msg.Decrypt(kr)
	if err != nil {
		if !opts.IgnoreDecryptionErrors {
			return nil, errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return buildMultipartRFC822(kr, msg, nil, opts)
	}

	hdr := getTextPartHeader(getMessageHeader(msg, opts), dec, msg.MIMEType)

	buf := new(bytes.Buffer)

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(dec); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func buildMultipartRFC822(
	kr *crypto.KeyRing,
	msg *pmapi.Message,
	attData map[string][]byte,
	opts JobOptions,
) ([]byte, error) {
	boundary := newBoundary(msg.ID)

	hdr := getMessageHeader(msg, opts)

	hdr.SetContentType("multipart/mixed", map[string]string{"boundary": boundary.gen()})

	buf := new(bytes.Buffer)

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return nil, err
	}

	var (
		inlineAtts []*pmapi.Attachment
		inlineData [][]byte
		attachAtts []*pmapi.Attachment
		attachData [][]byte
	)

	for _, att := range msg.Attachments {
		if att.Disposition == pmapi.DispositionInline {
			inlineAtts = append(inlineAtts, att)
			inlineData = append(inlineData, attData[att.ID])
		} else {
			attachAtts = append(attachAtts, att)
			attachData = append(attachData, attData[att.ID])
		}
	}

	if len(inlineAtts) > 0 {
		if err := writeRelatedParts(w, kr, boundary, msg, inlineAtts, inlineData, opts); err != nil {
			return nil, err
		}
	} else if err := writeTextPart(w, kr, msg, opts); err != nil {
		return nil, err
	}

	for i, att := range attachAtts {
		if err := writeAttachmentPart(w, kr, att, attachData[i], opts); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeTextPart(
	w *message.Writer,
	kr *crypto.KeyRing,
	msg *pmapi.Message,
	opts JobOptions,
) error {
	dec, err := msg.Decrypt(kr)
	if err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return writeCustomTextPart(w, msg, err)
	}

	return writePart(w, getTextPartHeader(message.Header{}, dec, msg.MIMEType), dec)
}

func writeAttachmentPart(
	w *message.Writer,
	kr *crypto.KeyRing,
	att *pmapi.Attachment,
	attData []byte,
	opts JobOptions,
) error {
	kps, err := base64.StdEncoding.DecodeString(att.KeyPackets)
	if err != nil {
		return err
	}

	msg := crypto.NewPGPSplitMessage(kps, attData).GetPGPMessage()

	dec, err := kr.Decrypt(msg, nil, crypto.GetUnixTime())
	if err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		log.
			WithField("attID", att.ID).
			WithField("msgID", att.MessageID).
			WithError(err).
			Warn("Attachment decryption failed")

		return writeCustomAttachmentPart(w, att, msg, err)
	}

	return writePart(w, getAttachmentPartHeader(att), dec.GetBinary())
}

func writeRelatedParts(
	w *message.Writer,
	kr *crypto.KeyRing,
	boundary *boundary,
	msg *pmapi.Message,
	atts []*pmapi.Attachment,
	attData [][]byte,
	opts JobOptions,
) error {
	hdr := message.Header{}

	hdr.SetContentType("multipart/related", map[string]string{"boundary": boundary.gen()})

	return createPart(w, hdr, func(rel *message.Writer) error {
		if err := writeTextPart(rel, kr, msg, opts); err != nil {
			return err
		}

		for i, att := range atts {
			if err := writeAttachmentPart(rel, kr, att, attData[i], opts); err != nil {
				return err
			}
		}

		return nil
	})
}

func buildPGPRFC822(kr *crypto.KeyRing, msg *pmapi.Message, opts JobOptions) ([]byte, error) {
	dec, err := msg.Decrypt(kr)
	if err != nil {
		if !opts.IgnoreDecryptionErrors {
			return nil, errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return buildPGPMIMEFallbackRFC822(msg, opts)
	}

	hdr := getMessageHeader(msg, opts)

	sigs, err := msg.ExtractSignatures(kr)
	if err != nil {
		log.WithError(err).WithField("id", msg.ID).Warn("Extract signature failed")
	}

	if len(sigs) > 0 {
		return writeMultipartSignedRFC822(hdr, dec, sigs[0])
	}

	return writeMultipartEncryptedRFC822(hdr, dec)
}

func buildPGPMIMEFallbackRFC822(msg *pmapi.Message, opts JobOptions) ([]byte, error) {
	hdr := getMessageHeader(msg, opts)

	hdr.SetContentType("multipart/encrypted", map[string]string{
		"boundary": newBoundary(msg.ID).gen(),
		"protocol": "application/pgp-encrypted",
	})

	buf := new(bytes.Buffer)

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return nil, err
	}

	var encHdr message.Header

	encHdr.SetContentType("application/pgp-encrypted", nil)
	encHdr.Set("Content-Description", "PGP/MIME version identification")

	if err := writePart(w, encHdr, []byte("Version: 1")); err != nil {
		return nil, err
	}

	var dataHdr message.Header

	dataHdr.SetContentType("application/octet-stream", map[string]string{"name": "encrypted.asc"})
	dataHdr.SetContentDisposition("inline", map[string]string{"filename": "encrypted.asc"})
	dataHdr.Set("Content-Description", "OpenPGP encrypted message")

	if err := writePart(w, dataHdr, []byte(msg.Body)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeMultipartSignedRFC822(header message.Header, body []byte, sig pmapi.Signature) ([]byte, error) { //nolint:funlen
	buf := new(bytes.Buffer)

	boundary := newBoundary("").gen()

	header.SetContentType("multipart/signed", map[string]string{
		"micalg":   sig.Hash,
		"protocol": "application/pgp-signature",
		"boundary": boundary,
	})

	if err := textproto.WriteHeader(buf, header.Header); err != nil {
		return nil, err
	}

	mw := textproto.NewMultipartWriter(buf)

	if err := mw.SetBoundary(boundary); err != nil {
		return nil, err
	}

	bodyHeader, bodyData, err := readHeaderBody(body)
	if err != nil {
		return nil, err
	}

	bodyPart, err := mw.CreatePart(*bodyHeader)
	if err != nil {
		return nil, err
	}

	if _, err := bodyPart.Write(bodyData); err != nil {
		return nil, err
	}

	var sigHeader message.Header

	sigHeader.SetContentType("application/pgp-signature", map[string]string{"name": "OpenPGP_signature.asc"})
	sigHeader.SetContentDisposition("attachment", map[string]string{"filename": "OpenPGP_signature"})
	sigHeader.Set("Content-Description", "OpenPGP digital signature")

	sigPart, err := mw.CreatePart(sigHeader.Header)
	if err != nil {
		return nil, err
	}

	sigData, err := crypto.NewPGPSignature(sig.Data).GetArmored()
	if err != nil {
		return nil, err
	}

	if _, err := sigPart.Write([]byte(sigData)); err != nil {
		return nil, err
	}

	if err := mw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeMultipartEncryptedRFC822(header message.Header, body []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	bodyHeader, bodyData, err := readHeaderBody(body)
	if err != nil {
		return nil, err
	}

	// If parsed header is empty then either it is malformed or it is missing.
	// Anyway message could not be considered multipart/mixed anymore since there will be no boundary.
	if bodyHeader.Len() == 0 {
		header.Del("Content-Type")
	}

	entFields := bodyHeader.Fields()

	for entFields.Next() {
		header.Set(entFields.Key(), entFields.Value())
	}

	if err := textproto.WriteHeader(buf, header.Header); err != nil {
		return nil, err
	}

	if _, err := buf.Write(bodyData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getMessageHeader(msg *pmapi.Message, opts JobOptions) message.Header { //nolint:funlen
	hdr := toMessageHeader(msg.Header)

	// SetText will RFC2047-encode.
	if msg.Subject != "" {
		hdr.SetText("Subject", msg.Subject)
	}

	// mail.Address.String() will RFC2047-encode if necessary.
	if msg.Sender != nil {
		hdr.Set("From", msg.Sender.String())
	}

	if len(msg.ReplyTos) > 0 {
		hdr.Set("Reply-To", toAddressList(msg.ReplyTos))
	}

	if len(msg.ToList) > 0 {
		hdr.Set("To", toAddressList(msg.ToList))
	}

	if len(msg.CCList) > 0 {
		hdr.Set("Cc", toAddressList(msg.CCList))
	}

	if len(msg.BCCList) > 0 {
		hdr.Set("Bcc", toAddressList(msg.BCCList))
	}

	setMessageIDIfNeeded(msg, &hdr)

	// Sanitize the date; it needs to have a valid unix timestamp.
	if opts.SanitizeDate {
		if date, err := rfc5322.ParseDateTime(hdr.Get("Date")); err != nil || date.Before(time.Unix(0, 0)) {
			msgDate := SanitizeMessageDate(msg.Time)
			hdr.Set("Date", msgDate.In(time.UTC).Format(time.RFC1123Z))
			// We clobbered the date so we save it under X-Original-Date.
			hdr.Set("X-Original-Date", date.In(time.UTC).Format(time.RFC1123Z))
		}
	}

	// Set our internal ID if requested.
	// This is important for us to detect whether APPENDed things are actually "move like outlook".
	if opts.AddInternalID {
		hdr.Set("X-Pm-Internal-Id", msg.ID)
	}

	// Set our external ID if requested.
	// This was useful during debugging of applemail recovered messages; doesn't help with any behaviour.
	if opts.AddExternalID {
		hdr.Set("X-Pm-External-Id", "<"+msg.ExternalID+">")
	}

	// Set our server date if requested.
	// Can be useful to see how long it took for a message to arrive.
	if opts.AddMessageDate {
		hdr.Set("X-Pm-Date", time.Unix(msg.Time, 0).In(time.UTC).Format(time.RFC1123Z))
	}

	// Include the message ID in the references (supposedly this somehow improves outlook support...).
	if opts.AddMessageIDReference {
		if references := hdr.Get("References"); !strings.Contains(references, msg.ID) {
			hdr.Set("References", references+" <"+msg.ID+"@"+pmapi.InternalIDDomain+">")
		}
	}

	return hdr
}

// SanitizeMessageDate will return time from msgTime timestamp. If timestamp is
// not after epoch the RFC822 publish day will be used. No message should
// realistically be older than RFC822 itself.
func SanitizeMessageDate(msgTime int64) time.Time {
	if msgTime := time.Unix(msgTime, 0); msgTime.After(time.Unix(0, 0)) {
		return msgTime
	}
	return time.Date(1982, 8, 13, 0, 0, 0, 0, time.UTC)
}

// setMessageIDIfNeeded sets Message-Id from ExternalID or ID if it's not
// already set.
func setMessageIDIfNeeded(msg *pmapi.Message, hdr *message.Header) {
	if hdr.Get("Message-Id") == "" {
		if msg.ExternalID != "" {
			hdr.Set("Message-Id", "<"+msg.ExternalID+">")
		} else {
			hdr.Set("Message-Id", "<"+msg.ID+"@"+pmapi.InternalIDDomain+">")
		}
	}
}

func getTextPartHeader(hdr message.Header, body []byte, mimeType string) message.Header {
	params := make(map[string]string)

	if utf8.Valid(body) {
		params["charset"] = "utf-8"
	}

	hdr.SetContentType(mimeType, params)

	// Use quoted-printable for all text/... parts
	hdr.Set("Content-Transfer-Encoding", "quoted-printable")

	return hdr
}

func getAttachmentPartHeader(att *pmapi.Attachment) message.Header {
	hdr := toMessageHeader(mail.Header(att.Header))

	// All attachments have a content type.
	hdr.SetContentType(att.MIMEType, map[string]string{"name": mime.QEncoding.Encode("utf-8", att.Name)})

	// All attachments have a content disposition.
	hdr.SetContentDisposition(att.Disposition, map[string]string{"filename": mime.QEncoding.Encode("utf-8", att.Name)})

	// Use base64 for all attachments except embedded RFC822 messages.
	if att.MIMEType != rfc822Message {
		hdr.Set("Content-Transfer-Encoding", "base64")
	} else {
		hdr.Del("Content-Transfer-Encoding")
	}

	return hdr
}

func toMessageHeader(hdr mail.Header) message.Header {
	var res message.Header

	for key, val := range hdr {
		for _, val := range val {
			// Using AddRaw instead of Add to save key-value pair as byte buffer within Header.
			// This buffer is used latter on in message writer to construct message and avoid crash
			// when key length is more than 76 characters long.
			res.AddRaw([]byte(key + ": " + val + "\r\n"))
		}
	}

	return res
}

func toAddressList(addrs []*mail.Address) string {
	res := make([]string, len(addrs))

	for i, addr := range addrs {
		res[i] = addr.String()
	}

	return strings.Join(res, ", ")
}

func createPart(w *message.Writer, hdr message.Header, fn func(*message.Writer) error) error {
	part, err := w.CreatePart(hdr)
	if err != nil {
		return err
	}

	if err := fn(part); err != nil {
		return err
	}

	return part.Close()
}

func writePart(w *message.Writer, hdr message.Header, body []byte) error {
	return createPart(w, hdr, func(part *message.Writer) error {
		if _, err := part.Write(body); err != nil {
			return errors.Wrap(err, "failed to write part body")
		}

		return nil
	})
}
