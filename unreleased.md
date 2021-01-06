# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added

### Removed

### Fixed
* GODT-922 Fix panic during restarting the bridge.
* GODT-945 Fix panic in integration tests caused by concurrent map writes.
* GODT-732 Fix usage of fontawesome.
* GODT-951 Properly parse message with long lines in header and long header split to multiple lines (upgrading to latest go-message).

### Changed
* GODT-858 Bump go-rfc5322 dependency to v0.5.0 to handle some invalid RFC5322 groups and add support for semicolon delimiter in address-list.
* GODT-923 Fix listener locking.

### Changed
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-773 Replace `INTERNALDATE` older than birthday of RFC822 by birthday of RFC822 to not crash Apple Mail.
