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
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Card struct {
	Type      int
	Data      string
	Signature string
}

const (
	CardEncrypted = 1
	CardSigned    = 2
)

type Contact struct {
	ID         string
	Name       string
	UID        string
	Size       int64
	CreateTime int64
	ModifyTime int64
	LabelIDs   []string

	ContactEmails []ContactEmail
	Cards         []Card
}

type ContactEmail struct {
	ID        string
	Name      string
	Email     string
	Type      []string
	Defaults  int
	Order     int
	ContactID string
	LabelIDs  []string
}

var errVerificationFailed = errors.New("signature verification failed")

// ================= Public utility functions ======================

func (c *client) DecryptAndVerifyCards(cards []Card) ([]Card, error) {
	for i := range cards {
		card := &cards[i]
		if isEncryptedCardType(card.Type) {
			signedCard, err := c.decrypt(card.Data)
			if err != nil {
				return nil, err
			}
			card.Data = string(signedCard)
		}
		if isSignedCardType(card.Type) {
			err := c.verify(card.Data, card.Signature)
			if err != nil {
				return cards, errVerificationFailed
			}
		}
	}
	return cards, nil
}

// GetContactByID gets contact details specified by contact ID.
func (c *client) GetContactByID(ctx context.Context, contactID string) (contactDetail Contact, err error) {
	var res struct {
		Contact Contact
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/contacts/v4/" + contactID)
	}); err != nil {
		return Contact{}, err
	}

	return res.Contact, nil
}

// GetContactEmailByEmail gets all emails from all contacts matching a specified email string.
func (c *client) GetContactEmailByEmail(ctx context.Context, email string, page int, pageSize int) (contactEmails []ContactEmail, err error) {
	var res struct {
		ContactEmails []ContactEmail
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		r = r.SetQueryParams(map[string]string{
			"Email": email,
			"Page":  strconv.Itoa(page),
		})
		if pageSize != 0 {
			r.SetQueryParam("PageSize", strconv.Itoa(pageSize))
		}
		return r.SetResult(&res).Get("/contacts/v4/emails")
	}); err != nil {
		return nil, err
	}

	return res.ContactEmails, nil
}

func isSignedCardType(cardType int) bool {
	return (cardType & CardSigned) == CardSigned
}

func isEncryptedCardType(cardType int) bool {
	return (cardType & CardEncrypted) == CardEncrypted
}
