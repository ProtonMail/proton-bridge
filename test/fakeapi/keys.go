// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package fakeapi

import (
	"context"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// publicKey is used from pmapi unit tests.
// For now we need just some key, no need to have some specific one.
const publicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
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

func (api *FakePMAPI) GetPublicKeysForEmail(_ context.Context, email string) (keys []pmapi.PublicKey, internal bool, err error) {
	if err := api.checkAndRecordCall(GET, "/keys?Email="+email, nil); err != nil {
		return nil, false, err
	}
	return []pmapi.PublicKey{{
		PublicKey: publicKey,
	}}, true, nil
}
