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

package smtp

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/ProtonMail/go-vcard"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type ContactMetadata struct {
	Email     string
	Keys      []string
	Scheme    string
	Sign      bool
	SignIsSet bool
	Encrypt   bool
	MIMEType  string
}

const (
	FieldPMScheme   = "X-PM-SCHEME"
	FieldPMEncrypt  = "X-PM-ENCRYPT"
	FieldPMSign     = "X-PM-SIGN"
	FieldPMMIMEType = "X-PM-MIMETYPE"
)

func GetContactMetadataFromVCards(cards []pmapi.Card, email string) (contactMeta *ContactMetadata, err error) {
	for _, card := range cards {
		dec := vcard.NewDecoder(strings.NewReader(card.Data))
		parsedCard, err := dec.Decode()
		if err != nil {
			return nil, err
		}
		group := parsedCard.GetGroupByValue(vcard.FieldEmail, email)
		if len(group) == 0 {
			continue
		}

		keys := []string{}
		for _, key := range parsedCard.GetAllValueByGroup(vcard.FieldKey, group) {
			keybyte, err := base64.StdEncoding.DecodeString(strings.Split(key, "base64,")[1])
			if err != nil {
				return nil, err
			}
			// It would be better to always have correct data on the server, but mistakes
			// can happen -- we had an issue where KEY was included in VCARD, but was empty.
			// It's valid and we need to handle it by not including it in the keys, which would fail later.
			if len(keybyte) > 0 {
				keys = append(keys, string(keybyte))
			}
		}
		scheme := parsedCard.GetValueByGroup(FieldPMScheme, group)
		// Warn: ParseBool treats 1, T, True, true as true and 0, F, Fale, false as false.
		//       However PMEL declares 'true' is true, 'false' is false. every other string is true
		encrypt, _ := strconv.ParseBool(parsedCard.GetValueByGroup(FieldPMEncrypt, group))
		var sign, signIsSet bool
		if len(parsedCard[FieldPMSign]) == 0 {
			signIsSet = false
		} else {
			sign, _ = strconv.ParseBool(parsedCard.GetValueByGroup(FieldPMSign, group))
			signIsSet = true
		}
		mimeType := parsedCard.GetValueByGroup(FieldPMMIMEType, group)
		return &ContactMetadata{
			Email:     email,
			Keys:      keys,
			Scheme:    scheme,
			Sign:      sign,
			SignIsSet: signIsSet,
			Encrypt:   encrypt,
			MIMEType:  mimeType,
		}, nil
	}
	return &ContactMetadata{}, nil
}
