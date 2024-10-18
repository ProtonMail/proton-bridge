// Copyright (c) 2024 Proton AG
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
	"bufio"
	"bytes"
	"encoding/base64"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/htmlindex"
)

func newTestMessage(
	t *testing.T,
	kr *crypto.KeyRing,
	messageID, addressID, mimeType, body string, //nolint:unparam
	date time.Time,
) proton.Message {
	return newTestMessageWithHeaders(t, kr, messageID, addressID, mimeType, body, date, nil)
}

func newRawTestMessage(messageID, addressID, mimeType, body string, date time.Time) proton.Message { // nolint:unparam
	return newRawTestMessageWithHeaders(messageID, addressID, mimeType, body, date, nil)
}

func newTestMessageWithHeaders(
	t *testing.T,
	kr *crypto.KeyRing,
	messageID, addressID, mimeType, body string, //nolint:unparam
	date time.Time,
	headers map[string][]string,
) proton.Message {
	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), kr)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	return newRawTestMessageWithHeaders(messageID, addressID, mimeType, arm, date, headers)
}

func newRawTestMessageWithHeaders(messageID, addressID, mimeType, body string, date time.Time, headers map[string][]string) proton.Message {
	msgHeaders := proton.Headers{
		Values: map[string][]string{
			"Content-Type": {mimeType},
			"Date":         {date.In(time.UTC).Format(time.RFC1123Z)},
		},
		Order: []string{"Content-Type", "Date"},
	}

	for k, v := range headers {
		_, ok := msgHeaders.Values[k]
		if !ok {
			msgHeaders.Order = append(msgHeaders.Order, k)
		}

		msgHeaders.Values[k] = v
	}

	return proton.Message{
		MessageMetadata: proton.MessageMetadata{
			ID:        messageID,
			AddressID: addressID,
			Time:      date.Unix(),
		},
		ParsedHeaders: msgHeaders,
		MIMEType:      rfc822.MIMEType(mimeType),
		Body:          body,
	}
}

func newTestMessageFromRFC822(t *testing.T, literal []byte) proton.Message {
	// Note attachment are not supported.
	p := rfc822.Parse(literal)
	h, err := p.ParseHeader()
	require.NoError(t, err)
	var parsedHeaders proton.Headers
	parsedHeaders.Values = make(map[string][]string)
	h.Entries(func(key, val string) {
		parsedHeaders.Values[key] = []string{val}
		parsedHeaders.Order = append(parsedHeaders.Order, key)
	})
	var mailHeaders = mail.Header(parsedHeaders.Values)
	require.True(t, h.Has("Content-Type"))
	mime, _, err := rfc822.ParseMIMEType(h.Get("Content-Type"))
	require.NoError(t, err)
	date, err := mailHeaders.Date()
	require.NoError(t, err)
	sender, err := mail.ParseAddress(parsedHeaders.Values["From"][0])
	require.NoError(t, err)

	return proton.Message{
		MessageMetadata: proton.MessageMetadata{
			ID:             "messageID",
			AddressID:      "addressID",
			LabelIDs:       []string{},
			ExternalID:     "",
			Subject:        parsedHeaders.Values["Subject"][0],
			Sender:         sender,
			ToList:         parseAddressList(t, mailHeaders, "To"),
			CCList:         parseAddressList(t, mailHeaders, "Cc"),
			BCCList:        parseAddressList(t, mailHeaders, "Bcc"),
			ReplyTos:       parseAddressList(t, mailHeaders, "Reply-To"),
			Flags:          0,
			Time:           date.Unix(),
			Size:           0,
			Unread:         false,
			IsReplied:      false,
			IsRepliedAll:   false,
			IsForwarded:    false,
			NumAttachments: 0,
		},
		Header:        string(h.Raw()),
		ParsedHeaders: parsedHeaders,
		Body:          string(p.Body()),
		MIMEType:      mime,
		Attachments:   nil,
	}
}

func parseAddressList(t *testing.T, header mail.Header, key string) []*mail.Address {
	var result []*mail.Address
	if len(header.Get(key)) == 0 {
		return nil
	}

	result, err := header.AddressList(key)
	require.NoError(t, err)

	return result
}

func addTestAttachment(
	t *testing.T,
	kr *crypto.KeyRing,
	msg *proton.Message,
	attachmentID, name, mimeType, disposition, data string,
) []byte {
	enc, err := kr.EncryptAttachment(crypto.NewPlainMessageFromString(data), attachmentID+".bin")
	require.NoError(t, err)

	msg.Attachments = append(msg.Attachments, proton.Attachment{
		ID:       attachmentID,
		Name:     name,
		MIMEType: rfc822.MIMEType(mimeType),
		Headers: proton.Headers{
			Values: map[string][]string{
				"Content-Type":              {mimeType},
				"Content-Disposition":       {disposition},
				"Content-Transfer-Encoding": {"base64"},
			},
			Order: []string{"Content-Type", "Content-Disposition", "Content-Transfer-Encoding"},
		},
		Disposition: proton.Disposition(disposition),
		KeyPackets:  base64.StdEncoding.EncodeToString(enc.GetBinaryKeyPacket()),
	})

	return enc.GetBinaryDataPacket()
}

type testSection struct {
	t    *testing.T
	part *parser.Part
	raw  []byte
}

// NOTE: Each section is parsed individually --> cleaner test code but slower... improve this one day?
func section(t *testing.T, b []byte, section ...int) *testSection {
	p, err := parser.New(bytes.NewReader(b))
	require.NoError(t, err)

	part, err := p.Section(section)
	require.NoError(t, err)

	s, err := rfc822.Parse(b).Part(section...)
	require.NoError(t, err)

	return &testSection{
		t:    t,
		part: part,
		raw:  s.Literal(),
	}
}

func (s *testSection) expectBody(wantBody matcher) *testSection {
	wantBody.match(s.t, string(s.part.Body))

	return s
}

func (s *testSection) expectSection(wantSection matcher) *testSection { //nolint:unparam
	wantSection.match(s.t, string(s.raw))

	return s
}

func (s *testSection) expectContentType(wantContentType matcher) *testSection {
	mimeType, _, err := s.part.Header.ContentType()
	require.NoError(s.t, err)

	wantContentType.match(s.t, mimeType)

	return s
}

func (s *testSection) expectContentTypeParam(key string, wantParam matcher) *testSection { //nolint:unparam
	_, params, err := s.part.Header.ContentType()
	require.NoError(s.t, err)

	wantParam.match(s.t, params[key])

	return s
}

func (s *testSection) expectContentDisposition(wantDisposition matcher) *testSection {
	disposition, _, err := s.part.Header.ContentDisposition()
	require.NoError(s.t, err)

	wantDisposition.match(s.t, disposition)

	return s
}

func (s *testSection) expectContentDispositionParam(key string, wantParam matcher) *testSection { //nolint:unparam
	_, params, err := s.part.Header.ContentDisposition()
	require.NoError(s.t, err)

	wantParam.match(s.t, params[key])

	return s
}

func (s *testSection) expectTransferEncoding(wantTransferEncoding matcher) *testSection {
	wantTransferEncoding.match(s.t, s.part.Header.Get("Content-Transfer-Encoding"))

	return s
}

func (s *testSection) expectDate(wantDate matcher) *testSection {
	wantDate.match(s.t, s.part.Header.Get("Date"))

	return s
}

func (s *testSection) expectHeader(key string, wantValue matcher) *testSection {
	wantValue.match(s.t, s.part.Header.Get(key))

	return s
}

func (s *testSection) expectDecodedHeader(key string, wantValue matcher) *testSection { //nolint:unparam
	dec, err := s.part.Header.Text(key)
	require.NoError(s.t, err)

	wantValue.match(s.t, dec)

	return s
}

func (s *testSection) pubKey() *crypto.KeyRing {
	key, err := crypto.NewKeyFromArmored(string(s.part.Body))
	require.NoError(s.t, err)

	kr, err := crypto.NewKeyRing(key)
	require.NoError(s.t, err)

	return kr
}

func (s *testSection) signature() *crypto.PGPSignature {
	sig, err := crypto.NewPGPSignatureFromArmored(string(s.part.Body))
	require.NoError(s.t, err)

	return sig
}

type matcher interface {
	match(*testing.T, string)
}

type isMatcher struct {
	want string
}

func (matcher isMatcher) match(t *testing.T, have string) {
	require.Equal(t, matcher.want, have)
}

func is(want string) isMatcher {
	return isMatcher{want: want}
}

func isMissing() isMatcher {
	return isMatcher{}
}

type isNotMatcher struct {
	notWant string
}

func (matcher isNotMatcher) match(t *testing.T, have string) {
	require.NotEqual(t, matcher.notWant, have)
}

func isNot(notWant string) isNotMatcher {
	return isNotMatcher{notWant: notWant}
}

type containsMatcher struct {
	contains string
}

func (matcher containsMatcher) match(t *testing.T, have string) {
	require.Contains(t, have, matcher.contains)
}

func contains(contains string) containsMatcher {
	return containsMatcher{contains: contains}
}

type decryptsToMatcher struct {
	kr   *crypto.KeyRing
	want string
}

func (matcher decryptsToMatcher) match(t *testing.T, have string) {
	haveMsg, err := crypto.NewPGPMessageFromArmored(have)
	require.NoError(t, err)

	dec, err := matcher.kr.Decrypt(haveMsg, nil, crypto.GetUnixTime())
	require.NoError(t, err)

	require.Equal(t, matcher.want, string(dec.GetBinary()))
}

func decryptsTo(kr *crypto.KeyRing, want string) decryptsToMatcher {
	return decryptsToMatcher{kr: kr, want: want}
}

type decodesToMatcher struct {
	charset string
	want    string
}

func (matcher decodesToMatcher) match(t *testing.T, have string) {
	enc, err := htmlindex.Get(matcher.charset)
	require.NoError(t, err)

	dec, err := enc.NewDecoder().String(have)
	require.NoError(t, err)

	require.Equal(t, matcher.want, dec)
}

func decodesTo(charset string, want string) decodesToMatcher {
	return decodesToMatcher{charset: charset, want: want}
}

type verifiesAgainstMatcher struct {
	kr  *crypto.KeyRing
	sig *crypto.PGPSignature
}

func (matcher verifiesAgainstMatcher) match(t *testing.T, have string) {
	require.NoError(t, matcher.kr.VerifyDetached(
		crypto.NewPlainMessage([]byte(have)),
		matcher.sig,
		crypto.GetUnixTime()),
	)
}

func verifiesAgainst(kr *crypto.KeyRing, sig *crypto.PGPSignature) verifiesAgainstMatcher {
	return verifiesAgainstMatcher{kr: kr, sig: sig}
}

type maxLineLengthMatcher struct {
	wantMax int
}

func (matcher maxLineLengthMatcher) match(t *testing.T, have string) {
	scanner := bufio.NewScanner(strings.NewReader(have))

	for scanner.Scan() {
		require.Less(t, len(scanner.Text()), matcher.wantMax)
	}
}

func hasMaxLineLength(wantMax int) maxLineLengthMatcher {
	return maxLineLengthMatcher{wantMax: wantMax}
}
