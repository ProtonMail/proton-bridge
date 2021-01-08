# ProtonMail Bridge and Import-Export unreleased

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added

### Removed

### Changed

### Fixed
* GODT-979 Fix panic when trying to parse a multipart/alternative section that has no child sections.
### Changed
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-773 Replace `INTERNALDATE` older than birthday of RFC822 by birthday of RFC822 to not crash Apple Mail.
* GODT-927 Avoid to call API with empty label name.
* GODT-732 Fix usage of fontawesome
* GODT-885 Do not explicitly unlabel folders during move to match behaviour of other clients.
