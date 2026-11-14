// Copyright (c) 2026 Proton AG
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
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBuildPlainMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is(`body`)).
		expectTransferEncoding(is(`quoted-printable`))
}

func TestBuildPlainMessageWithLongKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())
	msg.ParsedHeaders.Values["ReallyVeryVeryVeryVeryVeryLongLongLongLongLongLongLongKeyThatWillHaveNotSoLongValue"] = []string{"value"}
	msg.ParsedHeaders.Order = append(msg.ParsedHeaders.Order, "ReallyVeryVeryVeryVeryVeryLongLongLongLongLongLongLongKeyThatWillHaveNotSoLongValue")

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/html`)).
		expectBody(is(`<html><body>body</body></html>`)).
		expectTransferEncoding(is(`quoted-printable`))
}

func TestBuildPlainEncryptedMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	body := readFile(t, "pgp-mime-body-plaintext.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "plaintext-missing-header.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is("How do we know that the ocean is friendly? It waves!\r\n"))
}

func TestBuildPlainEncryptedMessageInvalidHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	body := readFile(t, "plaintext-invalid-header.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectContentType(is(`text/plain`)).
		expectBody(is("MalformedKey Value\r\n\r\nHow do we know that the ocean is friendly? It waves!\r\n"))
}

func TestBuildPlainSignedEncryptedMessageMissingHeader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	body := readFile(t, "plaintext-missing-header.eml")

	kr := utils.MakeKeyRing(t)
	sig := utils.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "plaintext-invalid-header.eml")

	kr := utils.MakeKeyRing(t)
	sig := utils.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "pgp-mime-body-plaintext-latin2.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "pgp-mime-body-html.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "text_plain.eml")

	kr := utils.MakeKeyRing(t)
	sig := utils.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "text_plain_base64.eml")

	kr := utils.MakeKeyRing(t)
	sig := utils.MakeKeyRing(t)

	enc, err := kr.Encrypt(crypto.NewPlainMessageFromString(body), sig)
	require.NoError(t, err)

	arm, err := enc.GetArmored()
	require.NoError(t, err)

	msg := newRawTestMessage("messageID", "addressID", "multipart/mixed", arm, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "pgp-mime-body-signed-plaintext.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "pgp-mime-body-signed-html.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
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

	body := readFile(t, "pgp-mime-body-signed-plaintext-with-pubkey.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`simple plaintext body`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is("\"InfernalBridgeTester@proton.me\" <InfernalbridgeTester@proton.me>")).
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

	body := readFile(t, "pgp-mime-body-signed-html-with-pubkey.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`simple html body`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is("\"InfernalBridgeTester@proton.me\" <InfernalbridgeTester@proton.me>")).
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

	body := readFile(t, "pgp-mime-body-signed-multipart-alternative-with-pubkey.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`Alternative`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is("\"InfernalBridgeTester@proton.me\" <InfernalbridgeTester@proton.me>")).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 2).pubKey(), section(t, res, 2).signature()))

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
		expectBody(contains(`This Rich formated text`)).
		expectBody(contains(`What kind of shoes do ninjas wear`)).
		expectBody(contains(`How does a penguin build its house`))

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

func TestBuildSignedEmbeddedMessageRFC822EncryptedMessageWithPubKey(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	body := readFile(t, "pgp-mime-body-signed-embedded-message-rfc822-with-pubkey.eml")

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`)).
		expectContentType(is(`multipart/signed`)).
		expectContentTypeParam(`micalg`, is(`SHA-256`)).
		expectContentTypeParam(`protocol`, is(`application/pgp-signature`))

	section(t, res, 1).
		expectContentType(is(`multipart/mixed`)).
		expectContentTypeParam(`protected-headers`, is(`v1`)).
		expectHeader(`Subject`, is(`Fwd: simple html body`)).
		expectHeader(`From`, is(`"pm.bridge.qa" <pm.bridge.qa@gmail.com>`)).
		expectHeader(`To`, is("\"InfernalBridgeTester@proton.me\" <InfernalbridgeTester@proton.me>")).
		expectSection(verifiesAgainst(section(t, res, 1, 1, 3).pubKey(), section(t, res, 2).signature()))

	section(t, res, 1, 1).
		expectContentType(is(`multipart/mixed`))

	section(t, res, 1, 1, 1).
		expectContentType(is(`text/plain`))

	section(t, res, 1, 1, 2).
		expectContentType(is(`message/rfc822`)).
		expectContentTypeParam(`name`, is(`simple html body.eml`)).
		expectContentDisposition(is(`attachment`)).
		expectContentDispositionParam(`filename`, is(`simple html body.eml`))

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

func TestBuildHTMLMessageWithAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{})
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

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, &msg, "attachID", "file.eml", "message/rfc822", "attachment", "... message/rfc822 ...")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{})
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

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	inl := addTestAttachment(t, kr, &msg, "inlineID", "file.png", "image/png", "inline", "inline")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{inl}, JobOptions{})
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

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	inl0 := addTestAttachment(t, kr, &msg, "inlineID0", "inline0.png", "image/png", "inline", "inline0")
	inl1 := addTestAttachment(t, kr, &msg, "inlineID1", "inline1.png", "image/png", "inline", "inline1")
	att0 := addTestAttachment(t, kr, &msg, "attachID0", "attach0.png", "image/png", "attachment", "attach0")
	att1 := addTestAttachment(t, kr, &msg, "attachID1", "attach1.png", "image/png", "attachment", "attach1")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{
		inl0,
		inl1,
		att0,
		att1,
	}, JobOptions{})
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

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, &msg, "attachID", `I řeally šhould leařn czech.png`, "image/png", "attachment", "attachment")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{})
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

	veryLongName := strings.Repeat("a", 200) + ".png"

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, kr, &msg, "attachID", veryLongName, "image/png", "attachment", "attachment")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{})
	require.NoError(t, err)

	// NOTE: hasMaxLineLength is too high! Long filenames should be linewrapped using multipart filenames.
	section(t, res, 2).
		expectContentTypeParam(`name`, is(veryLongName)).
		expectHeader(`Content-Type`, contains(veryLongName)).
		expectContentDispositionParam(`filename`, is(veryLongName)).
		expectHeader(`Content-Disposition`, contains(veryLongName)).
		// GODT-2477 - Implement line splitting according to RFC-2184.
		expectSection(hasMaxLineLength(426))
}

func TestBuildMessageDate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).expectDate(is(`Wed, 01 Jan 2020 00:00:00 +0000`))
}

func TestBuildMessageWithInvalidDate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)

	// Create a message with "invalid" (according to applemail) date (before unix time 0).
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Unix(-1, 0))

	// Build the message as usual; the date will be before 1970.
	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`Wed, 31 Dec 1969 23:59:59 +0000`)).
		expectHeader(`X-Original-Date`, isMissing())

	// Build the message with date sanitization enabled; the date will be RFC822's birthdate.
	resFix, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{SanitizeDate: true})
	require.NoError(t, err)

	section(t, resFix).
		expectDate(is(`Fri, 13 Aug 1982 00:00:00 +0000`)).
		expectHeader(`X-Original-Date`, is(`Wed, 31 Dec 1969 23:59:59 +0000`))
}

func TestBuildMessageWithExistingOriginalDate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)

	// Create a new message with existing original date
	msg := newTestMessageWithHeaders(t, kr,
		"messageID",
		"addressID",
		"text/html",
		"<html><body>body</body></html>",
		time.Unix(-1, 0),
		map[string][]string{
			"X-Original-Date": {"Sun, 15 Jan 2023 04:23:03 +0100 (W. Europe Standard Time)"},
			"Date":            {"15-Jan-2023 04:23:13 +0100"},
		})

	// Build the message as usual; the date will be before 1970.
	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectDate(is(`15-Jan-2023 04:23:13 +0100`)).
		expectHeader(`X-Original-Date`, is("Sun, 15 Jan 2023 04:23:03 +0100 (W. Europe Standard Time)"))

	// Build the message with date sanitization enabled; the date will be RFC822's birthdate.
	resFix, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{SanitizeDate: true})
	require.NoError(t, err)

	section(t, resFix).
		expectDate(is(`Fri, 13 Aug 1982 00:00:00 +0000`)).
		expectHeader(`X-Original-Date`, is("Sun, 15 Jan 2023 04:23:03 +0100 (W. Europe Standard Time)"))
}

func TestBuildMessageInternalID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).expectHeader(`Message-Id`, is(`<messageID@protonmail.internalid>`))
}

func TestBuildMessageExternalID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set the message's external ID; this should be used preferentially to set the Message-Id header field.
	msg.ExternalID = "externalID"

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).expectHeader(`Message-Id`, is(`<externalID>`))
}

func TestBuild8BitBody(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)

	// Set an 8-bit body; the charset should be set to UTF-8.
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "I řeally šhould leařn czech", time.Now())

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).expectContentTypeParam(`charset`, is(`utf-8`))
}

func TestBuild8BitSubject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit subject; it should be RFC2047-encoded.
	msg.Subject = `I řeally šhould leařn czech`

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectHeader(`Subject`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech?=`)).
		expectDecodedHeader(`Subject`, is(`I řeally šhould leařn czech`))
}

func TestBuild8BitSender(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit sender; it should be RFC2047-encoded.
	msg.Sender = &mail.Address{
		Name:    `I řeally šhould leařn czech`,
		Address: `mail@example.com`,
	}

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectHeader(`From`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould_lea=C5=99n_czech?= <mail@example.com>`)).
		expectDecodedHeader(`From`, is(`I řeally šhould leařn czech <mail@example.com>`))
}

func TestBuild8BitRecipients(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Set an 8-bit sender; it should be RFC2047-encoded.
	msg.ToList = []*mail.Address{
		{Name: `I řeally šhould`, Address: `mail1@example.com`},
		{Name: `leařn czech`, Address: `mail2@example.com`},
	}

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).
		expectHeader(`To`, is(`=?utf-8?q?I_=C5=99eally_=C5=A1hould?= <mail1@example.com>, =?utf-8?q?lea=C5=99n_czech?= <mail2@example.com>`)).
		expectDecodedHeader(`To`, is(`I řeally šhould <mail1@example.com>, leařn czech <mail2@example.com>`))
}

func TestBuildIncludeMessageIDReference(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Add references.
	msg.ParsedHeaders.Values["References"] = []string{"<myreference@domain.com>"}
	msg.ParsedHeaders.Order = append(msg.ParsedHeaders.Order, "References")

	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.NoError(t, err)

	section(t, res).expectHeader(`References`, is(`<myreference@domain.com>`))

	resRef, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{AddMessageIDReference: true})
	require.NoError(t, err)

	section(t, resRef).expectHeader(`References`, is(`<myreference@domain.com> <messageID@protonmail.internalid>`))
}

func TestBuildMessageStructuralEquivalence(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())
	inl := addTestAttachment(t, kr, &msg, "inlineID", "file.png", "image/png", "inline", "inline")
	att := addTestAttachment(t, kr, &msg, "attachID", "attach.png", "image/png", "attachment", "attachment")

	res1, err := DecryptAndBuildRFC822(kr, msg, [][]byte{inl, att}, JobOptions{})
	require.NoError(t, err)

	res2, err := DecryptAndBuildRFC822(kr, msg, [][]byte{inl, att}, JobOptions{})
	require.NoError(t, err)

	// Boundaries are random so raw bytes differ; verify structural equivalence instead.
	// Structure: multipart/mixed > multipart/related(text/plain + inline image) + attachment
	for _, res := range [][]byte{res1, res2} {
		section(t, res, 1).expectContentType(is(`multipart/related`))
		section(t, res, 1, 1).expectContentType(is(`text/plain`)).expectBody(is(`body`))
		section(t, res, 1, 2).expectContentType(is(`image/png`)).expectBody(is(`inline`))
		section(t, res, 2).expectContentType(is(`image/png`)).expectBody(is(`attachment`))
	}
}

func TestBuildUndecryptableMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	msg := newTestMessage(t, utils.MakeKeyRing(t), "messageID", "addressID", "text/plain", "body", time.Now())

	_, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{})
	require.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestBuildUndecryptableAttachment(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/plain", "body", time.Now())

	// Use a different keyring for encrypting the attachment; it won't be decryptable.
	att := addTestAttachment(t, utils.MakeKeyRing(t), &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	_, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{})
	require.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestBuildCustomMessagePlain(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/plain", "body", time.Now())

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{IgnoreDecryptionErrors: true})
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

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{IgnoreDecryptionErrors: true})
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

	kr := utils.MakeKeyRing(t)

	body := readFile(t, "pgp-mime-body-plaintext.eml")

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "multipart/mixed", body, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	msg.Subject = "this is a subject to make sure we preserve subject"

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, nil, JobOptions{IgnoreDecryptionErrors: true})
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

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/plain", "body", time.Now())
	att := addTestAttachment(t, foreignKR, &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{IgnoreDecryptionErrors: true})
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

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att := addTestAttachment(t, foreignKR, &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{IgnoreDecryptionErrors: true})
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

	kr := utils.MakeKeyRing(t)

	// Use a different keyring for encrypting the message; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	msg := newTestMessage(t, foreignKR, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Use the original keyring for encrypting the attachment; it should decrypt fine.
	att := addTestAttachment(t, kr, &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{IgnoreDecryptionErrors: true})
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

	// Use the original keyring for encrypting the message; it should decrypt fine.
	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())

	// Use a different keyring for encrypting the attachment; it won't be decryptable.
	foreignKR := utils.MakeKeyRing(t)
	att := addTestAttachment(t, foreignKR, &msg, "attachID", "file.png", "image/png", "attachment", "attachment")

	// Tell the job to ignore decryption errors; a custom message will be returned instead of an error.
	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{att}, JobOptions{IgnoreDecryptionErrors: true})
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

func readFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", path))
	require.NoError(t, err)

	return string(b)
}

func TestBuildComplexMIMEType(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	kr := utils.MakeKeyRing(t)
	msg := newTestMessage(t, kr, "messageID", "addressID", "text/html", "<html><body>body</body></html>", time.Now())
	att0 := addTestAttachment(t, kr, &msg, "attachID0", "attach0.png", "image/png", "attachment", "attach0")
	att1 := addTestAttachment(t, kr, &msg, "attachID1", "Cat_August_2010-4.jpeg", "image/jpeg; name=Cat_August_2010-4.jpeg; x-unix-mode=0644", "attachment", "attach1")

	res, err := DecryptAndBuildRFC822(kr, msg, [][]byte{
		att0,
		att1,
	}, JobOptions{})
	require.NoError(t, err)

	section(t, res, 1).
		expectBody(is(`<html><body>body</body></html>`)).
		expectContentType(is(`text/html`)).
		expectTransferEncoding(is(`quoted-printable`))

	section(t, res, 2).
		expectBody(is(`attach0`)).
		expectContentType(is(`image/png`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`attach0.png`)).
		expectContentDispositionParam(`filename`, is(`attach0.png`))

	section(t, res, 3).
		expectBody(is(`attach1`)).
		expectContentType(is(`image/jpeg`)).
		expectTransferEncoding(is(`base64`)).
		expectContentTypeParam(`name`, is(`Cat_August_2010-4.jpeg`)).
		expectContentDispositionParam(`filename`, is(`Cat_August_2010-4.jpeg`))
}

func TestHasMBOXHeaderLine(t *testing.T) {
	cases := map[string]struct {
		index, indexCRLF int
	}{
		// No MBOX line and missing header.
		"From: ok\nTo: Ok": {-1, -1},

		// MBOX line in body, not header.
		"From: ok\nTo: Ok\n\nFrom - 123":  {-1, -1},
		"From: ok\nTo: Ok\n\n>From - 123": {-1, -1},

		// MBOX lines without proper header ending.
		"From ok\nFrom: ok\nTo: Ok\nDate: ok":  {-1, -1},
		">From ok\nFrom: ok\nTo: Ok\nDate: ok": {-1, -1},

		// MBOX lines with proper header ending.
		"From ok\nFrom: ok\nTo: Ok\nDate: ok\n\n":  {0, 0},
		">From ok\nFrom: ok\nTo: Ok\nDate: ok\n\n": {0, 0},

		// MBOX lines in middle of headers.
		"From: ok\nFrom middle\nTo: Ok\nDate: ok\n\n":  {9, 10},
		"From: ok\n>From middle\nTo: Ok\nDate: ok\n\n": {9, 10},

		// Multiple MBOX lines, should return first one.
		"From first\nFrom second\nFrom: ok\nTo: Ok\nDate: ok\n\n":  {0, 0},
		"From first\n>From second\nFrom: ok\nTo: Ok\nDate: ok\n\n": {0, 0},
		">From first\nFrom second\nFrom: ok\nTo: Ok\nDate: ok\n\n": {0, 0},

		// MBOX lines at various positions.
		"From: sender\nTo: recipient\nFrom mbox\nDate: today\n\n":  {27, 29},
		"From: sender\nTo: recipient\n>From mbox\nDate: today\n\n": {27, 29},
		"From: sender\nTo: recipient\nDate: today\nFrom mbox\n\n":  {39, 42},
		"From: sender\nTo: recipient\nDate: today\n>From mbox\n\n": {39, 42},

		// MBOX line at end of header section.
		"From: ok\nTo: Ok\nDate: ok\nFrom mbox\n\n":  {25, 28},
		"From: ok\nTo: Ok\nDate: ok\n>From mbox\n\n": {25, 28},

		// Headers with missing required fields.
		"From mbox\nFrom: ok\nTo: Ok\n\n":   {-1, -1},
		"From mbox\nFrom: ok\nDate: ok\n\n": {-1, -1},
		"From mbox\nTo: Ok\nDate: ok\n\n":   {-1, -1},

		// Valid headers but no MBOX lines.
		"From: sender\nTo: recipient\nDate: today\n\n":                {-1, -1},
		"From: sender\nTo: recipient\nDate: today\nSubject: test\n\n": {-1, -1},

		// Headers that are like-MBOX but not.
		"From: sender\n>From: other\nTo: recipient\nDate: today\n\n": {-1, -1},
		"From: sender\nFrom: other\nTo: recipient\nDate: today\n\n":  {-1, -1},

		"From mbox\nFrom: sender\n>From: header\nTo: recipient\nDate: today\n\n": {0, 0},
		"From: sender\nFrom mbox\n>From: header\nTo: recipient\nDate: today\n\n": {13, 14},
	}

	test := func(t *testing.T, wantIndex int, given string, useCRLF bool) {
		decrypted := &DecryptedMessage{}

		if useCRLF {
			decrypted.Body = *bytes.NewBufferString(strings.ReplaceAll(given, "\n", "\r\n"))
		} else {
			decrypted.Body = *bytes.NewBufferString(given)
		}

		headerIdx := indexMBOXHeaderLine(decrypted.Body.Bytes())
		require.Equal(t, wantIndex, headerIdx)

		if headerIdx == -1 {
			return
		}

		partFromOne := decrypted.Body.Bytes()[headerIdx : headerIdx+5]
		partFromTwo := decrypted.Body.Bytes()[headerIdx : headerIdx+6]
		hasMboxHeader := strings.HasPrefix(string(partFromOne), "From ") || strings.HasPrefix(string(partFromOne), ">From ") ||
			strings.HasPrefix(string(partFromTwo), "From ") || strings.HasPrefix(string(partFromTwo), ">From ")

		require.True(t, hasMboxHeader)
	}

	for given, want := range cases {
		t.Run("LF-"+given, func(t *testing.T) { test(t, want.index, given, false) })
		t.Run("CRLF-"+given, func(t *testing.T) { test(t, want.indexCRLF, given, true) })
	}
}

func TestSanitizeMBOXHeaderLine(t *testing.T) {
	cases := map[string]string{
		// Unchanged - no MBOX headers in header section
		"From: ok\nTo: Ok":                "From: ok\nTo: Ok",
		"From: ok\nTo: Ok\n\nFrom - 123":  "From: ok\nTo: Ok\n\nFrom - 123",
		"From: ok\nTo: Ok\n\n>From - 123": "From: ok\nTo: Ok\n\n>From - 123",

		// Unchanged - no MBOX headers.
		"From: ok\nTo: ok\nDate: ok":     "From: ok\nTo: ok\nDate: ok",
		"From: ok\nTo: ok\nDate: ok\n\n": "From: ok\nTo: ok\nDate: ok\n\n",

		// MBOX headers should be removed
		"From ok\nFrom: ok\nTo: ok\nDate: ok\n\n":  "From: ok\nTo: ok\nDate: ok\n\n",
		">From ok\nFrom: ok\nTo: ok\nDate: ok\n\n": "From: ok\nTo: ok\nDate: ok\n\n",

		// MBOX header mixed in-between.
		"From: sender\nFrom line\nTo: recipient\nDate: today\n\n":  "From: sender\nTo: recipient\nDate: today\n\n",
		"From: sender\n>From line\nTo: recipient\nDate: today\n\n": "From: sender\nTo: recipient\nDate: today\n\n",
		"From line\nFrom: sender\nTo: recipient\nDate: today\n\n":  "From: sender\nTo: recipient\nDate: today\n\n",

		// Multiple MBOX headers.
		"From line1\nFrom line2\nFrom: sender\nTo: recipient\nDate: today\n\n":  "From: sender\nTo: recipient\nDate: today\n\n",
		"From line1\n>From line2\nFrom: sender\nTo: recipient\nDate: today\n\n": "From: sender\nTo: recipient\nDate: today\n\n",

		// Incomplete required headers - should not process.
		"From line\nFrom: sender\nTo: recipient\n\n": "From line\nFrom: sender\nTo: recipient\n\n",
		"From line\nFrom: sender\nDate: today\n\n":   "From line\nFrom: sender\nDate: today\n\n",

		// MBOX and complete required headers - No separation between header part - should not be processed.
		"From ok\nFrom: ok\nTo: ok\nDate: ok":  "From ok\nFrom: ok\nTo: ok\nDate: ok",
		">From ok\nFrom: ok\nTo: ok\nDate: ok": ">From ok\nFrom: ok\nTo: ok\nDate: ok",
	}

	test := func(t *testing.T, given, want string, useCRLF bool) {
		decrypted := &DecryptedMessage{}

		if useCRLF {
			decrypted.Body = *bytes.NewBufferString(strings.ReplaceAll(given, "\n", "\r\n"))
			want = strings.ReplaceAll(want, "\n", "\r\n")
		} else {
			decrypted.Body = *bytes.NewBufferString(given)
		}

		require.NoError(t, sanitizeMBOXHeaderLine(decrypted))
		require.Equal(t, []byte(want), decrypted.Body.Bytes())

		require.NoError(t, sanitizeMBOXHeaderLine(decrypted))
		require.Equal(t, []byte(want), decrypted.Body.Bytes())
	}

	for given, want := range cases {
		t.Run("LF"+given, func(t *testing.T) { test(t, given, want, false) })
		t.Run("CRLF"+given, func(t *testing.T) { test(t, given, want, true) })
	}
}

func TestSanitizeMBOXHeaderLineManyLines(t *testing.T) {
	const lines = 5000

	headers := "From: sender\nTo: recipient\nDate: today\n\nbody\n"
	body := strings.Repeat("From mbox\n", lines) + headers

	decrypted := &DecryptedMessage{}
	decrypted.Body = *bytes.NewBufferString(body)

	require.NoError(t, sanitizeMBOXHeaderLine(decrypted))
	require.Equal(t, headers, decrypted.Body.String())

	// The body must not be rescanned and copied again for every removed line.
	allocs := testing.AllocsPerRun(5, func() {
		decrypted := &DecryptedMessage{}
		decrypted.Body = *bytes.NewBufferString(body)

		_ = sanitizeMBOXHeaderLine(decrypted)
	})
	require.Less(t, allocs, float64(lines/10))
}
