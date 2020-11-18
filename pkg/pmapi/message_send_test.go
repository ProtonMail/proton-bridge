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

package pmapi

import (
	"encoding/base64"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/require"
)

type recipient struct {
	email       string
	sendScheme  int
	pubkey      *crypto.KeyRing
	signature   int
	contentType string
	doEncrypt   bool
	wantError   error
}

type testData struct {
	recipients   []recipient
	wantPackages []*MessagePackage

	attKeys                       map[string]*crypto.SessionKey
	mimeBody, plainBody, richBody string
}

func (td *testData) prepareAndCheck(t *testing.T) {
	r := require.New(t)

	shouldBeEmpty := func(want string) require.ValueAssertionFunc {
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

		r.Equal(len(havePackage.Addresses), len(wantPackage.Addresses))
		for email, wantAddress := range wantPackage.Addresses {
			haveAddress, ok := havePackage.Addresses[email]
			r.True(ok, "pkg %d email %s", i, email)

			r.Equal(wantAddress.Type, haveAddress.Type, "pkg %d email %s", i, email)
			shouldBeEmpty(wantAddress.EncryptedBodyKeyPacket)(t, haveAddress.EncryptedBodyKeyPacket, "pkg %d email %s", i, email)
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
					shouldBeEmpty(wantAttKey)(t, haveAttKey, "pkg %d email %s att %s", i, email, attID)
				}
			}
		}

		r.Equal(wantPackage.Type, havePackage.Type, "pkg %d", i)
		r.Equal(wantPackage.MIMEType, havePackage.MIMEType, "pkg %d", i)

		shouldBeEmpty(wantPackage.EncryptedBody)(t, havePackage.EncryptedBody, "pkg %d", i)

		wantBodyKey := wantPackage.DecryptedBodyKey
		haveBodyKey := havePackage.DecryptedBodyKey

		shouldBeEmpty(wantBodyKey.Algorithm)(t, haveBodyKey.Algorithm, "pkg %d", i)
		shouldBeEmpty(wantBodyKey.Key)(t, haveBodyKey.Key, "pkg %d", i)

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
				shouldBeEmpty(wantAttKey.Key)(t, haveAttKey.Key, "pkg %d att %s", i, attID)
				shouldBeEmpty(wantAttKey.Algorithm)(t, haveAttKey.Algorithm, "pkg %d att %s", i, attID)
			}
		}
	}
}

func TestSendReq(t *testing.T) {
	attKeyB64 := "EvjO/2RIJNn6HdoU6ACqFdZglzJhpjQ/PpjsvL3mB5Q="
	token, err := base64.StdEncoding.DecodeString(attKeyB64)
	require.NoError(t, err)

	attKey := crypto.NewSessionKeyFromToken(token, "aes256")
	attKeyPackets := map[string]string{"attID": "not-empty"}
	attAlgoKeys := map[string]AlgoKey{"attID": {"not-empty", "not-empty"}}

	// NOTE naming
	// Single: there should be one packet
	// Multiple: there should be more than one packet
	// Internal: there should be internal package
	// Clear: there should be non-encrypted package
	// Encrypted: there should be encrypted package
	// NotAllowed: combination of inputs which are not allowed
	tests := map[string]testData{
		"SingleInternalHTML": {
			recipients: []recipient{
				{"html@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"html@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleInternalPlain": {
			recipients: []recipient{
				{"plain@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},
		"InternalNotAllowed": {
			recipients: []recipient{
				{"multipart@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, errMultipartInNonMIME},
				{"noencrypt@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, false, errInternalMustEncrypt},
				{"no-pubkey@pm.me", InternalPackage, nil, SignatureDetached, ContentTypeHTML, true, errMisingPubkey},
				{"nosigning@pm.me", InternalPackage, testPublicKeyRing, SignatureNone, ContentTypeHTML, true, errEncryptMustSign},
			},
		},
		"MultipleInternal": {
			recipients: []recipient{
				{"internal1@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
				{"internal2@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
				{"internal3@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
				{"internal4@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"internal1@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"internal3@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
				{
					Addresses: map[string]*MessageAddress{
						"internal2@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"internal4@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          InternalPackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},

		"SingleClearHTML": {
			recipients: []recipient{
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearPlain": {
			recipients: []recipient{
				{"plain@email.com", ClearPackage, nil, SignatureNone, ContentTypePlainText, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearMIME": {
			recipients: []recipient{
				{"mime@email.com", ClearMIMEPackage, nil, SignatureNone, ContentTypeMultipartMixed, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureNone,
						},
					},
					Type:             ClearMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
			},
		},
		"SingleClearSign": {
			recipients: []recipient{
				{"signed@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"signed@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureDetached,
						},
					},
					Type:             ClearMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
			},
		},
		"ClearNotAllowed": {
			recipients: []recipient{
				{"plain@email.com", ClearPackage, nil, SignatureDetached, ContentTypePlainText, false, errSignMustBeMultipart},
				{"html-1@email.com", ClearPackage, nil, SignatureDetached, ContentTypeHTML, false, errSignMustBeMultipart},
				{"plain@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypePlainText, false, errMIMEMustBeMultipart},
				{"html-@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypeHTML, false, errMIMEMustBeMultipart},
			},
		},
		"MultipleClear": {
			recipients: []recipient{
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
				{"sign@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, false, nil},
				{"mime@email.com", ClearMIMEPackage, nil, SignatureNone, ContentTypeMultipartMixed, false, nil},
				{"plain@email.com", ClearPackage, nil, SignatureNone, ContentTypePlainText, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{ // TODO can this two be combined
						"sign@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureDetached,
						},
						"mime@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureNone,
						},
					},
					Type:             ClearMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
				{
					Addresses: map[string]*MessageAddress{
						"plain@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
				{
					Addresses: map[string]*MessageAddress{
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},

		"SingleEncryptedMIME": {
			recipients: []recipient{
				{"mime@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": {
							Type:                   PGPMIMEPackage,
							Signature:              SignatureDetached,
							EncryptedBodyKeyPacket: "non-empty",
						},
					},
					Type:          PGPMIMEPackage,
					MIMEType:      ContentTypeMultipartMixed,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleEncryptedInlinePlain": {
			recipients: []recipient{
				{"inline-plain@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"inline-plain@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleEncryptedInlineHTML": {
			recipients: []recipient{
				{"inline-html@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"inline-html@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},
		"EncryptedNotAllowed": {
			recipients: []recipient{
				{"inline-mixed@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, errMultipartInNonMIME},
				{"inline-clear@gpg.com", PGPInlinePackage, nil, SignatureDetached, ContentTypePlainText, false, errInlinelMustEncrypt},
				{"mime-plain@gpg.com", PGPMIMEPackage, nil, SignatureDetached, ContentTypePlainText, true, errMIMEMustBeMultipart},
				{"mime-html@gpg.com", PGPMIMEPackage, nil, SignatureDetached, ContentTypeHTML, true, errMIMEMustBeMultipart},
				{"no-pubkey@gpg.com", PGPMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, true, errMisingPubkey},
				{"not-signed@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureNone, ContentTypeMultipartMixed, true, errEncryptMustSign},
			},
		},
		"MultipleEncrypted": {
			recipients: []recipient{
				{"mime@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},
				{"inline-plain@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
				{"inline-html@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": {
							Type:                   PGPMIMEPackage,
							Signature:              SignatureDetached,
							EncryptedBodyKeyPacket: "non-empty",
						},
					},
					Type:          PGPMIMEPackage,
					MIMEType:      ContentTypeMultipartMixed,
					EncryptedBody: "non-empty",
				},
				{
					Addresses: map[string]*MessageAddress{
						"inline-plain@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
				{
					Addresses: map[string]*MessageAddress{
						"inline-html@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},

		"SingleInternalEncryptedHTML": {
			recipients: []recipient{
				{"inline-html@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
				{"internal@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"inline-html@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"internal@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage | InternalPackage,
					MIMEType:      ContentTypeHTML,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleInternalEncryptedPlain": {
			recipients: []recipient{
				{"inline-plain@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
				{"internal@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"inline-plain@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"internal@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:          PGPInlinePackage | InternalPackage,
					MIMEType:      ContentTypePlainText,
					EncryptedBody: "non-empty",
				},
			},
		},
		"SingleInternalClearHTML": {
			recipients: []recipient{
				{"internal@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"internal@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    InternalPackage | ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleInternalClearPlain": {
			recipients: []recipient{
				{"internal@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"internal@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    InternalPackage | ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearEncryptedHTML": {
			recipients: []recipient{
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
				{"inline-html@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"inline-html@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
					},
					Type:                    PGPInlinePackage | ClearPackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearEncryptedPlain": {
			recipients: []recipient{
				{"plain@email.com", ClearPackage, nil, SignatureNone, ContentTypePlainText, false, nil},
				{"inline-plain@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"plain@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
						"inline-plain@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:                    PGPInlinePackage | ClearPackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
		"SingleClearEncryptedMIME": {
			recipients: []recipient{
				{"signed@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, false, nil},
				{"mime@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": {
							Type:                   PGPMIMEPackage,
							Signature:              SignatureDetached,
							EncryptedBodyKeyPacket: "non-empty",
						},
						"signed@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureDetached,
						},
					},
					Type:             ClearMIMEPackage | PGPMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
			},
		},
		"SingleClearEncryptedMIMENoSign": {
			recipients: []recipient{
				{"mime@email.com", ClearMIMEPackage, nil, SignatureNone, ContentTypeMultipartMixed, false, nil},
				{"mime@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{
						"mime@gpg.com": {
							Type:                   PGPMIMEPackage,
							Signature:              SignatureDetached,
							EncryptedBodyKeyPacket: "non-empty",
						},
						"mime@email.com": { // can this be combined ?
							Type:      ClearMIMEPackage,
							Signature: SignatureNone,
						},
					},
					Type:             ClearMIMEPackage | PGPMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
			},
		},
		"MultipleCombo": {
			recipients: []recipient{
				{"mime@email.com", ClearMIMEPackage, nil, SignatureNone, ContentTypeMultipartMixed, false, nil},
				{"signed@email.com", ClearMIMEPackage, nil, SignatureDetached, ContentTypeMultipartMixed, false, nil},
				{"mime@gpg.com", PGPMIMEPackage, testPublicKeyRing, SignatureDetached, ContentTypeMultipartMixed, true, nil},

				{"plain@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},
				{"plain@email.com", ClearPackage, nil, SignatureNone, ContentTypePlainText, false, nil},
				{"inline-plain@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypePlainText, true, nil},

				{"html@pm.me", InternalPackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
				{"html@email.com", ClearPackage, nil, SignatureNone, ContentTypeHTML, false, nil},
				{"inline-html@gpg.com", PGPInlinePackage, testPublicKeyRing, SignatureDetached, ContentTypeHTML, true, nil},
			},
			wantPackages: []*MessagePackage{
				{
					Addresses: map[string]*MessageAddress{ // TODO can this three be combined
						"mime@gpg.com": {
							Type:                   PGPMIMEPackage,
							Signature:              SignatureDetached,
							EncryptedBodyKeyPacket: "non-empty",
						},
						"mime@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureNone,
						},
						"signed@email.com": {
							Type:      ClearMIMEPackage,
							Signature: SignatureDetached,
						},
					},
					Type:             ClearMIMEPackage | PGPMIMEPackage,
					MIMEType:         ContentTypeMultipartMixed,
					EncryptedBody:    "non-empty",
					DecryptedBodyKey: AlgoKey{"non-empty", "non-empty"},
				},
				{
					Addresses: map[string]*MessageAddress{
						"plain@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"plain@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
						"inline-plain@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:                    InternalPackage | ClearPackage | PGPInlinePackage,
					MIMEType:                ContentTypePlainText,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
				{
					Addresses: map[string]*MessageAddress{
						"html@pm.me": {
							Type:                          InternalPackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "not-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
						"html@email.com": {
							Type:      ClearPackage,
							Signature: SignatureNone,
						},
						"inline-html@gpg.com": {
							Type:                          PGPInlinePackage,
							Signature:                     SignatureDetached,
							EncryptedBodyKeyPacket:        "non-empty",
							EncryptedAttachmentKeyPackets: attKeyPackets,
						},
					},
					Type:                    InternalPackage | ClearPackage | PGPInlinePackage,
					MIMEType:                ContentTypeHTML,
					EncryptedBody:           "non-empty",
					DecryptedBodyKey:        AlgoKey{"non-empty", "non-empty"},
					DecryptedAttachmentKeys: attAlgoKeys,
				},
			},
		},
	}

	for name, test := range tests {
		test.mimeBody = "Mime body"
		test.plainBody = "Plain body"
		test.richBody = "HTML body"
		t.Run("NoAtt"+name, test.prepareAndCheck)
		test.attKeys = map[string]*crypto.SessionKey{"attID": attKey}
		t.Run("Att"+name, test.prepareAndCheck)
	}
}
