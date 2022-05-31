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
	"context"
	"errors"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/mocks"
	tests "github.com/ProtonMail/proton-bridge/v2/test"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPlainMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is(`body`)).
		expectTransferEncoding(is(`quoted-printable`))
}

func TestBuildPlainMessageWithLongKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(1, 1)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())
	msg.Header["ReallyVeryVeryVeryVeryVeryLongLongLongLongLongLongLongKeyThatWillHaveNotSoLongValue"] = []string{"value"}

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is(`body`)).
		expectTransferEncoding(is(`quoted-printable`)).
		expectHeader(`ReallyVeryVeryVeryVeryVeryLongLongLongLongLongLongLongKeyThatWillHaveNotSoLongValue`, is(`value`))
}

func TestBuildHTMLMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/html`)).
		expectBody(is(`<html><body>body</body></html>`)).
		expectTransferEncoding(is(`quoted-printable`))
}

func TestBuildPlainEncryptedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-plaintext.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`plain no pubkey no sign`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`Where do fruits go on vacation? Pear-is!`))
}

func TestBuildPlainEncryptedMessageMissingHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(1, 1)
	defer b.Done()

	body := readerToString(getFileReader("plaintext-missing-header.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is("How do we know that the ocean is friendly? It waves!\r\n"))
}

func TestBuildPlainEncryptedMessageInvalidHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(1, 1)
	defer b.Done()

	body := readerToString(getFileReader("plaintext-invalid-header.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is("MalformedKey Value\r\n\r\nHow do we know that the ocean is friendly? It waves!\r\n"))
}

func TestBuildPlainSignedEncryptedMessageMissingHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(1, 1)
	defer b.Done()

	body := readerToString(getFileReader("plaintext-missing-header.eml"))

	kr := tests.MakeKeyRing(t)
	sig := tests.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)). // NOTE: Maybe this is bad... should probably be pgp-sha256
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(is("How do we know that the ocean is friendly? It waves!\r\n"))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildPlainSignedEncryptedMessageInvalidHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(1, 1)
	defer b.Done()

	body := readerToString(getFileReader("plaintext-invalid-header.eml"))

	kr := tests.MakeKeyRing(t)
	sig := tests.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)). // NOTE: Maybe this is bad... should probably be pgp-sha256
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(is("MalformedKey Value\r\n\r\nHow do we know that the ocean is friendly? It waves!\r\n"))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildPlainEncryptedLatin2Message(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-plaintext-latin2.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectContentTypeParam("charset", is(`iso-8859-2`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectHeader(`Subject`, is(`plain no pubkey no sign`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`)).
		expectBody(decodesTo("iso-8859-2", "řšřšřš\r\n"))
}

func TestBuildHTMLEncryptedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-html.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`html no pubkey no sign`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`))

	section(t, res, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`What do you call a poor Santa Claus`)).
		expectBody(contains(`Where do boats go when they're sick`))
}

func TestBuildPlainSignedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("text_plain.eml"))

	kr := tests.MakeKeyRing(t)
	sig := tests.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)). // NOTE: Maybe this is bad... should probably be pgp-sha256
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(is(`body`)).
		expectSection(verifiesAgainst(sig, section(t, res, 2).signature()))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildPlainSignedBase64Message(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("text_plain_base64.eml"))

	kr := tests.MakeKeyRing(t)
	sig := tests.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)). // NOTE: Maybe this is bad... should probably be pgp-sha256
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`)).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectTransferEncoding(is(`base64`)).
		expectBody(is(`body`)).
		expectSection(verifiesAgainst(sig, section(t, res, 2).signature()))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedPlainEncryptedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-plaintext.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`plain body no pubkey`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`))

	section(t, res, 1, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`Why do seagulls fly over the ocean`)).
		expectBody(contains(`Because if they flew over the bay, we'd call them bagels`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedHTMLEncryptedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-html.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`html body no pubkey`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`))

	section(t, res, 1, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`Behold another <font color="#ee24cc">HTML</font>`)).
		expectBody(contains(`I only know 25 letters of the alphabet`)).
		expectBody(contains(`What did one wall say to the other`)).
		expectBody(contains(`What did the zero say to the eight`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedPlainEncryptedMessageWithPubKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-plaintext-with-pubkey.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`simple plaintext body`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`)).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 2).pubKey(), section(t, res, 2).signature()))

	section(t, res, 1, 1).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1, 1, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`Why don't crabs give to charity? Because they're shellfish.`))

	section(t, res, 1, 1, 2).
		expectContentType(is(`application/pgp-keys`)).
		expectContentTypeParam(`name`, is(`OpenPGP_0x161C0875822359F7.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_0x161C0875822359F7.asc`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedHTMLEncryptedMessageWithPubKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-html-with-pubkey.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`simple html body`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`)).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 2).pubKey(), section(t, res, 2).signature()))

	section(t, res, 1, 1).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1, 1, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`Do I enjoy making courthouse puns`)).
		expectBody(contains(`Can February March`))

	section(t, res, 1, 1, 2).
		expectContentType(is(`application/pgp-keys`)).
		expectContentTypeParam(`name`, is(`OpenPGP_0x161C0875822359F7.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_0x161C0875822359F7.asc`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedMultipartAlternativeEncryptedMessageWithPubKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-multipart-alternative-with-pubkey.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`Alternative`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`)).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 3).pubKey(), section(t, res, 2).signature()))

	section(t, res, 1, 1).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1, 1, 1).
		expectContentType(is(`multipart/alternative`))

	section(t, res, 1, 1, 1, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`This Rich formated text`)).
		expectBody(contains(`What kind of shoes do ninjas wear`)).
		expectBody(contains(`How does a penguin build its house`))

	section(t, res, 1, 1, 1, 2).
		expectContentType(is(`text/html`)).
		expectBody(contains(`This <font color="#ee24cc">Rich</font> formated text`)).
		expectBody(contains(`What kind of shoes do ninjas wear`)).
		expectBody(contains(`How does a penguin build its house`))

	section(t, res, 1, 1, 2).
		expectContentType(is(`application/pdf`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`minimal.pdf`)).
		expectContentDispositionParam(`filename`, is(`minimal.pdf`))

	section(t, res, 1, 1, 3).
		expectContentType(is(`application/pgp-keys`)).
		expectContentTypeParam(`name`, is(`OpenPGP_0x161C0875822359F7.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_0x161C0875822359F7.asc`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildSignedEmbeddedMessageRFC822EncryptedMessageWithPubKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	body := readerToString(getFileReader("pgp-mime-body-signed-embedded-message-rfc822-with-pubkey.eml"))

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`pgp-sha256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`Fwd: HTML with attachment external PGP`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is(`schizofrenic@pm.me`)).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 2).pubKey(), section(t, res, 2).signature()))

	section(t, res, 1, 1).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1, 1, 1).
		expectContentType(is(`text/plain`))

	section(t, res, 1, 1, 2).
		expectContentType(is(`application/pgp-keys`)).
		expectContentTypeParam(`name`, is(`OpenPGP_0x161C0875822359F7.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_0x161C0875822359F7.asc`))

	section(t, res, 1, 1, 3).
		expectContentType(is(`message/rfc822`)).
		expectContentTypeParam(`name`, is(`HTML with attachment external PGP.eml`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`HTML with attachment external PGP.eml`))

	section(t, res, 2).
		expectContentType(is(`application/pgp-signature`)).
		expectContentTypeParam(`name`, is(`OpenPGP_signature.asc`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`OpenPGP_signature`))
}

func TestBuildHTMLMessageWithAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, att), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 2).
		expectBody(is(`attachment`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`file.png`)).
		expectContentDispositionParam(`filename`, is(`file.png`))
}

func TestBuildHTMLMessageWithRFC822Attachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, msg, "attachID", "file.eml", "message/rfc822", "attachment", "... message/rfc822 ...")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, att), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 2).
		expectBody(is(`... message/rfc822 ...`)).
		expectContentType(is(`message/rfc822`)).
		expectTransferEncoding(isNot(`base64`)).
		expectContentTypeParam(`name`, is(`file.eml`)).
		expectContentDispositionParam(`filename`, is(`file.eml`))
}

func TestBuildHTMLMessageWithInlineAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	inl := addTestAttachment(t, kr, msg, "inlineID", "file.png", "image/png", "inline", "inline")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, inl), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res, 1).
		expectContentType(is(`multipart/related`))

	section(t, res, 1, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 1, 2).
		expectBody(is(`inline`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`file.png`)).
		expectContentDispositionParam(`filename`, is(`file.png`))
}

func TestBuildHTMLMessageWithComplexAttachments(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	inl0 := addTestAttachment(t, kr, msg, "inlineID0", "inline0.png", "image/png", "inline", "inline0")
	inl1 := addTestAttachment(t, kr, msg, "inlineID1", "inline1.png", "image/png", "inline", "inline1")
	att0 := addTestAttachment(t, kr, msg, "attachID0", "attach0.png", "image/png", "attachment", "attach0")
	att1 := addTestAttachment(t, kr, msg, "attachID1", "attach1.png", "image/png", "attachment", "attach1")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, inl0, inl1, att0, att1), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res, 1).
		expectContentType(is(`multipart/related`))

	section(t, res, 1, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 1, 2).
		expectBody(is(`inline0`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`inline0.png`)).
		expectContentDispositionParam(`filename`, is(`inline0.png`))

	section(t, res, 1, 3).
		expectBody(is(`inline1`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`inline1.png`)).
		expectContentDispositionParam(`filename`, is(`inline1.png`))

	section(t, res, 2).
		expectBody(is(`attach0`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`attach0.png`)).
		expectContentDispositionParam(`filename`, is(`attach0.png`))

	section(t, res, 3).
		expectBody(is(`attach1`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`attach1.png`)).
		expectContentDispositionParam(`filename`, is(`attach1.png`))
}

func TestBuildAttachmentWithExoticFilename(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, msg, "attachID", `I řeally šhould leařn czech.png`, "image/png", "attachment", "attachment")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, att), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	// The "name" and "filename" params should actually be RFC2047-encoded because they aren't 7-bit clean.
	// We expect them to be readable as UTF-8 but we check that the raw header value contains the encoded data.
	section(t, res, 2).
		expectContentTypeParam(`name`, is(`I řeally šhould leařn czech.png`)).
		expectHeader(`Content-Type`, contains(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech.png?=`)).
		expectContentDispositionParam(`filename`, is(`I řeally šhould leařn czech.png`)).
		expectHeader(`Content-Disposition`, contains(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech.png?=`))
}

func TestBuildAttachmentWithLongFilename(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	veryLongName := strings.Repeat("a", 200) + ".png"

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, msg, "attachID", veryLongName, "image/png", "attachment", "attachment")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, att), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	// NOTE: hasMaxLineLength is too high! Long filenames should be linewrapped using multipart filenames.
	section(t, res, 2).
		expectContentTypeParam(`name`, is(veryLongName)).
		expectHeader(`Content-Type`, contains(veryLongName)).
		expectContentDispositionParam(`filename`, is(veryLongName)).
		expectHeader(`Content-Disposition`, contains(veryLongName)).
		expectSection(hasMaxLineLength(215))
}

func TestBuildMessageDate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))
}

func TestBuildMessageWithInvalidDate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Create a message with "invalid" (according to applemail) date (before unix time 0).
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Unix(-1, 0))

	// Build the message as usual; the date will be before 1970.
	jobRaw, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	resRaw, err := jobRaw.GetResult()
	require.NoError(t, err)
	done()

	section(t, resRaw).
		expectDate(is(`Wed, 31 Dec 1969 23:59:59 +0000`)).
		expectHeader(`X-Original-Date`, isMissing())

	// Build the message with date sanitization enabled; the date will be RFC822's birthdate.
	jobFix, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg),
		msg.ID,
		JobOptions{SanitizeDate: true},
		ForegroundPriority,
	)
	resFix, err := jobFix.GetResult()
	require.NoError(t, err)
	done()

	section(t, resFix).
		expectDate(is(`Fri, 13 Aug 1982 00:00:00 +0000`)).
		expectHeader(`X-Original-Date`, is(`Wed, 31 Dec 1969 23:59:59 +0000`))
}

func TestBuildMessageInternalID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).expectHeader(`Message-Id`, is(`<messageID@protonmail.internalid>`))
}

func TestBuildMessageExternalID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set the message's external ID; this should be used preferentially to set the Message-Id header field.
	msg.ExternalID = "externalID"

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).expectHeader(`Message-Id`, is(`<externalID>`))
}

func TestBuild8BitBody(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Set an 8-bit body; the charset should be set to UTF-8.
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "I řeally šhould leařn czech", time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).expectContentTypeParam(`charset`, is(`utf-8`))
}

func TestBuild8BitSubject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit subject; it should be RFC2047-encoded.
	msg.Subject = `I řeally šhould leařn czech`

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectHeader(`Subject`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech?=`)).
		expectDecodedHeader(`Subject`, is(`I řeally šhould leařn czech`))
}

func TestBuild8BitSender(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit sender; it should be RFC2047-encoded.
	msg.Sender = &mail.Address{
		Name:    `I řeally šhould leařn czech`,
		Address: `mail@example.com`,
	}

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectHeader(`From`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech?= <mail@example.com>`)).
		expectDecodedHeader(`From`, is(`I řeally šhould leařn czech <mail@example.com>`))
}

func TestBuild8BitRecipients(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit sender; it should be RFC2047-encoded.
	msg.ToList = []*mail.Address{
		{Name: `I řeally šhould`, Address: `mail1@example.com`},
		{Name: `leařn czech`, Address: `mail2@example.com`},
	}

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectHeader(`To`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould?= <mail1@example.com>, =?utf-8?q?lea=C5=99n_czech?= <mail2@example.com>`)).
		expectDecodedHeader(`To`, is(`I řeally šhould <mail1@example.com>, leařn czech <mail2@example.com>`))
}

func TestBuildIncludeMessageIDReference(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Add references.
	msg.Header["References"] = []string{"<myreference@domain.com>"}

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	res, err := job.GetResult()
	require.NoError(t, err)
	done()

	section(t, res).expectHeader(`References`, is(`<myreference@domain.com>`))

	jobRef, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg),
		msg.ID,
		JobOptions{AddMessageIDReference: true},
		ForegroundPriority,
	)
	resRef, err := jobRef.GetResult()
	require.NoError(t, err)
	done()

	section(t, resRef).expectHeader(`References`, is(`<myreference@domain.com> <messageID@protonmail.internalid>`))
}

func TestBuildMessageIsDeterministic(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())
	inl := addTestAttachment(t, kr, msg, "inlineID", "file.png", "image/png", "inline", "inline")
	att := addTestAttachment(t, kr, msg, "attachID", "attach.png", "image/png", "attachment", "attachment")

	job1, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, inl, att), msg.ID, ForegroundPriority)
	res1, err := job1.GetResult()
	require.NoError(t, err)
	done()

	job2, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, inl, att), msg.ID, ForegroundPriority)
	res2, err := job2.GetResult()
	require.NoError(t, err)
	done()

	assert.Equal(t, res1, res2)
}

func TestBuildParallel(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg1 := newTestMessage(t, kr, "messageID1", "addressID", "text/plain", "body1", time.Now())
	msg2 := newTestMessage(t, kr, "messageID2", "addressID", "text/plain", "body2", time.Now())

	job1, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg1), msg1.ID, ForegroundPriority)
	defer done()

	job2, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg2), msg2.ID, ForegroundPriority)
	defer done()

	res1, err := job1.GetResult()
	require.NoError(t, err)

	section(t, res1).expectBody(is(`body1`))

	res2, err := job2.GetResult()
	require.NoError(t, err)

	section(t, res2).expectBody(is(`body2`))
}

func TestBuildParallelSameMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Jobs for the same messageID are shared so fetcher is only called once.
	fetcher := newTestFetcher(m, kr, msg)

	job1, done := b.NewJob(context.Background(), fetcher, msg.ID, ForegroundPriority)
	defer done()

	job2, done := b.NewJob(context.Background(), fetcher, msg.ID, ForegroundPriority)
	defer done()

	res1, err := job1.GetResult()
	require.NoError(t, err)

	section(t, res1).expectBody(is(`body`))

	res2, err := job2.GetResult()
	require.NoError(t, err)

	section(t, res2).expectBody(is(`body`))
}

func TestBuildUndecryptableMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	msg := newTestMessage(t, tests.MakeKeyRing(t), "messageID", "addressID", "text/plain", "body", time.Now())

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg), msg.ID, ForegroundPriority)
	defer done()

	_, err := job.GetResult()
	assert.True(t, errors.Is(err, ErrDecryptionFailed))
}

func TestBuildUndecryptableAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Use a different keyring for encrypting the attachment; it won't be decryptable.
	att := addTestAttachment(t, tests.MakeKeyRing(t), msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	job, done := b.NewJob(context.Background(), newTestFetcher(m, kr, msg, att), msg.ID, ForegroundPriority)
	defer done()

	_, err := job.GetResult()
	assert.True(t, errors.Is(err, ErrDecryptionFailed))
}

func TestBuildCustomMessagePlain(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/plain", "body", time.Now())

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`This message could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `body`)).
		expectTransferEncoding(isMissing())
}

func TestBuildCustomMessageHTML(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`This message could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `<html><body>body</body></html>`)).
		expectTransferEncoding(isMissing())
}

func TestBuildCustomMessageEncrypted(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	body := readerToString(getFileReader("pgp-mime-body-plaintext.eml"))

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	msg.Subject = "this is a subject to make sure we preserve subject"

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectHeader(`Subject`, is(msg.Subject)).
		expectContentType(is(`multipart/encrypted`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-encrypted`))

	section(t, res, 1).
		expectContentType(is(`application/pgp-encrypted`)).
		expectHeader(`Content-Description`, is(`PGP/MIME version identification`)).
		expectBody(is(`Version: 1`))

	section(t, res, 2).
		expectContentType(is(`application/octet-stream`)).
		expectContentTypeParam(`name`, is(`encrypted.asc`)).
		expectContentDisposition(is(`inline`)).
		expectContentDispositionParam(`filename`, is(`encrypted.asc`)).
		expectHeader(`Content-Description`, is(`OpenPGP encrypted message`)).
		expectBody(decryptsTo(foreignKR, body))
}

func TestBuildCustomMessagePlainWithAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/plain", "body", time.Now())
	att := addTestAttachment(t, foreignKR, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg, att),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectContentType(is(`text/plain`)).
		expectBody(contains(`This message could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `body`)).
		expectTransferEncoding(isMissing())

	section(t, res, 2).
		expectContentType(is(`application/octet-stream`)).
		expectBody(contains(`This attachment could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `attachment`)).
		expectContentTypeParam(`name`, is(`file.png.pgp`)).
		expectContentDispositionParam(`filename`, is(`file.png.pgp`)).
		expectTransferEncoding(isMissing())
}

func TestBuildCustomMessageHTMLWithAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, foreignKR, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg, att),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`This message could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `<html><body>body</body></html>`)).
		expectTransferEncoding(isMissing())

	section(t, res, 2).
		expectContentType(is(`application/octet-stream`)).
		expectBody(contains(`This attachment could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `attachment`)).
		expectContentTypeParam(`name`, is(`file.png.pgp`)).
		expectContentDispositionParam(`filename`, is(`file.png.pgp`)).
		expectTransferEncoding(isMissing())
}

func TestBuildCustomMessageOnlyBodyIsUndecryptable(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Use the original keyring for encrypting the attachment; it should decrypt fine.
	att := addTestAttachment(t, kr, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg, att),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectContentType(is(`text/html`)).
		expectBody(contains(`This message could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `<html><body>body</body></html>`)).
		expectTransferEncoding(isMissing())

	section(t, res, 2).
		expectBody(is(`attachment`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`file.png`)).
		expectContentDispositionParam(`filename`, is(`file.png`))
}

func TestBuildCustomMessageOnlyAttachmentIsUndecryptable(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	// Use the original keyring for encrypting the message; it should decrypt fine.
	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Use a different keyring for encrypting the attachment; it won't be decryptable.
	foreignKR := tests.MakeKeyRing(t)
	att := addTestAttachment(t, foreignKR, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	job, done := b.NewJobWithOptions(
		context.Background(),
		newTestFetcher(m, kr, msg, att),
		msg.ID,
		JobOptions{IgnoreDecryptionErrors: true},
		ForegroundPriority,
	)
	defer done()

	res, err := job.GetResult()
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 2).
		expectContentType(is(`application/octet-stream`)).
		expectBody(contains(`This attachment could not be decrypted`)).
		expectBody(decryptsTo(foreignKR, `attachment`)).
		expectContentTypeParam(`name`, is(`file.png.pgp`)).
		expectContentDispositionParam(`filename`, is(`file.png.pgp`)).
		expectTransferEncoding(isMissing())
}

func TestBuildFetchMessageFail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Pretend the message cannot be fetched.
	f := mocks.NewMockFetcher(m)
	f.EXPECT().GetMessage(gomock.Any(), msg.ID).Return(nil, errors.New("oops"))

	// The job should fail, returning an error and a nil result.
	job, done := b.NewJob(context.Background(), f, msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestBuildFetchAttachmentFail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())
	_ = addTestAttachment(t, kr, msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Pretend the attachment cannot be fetched.
	f := mocks.NewMockFetcher(m)
	f.EXPECT().GetMessage(gomock.Any(), msg.ID).Return(msg, nil)
	f.EXPECT().GetAttachment(gomock.Any(), msg.Attachments[0].ID).Return(nil, errors.New("oops"))

	// The job should fail, returning an error and a nil result.
	job, done := b.NewJob(context.Background(), f, msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestBuildNoSuchKeyRing(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	b := NewBuilder(2, 2)
	defer b.Done()

	kr := tests.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Pretend there is no available keyring.
	f := mocks.NewMockFetcher(m)
	f.EXPECT().GetMessage(gomock.Any(), msg.ID).Return(msg, nil)
	f.EXPECT().KeyRingForAddressID(msg.AddressID).Return(nil, errors.New("oops"))

	job, done := b.NewJob(context.Background(), f, msg.ID, ForegroundPriority)
	defer done()

	res, err := job.GetResult()
	assert.Error(t, err)
	assert.Nil(t, res)

	// The returned error should be of this specific type.
	assert.True(t, errors.Is(err, ErrNoSuchKeyRing))
}
