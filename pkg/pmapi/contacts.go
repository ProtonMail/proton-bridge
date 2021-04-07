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
	"errors"
	"net/url"
	"strconv"
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

func (c *client) EncryptAndSignCards(cards []Card) ([]Card, error) {
	var err error
	for i := range cards {
		card := &cards[i]
		if isEncryptedCardType(card.Type) {
			if isSignedCardType(card.Type) {
				if card.Signature, err = c.sign(card.Data); err != nil {
					return nil, err
				}
			}

			if card.Data, err = c.encrypt(card.Data, nil); err != nil {
				return nil, err
			}
		} else if isSignedCardType(card.Type) {
			if card.Signature, err = c.sign(card.Data); err != nil {
				return nil, err
			}
		}
	}
	return cards, nil
}

func (c *client) DecryptAndVerifyCards(cards []Card) ([]Card, error) {
	for i := range cards {
		card := &cards[i]
		if isEncryptedCardType(card.Type) {
			signedCard, err := c.decrypt(card.Data)
			if err != nil {
				return nil, err
			}
			card.Data = signedCard
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

// ====================== READ ===========================

type ContactsListRes struct {
	Res
	Contacts []*Contact
}

// GetContacts gets all contacts.
func (c *client) GetContacts(page int, pageSize int) (contacts []*Contact, err error) {
	v := url.Values{}
	v.Set("Page", strconv.Itoa(page))
	if pageSize > 0 {
		v.Set("PageSize", strconv.Itoa(pageSize))
	}
	req, err := c.NewRequest("GET", "/contacts?"+v.Encode(), nil)

	if err != nil {
		return
	}

	var res ContactsListRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	contacts, err = res.Contacts, res.Err()
	return
}

// GetContactByID gets contact details specified by contact ID.
func (c *client) GetContactByID(id string) (contactDetail Contact, err error) {
	req, err := c.NewRequest("GET", "/contacts/"+id, nil)

	if err != nil {
		return
	}

	type ContactRes struct {
		Res
		Contact Contact
	}
	var res ContactRes

	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	contactDetail, err = res.Contact, res.Err()
	return
}

// GetContactsForExport gets contacts in vCard format, signed and encrypted.
func (c *client) GetContactsForExport(page int, pageSize int) (contacts []Contact, err error) {
	v := url.Values{}
	v.Set("Page", strconv.Itoa(page))
	if pageSize > 0 {
		v.Set("PageSize", strconv.Itoa(pageSize))
	}

	req, err := c.NewRequest("GET", "/contacts/export?"+v.Encode(), nil)

	if err != nil {
		return
	}

	type ContactsDetailsRes struct {
		Res
		Contacts []Contact
	}
	var res ContactsDetailsRes

	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	contacts, err = res.Contacts, res.Err()
	return
}

type ContactsEmailsRes struct {
	Res
	ContactEmails []ContactEmail
	Total         int
}

// GetAllContactsEmails gets all emails from all contacts.
func (c *client) GetAllContactsEmails(page int, pageSize int) (contactsEmails []ContactEmail, err error) {
	v := url.Values{}
	v.Set("Page", strconv.Itoa(page))
	if pageSize > 0 {
		v.Set("PageSize", strconv.Itoa(pageSize))
	}

	req, err := c.NewRequest("GET", "/contacts/emails?"+v.Encode(), nil)
	if err != nil {
		return
	}

	var res ContactsEmailsRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	contactsEmails, err = res.ContactEmails, res.Err()
	return
}

// GetContactEmailByEmail gets all emails from all contacts matching a specified email string.
func (c *client) GetContactEmailByEmail(email string, page int, pageSize int) (contactEmails []ContactEmail, err error) {
	v := url.Values{}
	v.Set("Page", strconv.Itoa(page))
	if pageSize > 0 {
		v.Set("PageSize", strconv.Itoa(pageSize))
	}
	v.Set("Email", email)

	req, err := c.NewRequest("GET", "/contacts/emails?"+v.Encode(), nil)
	if err != nil {
		return
	}

	var res ContactsEmailsRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	contactEmails, err = res.ContactEmails, res.Err()
	return
}

// ============================ CREATE ====================================

type CardsList struct {
	Cards []Card
}

type ContactsCards struct {
	Contacts []CardsList
}

type SingleContactResponse struct {
	Res
	Contact Contact
}

type IndexedContactResponse struct {
	Index    int
	Response SingleContactResponse
}

type AddContactsResponse struct {
	Res
	Responses []IndexedContactResponse
}

type AddContactsReq struct {
	ContactsCards
	Overwrite int
	Groups    int
	Labels    int
}

// AddContacts adds contacts specified by cards. Performs signing and encrypting based on card type.
func (c *client) AddContacts(cards ContactsCards, overwrite int, groups int, labels int) (res *AddContactsResponse, err error) {
	reqBody := AddContactsReq{
		ContactsCards: cards,
		Overwrite:     overwrite,
		Groups:        groups,
		Labels:        labels,
	}

	req, err := c.NewJSONRequest("POST", "/contacts", reqBody)
	if err != nil {
		return
	}

	var addContactsRes AddContactsResponse
	if err = c.DoJSON(req, &addContactsRes); err != nil {
		return
	}

	res, err = &addContactsRes, addContactsRes.Err()
	return
}

// ================================= UPDATE =======================================

type UpdateContactResponse struct {
	Res
	Contact Contact
}

type UpdateContactReq struct {
	Cards []Card
}

// UpdateContact updates contact identified by contact ID. Modified contact is specified by cards.
func (c *client) UpdateContact(id string, cards []Card) (res *UpdateContactResponse, err error) {
	reqBody := UpdateContactReq{
		Cards: cards,
	}
	req, err := c.NewJSONRequest("PUT", "/contacts/"+id, reqBody)
	if err != nil {
		return
	}
	var updateContactRes UpdateContactResponse
	if err = c.DoJSON(req, &updateContactRes); err != nil {
		return
	}

	res, err = &updateContactRes, updateContactRes.Err()
	return
}

type SingleIDResponse struct {
	Res
	ID string
}

type UpdateContactGroupsResponse struct {
	Res
	Response SingleIDResponse
}

func (c *client) AddContactGroups(groupID string, contactEmailIDs []string) (res *UpdateContactGroupsResponse, err error) {
	return c.modifyContactGroups(groupID, addContactGroupsAction, contactEmailIDs)
}

func (c *client) RemoveContactGroups(groupID string, contactEmailIDs []string) (res *UpdateContactGroupsResponse, err error) {
	return c.modifyContactGroups(groupID, removeContactGroupsAction, contactEmailIDs)
}

const (
	removeContactGroupsAction = 0
	addContactGroupsAction    = 1
)

type ModifyContactGroupsReq struct {
	LabelID         string
	Action          int
	ContactEmailIDs []string
}

func (c *client) modifyContactGroups(groupID string, modifyContactGroupsAction int, contactEmailIDs []string) (res *UpdateContactGroupsResponse, err error) {
	reqBody := ModifyContactGroupsReq{
		LabelID:         groupID,
		Action:          modifyContactGroupsAction,
		ContactEmailIDs: contactEmailIDs,
	}
	req, err := c.NewJSONRequest("PUT", "/contacts/group", reqBody)
	if err != nil {
		return
	}
	if err = c.DoJSON(req, &res); err != nil {
		return
	}
	err = res.Err()
	return
}

// ================================= DELETE =======================================

type DeleteReq struct {
	IDs []string
}

// DeleteContacts deletes contacts specified by an array of contact IDs.
func (c *client) DeleteContacts(ids []string) (err error) {
	deleteReq := DeleteReq{
		IDs: ids,
	}

	req, err := c.NewJSONRequest("PUT", "/contacts/delete", deleteReq)
	if err != nil {
		return
	}

	type DeleteContactsRes struct {
		Res
		Responses []struct {
			ID       string
			Response Res
		}
	}
	var res DeleteContactsRes

	if err = c.DoJSON(req, &res); err != nil {
		return
	}
	if err = res.Err(); err != nil {
		return
	}
	return
}

// DeleteAllContacts deletes all contacts.
func (c *client) DeleteAllContacts() (err error) {
	req, err := c.NewRequest("DELETE", "/contacts", nil)
	if err != nil {
		return
	}

	var res Res

	if err = c.DoJSON(req, &res); err != nil {
		return
	}
	if err = res.Err(); err != nil {
		return
	}

	return
}

// ===================== Private utility methods =======================

func isSignedCardType(cardType int) bool {
	return (cardType & CardSigned) == CardSigned
}

func isEncryptedCardType(cardType int) bool {
	return (cardType & CardEncrypted) == CardEncrypted
}
