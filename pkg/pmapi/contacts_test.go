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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	CleartextCard       = 0
	EncryptedCard       = 1
	SignedCard          = 2
	EncryptedSignedCard = 3
)

var testAddContactsReq = AddContactsReq{
	ContactsCards: ContactsCards{
		Contacts: []CardsList{
			{
				Cards: []Card{
					{
						Type: 2,
						Data: `BEGIN:VCARD
VERSION:4.0
FN;TYPE=fn:Bob
item1.EMAIL:bob.tester@protonmail.com
UID:proton-web-cd974706-5cde-0e53-e131-c49c88a92ece
END:VCARD
`,
						Signature: ``,
					},
				},
			},
		},
	},
	Overwrite: 0,
	Groups:    0,
	Labels:    0,
}

var testAddContactsResponseBody = `{
    "Code": 1001,
    "Responses": [
        {
            "Index": 0,
            "Response": {
                "Code": 1000,
                "Contact": {
                    "ID": "EU7qYvPAdgJ-zl53hw_btO1WG8TN2FYh2cTIFq1_T6KqulwgxF8CzPjVk_RBUdEejtLvfynlelVNoZwMK_9X2g==",
                    "Name": "Bob",
                    "UID": "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
                    "Size": 139,
                    "CreateTime": 1517319495,
                    "ModifyTime": 1517319495,
                    "ContactEmails": [
                        {
                            "ID": "VT4NoPeQPk48_vg0CVmk63n5mB6CZn9q-P_DYODhOUemhuzUkgBFGF1MktVArjX5zsVdfVlEBFObvt0_K5NwPg==",
                            "Name": "Bob",
                            "Email": "bob.tester@protonmail.com",
                            "Type": [],
                            "Defaults": 1,
                            "Order": 1,
                            "ContactID": "EU7qYvPAdgJ-zl53hw_btO1WG8TN2FYh2cTIFq1_T6KqulwgxF8CzPjVk_RBUdEejtLvfynlelVNoZwMK_9X2g==",
                            "LabelIDs": []
                        }
                    ],
                    "LabelIDs": []
                }
            }
        }
    ]
}`

var testContactCreated = &AddContactsResponse{
	Res: Res{
		Code:       1001,
		StatusCode: 200,
	},
	Responses: []IndexedContactResponse{
		{
			Index: 0,
			Response: SingleContactResponse{
				Res: Res{
					Code: 1000,
				},
				Contact: Contact{
					ID:         "EU7qYvPAdgJ-zl53hw_btO1WG8TN2FYh2cTIFq1_T6KqulwgxF8CzPjVk_RBUdEejtLvfynlelVNoZwMK_9X2g==",
					Name:       "Bob",
					UID:        "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
					Size:       139,
					CreateTime: 1517319495,
					ModifyTime: 1517319495,
					ContactEmails: []ContactEmail{
						{
							ID:        "VT4NoPeQPk48_vg0CVmk63n5mB6CZn9q-P_DYODhOUemhuzUkgBFGF1MktVArjX5zsVdfVlEBFObvt0_K5NwPg==",
							Name:      "Bob",
							Email:     "bob.tester@protonmail.com",
							Type:      []string{},
							Defaults:  1,
							Order:     1,
							ContactID: "EU7qYvPAdgJ-zl53hw_btO1WG8TN2FYh2cTIFq1_T6KqulwgxF8CzPjVk_RBUdEejtLvfynlelVNoZwMK_9X2g==",
							LabelIDs:  []string{},
						},
					},
					LabelIDs: []string{},
				},
			},
		},
	},
}

var testContactUpdated = &UpdateContactResponse{
	Res: Res{
		Code:       1000,
		StatusCode: 200,
	},
	Contact: Contact{
		ID:         "l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ==",
		Name:       "Bob",
		UID:        "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
		Size:       303,
		CreateTime: 1517416603,
		ModifyTime: 1517416656,
		ContactEmails: []ContactEmail{
			{
				ID:        "14n6vuf1zbeo3zsYzgV471S6xJ9gzl7-VZ8tcOTQq6ifBlNEre0SUdUM7sXh6e2Q_4NhJZaU9c7jLdB1HCV6dA==",
				Name:      "Bob",
				Email:     "bob.changed.tester@protonmail.com",
				Type:      []string{},
				Defaults:  1,
				Order:     1,
				ContactID: "l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ==",
				LabelIDs:  []string{},
			},
		},
		LabelIDs: []string{},
	},
}

func TestContact_AddContact(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "POST", "/contacts"))

		var addContactsReq AddContactsReq
		if err := json.NewDecoder(r.Body).Decode(&addContactsReq); err != nil {
			t.Error("Expecting no error while reading request body, got:", err)
		}
		if !reflect.DeepEqual(testAddContactsReq.ContactsCards, addContactsReq.ContactsCards) {
			t.Errorf("Invalid contacts request: expected %+v but got %+v", testAddContactsReq.ContactsCards, addContactsReq.ContactsCards)
		}

		fmt.Fprint(w, testAddContactsResponseBody)
	}))
	defer s.Close()

	created, err := c.AddContacts(testAddContactsReq.ContactsCards, 0, 0, 0)
	if err != nil {
		t.Fatal("Expected no error while adding contact, got:", err)
	}

	if !reflect.DeepEqual(created, testContactCreated) {
		t.Fatalf("Invalid created contact: expected %+v, got %+v", testContactCreated, created)
	}
}

var testGetContactsResponseBody = `{
    "Code": 1000,
    "Contacts": [
        {
            "ID": "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
            "Name": "Alice",
            "UID": "proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b",
            "Size": 243,
            "CreateTime": 1517395498,
            "ModifyTime": 1517395498,
            "LabelIDs": []
        },
        {
            "ID": "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
            "Name": "Bob",
            "UID": "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
            "Size": 303,
            "CreateTime": 1517394677,
            "ModifyTime": 1517394678,
            "LabelIDs": []
        }
    ],
    "Total": 2
}`

var testGetContacts = []*Contact{

	{
		ID:         "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
		Name:       "Alice",
		UID:        "proton-web-98c8de5e-4536-140b-9ab0-bd8ab6a2050b",
		Size:       243,
		CreateTime: 1517395498,
		ModifyTime: 1517395498,
		LabelIDs:   []string{},
	},
	{
		ID:         "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
		Name:       "Bob",
		UID:        "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
		Size:       303,
		CreateTime: 1517394677,
		ModifyTime: 1517394678,
		LabelIDs:   []string{},
	},
}

func TestContact_GetContacts(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/contacts?Page=0&PageSize=1000"))

		fmt.Fprint(w, testGetContactsResponseBody)
	}))
	defer s.Close()

	contacts, err := c.GetContacts(0, 1000)
	if err != nil {
		t.Fatal("Expected no error while getting contacts, got:", err)
	}

	if !reflect.DeepEqual(contacts, testGetContacts) {
		t.Fatalf("Invalid created contact: expected %+v, got %+v", testGetContacts, contacts)
	}
}

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

func TestContact_GetContactById(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/contacts/s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg=="))

		fmt.Fprint(w, testGetContactByIDResponseBody)
	}))
	defer s.Close()

	contact, err := c.GetContactByID("s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==")
	if err != nil {
		t.Fatal("Expected no error while getting contacts, got:", err)
	}

	if !reflect.DeepEqual(contact, testGetContactByID) {
		t.Fatalf("Invalid got contact: expected %+v, got %+v", testGetContactByID, contact)
	}
}

var testGetContactsForExportResponseBody = `{
    "Code": 1000,
    "Contacts": [
        {
            "ID": "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
            "Cards": [
                {
                    "Type": 2,
                    "Data": "BEGIN:VCARD\nVERSION:4.0\nFN;TYPE=fn:Bob\nitem1.EMAIL:bob.changed.tester@protonmail.com\nUID:proton-web-cd974706-5cde-0e53-e131-c49c88a92ece\nEND:VCARD\n",
                    "Signature": "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZr2CRDMO9BwcW4mpAAAtxwIAFGgPO+xH4PHppffQC1R\nxCp/Bjzaq5rDUE3ZMKVJ1sFqGVlq2bP5CIN4w2XCe/MuZ+z2o87fSEtt2n7i\n0/8Ah35u4czn7t8FZoW8u9WwHPURa8gUbP3fYpVASBY1Bt2fUxJrSUYn5KQp\njJM/DgF99bhIjOTuhx9IN7DFKG647Arq+GJh9M6RJNxkb3CBfcCVUXoIwMB7\nnM/fA1r+mcl8dQam0WKVJgy9aO2XUUR62w1SpqJlXY3z8hKvXjjskzU3DQk5\net07RLVQvhy2nCZePsM+TJzL8OBbTa1aF/p1xPe+HND7t3ZCm9tQOY+UhK5H\nbhPbQY48KGdci1dTcm2HbsQ=\n=iOnV\n-----END PGP SIGNATURE-----\n"
                },
                {
                    "Type": 3,
                    "Data": "-----BEGIN PGP MESSAGE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwcBMA1vYAFKnBP8gAQgABsQWnqZadqrHDN43McGhEYfJjOB66R5HhkQAUavP\nHaAHpJciGxfz6tbztQu4C6kdMA80ElbD8c+bJqalw6ZbT4seoP4TTQLykD1n\n0LuNBlaW4x8kfd8rZzFdckk/dY2PruX6byAjSZslnZlZSwp99AJJbvJtfXRR\nzunKMbDieRkaApGZYT25wT5mz1embpXFesvO4nDkOEQCa0uyti3mNSLhYlf/\ntbaOS3WM9VYM9eB9YRZGzJNxMtTxOsd45tBlGCHnCzWEUnJdqZuYzH2QOky7\nMckXhk6YwyemYi/q7OOgSYEg/0lCs2EK3b//14yPDx8Bj5G7rZrnDgsP+BHj\nu9KaAZb2pSBPQoJ2DY3Y4A2Sg8GjaX5CMO9D6GKJkZSYkXddQgcmw7sVPUS+\n+5JaPXlfxoJOOn9kj9A6LDC6eMhYaLujG1BKcZ16DB0jqfwMnPLJ+bYEdatr\nKMvd9rbdsDwQ/tfk11VvHpiEBCNZjxM2+bdBLl9q2EXaLXi+dz/rJg5C0A9u\nNS2CzCUvg6+jNUzHo/RBfRXvlNV8tw==\n=mE2b\n-----END PGP MESSAGE-----\n",
                    "Signature": "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZr2CRDMO9BwcW4mpAAApucIAD/uwWuV6DOg127XIPG6\n/jluL8jmwyCJX9noL6S8ZVMOymziKSh4/P1QyMPC5SL4lMPEiuaEdyetfBkU\n+5hW3tcZ+ptxmDi59SVYqmXTVewPgeB7t8c5nbzCuVuzA7ZAo8HAXHzFVQDS\nj9fKVGjZzQkmlwdcfnkXHAF0Ejilv9wxOOYgqVDuzm7JXVF3Um7nAgGKTJE5\n5CNnrEjmJGapj96mQFwXzET/kAhNIBw9tL5FAkDlKImdw8C0w9sXdvDu3yVM\ntvUZ5o2rR6ft0SC1byFso49vgJ/syeK6P2pPzltZJbsp4MvmlPUB0/G1XRU+\nI7q4IOWCvs8RD88ADmOty2o=\n=hyZE\n-----END PGP SIGNATURE-----\n"
                }
            ]
        },
        {
            "ID": "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
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
            ]
        }
    ],
    "Total": 2
}`

var testGetContactsForExport = []Contact{
	{
		ID: "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
		Cards: []Card{
			{
				Type:      2,
				Data:      "BEGIN:VCARD\nVERSION:4.0\nFN;TYPE=fn:Bob\nitem1.EMAIL:bob.changed.tester@protonmail.com\nUID:proton-web-cd974706-5cde-0e53-e131-c49c88a92ece\nEND:VCARD\n",
				Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZr2CRDMO9BwcW4mpAAAtxwIAFGgPO+xH4PHppffQC1R\nxCp/Bjzaq5rDUE3ZMKVJ1sFqGVlq2bP5CIN4w2XCe/MuZ+z2o87fSEtt2n7i\n0/8Ah35u4czn7t8FZoW8u9WwHPURa8gUbP3fYpVASBY1Bt2fUxJrSUYn5KQp\njJM/DgF99bhIjOTuhx9IN7DFKG647Arq+GJh9M6RJNxkb3CBfcCVUXoIwMB7\nnM/fA1r+mcl8dQam0WKVJgy9aO2XUUR62w1SpqJlXY3z8hKvXjjskzU3DQk5\net07RLVQvhy2nCZePsM+TJzL8OBbTa1aF/p1xPe+HND7t3ZCm9tQOY+UhK5H\nbhPbQY48KGdci1dTcm2HbsQ=\n=iOnV\n-----END PGP SIGNATURE-----\n",
			},
			{
				Type:      3,
				Data:      "-----BEGIN PGP MESSAGE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwcBMA1vYAFKnBP8gAQgABsQWnqZadqrHDN43McGhEYfJjOB66R5HhkQAUavP\nHaAHpJciGxfz6tbztQu4C6kdMA80ElbD8c+bJqalw6ZbT4seoP4TTQLykD1n\n0LuNBlaW4x8kfd8rZzFdckk/dY2PruX6byAjSZslnZlZSwp99AJJbvJtfXRR\nzunKMbDieRkaApGZYT25wT5mz1embpXFesvO4nDkOEQCa0uyti3mNSLhYlf/\ntbaOS3WM9VYM9eB9YRZGzJNxMtTxOsd45tBlGCHnCzWEUnJdqZuYzH2QOky7\nMckXhk6YwyemYi/q7OOgSYEg/0lCs2EK3b//14yPDx8Bj5G7rZrnDgsP+BHj\nu9KaAZb2pSBPQoJ2DY3Y4A2Sg8GjaX5CMO9D6GKJkZSYkXddQgcmw7sVPUS+\n+5JaPXlfxoJOOn9kj9A6LDC6eMhYaLujG1BKcZ16DB0jqfwMnPLJ+bYEdatr\nKMvd9rbdsDwQ/tfk11VvHpiEBCNZjxM2+bdBLl9q2EXaLXi+dz/rJg5C0A9u\nNS2CzCUvg6+jNUzHo/RBfRXvlNV8tw==\n=mE2b\n-----END PGP MESSAGE-----\n",
				Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwsBcBAEBCAAQBQJacZr2CRDMO9BwcW4mpAAApucIAD/uwWuV6DOg127XIPG6\n/jluL8jmwyCJX9noL6S8ZVMOymziKSh4/P1QyMPC5SL4lMPEiuaEdyetfBkU\n+5hW3tcZ+ptxmDi59SVYqmXTVewPgeB7t8c5nbzCuVuzA7ZAo8HAXHzFVQDS\nj9fKVGjZzQkmlwdcfnkXHAF0Ejilv9wxOOYgqVDuzm7JXVF3Um7nAgGKTJE5\n5CNnrEjmJGapj96mQFwXzET/kAhNIBw9tL5FAkDlKImdw8C0w9sXdvDu3yVM\ntvUZ5o2rR6ft0SC1byFso49vgJ/syeK6P2pPzltZJbsp4MvmlPUB0/G1XRU+\nI7q4IOWCvs8RD88ADmOty2o=\n=hyZE\n-----END PGP SIGNATURE-----\n",
			},
		},
	},
	{
		ID: "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
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
	},
}

func TestContact_GetContactsForExport(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/contacts/export?Page=0&PageSize=1000"))

		fmt.Fprint(w, testGetContactsForExportResponseBody)
	}))
	defer s.Close()

	contacts, err := c.GetContactsForExport(0, 1000)
	if err != nil {
		t.Fatal("Expected no error while getting contacts for export, got:", err)
	}

	if !reflect.DeepEqual(contacts, testGetContactsForExport) {
		t.Fatalf("Invalid contact for export: expected %+v, got %+v", testGetContactsForExport, contacts)
	}
}

var testGetContactsEmailsResponseBody = `{
    "Code": 1000,
    "ContactEmails": [
        {
            "ID": "Hgyz1tG0OiC2v_hMIVOa6juMOAp_recWNzWII7a79Tfwdx08Jy3FJY0_Y_UtFYwbi6mN-Xx1sOI9_GmUGAcwWg==",
            "Name": "Bob",
            "Email": "bob.changed.tester@protonmail.com",
            "Type": [],
            "Defaults": 1,
            "Order": 1,
            "ContactID": "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
            "LabelIDs": []
        },
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
    "Total": 2
}`

var testGetContactsEmails = []ContactEmail{
	{
		ID:        "Hgyz1tG0OiC2v_hMIVOa6juMOAp_recWNzWII7a79Tfwdx08Jy3FJY0_Y_UtFYwbi6mN-Xx1sOI9_GmUGAcwWg==",
		Name:      "Bob",
		Email:     "bob.changed.tester@protonmail.com",
		Type:      []string{},
		Defaults:  1,
		Order:     1,
		ContactID: "c6CWuyEE6mMRApAxvvCO9MQKydTU8Do1iikL__M5MoWWjDEebzChAUx-73qa1jTV54RzFO5p9pLBPsIIgCwpww==",
		LabelIDs:  []string{},
	},
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
}

func TestContact_GetAllContactsEmails(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/contacts/emails?Page=0&PageSize=1000"))

		fmt.Fprint(w, testGetContactsEmailsResponseBody)
	}))
	defer s.Close()

	contactsEmails, err := c.GetAllContactsEmails(0, 1000)
	if err != nil {
		t.Fatal("Expected no error while getting contacts for export, got:", err)
	}

	if !reflect.DeepEqual(contactsEmails, testGetContactsEmails) {
		t.Fatalf("Invalid contact for export: expected %+v, got %+v", testGetContactsEmails, contactsEmails)
	}
}

var testUpdateContactReq = UpdateContactReq{
	Cards: []Card{
		{
			Type: 2,
			Data: `BEGIN:VCARD
VERSION:4.0
FN;TYPE=fn:Bob
item1.EMAIL:bob.changed.tester@protonmail.com
UID:proton-web-cd974706-5cde-0e53-e131-c49c88a92ece
END:VCARD
`,
			Signature: ``,
		},
	},
}

var testUpdateContactResponseBody = `{
    "Code": 1000,
    "Contact": {
        "ID": "l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ==",
        "Name": "Bob",
        "UID": "proton-web-cd974706-5cde-0e53-e131-c49c88a92ece",
        "Size": 303,
        "CreateTime": 1517416603,
        "ModifyTime": 1517416656,
        "ContactEmails": [
            {
                "ID": "14n6vuf1zbeo3zsYzgV471S6xJ9gzl7-VZ8tcOTQq6ifBlNEre0SUdUM7sXh6e2Q_4NhJZaU9c7jLdB1HCV6dA==",
                "Name": "Bob",
                "Email": "bob.changed.tester@protonmail.com",
                "Type": [],
                "Defaults": 1,
                "Order": 1,
                "ContactID": "l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ==",
                "LabelIDs": []
            }
        ],
        "LabelIDs": []
    }
}`

func TestContact_UpdateContact(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "PUT", "/contacts/l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ=="))

		var updateContactReq UpdateContactReq
		if err := json.NewDecoder(r.Body).Decode(&updateContactReq); err != nil {
			t.Error("Expecting no error while reading request body, got:", err)
		}
		if !reflect.DeepEqual(testUpdateContactReq.Cards, updateContactReq.Cards) {
			t.Errorf("Invalid contacts request: expected %+v but got %+v", testUpdateContactReq.Cards, updateContactReq.Cards)
		}

		fmt.Fprint(w, testUpdateContactResponseBody)
	}))
	defer s.Close()

	created, err := c.UpdateContact("l4PrVkmDsIIDba9aln829uwPK0nnyWZHnFtrsyb7CJsYgrD6JTVTuuoaVmaANfO2jIVxzZ2vtbt74rznGjjwFQ==", testUpdateContactReq.Cards)
	if err != nil {
		t.Fatal("Expected no error while updating contact, got:", err)
	}

	if !reflect.DeepEqual(created, testContactUpdated) {
		t.Fatalf("Invalid updated contact: expected\n%+v\ngot\n%+v\n", testContactUpdated, created)
	}
}

var testDeleteContactsReq = DeleteReq{
	IDs: []string{
		"s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
	},
}

var testDeleteContactsResponseBody = `{
    "Code": 1001,
    "Responses": [
        {
            "ID": "s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg==",
            "Response": {
                "Code": 1000
            }
        }
    ]
}`

func TestContact_DeleteContacts(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "PUT", "/contacts/delete"))

		var deleteContactsReq DeleteReq
		if err := json.NewDecoder(r.Body).Decode(&deleteContactsReq); err != nil {
			t.Error("Expecting no error while reading request body, got:", err)
		}
		if !reflect.DeepEqual(testDeleteContactsReq.IDs, deleteContactsReq.IDs) {
			t.Errorf("Invalid delete contacts request: expected %+v but got %+v", deleteContactsReq.IDs, testDeleteContactsReq.IDs)
		}

		fmt.Fprint(w, testDeleteContactsResponseBody)
	}))
	defer s.Close()

	err := c.DeleteContacts([]string{"s_SN9y1q0jczjYCH4zhvfOdHv1QNovKhnJ9bpDcTE0u7WCr2Z-NV9uubHXvOuRozW-HRVam6bQupVYRMC3BCqg=="})
	if err != nil {
		t.Fatal("Expected no error while getting contacts for export, got:", err)
	}
}

var testDeleteAllResponseBody = `{
    "Code": 1000
}`

func TestContact_DeleteAllContacts(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "DELETE", "/contacts"))

		fmt.Fprint(w, testDeleteAllResponseBody)
	}))
	defer s.Close()

	err := c.DeleteAllContacts()
	if err != nil {
		t.Fatal("Expected no error while getting contacts for export, got:", err)
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

func TestClient_Encrypt(t *testing.T) {
	c := newTestClient(NewClientManager(testClientConfig))
	c.kr = testPrivateKeyRing

	cardEncrypted, err := c.EncryptAndSignCards(testCardsCleartext)
	assert.Nil(t, err)

	// Result is always different, so the best way is to test it by decrypting again.
	// Another test for decrypting will help us to be sure it's working.
	cardCleartext, err := c.DecryptAndVerifyCards(cardEncrypted)
	assert.Nil(t, err)
	assert.Equal(t, testCardsCleartext[0].Data, cardCleartext[0].Data)
}

func TestClient_Decrypt(t *testing.T) {
	c := newTestClient(NewClientManager(testClientConfig))
	c.kr = testPrivateKeyRing

	cardCleartext, err := c.DecryptAndVerifyCards(testCardsEncrypted)
	assert.Nil(t, err)
	assert.Equal(t, testCardsCleartext[0].Data, cardCleartext[0].Data)
}
