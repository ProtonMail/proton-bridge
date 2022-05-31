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

package smtp

import (
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferencesBuilder(t *testing.T) {
	testContactKey := loadContactKey(t, testPublicKey)
	testOtherContactKey := loadContactKey(t, testOtherPublicKey)

	tests := []struct { //nolint:maligned
		name string

		contactMeta      *ContactMetadata
		receivedKeys     []pmapi.PublicKey
		isInternal       bool
		mailSettings     pmapi.MailSettings
		composerMIMEType string

		wantEncrypt   bool
		wantSign      bool
		wantScheme    pmapi.PackageFlag
		wantMIMEType  string
		wantPublicKey string
	}{
		{
			name: "internal",

			contactMeta:  &ContactMetadata{},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.InternalPackage,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "internal with contact-specific email format",

			contactMeta:  &ContactMetadata{MIMEType: "text/plain"},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.InternalPackage,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "internal with pinned contact public key",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.InternalPackage,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			// NOTE: Need to figured out how to test that this calls the frontend to check for user confirmation.
			name: "internal with conflicting contact public key",

			contactMeta:  &ContactMetadata{Keys: []string{testOtherContactKey}},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.InternalPackage,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external",

			contactMeta:  &ContactMetadata{},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific email format",

			contactMeta:  &ContactMetadata{MIMEType: "text/plain"},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with global pgp-inline scheme",

			contactMeta:  &ContactMetadata{},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPInlinePackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPInlinePackage,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific pgp-inline scheme overriding global pgp-mime setting",

			contactMeta:  &ContactMetadata{Scheme: pgpInline},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPInlinePackage,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific pgp-mime scheme overriding global pgp-inline setting",

			contactMeta:  &ContactMetadata{Scheme: pgpMIME},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPInlinePackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with additional pinned contact public key",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			// NOTE: Need to figured out how to test that this calls the frontend to check for user confirmation.
			name: "wkd-external with additional conflicting contact public key",

			contactMeta:  &ContactMetadata{Keys: []string{testOtherContactKey}},
			receivedKeys: []pmapi.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external",

			contactMeta:  &ContactMetadata{},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     false,
			wantScheme:   pmapi.ClearPackage,
			wantMIMEType: "text/html",
		},

		{
			name: "external with contact-specific email format",

			contactMeta:  &ContactMetadata{MIMEType: "text/plain"},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     false,
			wantScheme:   pmapi.ClearPackage,
			wantMIMEType: "text/plain",
		},

		{
			name: "external with sign enabled",

			contactMeta:  &ContactMetadata{Sign: true, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     true,
			wantScheme:   pmapi.ClearMIMEPackage,
			wantMIMEType: "multipart/mixed",
		},

		{
			name: "external with contact sign enabled and plain text",

			contactMeta:  &ContactMetadata{MIMEType: "text/plain", Scheme: pgpInline, Sign: true, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     true,
			wantScheme:   pmapi.ClearPackage,
			wantMIMEType: "text/plain",
		},

		{
			name: "external with sign enabled, sending plaintext, should still send as ClearMIME",

			contactMeta:  &ContactMetadata{Sign: true, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/plain"},

			wantEncrypt:  false,
			wantSign:     true,
			wantScheme:   pmapi.ClearMIMEPackage,
			wantMIMEType: "multipart/mixed",
		},

		{
			name: "external with pinned contact public key but no intention to encrypt/sign",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   false,
			wantSign:      false,
			wantScheme:    pmapi.ClearPackage,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}, Encrypt: true, Sign: true, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPMIMEPackage,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed using contact-specific pgp-inline",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}, Encrypt: true, Sign: true, Scheme: pgpInline, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPMIMEPackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPInlinePackage,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed using global pgp-inline",

			contactMeta:  &ContactMetadata{Keys: []string{testContactKey}, Encrypt: true, Sign: true, SignIsSet: true},
			receivedKeys: []pmapi.PublicKey{},
			isInternal:   false,
			mailSettings: pmapi.MailSettings{PGPScheme: pmapi.PGPInlinePackage, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      true,
			wantScheme:    pmapi.PGPInlinePackage,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},
	}

	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.

		t.Run(test.name, func(t *testing.T) {
			b := &sendPreferencesBuilder{}

			require.NoError(t, b.setPGPSettings(test.contactMeta, test.receivedKeys, test.isInternal))
			b.setEncryptionPreferences(test.mailSettings)
			b.setMIMEPreferences(test.composerMIMEType)

			prefs := b.build()

			assert.Equal(t, test.wantEncrypt, prefs.Encrypt)
			assert.Equal(t, test.wantSign, prefs.Sign)
			assert.Equal(t, test.wantScheme, prefs.Scheme)
			assert.Equal(t, test.wantMIMEType, prefs.MIMEType)

			if prefs.PublicKey != nil {
				wantKey, err := crypto.NewKeyFromArmored(test.wantPublicKey)
				require.NoError(t, err)

				haveKey, err := prefs.PublicKey.GetKey(0)
				require.NoError(t, err)

				assert.Equal(t, wantKey.GetFingerprint(), haveKey.GetFingerprint())
			}
		})
	}
}

func loadContactKey(t *testing.T, key string) string {
	ck, err := crypto.NewKeyFromArmored(key)
	require.NoError(t, err)

	pk, err := ck.GetPublicKey()
	require.NoError(t, err)

	return string(pk)
}
