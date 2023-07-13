## v3.3.0
- 2023-06-20

### New
- Reduced the number of occasions when email clients ask for Bridge credentials
- Added new Bridge notifications to help users to configure and troubleshoot their email clients
- To avoid the need to reconfigure email clients, Bridge remembers the old account password when an account is re-added (removed and added again)
- Further improved logging to support troubleshooting
- 2 factor authentication (2FA) is submitted automatically after entering a code
- Removed the requirement of having an administrator account on macOS to install Bridge

### Fixed
- Fixed numerous crashes
- Fixed the case when an email could not be sent if a PDF was attached to the email
- Added varioius bugfixes and security improvemenets
- Reduced the Bridge cache size by cleaning up temporary emails that were saved during failed initial synchronizations
- Further reduced the chance of desyncronization between the email client and Bridge


## v3.2.0
- 2023-05-26

### New
- Enhanced Proton infrastructure protection
- Enhanced the integration with the operating system by replacing status windows with native tray icon context menu
- Switched to two columns layout on the account details page to make the informaion easier to access
- Improved logs to support troubleshooting
- Added optional usage sharing to support user experience improvements. Additional information about data sharing can be found on our [support page](https://proton.me/support/share-usage-statistics).
- Implemented smart picking of default IMAP and SMTP ports
- Added various security and performance improvements

### Fixed
- Replaced invalid email addresses with empty field for new drafts so it can be syncronized across Proton clients
- Improved crash handling
- Fixed label / unlabel performance when applied on large amount of emails
- Fixed "reply to" related issues
- Updated build instructions
- Announced IMAP ID capability to email clients

## v3.1.3
- 2023-05-10

### Fixed
- Added a missing error handler that can make the initial synchronization to stuck


## v3.1.2
- 2023-04-27

### New
- Significantly reduced memory consumption both during synchronization and communication with email clients
- Added synchronization indicator to the graphical user interface (GUI)
- Added "Close window" and "Quit Bridge" buttons to the main window
- Added command line switches to control GUI rendering
- Switched to software rendering on Windows to support old graphics cards
- Added support for Proton's Scheduled send feature
- Avoided making email clients to ask for Bridge credentials when they started faster than Bridge at startup
- Added a notification when a user is signed out from Bridge in the background
- Improved desynchronization avoidence by setting UIDValidity from the current time
- Started updating emails in the email clients frequently when Bridge is started after not being online for longer period of time
- Improved error detection and handling
- Optimized Recovered Messages folder size by not adding a message to the folder if that message has been added to it before (deduplication)

### Fixed
- Fixed transparent window with old graphics cards or virtual machines on Windows
- Reduced notifications that does not require user actions
- Improved exception / crash handling
- Improved handling complex MIME types
- Reduced the source of errors that can lead to gRPC related error messages
- Fixed sub-folder rename issues
- Fixed various bugs related to secure vault handling, network communication errors, Proton server communication, operating system integration.


## v3.0.21
- 2023-03-23

### New
- Extended the migration from the previous major Bridge version with certificates
- Improved error detection

### Fixed
- Fixed the misplaced .desktop file on Linux
- Fixed DBUS secret service integration (e.g., KWallet, KeePass)
- Made Bridge more resilient against Proton server outages


## v3.0.20
- 2023-03-09

### New
- Added better explanation when an email cannot be sent because of non-existing email addresses
- Added a dialog to Bridge where users can repair the application when it encounters an internal error
- Improved error detection

### Fixed
- Reduced the cases when Bridge could not restart automatically
- Fixed the bug that could cause email states (e.g., read, unread, answered) to come out of sync with the web application. **NOTE: This fix is only applied to new emails. In order to fix older emails in Bridge, the account in Bridge needs to be removed and added back.**
- Fixed incorrect subject parsing caused by double quotes


## v3.0.19
- 2023-03-01

### New
- Improved inter-process communication error detection
- Improved exceptions related error detection

### Fixed
- Fixed numerous sources of errors leading to logout (internal errors)
- Fixed inter-process communication related startup issues (e.g., gRPC, service configuration file exchange)


## v3.0.18
- 2023-02-24

### New
- Improved event processing related error handling

### Fixed
- Fixed manual update errors on Windows by ensuring that all new files are deployed by the Bridge installer


## v3.0.17
- 2023-02-22

### New
- Rewrote a significant part of Bridge to improve overall Bridge stability and performance
- Open sourced and integrated a new IMAP library, Gluon (https://github.com/ProtonMail/gluon)
- Open sourced and integrated a new Proton API library, Go Proton API (GPA) (https://github.com/ProtonMail/go-proton-api)
- Significantly improved error detection and unexpected error handling
- Improved email sending performance
- Improved synchronization performance
- Added new command line argument for software rendering
- Extended the coverage of the Bridge data that is encrypted on the users' computers
- Added an option to the graphical user interface to export TLS certificates
- Reimplemented the user interface (upgraded to the Qt 6 user interface library)
- Added native Apple Silicon macOS support
- Added an option to change IMAP connection mode
- Added subfolder support
- Added a new icon for the .dmg installer
- Increased automated test coverage

### Fixed
- Desynchronization while creating a draft email
- Fixed sub-folder creation issues
- User name and password validation messages are shown only when the Sign in button is pressed
- Improved handling SMTP send deduplication
- Improved robustness of Bridge restart
- Fixed the user notification for occupied Bridge ports
- Fixed vulnerabilities of golang.org/x/crypto
- Added the missing Library on Fedora/Gnome for 2.3 to 2.4 update
- Added digital-signature for DLLs (to avoid the Windows Security alert, and to show if Bridge is coming from a trusted publisher)
- Fixed many Qt 6 related Linux and Windows update process issues
- Changed the default location of the database and storage files to avoid conflicts with cache cleaner applications


## v2.3.0
- 2022-09-08

### New
- Feature to hide All Mail from IMAP client
- Enable automatic configuration on macOS Ventura
- Improved the scope of local logs

### Fixed
- Visibility of Dependencies in Bridge GUI
- Potential crashes on parallel LIST command


## v2.2.2
- 2022-08-03

### New
- Updated GUI colours to reflect new Proton's colours theme
- Renamed ProtonMail Bridge to Proton Mail Bridge - installers, keychain etc.
- New Bridge systray icons for all OSes
- New Bridge application icons for all OSes
- Visual update of macOS and Windows installers
- Add label/folder filtering to pmapi
- Use one buffered event for internet status changes - to eliminate false-positive connection errors
- Added a modal to prompt the user to reconfigure the account once a new PM address is added
- Added a link to dependencies' licences to the help section footer

### Fixed
- Updated crypto-libraries to gopenpgp v2.4.7 and go-srp v0.0.5
- Convert charset only for `text/*` MIME types - to ensure no attachment corruption when sending with some email clients
- Reduce unnecessary shell executions
- Syncing issues for when a new PM address is added
- Changed the wording of 'delete this account' dialog
- Improved manual update process (GUI changes)
- Improvements to manual update process


## v2.1.3
- 2022-04-27

### New
- Added keybase/go-keychain/secretservice as a new keychain helper
- GUI changes to 'report a problem'
- Added another proxy provider
- Improved UX for working with keychain on macOS

### Fixed
- Manual update mechanism
- Windows clipboard issues (copying account details)
- Random logouts on macOS
- Error for corrupted keychain
- Bug reporting (emails send from custom domain)


## v2.1.1
- 2022-02-09

### New
- Improved Sentry reporting

### Fixed
- Ensure messageID is properly removed from DB when it is no longer present on the API


## v2.1.0
- 2022-01-31

### New
New Bridge GUI

- Added a Status View in addition to the Main Bridge Window
- Added storage information per signed in account
- Refactor of sign in flows
- Refactor of Helps and Settings section
- Refactor of bug reports
- Refactor of Bridge update flows for beta and stable channels
- Introduced Reset Bridge feature - to clear all the local preferences and settings
- Introduce local cache configuration
- Dark Mode for Bridge, including autodetect mechanism for system colour scheme
- GUI element for changing keychain (Linux)

New local cache

- Refactor of local store (caching of size, headers and bodystructure)
- Allow to store full encrypted message bodies on a disk

Other

- Update to goopenpgp 2.4.1
- Optimising sentry reporting

### Fixed
- Improved retry mechanism for connecting to Proton servers
- OpenGL issue during startup for specific GPUs
- Blurry system icons with multiple monitor setup
- Bridge crashes related to unlocking local cache
- Bug with sending to 'non-encrypted' recipients
- Cosmetic GUI changes


## v1.8.12
- 2021-12-06

### New
- Bridge to only be checking and trying to unclock active keys, both user and address

### Fixed
- Updated bbold to v1.3.6 - including Unix fixes
- Ensure 'delete' on 'All Mail' is not allowed
- Fixed behaviour for 'append' of external messages to Archive
- Fixed behaviour for 'append' of internal messages to All Mail 
- Ensure 'move' to All Mail returns an error
- Fixed behaviour for moving/removing message to/from Spam


## v1.8.10
- 2021-10-13

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
- 2021-06-24

### New
- Updated golang Secure Remote Password Protocol
- Updated crypto-libraries to gopenpgp/v2 v2.1.10
- Implemented new message parser (for imports from external accounts)

### Fixed
- Fixed IMAP/SMTP restart in Bridge to mitigate connection issues
- Fixed unknown charset error for 'combined' messages
- Implemented a long-term fix for 'invalid or missing message signature' error
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
- 2021-05-17

### New
- Refactor of message builder to achieve greater RFC compliance
- Implemented connection manager to improve performance during weak connection, better handling of connection loss and other connectivity issues
- Increased the number of message fetchers to allow more parallel requests - performance improvement
- Log changes for easier debugging (update-related)
- Prompt profile installation during Apple Mail auto-configuration on MacOS Big Sur

### Fixed
- Bugs with building of message bodies/headers
- Incorrect naming format of some of the attachments 
- Removed html-wrappig of non-decriptable messages - to facilitate decryption outside Bridge and/or allow to store such messages as they are
- Tray icon issues with multiple displays on MacOS


## v1.6.9
- 2021-04-01

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
- 2021-03-04

### New
- Allow to choose which keychain is used by Bridge on Linux
- Added automatic update CLI commands
- Improved performance during slow connection
- Added IMAP requests to the logs for easier debugging

### Fixed
- Fixed update notifications
- Fixed GUI freeze while switching to early update channel
- Fixed Bridge autostart
- Improved signing of update packages
- NoGUI build
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


## v1.5.7
- 2021-01-21

### New
- Improvements to message parsing
- Better error handling
- Ensured better message flow by refactoring both address and date parsing
- Improved secure connectivity checks
- Better deb packaging
- More robust error handling
- Improved package creation logic
- Refactor of sending functions to simplify code maintenance
- Added tests for package creation
- Support read confirmations
- Adding GPLv3 licence button to the GUI
- Improved testing

### Fixed
- AppleMail crashes (related to timestamps)
- Sending messages from aliases in combined inbox mode
- Fedora font issues
- Ensured that conversations are properly threaded
- Fixed Linux font issues (Fedora)
- Better handling of Mime encrypted messages
- Bridge crashes related to labels handling
- GUI popup related to TLS connection error
- An issue where a random session key is included in the data payload
- Error handling (including improved detection)
- Encoding errors
- Installation issues on linux
