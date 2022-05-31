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
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/require"
)

const (
	testMessageCleartext       = `<div>jeej saas<br></div><div><br></div><div class="protonmail_signature_block"><div>Sent from <a href="https://protonmail.ch">ProtonMail</a>, encrypted email based in Switzerland.<br></div><div><br></div></div>`
	testMessageCleartextLegacy = `<div>flkasjfkjasdklfjasd<br></div><div>fasd<br></div><div>jfasjdfjasd<br></div><div>fj<br></div><div>asdfj<br></div><div>sadjf<br></div><div>sadjf<br></div><div>asjdf<br></div><div>jasd<br></div><div>fj<br></div><div>asdjf<br></div><div>asdjfsad<br></div><div>fasdlkfjasdjfkljsadfljsdfjsdljflkdsjfkljsdlkfjsdlk<br></div><div>jasfd<br></div><div>jsd<br></div><div>jf<br></div><div>sdjfjsdf<br></div><div><br></div><div>djfskjsladf<br></div><div>asd<br></div><div>fja<br></div><div>sdjfajsf<br></div><div>jas<br></div><div>fas<br></div><div>fj<br></div><div>afj<br></div><div>ajf<br></div><div>af<br></div><div>asdfasdfasd<br></div><div>Sent from <a href="https://protonmail.ch">ProtonMail</a>, encrypted email based in Switzerland.<br></div><div>dshfljsadfasdf<br></div><div>as<br></div><div>df<br></div><div>asd<br></div><div>fasd<br></div><div>f<br></div><div>asd<br></div><div>fasdflasdklfjsadlkjf</div><div>asd<br></div><div>fasdlkfjasdlkfjklasdjflkasjdflaslkfasdfjlasjflkasflksdjflkjasdf<br></div><div>asdflkasdjflajsfljaslkflasf<br></div><div>asdfkas<br></div><div>dfjas<br></div><div>djf<br></div><div>asjf<br></div><div>asj<br></div><div>faj<br></div><div>f<br></div><div>afj<br></div><div>sdjaf<br></div><div>jas<br></div><div>sdfj<br></div><div>ajf<br></div><div>aj<br></div><div>ajsdafafdaaf<br></div><div>a<br></div><div>f<br></div><div>lasl;ga<br></div><div>sags<br></div><div>ad<br></div><div>gags<br></div><div>g<br></div><div>ga<br></div><div>a<br></div><div>gg<br></div><div>a<br></div><div>ag<br></div><div>ag<br></div><div>agga.g.ga,ag.ag./ga<br></div><div><br></div><div>dsga<br></div><div>sg<br></div><div><br></div><div>gasga\g\g\g\g\g\n\y\t\r\\r\r\\n\n\n\<br></div><div><br></div><div><br></div><div>sd<br></div><div>asdf<br></div><div>asdf<br></div><div>dsa<br></div><div>fasd<br></div><div>f</div>`
)

const testMessageEncrypted = `-----BEGIN PGP MESSAGE-----
Version: OpenPGP.js v1.2.0
Comment: http://openpgpjs.org

wcBMA0fcZ7XLgmf2AQf+JPulpEOWwmY/Sfze8rBpYvrO2cebSSkjCgapFfXG
CI4PA+rb+WGkn9uBJf3FgEEg76c2ZqGh9zXTyrdHyFLm8ekarvxzgLpvcei/
p18IzcxsWnaM+1uknL4bKUtK3298gIl6xrfc4eVEA8tqUPUkSLSGk7uggjhj
zEYR4zIgMa0c6sMVcZ1Idvy9gGsTIvvcZJ4h1lKVUl8gba+qr1D76RaAf5xS
SBT74q9HhgfEMZwk6hXAp4MYY5h+lIsuhFu5kQ9fhZKU0PWS7ljddv854ZxS
9gHKPBerv4NBjkkCLp9xa2QNjDnu1fNlzlJpfCavp6wDdC83GiT61VRHPE4s
J9LASwFwgOrPmB8Mi867AQM0dddbj4Qe5ghlUcF1XnybkwfHqvQA1QT50d5n
ddFyxwIjvI/Nsn8MTCSnmrWCrjQ7v8JC73NyGxO5k6ZlUnc6BQVie78QJo5a
ftzl5b6nwlCYuXI8R6N/t5MXzrC5GwR8nvjH6kgbUVTLL1hO2Sbgyq5bBKLW
jjylTsZDHUGi4OX7q7eet5/RhKusWdvR0cHEaZAVD6BhTNN0mFBJ5bM1SINI
9gxJVqKJe7j4nJP4PGZBJrokZihhiBS/WEbJdvS54frYajGKjMavB3VhFP6k
qi5aiqGJKOJOV/G8yIwtdtxac3UL34eWo69U39Zx2mNfSXCzSjuafCr1nmAS
4g==
=Uw3B
-----END PGP MESSAGE-----
`

const testMessageEncryptedLegacy = `---BEGIN ENCRYPTED MESSAGE---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---END ENCRYPTED MESSAGE---||---BEGIN ENCRYPTED RANDOM KEY--------BEGIN PGP MESSAGE-----
Version: OpenPGP.js v0.9.0
Comment: http://openpgpjs.org

wcBMA2tjJVxNCRhtAQf/YzkQoUqaqa3NJ/c1apIF/dsl7yJ4GdVrC3/w7lxE
2CO5ioQD4s6QMWP2Y9dOdVl2INwz8eXOds9NS+1nMs4SoMbrpJnAjx8Cthti
1Z/8eWMU023LYahds8BYM0T435K/2tTB5GTA4uTl2y8Xzz2PbptQ4PrUDaII
+egeQQyPA0yuoRDwpaeTiaBYOSa06YYuK5Agr0buQAxRIMCxI2o+fucjoabv
FsQHKGu20U5GlJroSIyIVVkaH3evhNti/AnYX1HuokcGEQNsF5vo4SjWcH23
2P86EIV+w5lUWC1FN9vZCyvbvyuqLHQMtqKVn4GBOkIc3bYQ0jru3a0FG4Cx
bNJ0ASps2+p3Vxe0d+so2iFV92ByQ+0skyCUwCNUlwOV5V5f2fy1ImXk4mXI
cO/bcbqRxx3pG9gkPIh43FoQktTT+tsJ5vS53qfaLGdhCYfkrWjsKu+2P9Xg
+Cr8clh6NTblhfkoAS1gzjA3XgsgEFrtP+OGqwg=
=c5WU
-----END PGP MESSAGE-----
---END ENCRYPTED RANDOM KEY---
`

const testMessageSigned = `-----BEGIN PGP MESSAGE-----
Version: OpenPGP.js v4.5.3
Comment: https://openpgpjs.org

wcBMA0fcZ7XLgmf2AQgAgnHOlcAwVu2AnVfi2fIQHSkTQ0OFnZMJMRR3MJ1q
HtUW8jkSLcurL0Sn/tBFLIIR4YT2tQMzV7cvZzZyBEuZM4OYnDp8xSmoszPh
Gc/nvYG0A0pmKAQkL27v05Dul8oUWA0APT51urghH2Pzm7NdOMtTKIE4LQjS
mBfQ6Cf14uKV0xGS9v2dSFjFxxXEEpMQ+k60NCKRYClN2LVVxf3OKXbuugds
m2GUGn3CuFsiabosIUv4EcdE3aD9HbNo+PIWLJWRJIYJSc5+FWcbwXuIIFgC
XX1s7OV53ceZJnhjCmDE0N2ZOLLAYWED2zRvUa+CAqG+hZgc/3Ia+UmJUVuZ
BNLAugFuRsOVgh3olUIz0vazHhyGG0XIsNqmRm0U9SIfhWkPPHBmU6Xht6Qw
EvLbBfKTYHxX01yQUNgIv4S/TULeQuUjZQfsNYNXXGepS+jiCoIdEgUwpvre
OMFGsypwQXVCFYO/GQdYanMQRTckEexyBY4hGYVrevDM1yG/zGJIdbfI2L+1
1cz76jI8PtzL+S0zcVkevLcjjsHm2Je959uSida9jara7Bymr0y56UdoXoWX
4vZ0kQNo58eEEV0zg7dit4lDvwcuSZMW6K//xNtRQ4QX7/EDtlcYqBJXPwJY
eQSBVeYbeUbZ+PHJdu5gbI85BJNE2dKcS1bdOhEU2lPLYpvmMpPdot9TwnJb
dN3l8yDyhScGvTIZqlxhU7HCM9VHAS0bDqCUoO8EruztUSgjMI+gKC9+xdVU
yrkF7K23UNLWflROMv4cp0LDRB57619Y2w5lY/MG5bS0jSfMWBwnJG2AF28c
2tYKnHw6rpZXvXnlDmEDT8suTzuTGA==
=Sir8
-----END PGP MESSAGE-----
`

const testMessageSigner = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP.js v0.7.1
Comment: http://openpgpjs.org

xsBNBFSI0BMBB/9td6B5RDzVSFTlFzYOS4JxIb5agtNW1rbA4FeLoC47bGLR
8E42IA6aKcO4H0vOZ1lFms0URiKk1DjCMXn3AUErbxqiV5IATRZLwliH6vwy
PI6j5rtGF8dyxYfwmLtoUNkDcPdcFEb4NCdowsN7e8tKU0bcpouZcQhAqawC
9nEdaG/gS5w+2k4hZX2lOKS1EF5SvP48UadlspEK2PLAIp5wB9XsFS9ey2wu
elzkSfDh7KUAlteqFGSMqIgYH62/gaKm+TcckfZeyiMHWFw6sfrcFQ3QOZPq
ahWt0Rn9XM5xBAxx5vW0oceuQ1vpvdfFlM5ix4gn/9w6MhmStaCee8/fABEB
AAHNBlVzZXJJRMLAcgQQAQgAJgUCVIjQHQYLCQgHAwIJEASDR1Fk7GNTBBUI
AgoDFgIBAhsDAh4BAADmhAf/Yt0mCfWqQ25NNGUN14pKKgnPm68zwj1SmMGa
pU7+7ItRpoFNaDwV5QYiQSLC1SvSb1ZeKoY928GPKfqYyJlBpTPL9zC1OHQj
9+2yYauHjYW9JWQM7hst2S2LBcdiQPOs3ybWPaO9yaccV4thxKOCPvyClaS5
b9T4Iv9GEVZQIUvArkwI8hyzIi6skRgxflGheq1O+S1W4Gzt2VtYvo8g8r6W
GzAGMw2nrs2h0+vUr+dLDgIbFCTc5QU99d5jE/e5Hw8iqBxv9tqB1hVATf8T
wC8aU5MTtxtabOiBgG0PsBs6oIwjFqEjpOIza2/AflPZfo7stp6IiwbwvTHo
1NlHoM7ATQRUiNAdAQf/eOLJYxX4lUQUzrNQgASDNE8gJPj7ywcGzySyqr0Y
5rbG57EjtKMIgZrpzJRpSCuRbBjfsltqJ5Q9TBAbPO+oR3rue0LqPKMnmr/q
KsHswBJRfsb/dbktUNmv/f7R9IVyOuvyP6RgdGeloxdGNeWiZSA6AZYI+WGc
xaOvVDPz8thtnML4G4MUhXxxNZ7JzQ0Lfz6mN8CCkblIP5xpcJsyRU7lUsGD
EJGZX0JH/I8bRVN1Xu08uFinIkZyiXRJ5ZGgF3Dns6VbIWmbttY54tBELtk+
5g9pNSl9qiYwiCdwuZrA//NmD3xlZIN8sG4eM7ZUibZ23vEq+bUt1++6Mpba
GQARAQABwsBfBBgBCAATBQJUiNAfCRAEg0dRZOxjUwIbDAAAlpMH/085qZdO
mGRAlbvViUNhF2rtHvCletC48WHGO1ueSh9VTxalkP21YAYLJ4JgJzArJ7tH
lEeiKiHm8YU9KhLe11Yv/o3AiKIAQjJiQluvk+mWdMcddB4fBjL6ttMTRAXe
gHnjtMoamHbSZdeUTUadv05Fl6ivWtpXlODG4V02YvDiGBUbDosdGXEqDtpT
g6MYlj3QMvUiUNQvt7YGMJS8A9iQ9qBNzErgRW8L6CON2RmpQ/wgwP5nwUHz
JjY51d82Vj8bZeI8LdsX41SPoUhyC7kmNYpw9ZRy7NlrCt8dBIOB4/BKEJ2G
ClW54lp9eeOfYTsdTSbn9VaSO0E6m2/Q4Tk=
=WFtr
-----END PGP PUBLIC KEY BLOCK-----`

func TestMessage_IsBodyEncrypted(t *testing.T) {
	r := require.New(t)
	msg := &Message{Body: testMessageEncrypted}
	r.True(msg.IsBodyEncrypted(), "the body should be encrypted")

	msg.Body = testMessageCleartext
	r.True(!msg.IsBodyEncrypted(), "the body should not be encrypted")
}

func TestMessage_Decrypt(t *testing.T) {
	r := require.New(t)
	msg := &Message{Body: testMessageEncrypted}
	dec, err := msg.Decrypt(testPrivateKeyRing)
	r.NoError(err)
	r.Equal(testMessageCleartext, string(dec))
}

func TestMessage_Decrypt_Legacy(t *testing.T) {
	r := require.New(t)
	testPrivateKeyLegacy := readTestFile("testPrivateKeyLegacy", false)

	key, err := crypto.NewKeyFromArmored(testPrivateKeyLegacy)
	r.NoError(err)

	unlockedKey, err := key.Unlock([]byte(testMailboxPasswordLegacy))
	r.NoError(err)

	testPrivateKeyRingLegacy, err := crypto.NewKeyRing(unlockedKey)
	r.NoError(err)

	msg := &Message{Body: testMessageEncryptedLegacy}

	dec, err := msg.Decrypt(testPrivateKeyRingLegacy)
	r.NoError(err)

	r.Equal(testMessageCleartextLegacy, string(dec))
}

func TestMessage_Decrypt_signed(t *testing.T) {
	r := require.New(t)
	msg := &Message{Body: testMessageSigned}
	dec, err := msg.Decrypt(testPrivateKeyRing)
	r.NoError(err)
	r.Equal(testMessageCleartext, string(dec))
}

func TestMessage_Encrypt(t *testing.T) {
	r := require.New(t)

	key, err := crypto.NewKeyFromArmored(testMessageSigner)
	r.NoError(err)

	signer, err := crypto.NewKeyRing(key)
	r.NoError(err)

	msg := &Message{Body: testMessageCleartext}
	r.NoError(msg.Encrypt(testPrivateKeyRing, testPrivateKeyRing))

	dec, err := msg.Decrypt(testPrivateKeyRing)
	r.NoError(err)

	r.Equal(testMessageCleartext, string(dec))
	r.Equal(testIdentity, signer.GetIdentities()[0])
}

func routeLabelMessages(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
	require.NoError(tb, checkMethodAndPath(req, "PUT", "/mail/v4/messages/label"))
	return "messages/label/put_response.json"
}

func TestMessage_LabelMessages_NoPaging(t *testing.T) {
	r := require.New(t)

	// This should be only enough IDs to produce one page.
	testIDs := []string{}
	for i := 0; i < messageIDPageSize-1; i++ {
		testIDs = append(testIDs, fmt.Sprintf("%v", i))
	}

	// There should be enough IDs to produce just one page so the endpoint should be called once.
	finish, c := newTestClientCallbacks(t,
		routeLabelMessages,
	)
	defer finish()

	r.NoError(c.LabelMessages(context.Background(), testIDs, "mylabel"))
}

func TestMessage_LabelMessages_Paging(t *testing.T) {
	r := require.New(t)

	// This should be enough IDs to produce three pages.
	testIDs := []string{}
	for i := 0; i < 3*messageIDPageSize; i++ {
		testIDs = append(testIDs, fmt.Sprintf("%v", i))
	}

	// There should be enough IDs to produce three pages so the endpoint should be called three times.
	finish, c := newTestClientCallbacks(t,
		routeLabelMessages,
		routeLabelMessages,
		routeLabelMessages,
	)
	defer finish()

	r.NoError(c.LabelMessages(context.Background(), testIDs, "mylabel"))
}

// TestClient_GetMessage might look like no actual functionality is tested
// here. But there was case when API was responding with bad payload and it was
// useful to have this to quickly test it.
func TestClient_GetMessage(t *testing.T) {
	r := require.New(t)
	testID := "AeUizgtA3H44qRgcr-HdBApwLiUhlQg5kB81mg_QalWotmQJIHep9OScWIo7Wu9pnYxM4RqQxJnr3BE4kh4y_Q=="

	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(checkMethodAndPath(req, "GET", "/mail/v4/messages/"+testID))

			return "/messages/get_response.json"
		},
	)
	defer finish()

	msg, err := c.GetMessage(context.Background(), testID)
	r.NoError(err)
	r.Equal(testID, msg.ID)
}
