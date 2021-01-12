# ProtonMail Bridge and Import-Export unreleased

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added

### Removed

### Changed

### Fixed
* GODT-787 GODT-978 Fix IE and Bridge importing to sent not showing up in inbox (setting up flags properly).
* GODT-1006 Use correct macOS keychain name.
* GODT-1009 Set ContentID if present and not explicitly attachment.
* GODT-900 Remove \Deleted flag after re-importing the message (do not delete messages by moving to local folder and back).
* GODT-908 Do not unpause event loop if other mailbox is still fetching.
* Check deprecated status code first to better determine API error.
* GODT-732 Fix usage of fontawesome
* GODT-915 Bump go-imap dependency and remove go-imap-specialuse dependency.
* GODT-928 Reject messages which are too large.
