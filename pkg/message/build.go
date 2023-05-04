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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package message

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ProtonMail/gluon/rfc5322"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrDecryptionFailed = errors.New("message could not be decrypted")
	ErrNoSuchKeyRing    = errors.New("the keyring to decrypt this message could not be found")
)

// InternalIDDomain is used as a placeholder for reference/message ID headers to improve compatibility with various clients.
const InternalIDDomain = `protonmail.internalid`

func BuildRFC822(kr *crypto.KeyRing, msg proton.Message, attData [][]byte, opts JobOptions) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := BuildRFC822Into(kr, msg, attData, opts, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func BuildRFC822Into(kr *crypto.KeyRing, msg proton.Message, attData [][]byte, opts JobOptions, buf *bytes.Buffer) error {
	switch {
	case len(msg.Attachments) > 0:
		return buildMultipartRFC822(kr, msg, attData, opts, buf)

	case msg.MIMEType == "multipart/mixed":
		return buildPGPRFC822(kr, msg, opts, buf)

	default:
		return buildSimpleRFC822(kr, msg, opts, buf)
	}
}

func buildSimpleRFC822(kr *crypto.KeyRing, msg proton.Message, opts JobOptions, buf *bytes.Buffer) error {
	var decrypted bytes.Buffer
	decrypted.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &decrypted); err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return buildMultipartRFC822(kr, msg, nil, opts, buf)
	}

	hdr := getTextPartHeader(getMessageHeader(msg, opts), decrypted.Bytes(), msg.MIMEType)

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return err
	}

	if _, err := w.Write(decrypted.Bytes()); err != nil {
		return err
	}

	return w.Close()
}

func buildMultipartRFC822(
	kr *crypto.KeyRing,
	msg proton.Message,
	attData [][]byte,
	opts JobOptions,
	buf *bytes.Buffer,
) error {
	boundary := newBoundary(msg.ID)

	hdr := getMessageHeader(msg, opts)

	hdr.SetContentType("multipart/mixed", map[string]string{"boundary": boundary.gen()})

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return err
	}

	var (
		inlineAtts []proton.Attachment
		inlineData [][]byte
		attachAtts []proton.Attachment
		attachData [][]byte
	)

	for index, att := range msg.Attachments {
		if att.Disposition == proton.InlineDisposition {
			inlineAtts = append(inlineAtts, att)
			inlineData = append(inlineData, attData[index])
		} else {
			attachAtts = append(attachAtts, att)
			attachData = append(attachData, attData[index])
		}
	}

	if len(inlineAtts) > 0 {
		if err := writeRelatedParts(w, kr, boundary, msg, inlineAtts, inlineData, opts); err != nil {
			return err
		}
	} else if err := writeTextPart(w, kr, msg, opts); err != nil {
		return err
	}

	for i, att := range attachAtts {
		if err := writeAttachmentPart(w, kr, att, attachData[i], opts); err != nil {
			return err
		}
	}

	return w.Close()
}

func writeTextPart(
	w *message.Writer,
	kr *crypto.KeyRing,
	msg proton.Message,
	opts JobOptions,
) error {
	var decrypted bytes.Buffer
	decrypted.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &decrypted); err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return writeCustomTextPart(w, msg, err)
	}

	return writePart(w, getTextPartHeader(message.Header{}, decrypted.Bytes(), msg.MIMEType), decrypted.Bytes())
}

func writeAttachmentPart(
	w *message.Writer,
	kr *crypto.KeyRing,
	att proton.Attachment,
	attData []byte,
	opts JobOptions,
) error {
	kps, err := base64.StdEncoding.DecodeString(att.KeyPackets)
	if err != nil {
		return err
	}

	// Use io.Multi
	attachmentReader := io.MultiReader(bytes.NewReader(kps), bytes.NewReader(attData))

	stream, err := kr.DecryptStream(attachmentReader, nil, crypto.GetUnixTime())
	if err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		log.
			WithField("attID", att.ID).
			WithError(err).
			Warn("Attachment decryption failed - construct")

		var pgpMessageBuffer bytes.Buffer
		pgpMessageBuffer.Grow(len(kps) + len(attData))
		pgpMessageBuffer.Write(kps)
		pgpMessageBuffer.Write(attData)

		return writeCustomAttachmentPart(w, att, &crypto.PGPMessage{Data: pgpMessageBuffer.Bytes()}, err)
	}

	var decryptBuffer bytes.Buffer
	decryptBuffer.Grow(len(kps) + len(attData))

	if _, err := decryptBuffer.ReadFrom(stream); err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		log.
			WithField("attID", att.ID).
			WithError(err).
			Warn("Attachment decryption failed - stream")

		var pgpMessageBuffer bytes.Buffer
		pgpMessageBuffer.Grow(len(kps) + len(attData))
		pgpMessageBuffer.Write(kps)
		pgpMessageBuffer.Write(attData)

		return writeCustomAttachmentPart(w, att, &crypto.PGPMessage{Data: pgpMessageBuffer.Bytes()}, err)
	}

	return writePart(w, getAttachmentPartHeader(att), decryptBuffer.Bytes())
}

func writeRelatedParts(
	w *message.Writer,
	kr *crypto.KeyRing,
	boundary *boundary,
	msg proton.Message,
	atts []proton.Attachment,
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

func buildPGPRFC822(kr *crypto.KeyRing, msg proton.Message, opts JobOptions, buf *bytes.Buffer) error {
	var decrypted bytes.Buffer
	decrypted.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &decrypted); err != nil {
		if !opts.IgnoreDecryptionErrors {
			return errors.Wrap(ErrDecryptionFailed, err.Error())
		}

		return buildPGPMIMEFallbackRFC822(msg, opts, buf)
	}

	hdr := getMessageHeader(msg, opts)

	sigs, err := proton.ExtractSignatures(kr, msg.Body)
	if err != nil {
		log.WithError(err).WithField("id", msg.ID).Warn("Extract signature failed")
	}

	if len(sigs) > 0 {
		return writeMultipartSignedRFC822(hdr, decrypted.Bytes(), sigs[0], buf)
	}

	return writeMultipartEncryptedRFC822(hdr, decrypted.Bytes(), buf)
}

func buildPGPMIMEFallbackRFC822(msg proton.Message, opts JobOptions, buf *bytes.Buffer) error {
	hdr := getMessageHeader(msg, opts)

	hdr.SetContentType("multipart/encrypted", map[string]string{
		"boundary": newBoundary(msg.ID).gen(),
		"protocol": "application/pgp-encrypted",
	})

	w, err := message.CreateWriter(buf, hdr)
	if err != nil {
		return err
	}

	var encHdr message.Header

	encHdr.SetContentType("application/pgp-encrypted", nil)
	encHdr.Set("Content-Description", "PGP/MIME version identification")

	if err := writePart(w, encHdr, []byte("Version: 1")); err != nil {
		return err
	}

	var dataHdr message.Header

	dataHdr.SetContentType("application/octet-stream", map[string]string{"name": "encrypted.asc"})
	dataHdr.SetContentDisposition("inline", map[string]string{"filename": "encrypted.asc"})
	dataHdr.Set("Content-Description", "OpenPGP encrypted message")

	if err := writePart(w, dataHdr, []byte(msg.Body)); err != nil {
		return err
	}

	return w.Close()
}

func writeMultipartSignedRFC822(header message.Header, body []byte, sig proton.Signature, buf *bytes.Buffer) error {
	boundary := newBoundary("").gen()

	header.SetContentType("multipart/signed", map[string]string{
		"micalg":   sig.Hash,
		"protocol": "application/pgp-signature",
		"boundary": boundary,
	})

	if err := textproto.WriteHeader(buf, header.Header); err != nil {
		return err
	}

	mw := textproto.NewMultipartWriter(buf)

	if err := mw.SetBoundary(boundary); err != nil {
		return err
	}

	bodyHeader, bodyData, err := readHeaderBody(body)
	if err != nil {
		return err
	}

	bodyPart, err := mw.CreatePart(*bodyHeader)
	if err != nil {
		return err
	}

	if _, err := bodyPart.Write(bodyData); err != nil {
		return err
	}

	var sigHeader message.Header

	sigHeader.SetContentType("application/pgp-signature", map[string]string{"name": "OpenPGP_signature.asc"})
	sigHeader.SetContentDisposition("attachment", map[string]string{"filename": "OpenPGP_signature"})
	sigHeader.Set("Content-Description", "OpenPGP digital signature")

	sigPart, err := mw.CreatePart(sigHeader.Header)
	if err != nil {
		return err
	}

	sigData, err := sig.Data.GetArmored()
	if err != nil {
		return err
	}

	if _, err := sigPart.Write([]byte(sigData)); err != nil {
		return err
	}

	return mw.Close()
}

func writeMultipartEncryptedRFC822(header message.Header, body []byte, buf *bytes.Buffer) error {
	bodyHeader, bodyData, err := readHeaderBody(body)
	if err != nil {
		return err
	}

	// Remove old content type header as it is non-standard. Ensure that messages
	// without content type header entries don't become invalid.
	header.Del("Content-Type")

	entFields := bodyHeader.Fields()

	for entFields.Next() {
		// Only set the header field if it is present. Header sanitation will be overridden otherwise.
		if !header.Has(entFields.Key()) {
			header.Set(entFields.Key(), entFields.Value())
		}
	}

	if err := textproto.WriteHeader(buf, header.Header); err != nil {
		return err
	}

	if _, err := buf.Write(bodyData); err != nil {
		return err
	}

	return nil
}

func addressEmpty(address *mail.Address) bool {
	if address == nil {
		return true
	}

	if address.Name == "" && address.Address == "" {
		return true
	}

	return false
}

func getMessageHeader(msg proton.Message, opts JobOptions) message.Header {
	hdr := toMessageHeader(msg.ParsedHeaders)

	// SetText will RFC2047-encode.
	if msg.Subject != "" {
		hdr.SetText("Subject", msg.Subject)
	}

	// mail.Address.String() will RFC2047-encode if necessary.
	if !addressEmpty(msg.Sender) {
		hdr.Set("From", msg.Sender.String())
	}

	if len(msg.ReplyTos) > 0 && !msg.IsDraft() {
		if !(len(msg.ReplyTos) == 1 && addressEmpty(msg.ReplyTos[0])) {
			hdr.Set("Reply-To", toAddressList(msg.ReplyTos))
		}
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
			// We clobbered the date so we save it under X-Original-Date only if no such value exists.
			if !hdr.Has("X-Original-Date") {
				hdr.Set("X-Original-Date", date.In(time.UTC).Format(time.RFC1123Z))
			}
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
		if msg.ExternalID != "" {
			hdr.Set("X-Pm-External-Id", "<"+msg.ExternalID+">")
		}
	}

	// Set our server date if requested.
	// Can be useful to see how long it took for a message to arrive.
	if opts.AddMessageDate {
		hdr.Set("X-Pm-Date", time.Unix(msg.Time, 0).In(time.UTC).Format(time.RFC1123Z))
	}

	// Include the message ID in the references (supposedly this somehow improves outlook support...).
	if opts.AddMessageIDReference {
		if refs := hdr.Values("References"); xslices.IndexFunc(refs, func(ref string) bool {
			return strings.Contains(ref, msg.ID)
		}) < 0 {
			hdr.Set("References", strings.Join(append(refs, "<"+msg.ID+"@"+InternalIDDomain+">"), " "))
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
func setMessageIDIfNeeded(msg proton.Message, hdr *message.Header) {
	if hdr.Get("Message-Id") == "" {
		if msg.ExternalID != "" {
			hdr.Set("Message-Id", "<"+msg.ExternalID+">")
		} else {
			hdr.Set("Message-Id", "<"+msg.ID+"@"+InternalIDDomain+">")
		}
	}
}

func getTextPartHeader(hdr message.Header, body []byte, mimeType rfc822.MIMEType) message.Header {
	params := make(map[string]string)

	if utf8.Valid(body) {
		params["charset"] = "utf-8"
	}

	hdr.SetContentType(string(mimeType), params)

	// Use quoted-printable for all text/... parts
	hdr.Set("Content-Transfer-Encoding", "quoted-printable")

	return hdr
}

func getAttachmentPartHeader(att proton.Attachment) message.Header {
	hdr := toMessageHeader(att.Headers)

	// All attachments have a content type.
	mimeType, params, err := mime.ParseMediaType(string(att.MIMEType))
	if err != nil {
		logrus.WithError(err).Errorf("Failed to parse mime type: '%v'", att.MIMEType)
		hdr.Set("Content-Type", string(att.MIMEType))
	} else {
		// Merge the overridden name into the params
		encodedName := mime.QEncoding.Encode("utf-8", att.Name)
		params["name"] = encodedName
		params["filename"] = encodedName
		hdr.SetContentType(mimeType, params)
	}

	// All attachments have a content disposition.
	hdr.SetContentDisposition(string(att.Disposition), map[string]string{"filename": mime.QEncoding.Encode("utf-8", att.Name)})

	// Use base64 for all attachments except embedded RFC822 messages.
	if att.MIMEType != rfc822.MessageRFC822 {
		hdr.Set("Content-Transfer-Encoding", "base64")
	} else {
		hdr.Del("Content-Transfer-Encoding")
	}

	return hdr
}

func toMessageHeader(hdr proton.Headers) message.Header {
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

type boundary struct {
	val string
}

func newBoundary(seed string) *boundary {
	return &boundary{val: seed}
}

func (bw *boundary) gen() string {
	bw.val = algo.HashHexSHA256(bw.val)
	return bw.val
}
