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
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

const (
	pgpInline  = "pgp-inline"
	pgpMIME    = "pgp-mime"
	pmInternal = "internal" // A mix between pgpInline and pgpMime used by PM.
)

// SendPreferences contains information about how to handle a message.
// It is derived from contact data, api key data, mail settings and composer preferences.
type SendPreferences struct {
	// Encrypt indicates whether the email should be encrypted or not.
	// If it's encrypted, we need to know which public key to use.
	Encrypt bool

	// Sign indicates whether the email should be signed or not.
	Sign bool

	// Scheme indicates if we should encrypt body and attachments separately and
	// what MIME format to give the final encrypted email. The two standard PGP
	// schemes are PGP/MIME and PGP/Inline. However we use a custom scheme for
	// internal emails (including the so-called encrypted-to-outside emails,
	// which even though meant for external users, they don't really get out of
	// our platform). If the email is sent unencrypted, no PGP scheme is needed.
	Scheme pmapi.PackageFlag

	// MIMEType is the MIME type to use for formatting the body of the email
	// (before encryption/after decryption). The standard possibilities are the
	// enriched HTML format, text/html, and plain text, text/plain. But it's
	// also possible to have a multipart/mixed format, which is typically used
	// for PGP/MIME encrypted emails, where attachments go into the body too.
	// Because of this, this option is sometimes called MIME format.
	MIMEType string

	// PublicKey contains an OpenPGP key that can be used for encryption.
	PublicKey *crypto.KeyRing
}

type sendPreferencesBuilder struct {
	internal bool
	encrypt  *bool
	sign     *bool
	scheme   *string
	mimeType *string

	publicKey *crypto.KeyRing
}

func (b *sendPreferencesBuilder) withInternal() {
	b.internal = true
}

func (b *sendPreferencesBuilder) isInternal() bool {
	return b.internal
}

func (b *sendPreferencesBuilder) withEncrypt(v bool) {
	b.encrypt = &v
}

func (b *sendPreferencesBuilder) withEncryptDefault(v bool) {
	if b.encrypt == nil {
		b.encrypt = &v
	}
}

func (b *sendPreferencesBuilder) shouldEncrypt() bool {
	if b.encrypt != nil {
		return *b.encrypt
	}

	return false
}

func (b *sendPreferencesBuilder) withSign(sign bool) {
	b.sign = &sign
}

func (b *sendPreferencesBuilder) withSignDefault() {
	v := true
	if b.sign == nil {
		b.sign = &v
	}
}

func (b *sendPreferencesBuilder) shouldSign() bool {
	if b.sign != nil {
		return *b.sign
	}

	return false
}

func (b *sendPreferencesBuilder) withScheme(v string) {
	b.scheme = &v
}

func (b *sendPreferencesBuilder) withSchemeDefault(v string) {
	if b.scheme == nil {
		b.scheme = &v
	}
}

func (b *sendPreferencesBuilder) getScheme() string {
	if b.scheme != nil {
		return *b.scheme
	}

	return ""
}

func (b *sendPreferencesBuilder) withMIMEType(v string) {
	b.mimeType = &v
}

func (b *sendPreferencesBuilder) withMIMETypeDefault(v string) {
	if b.mimeType == nil {
		b.mimeType = &v
	}
}

func (b *sendPreferencesBuilder) removeMIMEType() {
	b.mimeType = nil
}

func (b *sendPreferencesBuilder) getMIMEType() string {
	if b.mimeType != nil {
		return *b.mimeType
	}

	return ""
}

func (b *sendPreferencesBuilder) withPublicKey(v *crypto.KeyRing) {
	b.publicKey = v
}

// Build converts the PGP scheme with a string value into a number value, and
// we may override some of the other encryption preferences with the composer
// preferences. Notice that the composer allows to select a sign preference,
// an email format preference and an encrypt-to-outside preference. The
// object we extract has the following possible value types:
// {
//     encrypt: true | false,
//     sign: true | false,
//     pgpScheme: 	1 (protonmail custom scheme)
//								| 2 (Protonmail scheme for encrypted-to-outside email)
// 								| 4 (no cryptographic scheme)
// 								| 8 (PGP/INLINE)
//								| 16 (PGP/MIME),
//     mimeType: 'text/html' | 'text/plain' | 'multipart/mixed',
//     publicKey: OpenPGPKey | undefined/null
// }.
func (b *sendPreferencesBuilder) build() (p SendPreferences) {
	p.Encrypt = b.shouldEncrypt()
	p.Sign = b.shouldSign()
	p.MIMEType = b.getMIMEType()
	p.PublicKey = b.publicKey

	switch {
	case b.isInternal():
		p.Scheme = pmapi.InternalPackage

	case b.shouldSign() && b.shouldEncrypt():
		if b.getScheme() == pgpInline {
			p.Scheme = pmapi.PGPInlinePackage
		} else {
			p.Scheme = pmapi.PGPMIMEPackage
		}

	case b.shouldSign() && !b.shouldEncrypt():
		if b.getScheme() == pgpInline {
			p.Scheme = pmapi.ClearPackage
		} else {
			p.Scheme = pmapi.ClearMIMEPackage
		}

	default:
		p.Scheme = pmapi.ClearPackage
	}

	return
}

// setPGPSettings returns a SendPreferences with the following possible values:
//
// {
//    encrypt:   true 			 | false 	        | undefined/null/'',
//    sign:      true 			 | false          | undefined/null/'',
//    pgpScheme: 'pgp-mime'  | 'pgp-inline'   | undefined/null/'',
//    mimeType:  'text/html' | 'text/plain'   | undefined/null/'',
//    publicKey: OpenPGPKey  | undefined/null
// }
//
// These settings are simply a reflection of the vCard content plus the public
// key info retrieved from the API via the GET KEYS route.
func (b *sendPreferencesBuilder) setPGPSettings(
	vCardData *ContactMetadata,
	apiKeys []pmapi.PublicKey,
	isInternal bool,
) (err error) {
	// If there is no contact metadata, we can just use a default constructed one.
	if vCardData == nil {
		vCardData = &ContactMetadata{}
	}

	// Sending internal.
	// We are guaranteed to always receive API keys.
	if isInternal {
		b.withInternal()
		return b.setInternalPGPSettings(vCardData, apiKeys)
	}

	// Sending external but with keys supplied by WKD.
	// Treated pretty much same as internal.
	if len(apiKeys) > 0 {
		return b.setExternalPGPSettingsWithWKDKeys(vCardData, apiKeys)
	}

	// Sending external without any WKD keys.
	// If we have a contact saved, we can use its settings.
	return b.setExternalPGPSettingsWithoutWKDKeys(vCardData)
}

// setInternalPGPSettings returns SendPreferences for internal messages.
// An internal address can be either an obvious one: abc@protonmail.com,
// abc@protonmail.ch or abc@pm.me, or one belonging to a custom domain
// registered with proton.
func (b *sendPreferencesBuilder) setInternalPGPSettings(
	vCardData *ContactMetadata,
	apiKeys []pmapi.PublicKey,
) (err error) {
	// We're guaranteed to get at least one valid (i.e. not expired, revoked or
	// marked as verification-only) public key from the server.
	if len(apiKeys) == 0 {
		return errors.New("an API key is necessary but wasn't provided")
	}

	// We always encrypt and sign internal mail.
	b.withEncrypt(true)
	b.withSign(true)

	// We use a custom scheme for internal messages.
	b.withScheme(pmInternal)

	// If user has overridden the MIMEType for a contact, we use that.
	// Otherwise, we take the MIMEType from the composer.
	if vCardData.MIMEType != "" {
		b.withMIMEType(vCardData.MIMEType)
	}

	sendingKey, err := pickSendingKey(vCardData, apiKeys)
	if err != nil {
		return
	}

	b.withPublicKey(sendingKey)

	return nil
}

// pickSendingKey tries to determine which key to use to encrypt outgoing mail.
// It returns a keyring containing the chosen key or an error.
//
// 1. If there are pinned keys in the vCard, those should be given preference
//    (assuming the fingerprint matches one of the keys served by the API).
// 2. If there are pinned keys in the vCard but no matching keys were served
//    by the API, we use one of the API keys but first show a modal to the
//    user to ask them to confirm that they trust the API key.
//    (Use case: user doesn't trust server, pins the only keys they trust to
//    the contact, rogue server sends unknown keys, user should have option
//    to say they don't recognise these keys and abort the mail send.)
// 3. If there are no pinned keys, then the client should encrypt with the
//    first valid key served by the API (in principle the server already
//    validates the keys and the first one provided should be valid).
func pickSendingKey(vCardData *ContactMetadata, rawAPIKeys []pmapi.PublicKey) (kr *crypto.KeyRing, err error) {
	contactKeys := make([]*crypto.Key, len(vCardData.Keys))
	apiKeys := make([]*crypto.Key, len(rawAPIKeys))

	for i, key := range vCardData.Keys {
		var ck *crypto.Key

		// Contact keys are not armored.
		if ck, err = crypto.NewKey([]byte(key)); err != nil {
			return
		}

		contactKeys[i] = ck
	}

	for i, key := range rawAPIKeys {
		var ck *crypto.Key

		// API keys are armored.
		if ck, err = crypto.NewKeyFromArmored(key.PublicKey); err != nil {
			return
		}

		apiKeys[i] = ck
	}

	matchedKeys := matchFingerprints(contactKeys, apiKeys)

	var sendingKey *crypto.Key

	switch {
	// Case 1.
	case len(matchedKeys) > 0:
		sendingKey = matchedKeys[0]

	// Case 2.
	case len(matchedKeys) == 0 && len(contactKeys) > 0:
		// NOTE: Here we should ask for trust confirmation.
		sendingKey = apiKeys[0]

	// Case 3.
	default:
		sendingKey = apiKeys[0]
	}

	return crypto.NewKeyRing(sendingKey)
}

func matchFingerprints(a, b []*crypto.Key) (res []*crypto.Key) {
	aMap := make(map[string]*crypto.Key)

	for _, el := range a {
		aMap[el.GetFingerprint()] = el
	}

	for _, el := range b {
		if _, inA := aMap[el.GetFingerprint()]; inA {
			res = append(res, el)
		}
	}

	return
}

func (b *sendPreferencesBuilder) setExternalPGPSettingsWithWKDKeys(
	vCardData *ContactMetadata,
	apiKeys []pmapi.PublicKey,
) (err error) {
	// We're guaranteed to get at least one valid (i.e. not expired, revoked or
	// marked as verification-only) public key from the server.
	if len(apiKeys) == 0 {
		return errors.New("an API key is necessary but wasn't provided")
	}

	// We always encrypt and sign external mail if WKD keys are present.
	b.withEncrypt(true)
	b.withSign(true)

	// If the contact has a specific Scheme preference, we set it (otherwise we
	// leave it unset to allow it to be filled in with the default value later).
	if vCardData.Scheme != "" {
		b.withScheme(vCardData.Scheme)
	}

	// Because the email is signed, the cryptographic scheme determines the email
	// format. A PGP/INLINE scheme forces to use plain text. A PGP/MIME scheme
	// forces the automatic format.
	switch vCardData.Scheme {
	case pgpMIME:
		b.removeMIMEType()
	case pgpInline:
		b.withMIMEType("text/plain")
	}

	sendingKey, err := pickSendingKey(vCardData, apiKeys)
	if err != nil {
		return
	}

	b.withPublicKey(sendingKey)

	return nil
}

func (b *sendPreferencesBuilder) setExternalPGPSettingsWithoutWKDKeys(
	vCardData *ContactMetadata,
) (err error) {
	b.withEncrypt(vCardData.Encrypt)

	if vCardData.SignIsSet {
		b.withSign(vCardData.Sign)
	}

	// Sign must be enabled whenever encrypt is.
	if vCardData.Encrypt {
		b.withSign(true)
	}

	// If the contact has a specific Scheme preference, we set it (otherwise we
	// leave it unset to allow it to be filled in with the default value later).
	if vCardData.Scheme != "" {
		b.withScheme(vCardData.Scheme)
	}

	// If we are signing the message, the PGP scheme overrides the MIMEType.
	// Otherwise, we read the MIMEType from the vCard, if set.
	if vCardData.Sign {
		switch vCardData.Scheme {
		case pgpMIME:
			b.removeMIMEType()
		case pgpInline:
			b.withMIMEType("text/plain")
		}
	} else if vCardData.MIMEType != "" {
		b.withMIMEType(vCardData.MIMEType)
	}

	if len(vCardData.Keys) > 0 {
		var key *crypto.Key

		// Contact keys are not armored.
		if key, err = crypto.NewKey([]byte(vCardData.Keys[0])); err != nil {
			return
		}

		var kr *crypto.KeyRing

		if kr, err = crypto.NewKeyRing(key); err != nil {
			return
		}

		b.withPublicKey(kr)
	}

	return nil
}

// setEncryptionPreferences sets the undefined values in the SendPreferences
// determined thus far using using the (global) user mail settings.
// The object we extract has the following possible value types:
//
// {
//     encrypt: true | false,
//     sign: true | false,
//     pgpScheme: 'pgp-mime' | 'pgp-inline',
//     mimeType: 'text/html' | 'text/plain',
//     publicKey: OpenPGPKey | undefined/null
// }
//
// The public key can still be undefined as we do not need it if the outgoing
// email is not encrypted.
func (b *sendPreferencesBuilder) setEncryptionPreferences(mailSettings pmapi.MailSettings) {
	// For internal addresses or external ones with WKD keys, this flag should
	// always be true. For external ones, an undefined flag defaults to false.
	b.withEncryptDefault(false)

	// For internal addresses or external ones with WKD keys, this flag should
	// always be true. For external ones, an undefined flag defaults to the user
	// mail setting "Sign External messages". Otherwise we keep the defined value
	// unless it conflicts with the encrypt flag (we do not allow to send
	// encrypted but not signed).
	if mailSettings.Sign > 0 {
		b.withSignDefault()
	}

	if b.shouldEncrypt() {
		b.withSign(true)
	}

	// If undefined, default to the user mail setting "Default PGP scheme".
	// Otherwise keep the defined value.
	switch mailSettings.PGPScheme {
	case pmapi.PGPInlinePackage:
		b.withSchemeDefault(pgpInline)
	case pmapi.PGPMIMEPackage:
		b.withSchemeDefault(pgpMIME)
	case pmapi.ClearMIMEPackage, pmapi.ClearPackage, pmapi.EncryptedOutsidePackage, pmapi.InternalPackage:
		// nothing to set
	}

	// Its value is constrained by the sign flag and the PGP scheme:
	//  - Sign flag = true → For a PGP/Inline scheme, the MIME type must be
	//    'plain/text'. Otherwise we default to the user mail setting "Composer mode"
	//  - Sign flag = false → If undefined, default to the user mail setting
	//    "Composer mode". Otherwise keep the defined value.
	if b.shouldSign() && b.getScheme() == pgpInline {
		b.withMIMEType("text/plain")
	} else {
		b.withMIMETypeDefault(mailSettings.DraftMIMEType)
	}
}

func (b *sendPreferencesBuilder) setMIMEPreferences(composerMIMEType string) {
	// If the sign flag (that we just determined above) is true, then the MIME
	// type is determined by the PGP scheme (also determined above): we should
	// use 'text/plain' for a PGP/Inline scheme, and 'multipart/mixed' otherwise.
	// Otherwise we use the MIME type from the encryption preferences, unless
	// the plain text option has been selecting in the composer, which should
	// enforce 'text/plain' and override the encryption preference.
	if !b.isInternal() && b.shouldSign() {
		switch b.getScheme() {
		case pgpInline:
			b.withMIMEType("text/plain")
		default:
			b.withMIMEType("multipart/mixed")
		}
	} else if composerMIMEType == "text/plain" {
		b.withMIMEType("text/plain")
	}
}
