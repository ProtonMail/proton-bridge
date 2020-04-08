# Do not modify this file!
It is here for historical reasons only. All changes should be documented in the
Changelog at the root of this repository.


# Changelog for API
> NOTE we are using versioning for go-pmapi in format `major.minor.bugfix`
> * major stays at version 1 for the forseeable future
> * minor is increased when a force upgrade happens or in case of major breaking changes
> * patch is increased when new features are added

## v1.0.16

### Fixed
* Potential crash when reporting cert pin failure 

## v1.0.15

### Changed
* Merge only 50 events into one
* Response header timeout increased from 10s to 30s

### Fixed
* Make keyring unlocking threadsafe

## v1.0.14

### Added
* Config for disabling TLS cert fingerprint checking

### Fixed
* Ensure sensitive stuff is cleared on client logout even if requests fail

## v1.0.13

### Fixed
* Correctly set Transport in http client

## v1.0.12

### Changed
* Only `http.RoundTripper` interface is needed instead of full `http.Transport` struct

### Added
* GODT-61 (and related): Use DoH to find and switch to a proxy server if the API becomes unreachable
* GODT-67 added random wait to not cause spikes on server after StatusTooManyRequests

### Fixed
* FirstReadTimeout was wrongly timeout of the whole request including repeating ones, now it's really only timeout for the first read

## v1.0.11

### Added
* GODT-53 `Message.Type` added with constants `MessageType*`

## v1.0.10

### Added
* GODT-55 exporting DANGEROUSLYSetUID

### Changed
* The full communication between clien and API is logged if logrus level is trace

## v1.0.9

### Fixed
* Use correct address type value (because API starts counting from 1 but we were counting from 0)

## v1.0.8

### Added
* Introdcution of connection manager

### Fixed
* Deadlock during the auth-refresh
* Fixed an issue where some events were being discarded when merging

## v1.0.7

### Changed
* The given access token is saved during auth refresh if none was available yet


## v1.0.6

### Added
* `ClientConfig.Timeout` to be able to configure the whole timeout of request
* `ClientConfig.FirstReadTimeout` to be able to configure the timeout of request to the first byte
* `ClientConfig.MinSpeed` to be able to configure the timeout when the connection is too slow (limitation in minimum bytes per second)
* Set default timeouts for http.Transport with certificate pinning

### Changed
* http.Client by default uses ProxyFromEnvironment to support HTTP_PROXY and HTTPS_PROXY environment variables

## v1.0.5

### Added
* `ContentTypeMultipartEncrypted` MIME content type for encrypted email
* `MessageCounts` in event struct

## v1.0.4

### Added
* `PMKeys` for parsing and reading KeyRing
* `clearableKey` to rewrite memory
* Proton/backend-communication#25 Unlock with tokens (OneKey2RuleThemAll Phase I)

### Changed
* Update of gopenpgp: convert JSON to KeyRing in PMAPI
* `user.KeyRing` -> `user.KeyRing()`
* typo `client.GetAddresses()`

### Removed
* `address.KeyRing`

## v1.0.2 v1.0.3

### Changed
* Fixed capitalisation in a few places
* Added /metrics API route
* Changed function names to be compliant with go linter
* Encrypt with primary key only
* Fix `client.doBuffered` - closing body before handling unauthorized request
* go-pm-crypto -> GopenPGP
* redefine old functions in `keyring.go`
* `attachment.Decrypt` drops returning signature (does signature check by default)
* `attachment.Encrypt` is using readers instead of writers
* `attachment.DetachedSign` drops writer param and returns signature as a reader
* `message.Decrypt` drops returning signature (does signature check by default)
* Changed TLS report URL to https://reports.protonmail.ch/reports/tls
* Moved from current to soon TLS pin

## v1.0.1

### Removed
* `ClientID` from all auth routes
* `ErrorDescription` from error

## v1.0.0

### Changed
* `client.AuthInfo` does return 2FA information only when authenticated, for the first login information available in `Auth.HasTwoFactor`
* `client.Auth` does not accept 2FA code in favor of `client.Auth2FA`
* `client.Unlock` supports only new way of unlock with directly available access token

### Added
* `Res.StatusCode` to pass HTTP status code to responses
* `Auth.HasTwoFactor` method to determine whether account has enabled 2FA (same as `AuthInfo.HasTwoFactor`)
* `Auth2FA*` structs for 2FA endpoint
* `client.Auth2FA` method to fully unlock session with 2FA code
* `ErrUnauthorized` when request cannot be authorized
* `ErrBad2FACode` when bad 2FA and user cannot try again
* `ErrBad2FACodeTryAgain` when bad 2FA but user can try again

## 2019-08-06

### Added
* Send TLS issue report to API
* Cert fingerpring with `TLSPinning` struct
* Check API certificate fingerprint and verify hostname

### Changed
* Using `AddressID` for `/messge/count` and `/conversations/count`
* Less of copying of responses from the server in the memory

## 2019-08-01
* low case for `sirupsen`
* using go modules

## 2019-07-15

### Changed
* `client.Auths` field is removed in favor of function `client.SetAuths` which opens possibility to use interface

## 2019-05-18

### Changed
* proton/backend-communication#11 x-pm-uid sent always for `/auth/refresh`
* proton/backend-communication#11 UID never changes

## 2019-05-28

### Added
* New test server patern using callbacks
* Responses are read from json files

### Changed
* `auth_tests.go` to new callback server pattern
* Linter fixes for tests

### Removed
* `TestClient_Do_expired` due to no effect, use `DoUnauthorized` instead

## 2019-05-24
* Help functions for test
* CI with Lint

## 2019-05-23
* Log userID

## 2019-05-21
* Fix unlocking user keys

## 2019-04-25

### Changed
* rename `Uid` -> `UID` proton/backend-communication#11

## 2019-04-09

### Added
* sending attachments as zip `application/octet-stream`
* function `ReportReq.AddAttachment()`
* data memeber `ReportReq.Attachments`
* general function to report bug `client.Report(req ReportReq)` with object as parameter

### Changed
* `client.ReportBug` and `client.ReportBugWithClient` functions are obsolete and they uses `client.Report(req ReportReq)`
* `client.ReportCrash` is obsolete. Use sentry instead
* `Api`->`API`, `Uid`->`UID`

## 2019-03-13
* user id in raven
* add file position of panic sender

## 2019-03-06
* #30 update `pm-crypto` to store `KeyRing.FirstKeyID`
* #30 Add key salt to `Auth` object from `GetKeySalts` request
* #30 Add route `GET /keys/salt`
* removed unused `PmCrypto`

## 2019-02-20
* removed unused `decryptAccessToken`

## 2019-01-21
* #29 Parsing all goroutines from pprof
* #29 Sentry `Threads` implementation
* #29 using sentry for crashes

## 2019-01-07
* refactor `pmapi.DecryptString` -> `pmcrypto.KeyRing.DecryptString`
* fixed tests
* `crypto` -> `pmcrypto`
* refactoring code using repos `go-pm-crypto`, `go-pm-mime` and `go-srp`


## 2018-12-10
* #26 adding `Flags` field to message
* #26 removing fields deprecated by `Flags`: `IsEncrypted`, `Type`, `IsReplied`, `IsRepliedAll`, `IsForwarded`
* #26 removing deprecated consts (see #26 for replacement)
* #26 fixing tests (compiling not working)

## 2018-11-19

### Added
* Wait and retry from `DoJson` if banned from api

### Changed
* `ErrNoInternet` -> `ErrAPINotReachable`
* Adding codes for force upgrade: 5004 and 5005
* Adding codes for API offline: 7001
* Adding codes for BansRequests: 85131

## 2018-09-18

### Added
* `client.decryptAccessToken` if privateKey is received (tested with local api) #23

### Changed
* added fields to User
* local config TLS skip verify

## 2018-09-06

### Changed
* decrypt token only if needed

### Broken
* Tests are not working

## APIv3 UPDATE (2018-08-01)
* issue Desktop-Bridge#561

### Added
* Key flag consts
* `EventAddress`
* `MailSettings` object and route call
* `Client.KeyRingForAddressID`
* `AuthInfo.HasTwoFactor()`
* `Auth.HasMailboxPassword()`

### Changed
* Addresses are part of client
* Update user updates also addresses
* `BodyKey` and `AttachmentKey` contains `Key` and `Algorithm`
* `keyPair` (not use Pubkey) -> `pmKeyObject`
* lots of indent
* bugs route
* two factor (ready to U2F)
* Reorder some to match order in doc (easier to )
* omit address Order when empty
* update user and addresses in `CurrentUser()`
* `User.Unlock()` -> `Client.UnlockAddresses()`
* `AuthInfo.Uid`  -> `AuthInfo.Uid()`
* `User.Addresses`  -> `Client.Addresses()`

### Removed
* User v3 removed plenty (now in settings)
* Message v3 removed plenty (Starred is label)
