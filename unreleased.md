# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added
* GODT-906 Handle RFC2047-encoded content transfer encoding values.
* GODT-875 Added GUI dialog on force update.
* GODT-820 Added GUI notification on impossibility of update installation (both silent and manual).
* GODT-870 Added GUI notification on error during silent update.
* GODT-805 Added GUI notification on update available.
* GODT-804 Added GUI notification on silent update installed (promt to restart).
* GODT-275 Added option to disable autoupdates in settings (default autoupdate is enabled).
* GODT-874 Added manual triggers to Updater module.

### Changed
* Rename channels `beta->early`, `live->stable`.
* GODT-893 Bump go-rfc5322 dependency to v0.2.1 to properly detect syntax errors during parsing.
* GODT-892 Swap type and value from sentry exception and cut panic handlers from the traceback.
* GODT-854 EXPUNGE and FETCH unilateral responses are returned before OK EXPUNGE or OK STORE, respectively.
* GODT-806 Changed GUI dialog on manual update. Added autoupdates checkbox. Simplifyed installation process GUI.
* Bump gopenpgp dependency to v2.1.3 for improved memory usage.
* GODT-912 Changed scroll bar behaviour in settings tab
* GODT-858 Bump go-rfc5322 dependency to v0.5.0 to handle some invalid RFC5322 groups and add support for semicolon delimiter in address-list.
* GODT-923 Fix listener locking.
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-149 Send heartbeat ASAP on each new calendar day.

### Removed
* GODT-208 Remove deprecated use of BuildNameToCertificate.

### Changed
* GODT-97 Don't log errors caused by SELECT "".

### Fixed
* GODT-979 Fix panic when trying to parse a multipart/alternative section that has no child sections.
### Changed
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-773 Replace `INTERNALDATE` older than birthday of RFC822 by birthday of RFC822 to not crash Apple Mail.
* GODT-927 Avoid to call API with empty label name.
* GODT-732 Fix usage of fontawesome
