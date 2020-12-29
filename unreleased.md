# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added

### Removed

### Fixed
* GODT-135 Support parameters in SMTP `FROM MAIL` command, such as `BODY=7BIT`, or empty value `FROM MAIL:<>` used by some clients.
* GODT-338 GODT-781 GODT-857 GODT-866 Flaky tests.
* GODT-773 Replace old dates with birthday of RFC822 to not crash Apple Mail. Original is available under `X-Original-Date` header.
* GODT-922 Fix panic during restarting the bridge.
* GODT-945 Fix panic in integration tests caused by concurrent map writes.

### Changed
* GODT-858 Bump go-rfc5322 dependency to v0.5.0 to handle some invalid RFC5322 groups and add support for semicolon delimiter in address-list.
* GODT-923 Fix listener locking.

### Changed
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
