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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package smtp

import (
	"encoding/base64"
	"testing"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/ProtonMail/proton-bridge/v3/utils"
	"github.com/stretchr/testify/assert"
)

var mimeBody message.MIMEBody = "Mime-Version: 1.0\r\nContent-Type: text/html;charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\nFrom: Bridge Test <bridge.test@proton.local>\r\nTo: Test <test@proton.local>\r\nSubject: HTML text with trailing whitespace\r\n\r\n<html><head></head><body>This is body of <b>HTML mail</b>\r\n \t\r\n with trailing whitespace</body></html>"
var richBody message.Body = "<html><head></head><body>This is body of <b>HTML mail</b>\r\n \t\r\n with trailing whitespace</body></html>"
var plainBody message.Body = "This is body of *HTML mail*\r\n \t\r\n with trailing whitespace"

type SendReqStruct struct {
	name string

	Encrypt          bool
	SignatureType    proton.SignatureType
	EncryptionScheme proton.EncryptionScheme
	MIMEType         rfc822.MIMEType

	wantMIMEType rfc822.MIMEType
	wantType     proton.EncryptionScheme
	wantBody     string

	wantError bool
}

// | Encrypt     | MIME Type               | Signature Type             |
// | true | false| text | html | multipart | Detached | Attached | None |
// |------|------|------|------|-----------|----------|----------|------|
// | OK   | KO   | KO   | KO   | OK        | OK       | KO       | KO   |.
func TestCreateSendReq_PGPMIMEScheme(t *testing.T) {
	kr := utils.MakeKeyRing(t)

	tests := []SendReqStruct{
		{
			name: "PGPMIMEScheme Text Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantMIMEType: rfc822.MultipartMixed,
			wantType:     proton.PGPMIMEScheme,
			wantBody:     string(mimeBody),
			wantError:    false,
		},
		{
			name: "PGPMIMEScheme Text clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Text Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Text clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Text Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Text clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme HTML clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPMIMEScheme Multipart clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
	}
	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.
		t.Run(test.name, func(t *testing.T) { checkCreateSendReq(t, test, kr) })
	}
}

// | Encrypt         | MIME Type               | Signature Type             |
// | true   | false  | text | html | multipart | Detached | Attached | None |
// |--------|--------|------|------|-----------|----------|----------|------|
// | forced to false | KO   | KO   | OK        | OK       | KO       | KO   |.
func TestCreateSendReq_ClearMIMEScheme(t *testing.T) {
	kr := utils.MakeKeyRing(t)

	tests := []SendReqStruct{
		{
			name: "ClearMIMEScheme Text Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantMIMEType: rfc822.MultipartMixed,
			wantType:     proton.ClearMIMEScheme,
			wantBody:     string(mimeBody),
			wantError:    false,
		},
		{
			name: "ClearMIMEScheme Text clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantMIMEType: rfc822.MultipartMixed,
			wantType:     proton.ClearMIMEScheme,
			wantBody:     string(mimeBody),
			wantError:    false,
		},
		{
			name: "ClearMIMEScheme Text Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Text clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Text Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Text clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme HTML clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearMIMEScheme Multipart clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearMIMEScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
	}
	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.
		t.Run(test.name, func(t *testing.T) { checkCreateSendReq(t, test, kr) })
	}
}

// | Encrypt     | MIME Type               | Signature Type             |
// | true | false| text | html | multipart | Detached | Attached | None |
// |------|------|------|------|-----------|----------|----------|------|
// | OK   | KO   | OK   | OK   | KO        | OK       | KO       | KO   |.
func TestCreateSendReq_InternalScheme(t *testing.T) {
	kr := utils.MakeKeyRing(t)

	tests := []SendReqStruct{
		{
			name: "InternalScheme Text Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.InternalScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "InternalScheme HTML Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantMIMEType: rfc822.TextHTML,
			wantType:     proton.InternalScheme,
			wantBody:     string(richBody),
			wantError:    false,
		},
		{
			name: "InternalScheme Multipart Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "InternalScheme Text clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "InternalScheme HTML clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "InternalScheme Multipart clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "InternalScheme Text Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "InternalScheme HTML Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "InternalScheme Multipart Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "InternalScheme Text clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "InternalScheme HTML clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "InternalScheme Multipart clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "InternalScheme Text Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "InternalScheme HTML Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "InternalScheme Multipart Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "InternalScheme Text clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "InternalScheme HTML clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "InternalScheme Multipart clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.InternalScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
	}
	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.
		t.Run(test.name, func(t *testing.T) { checkCreateSendReq(t, test, kr) })
	}
}

// | Encrypt     | MIME Type               | Signature Type             |
// | true | false| text | html | multipart | Detached | Attached | None |
// |------|------|------|------|-----------|----------|----------|------|
// | OK   | KO   | OK   | KO   | KO        | OK       | KO       | KO   |.
func TestCreateSendReq_PGPInlineScheme(t *testing.T) {
	kr := utils.MakeKeyRing(t)

	tests := []SendReqStruct{
		{
			name: "PGPInlineScheme Text Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.PGPInlineScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "PGPInlineScheme HTML Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Text clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPInlineScheme HTML clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Text Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPInlineScheme HTML Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		// investigate during GODT-3097
		// {
		// 	name: "PGPInlineScheme Text clear AttachedSignature",
		//
		//	Encrypt:          false,
		//	SignatureType:    proton.AttachedSignature,
		//	EncryptionScheme: proton.PGPInlineScheme,
		//	MIMEType:         rfc822.TextPlain,
		//
		//	wantMIMEType: rfc822.TextPlain,
		//	wantType:     proton.PGPInlineScheme,
		//	wantBody:     string(plainBody),
		//	wantError:    false,
		// },
		{
			name: "PGPInlineScheme HTML clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Text Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "PGPInlineScheme HTML Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		// investigate during GODT-3097
		// {
		//	name: "PGPInlineScheme Text clear NoSignature",
		//
		//	Encrypt:          false,
		//	SignatureType:    proton.NoSignature,
		//	EncryptionScheme: proton.PGPInlineScheme,
		//	MIMEType:         rfc822.TextPlain,
		//
		//	wantMIMEType: rfc822.TextPlain,
		//	wantType:     proton.PGPInlineScheme,
		//	wantBody:     string(plainBody),
		//	wantError:    false,
		// },
		{
			name: "PGPInlineScheme HTML clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "PGPInlineScheme Multipart clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.PGPInlineScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
	}
	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.
		t.Run(test.name, func(t *testing.T) { checkCreateSendReq(t, test, kr) })
	}
}

// | Encrypt         | MIME Type               | Signature Type             |
// | true   | false  | text | html | multipart | Detached | Attached | None |
// |--------|--------|------|------|-----------|----------|----------|------|
// | forced to false | OK   | OK   | KO        | OK       | OK       | OK   |.
func TestCreateSendReq_ClearScheme(t *testing.T) {
	kr := utils.MakeKeyRing(t)

	tests := []SendReqStruct{
		{
			name: "ClearScheme Text Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.ClearScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "ClearScheme HTML Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantMIMEType: rfc822.TextHTML,
			wantType:     proton.ClearScheme,
			wantBody:     string(richBody),
			wantError:    false,
		},
		{
			name: "ClearScheme Multipart Encrypt DetachedSignature",

			Encrypt:          true,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearScheme Text clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.ClearScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "ClearScheme HTML clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearScheme Multipart clear DetachedSignature",

			Encrypt:          false,
			SignatureType:    proton.DetachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearScheme Text Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearScheme HTML Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearScheme Multipart Encrypt AttachedSignature",

			Encrypt:          true,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearScheme Text clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.ClearScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "ClearScheme HTML clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantMIMEType: rfc822.TextHTML,
			wantType:     proton.ClearScheme,
			wantBody:     string(richBody),
			wantError:    false,
		},
		{
			name: "ClearScheme Multipart clear AttachedSignature",

			Encrypt:          false,
			SignatureType:    proton.AttachedSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearScheme Text Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantError: true,
		},
		{
			name: "ClearScheme HTML Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantError: true,
		},
		{
			name: "ClearScheme Multipart Encrypt NoSignature",

			Encrypt:          true,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
		{
			name: "ClearScheme Text clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextPlain,

			wantMIMEType: rfc822.TextPlain,
			wantType:     proton.ClearScheme,
			wantBody:     string(plainBody),
			wantError:    false,
		},
		{
			name: "ClearScheme HTML clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.TextHTML,

			wantMIMEType: rfc822.TextHTML,
			wantType:     proton.ClearScheme,
			wantBody:     string(richBody),
			wantError:    false,
		},
		{
			name: "ClearScheme Multipart clear NoSignature",

			Encrypt:          false,
			SignatureType:    proton.NoSignature,
			EncryptionScheme: proton.ClearScheme,
			MIMEType:         rfc822.MultipartMixed,

			wantError: true,
		},
	}
	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.
		t.Run(test.name, func(t *testing.T) { checkCreateSendReq(t, test, kr) })
	}
}

func checkCreateSendReq(t *testing.T, test SendReqStruct, kr *crypto.KeyRing) {
	var rec = make(recipients)
	var pref = proton.SendPreferences{
		Encrypt:          test.Encrypt,
		SignatureType:    test.SignatureType,
		EncryptionScheme: test.EncryptionScheme,
		MIMEType:         test.MIMEType}
	if test.Encrypt {
		pref.PubKey = kr
	}
	rec["test@proton.local"] = pref

	req, err := createSendReq(kr, mimeBody, richBody, plainBody, rec, map[string]*crypto.SessionKey{})
	if test.wantError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, "Error while CreateSendReq")
		}
		assert.Equal(t, 1, len(req.Packages))
		if len(req.Packages) != 1 {
			assert.FailNow(t, "No package created")
		}
		assert.Equal(t, test.wantMIMEType, req.Packages[0].MIMEType)
		assert.Equal(t, test.wantType, req.Packages[0].Type)

		decBody, err := base64.StdEncoding.DecodeString(req.Packages[0].Body)
		assert.NoError(t, err)

		if test.Encrypt && test.EncryptionScheme != proton.ClearMIMEScheme && test.EncryptionScheme != proton.ClearScheme {
			assert.Nil(t, req.Packages[0].BodyKey)
			if req.Packages[0].BodyKey != nil {
				assert.FailNow(t, "BodyKey should be nil when Encrypt is set")
			}
			decBodyKey, err := base64.StdEncoding.DecodeString(req.Packages[0].Addresses["test@proton.local"].BodyKeyPacket)
			assert.NoError(t, err)
			split := crypto.NewPGPSplitMessage(decBodyKey, decBody)
			plain, err := kr.Decrypt(split.GetPGPMessage(), kr, crypto.GetUnixTime())
			assert.NoError(t, err)
			assert.Equal(t, test.wantBody, string(plain.Data))
		} else {
			assert.NotNil(t, req.Packages[0].BodyKey)
			if req.Packages[0].BodyKey == nil {
				assert.FailNow(t, "BodyKey should not be nil when Encrypt is not set")
			}
			decBodyKey, err := base64.StdEncoding.DecodeString(req.Packages[0].BodyKey.Key)
			assert.NoError(t, err)
			sk := crypto.SessionKey{Key: decBodyKey, Algo: req.Packages[0].BodyKey.Algorithm}
			plain, err := sk.Decrypt(decBody)
			assert.NoError(t, err)
			assert.Equal(t, test.wantBody, string(plain.Data))
		}
	}
}

func TestCreateSendReq_MultiRecipients(t *testing.T) {
	kr := utils.MakeKeyRing(t)
	var rec = make(recipients)
	rec["pgpmimeMultipartEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.PGPMIMEScheme,
		MIMEType:         rfc822.MultipartMixed,
	}
	rec["ClearMimeMultipartEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.ClearMIMEScheme,
		MIMEType:         rfc822.MultipartMixed,
	}
	rec["ClearMimeMultipartClearDetached@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.ClearMIMEScheme,
		MIMEType:         rfc822.MultipartMixed,
	}
	rec["InternalTextEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.InternalScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["InternalHTMLEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.InternalScheme,
		MIMEType:         rfc822.TextHTML,
	}
	rec["InilineTextEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.PGPInlineScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["ClearTextEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["ClearHTMLEncryptDetached@proton.local"] = proton.SendPreferences{
		PubKey:           kr,
		Encrypt:          true,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextHTML,
	}
	rec["ClearTextClearDetached@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.DetachedSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["ClearTextClearAttached@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.AttachedSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["ClearHTMLClearAttached@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.AttachedSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextHTML,
	}
	rec["ClearTextClearNone@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.NoSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextPlain,
	}
	rec["ClearHTMLClearNone@proton.local"] = proton.SendPreferences{
		Encrypt:          false,
		SignatureType:    proton.NoSignature,
		EncryptionScheme: proton.ClearScheme,
		MIMEType:         rfc822.TextHTML,
	}

	req, err := createSendReq(kr, mimeBody, richBody, plainBody, rec, map[string]*crypto.SessionKey{})
	assert.NoError(t, err)

	// expect 3 packages: Multipart/HTML/text
	assert.Equal(t, 3, len(req.Packages))

	// check that there is the appropriate number of recipients in the packages
	var totalRec = 0
	var totalFromRecipientKey = 0
	var totalFromSessionKey = 0
	for _, pkg := range req.Packages {
		var wantBody string
		switch {
		case pkg.MIMEType == rfc822.TextHTML:
			wantBody = string(richBody)
		case pkg.MIMEType == rfc822.TextPlain:
			wantBody = string(plainBody)
		default:
			wantBody = string(mimeBody)
		}

		decBody, err := base64.StdEncoding.DecodeString(pkg.Body)
		assert.NoError(t, err)

		// check every recipient can decrypt
		for _, addr := range pkg.Addresses {
			if addr.BodyKeyPacket != "" {
				decBodyKey, err := base64.StdEncoding.DecodeString(addr.BodyKeyPacket)
				assert.NoError(t, err)
				split := crypto.NewPGPSplitMessage(decBodyKey, decBody)
				plain, err := kr.Decrypt(split.GetPGPMessage(), kr, crypto.GetUnixTime())
				assert.NoError(t, err)
				assert.Equal(t, wantBody, string(plain.Data))
				totalFromRecipientKey++
			} else {
				assert.NotNil(t, pkg.BodyKey)
				decBodyKey, err := base64.StdEncoding.DecodeString(pkg.BodyKey.Key)
				assert.NoError(t, err)
				sk := crypto.SessionKey{Key: decBodyKey, Algo: pkg.BodyKey.Algorithm}
				plain, err := sk.Decrypt(decBody)
				assert.NoError(t, err)
				assert.Equal(t, wantBody, string(plain.Data))
				totalFromSessionKey++
			}
		}

		totalRec += len(pkg.Addresses)
	}
	assert.Equal(t, len(rec), totalRec)
	// 6 without encryption
	assert.Equal(t, 6, totalFromRecipientKey)
	// 7 with encryption
	assert.Equal(t, 7, totalFromSessionKey)
}
