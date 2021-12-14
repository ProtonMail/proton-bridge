## v2.0.1
- 2021-12-15

### New

New Bridge GUI

* Added a Status View in addition to the Main Bridge Window
* Added storage information per signed in account
* Refactor of sign in flows
- Refactor of  Helps and Settings section
- Refactor of  bug reports
- Refactor of Bridge update flows for beta and stable channeles
- Introduced Reset Bridge feature - to clear all the local preferences and settings
- Introduce local cache configuration

New local cache

* Refactor of message.BodyStructure
* Refactor of local store
* Allow to store full encrypted message bodies on dysk

### Fixed
- Blurry system icons with multiple monitor setup


## v1.8.12
- 2021-11-30

### New

- Bridge to only be checking and trying to unclock active keys, both user and address


## v1.8.11
- 2021-11-18

### Fixed

- Updated bbold to v1.3.6 - including Unix fixes
- Ensured 'delete' on 'All Mail' is not allowed
- Fixed behaviour for 'append' of external messages to Archive
- Fixed behaviour for 'append' of internal messages to All Mail 
- Ensure 'move' to All Mail returns an error
- Fixed behaviour for moving/removing message to/from Spam


## v1.8.10
- 2021-10-01

### Fixed

- Updated crypto-libraries to gopenpgp v2.2.2 and go-srp v0.0.1
- Ensuring proper handling of updates when the user downloads the newest version manually
- Better handling of an error for importing too large messages via Bridge
- Ensuring message packages are fully built when the list of recipients includes internal addresses (for the users using active domain with Microsoft exchange)
- Fixed Uninstalling on Windows to properly clear updates
- Improvements to reusing connections - performance


## v1.8.9
- 2021-09-01

### Fixed

- Fixed an issues with incorrect handling of 401 server error leading to random Bridge logouts
- Changed encoding of message/rfc822 - to better handle sending of the .msg files
- Fixed crash within RFC822 builder for invalid or empty headers
- Fixed crash within RFC822 builder for header with key length > 76 chars


## v1.8.7
- 2021-06-22

### New

- Updated crypto-libraries to gopenpgp/v2 v2.1.10

### Fixed

- Fixed IMAP/SMTP restart in Bridge to mitigate connection issues
- Fixed unknown charset error for 'combined' messages
- Implemented a long-term fix for 'invalid or missing message signature' error


## v1.8.5
- 2021-06-11

### New

- Updated golang Secure Remote Password Protocol
- Updated crypto-libraries to gopenpgp/v2 v2.1.9
- Implemented new message parser (for imports from external accounts)

### Fixed

- Bridge not to strip PGP signatures of incoming clear text messages
- Import of messages with malformed MIME header
- Improved parsing of message headers
- Fetching bodies of non-multipart messages
- Sync and performance improvements


## v1.8.3
- 2021-05-27

### Fixed

- A bug with sending encrypted emails to external contacts 


## v1.8.2
- 2021-05-21

### Fixed

- Hotfix for error during bug reporting


## v1.8.1
- 2021-05-19

### Fixed

- Hotfix for crash when listing empty folder


## v1.8.0
- 2021-05-10

### New

- Implemented connection manager to improve performance during weak connection, better handling of connection loss and other connectivity issues
- Prompt profile installation during Apple Mail auto-configuration on MacOS Big Sur

### Fixed

- Bugs with building of message bodies/headers
- Incorrect naming format of some of the attachments 


## v1.7.1
- 2021-04-27

### New

- Refactor of message builder to achieve greater RFC compliance
- Increased the number of message fetchers to allow more parallel requests - performance improvement
- Log changes for easier debugging (update-related)

### Fixed

- Removed html-wrappig of non-decriptable messages - to facilitate decryption outside Bridge and/or allow to store such messages as they are
- Tray icon issues with multiple displays on MacOS


## v1.6.9
- 2021-03-30

### New

- Revise storage locations for the config files, preferences and cache
- Log improvements for easier debugging (sync issues)
- Added relevant metadata to Windows builds

### Fixed

- Fixed the way Bridge interacts with Windows Firewall and Defender
- Fixed potential security vulnerability related to rpath
- Improved parsing of embedded messages
- GUI bug fixes


## v1.6.6
- 2021-02-26

### Fixed

- Fixed update notifications
- Fixed GUI freeze while switching to early update channel
- Fixed Bridge autostart
- Improved signing of update packages

## v1.6.5
- 2021-02-22

### New

- Allow to choose which keychain is used by Bridge on Linux
- Added automatic update CLI commands
- Improved performance during slow connection
- Added IMAP requests to the logs for easier debugging 

### Fixed

- NoGUI bulid
- Background of GUI welcome message
- Incorrect total mailbox size displayed in Apple Mail

## v1.6.3
- 2021-02-16

### New

- Added desktop files and icon in Bridge repo
- Better detection of MacOS version to improve automatic AppleMail configuration
- Clearing cache after switching early access off

### Fixed

- Better poor connection handling - added retries for starting IMAP server after the connection was down
- Excluding updates from 'clearing cache'
- Not allowing copying from Inbox to Sent and vice versa
- Improvements to moving messages (unlabelling folders)
- Fixed the separation of release notes for 'early' and 'stable' channels


## v1.6.2
- 2021-02-02

### New
Introducing silent updates

- Introducing 'early' and 'stable' updates channels
- Allowing users to enable early access from within the GUI
- Adding and option to disable silent updates in settings

Changing the distribution of release notes

Performance and stability

- Implement support of UID EXPUNGE - to avoid avoid unnecessary resync
- Improve memory usage - gopenpgp dependency updated to v2.1.3
- Reducing network traffic by caching message size and body structure

Adding a scroll bar to the settings tab

### Fixed
- Fetch errors - introducing a stop to the imap server once there is no internet connection
- Setting up flags to avoid messages misplacement
- Inline messages incorrectly changed to attachments 
- Reporting bug from accounts with empty account name
- Panic when stopping import progress during loading mailboxes info
- Panic when modifying addresses during changing address mode
- Panic when trying to parse a multipart/alternative section that has no child sections
- Prevent potential loss of messages when moving to local folder and back
