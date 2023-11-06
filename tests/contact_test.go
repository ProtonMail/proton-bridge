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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"context"
	"errors"
	"os"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/cucumber/godog"
	"github.com/emersion/go-vcard"
)

func (s *scenario) userHasContacts(user string, contacts *godog.Table) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contactList, err := unmarshalTable[Contact](contacts)
			if err != nil {
				return err
			}
			for _, contact := range contactList {
				var settings = proton.ContactSettings{}
				format, err := stringToMimeType(contact.Format)
				if err != nil {
					settings.MIMEType = nil
				} else {
					settings.SetMimeType(format)
				}
				scheme, err := stringToEncryptionScheme(contact.Scheme)
				if err != nil {
					settings.Scheme = nil
				} else {
					settings.SetScheme(scheme)
				}
				sign, err := stringToBool(contact.Sign)
				if err != nil {
					settings.Sign = nil
				} else {
					settings.SetSign(sign)
				}
				encrypt, err := stringToBool(contact.Encrypt)
				if err != nil {
					settings.Encrypt = nil
				} else {
					settings.SetEncrypt(encrypt)
				}
				if err := createContact(ctx, c, contact.Email, contact.Name, addrKR, &settings); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (s *scenario) userHasContactWithName(user, contact, name string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			return createContact(ctx, c, contact, name, addrKR, nil)
		})
	})
}

func (s *scenario) contactOfUserHasNoMessageFormat(email, user string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.MIMEType = nil

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasMessageFormat(email, user, format string) error {
	value, err := stringToMimeType(format)
	if err != nil {
		return err
	}
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.SetMimeType(value)

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasNoEncryptionScheme(email, user string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.Scheme = nil

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasEncryptionScheme(email, user, scheme string) error {
	value := proton.PGPInlineScheme
	switch {
	case scheme == "inline":
		value = proton.PGPInlineScheme
	case scheme == "MIME":
		value = proton.PGPMIMEScheme
	default:
		return errors.New("parameter should either be 'inline' or 'MIME'")
	}
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.SetScheme(value)

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasNoSignature(email, user string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.Sign = nil

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasSignature(email, user, enabled string) error {
	value := true
	switch {
	case enabled == "enabled":
		value = true
	case enabled == "disabled":
		value = false
	default:
		return errors.New("parameter should either be 'enabled' or 'disabled'")
	}
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.SetSign(value)

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasNoEncryption(email, user string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.Encrypt = nil

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasEncryption(email, user, enabled string) error {
	value := true
	switch {
	case enabled == "enabled":
		value = true
	case enabled == "disabled":
		value = false
	default:
		return errors.New("parameter should either be 'enabled' or 'disabled'")
	}
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				settings.SetEncrypt(value)

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func (s *scenario) contactOfUserHasPubKey(email, user string, pubKey *godog.DocString) error {
	return s.addContactKey(email, user, pubKey.Content)
}

func (s *scenario) contactOfUserHasPubKeyFromFile(email, user, file string) error {
	body, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return s.addContactKey(email, user, string(body))
}

func getContact(ctx context.Context, c *proton.Client, email string) (proton.Contact, error) {
	contacts, err := c.GetAllContactEmails(ctx, email)
	if err != nil {
		return proton.Contact{}, err
	}
	if len(contacts) == 0 {
		return proton.Contact{}, errors.New("No contact found with email " + email)
	}
	return c.GetContact(ctx, contacts[0].ContactID)
}

func createContact(ctx context.Context, c *proton.Client, contact, name string, addrKR *crypto.KeyRing, settings *proton.ContactSettings) error {
	card, err := proton.NewCard(addrKR, proton.CardTypeSigned)
	if err != nil {
		return err
	}
	if err := card.Set(addrKR, vcard.FieldUID, &vcard.Field{Value: "proton-legacy-139892c2-f691-4118-8c29-061196013e04", Group: "test"}); err != nil {
		return err
	}

	if err := card.Set(addrKR, vcard.FieldFormattedName, &vcard.Field{Value: name, Group: "test"}); err != nil {
		return err
	}
	if err := card.Set(addrKR, vcard.FieldEmail, &vcard.Field{Value: contact, Group: "test"}); err != nil {
		return err
	}
	res, err := c.CreateContacts(ctx, proton.CreateContactsReq{Contacts: []proton.ContactCards{{Cards: []*proton.Card{card}}}, Overwrite: 1})
	if err != nil {
		return err
	}
	if res[0].Response.Code != proton.SuccessCode {
		return errors.New("APIError " + res[0].Response.Message + " while creating contact")
	}

	if settings != nil {
		ctact, err := getContact(ctx, c, contact)
		if err != nil {
			return err
		}
		for _, card := range ctact.Cards {
			settings, err := ctact.GetSettings(addrKR, contact, card.Type)
			if err != nil {
				return err
			}

			err = ctact.SetSettings(addrKR, contact, card.Type, settings)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *scenario) addContactKey(email, user string, pubKey string) error {
	return s.t.withClient(context.Background(), user, func(ctx context.Context, c *proton.Client) error {
		addrID := s.t.getUserByName(user).getAddrID(s.t.getUserByName(user).getEmails()[0])
		return s.t.withAddrKR(ctx, c, user, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			contact, err := getContact(ctx, c, email)
			if err != nil {
				return err
			}
			for _, card := range contact.Cards {
				settings, err := contact.GetSettings(addrKR, email, card.Type)
				if err != nil {
					return err
				}
				key, err := crypto.NewKeyFromArmored(pubKey)
				if err != nil {
					return err
				}
				settings.AddKey(key)

				err = contact.SetSettings(addrKR, email, card.Type, settings)
				if err != nil {
					return err
				}
			}
			_, err = c.UpdateContact(ctx, contact.ContactMetadata.ID, proton.UpdateContactReq{Cards: contact.Cards})
			return err
		})
	})
}

func stringToMimeType(value string) (rfc822.MIMEType, error) {
	switch {
	case value == "plain":
		return rfc822.TextPlain, nil
	case value == "HTML":
		return rfc822.TextHTML, nil
	}
	return rfc822.TextPlain, errors.New("parameter should either be 'plain' or 'HTML'")
}

func stringToEncryptionScheme(value string) (proton.EncryptionScheme, error) {
	switch {
	case value == "inline":
		return proton.PGPInlineScheme, nil
	case value == "MIME":
		return proton.PGPMIMEScheme, nil
	}
	return proton.PGPInlineScheme, errors.New("parameter should either be 'inline' or 'MIME'")
}

func stringToBool(value string) (bool, error) {
	switch {
	case value == "enabled":
		return true, nil
	case value == "disabled":
		return false, nil
	}
	return false, errors.New("parameter should either be 'enabled' or 'disabled'")
}
