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
	"reflect"
	"testing"

	r "github.com/stretchr/testify/require"
)

var (
	CleartextCard       = 0
	EncryptedCard       = 1
	SignedCard          = 2
	EncryptedSignedCard = 3
)

var testGetContactByIDResponseBody = `{
    "Code": 1000,
    "Contact": {
        "ID": "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
        "Name": "Alice",
        "UID": "proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b",
        "Size": 243,
        "CreateTime": 1517395498,
        "ModifyTime": 1517395498,
        "Cards": [
            {
                "Type": 3,
                "Data": "-----BEGIN PGP MESSAGE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwcBMA1vYAFKnBP8gAQf/RnOQRpo8DVJHQSJRgckEaUQvdMcADiM4L23diyiS\nQfclby/Ve2WInmvZc2RJ3rWENfeqyDZE6krQT642pKiW09GOIyVIjl+hje9y\nE4HBX0AIAWv7QhhKX6UZcM5dYSFbV3j3QxQB8A4Thng2G6ltotMTlbtcHbhu\n96Lt6ngA1tngXLSF5seyflnoiSQ5gLi2qVzrd95dIP6D4Ottcp929/4hDGmq\nPyxw9dColx6gVd1bmIDSI6ewkET4Grmo6QYqjSvjqLOf0PqHKzqypSFLkI5l\nmmnWKYTQCgl9wX+hq6Qz5E+m/BtbkdeX0YxYUss2e+oSAzJmnfdETErG9U5z\n3NJqAc3sgdwDzfWHBzogAxAbDHiqrF6zMlR5SFvZ6nRU7M2DTOE5dJhf+zOp\n1WSKn5LR46LGyt0m5wJPDjaGyQdPffAO4EULvwhGENe10UxRjY1qcUmjYOtS\nunl/vh3afI9PC1jj+HHJD2VgCA==\n=UpcY\n-----END PGP MESSAGE-----\n",
                "Signature": "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZ4pCRDMO9BwcW4mpAAA6h0H/2+97koXzly5pu9hpbaW\n75d1Q976RjMr5DjAx6tKFtSzznel8YfWgvA6OQmMGdPY8ae7/+3mwCJZYWy/\nXVvUfCSflmYpSIKGfP+Vm1XezWY1W84DGhiFj5n8sdaWisv3bpFwFf1YR3Ae\noBoZ4ufNzaQALRqGPMgXETtXZCtzuL/+0vGSKj5SLECiRcSE4jCPEVRy2bcl\nWJyB9r4VmcjF042OMHxphXoYmTEWvgigyaQFHNORu5cK9EHfHpCG6IcjGbdx\n+9Px5YnDY1ix+YpBKePGSTlLE0u6ow0VTUrdvNjl7IUBaRcfJcIIdgCBOTMw\n1uQ/yeyP46V5AFXFnIKeZeQ=\n=FlOf\n-----END PGP SIGNATURE-----\n"
            },
            {
                "Type": 2,
                "Data": "BEGIN:VCARD\nVERSION:4.0\nFN;TYPE=fn:Alice\nitem1.EMAIL:alice@protonmail.com\nUID:proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b\nEND:VCARD",
                "Signature": "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZ4qCRDMO9BwcW4mpAAA3jUIAJ88mIyO8Yj0+evSFXnK\nNxNdjNgn7t1leY0BWlh1nkK76XrZEPipdw2QU8cOcZzn1Wby2SGfZVkwoPc4\nzAhPT4WKbkFVqXhDry5399kLwGYJCxdEcw/oPyYj+YgpQKMxhTrQq21tbEwr\n7JDRBXgi3Cckh/XsteFHOIiAVnM7BV6zFudipnYxa4uNF0Bf4VbUZx1Mm0Wb\nMJaGsO5reqQUQzDPO5TdSAZ8qGSdjVv7RESgUu5DckcDSsnB987Zbh9uFc22\nfPYmb6zA0cEZh3dAjpDPT7cg8hlvfYBb+kP3sLFyLiIkdEG8Pcagjf0k+l76\nr1IsPlYBx2LJmsJf+WDNlj8=\n=Xn+3\n-----END PGP SIGNATURE-----\n"
            }
        ],
        "ContactEmails": [
            {
                "ID": "4m2sBxLq4McqD0D330Kuy5xG-yyDNXyLEjG5_RYcjy9X-3qHGNP07DNOWLY40TYtUAQr4fAVp8zOcZ_z2o6H-A==",
                "Name": "Alice",
                "Email": "alice@protonmail.com",
                "Type": [],
                "Defaults": 1,
                "Order": 1,
                "ContactID": "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
                "LabelIDs": []
            }
        ],
        "LabelIDs": []
    }
}`

var testGetContactEmailByEmailResponseBody = `{
  "Code": 1000,
  "ContactEmails": [
    {
      "ID": "aefew4323jFv0BhSMw==",
      "Name": "ProtonMail Features",
      "Email": "features@protonmail.black",
      "Type": [
        "work"
      ],
      "Defaults": 1,
      "Order": 1,
      "ContactID": "a29olIjFv0rnXxBhSMw==",
      "LabelIDs": [
        "I6hgx3Ol-d3HYa3E394T_ACXDmTaBub14w=="
      ],
      "CanonicalEmail": "features@protonmail.black",
      "LastUsedTime": 1612546350
    }
  ],
  "Total": 2
}`

var testGetContactByID = Contact{
	ID:         "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
	Name:       "Alice",
	UID:        "proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b",
	Size:       243,
	CreateTime: 1517395498,
	ModifyTime: 1517395498,
	Cards: []Card{
		{
			Type:      3,
			Data:      "-----BEGIN PGP MESSAGE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwcBMA1vYAFKnBP8gAQf/RnOQRpo8DVJHQSJRgckEaUQvdMcADiM4L23diyiS\nQfclby/Ve2WInmvZc2RJ3rWENfeqyDZE6krQT642pKiW09GOIyVIjl+hje9y\nE4HBX0AIAWv7QhhKX6UZcM5dYSFbV3j3QxQB8A4Thng2G6ltotMTlbtcHbhu\n96Lt6ngA1tngXLSF5seyflnoiSQ5gLi2qVzrd95dIP6D4Ottcp929/4hDGmq\nPyxw9dColx6gVd1bmIDSI6ewkET4Grmo6QYqjSvjqLOf0PqHKzqypSFLkI5l\nmmnWKYTQCgl9wX+hq6Qz5E+m/BtbkdeX0YxYUss2e+oSAzJmnfdETErG9U5z\n3NJqAc3sgdwDzfWHBzogAxAbDHiqrF6zMlR5SFvZ6nRU7M2DTOE5dJhf+zOp\n1WSKn5LR46LGyt0m5wJPDjaGyQdPffAO4EULvwhGENe10UxRjY1qcUmjYOtS\nunl/vh3afI9PC1jj+HHJD2VgCA==\n=UpcY\n-----END PGP MESSAGE-----\n",
			Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZ4pCRDMO9BwcW4mpAAA6h0H/2+97koXzly5pu9hpbaW\n75d1Q976RjMr5DjAx6tKFtSzznel8YfWgvA6OQmMGdPY8ae7/+3mwCJZYWy/\nXVvUfCSflmYpSIKGfP+Vm1XezWY1W84DGhiFj5n8sdaWisv3bpFwFf1YR3Ae\noBoZ4ufNzaQALRqGPMgXETtXZCtzuL/+0vGSKj5SLECiRcSE4jCPEVRy2bcl\nWJyB9r4VmcjF042OMHxphXoYmTEWvgigyaQFHNORu5cK9EHfHpCG6IcjGbdx\n+9Px5YnDY1ix+YpBKePGSTlLE0u6ow0VTUrdvNjl7IUBaRcfJcIIdgCBOTMw\n1uQ/yeyP46V5AFXFnIKeZeQ=\n=FlOf\n-----END PGP SIGNATURE-----\n",
		},
		{
			Type:      2,
			Data:      "BEGIN:VCARD\nVERSION:4.0\nFN;TYPE=fn:Alice\nitem1.EMAIL:alice@protonmail.com\nUID:proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b\nEND:VCARD",
			Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZ4qCRDMO9BwcW4mpAAA3jUIAJ88mIyO8Yj0+evSFXnK\nNxNdjNgn7t1leY0BWlh1nkK76XrZEPipdw2QU8cOcZzn1Wby2SGfZVkwoPc4\nzAhPT4WKbkFVqXhDry5399kLwGYJCxdEcw/oPyYj+YgpQKMxhTrQq21tbEwr\n7JDRBXgi3Cckh/XsteFHOIiAVnM7BV6zFudipnYxa4uNF0Bf4VbUZx1Mm0Wb\nMJaGsO5reqQUQzDPO5TdSAZ8qGSdjVv7RESgUu5DckcDSsnB987Zbh9uFc22\nfPYmb6zA0cEZh3dAjpDPT7cg8hlvfYBb+kP3sLFyLiIkdEG8Pcagjf0k+l76\nr1IsPlYBx2LJmsJf+WDNlj8=\n=Xn+3\n-----END PGP SIGNATURE-----\n",
		},
	},
	ContactEmails: []ContactEmail{
		{
			ID:        "4m2sBxLq4McqD0D330Kuy5xG-yyDNXyLEjG5_RYcjy9X-3qHGNP07DNOWLY40TYtUAQr4fAVp8zOcZ_z2o6H-A==",
			Name:      "Alice",
			Email:     "alice@protonmail.com",
			Type:      []string{},
			Defaults:  1,
			Order:     1,
			ContactID: "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
			LabelIDs:  []string{},
		},
	},
	LabelIDs: []string{},
}

var testGetContactEmailByEmail = []ContactEmail{
	{
		ID:        "aefew4323jFv0BhSMw==",
		Name:      "ProtonMail Features",
		Email:     "features@protonmail.black",
		Type:      []string{"work"},
		Defaults:  1,
		Order:     1,
		ContactID: "a29olIjFv0rnXxBhSMw==",
		LabelIDs:  []string{"I6hgx3Ol-d3HYa3E394T_ACXDmTaBub14w=="},
	},
}

func TestContact_GetContactById(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/contacts/v4/s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg=="))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testGetContactByIDResponseBody)
	}))
	defer s.Close()

	contact, err := c.GetContactByID(context.Background(), "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==")
	r.NoError(t, err)

	if !reflect.DeepEqual(contact, testGetContactByID) {
		t.Fatalf("Invalid got contact: expected %+v, got %+v", testGetContactByID, contact)
	}
}

func TestContact_GetContactEmailByEmail(t *testing.T) {
	s, c := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/contacts/v4/emails?Email=someone%40pm.me&Page=1&PageSize=10"))

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testGetContactEmailByEmailResponseBody)
	}))
	defer s.Close()

	contact, err := c.GetContactEmailByEmail(context.Background(), "someone@pm.me", 1, 10)
	r.NoError(t, err)

	if !reflect.DeepEqual(contact, testGetContactEmailByEmail) {
		t.Fatalf("Invalid got contact: expected %+v, got %+v", testGetContactByID, contact)
	}
}

func TestContact_isSignedCardType(t *testing.T) {
	if !isSignedCardType(SignedCard) || !isSignedCardType(EncryptedSignedCard) {
		t.Fatal("isSignedCardType shouldn't return false for signed card types")
	}
	if isSignedCardType(CleartextCard) || isSignedCardType(EncryptedCard) {
		t.Fatal("isSignedCardType shouldn't return true for non-signed card types")
	}
}

func TestContact_isEncryptedCardType(t *testing.T) {
	if !isEncryptedCardType(EncryptedCard) || !isEncryptedCardType(EncryptedSignedCard) {
		t.Fatal("isEncryptedCardType shouldn't return false for encrypted card types")
	}
	if isEncryptedCardType(CleartextCard) || isEncryptedCardType(SignedCard) {
		t.Fatal("isEncryptedCardType shouldn't return true for non-encrypted card types")
	}
}

var testCardsEncrypted = []Card{
	{
		Type:      EncryptedSignedCard,
		Data:      "-----BEGIN PGP MESSAGE-----\nVersion: GopenPGP 0.0.1 (ddacebe0)\nComment: https://gopenpgp.org\n\nwcBMA0fcZ7XLgmf2AQf/fLKA6ZCkDxumpDoUoFQfO86B9LFuqGEJq+voP12C6UXo\nfB2nTy/K4+VosLKYOkU9sW1PZOCL+i00z+zkqUZ6jchbZBpzwy/UCTmpPRw5zrmr\nW6bZCwwgqJSGVWrvcrDA3bW9cn/HHqQqU6jNeXIF+IuhTscRAJVGehJZYWjr1lgB\nToJhg4+//Bgp/Fxzz8Fej/fsokgOlRJ8xcZKYx0rKL/+Il0u2jnd08kJTegpaY+6\nBlsYBzfYq25WkS02iy02wHbt6XD7AxFDi4WDjsM8bryLSm/KNWrejqfDYb/tMAKa\nKNJqK39/EUewzp1gHEXiGmdDEIFTKCHTDTPV84mwf9I1Ae4yoLs+ilYE6sSk7DCh\nPSWjDC8lpKzmw93slsejTG93HJKQPcZ0rLBpv6qPZX6widNYjDE=\n=QFxr\n-----END PGP MESSAGE-----",
		Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GopenPGP 0.0.1 (ddacebe0)\nComment: https://gopenpgp.org\n\nwsBcBAABCgAQBQJdZQ1kCRA+tiWe3yHfJAAA9nMH/0X7pS8TGt6Ox0BewRh0vjfQ\n9LPLwbOiHdj97LNqutZcLlDTfm9SPH82221ZpVILWhB0u2kFeNUGihVbjAqJGYJn\nEk2TELLwn8csYRy9r5JkyUirqrvh7jgl4vs1yt8O/3Yb4ARudOoZr8Yrb4+NVNe0\nCcwQJnH/fJPtF1hbarKwtKtCo3IFwTis4pc8qWJRpBH61z1mO0Yr/LIh85QndhnF\nnZ/3MkWOY0kp2gl4ptqtNUw7z+JJ4LLVdT3ycdVK7GVTZmIG90y5KKxwJvrwbS7/\n8rmPGPQ5diLEMrzuKC2plXT6Pdy0ShtZxie2C3JY86e7ol7xvl0pNqxzOrj424w=\n=AOTG\n-----END PGP SIGNATURE-----",
	},
}

var testCardsCleartext = []Card{
	{
		Type:      EncryptedSignedCard,
		Data:      "data",
		Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GopenPGP 0.0.1 (ddacebe0)\nComment: https://gopenpgp.org\n\nwsBcBAABCgAQBQJdZQ1kCRA+tiWe3yHfJAAA9nMH/0X7pS8TGt6Ox0BewRh0vjfQ\n9LPLwbOiHdj97LNqutZcLlDTfm9SPH82221ZpVILWhB0u2kFeNUGihVbjAqJGYJn\nEk2TELLwn8csYRy9r5JkyUirqrvh7jgl4vs1yt8O/3Yb4ARudOoZr8Yrb4+NVNe0\nCcwQJnH/fJPtF1hbarKwtKtCo3IFwTis4pc8qWJRpBH61z1mO0Yr/LIh85QndhnF\nnZ/3MkWOY0kp2gl4ptqtNUw7z+JJ4LLVdT3ycdVK7GVTZmIG90y5KKxwJvrwbS7/\n8rmPGPQ5diLEMrzuKC2plXT6Pdy0ShtZxie2C3JY86e7ol7xvl0pNqxzOrj424w=\n=AOTG\n-----END PGP SIGNATURE-----",
	},
}

func TestClient_Decrypt(t *testing.T) {
	c := newClient(newManager(Config{}), "")
	c.userKeyRing = testPrivateKeyRing

	cardCleartext, err := c.DecryptAndVerifyCards(testCardsEncrypted)
	r.Nil(t, err)
	r.Equal(t, testCardsCleartext[0].Data, cardCleartext[0].Data)
}
