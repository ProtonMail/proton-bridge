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

package pmapi

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/require"
)

type recipient struct {
	email       string
	sendScheme  PackageFlag
	pubkey      *crypto.KeyRing
	signature   SignatureFlag
	contentType string
	doEncrypt   bool
	wantError   error
}

type testData struct {
	emails       []string
	recipients   []recipient
	wantPackages []*MessagePackage

	allRecipients map[string]recipient
	allAddresses  map[string]*MessageAddress

	attKeys                       map[string]*crypto.SessionKey
	mimeBody, plainBody, richBody string
}

func (td *testData) addRecipients(t testing.TB) {
	for _, email := range td.emails {
		rcp, ok := td.allRecipients[email]
		require.True(t, ok, "missing recipient %s", email)
		rcp.email = email
		td.recipients = append(td.recipients, rcp)
	}
}

func (td *testData) addAddresses(t testing.TB) {
	for i, wantPackage := range td.wantPackages {
		for email := range wantPackage.Addresses {
			address, ok := td.allAddresses[email]
			require.True(t, ok, "missing address %s", email)
			td.wantPackages[i].Addresses[email] = address
		}
	}
}

func (td *testData) prepareAndCheck(t *testing.T) {
	r := require.New(t)

	matchPresence := func(want string) require.ValueAssertionFunc {
		if len(want) == 0 {
			return require.Empty
		}
		return require.NotEmpty
	}

	have := NewSendMessageReq(testPrivateKeyRing, td.mimeBody, td.plainBody, td.richBody, td.attKeys)
	for _, rec := range td.recipients {
		err := have.AddRecipient(rec.email, rec.sendScheme, rec.pubkey, rec.signature, rec.contentType, rec.doEncrypt)

		if rec.wantError == nil {
			r.NoError(err, "email %s", rec.email)
		} else {
			r.EqualError(err, rec.wantError.Error(), "email %s", rec.email)
		}
	}
	have.PreparePackages()

	r.Equal(len(td.wantPackages), len(have.Packages))

	for i, wantPackage := range td.wantPackages {
		havePackage := have.Packages[i]

		r.Equal(wantPackage.MIMEType, havePackage.MIMEType, "pkg %d", i)
		r.Equal(wantPackage.Type, havePackage.Type, "pkg %d", i)

		r.Equal(len(wantPackage.Addresses), len(havePackage.Addresses), "pkg %d", i)
		for email, wantAddress := range wantPackage.Addresses {
			haveAddress, ok := havePackage.Addresses[email]
			r.True(ok, "pkg %d email %s", i, email)

			r.Equal(wantAddress.Type, haveAddress.Type, "pkg %d email %s", i, email)
			matchPresence(wantAddress.EncryptedBodyKeyPacket)(t, haveAddress.EncryptedBodyKeyPacket, "pkg %d email %s", i, email)
			r.Equal(wantAddress.Signature, haveAddress.Signature, "pkg %d email %s", i, email)

			if len(td.attKeys) == 0 {
				r.Len(haveAddress.EncryptedAttachmentKeyPackets, 0)
			} else {
				r.Equal(
					len(wantAddress.EncryptedAttachmentKeyPackets),
					len(haveAddress.EncryptedAttachmentKeyPackets),
					"pkg %d email %s", i, email,
				)
				for attID, wantAttKey := range wantAddress.EncryptedAttachmentKeyPackets {
					haveAttKey, ok := haveAddress.EncryptedAttachmentKeyPackets[attID]
					r.True(ok, "pkg %d email %s att %s", i, email, attID)
					matchPresence(wantAttKey)(t, haveAttKey, "pkg %d email %s att %s", i, email, attID)
				}
			}
		}

		matchPresence(wantPackage.EncryptedBody)(t, havePackage.EncryptedBody, "pkg %d", i)

		wantBodyKey := wantPackage.DecryptedBodyKey
		haveBodyKey := havePackage.DecryptedBodyKey

		if wantBodyKey == nil {
			r.Nil(haveBodyKey, "pkg %d: expected empty body key but got %v", i, haveBodyKey)
		} else {
			r.NotNil(haveBodyKey, "pkg %d: expected body key but got nil", i)
			r.NotEmpty(haveBodyKey.Algorithm, "pkg %d", i)
			r.NotEmpty(haveBodyKey.Key, "pkg %d", i)
		}

		if len(td.attKeys) == 0 {
			r.Len(havePackage.DecryptedAttachmentKeys, 0)
		} else {
			r.Equal(
				len(wantPackage.DecryptedAttachmentKeys),
				len(havePackage.DecryptedAttachmentKeys),
				"pkg %d", i,
			)
			for attID, wantAttKey := range wantPackage.DecryptedAttachmentKeys {
				haveAttKey, ok := havePackage.DecryptedAttachmentKeys[attID]
				r.True(ok, "pkg %d att %s", i, attID)
				matchPresence(wantAttKey.Key)(t, haveAttKey.Key, "pkg %d att %s", i, attID)
				matchPresence(wantAttKey.Algorithm)(t, haveAttKey.Algorithm, "pkg %d att %s", i, attID)
			}
		}
	}

	haveBytes, err := json.Marshal(have)
	r.NoError(err)
	haveString := string(haveBytes)
	// Added `:` to avoid false-fail if the whole output results to empty object.
	r.NotContains(haveString, ":\"\"", "found empty string: %s", haveString)
	r.NotContains(haveString, ":[]", "found empty list: %s", haveString)
	r.NotContains(haveString, ":{}", "found empty object: %s", haveString)
	r.NotContains(haveString, ":null", "found null: %s", haveString)
}

func TestSendReq(t *testing.T) {
	attKeyB64 := "EvjO/2RIJNn6HdoU6ACqFdZglzJhpjQ/PpjsvL3mB5Q="
	token, err := base64.StdEncoding.DecodeString(attKeyB64)
	require.NoError(t, err)

	attKey := crypto.NewSessionKeyFromToken(token, "aes256")
	attKeyPackets := map[string]string{"attID": "not-empty"}
	attAlgoKeys := map[string]AlgoKey{"attID": {"not-empty", "not-empty"}}

	allRecipients := map[string]recipient{
		// Internal OK
		"none@pm.me":  {"", InternalPackage, testPublicKeyRing, SignatureDetached, "", true, nil},
		"html@pm.me":  {"", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
		"plain@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
		// Internal bad
		"wrongtype@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureDetached, "application/rfc822", true, errUnknownContentType},
		"multipart@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, errMultipartInNonMIME},
		"noencrypt@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, false, errInternalMustEncrypt},
		"no-pubkey@pm.me": {"", InternalPackage, nil, SignatureDetached, ContentTypeHTML, true, errMissingPubkey},
		"nosigning@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureNone, ContentTypeHTML, true, errEncryptMustSign},
		// testing combination
		"internal1@pm.me": {"", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
		// Clear OK
		"html@email.com":       {"", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
		"none@email.com":       {"", ClearPackage, nil, SignatureNone, "", false, nil},
		"plain@email.com":      {"", ClearPackage, nil, SignatureNone, ContentTypePlainText, false, nil},
		"plain-sign@email.com": {"", ClearPackage, nil, SignatureDetached, ContentTypePlainText, false, nil},
		"mime-sign@email.com":  {"", ClearMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, false, nil},
		// Clear bad
		"mime@email.com":             {"", ClearMIMEPackage, nil, SignatureNone, ContentTypeMultipartMixed, false, errClearMIMEMustSign},
		"clear-plain-sign@email.com": {"", PGPInlinePackage, nil, SignatureDetached, ContentTypePlainText, false, errClearSignMustNotBePGPInline},
		"html-sign@email.com":        {"", ClearPackage, nil, SignatureDetached, ContentTypeHTML, false, errClearSignMustNotBeHTML},
		"mime-plain@email.com":       {"", ClearMIMEPackage, nil, SignatureDetached, ContentTypePlainText, false, errMIMEMustBeMultipart},
		"mime-html@email.com":        {"", ClearMIMEPackage, nil, SignatureDetached, ContentTypeHTML, false, errMIMEMustBeMultipart},
		// External Encryption OK
		"mime@gpg.com":  {"", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},
		"plain@gpg.com": {"", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
		// External Encryption bad
		"eo@gpg.com":           {"", EncryptedOutsidePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, errEncryptedOutsideNotSupported},
		"inline-html@gpg.com":  {"", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, errInlineMustBePlain},
		"inline-mixed@gpg.com": {"", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, errMultipartInNonMIME},
		"mime-plain@gpg.com":   {"", PGPMIMEPackage, nil, SignatureDetached, ContentTypePlainText, true, errMIMEMustBeMultipart},
		"mime-html@sgpg.com":   {"", PGPMIMEPackage, nil, SignatureDetached, ContentTypeHTML, true, errMIMEMustBeMultipart},
		"no-pubkey@gpg.com":    {"", PGPMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, true, errMissingPubkey},
		"not-signed@gpg.com":   {"", PGPMIMEPackage, testPublicKeyRing, SignatureNone, ContentTypeMultipartMixed, true, errEncryptMustSign},
	}

	allAddresses := map[string]*MessageAddress{
		"none@pm.me": {
			Type:                          InternalPackage,
			Signature:                     SignatureDetached,
			EncryptedBodyKeyPacket:        "not-empty",
			EncryptedAttachmentKeyPackets: attKeyPackets,
		},
		"plain@pm.me": {
			Type:                          InternalPackage,
			Signature:                     SignatureDetached,
			EncryptedBodyKeyPacket:        "not-empty",
			EncryptedAttachmentKeyPackets: attKeyPackets,
		},
		"html@pm.me": {
			Type:                          InternalPackage,
			Signature:                     SignatureDetached,
			EncryptedBodyKeyPacket:        "not-empty",
			EncryptedAttachmentKeyPackets: attKeyPackets,
		},
		"internal1@pm.me": {
			Type:                          InternalPackage,
			Signature:                     SignatureDetached,
			EncryptedBodyKeyPacket:        "not-empty",
			EncryptedAttachmentKeyPackets: attKeyPackets,
		},

		"html@email.com": {
			Type:      ClearPackage,
			Signature: SignatureNone,
		},
		"none@email.com": {
			Type:      ClearPackage,
			Signature: SignatureNone,
		},
		"plain@email.com": {
			Type:      ClearPackage,
			Signature: SignatureNone,
		},
		"plain-sign@email.com": {
			Type:      ClearPackage,
			Signature: SignatureDetached,
		},
		"mime-sign@email.com": {
			Type:      ClearMIMEPackage,
			Signature: SignatureDetached,
		},

		"mime@gpg.com": {
			Type:                   PGPMIMEPackage,
			Signature:              SignatureDetached,
			EncryptedBodyKeyPacket: "non-empty",
		},
		"plain@gpg.com": {
			Type:                          PGPInlinePackage,
			Signature:                     SignatureDetached,
			EncryptedBodyKeyPacket:        "non-empty",
			EncryptedAttachmentKeyPackets: attKeyPackets,
		},
	}

	// NOTE naming
	// Single: there should be one package
	// Multiple: there should be more than one package
	// Internal: there should be internal package
	// Clear: there should be non-encrypted package
	// Encrypted: there should be encrypted package
	// NotAllowed: combination of inputs which are not allowed
	newTests := map[string]testData{
		"Nothing": { // expect no crash
			emails:       []string{},
			wantPackages: []*MessagePackage{},
		},
		"Fails": {
			emails: []string{
				"wrongtype@pm.me",
				"multipart@pm.me",
				"noencrypt@pm.me",
				"no-pubkey@pm.me",
				"nosigning@pm.me",

				"html-sign@email.com",
				"mime-plain@email.com",
				"mime-html@email.com",
				"mime@email.com",
				"clear-plain-sign@email.com",

				"eo@gpg.com",
				"inline-html@gpg.com",
				"inline-mixed@gpg.com",
				"mime-plain@gpg.com",
				"mime-html@sgpg.com",
				"no-pubkey@gpg.com",
				"not-signed@gpg.com",
			},
		},

		// one scheme in one package
		"SingleInternalHTML": {
			emails: []string{"none@pm.me", "html@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"none@pm.me": nil,
						"html@pm.me": nil,
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleInternalPlain": {
			emails: []string{"plain@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me": nil,
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},

		"SingleClearHTML": {
			emails: []string{"none@email.com", "html@email.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"html@email.com": nil,
						"none@email.com": nil,
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearPlain": {
			emails: []string{"plain@email.com", "plain-sign@email.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@email.com":      nil,
						"plain-sign@email.com": nil,
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearMIME": {
			emails: []string{"mime-sign@email.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime-sign@email.com": nil,
					},
					Type:             ClearMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: &AlgoKey{"non-empty", "non-empty"},
				},
			},
		},

		"SingleEncyptedPlain": {
			emails: []string{"plain@gpg.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@gpg.com": nil,
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleEncyptedMIME": {
			emails: []string{"mime@gpg.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": nil,
					},
					Type:          PGPMIMEPackage,
					MIMEType:      ContentTypeMultipartMixed,
					EncryptedBody: "non-empty",
				},
			},
		},

		// two schemes combined to one package
		"SingleClearInternalPlain": {
			emails: []string{"plain@email.com", "plain-sign@email.com", "plain@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me":          nil,
						"plain@email.com":      nil,
						"plain-sign@email.com": nil,
					},
					Type:                    InternalPackage | ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearInternalHTML": {
			emails: []string{"none@email.com", "html@email.com", "html@pm.me", "none@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"none@pm.me":     nil,
						"html@pm.me":     nil,
						"html@email.com": nil,
						"none@email.com": nil,
					},
					Type:                    InternalPackage | ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleEncryptedInternalPlain": {
			emails: []string{"plain@gpg.com", "plain@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me":   nil,
						"plain@gpg.com": nil,
					},
					Type:          InternalPackage | PGPInlinePackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleEncryptedClearMIME": {
			emails: []string{"mime@gpg.com", "mime-sign@email.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com":        nil,
						"mime-sign@email.com": nil,
					},
					Type:             ClearMIMEPackage | PGPMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: &AlgoKey{"non-empty", "non-empty"},
				},
			},
		},

		// one scheme separated to multiple packages
		"MultipleInternal": {
			emails: []string{"none@pm.me", "html@pm.me", "plain@pm.me"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me": nil,
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
				{
					Addresses: map[string]*MessageAddress{
						"none@pm.me": nil,
						"html@pm.me": nil,
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},
		"MultipleClear": {
			emails: []string{
				"none@email.com", "html@email.com",
				"plain@email.com", "plain-sign@email.com",
				"mime-sign@email.com",
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime-sign@email.com": nil,
					},
					Type:             ClearMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: &AlgoKey{"non-empty", "non-empty"},
				},
				{
					Addresses: map[string]*MessageAddress{
						"plain@email.com":      nil,
						"plain-sign@email.com": nil,
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
				{
					Addresses: map[string]*MessageAddress{
						"html@email.com": nil,
						"none@email.com": nil,
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"MultipleEncrypted": {
			emails: []string{"plain@gpg.com", "mime@gpg.com"},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": nil,
					},
					Type:          PGPMIMEPackage,
					MIMEType:      ContentTypeMultipartMixed,
					EncryptedBody: "non-empty",
				},
				{
					Addresses: map[string]*MessageAddress{
						"plain@gpg.com": nil,
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},

		"MultipleComboAll": {
			emails: []string{
				"none@pm.me",
				"plain@pm.me",
				"html@pm.me",

				"none@email.com",
				"html@email.com",
				"plain@email.com",
				"plain-sign@email.com",
				"mime-sign@email.com",

				"mime@gpg.com",
				"plain@gpg.com",
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com":        nil,
						"mime-sign@email.com": nil,
					},
					Type:             ClearMIMEPackage | PGPMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: &AlgoKey{"non-empty", "non-empty"},
				},
				{
					Addresses: map[string]*MessageAddress{
						"plain@gpg.com":        nil,
						"plain@email.com":      nil,
						"plain-sign@email.com": nil,
						"plain@pm.me":          nil,
					},
					Type:                    InternalPackage | ClearPackage | PGPInlinePackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
				{
					Addresses: map[string]*MessageAddress{
						"none@pm.me":     nil,
						"html@pm.me":     nil,
						"none@email.com": nil,
						"html@email.com": nil,
					},
					Type:                    InternalPackage | ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        &AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
	}

	for name, test := range newTests {
		test.mimeBody = "Mime body"
		test.plainBody = "Plain body"
		test.richBody = "HTML body"
		test.allRecipients = allRecipients
		test.allAddresses = allAddresses

		test.addRecipients(t)
		test.addAddresses(t)

		t.Run("NoAtt"+name, test.prepareAndCheck)
		test.attKeys = map[string]*crypto.SessionKey{"attID": attKey}
		t.Run("Att"+name, test.prepareAndCheck)
	}
}
