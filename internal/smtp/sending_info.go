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

package smtp

import (
	"errors"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/algo"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

const (
	pgpInline = "pgp-inline"
	pgpMime   = "pgp-mime"
)

type SendingInfo struct {
	Encrypt   bool
	Sign      bool
	Scheme    int
	MIMEType  string
	PublicKey *pmcrypto.KeyRing
}

func generateSendingInfo(
	eventListener listener.Listener,
	contactMeta *ContactMetadata,
	isInternal bool,
	composeMode string,
	apiKeys,
	contactKeys []*pmcrypto.KeyRing,
	settingsSign bool,
	settingsPgpScheme int) (sendingInfo SendingInfo, err error) {
	contactKeys, err = pmcrypto.FilterExpiredKeys(contactKeys)
	if err != nil {
		return
	}

	if isInternal {
		sendingInfo, err = generateInternalSendingInfo(eventListener, contactMeta, composeMode, apiKeys, contactKeys, settingsSign, settingsPgpScheme)
	} else {
		sendingInfo, err = generateExternalSendingInfo(contactMeta, composeMode, apiKeys, contactKeys, settingsSign, settingsPgpScheme)
	}

	if (sendingInfo.Scheme == pmapi.PGPInlinePackage || sendingInfo.Scheme == pmapi.PGPMIMEPackage) && sendingInfo.PublicKey == nil {
		return sendingInfo, errors.New("public key nil during attempt to encrypt")
	}

	return
}

func generateInternalSendingInfo(
	eventListener listener.Listener,
	contactMeta *ContactMetadata,
	composeMode string,
	apiKeys,
	contactKeys []*pmcrypto.KeyRing,
	settingsSign bool, //nolint[unparam]
	settingsPgpScheme int) (sendingInfo SendingInfo, err error) { //nolint[unparam]
	// If sending internally, there should always be a public key; if not, there's an error.
	if len(apiKeys) == 0 {
		err = errors.New("no valid public keys found for contact")
		return
	}

	// The default settings, unless overridden by presence of a saved contact.
	sendingInfo = SendingInfo{
		Encrypt:   true,
		Sign:      true,
		Scheme:    pmapi.InternalPackage,
		MIMEType:  composeMode,
		PublicKey: apiKeys[0],
	}

	// If there is no saved contact, our work here is done.
	if contactMeta == nil {
		return
	}

	// If contact has a pinned key, prefer that over the api key (if it's not expired).
	checkedContactKeys, err := checkContactKeysAgainstAPI(contactKeys, apiKeys)
	if err != nil {
		return
	}

	// If we find no matching keys with the api but the contact still has pinned keys
	// it means the pinned keys are out of date (e.g. the contact has since changed their protonmail
	// keys and so the keys returned via the api don't match the keys pinned in the contact).
	if len(checkedContactKeys) == 0 && len(contactKeys) != 0 {
		eventListener.Emit(events.NoActiveKeyForRecipientEvent, contactMeta.Email)
		return sendingInfo, errors.New("found no active key for recipient " + contactMeta.Email + ", please check contact settings")
	}

	if len(checkedContactKeys) > 0 {
		sendingInfo.PublicKey = checkedContactKeys[0]
	}

	// If contact has a saved mime type preference, prefer that over the default.
	if len(contactMeta.MIMEType) > 0 {
		sendingInfo.MIMEType = contactMeta.MIMEType
	}

	return sendingInfo, nil
}

func generateExternalSendingInfo(
	contactMeta *ContactMetadata,
	composeMode string,
	apiKeys,
	contactKeys []*pmcrypto.KeyRing,
	settingsSign bool,
	settingsPgpScheme int) (sendingInfo SendingInfo, err error) {
	// The default settings, unless overridden by presence of a saved contact.
	sendingInfo = SendingInfo{
		Encrypt:   false,
		Sign:      settingsSign,
		PublicKey: nil,
	}

	if contactMeta != nil && len(contactKeys) > 0 {
		// If the contact has a key, use it. And if the contact metadata says to encryt, do so.
		sendingInfo.PublicKey = contactKeys[0]
		sendingInfo.Encrypt = contactMeta.Encrypt
	} else if len(apiKeys) > 0 {
		// If the api returned a key (via WKD), use it. In this case we always encrypt.
		sendingInfo.PublicKey = apiKeys[0]
		sendingInfo.Encrypt = true
	}

	// - If we are encrypting, we always sign
	// - else if the contact has a preference, we follow that
	// - otherwise, we fall back to the mailbox default signing settings
	if sendingInfo.Encrypt { //nolint[gocritic]
		sendingInfo.Sign = true
	} else if contactMeta != nil && !contactMeta.SignMissing {
		sendingInfo.Sign = contactMeta.Sign
	} else {
		sendingInfo.Sign = settingsSign
	}

	sendingInfo.Scheme, sendingInfo.MIMEType, err = schemeAndMIME(contactMeta,
		settingsPgpScheme,
		composeMode,
		sendingInfo.Encrypt,
		sendingInfo.Sign)

	return sendingInfo, err
}

func schemeAndMIME(contact *ContactMetadata, settingsScheme int, settingsMIMEType string, encrypted, signed bool) (scheme int, mime string, err error) {
	if encrypted && signed {
		// Prefer contact settings.
		if contact != nil && contact.Scheme == pgpInline {
			return pmapi.PGPInlinePackage, pmapi.ContentTypePlainText, nil
		} else if contact != nil && contact.Scheme == pgpMime {
			return pmapi.PGPMIMEPackage, pmapi.ContentTypeMultipartMixed, nil
		}

		// If no contact settings, follow mailbox defaults.
		scheme = settingsScheme
		if scheme == pmapi.PGPMIMEPackage {
			return scheme, pmapi.ContentTypeMultipartMixed, nil
		} else if scheme == pmapi.PGPInlinePackage {
			return scheme, pmapi.ContentTypePlainText, nil
		}
	}

	if !encrypted && signed {
		// Prefer contact settings but send unencrypted (PGP-->Clear).
		if contact != nil && contact.Scheme == pgpMime {
			return pmapi.ClearMIMEPackage, pmapi.ContentTypeMultipartMixed, nil
		} else if contact != nil && contact.Scheme == pgpInline {
			return pmapi.ClearPackage, pmapi.ContentTypePlainText, nil
		}

		// If no contact settings, follow mailbox defaults but send unencrypted (PGP-->Clear).
		if settingsScheme == pmapi.PGPMIMEPackage {
			return pmapi.ClearMIMEPackage, pmapi.ContentTypeMultipartMixed, nil
		} else if settingsScheme == pmapi.PGPInlinePackage {
			return pmapi.ClearPackage, pmapi.ContentTypePlainText, nil
		}
	}

	if !encrypted && !signed {
		// Always send as clear package if we are neither encrypting nor signing.
		scheme = pmapi.ClearPackage

		// If the contact is nil, no further modifications can be made.
		if contact == nil {
			return scheme, settingsMIMEType, nil
		}

		// Prefer contact mime settings.
		if contact.Scheme == pgpMime {
			return scheme, pmapi.ContentTypeMultipartMixed, nil
		} else if contact.Scheme == pgpInline {
			return scheme, pmapi.ContentTypePlainText, nil
		}

		// If contact has a preferred mime type, use that, otherwise follow mailbox default.
		if len(contact.MIMEType) > 0 {
			return scheme, contact.MIMEType, nil
		}
		return scheme, settingsMIMEType, nil
	}

	// If we end up here, something went wrong.
	err = errors.New("could not determine correct PGP Scheme and MIME Type to use to send mail")

	return scheme, mime, err
}

// checkContactKeysAgainstAPI keeps only those contact keys which are up to date and have
// an ID that matches an API key's ID.
func checkContactKeysAgainstAPI(contactKeys, apiKeys []*pmcrypto.KeyRing) (filteredKeys []*pmcrypto.KeyRing, err error) { //nolint[unparam]
	keyIDsAreEqual := func(a, b interface{}) bool {
		aKey, bKey := a.(*pmcrypto.KeyRing), b.(*pmcrypto.KeyRing)
		return aKey.GetEntities()[0].PrimaryKey.KeyId == bKey.GetEntities()[0].PrimaryKey.KeyId
	}

	for _, v := range algo.SetIntersection(contactKeys, apiKeys, keyIDsAreEqual) {
		filteredKeys = append(filteredKeys, v.(*pmcrypto.KeyRing))
	}

	return
}
