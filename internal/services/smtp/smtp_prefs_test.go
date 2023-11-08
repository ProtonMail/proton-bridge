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

package smtp

import (
	"testing"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferencesBuilder(t *testing.T) {
	testContactKey := loadContactKey(t, testPublicKey)
	testOtherContactKey := loadContactKey(t, testOtherPublicKey)

	tests := []struct { //nolint:maligned
		name string

		contactMeta      *contactSettings
		receivedKeys     []proton.PublicKey
		isInternal       bool
		mailSettings     proton.MailSettings
		composerMIMEType string

		wantEncrypt   bool
		wantSign      proton.SignatureType
		wantScheme    proton.EncryptionScheme
		wantMIMEType  rfc822.MIMEType
		wantPublicKey string
	}{
		{
			name: "internal",

			contactMeta:  &contactSettings{},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.InternalScheme,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "internal with contact-specific email format",

			contactMeta:  &contactSettings{MIMEType: "text/plain"},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.InternalScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "internal with pinned contact public key",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.InternalScheme,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			// NOTE: Need to figured out how to test that this calls the frontend to check for user confirmation.
			name: "internal with conflicting contact public key",

			contactMeta:  &contactSettings{Keys: []string{testOtherContactKey}},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   true,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.InternalScheme,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external",

			contactMeta:  &contactSettings{EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external",

			contactMeta:  &contactSettings{EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific email format",

			contactMeta:  &contactSettings{MIMEType: "text/plain", EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with global pgp-inline scheme",

			contactMeta:  &contactSettings{EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPInlineScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPInlineScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific pgp-inline scheme overriding global pgp-mime setting",

			contactMeta:  &contactSettings{Scheme: pgpInline, EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPInlineScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with contact-specific pgp-mime scheme overriding global pgp-inline setting",

			contactMeta:  &contactSettings{Scheme: pgpMIME, EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPInlineScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external with additional pinned contact public key",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}, EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			// NOTE: Need to figured out how to test that this calls the frontend to check for user confirmation.
			name: "wkd-external with additional conflicting contact public key",

			contactMeta:  &contactSettings{Keys: []string{testOtherContactKey}, EncryptUntrusted: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external-with-encrypt-and-sign-disabled",

			contactMeta:  &contactSettings{EncryptUntrusted: false},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   false,
			wantSign:      proton.NoSignature,
			wantScheme:    proton.ClearScheme,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external-with-encrypt-and-sign-disabled-plain-text",

			contactMeta:  &contactSettings{EncryptUntrusted: false},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/plain"},

			wantEncrypt:   false,
			wantSign:      proton.NoSignature,
			wantScheme:    proton.ClearScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "wkd-external-with-encrypt-disabled-sign-enabled",

			contactMeta:  &contactSettings{EncryptUntrusted: false, Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{{PublicKey: testPublicKey}},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   false,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.ClearMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external",

			contactMeta:  &contactSettings{},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     proton.NoSignature,
			wantScheme:   proton.ClearScheme,
			wantMIMEType: "text/html",
		},

		{
			name: "external with contact-specific email format",

			contactMeta:  &contactSettings{MIMEType: "text/plain"},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     proton.NoSignature,
			wantScheme:   proton.ClearScheme,
			wantMIMEType: "text/plain",
		},

		{
			name: "external with sign enabled",

			contactMeta:  &contactSettings{Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     proton.DetachedSignature,
			wantScheme:   proton.ClearMIMEScheme,
			wantMIMEType: "multipart/mixed",
		},

		{
			name: "external with contact sign enabled and plain text",

			contactMeta:  &contactSettings{MIMEType: "text/plain", Scheme: pgpInline, Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:  false,
			wantSign:     proton.DetachedSignature,
			wantScheme:   proton.ClearScheme,
			wantMIMEType: "text/plain",
		},

		{
			name: "external with sign enabled, sending plaintext, should still send as ClearMIME",

			contactMeta:  &contactSettings{Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/plain"},

			wantEncrypt:  false,
			wantSign:     proton.DetachedSignature,
			wantScheme:   proton.ClearMIMEScheme,
			wantMIMEType: "multipart/mixed",
		},

		{
			name: "external with pinned contact public key but no intention to encrypt/sign",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   false,
			wantSign:      proton.NoSignature,
			wantScheme:    proton.ClearScheme,
			wantMIMEType:  "text/html",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}, Encrypt: true, Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPMIMEScheme,
			wantMIMEType:  "multipart/mixed",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed using contact-specific pgp-inline",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}, Encrypt: true, Sign: true, Scheme: pgpInline, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPMIMEScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPInlineScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},

		{
			name: "external with pinned contact public key, encrypted and signed using global pgp-inline",

			contactMeta:  &contactSettings{Keys: []string{testContactKey}, Encrypt: true, Sign: true, SignIsSet: true},
			receivedKeys: []proton.PublicKey{},
			isInternal:   false,
			mailSettings: proton.MailSettings{PGPScheme: proton.PGPInlineScheme, DraftMIMEType: "text/html"},

			wantEncrypt:   true,
			wantSign:      proton.DetachedSignature,
			wantScheme:    proton.PGPInlineScheme,
			wantMIMEType:  "text/plain",
			wantPublicKey: testPublicKey,
		},
	}

	for _, test := range tests {
		test := test // Avoid using range scope test inside function literal.

		t.Run(test.name, func(t *testing.T) {
			b := &sendPrefsBuilder{}

			require.NoError(t, b.setPGPSettings(test.contactMeta, test.receivedKeys, test.isInternal))
			b.setEncryptionPreferences(test.mailSettings)
			b.setMIMEPreferences(test.composerMIMEType)

			prefs := b.build()

			assert.Equal(t, test.wantEncrypt, prefs.Encrypt)
			assert.Equal(t, test.wantSign, prefs.SignatureType)
			assert.Equal(t, test.wantScheme, prefs.EncryptionScheme)
			assert.Equal(t, test.wantMIMEType, prefs.MIMEType)

			if prefs.PubKey != nil {
				wantKey, err := crypto.NewKeyFromArmored(test.wantPublicKey)
				require.NoError(t, err)

				haveKey, err := prefs.PubKey.GetKey(0)
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

const testPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

xsBNBFRJbc0BCAC0mMLZPDBbtSCWvxwmOfXfJkE2+ssM3ux21LhD/bPiWefEWSHl
CjJ8PqPHy7snSiUuxuj3f9AvXPvg+mjGLBwu1/QsnSP24sl3qD2onl39vPiLJXUq
Zs20ZRgnvX70gjkgEzMFBxINiy2MTIG+4RU8QA7y8KzWev0btqKiMeVa+GLEHhgZ
2KPOn4Jv1q4bI9hV0C9NUe2tTXS6/Vv3vbCY7lRR0kbJ65T5c8CmpqJuASIJNrSX
M/Q3NnnsY4kBYH0s5d2FgbASQvzrjuC2rngUg0EoPsrbDEVRA2/BCJonw7aASiNC
rSP92lkZdtYlax/pcoE/mQ4WSwySFmcFT7yFABEBAAHNBlVzZXJJRMLAcgQQAQgA
JgUCVEltzwYLCQgHAwIJED62JZ7fId8kBBUIAgoDFgIBAhsDAh4BAAD0nQf9EtH9
TC0JqSs8q194Zo244jjlJFM3EzxOSULq0zbywlLORfyoo/O8jU/HIuGz+LT98JDt
nltTqfjWgu6pS3ZL2/L4AGUKEoB7OI6oIdRwzMc61sqI+Qpbzxo7rzufH4CiXZc6
cxORUgL550xSCcqnq0q1mds7h5roKDzxMW6WLiEsc1dN8IQKzC7Ec5wA7U4oNGsJ
3TyI8jkIs0IhXrRCd26K0TW8Xp6GCsfblWXosR13y89WVNgC+xrrJKTZEisc0tRl
neIgjcwEUvwfIg2n9cDUFA/5BsfzTW5IurxqDEziIVP0L44PXjtJrBQaGMPlEbtP
5i2oi3OADVX2XbvsRc7ATQRUSW3PAQgAkPnu5fps5zhOB/e618v/iF3KiogxUeRh
A68TbvA+xnFfTxCx2Vo14aOL0CnaJ8gO5yRSqfomL2O1kMq07N1MGbqucbmc+aSf
oElc+Gd5xBE/w3RcEhKcAaYTi35vG22zlZup4x3ElioyIarOssFEkQgNNyDf5AXZ
jdHLA6qVxeqAb/Ff74+y9HUmLPSsRU9NwFzvK3Jv8C/ubHVLzTYdFgYkc4W1Uug9
Ou08K+/4NEMrwnPFBbZdJAuUjQz2zW2ZiEKiBggiorH2o5N3mYUnWEmUvqL3EOS8
TbWo8UBIW3DDm2JiZR8VrEgvBtc9mVDUj/x+5pR07Fy1D6DjRmAc9wARAQABwsBf
BBgBCAATBQJUSW3SCRA+tiWe3yHfJAIbDAAA/iwH/ik9RKZMB9Ir0x5mGpKPuqhu
gwrc3d04m1sOdXJm2NtD4ddzSEvzHwaPNvEvUl5v7FVMzf6+6mYGWHyNP4+e7Rtw
YLlRpud6smuGyDSsotUYyumiqP6680ZIeWVQ+a1TThNs878mAJy1FhvQFdTmA8XI
C616hDFpamQKPlpoO1a0wZnQhrPwT77HDYEEa+hqY4Jr/a7ui40S+7xYRHKL/7ZA
S4/grWllhU3dbNrwSzrOKwrA/U0/9t738Ap6JL71YymDeaL4sutcoaahda1pTrMW
ePtrCltz6uySwbZs7GXoEzjX3EAH+6qhkUJtzMaE3YEFEoQMGzcDTUEfXCJ3zJw=
=yT9U
-----END PGP PUBLIC KEY BLOCK-----`

const testOtherPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF8Rmj4BCACgXXxRqLsmEUWZGd0f88BteXBfi9zL+9GysOTk4n9EgINLN2PU
5rYSmWvVocO8IAfl/z9zpTJQesQjGe5lHbygUWFmjadox2ZeecZw0PWCSRdAjk6w
Q4UX0JiCo3IuICZk1t53WWRtGnhA2Q21J4b2DJg4T5ZFKgKDzDhWoGF1ZStbI5X1
0rKTGFNHgreV5PqxUjxHVtx3rgT9Mx+13QTffqKR9oaYC6mNs4TNJdhyqfaYxqGw
ElxfdS9Wz6ODXrUNuSHETfgvAmo1Qep7GkefrC1isrmXA2+a+mXzFn4L0FCG073w
Vi/lEw6R/vKfN6QukHPxwoSguow4wTyhRRmfABEBAAG0GVRlc3RUZXN0IDx0ZXN0
dGVzdEBwbS5tZT6JAU4EEwEIADgWIQTsXZU1AxlWCPT02+BKdWAu4Q1jXQUCXxGa
PgIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRBKdWAu4Q1jXQw+B/0ZudN+
W9EqJtL/elm7Qla47zNsFmB+pHObdGoKtp3mNc97CQoW1yQ/i/V0heBFTAioP00g
FgEk1ZUJfO++EtI8esNFdDZqY99826/Cl0FlJwubn/XYxi4XyaGTY1nhhyEJ2HWI
/mZ+Jfm9ojbHSLwO5/AHiQt5t+LPDsKLXZw1BDJTgf1xD6e36CwAZgrPGWDqCXJ9
BjlQn5hje7p0F8vYWBnnfSPkMHwibz9FlFqDh5v3XTgGpFIWDVkPVgAs8erM9AM2
TjdpGcdW8xfcymo3j/o2QUBGYGJwPTsGEO5IkFRre9c/3REa7MKIi17Y479ub0A6
2J3xgnqgI4sxmgmOuQENBF8Rmj4BCADX3BamNZsjC3I0knVIwjbz//1r8WOfNwGh
gg5LsvpfLkrsNUZy+deSwb+hS9Auyr1xsMmtVyiTPGUXTjU4uUzY2zyTYWgYfSEi
CojlXmYYLsjyPzR7KhVP6QIYZqYkOQXaCQDRlprRoFIEe4FzTCuqDHatJNwSesGy
5pPJrjiAeb47m9KaoEIacoe9D3w1z4FCKN3A8cjiWT8NRfhYTBoE/T34oXVUj8l+
jLIgVUQgGoBos160Z1Cnxd2PKWFVh/Br3QtIPTbNVDWhh5T1+N2ypbwsXCawy6fj
cbOaTLz/vF9g+RJKC0MtxdL5qUtv3d3Zn07Sg+9H6wjsboAdAvirABEBAAGJATYE
GAEIACAWIQTsXZU1AxlWCPT02+BKdWAu4Q1jXQUCXxGaPgIbDAAKCRBKdWAu4Q1j
Xc4WB/9+aTGMMTlIdAFs9rf0i7i83pUOOxuLl34YQ0t5WGsjteQ4IK+gfuFvp37W
ktv98ShOxAexbfqzGyGcYLLgaCxCbbB85fvSeX0xK/C2UbiH3Gv1z8GTelailCxt
vyx642TwpcLXW1obHaHTSIi5L35Tce9gbug9sKCRSlAH76dANYBbMLa2Bl0LSrF8
mcie9jJaPRXGOeHOyZmPZwwGhVYgadjptWqXnFz3ua8vxgqG0sefWF23F36iVz2q
UjxSE+nKLaPFLlEDLgxG4SwHkcR9fi7zaQVnXg4rEjr0uz5MSUqZC4MNB4rkhU3g
/rUMQyZupw+xJ+ayQNVBEtYZd/9u
=TNX4
-----END PGP PUBLIC KEY BLOCK-----`
