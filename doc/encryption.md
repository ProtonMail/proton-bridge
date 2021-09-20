# Encryption

Encryption is done in PMAPI, bridge utils and bridge itself. The best would be to keep encryption
in PMAPI and bridge utils (in package such as messages). All packages are using our high-level
GopenPGP library on top of openpgp.

## `gopenpgp.KeyRing`

We use one `KeyRing` per address. Our usage then contains all keys for specific address. Primary
key is always on the first position, then there old ones to be able to decrypt last e-mail.
Openpgp encrypts given message with all available keys, so we need to first get first (primary)
key for encryption to have message encrypted only once with primary key.
