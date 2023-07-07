## v3.3.1
- 2023-07-07

### New
- Improved Bridge debugging capabilities by adding more information to the application logs
- Started measuring the Bridge setup experience

### Fixed
- Temporarily removed clickable notifications on Linux to not raise the Bridge window when receiving non-Bridge notifications


## v3.3.0
- 2023-06-08

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
- 2023-05-15

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
- Optimized Recovered Messages folder size by not adding a message to the folder if that message has been added to it before (deduplication)

## v3.1.1
- 2023-04-11

### Fixed
- Improved exception / crash handling

## v3.1.0
- 2023-04-05

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


## v3.0.16
- 2023-02-17

### Fixed
- Desynchronization while creating draft.


## v3.0.15
- 2023-02-14

### Fixed
- Better network error handling


## v3.0.14
- 2023-02-09

### New
- Improved error detection

### Fixed
- Fixed the sync issues that can happen when updating from an earlier v3 version
- Improved attachment handling by setting proper MIME parameters
- Improved update processing while Bridge is not active or performs a synchronization with Proton servers

## v3.0.12
- 2023-02-01

### New
- Changed the default location of the database and storage files. **NOTE: Please delete the old cache location if necessary.**
- Optimised cache, database and storage placement
- Improved email sending performance
- Improved unexpected event handling

### Fixed
- Outlook does not show sent messages as drafts
- Improved 'Reply to' behaviour


## v3.0.10
- 2023-01-17

### New
- Program argument to use software rendering.
- Improved exception handling in GUI.

### Fixed
- API event processing more robust.
- Improve the startup process.
- Fixed sub-folder creation bug.


## v3.0.9
- 2023-01-05

### New
- Added an option to the GUI to export TLS certificates
- Increased tolerance of invalid messages

### Fixed
- Autostart is set only when changed by the user
- Folders that are created during initial sync are synchronized correctly
- Improved settings migration from 2.x to 3.x
- Error reporting improvements on Intel Macs
- Show the setup guide after the first login
- User name and password validation messages are shown only when the Sign in button is pressed
- The Bridge main window is not shown on startup or after a crash
- Sign in button is not greyed out after the first login


## v3.0.8
- 2022-12-20

### New
- Improved error detection when Proton server updates cannot be processed

### Fixed
- Proton server update processing will not stop after a folder update failure

## v3.0.7
- 2022-12-19

### New
- Increase worker count (performance improvement)

### Fixed
- Bridge password migration from 2.x to 3.x
- Ensure proper handling of folders and labels with non-US ASCII chars


## v3.0.6
- 2022-12-12

### New
- New IMAP library (https://github.com/ProtonMail/gluon)
    - IMAP state managed entirely by the new IMAP library, to increase robustness and performance
    - Used ANTLR to generate a correct IMAP parser directly from RFC protocol description
    - Implemented an IMAP 'snapshot' system to ensure correct execution of IMAP commands when multiple clients are connected simultaneously
    - Full support of IMAP subscription
    - Full support of IMAP SEARCH
    - Allow users to modify the Gluon data location
    - Improved synchronization of local and remote changes
- New API library (https://github.com/ProtonMail/go-proton-api)
    - Switched from pmapi to go-proton-api
    - Stability and performance improvement
- Other
    - Added an option to change IMAP connection mode
    - Subfolder support

### Fixed
- Stability & Reliability improvements
    - Optimized SELECT, FETCH and SEARCH performance
    - Parallel user unlock (faster startup times)
    - Parallel file upload (faster send with attachments)
    - Parallel contact fetch (faster send to multiple addresses)
    - Implemented batching for increased performance for COPY/MOVE/STORE on multiple messages
    - Reduced reliance on OS keychain
- Other
    - Implemented sync manager
    - Improved handling SMTP send deduplication
    - Better user management
    - Improved Sentry reporting for easier debugging
    - Increase test coverage
    - GUI improvements


## v2.4.8
- 2022-11-15

### New
- More detailed logs for Bridge GUI
- GUI improvements

### Fixed
- Improved Bridge <-> Bridge-GUI communication
- Ensuring all the logs files are included when sending a bug report
- Fixes to the update process on Linux and Windows (qt6 related)


## v2.4.5
- 2022-11-08

### New
- GUI improvements
- More verbose logs for GUI-related issues
- New icon for .dmg installer

### Fixed
- Change download and version check urls to proton.me
- Fixed manual check for updates after switching the update channel


## v2.4.3
- 2022-10-25

### New
- Ensured the use of random port for gRPC
- Implemented token exchange for identity validation
- Ensured gRPC generates its own TLS certificate
- Increased bridge-gui timeout for gRPC server connection
- Added new warnings for 'TLS pinning' and 'no active key for recipient' errors

### Fixed
- GUI-related Bridge crashes
- The notification for when Bridge ports are occupied
- Fixed vulnerabilities of golang.org/x/crypto
- Missing Library on Fedora/Gnome upgrade form 2.3 to 2.4
- Added Digital-Signature for DLLs (Windows Security Alert to show Bridge as coming from a trusted publisher)


## v2.4.0
- 2022-09-28

### New
- Native Mac M1 release
- Upgrade to Qt 6:
    - Change the app architecture
    - Drop therecipe/qt dependency
    - Update to go1.18
    - Update to Qt 6.3.2

### Fixed
- Improved wording for specific errors
- Improved robustness of Bridge restart
- Status View visual improvements


## v2.3.0
- 2022-09-01

### New
- Feature to hide All Mail from IMAP client
- Enable automatic configuration on macOS Ventura
- Improved the scope of local logs

### Fixed
- Visibility of Dependencies in Bridge GUI
- Potential crashes on parallel LIST command


## v2.2.2
- 2022-07-27

### Fixed
- Improvements to manual update process


## v2.2.1
- 2022-07-21

### New
- New Bridge systray icons for all OSes
- New Bridge application icons for all OSes
- Visual update of macOS and Windows installers
- Add label/folder filtering to pmapi

### Fixed
- Updated crypto-libraries to gopenpgp v2.4.7 and go-srp v0.0.5
- Convert charset only for `text/*` MIME types - to ensure no attachment corruption when sending with some email clients
- Reduced unnecessary shell executions


## v2.2.0
- 2022-05-25

### New
- Updated GUI colours to reflect new Proton's colours theme
- Renamed ProtonMail Bridge to Proton Mail Bridge - installers, keychain etc.
- Use one buffered event for internet status changes
- Added a modal to prompt the user to reconfigure the account once a new PM address is added
- Added a link to dependencies' licences to the help section footer

### Fixed
- Syncing issues for when a new PM address is added
- Changed the wording of 'delete this account' dialog
- Improved manual update process (GUI changes)


## v2.1.3
- 2022-04-11

### New
- Added keybase/go-keychain/secretservice as new keychain helper
- GUI changes to report a problem tab

### Fixed
- Manual update mechanism


## v2.1.2
- 2022-03-29

### New
- Added another proxy provider
- Improved UX for working with keychain on macOS

### Fixed
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
- 2022-01-18

### New
- Dark Mode for Bridge, including autodetect mechanism for system colour scheme
- GUI element for changing keychain (Linux)
- Update to goopenpgp 2.4.1
- Optimising sentry reporting

### Fixed
- Bridge crashes related to unlocking local cache
- Bug with sending to 'non-encrypted' recipients
- Cosmetic GUI changes


## v2.0.1
- 2021-12-15

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

New local cache

- Refactor of local store (caching of size, headers and bodystructure)
- Allow to store full encrypted message bodies on a disk

### Fixed
- Improved retry mechanism for connecting to Proton servers
- OpenGL issue during startup for specific GPUs
- Blurry system icons with multiple monitor setup


## v1.8.12
- 2021-11-30

### New
- Bridge to only be checking and trying to unlock active keys, both user and address


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
