# ProtonMail Bridge and Import-Export app Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)

## Unreleased

### Added
* GODT-906 Handle RFC2047-encoded content transfer encoding values.

### Changed
* GODT-893 Bump go-rfc5322 dependency to v0.2.1 to properly detect syntax errors during parsing.
* GODT-892 Swap type and value from sentry exception and cut panic handlers from the traceback.
* GODT-854 EXPUNGE and FETCH unilateral responses are returned before OK EXPUNGE or OK STORE, respectively.

### Removed
* GODT-651 Build creates proper binary names.
* GODT-148 Allow import (using the Import-Export app) of already encrypted messages as is.
* GODT-202 Update to latest go-smtp.

### Fixed
* GODT-135 Support parameters in SMTP `FROM MAIL` command, such as `BODY=7BIT`, or empty value `FROM MAIL:<>` used by some clients.
* GODT-338 GODT-781 GODT-857 GODT-866 Flaky tests.
