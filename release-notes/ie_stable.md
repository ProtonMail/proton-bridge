## v1.3.3
- 2021-05-17

### Fixed
- Fixed potential security vulnerability related to rpath
- Improved parsing of embedded messages


## v1.3.1
- 2021-03-11

### New
- Reduce the number of import errors by supporting malformed undisclosed-recipient and better handling of overly long headers
- Improvements to how large attachments are processed
- New format of the release notes

### Fixed
- Linux font issues - Fedora specific
- Fixed rare message misplacement
- Ensure removal of the startup entry during uninstallation
- Update errors


## v1.2.2
- 2020-11-27

### New
- Improvements to the import from large mbox files with multiple labels
- Not allow to run multiple instances of the app or transfers at the same time
- Better handling and displaying of skipped messages
- Various enhancements of the import process related to parsing
- Cosmetic GUI changes
- Better error handling

### Fixed
- Linux font issues - Fedora specific
- App response to the user pausing and canceling import or export
- Upgrade errors


## v1.1.2
- 2020-09-23

### New
- Improving performance
- Speed up import by implementing parallel processing (parallel fetch, encrypt and upload of messages)
- Optimising the initial fetch of messages from external accounts
- Better message parsing
- Better handling of attachments and non-standard formatting
- Improved stability of the message parser
- Improved metrics
- Added persistent anonymous API cookies

### Fixed
- Fixed issues causing failing of import
- Import from mbox files with long lines
- Improvements to import from Yahoo accounts
