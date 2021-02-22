// Copyright (c) 2021 Proton Technologies AG
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
	"context"
	"net/http"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	r "github.com/stretchr/testify/require"
)

var testCurrentUser = &User{
	ID:         "MJLke8kWh1BBvG95JBIrZvzpgsZ94hNNgjNHVyhXMiv4g9cn6SgvqiIFR5cigpml2LD_iUk_3DkV29oojTt3eA==",
	Name:       "jason",
	UsedSpace:  96691332,
	Currency:   "USD",
	Role:       2,
	Subscribed: 1,
	Services:   1,
	MaxSpace:   10737418240,
	MaxUpload:  26214400,
	Private:    1,
	Keys:       *loadPMKeys(readTestFile("keyring_userKey_JSON", false)),
}

func routeGetUsers(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
	Ok(tb, checkMethodAndPath(r, "GET", "/users"))
	Ok(tb, isAuthReq(r, testUID, testAccessToken))

	return "users/get_response.json"
}

const testPublicKeysBody = `{
    "Code": 1000,
    "RecipientType": 1,
    "MIMEType": "text/html",
    "Keys": [
	{ "Flags": 3, "PublicKey": "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: OpenPGP.js v0.7.1\nComment: http://openpgpjs.org\n\nxsBNBFSI0BMBB/9td6B5RDzVSFTlFzYOS4JxIb5agtNW1rbA4FeLoC47bGLR\n8E42IA6aKcO4H0vOZ1lFms0URiKk1DjCMXn3AUErbxqiV5IATRZLwliH6vwy\nPI6j5rtGF8dyxYfwmLtoUNkDcPdcFEb4NCdowsN7e8tKU0bcpouZcQhAqawC\n9nEdaG/gS5w+2k4hZX2lOKS1EF5SvP48UadlspEK2PLAIp5wB9XsFS9ey2wu\nelzkSfDh7KUAlteqFGSMqIgYH62/gaKm+TcckfZeyiMHWFw6sfrcFQ3QOZPq\nahWt0Rn9XM5xBAxx5vW0oceuQ1vpvdfFlM5ix4gn/9w6MhmStaCee8/fABEB\nAAHNBlVzZXJJRMLAcgQQAQgAJgUCVIjQHQYLCQgHAwIJEASDR1Fk7GNTBBUI\nAgoDFgIBAhsDAh4BAADmhAf/Yt0mCfWqQ25NNGUN14pKKgnPm68zwj1SmMGa\npU7+7ItRpoFNaDwV5QYiQSLC1SvSb1ZeKoY928GPKfqYyJlBpTPL9zC1OHQj\n9+2yYauHjYW9JWQM7hst2S2LBcdiQPOs3ybWPaO9yaccV4thxKOCPvyClaS5\nb9T4Iv9GEVZQIUvArkwI8hyzIi6skRgxflGheq1O+S1W4Gzt2VtYvo8g8r6W\nGzAGMw2nrs2h0+vUr+dLDgIbFCTc5QU99d5jE/e5Hw8iqBxv9tqB1hVATf8T\nwC8aU5MTtxtabOiBgG0PsBs6oIwjFqEjpOIza2/AflPZfo7stp6IiwbwvTHo\n1NlHoM7ATQRUiNAdAQf/eOLJYxX4lUQUzrNQgASDNE8gJPj7ywcGzySyqr0Y\n5rbG57EjtKMIgZrpzJRpSCuRbBjfsltqJ5Q9TBAbPO+oR3rue0LqPKMnmr/q\nKsHswBJRfsb/dbktUNmv/f7R9IVyOuvyP6RgdGeloxdGNeWiZSA6AZYI+WGc\nxaOvVDPz8thtnML4G4MUhXxxNZ7JzQ0Lfz6mN8CCkblIP5xpcJsyRU7lUsGD\nEJGZX0JH/I8bRVN1Xu08uFinIkZyiXRJ5ZGgF3Dns6VbIWmbttY54tBELtk+\n5g9pNSl9qiYwiCdwuZrA//NmD3xlZIN8sG4eM7ZUibZ23vEq+bUt1++6Mpba\nGQARAQABwsBfBBgBCAATBQJUiNAfCRAEg0dRZOxjUwIbDAAAlpMH/085qZdO\nmGRAlbvViUNhF2rtHvCletC48WHGO1ueSh9VTxalkP21YAYLJ4JgJzArJ7tH\nlEeiKiHm8YU9KhLe11Yv/o3AiKIAQjJiQluvk+mWdMcddB4fBjL6ttMTRAXe\ngHnjtMoamHbSZdeUTUadv05Fl6ivWtpXlODG4V02YvDiGBUbDosdGXEqDtpT\ng6MYlj3QMvUiUNQvt7YGMJS8A9iQ9qBNzErgRW8L6CON2RmpQ/wgwP5nwUHz\nJjY51d82Vj8bZeI8LdsX41SPoUhyC7kmNYpw9ZRy7NlrCt8dBIOB4/BKEJ2G\nClW54lp9eeOfYTsdTSbn9VaSO0E6m2/Q4Tk=\n=WFtr\n-----END PGP PUBLIC KEY BLOCK-----"},
	{ "Flags": 1, "PublicKey": "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: OpenPGP.js v0.7.1\nComment: http://openpgpjs.org\n\nxsBNBFSI0BMBB/9td6B5RDzVSFTlFzYOS4JxIb5agtNW1rbA4FeLoC47bGLR\n8E42IA6aKcO4H0vOZ1lFms0URiKk1DjCMXn3AUErbxqiV5IATRZLwliH6vwy\nPI6j5rtGF8dyxYfwmLtoUNkDcPdcFEb4NCdowsN7e8tKU0bcpouZcQhAqawC\n9nEdaG/gS5w+2k4hZX2lOKS1EF5SvP48UadlspEK2PLAIp5wB9XsFS9ey2wu\nelzkSfDh7KUAlteqFGSMqIgYH62/gaKm+TcckfZeyiMHWFw6sfrcFQ3QOZPq\nahWt0Rn9XM5xBAxx5vW0oceuQ1vpvdfFlM5ix4gn/9w6MhmStaCee8/fABEB\nAAHNBlVzZXJJRMLAcgQQAQgAJgUCVIjQHQYLCQgHAwIJEASDR1Fk7GNTBBUI\nAgoDFgIBAhsDAh4BAADmhAf/Yt0mCfWqQ25NNGUN14pKKgnPm68zwj1SmMGa\npU7+7ItRpoFNaDwV5QYiQSLC1SvSb1ZeKoY928GPKfqYyJlBpTPL9zC1OHQj\n9+2yYauHjYW9JWQM7hst2S2LBcdiQPOs3ybWPaO9yaccV4thxKOCPvyClaS5\nb9T4Iv9GEVZQIUvArkwI8hyzIi6skRgxflGheq1O+S1W4Gzt2VtYvo8g8r6W\nGzAGMw2nrs2h0+vUr+dLDgIbFCTc5QU99d5jE/e5Hw8iqBxv9tqB1hVATf8T\nwC8aU5MTtxtabOiBgG0PsBs6oIwjFqEjpOIza2/AflPZfo7stp6IiwbwvTHo\n1NlHoM7ATQRUiNAdAQf/eOLJYxX4lUQUzrNQgASDNE8gJPj7ywcGzySyqr0Y\n5rbG57EjtKMIgZrpzJRpSCuRbBjfsltqJ5Q9TBAbPO+oR3rue0LqPKMnmr/q\nKsHswBJRfsb/dbktUNmv/f7R9IVyOuvyP6RgdGeloxdGNeWiZSA6AZYI+WGc\nxaOvVDPz8thtnML4G4MUhXxxNZ7JzQ0Lfz6mN8CCkblIP5xpcJsyRU7lUsGD\nEJGZX0JH/I8bRVN1Xu08uFinIkZyiXRJ5ZGgF3Dns6VbIWmbttY54tBELtk+\n5g9pNSl9qiYwiCdwuZrA//NmD3xlZIN8sG4eM7ZUibZ23vEq+bUt1++6Mpba\nGQARAQABwsBfBBgBCAATBQJUiNAfCRAEg0dRZOxjUwIbDAAAlpMH/085qZdO\nmGRAlbvViUNhF2rtHvCletC48WHGO1ueSh9VTxalkP21YAYLJ4JgJzArJ7tH\nlEeiKiHm8YU9KhLe11Yv/o3AiKIAQjJiQluvk+mWdMcddB4fBjL6ttMTRAXe\ngHnjtMoamHbSZdeUTUadv05Fl6ivWtpXlODG4V02YvDiGBUbDosdGXEqDtpT\ng6MYlj3QMvUiUNQvt7YGMJS8A9iQ9qBNzErgRW8L6CON2RmpQ/wgwP5nwUHz\nJjY51d82Vj8bZeI8LdsX41SPoUhyC7kmNYpw9ZRy7NlrCt8dBIOB4/BKEJ2G\nClW54lp9eeOfYTsdTSbn9VaSO0E6m2/Q4Tk=\n=WFtr\n-----END PGP PUBLIC KEY BLOCK-----"}
    ]}`

func TestClient_CurrentUser(t *testing.T) {
	finish, c := newTestClientCallbacks(t,
		routeGetUsers,
		routeGetAddresses,
	)
	defer finish()

	user, err := c.CurrentUser(context.TODO())
	r.Nil(t, err)

	// Ignore KeyRings during the check because they have unexported fields and cannot be compared
	r.True(t, cmp.Equal(user, testCurrentUser, cmpopts.IgnoreTypes(&crypto.Key{})))

	r.Nil(t, c.Unlock(context.TODO(), []byte(testMailboxPassword)))
}
