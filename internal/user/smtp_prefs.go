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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
)

const (
	pgpInline  = "pgp-inline"
	pgpMIME    = "pgp-mime"
	pmInternal = "internal" // A mix between pgpInline and pgpMime used by PM.
)

type contactSettings struct {
	Email     string
	Keys      []string
	Scheme    string
	Sign      bool
	SignIsSet bool
	Encrypt   bool
	MIMEType  rfc822.MIMEType
}

// newContactSettings converts the API settings into our local settings.
// This is due to the legacy send preferences code.
func newContactSettings(settings proton.ContactSettings) *contactSettings {
	metadata := &contactSettings{}

	if settings.MIMEType != nil {
		metadata.MIMEType = *settings.MIMEType
	}

	if settings.Sign != nil {
		metadata.Sign = *settings.Sign
		metadata.SignIsSet = true
	}

	if settings.Encrypt != nil {
		metadata.Encrypt = *settings.Encrypt
	}

	if settings.Scheme != nil {
		switch *settings.Scheme { // nolint:exhaustive
		case proton.PGPMIMEScheme:
			metadata.Scheme = pgpMIME

		case proton.PGPInlineScheme:
			metadata.Scheme = pgpInline

		default:
			panic("unknown scheme")
		}
	}

	if settings.Keys != nil {
		for _, key := range settings.Keys {
			b, err := key.Serialize()
			if err != nil {
				panic(err)
			}

			metadata.Keys = append(metadata.Keys, string(b))
		}
	}

	return metadata
}

func buildSendPrefs(
	contactSettings proton.ContactSettings,
	mailSettings proton.MailSettings,
	pubKeys []proton.PublicKey,
	mimeType rfc822.MIMEType,
	isInternal bool,
) (proton.SendPreferences, error) {
	builder := &sendPrefsBuilder{}

	if err := builder.setPGPSettings(newContactSettings(contactSettings), pubKeys, isInternal); err != nil {
		return proton.SendPreferences{}, fmt.Errorf("failed to set PGP settings: %w", err)
	}

	builder.setEncryptionPreferences(mailSettings)

	builder.setMIMEPreferences(string(mimeType))

	return builder.build(), nil
}

type sendPrefsBuilder struct {
	internal  bool
	encrypt   *bool
	sign      *bool
	scheme    *string
	mimeType  *rfc822.MIMEType
	publicKey *crypto.KeyRing
}

func (b *sendPrefsBuilder) withInternal() {
	b.internal = true
}

func (b *sendPrefsBuilder) isInternal() bool {
	return b.internal
}

func (b *sendPrefsBuilder) withEncrypt(v bool) {
	b.encrypt = &v
}

func (b *sendPrefsBuilder) withEncryptDefault(v bool) {
	if b.encrypt == nil {
		b.encrypt = &v
	}
}

func (b *sendPrefsBuilder) shouldEncrypt() bool {
	if b.encrypt != nil {
		return *b.encrypt
	}

	return false
}

func (b *sendPrefsBuilder) withSign(sign bool) {
	b.sign = &sign
}

func (b *sendPrefsBuilder) withSignDefault() {
	v := true
	if b.sign == nil {
		b.sign = &v
	}
}

func (b *sendPrefsBuilder) shouldSign() bool {
	if b.sign != nil {
		return *b.sign
	}

	return false
}

func (b *sendPrefsBuilder) withScheme(v string) {
	b.scheme = &v
}

func (b *sendPrefsBuilder) withSchemeDefault(v string) {
	if b.scheme == nil {
		b.scheme = &v
	}
}

func (b *sendPrefsBuilder) getScheme() string {
	if b.scheme != nil {
		return *b.scheme
	}

	return ""
}

func (b *sendPrefsBuilder) withMIMEType(v rfc822.MIMEType) {
	b.mimeType = &v
}

func (b *sendPrefsBuilder) withMIMETypeDefault(v rfc822.MIMEType) {
	if b.mimeType == nil {
		b.mimeType = &v
	}
}

func (b *sendPrefsBuilder) removeMIMEType() {
	b.mimeType = nil
}

func (b *sendPrefsBuilder) getMIMEType() rfc822.MIMEType {
	if b.mimeType != nil {
		return *b.mimeType
	}

	return ""
}

func (b *sendPrefsBuilder) withPublicKey(v *crypto.KeyRing) {
	b.publicKey = v
}

// Build converts the PGP scheme with a string value into a number value, and
// we may override some of the other encryption preferences with the composer
// preferences. Notice that the composer allows to select a sign preference,
// an email format preference and an encrypt-to-outside preference. The
// object we extract has the following possible value types:
//
//	{
//	    encrypt: true | false,
//	    sign: true | false,
//	    pgpScheme: 	1 (protonmail custom scheme)
//									| 2 (Protonmail scheme for encrypted-to-outside email)
//									| 4 (no cryptographic scheme)
//									| 8 (PGP/INLINE)
//									| 16 (PGP/MIME),
//	    mimeType: 'text/html' | 'text/plain' | 'multipart/mixed',
//	    publicKey: OpenPGPKey | undefined/null
//	}.
func (b *sendPrefsBuilder) build() (p proton.SendPreferences) {
	p.Encrypt = b.shouldEncrypt()
	p.MIMEType = b.getMIMEType()
	p.PubKey = b.publicKey

	if b.shouldSign() {
		p.SignatureType = proton.DetachedSignature
	} else {
		p.SignatureType = proton.NoSignature
	}

	switch {
	case b.isInternal():
		p.EncryptionScheme = proton.InternalScheme

	case b.shouldSign() && b.shouldEncrypt():
		if b.getScheme() == pgpInline {
			p.EncryptionScheme = proton.PGPInlineScheme
		} else {
			p.EncryptionScheme = proton.PGPMIMEScheme
		}

	case b.shouldSign() && !b.shouldEncrypt():
		if b.getScheme() == pgpInline {
			p.EncryptionScheme = proton.ClearScheme
		} else {
			p.EncryptionScheme = proton.ClearMIMEScheme
		}

	default:
		p.EncryptionScheme = proton.ClearScheme
	}

	return p
}

// setPGPSettings returns a SendPreferences with the following possible values:
//
//	{
//	   encrypt:   true 			 | false 	        | undefined/null/'',
//	   sign:      true 			 | false          | undefined/null/'',
//	   pgpScheme: 'pgp-mime'  | 'pgp-inline'   | undefined/null/'',
//	   mimeType:  'text/html' | 'text/plain'   | undefined/null/'',
//	   publicKey: OpenPGPKey  | undefined/null
//	}
//
// These settings are simply a reflection of the vCard content plus the public
// key info retrieved from the API via the GET KEYS route.
func (b *sendPrefsBuilder) setPGPSettings(
	vCardData *contactSettings,
	apiKeys []proton.PublicKey,
	isInternal bool,
) (err error) {
	// If there is no contact metadata, we can just use a default constructed one.
	if vCardData == nil {
		vCardData = &contactSettings{}
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
func (b *sendPrefsBuilder) setInternalPGPSettings(
	vCardData *contactSettings,
	apiKeys []proton.PublicKey,
) error {
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
		return err
	}

	b.withPublicKey(sendingKey)

	return nil
}

// pickSendingKey tries to determine which key to use to encrypt outgoing mail.
// It returns a keyring containing the chosen key or an error.
//
//  1. If there are pinned keys in the vCard, those should be given preference
//     (assuming the fingerprint matches one of the keys served by the API).
//  2. If there are pinned keys in the vCard but no matching keys were served
//     by the API, we use one of the API keys but first show a modal to the
//     user to ask them to confirm that they trust the API key.
//     (Use case: user doesn't trust server, pins the only keys they trust to
//     the contact, rogue server sends unknown keys, user should have option
//     to say they don't recognise these keys and abort the mail send.)
//  3. If there are no pinned keys, then the client should encrypt with the
//     first valid key served by the API (in principle the server already
//     validates the keys and the first one provided should be valid).
func pickSendingKey(vCardData *contactSettings, rawAPIKeys []proton.PublicKey) (*crypto.KeyRing, error) {
	contactKeys := make([]*crypto.Key, len(vCardData.Keys))
	apiKeys := make([]*crypto.Key, len(rawAPIKeys))

	for i, key := range vCardData.Keys {
		var ck *crypto.Key

		// Contact keys are not armored.
		var err error
		if ck, err = crypto.NewKey([]byte(key)); err != nil {
			return nil, err
		}

		contactKeys[i] = ck
	}

	for i, key := range rawAPIKeys {
		var ck *crypto.Key

		// API keys are armored.
		var err error
		if ck, err = crypto.NewKeyFromArmored(key.PublicKey); err != nil {
			return nil, err
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

func (b *sendPrefsBuilder) setExternalPGPSettingsWithWKDKeys(
	vCardData *contactSettings,
	apiKeys []proton.PublicKey,
) error {
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
		return err
	}

	b.withPublicKey(sendingKey)

	return nil
}

func (b *sendPrefsBuilder) setExternalPGPSettingsWithoutWKDKeys(
	vCardData *contactSettings,
) error {
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
		var (
			key *crypto.Key
			err error
		)

		// Contact keys are not armored.
		if key, err = crypto.NewKey([]byte(vCardData.Keys[0])); err != nil {
			return err
		}

		var kr *crypto.KeyRing

		if kr, err = crypto.NewKeyRing(key); err != nil {
			return err
		}

		b.withPublicKey(kr)
	}

	return nil
}

// setEncryptionPreferences sets the undefined values in the SendPreferences
// determined thus far using using the (global) user mail settings.
// The object we extract has the following possible value types:
//
//	{
//	    encrypt: true | false,
//	    sign: true | false,
//	    pgpScheme: 'pgp-mime' | 'pgp-inline',
//	    mimeType: 'text/html' | 'text/plain',
//	    publicKey: OpenPGPKey | undefined/null
//	}
//
// The public key can still be undefined as we do not need it if the outgoing
// email is not encrypted.
func (b *sendPrefsBuilder) setEncryptionPreferences(mailSettings proton.MailSettings) {
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
	case proton.PGPInlineScheme:
		b.withSchemeDefault(pgpInline)
	case proton.PGPMIMEScheme:
		b.withSchemeDefault(pgpMIME)
	case proton.ClearMIMEScheme, proton.ClearScheme, proton.EncryptedOutsideScheme, proton.InternalScheme:
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

func (b *sendPrefsBuilder) setMIMEPreferences(composerMIMEType string) {
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
