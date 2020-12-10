# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added

### Changed
* Log warning about permanently deleting messages.

### Removed

### Fixed
* GODT-135 Support parameters in SMTP `FROM MAIL` command, such as `BODY=7BIT`, or empty value `FROM MAIL:<>` used by some clients.
* GODT-338 GODT-781 GODT-857 GODT-866 Flaky tests.
* GODT-773 Replace old dates with birthday of RFC822 to not crash Apple Mail. Original is available under `X-Original-Date` header.
* License path on Arch distributions.
