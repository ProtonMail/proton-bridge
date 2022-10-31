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
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/mocks"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/htmlindex"
)

func newTestFetcher(
	m *gomock.Controller,
	kr *crypto.KeyRing,
	msg *pmapi.Message,
	attData ...[]byte,
) Fetcher {
	f := mocks.NewMockFetcher(m)

	f.EXPECT().GetMessage(gomock.Any(), msg.ID).Return(msg, nil)

	for i, att := range msg.Attachments {
		f.EXPECT().GetAttachment(gomock.Any(), att.ID).Return(newTestReadCloser(attData[i]), nil)
	}

	f.EXPECT().KeyRingForAddressID(msg.AddressID).Return(kr, nil)

	return f
}

func newTestMessage(
	t *testing.T,
	kr *crypto.KeyRing,
	messageID, addressID, mimeType, body string, //nolint:unparam
	date time.Time,
) *pmapi.Message {
	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), kr)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	return newRawTestMessage(messageID, addressID, mimeType, arm, date)
}

func newRawTestMessage(messageID, addressID, mimeType, body string, date time.Time) *pmapi.Message {
	return &pmapi.Message{
		ID:        messageID,
		AddressID: addressID,
		MIMEType:  mimeType,
		Header: map[string][]string{
			"Content-Type": {mimeType},
			"Date":         {date.In(time.UTC).Format(time.RFC1123Z)},
		},
		Body: body,
		Time: date.Unix(),
	}
}

func addTestAttachment(
	t *testing.T,
	kr *crypto.KeyRing,
	msg *pmapi.Message,
	attachmentID, name, mimeType, disposition, data string,
) []byte {
	enc, err := kr.EncryptAttachment(crypto.NewPlainMessageFromString(data), attachmentID+".bin")
	require.NoError(t, err)

	msg.Attachments = append(msg.Attachments, &pmapi.Attachment{
		ID:       attachmentID,
		Name:     name,
		MIMEType: mimeType,
		Header: map[string][]string{
			"Content-Type":              {mimeType},
			"Content-Disposition":       {disposition},
			"Content-Transfer-Encoding": {"base64"},
		},
		Disposition: disposition,
		KeyPackets:  base64.StdEncoding.EncodeToString(enc.GetBinaryKeyPacket()),
	})

	return enc.GetBinaryDataPacket()
}

type testReadCloser struct {
	io.Reader
}

func newTestReadCloser(b []byte) *testReadCloser {
	return &testReadCloser{Reader: bytes.NewReader(b)}
}

func (testReadCloser) Close() error {
	return nil
}

type testSection struct {
	t    *testing.T
	part *parser.Part
	raw  []byte
}

// NOTE: Each section is parsed individually --> cleaner test code but slower... improve this one day?
func section(t *testing.T, b []byte, section ...int) *testSection {
	p, err := parser.New(bytes.NewReader(b))
	assert.NoError(t, err)

	part, err := p.Section(section)
	require.NoError(t, err)

	bs, err := NewBodyStructure(bytes.NewReader(b))
	require.NoError(t, err)

	raw, err := bs.GetSection(bytes.NewReader(b), section)
	require.NoError(t, err)

	return &testSection{
		t:    t,
		part: part,
		raw:  raw,
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
	assert.Equal(t, matcher.want, have)
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
	assert.NotEqual(t, matcher.notWant, have)
}

func isNot(notWant string) isNotMatcher {
	return isNotMatcher{notWant: notWant}
}

type containsMatcher struct {
	contains string
}

func (matcher containsMatcher) match(t *testing.T, have string) {
	assert.Contains(t, have, matcher.contains)
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

	assert.Equal(t, matcher.want, string(dec.GetBinary()))
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

	assert.Equal(t, matcher.want, dec)
}

func decodesTo(charset string, want string) decodesToMatcher {
	return decodesToMatcher{charset: charset, want: want}
}

type verifiesAgainstMatcher struct {
	kr  *crypto.KeyRing
	sig *crypto.PGPSignature
}

func (matcher verifiesAgainstMatcher) match(t *testing.T, have string) {
	assert.NoError(t, matcher.kr.VerifyDetached(
		crypto.NewPlainMessage(bytes.TrimSuffix([]byte(have), []byte("\r\n"))),
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
		assert.Less(t, len(scanner.Text()), matcher.wantMax)
	}
}

func hasMaxLineLength(wantMax int) maxLineLengthMatcher {
	return maxLineLengthMatcher{wantMax: wantMax}
}
