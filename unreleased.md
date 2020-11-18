# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added
* GODT-701 Try load messages one-by-one if IMAP server errors with batch load and not interrupt the transfer

### Changed
* GODT-651 Build creates proper binary names.
## Added
* GODT-878 Tests for send packet creation logic

## Changed
* GODT-878 Refactor and move the send packet creation login to `pmapi.SendMessageReq`
