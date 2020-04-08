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

package smtp

import (
	"strings"
	"testing"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type mocks struct {
	t             *testing.T
	eventListener *bridge.MockListener
}

func initMocks(t *testing.T) mocks {
	mockCtrl := gomock.NewController(t)
	return mocks{
		t:             t,
		eventListener: bridge.NewMockListener(mockCtrl),
	}
}

type args struct {
	eventListener     listener.Listener
	contactMeta       *ContactMetadata
	apiKeys           []*pmcrypto.KeyRing
	contactKeys       []*pmcrypto.KeyRing
	composeMode       string
	settingsPgpScheme int
	settingsSign      bool
	isInternal        bool
}

type testData struct {
	name            string
	args            args
	wantSendingInfo SendingInfo
	wantErr         bool
}

func (tt *testData) runTest(t *testing.T) {
	t.Run(tt.name, func(t *testing.T) {
		gotSendingInfo, err := generateSendingInfo(tt.args.eventListener, tt.args.contactMeta, tt.args.isInternal, tt.args.composeMode, tt.args.apiKeys, tt.args.contactKeys, tt.args.settingsSign, tt.args.settingsPgpScheme)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, gotSendingInfo, tt.wantSendingInfo)
		}
	})
}

func TestGenerateSendingInfo_WithoutContact(t *testing.T) {
	m := initMocks(t)

	pubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPublicKey))
	if err != nil {
		panic(err)
	}

	tests := []testData{
		{
			name: "internal, PGP_MIME",
			args: args{
				contactMeta:       nil,
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.InternalPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: pubKey,
			},
		},
		{
			name: "internal, PGP_INLINE",
			args: args{
				contactMeta:       nil,
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.InternalPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: pubKey,
			},
		},
		{
			name: "external, PGP_MIME",
			args: args{
				contactMeta:       nil,
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      true,
				Scheme:    pmapi.ClearMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: nil,
			},
		},
		{
			name: "external, PGP_INLINE",
			args: args{
				contactMeta:       nil,
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      true,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypePlainText,
				PublicKey: nil,
			},
		},
		{
			name: "external, PGP_MIME, Unsigned",
			args: args{
				contactMeta:       nil,
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      false,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      false,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: nil,
			},
		},
		{
			name: "internal, error no valid public key",
			args: args{
				eventListener: m.eventListener,
				contactMeta:   nil,
				isInternal:    true,
				apiKeys:       []*pmcrypto.KeyRing{},
				contactKeys:   []*pmcrypto.KeyRing{pubKey},
			},
			wantSendingInfo: SendingInfo{},
			wantErr:         true,
		},
		{
			name: "external, no pinned key but receive one via WKD",
			args: args{
				contactMeta:       nil,
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: pubKey,
			},
		},
	}
	for _, tt := range tests {
		tt.runTest(t)
	}
}

func TestGenerateSendingInfo_Contact_Internal(t *testing.T) {
	m := initMocks(t)

	pubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPublicKey))
	if err != nil {
		panic(err)
	}

	preferredPubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPublicKey))
	if err != nil {
		panic(err)
	}

	differentPubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testDifferentPublicKey))
	if err != nil {
		panic(err)
	}

	m.eventListener.EXPECT().Emit(events.NoActiveKeyForRecipientEvent, "badkey@email.com")

	tests := []testData{
		{
			name: "PGP_MIME, contact wants pgp-mime, no pinned key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.InternalPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: pubKey,
			},
		},
		{
			name: "PGP_MIME, contact wants pgp-mime, pinned key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{pubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.InternalPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: pubKey,
			},
		},
		{
			name: "PGP_MIME, contact wants pgp-mime, pinned key but prefer api key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{preferredPubKey},
				contactKeys:       []*pmcrypto.KeyRing{pubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.InternalPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: preferredPubKey,
			},
		},
		{
			name: "internal, found no active key for recipient",
			args: args{
				eventListener:     m.eventListener,
				contactMeta:       &ContactMetadata{Email: "badkey@email.com", Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        true,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{differentPubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{},
			wantErr:         true,
		},
		{
			name: "external, contact saved, no pinned key but receive one via WKD",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: pubKey,
			},
		},
		{
			name: "external, contact saved, pinned key but receive different one via WKD",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{pubKey},
				contactKeys:       []*pmcrypto.KeyRing{differentPubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: differentPubKey,
			},
		},
	}
	for _, tt := range tests {
		tt.runTest(t)
	}
}

func TestGenerateSendingInfo_Contact_External(t *testing.T) {
	pubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPublicKey))
	if err != nil {
		panic(err)
	}

	expiredPubKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(testExpiredPublicKey))
	if err != nil {
		panic(err)
	}

	tests := []testData{
		{
			name: "PGP_MIME, no pinned key",
			args: args{
				contactMeta:       &ContactMetadata{},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      false,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: nil,
			},
		},
		{
			name: "PGP_MIME, pinned key but it's expired",
			args: args{
				contactMeta:       &ContactMetadata{},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{expiredPubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      false,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: nil,
			},
		},
		{
			name: "PGP_MIME, contact wants pgp-mime, pinned key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-mime"},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{pubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: pubKey,
			},
		},
		{
			name: "PGP_MIME, contact wants pgp-inline, pinned key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true, Scheme: "pgp-inline"},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{pubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPInlinePackage,
				MIMEType:  pmapi.ContentTypePlainText,
				PublicKey: pubKey,
			},
		},
		{
			name: "PGP_MIME, contact wants default scheme, pinned key",
			args: args{
				contactMeta:       &ContactMetadata{Encrypt: true},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{pubKey},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPMIMEPackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   true,
				Sign:      true,
				Scheme:    pmapi.PGPMIMEPackage,
				MIMEType:  pmapi.ContentTypeMultipartMixed,
				PublicKey: pubKey,
			},
		},
		{
			name: "PGP_INLINE, contact wants default scheme, no pinned key",
			args: args{
				contactMeta:       &ContactMetadata{},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      false,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypeHTML,
				PublicKey: nil,
			},
		},
		{
			name: "PGP_INLINE, contact wants plain text, no pinned key",
			args: args{
				contactMeta:       &ContactMetadata{MIMEType: pmapi.ContentTypePlainText},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      false,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypePlainText,
				PublicKey: nil,
			},
		},
		{
			name: "PGP_INLINE, contact sign missing, no pinned key",
			args: args{
				contactMeta:       &ContactMetadata{SignMissing: true},
				isInternal:        false,
				composeMode:       pmapi.ContentTypeHTML,
				apiKeys:           []*pmcrypto.KeyRing{},
				contactKeys:       []*pmcrypto.KeyRing{},
				settingsSign:      true,
				settingsPgpScheme: pmapi.PGPInlinePackage,
			},
			wantSendingInfo: SendingInfo{
				Encrypt:   false,
				Sign:      true,
				Scheme:    pmapi.ClearPackage,
				MIMEType:  pmapi.ContentTypePlainText,
				PublicKey: nil,
			},
		},
	}
	for _, tt := range tests {
		tt.runTest(t)
	}
}

const testPublicKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP.js v0.7.1
Comment: http://openpgpjs.org

xsBNBFRJbc0BCAC0mMLZPDBbtSCWvxwmOfXfJkE2+ssM3ux21LhD/bPiWefE
WSHlCjJ8PqPHy7snSiUuxuj3f9AvXPvg+mjGLBwu1/QsnSP24sl3qD2onl39
vPiLJXUqZs20ZRgnvX70gjkgEzMFBxINiy2MTIG+4RU8QA7y8KzWev0btqKi
MeVa+GLEHhgZ2KPOn4Jv1q4bI9hV0C9NUe2tTXS6/Vv3vbCY7lRR0kbJ65T5
c8CmpqJuASIJNrSXM/Q3NnnsY4kBYH0s5d2FgbASQvzrjuC2rngUg0EoPsrb
DEVRA2/BCJonw7aASiNCrSP92lkZdtYlax/pcoE/mQ4WSwySFmcFT7yFABEB
AAHNBlVzZXJJRMLAcgQQAQgAJgUCVEltzwYLCQgHAwIJED62JZ7fId8kBBUI
AgoDFgIBAhsDAh4BAAD0nQf9EtH9TC0JqSs8q194Zo244jjlJFM3EzxOSULq
0zbywlLORfyoo/O8jU/HIuGz+LT98JDtnltTqfjWgu6pS3ZL2/L4AGUKEoB7
OI6oIdRwzMc61sqI+Qpbzxo7rzufH4CiXZc6cxORUgL550xSCcqnq0q1mds7
h5roKDzxMW6WLiEsc1dN8IQKzC7Ec5wA7U4oNGsJ3TyI8jkIs0IhXrRCd26K
0TW8Xp6GCsfblWXosR13y89WVNgC+xrrJKTZEisc0tRlneIgjcwEUvwfIg2n
9cDUFA/5BsfzTW5IurxqDEziIVP0L44PXjtJrBQaGMPlEbtP5i2oi3OADVX2
XbvsRc7ATQRUSW3PAQgAkPnu5fps5zhOB/e618v/iF3KiogxUeRhA68TbvA+
xnFfTxCx2Vo14aOL0CnaJ8gO5yRSqfomL2O1kMq07N1MGbqucbmc+aSfoElc
+Gd5xBE/w3RcEhKcAaYTi35vG22zlZup4x3ElioyIarOssFEkQgNNyDf5AXZ
jdHLA6qVxeqAb/Ff74+y9HUmLPSsRU9NwFzvK3Jv8C/ubHVLzTYdFgYkc4W1
Uug9Ou08K+/4NEMrwnPFBbZdJAuUjQz2zW2ZiEKiBggiorH2o5N3mYUnWEmU
vqL3EOS8TbWo8UBIW3DDm2JiZR8VrEgvBtc9mVDUj/x+5pR07Fy1D6DjRmAc
9wARAQABwsBfBBgBCAATBQJUSW3SCRA+tiWe3yHfJAIbDAAA/iwH/ik9RKZM
B9Ir0x5mGpKPuqhugwrc3d04m1sOdXJm2NtD4ddzSEvzHwaPNvEvUl5v7FVM
zf6+6mYGWHyNP4+e7RtwYLlRpud6smuGyDSsotUYyumiqP6680ZIeWVQ+a1T
ThNs878mAJy1FhvQFdTmA8XIC616hDFpamQKPlpoO1a0wZnQhrPwT77HDYEE
a+hqY4Jr/a7ui40S+7xYRHKL/7ZAS4/grWllhU3dbNrwSzrOKwrA/U0/9t73
8Ap6JL71YymDeaL4sutcoaahda1pTrMWePtrCltz6uySwbZs7GXoEzjX3EAH
+6qhkUJtzMaE3YEFEoQMGzcDTUEfXCJ3zJw=
=yT9U
-----END PGP PUBLIC KEY BLOCK-----
`

const testDifferentPublicKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF2Ix1EBCACwkUAn/5+sO1feDcS+aQ9BskESOyf1tBS8EDyz4deHFqnzoVCx
pJNvF7jb5J0AFCO/I5Mg7ddJb1udd/Eq+aZKfNYgjlvpdnW2Lo6Y0a5I5sm8R+vW
6EPQGxdgT7QG0VbeekGuy+F4o0KrgvJ4Sl3020q/Vix5B8ovtS6LGB22NWn5FGbL
+ssmq3tr3o2Q2jmHEIMTN4LOk1C4oHCljwrl7UP2MrER/if+czva3dB2jQgto6ia
o0+myIHkIjEKz5q7EGaGn9b7TEWk6+qNFRlKSa3GEFy4DXuQuysb+imjuP8uFxwb
/ib4QoOd/lAkrAVrcUHoWWhtBinsGEBXlG0LABEBAAG0GmphbWVzLXRlc3RAcHJv
dG9ubWFpbC5ibHVliQFUBBMBCAA+FiEEIwbxzW52iRgG0YMKojP3Zu/mCXIFAl2I
x1ECGwMFCQPCZwAFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQojP3Zu/mCXJu
iQf+PiGA0sLEHx0gn2TRoYe7NOn9cnbi+KMLPIFJLGG4mAdVnEVNgaEvMGsNnC14
3FNIVaSdIR5/4ebtplZIlJWb8zxyaNTFkOJexnzwLw2p2cMF78Vsc4sAVLL5Y068
0v6KUzSK2cI1D4kvCyVK57jZL5dURCyISQrekYN/qhQb/TXbbUuznIJURTnLIq6k
v3E6SPB0hKksPgYlQaRocICw7ybbFur7gavyYlyZwD22JSGjwkJBSBi9dj14OD5Q
Egrd7E0qMd6BPzdlV9bctRabyUQLVjWFq8Nw4cC8AW7j7ENq6QIsuM2iKPf9M/HR
5U+Q9hUxcaG/Sv72QI7M4Qc4DrkBDQRdiMdRAQgA7Qufpv+RrZzcxYyfRf4SWZu5
Geo4Zke/AzlkTsw3MgMJHxiSXxEZdU4u/NRQeK53sEQ9J5iIuuzdjLbs5ECT4PjI
G8Lw6LtsCQ6WW9Gc7RUQNsXErIYidfk+v2zsJTHkP9aGkAgEe92bu87SSGXKO1In
w3e04wPjXeZ3ZYw2NovtPFNKVqBrglmN2WMTUXqOXNtcHCn/x5hQfuyo41wTol1m
YrZCiWu+Nxt6nEWQHA3hw0Dp8byCd/9yhIbn21cCZbX2aITYZL4pFbemMGfeteZF
eDVDxAXPFtat9pzgFe8wmF1kDrvnEsjvbb5UjmtlWZr0EWGoBkiioVh4/pyVMwAR
AQABiQE2BBgBCAAgFiEEIwbxzW52iRgG0YMKojP3Zu/mCXIFAl2Ix1ECGwwACgkQ
ojP3Zu/mCXLJZAf9Hbfu7FraFdl2DwYO815XFukMCAIUzhIMrLhUFO1WWg/m44bm
6OZ8NockPl8Mx3CjSG5Kjuk9h5AOG/doOVQL+i8ktQ7VsF4G9tBEgcxjacoGvNZH
VP1gFScmnI4rSfduhHf8JKToTJvK/KOFnko4/2fzM2WH3VLu7qZgT3RufuUn5LLn
C7eju/gf4WQZUtMTJODzs/EaHOkFevrJ7c6IIAUWD12sA6WHEC3l/mQuc9iXlyJw
HyMl6JQldr4XCcdTu73uSvVJ/1IkvLiHPuPP9ma9+FClaUGOmUws7rNQ3ODX52tx
bIYA5I4XbBMze46izlbEAKt6wHhQWTGlSpts0A==
=cOfs
-----END PGP PUBLIC KEY BLOCK-----`

const testExpiredPublicKey = `
-----BEGIN PGP PRIVATE KEY BLOCK-----

xcA4BAAAAAEBAgCgONc0J8rfO6cJw5YTP38x1ze2tAYIO7EcmRCNYwMkXngb
0Qdzg34Q5RW0rNiR56VB6KElPUhePRPVklLFiIvHABEBAAEAAf9qabYMzsz/
/LeRVZSsTgTljmJTdzd2ambUbpi+vt8MXJsbaWh71vjoLMWSXajaKSPDjVU5
waFNt9kLqwGGGLqpAQD5ZdMH2XzTq6GU9Ka69iZs6Pbnzwdz59Vc3i8hXlUj
zQEApHargCTsrtvSrm+hK/pN51/BHAy9lxCAw9f2etx+AeMA/RGrijkFZtYt
jeWdv/usXL3mgHvEcJv63N5zcEvDX5X4W1bND3Rlc3QxIDxhQGIuY29tPsJ7
BBABCAAvBQIAAAABBQMAAAU5BgsJBwgDAgkQzcF99nGrkAkEFQgKAgMWAgEC
GQECGwMCHgEAABAlAfwPehmLZs+gOhOTTaSslqQ50bl/REjmv42Nyr1ZBlQS
DECl1Qu4QyeXin29uEXWiekMpNlZVsEuc8icCw6ABhIZ
=/7PI
-----END PGP PRIVATE KEY BLOCK-----`
