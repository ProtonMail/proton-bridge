# ProtonMail Bridge and Import-Export unreleased

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
* GODT-851 Added support of UID EXPUNGE.

### Removed

### Fixed
* GODT-922 Fix panic during restarting the bridge.
* GODT-945 Fix panic in integration tests caused by concurrent map writes.
* GODT-732 Fix usage of fontawesome.
* GODT-951 Properly parse message with long lines in header and long header split to multiple lines (upgrading to latest go-message).
* GODT-894 Fix panic when sending while account is logging in.
* GODT-831 Fix reporting bug from accounts with empty account name.

### Changed
* GODT-97 Don't log errors caused by SELECT "".
* Rename channels `beta->early`, `live->stable`.
* GODT-892 Swap type and value from sentry exception and cut panic handlers from the traceback.
* GODT-854 EXPUNGE and FETCH unilateral responses are returned before OK EXPUNGE or OK STORE, respectively.
* GODT-806 Changed GUI dialog on manual update. Added autoupdates checkbox. Simplifyed installation process GUI.
* Bump gopenpgp dependency to v2.1.3 for improved memory usage.
* GODT-912 Changed scroll bar behaviour in settings tab
* GODT-149 Send heartbeat ASAP on each new calendar day.

### Removed
* GODT-208 Remove deprecated use of BuildNameToCertificate.

### Fixed
* GODT-946 Fix flaky tests notifying changes.
* GODT-979 Fix panic when trying to parse a multipart/alternative section that has no child sections.
### Changed
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-773 Replace `INTERNALDATE` older than birthday of RFC822 by birthday of RFC822 to not crash Apple Mail.
* GODT-927 Avoid to call API with empty label name.
* GODT-732 Fix usage of fontawesome
* GODT-915 Bump go-imap dependency and remove go-imap-specialuse dependency.
* GODT-831 Cancel request of uploading attachment if reading/writing it fails.

### Fixed
* GODT-900 Remove \Deleted flag after re-importing the message (do not delete messages by moving to local folder and back).
