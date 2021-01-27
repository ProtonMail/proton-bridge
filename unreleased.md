# ProtonMail Bridge and Import-Export unreleased

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added
* GODT-958 Release notes per eaach update channel.
* GODT-906 Handle RFC2047-encoded content transfer encoding values.
* GODT-875 Added GUI dialog on force update.
* GODT-820 Added GUI notification on impossibility of update installation (both silent and manual).
* GODT-870 Added GUI notification on error during silent update.
* GODT-805 Added GUI notification on update available.
* GODT-804 Added GUI notification on silent update installed (promt to restart).
* GODT-275 Added option to disable autoupdates in settings (default autoupdate is enabled).
* GODT-874 Added manual triggers to Updater module.
* GODT-851 Added support of UID EXPUNGE.
* GODT-928 Reject messages which are too large.

### Removed

### Changed

### Fixed
* GODT-787 GODT-978 Fix IE and Bridge importing to sent not showing up in inbox (setting up flags properly).
* GODT-1006 Use correct macOS keychain name.
* GODT-1009 Set ContentID if present and not explicitly attachment.
* GODT-900 Remove \Deleted flag after re-importing the message (do not delete messages by moving to local folder and back).
* GODT-908 Do not unpause event loop if other mailbox is still fetching.
* Check deprecated status code first to better determine API error.
