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

package srp

import (
	"bytes"
	"encoding/base64"
	"math/rand"
	"testing"
)

const (
	testServerEphemeral = "l13IQSVFBEV0ZZREuRQ4ZgP6OpGiIfIjbSDYQG3Yp39FkT2B/k3n1ZhwqrAdy+qvPPFq/le0b7UDtayoX4aOTJihoRvifas8Hr3icd9nAHqd0TUBbkZkT6Iy6UpzmirCXQtEhvGQIdOLuwvy+vZWh24G2ahBM75dAqwkP961EJMh67/I5PA5hJdQZjdPT5luCyVa7BS1d9ZdmuR0/VCjUOdJbYjgtIH7BQoZs+KacjhUN8gybu+fsycvTK3eC+9mCN2Y6GdsuCMuR3pFB0RF9eKae7cA6RbJfF1bjm0nNfWLXzgKguKBOeF3GEAsnCgK68q82/pq9etiUDizUlUBcA=="
	testServerProof     = "ffYFIhnhZJAflFJr9FfXbtdsBLkDGH+TUR5sj98wg0iVHyIhIVT6BeZD8tZA75tYlz7uYIanswweB3bjrGfITXfxERgQysQSoPUB284cX4VQm1IfTB/9LPma618MH8OULNluXVu2eizPWnvIn9VLXCaIX+38Xd6xOjmCQgfkpJy3Sh3ndikjqNCGWiKyvERVJi0nTmpAbHmcdeEp1K++ZRbebRhm2d018o/u4H2gu+MF39Hx12zMzEGNMwkNkgKSEQYlqmj57S6tW9JuB30zVZFnw6Krftg1QfJR6zCT1/J57OGp0A/7X/lC6Xz/I33eJvXOpG9GCRCbNiozFg9IXQ=="

	testClientProof      = "8dQtp6zIeEmu3D93CxPdEiCWiAE86uDmK33EpxyqReMwUrm/bTL+zCkWa/X7QgLNrt2FBAriyROhz5TEONgZq/PqZnBEBym6Rvo708KHu6S4LFdZkVc0+lgi7yQpNhU8bqB0BCqdSWd3Fjd3xbOYgO7/vnFK+p9XQZKwEh2RmGv97XHwoxefoyXK6BB+VVMkELd4vL7vdqBiOBU3ufOlSp+0XBMVltQ4oi5l1y21pzOA9cw5WTPIPMcQHffNFq/rReHYnqbBqiLlSLyw6K0PcVuN3bvr3rVYfdS1CsM/Rv1DzXlBUl39B2j82y6hdyGcTeplGyAnAcu0CimvynKBvQ=="
	testModulus          = "W2z5HBi8RvsfYzZTS7qBaUxxPhsfHJFZpu3Kd6s1JafNrCCH9rfvPLrfuqocxWPgWDH2R8neK7PkNvjxto9TStuY5z7jAzWRvFWN9cQhAKkdWgy0JY6ywVn22+HFpF4cYesHrqFIKUPDMSSIlWjBVmEJZ/MusD44ZT29xcPrOqeZvwtCffKtGAIjLYPZIEbZKnDM1Dm3q2K/xS5h+xdhjnndhsrkwm9U9oyA2wxzSXFL+pdfj2fOdRwuR5nW0J2NFrq3kJjkRmpO/Genq1UW+TEknIWAb6VzJJJA244K/H8cnSx2+nSNZO3bbo6Ys228ruV9A8m6DhxmS+bihN3ttQ=="
	testModulusClearSign = `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

W2z5HBi8RvsfYzZTS7qBaUxxPhsfHJFZpu3Kd6s1JafNrCCH9rfvPLrfuqocxWPgWDH2R8neK7PkNvjxto9TStuY5z7jAzWRvFWN9cQhAKkdWgy0JY6ywVn22+HFpF4cYesHrqFIKUPDMSSIlWjBVmEJZ/MusD44ZT29xcPrOqeZvwtCffKtGAIjLYPZIEbZKnDM1Dm3q2K/xS5h+xdhjnndhsrkwm9U9oyA2wxzSXFL+pdfj2fOdRwuR5nW0J2NFrq3kJjkRmpO/Genq1UW+TEknIWAb6VzJJJA244K/H8cnSx2+nSNZO3bbo6Ys228ruV9A8m6DhxmS+bihN3ttQ==
-----BEGIN PGP SIGNATURE-----
Version: ProtonMail
Comment: https://protonmail.com

wl4EARYIABAFAlwB1j0JEDUFhcTpUY8mAAD8CgEAnsFnF4cF0uSHKkXa1GIa
GO86yMV4zDZEZcDSJo0fgr8A/AlupGN9EdHlsrZLmTA1vhIx+rOgxdEff28N
kvNM7qIK
=q6vu
-----END PGP SIGNATURE-----`
)

func init() {
	// Only for tests, replace the default random reader by something that always
	// return the same thing
	RandReader = rand.New(rand.NewSource(42))
}

func TestReadClearSigned(t *testing.T) {
	cleartext, err := ReadClearSignedMessage(testModulusClearSign)
	if err != nil {
		t.Fatal("Expected no error but have ", err)
	}
	if cleartext != testModulus {
		t.Fatalf("Expected message\n\t'%s'\nbut have\n\t'%s'", testModulus, cleartext)
	}

	lastChar := len(testModulusClearSign)
	wrongSignature := testModulusClearSign[:lastChar-100]
	wrongSignature += "c"
	wrongSignature += testModulusClearSign[lastChar-99:]
	_, err = ReadClearSignedMessage(wrongSignature)
	if err != ErrInvalidSignature {
		t.Fatal("Expected the ErrInvalidSignature but have ", err)
	}

	wrongSignature = testModulusClearSign + "data after modulus"
	_, err = ReadClearSignedMessage(wrongSignature)
	if err != ErrDataAfterModulus {
		t.Fatal("Expected the ErrDataAfterModulus but have ", err)
	}
}

func TestSRPauth(t *testing.T) {
	srp, err := NewSrpAuth(4, "bridgetest", "test", "yKlc5/CvObfoiw==", testModulusClearSign, testServerEphemeral)
	if err != nil {
		t.Fatal("Expected no error but have ", err)
	}

	proofs, err := srp.GenerateSrpProofs(2048)
	if err != nil {
		t.Fatal("Expected no error but have ", err)
	}

	expectedProof, err := base64.StdEncoding.DecodeString(testServerProof)
	if err != nil {
		t.Fatal("Expected no error but have ", err)
	}
	if !bytes.Equal(proofs.ExpectedServerProof, expectedProof) {
		t.Fatalf("Expected server proof\n\t'%s'\nbut have\n\t'%s'",
			testServerProof,
			base64.StdEncoding.EncodeToString(proofs.ExpectedServerProof),
		)
	}

	expectedProof, err = base64.StdEncoding.DecodeString(testClientProof)
	if err != nil {
		t.Fatal("Expected no error but have ", err)
	}
	if !bytes.Equal(proofs.ClientProof, expectedProof) {
		t.Fatalf("Expected client proof\n\t'%s'\nbut have\n\t'%s'",
			testClientProof,
			base64.StdEncoding.EncodeToString(proofs.ClientProof),
		)
	}
}
