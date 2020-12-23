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

### Removed

### Changed
* GODT-97 Don't log errors caused by SELECT "".

### Fixed
* GODT-979 Fix panic when trying to parse a multipart/alternative section that has no child sections.
