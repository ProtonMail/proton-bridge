# Proton Mail Bridge Changelog

Changelog [format](http://keepachangelog.com/en/1.0.0/)


## Colorado Bridge 3.13.0

### Added
* BRIDGE-37: added message broadcasting functionality.
* BRIDGE-122: added observability service.
* BRIDGE-119: added support for Feature Flags.
* BRIDGE-116: added command-line switches to enable/disable keychain check on macOS.
* BRIDGE-88: added context menu for quick actions on input labels: cut, copy, paste.

### Changed
* BRIDGE-81: KB article suggestion updates + more weight for long keywords.

### Fixed
* BRIDGE-67: Added detection for username changes on macOS & automatic reconfiguration.
* BRIDGE-138: Remove deprecated doc.


## Bastei Bridge 3.12.0

### Added
* BRIDGE-75: Bridge repair button.
* BRIDGE-79: Add New Outlook for Mac KB disclaimer.

### Changed
* BRIDGE-16: Bump version Go 1.21.9 Qt 6.4.3.
* BRIDGE-23: Update gluon to go 1.21.
* BRIDGE-22: Update gpa to go 1.21.

### Fixed
* BRIDGE-90: Disable repair button when bridge cannot connect to proton servers; bump GPA.
* BRIDGE-69: Explicitly handle semver panic for last bridge version from vault.
* BRIDGE-29: Bump gluon version.
* BRIDGE-49: Configure gitleaks baseline and grype config.
* BRIDGE-21: Missing panic handling.
* BRIDGE-17: Broken telemetry heartbeat test.
* BRIDGE-10: Bumped gluon version.


## Alcantara Bridge 3.11.1

### Fixed
* BRIDGE-70: Hotfix for blocked smtp/imap port causing bridge to quit.


## Alcantara Bridge 3.11.0

### Added
* GODT-3185: Report cases which leads to wrong address key used.

### Changed
* BRIDGE-14: HV3 implementation.
* BRIDGE-15: Certificate install is now also done during Outlook setup on macOS.
* GODT-3146: Start servers on startup, keep running even when no users are active.
* BRIDGE-19: Update checksum validation use warning instead of error on non-existing files.

### Fixed
* BRIDGE-8: Fix bridge double sessionID issue in logs.
* BRIDGE-7: Modify keychain test on macOS.
* BRIDGE-4: Logs not being created when invalid flag is passed.
* BRIDGE-5: Add tooltip to tray icon.
* GODT-3163: Filter MBOX format delimiter.


## Zaehringen Bridge 3.10.0

### Added
* GODT-3199: Add package log field.
* GODT-3220: Add more test scenarios.

### Changed
* GODT-3193: Preserve attachment encoding.
* GODT-3214: Encrypt only with primary key.
* GODT-2662: Use tart runner for darwin jobs.
* GODT-1602: Test: run integration tests against black 🖤.
* GODT-3257: Test: quad9 provider test not working on CI.

### Fixed
* GODT-3290: Fix test failing because of leap day.


## Ypsilon Bridge 3.9.1

### Fixed
* GODT-3235: Update bridge update key.


## Ypsilon Bridge 3.9.0

### Added
* GODT-3230: Scripts for removing Bridge from device.
* GODT-3195: Add OS info to the log.
* GODT-3156: Add time zone info to the bridge log.
* GODT-3162: Test: Add test scenarios for KB article suggestions.
* Test: Add scenarios for checking messages sent from Web Client.
* GODT-3162: Test: Add step definition for checking KB article suggestions.

### Changed
* GODT-3160: Bump version Go 1.21.6.
* GODT-3160: Load pipeline env from bridge internal.
* GODT-3052: Test: Replace attachments and inline content in feature tests with the smallest valid versions.
* GODT-3155: Customize log formatter for easier parsing.
* GODT-3172: Detect missing keychain item.
* GODT-3172: Do not list, just retrieve vault key.
* Log the message received time when handling message creation event.
* Set log as artefact for all integration test.
* Get better logging arround keychain list initialisation.

### Fixed
* GODT-3229: Escape reserved XML characters in Apple configuration profile.
* GODT-3228: Get rid of fork of docker-credential-helpers.
* GODT-3176: Assume inline if content id is present.
* GODT-3160: Ignore non-called vulnerabilities.
* GODT-3160: Updated external dependencies reported by govulncheck.
* GODT-3203: Crash in chunkDivide.
* Fix for SMTP connection mode toggle in bridge-gui-tester.
* GODT-3183: Fix database indices.
* GODT-3187: Fix numberOfDay computation when changing year and day.
* GODT-3188: Happy new year.


## Xikou Bridge 3.8.2

### Fixed
* GODT-3235: Update bridge update key.


## Xikou Bridge 3.8.1

### Added
* GODT-3121: Suggest relevant KB articles in the in-app bug report form.
* GODT-2001: Add govulncheck to scan for vulnerabilities.

### Changed
* Keep nighlty-job log as artifact.
* Test: Improve TestMetadata_JobCorrectlyFinishesAfterCancel.

### Fixed
* GODT-3153: Do not take into account full address when hasing messages.


## Xikou Bridge 3.8.0

### Added
* Test: Add test scenarios to add an /Answered flag to a replied message and revert.
* GODT-3046: Added links to KB in error messages.
* Test(GODT-3113): Inline HTML message and HTML attachment is getting altered.
* Test(GODT-3124): Attempt to fix 401 during login.

### Changed
* GODT-3134: Br tag triggers installer.
* Added update events to bridge GUI tester.

### Fixed
* GODT-3142: Pass br tag if available.
* GODT-3151: Fix feature test with non modified HTML part.
* GODT-3151: Only modify HTML Meta content if UTF-8 charset override is needed.
* GODT-2851: Add empty text part if no text part when importing multipart.
* GODT-3102: Distinguish Vault Decryption from Serialization Errors.
* GODT-3124: Handling of sync child jobs.
* GODT-3148: Bump go-sysinfo to get rid of linker warning on macOS Sonoma.
* GODT-3124: Flaky tests.
* GODT-3022: Handle multipart/related on fake server.
* GODT-3133: Fix GetSystemLanguage.
* GODT-3124: Race condition in sync task waiter.
* GODT-3124: Race conditions reported by race check.
* GODT-2797: Encode attached key name and use same pubkey name as web-app.
* Fix case of IMAP login error.
* GODT-3132: Do not allow sending on disabled accounts.
* GODT-3046: fix typo spotted during KB article review.
* GODT-3129: Bad Event during after address order change.
* GODT-3117: Improve GetAllContacts and GetAllContactsEmail.


## Wakato Bridge 3.7.1

### Added
* Test(GODT-2740): Sending Plain text messages to internal recipient.
* Test(GODT-2892): Create fake log file.
* GODT-3122: Added test, changed interface for accessing display name.

### Changed
* Remove debug prints.
* GODT-2576: Forward and $Forward Flag Support.
* GODT-3053: Use smaller bridge window on small screens.
* GODT-3113: Only force UTF-8 charset for HTML part when needed.
* GODT-3113: Do not render HTML for attachment.
* GODT-3112: Replaced error message when bridge exists prematurely. Added a link to support form.
* GODT-2947: Remove 'blame it on the weather' error part from go-smtp.
* GODT-3010: Log MimeType parsing issue.
* GODT-3104: Added log entry for cert install status on startup on macOS.
* GODT-2277: Move Keychain helpers creation in main.

### Fixed
* GODT-3054: Only delete drafts after message has been Sent.
* GODT-2576: Correctly handle Forwarded messages from Thunderbird.
* GODT-3122: Use display name as 'Email Account Name' in macOS profile.
* GODT-3125: Heartbeat crash on exit.
* GODT-2617: Validate user can send from the SMTP sender address.
* GODT-3123: Trigger bad event on empty EventID on existing accounts.
* GODT-3118: Do not reset EventID when migrating sync settings.
* GODT-3116: Panic on closed channel.
* GODT-1623: Throttle SMTP failed requests.
* GODT-3047: Fixed 'disk full' error message.
* GODT-3054: Delete draft create from reply.
* GODT-3048: WKD Policy behavior.


## Wakato Bridge 3.7.0

### Added
* Test(GODT-1224): Add testing around package creation.
* Add debug_assemble binary.
* Test(GODT-2723): Add importing a message with remote content.
* Test(GODT-2737): Sending HTML messages to internal.
* Test(GODT-3036): Keep inline attachment order on GPA Fake Server.
* GODT-3015: Add simple algorithm to deal with multiple attachment for bug report.
* Test: make message structure check more verbose.
* Test: Add test around account settings.

### Changed
* GODT-3097: Warn about PGPInline encryption scheme which will be deprecated.
* Test: Support multiple users when waiting for sync event.
* Test: Update fake server with defautl draft content-type and test it.
* Test: be less aggressive while checking for message structure.
* GODT-2996: Set password fields to hidden when resetting the login form.
* GODT-2990: Change runner tags.
* GODT-2835: Bump GPA adding support for AsyncAttachments for BugReport +...
* GODT-2940: Allow 3 attempts for mailbox password.
* GODT-3095: Update GOpenPGP.

### Fixed
* GODT-3106: Broken import route.
* GODT-3041: Fix Invalid Or Missing message signature during send.
* GODT-3087: Exclude attachment content-disposition part when determining...
* GODT-2887: Inline images with Apple Mail.
* GODT-3100: Fix issue where a fatal error that bubble up to cli.Run() is not written in the log file.
* GODT-3094: Clean up old update files on bridge startup.
* GODT-3012: Fix multipart request retries.
* GODT-2935: Do not allow parentID into drafts.
* GODT-2935: Correct error message when draft fails to create.
* GODT-2970: Correctly handle rename of Inbox.
* GODT-2969: Prevent duration corruption for config status event.
* Fixed type in QA installer CI job name.
* GODT-3019: Fix title of main window when no account is connected.
* GODT-3013: IMAP service getting "stuck".
* GODT-2966: Allow permissive parsing of MediaType parameters for import.
* GODT-2966: Add more test regarding quoted/unquoted filename in attachment.
* GODT-2490: Fix sync progress not being reset when toggling split mode.
* GODT-2515: Customized notification of unavailable keychain on macOS.


## Vasco da Gama Bridge 3.6.1

### Fixed
* GODT-3033: Unable to receive new mail.


## Umshiang Bridge 3.5.4

### Fixed
* GODT-3033: Unable to receive new mail.



## Vasco da Gama Bridge 3.6.0

### Added
* GODT-2762: Setup wizard.
* GODT-2772: Setup wizard content.
* GODT-2769: Setup Wizard architecture.
* GODT-2767: Setup Wizard foundations.
* GODT-2725: Implement receive message step with expected structure exposed.

### Changed
* GODT-2960: Added content in empty view when there is no account.
* GODT-2771: Cert related tools for macOS.
* GODT-2770: Proof of concept for web view as a tool window and overlay (not used).
* GODT-2916: Split Decryption from Message Building.
* GODT-2597: Implement contact specific settings in integration tests.
* GODT-2664: Trigger QA installer.

### Fixed
* GODT-2992: Fix link in 'no account view' in main window after 2FA or TOTP are cancelled.
* GODT-2989: Allow to send bug report when no account connected.
* GODT-2988: Fix setup wizard KB links.
* GODT-2968: Use proper base64 encoded string even for bad password test.
* GODT-2965: Fix multipart/mixed testdata + structure parsing steps related to this.
* GODT-2932: Fix syncing not being reported in GUI.
* GODT-2967: Tray menu entries close the setup wizard when needed.
* GODT-2212: Preserver Header order in message building.
* Fixed missing GoOs gRPC call in bridge-gui-tester.
* GODT-2929: Message dedup with different text transfer encoding.


## Umshiang Bridge 3.5.3

### Changed
* GODT-3004: Update gopenpgp and dependencies.


## Umshiang Bridge 3.5.2

### Fixed
* GODT-3003: Ensure IMAP State is reset after vault corruption.
* GODT-3001: Only create system labels during system label sync.


## Umshiang Bridge 3.5.1

### Fixed
* GODT-2963: Use multi error to report file removal errors.
* GODT-2956: Restore old deletion rules.
* GODT-2951: Negative WaitGroup Counter.
* GODT-2590: Fix send on closed channel.
* GODT-2949: Fix close of close channel in event service.


## Umshiang Bridge 3.5.0

### Added
* GODT-2734: Add testing steps to modify account settings.
* GODT-2746: Integration tests for reporting a problem.
* GODT-2891: Allow message create & delete during sync.
* GODT-2848: Decouple IMAP service from Event Loop.
* Add trace profiling option.
* GODT-2829: New Sync Service.
* Test: oss-fuzz support for fuzzing.
* GODT-2799: SMTP Service.
* GODT-2800: User Event Service.
* GODT-2801: Identity Service.
* GODT-2802: IMAP Serivce.
* GODT-2788: Add preview to bug report validation and JSON file validator.
* GODT-2803: Bridge Database access.

### Changed
* GODT-2909: Remove Timeout on event publish.
* GODT-2913: Reduce the number of configuration failure detected.
* GODT-2828: Increase sync progress report frequency.
* Test: Fix TestBridge_SyncWithOnGoingEvents.
* GODT-2871: Is telemetry enabled as service.
* Test(GODT-2873): Wait for Gluon Watcher to finish.
* Test(GODT-2744): Add integration tests for moving messages (with MOVE support).
* Test(GODT-2872): Fix nightly job.
* Test(GODT-2742): Add more integration tests regarding drafts.
* GODT-2787: Force Scrollview to top when re-opening questions set.
* GODT-2787: Tweaking Bug Report form with last Review.
* Ci(GODT-2717): Create a job that will run on schedule.
* GODT-2787: Fix vertical alignement on CategoryItem.
* GODT-2842: Implement Bug Report Fallback notification.
* Chore(GODT-2848): Simplify User Event Service.
* GODT-2808: Apply comment from Bug Report content review.
* Test(GODT-2743): Sync high number of messages.
* GODT-2814: Standalone Server Manager.
* GODT-2808: Initial list of categories and questions.
* GODT-2787: Replace the PathTracker by a more visual NavigationIndicator.
* GODT-2816: Wait until mandatory fields are filled then fill body and title.
* GODT-2794: Clear cached answers when report is sent.
* GODT-2793: Feed the bug report body with the answered questions.
* GODT-2791: Parse the Bug Report Flow description file and ensure forward compatibility (GODT-2789).
* GODT-2821: Display questions in one page.
* GODT-2786: Init bug report flow description file.
* GODT-2792: Implement display of question set for bug report.
* Use qmlformat on qml files, and removed deprecated tests.

### Fixed
* GODT-2828: Fix negative report time.
* GODT-2828: Fix sync progress report after restart.
* GODT-2867: Do not crash on timeout or context cancel.
* GODT-2693: Duplicate messages in sent folder.
* GODT-2867: Get attachment returns API error on network problem.
* GODT-2805: Ignore Contact Group Labels.
* GODT-2866: Add 429/5xx Retry to Event Service.
* GODT-2855: Fix for text overlapping in settings view.
* Test: Verify leaks at end of WithEnv.
* Test: Fix event registration in TestBridge_SyncWithOngoingEvents.
* Test: Fix deadlock in chToType.
* GODT-2865: Add error on failed unlock.
* GODT-2857: Do not check changed values in clear recent flag.
* GODT-2827: Restore ticker to event poller.
* Test: TestBridge_SendAddTextBodyPartIfNotExists eventually fix.
* GODT-2813: Write new vault to temporary file first.
* GODT-2807: Fix issue where sessionID would not be removed from command-line on restart by bridge-gui.
* GODT-2687: Tabs after header field colon.
* GODT-2764: Allow perma-delete for messages which still have labels.
* GODT-2693: Fix message appearing twice after sent.
* GODT-2781: Try to remove stale lock file before failing in checkSingleInstance.
* GODT-2780: Fix 'QSystemTrayIcon::setVisible: No Icon set' warning in bridge-gui log on startup.
* GODT-2778: Fix login screen being disabled after an 'already logged in' error.
* Fix typos found by codespell.
* GODT-2577: Answered flag should only be applied to replied messages.


## Trift Bridge 3.4.1

### Fixed
* GODT-2859: Trigger user resync while updating from 3.4.0 to 3.4.1.
* GODT-2833: Fix migration of message flags.
* GODT-2759: Use examine rather than select for fetching.


## Trift Bridge 3.4.0

### Added

### Changed
* Test: Add require.Eventually to TestBridge_UserAgentFromSMTPClient.
* Test: Add smtp-send utility.
* GODT-2759: Check for oprhan messages.
* GODT-2759: Add prompt to download missing messages for analysis.
* GODT-2759: CLI debug commands.
* Remove gRPC auto-generated C++ source files.
* Test: Force all unit test to use minimum sync spec.
* Test: Force sync limits to minimum with env variable.
* GODT-2691: Close logrus output file on exit.
* GODT-2522: New Gluon database layout.
* GODT-2678: When internet is off, do not display status dot icon for the user in the context menu.
* GODT-2686: Change the orientation of the expand/collapse arrow for Advanced settings.
* Test(GODT-2636): Add step for sending from EML.
* Log failed message ids during sync.
* GODT-2510: Remove Ent.
* Test(GODT-2600): Changing state (read/unread, starred/unstarred) of a message in integration tests.
* GODT-2703: Got rid of account details dialog with Apple Mail autoconf.
* GODT-2685: Update to bug report log attachment logic.
* GODT-2690: Update sentry reporting in GUI for new log file naming.
* GODT-2668: Implemented new log retention policy.
* Test(GODT-2683): Save Draft without "Date" & "From" in headers.
* GODT-2666: Feat(GODT-2667): introduce sessionID in bridge.
* GODT-2660: Calculate bridge coverage and refactor CI yaml file.
* Fix dependency_license script to handle dot formated version.

### Fixed
* GODT-2812: Fix rare sync deadlock.
* GODT-2822: Better handling 429 during sync and event loop.
* GODT-2763: Missing Answered flag on Sync and Message Create.
* GODT-2758: Fix panic in SetFlagsOnMessages.
* GODT-2578: Refresh literals appended to Sent folder.
* GODT-2753: Vault test now check that value auto-assigned is first available port.
* GODT-2522: Handle migration with unreferenced db values.
* GODT-2670: Allow missing whitespace after header field colon.
* GODT-2653: Only log when err is not nil.
* GODT-2680: Fix for C++ debugger not working on ARM64 because of OpenSSL 3.1.
* GODT-2675: Update GPA to applye togin-gonic/gin patch + update COPYING_NOTES.


## Stone Bridge 3.3.2

### Fixed
* GODT-2782: Filter all labels when doing perma delete check.


## Stone Bridge 3.3.1

### Changed
* GODT-2707: Set bridge-gui default log level to 'debug'.
* GODT-2674: Add more logs during update failed.
* GODT-2750: Disable raise on main window when a notification is clicked on Linux.
* GODT-2709: Remove the config status file when user is removed.
* GODT-2748: Log calls that cause main window to show, with reason.
* GODT-2705: Added log entries for focus service on client and server sides.
* GODT-2712: Feed config_status with user action while pending.
* GODT-2728: Remove the sentry report for gRPC event stream interruptions in bridge-gui.
* GODT-2715: Add Unitary test for configStatus event.
* GODT-2715: Add Functional test for configStatus telemetry event.
* Disable windows runner.
* GODT-2714: Apply PR comments.
* GODT-2714: Set Configuration Status to Failure and send Recovery event when issue is solved.
* GODT-2713: Send config_progress event once a day if the configuration is stucked in pending for more than a day.
* GODT-2711: Send config_abort event on User removal.
* GODT-2710: Send config success on IMAP/SMTP connection..
* GODT-2716: Make Configuration Statistics persistent.
* GODT-2709: Init Configuration status.
* Log errors on failed message Downloads.

### Fixed
* GODT-2774: Only check telemetry availability for the current user.
* GODT-2774: Add external context to telemetry tasks.
* GODT-2774: Add context to Authorize in `gluon.Connector`.
* GODT-2726: Fix Parsing of Details field in GPA error message.
* GODT-2708: Fix dimensions event format + handling of ReportClicked event.
* GODT-2756: Fix for 'Settings' context menu opening the 'Help' page.


## Stone Bridge 3.3.0

### Changed
* GODT-2653: Log API error details on Message import and send.
* GODT-2655: Display internal build time tag in log and GUI.
* Add error logs when messages fail to build during sync.
* GODT-2673: Use NoClient as UserAgent without any client connected and...
* GODT-2648: Make win build work on AWS machine.
* Disable building of bridgepp-test app in build script.
* GODT-2631: Bump go to 1.20.
* GODT-2639: Enhance sentry init log.
* GODT-2161: Auto-submit 2FA.
* Bump Gluon for GODT-2595, GODT-2634 and GODT-2619.
* Test: Fix TestBridge_Report.
* Extend the timeout for integration test form 20m to 30.
* Improve CPC code.
* GODT-2585: Only Start IMAP/SMTP once one user is loaded.
* GODT-2585: Server Manager.
* GODT-2585: Add CPC utility.
* GODT-2621: Display pop up warning when IMAP login fails because user is locked (connecting).
* Set default log level to Debug.
* GODT-2520: Update error message for free users.
* Test: Disable sync open files test.
* GODT-2346: Treat external address as disabled one.
* GODT-2610: Re-use previous password when removing and adding back account.
* GODT-2611: Bridge CLI exits on the first SIGINT / Ctrl+C.
* GODT-2540: Make icon loading failure behavior consistent.
* GODT-2540: Pop-up notification error icon is loaded on startup.
* GODT-2540: Notify user of wrong IMAP password.

### Fixed
* GODT-2683: Only validate messages that are not appended to Drafts.
* GODT-2683: Reduce message checks when appending into Drafts.
* Fix linter errors.
* GODT-2669: Display sentry ID in bridge init log.
* GODT-2672: Fix context cancelled when IMAP/SMTP parameters change is in progress.
* GODT-2650: Fix crash during header serialization.
* GODT-2437: Fix lint, test + bump Gluon with silenced report.
* GODT-2437: Silence harmless report to sentry.
* GODT-2649: Clean up cache files after failed connector create (Gluon).
* GODT-2638: Validate messages before import.
* GODT-2646: Bump GPA and Gluon dependency after CIRCL upgrade.
* GODT-2454: Only Send status update if transaction succeeded.
* Test: fix flaky tests.
* GODT-2628: Attempt to fix closed channel panic on logout.
* GODT-2627: Properly handle recording of message with Bcc fields.
* GODT-2627: Fix crash on closed channel.
* GODT-2307: Removed deprecated macOS security framework function.
* GODT-2637: Fix address parser error due to trailing separator.
* GODT-2635: Ensure Bridge can be compiled with GCC 13.
* GODT-2626: Handle rare crash due to missing address update ch.
* GODT-2626: Server Events should not be merged.
* GODT-2606: Improve Vault concurrency scopes.
* GODT-2623: Log IMAP/SMTP login failure as error.
* GODT-2527: Cleanup 503 test since handled by GPA.
* GODT-2613: Install the TLS certificate in the user keychain.
* GODT-2618: Crash when address does not have unlocked keyring.
* GODT-2616: Silence out of Order UID report.
* Update Gluon for async.Group.Do() fix.
* Upgraded golangci-lint v1.52.2 and fixed all issues.
* GODT-2464: Filter attachment name from content-type parameter to not send it twice to the API.


## [Bridge 3.2.0] Rialto

### Added
* GODT-2552, GODT-2553, GODT-2555, GODT-2556: Add telemetry.
* GODT-2575: Add dev info to cookies.

### Changed
* GODT-2598: Map Message Size Error to Gluon Error.
* GODT-2569: Support multiple externalID matching if we send one of it when looking for parentID.
* GODT-2576: Connector can send any flags to Gluon.
* GODT-2496: Bump gopenPGP to 2.7.1-proton.
* GODT-2517: Replace status window with native tray icon context menu.
* GODT-2586: Two-columns layout for account details.
* GODT-2580: Updated link to support website in GUI.
* GODT-2239: Bridgepp worker/overseer unit tests.
* GODT-2538: Implement smart picking of default IMAP/SMTP ports.
* GODT-2502: Improve logs.
* GODT-2551: Store and Recover Last User Agent from Vault.
* GODT-2550: Verify IMAP ID is set properly.
* GODT-2554: Compute telemetry availability from API UserSettings.
* Add missing double quotes in test.
* GODT-2239: Unit tests for BridgeUtils.cpp in bridgepp.
* Replace go-rfc5322 with gluon's rfc5322 parser.
* GODT-2483: Install cert without external tool on macOS.

### Fixed
* GODT-2625: Update Bridge pubkey for updates.
* GODT-2620: Avoid stalls in case of panic in gluon.
* GODT-2615: Remove keyboard shortcut for tray icon context menu on Windows and Linux.
* GODT-2596: Fix bug when trying to generate Sentry report and there is not log.
* GODT-1374: Fix tray icon DPI change handling.
* GODT-2589: Update BUILDS.md.
* GODT-2581: Update outdated link to bridge homepage in CLI 'manual' command.
* GODT-2337: Filter reply-to on draft.
* GODT-2550: Announce IMAP ID Capability.
* GODT-2574: Fix label/unlabel of large amounts of messages.
* GODT-2573: Handle invalid header fields in message.
* GODT-2573: Crash on null update.
* GODT-2407: Replace invalid email addresses with empty for new Drafts.

## [Bridge 3.1.3] Quebec

### Changed
* GODT-2616: Silence UID of order report.
* GODT-2614: Handle failed update during sync.


## [Bridge 3.1.2] Quebec

### Changed
* GODT-2582 Dedup recovered messages folder.

## [Bridge 3.1.1] Quebec

### Fixed
* GODT-2500: Fix handler passing.


## [Bridge 3.1.0] Quebec

### Changed
* GODT-2523: Use software QML rendering backend by default on Windows.
* GODT-2500: Reorganise async methods.
* GODT-2500: Add panic handlers everywhere.
* GODT-2511: Add bridge-gui switches to permanently select the QML rendering backend.
* GODT-2509: Migrate TLS cert from v1/v2 location during upgrade to v3.
* GODT-2487: Add windows test job and worker.
* Update GPA to include detailed error messages.
* GODT-2479: Ensure messages always have a text body part.
* GODT-2482: More attachment to relevant exceptions.
* GODT-2224: Refactor bridge sync to use less memory.
* GODT-2448: Supported Answered flag.
* GODT-2382: Added bridge-gui settings file with 'UseSoftwareRenderer' value.
* GODT-2411: Allow qmake executable to be named qmake6.
* GODT-2273: Menu with "Close window" and "Quit Bridge" button in main window.
* GODT-2261: Sync progress in GUI.
* GODT-2385: Gluon cache fallback.
* GODT-2366: Handle failed message updates as creates.
* GODT-2201: Bump Gluon to use pure Go IMAP parser.
* GODT-2374: Import TLS certs via shell.
* GODT-2361: Bump GPA to use simple encrypter.
* GODT-1264: Constraint on Scheduled mailbox in connector + Integration tests.
* GODT-1264: Creation and visibility of the 'Scheduled' system label.
* GODT-2283: Limit max import size to 30MB (bump GPA to v0.4.0).
* GODT-2352: Only copy resource file when needed.
* GODT-2352: Use go-build-finalize macro to build vault-editor for both mac arch.
* GODT-2278: Properly override server_name for go.
* GODT-2255: Randomize the focus service port.
* GODT-2144: Handle IMAP/SMTP server errors via event stream.
* GODT-2144: Delay IMAP/SMTP server start until all users are loaded.
* GODT-2295: Notifications for IMAP login when signed out.
* GODT-2278: Improve sentry logs.
* GODT-2289: UIDValidity as Timestamp.

### Fixed
* GODT-2505: Show notification only for cases when user needs to do actions.
* GODT-2516: Log error when the vault key cannot be created/loaded from the keychain.
* GODT-2526: Fix high memory usage with fetch/search.
* GODT-2514: Apply Retry-After to 503.
* GODT-2524: Preserve old vault values.
* GODT-2508: Handle Address Updated for none-existing address.
* GODT-2504: Fix missing attachments in imported message.
* GODT-2513: Scanner Crash in Gluon.
* GODT-2512: Catch unhandled API errors.
* GODT-2507: Memory consumption bug.
* GODT-2497: Do not report EOF and network errors.
* GODT-2481: Fix DBUS Secert Service.
* GODT-2455: Upper limit for number of merged events.
* GODT-2480: Do not override X-Original-Date with invalid Date.
* GODT-2473: Fix handling of complex mime types.
* GODT-2469: Fix sentry revision hash for cmake on windows.
* GODT-2424: Sync Builder Message Split.
* GODT-2419: Use connector.ErrOperationNotAllowed.
* GODT-2418: Ensure child folders are updated when parent is.
* GODT-1945: Handle disabled addresses correctly.
* GODT-2390: Add reports for uncaught json and net.opErr.
* GODT-2393: Improved handling of unrecoverable error.
* GODT-2394: Bump Gluon for golang.org/x/text DoS risk.
* GODT-2387: Ensure vault can be unlocked after factory reset.
* GODT-2389: Close bridge on exception and add max termination wait time.
* GODT-2201: Add missing rfc5322.CharsetReader initialization.
* GODT-1804: Preserve MIME parameters when uploading attachments.
* GODT-2312: Used space is properly updated.
* GODT-2319: Seed the math/rand RNG on app startup.
* GODT-2272: Use shorter filename for gRPC file socket.
* GODT-2318: Remove gluon DB if label sync was incomplete.
* GODT-2326: Only run sync after addIMAPUser().
* GODT-2323: Fix Expunge not issued for move.
* GODT-2224: Properly handle context cancellation during sync.
* GODT-2328: Ignore labels that aren't part of user label set.
* GODT-2326: Fix potential Win32 API deadlock.
* GODT-1804: Only promote content headers if non-empty.
* GODT-2327: Remove unnecessary sync when changing address mode.
* GODT-2343: Only poll after send if sync is complete.
* GODT-2336: Recover from changed address order while bridge is down.
* GODT-2347: Prevent updates from being dropped if goroutine doesn't start fast.
* GODT-2351: Bump GPA to properly handle net.OpError and add tests.
* GODT-2351: Bump GPA to automatically retry on net.OpError.
* GODT-2365: Use predictable remote ID for placeholder mailboxes.
* GODT-2381: Unset draft flag on sent messages.
* GODT-2380: Only set external ID in header if non-empty.


## [Bridge 3.0.21] Perth Narrows

### Added
* GODT-2509: Migrate TLS cert from v1/v2 location during upgrade to v3.

### Changed
* GODT-2516: log error when the vault key cannot be created/loaded from the keychain.

### Fixed
* GODT-2501: Remove additional .desktop file.
* GODT-2513: Crash in scanner.
* GODT-2481: Fix DBUS Secert Service.
* GODT-2512: Catch unhandled API errors.
* GODT-2469: Fix sentry revision hash for cmake on windows.


## [Bridge 3.0.20] Perth Narrows

### Added
* GODT-2442: Allow user to re-sync DB without logout.

### Changed
* GODT-2419: Reduce sentry reports.
* GODT-2458: Wait for both bridge and bridge-gui to be ended before restarting on crash.
* GODT-2457: Include address if GetPublickKeys() error message.
* GODT-2446: Attach logs to sentry reports for relevant bridge-gui exceptions.
* GODT-2425: Out of sync messages and read status.
* GODT-2435: Group report exception by message if exception message looks corrupted.
* GODT-2356: Unify sentry release description and add more context to it.
* GODT-2357: Hide DSN_SENTRY and use single setting point for DSN_SENTRY.
* GODT-2444: Bad event info.
* GODT-2447: Don't assume timestamp exists in log filename.
* GODT-2333: Do not allow modifications to All Mail label.
* GODT-2429: Do not report context cancel to sentry.

### Fixed
* GODT-2467: elide long email addresses in 'bad event' QML notification dialog.
* GODT-2449: fix bug in Bridge-GUI's Exception::what().
* GODT-2427: Parsing header issues.
* GODT-2426: Fix crash on user delete.
* GODT-2417: Do not request gluon recovered message from API.


## [Bridge 3.0.19] Perth Narrows

### Fixed
* GODT-2364: Wait and retry once if the gRPC service config file exists but cannot be opened.
* GODT-2364: Added optional details to C++ exceptions.
* GODT-2413: Use qEnvironmentVariable() instead of qgetenv().
* GODT-2412: Don't treat context cancellation as BadEvent.
* GODT-2404: Handle unexpected EOF.
* GODT-2400: Allow state updates to be applied if command fails.
* GODT-2399: Fix immediate message deletion during updates.
* GODT-2390: Missing changes from previous commit.
* GODT-2390: Add reports for uncaught json and net.opErr.
* GODT-2414: Multiple deletion bug in WriteControlledStore.


## [Bridge 3.0.18] Perth Narrows

### Fixed
* GODT-2392: Create message if gluon updateMessage returns `no such message`.
* GODT-2391: Create draft if missing during message update on gluon side.

## [Bridge 3.0.16/17] Perth Narrows

### Fixed
* GODT-2371: Continue, not return, when handling draft.

## [Bridge 3.0.15] Perth Narrows

### Changed
* GODT-2355: Improve wording and actions on bad event.

### Fixed
* GODT-2354: Report failed load users.
* GODT-2353: Show popup only after 3.0.16.
* GODT-2351: Bump GPA to better handle net.OpError.


## [Bridge 3.0.14] Perth Narrows

### Fixed
* GODT-2323: Fix Expunge not issued for move.
* GODT-2341: Handle URL error.
* GODT-2340: Improve logging.
* GODT-2278: Improve sentry logs.
* GODT-2327: Sync issues when migrating DB.
* GODT-2318: Remove gluon DB if label sync was incomplete.
* GODT-1804: Only promote content headers if non-empty.
* GODT-2343: Only poll after send if sync is complete.
* GODT-2336: Recover from changed address order while bridge is down.



## [Bridge 3.0.13] Perth Narrows

### Fixed
GODT-2328: Ignore labels that aren't part of user label set.
GODT-2326: Sync issue on missing fresh DB file.
GODT-2319: Seed the math/rand RNG on app startup.
GODT-1804: Preserve MIME parameters when uploading attachments.


## [Bridge 3.0.12] Perth Narrows

### Added
* GODT-2210: v3.0 splash screen.
* GODT-1770: handle UserBadEvent in CLI and gRPC.

### Changed
* GODT-2311: Fix missing headers in re-downloaded Gluon messages.
* GODT-1453: clicking 'Sign in' from status window now selects the right account.
* GODT-2297: More significantly improve GPA's paging algorithm.
* GODT-2145: Fix button spacing w/ Qt 6.4.
* GODT-2223: Improve event handling.
* GODT-2305: Detect missing gluon DB.
* GODT-2291: Change gluon store default location from Cache to Data.
* Other: Disable dialer test until badssl cert is bumped.
* GODT-2292: Updated BUILDS.md doc.
* GODT-2258: suggest email as login when signing in via status window.
* Other: Report corrupt and/or insecure vaults to sentry.
* Other: Better user load logs.
* GODT-2253: Restart Launcher from the gui when GUI crashes.
* Other(test): Make All Mail copy test more robust.
* Other(CI): Make race checks manual.
* Other: Remove old cert/key file location handling.
* GODT-2271: Update README with new system files path.

### Fixed
* GODT-2210: Fix splash screen always showing on CentOS and Ubuntu.
* GODT-2296: Log error rather than fail if cannot get parent ID.
* GODT-2266: Pause event stream while sending.
* GODT-2266: Add test for sent message flags.
* Other(test): Fix some more integration test placeholders.
* GODT-2177: Use correct attachment disposition when content ID is set.
* GODT-1556: If no references, use the in-reply-to header as ParentID.
* Other: make GUI Tester more resilient to Bridge abrupt termination.
* GODT-2275: fixed location of bridge-gui log files.
* Other: Ensure SMTP debug dump works on windows.
* Other: Fix MaxLogs off-by-one limit and bump limit to 10.
* Other: fix path of temp folder in README.
* Other(debug): Dump raw SMTP input to user's home dir.


## [Bridge 3.0.11] Perth Narrows

### Changed
* GODT-2252: Recover from deleted cached messages.
* GODT-2258: change login label and suggest email instead of username.
* Other: Don't clean settings path on teardown.
* Other: Bump GPA to v0.3.0.
* Other: added user's primary email address to the vault.
* GODT-2251: gluon store and DB separated.
* GODT-2093: use the primary email address in the account view and status view.
* GODT-2202: Report update errors from Gluon.
* GODT-2229: Own the full path for gluon and do not change Database path.
* GODT-1797: copyright notice shows a date range with the build year.

### Fixed
* GODT-2223: Handle bad events by logging user out.
* GODT-2165: Reduce UTF8 parsing errors from TLS header input.
* Others: chores fix a QML warning when no account is present* and a few typos in QML.
* Other(test): Fix integration test steps.
* GODT-2226: Fix moving drafts to trash.
* GODT-2246: do not report API error 422 when using an invalid email address.


## [Bridge 3.0.10] Perth Narrows

### Changed
* GODT-2205: use lock file in bridge-gui to detect orphan bridge.
* GODT-2242: Bump GPA - Don't send any 2fa information if not needed.
* GODT-2179: added handler for exceptions in QML backend methods.
* GODT-2181: Match live API behaviour.
* GODT-2221: Set DOH off by default.
* GODT-1817: Re-enable all integration tests.
* Other: C++ Code reformat.
* GODT-2234: added command-line switch to force Qt to use software rendering for QML.
* Other: added C/C++ header template file (*.h.in) type to missing_license.sh script.
* GODT-2236: add log entry when SMTP / IMAP serve method fails.
* Other: reorganised QMLBackend class code.

### Fixed
* Other: Flag messages imported into "Sent" mailbox as Sent.
* Other: Fix testCtx.getMBoxID().
* Other: Fixed GUI Tester to comply with latest gRPC changes.
* GODT-2010: add Cocoa app delegate handler for second application instance.
* Other: Fix double close on event channels.
* GODT-2233: Fix sub folder creation bug.
* GODT-2222: Dot not error on unknown Address Events.
* GODT-2218: Fix invalid UID ranges.


## [Bridge 3.0.9] Perth Narrows

### Changed
* GODT-2181(test): Refactor integration test setup a bit.
* Other: Updated GUI tester for new gRPC calls.
* GODT-1847: Add option to export TLS Certificates in GUI.

### Fixed
* Other: Fix TOTP login (bump go-proton-api).
* GODT-2188: Do not fail append with invalid mime-type.
* GODT-2213: Don't unnecessarily enable/disable autostart.
* Other: Do not decode message body during send record hashing.
* GODT-2196: Do not generate message updates for unknown labels.
* Other: Prevent double login.
* Other: Improve migration logging prefers username over primary address.
* Other(test): Prefer native API revoke rather than fake server method.
* GODT-2190: Unify crashpad_handler for darwin.
* Other(test): Add test that we skip and report bad messages during sync.
* Other: Catalina build.
* GODT-2042: Fix setup guide not always showing on first login.
* GODT-2152: Sign-in dialog validate email and password only when button is pressed.
* GODT-1556: Add unit test for in-reply-to header without references.
* GODT-2150: Fixed initial implementation that filtered --no-window in gui instead of bridge.
* GODT-2167: Bind sign-in buttons availability to loading state.
* Other: Only send to necessary update channel.
* GODT-1804: Add parsing ics attachment test.
* Other: Fix Warning introduced by connecting check timer.


## [Bridge 3.0.8] Perth Narrows

### Fixed
* Other: Add sentry reports for event processing failures.
* Other: Do not fail on label events.


## [Bridge 3.0.7] Perth Narrows

### Fixed
* Other: Increase default UIDVALIDITY.
* GODT-2173: fix: Migrate Bridge passwords from v2.X.
* GODT-2207: Fix encoding of non utf7 mailbox names.
* Other: Increase worker count (2 -> 4).


## [Bridge 3.0.6] Perth Narrows

### Fixed
* GODT-2187: Skip messages during sync that fail to build/parse.


## [Bridge 3.0.5] Perth Narrows

### Fixed
* GODT-2178: Bump go-proton-api to fix drafts.
* GODT-2180: Allow login with FIDO2.


## [Bridge 3.0.4] Perth Narrows

### Changed
* Other: Do not list \Deleted flag for All Mail.
* Other: Disable perma-delete for expunge on Spam folder.

### Fixed
* Other: Ensure expunge feature test pushes to error stack.
* GODT-2170: Use client-side draft update in integration tests.
* GODT-2170: Improving test server behaviour.
* GODT-2170: Update draft event means delete old and create new message.
* GODT-2170: User create draft route: first steps.


## [Bridge 3.0.3] Perth Narrows

### Fixed
* GPA v0.1.4: fix token expiration mechanism.

## [Bridge 3.0.2] Perth Narrows

### Changed
* GODT-2157: Add Sentry to Bridge-Gui.
* GODT-2153: Use file socket for bridge gRPC on linux & macOS.
* GODT-2150: Do not forward --no-window flag.
* GODT-2154: Allow noninteractive mode from launcher.
* Other: update gui tester to support latest changes in gRPC implementation.
* Other: GUI Tester supports the 3 states of user (Signed out/Locked/Connected).
* Other: Bump gluon version to drop non-UTF-8 commands.

### Fixed
* Other: Wipe vault properly on factory reset.
* GODT-2160: Prevent double closing of bridge if restart fails.
* GODT-2041: Crash after factory reset.
* GODT-2114: Sanitize attachment disposition.
* GODT-1910: Fix GUI not being notified of SMTP SSL being turned on by ConfigureAppleMail.
* GODT-1910: Fix save button state not being updated after being clicked once.
* GODT-2159: Improve 429 retry.
* GODT-1989: Handle Move with Append and Expunge.
* Other: SetMailServerSettings is async as it should.
* Other: Include sentry dll for Windows deploy.
* Other: Ensure context is string in sentry reports.
* GODT-2160: Ensure we can safely move cache file.


## [Bridge 3.0.1] Perth Narrows

### Changed
* GODT-2151: Sync backwards to please product people.

### Fixed
* GODT-2149: Sort logs by timestamp when clearing.
* GODT-2137: Set sentry sync transport.

## [Bridge 3.0.0] Perth Narrows

### Changed
* Other(chore): Bump major version to v3.
* Other: Switch from liteapi to go-proton-api.
* GODT-2085: Ensure minimum sync worker count.
* Other: Switch to mail-api.proton.me.
* GODT-2120: Encrypt gluon store with gzip.
* GODT-1910: Use a single view for IMAP & SMTP SSL options.
* GODT-1846: Remove restart cues* implement restart-less behaviour. 
* GODT-1975: Migrate keychain secrets.
* GODT-1976: Migrate app settings from prefs.json.
* GODT-2100: Load users in parallel at startup.
* GODT-2108: Implement C++ Focus gRPC service client in bridge-gui.
* GODT-2091: Animated "Connecting..." label.
* GODT-2003: Introduces 3 phases user state (SignedOut/Locked/Connected).
* GODT-2056: Kill old bridge from v2 lock file.
* GODT-2086: Changing the wording for signing in.
* GODT-2070: Implement SASL login for SMTP.
* Other: Use liteapi instead of pmapi.
* GODT-1609: Store password as byte array.
* GODT-1650: Gluon integration.
* GODT-1779: Remove go-imap.

### Fixed
* Other: Retry sync after cooldown if it fails.
* GODT-2142: Also permit split by comma in References header.
* GODT-2085: Use time.Since* structured logging.
* GODT-2139: Validate key pass during login.
* GODT-2111: Fix restart.
* GODT-2085: Revise sync algorithm.
* GODT-2134: Fix dock icon on macOS when launched with '--no-window'.
* GODT-2131: If refresh token is revoked* user gets signed out.
* GODT-2119: Only show supported label IDs to clients.
* GODT-2002: Wait for API events to be applied after send.
* GODT-2105: Ensure ClientVersion is set in bug report request.
* GODT-2107: Update user list after session revoke.
* GODT-2040: Bump UID validity when clearing sync status.
* GODT-2045: Timeouts should be considered network issues.
* GODT-2122: Handle check for updates failure.
* GODT-2033: Only set user agent from IMAP ID if not empty.
* GODT-2110: Force attachment disposition if content ID is missing.
* GODT-2081: If keychain cannot be loaded do not wipe Vault and use a temp one.
* GODT-2103: Trigger the version changed event.
* GODT-2047: Clear last event ID when clearing sync status.
* GODT-2109: Removed log message "Parent process XXX is still alive".
* GODT-1913: Pass reporter to gluon* limit restarts* add crash handlers.
* GODT-2037: Handle and log API refresh event.
* GODT-2029: Handle deadlock when reordering user addresses.
* GODT-2021: Remove gluon data when deleting user.
* GODT-2030: Rework deletion check on expunge.
* GODT-1977: Fix launcher for v2 to v3 updates.
* GODT-2048: Add missing special use attributes.
* GODT-2034: Basic vault migration ability (proof of concept).
* GODT-1978: Auto-updates from v2 to v3.
* GODT-1954: Draft message support.
* GODT-2022: Fix change between address modes.
* GODT-1993: Use more efficient filtering for message deletion.
* GODT-2004: Ensure log files don't have color formatting.
* GODT-2011: Use new app version format.
* GODT-2010: Add better logging for app focus feature.
* GODT-2008: Ensure user's addresses are returned in sorted order.
* GODT-2002: Poll after SMTP send.
* GODT-1982: Updated gRPC and GUI for disk cache.
* GODT-1984: Handle permanent message deletion.
* GODT-1916: Use XDG_DATA_HOME to store persistent data on linux.
* GODT-1974: Store everything in v3 path.
* GODT-1986: Handle case where an address has no decryption entities.
* GODT-1777: Message de-duplication in SMTP and IMAP.
* GODT-1940: Fix message encryption.
* GODT-1742: Implement hide All Mail.
* GODT-1813: Cleanup old go-imap cache files.
* GODT-1650: Implement Connector.CreateMessage.
* GODT-1901: Allow to set IMAP SSL from UI.
* GODT-1816: Connect Gluon Logs to bridge Logs.
* GODT-1657: Stable sync.
* GODT-1815: Gluon User management error.

## [Bridge 2.4.8] Osney

### Fixed
* GODT-2071: Fix --no-window flag that was broken on Windows.

## [Bridge 2.4.7] Osney

### Fixed
* GODT-2078: Launcher inception.
* GODT-2039: fix --parent-pid flag is removed from command-line when restarting the application.

## [Bridge 2.4.6] Osney

### Changed
* GODT-2019: When signing out and a single user is connected we do not go back to the welcome screen.
* GODT-2071: Bridge-gui report error if an orphan bridge is detected.
* GODT-2046: Bridge-gui log is included in optional archive sent with bug reports.
* GODT-2039: Bridge monitors bridge-gui via its PID.
* GODT-2038: Interrupt gRPC initialisation of bridge process terminates.
* Other: Added timestamp to bridge-gui logs.
* GODT-2035: Bridge-gui log includes Qt version info.
* GODT-2031: Updated bridge description.

### Fixed
* Other: Fix make run-qt target for Darwin.

## [Bridge 2.4.5] Osney

### Changed
* GODT-2015: Bridge-gui logs to file until gRPC connection is established.
* GODT-2016: Added more logging of gRPC events at info level.
* GODT-2013: CLI flag for frontend is required.

### Fixed
* GODT-2020: Fix xdg_{home,cache}_home variables.
* GODT-2014: Bridge quit if gRPC client ends stream.

## [Bridge 2.4.4] Osney

### Changed
* GODT-1751: Switch from protonmail.com to proton.me domain.

### Fixed
* Other: Fix make run-cli for Darwin.
* GODT-1645: Fix CI pipeline.
* GODT-1938: Account details box values wrap.
* Other: Also install vcpkg ARM64 on Intel mac hosts.
* Other: Fix minor typo.
* GODT-1939: removed vertical overshoot when scrolling.
* GODT-1479: fix 'Open Bridge' button still hovered when status windows opens for Windows.
* GODT-1519: Move back to account view after sending bug report.
* Other: fix QML error with Qt 6.4 and a typo.

## [Bridge 2.4.3] Osney

### Changed
* Other: implemented tokens in bridge-gui-tester.
* GODT-1853:
    * Upgrade dependencies (including x/crypto).
    * Ignore for CVE-2021-33194 false positive + add several try to gobinsec.
* GODT-1853: Improve pipeline:
    * Update gobinsec cache.
    * Test-windows branch manual, MR always.
    * Build needs test-linux and lint to start and keep vcpkg cache on linux.
    * Builds manuall except linux-qa.
* GODT-1893: Bridge-gui sends bridge's log to stdout, stderr.
* GODT-1932: Frontend is instantiated before bridge.
* GODT-1929: Changed gRPC wait timeout.
* Other: gRPC TLS server is generated for every session.
* GODT-1917: gRPC service should use random port.
* GODT-1924: gRPC identity validation with tokens.
* GODT-1344: Notifications for ApiCertError and NoActiveKeyForRecipient.
* GODT-1941: Update documentation.
* Other: Update golangci-lint to v1.50.0.
* GODT-1936: check gRPC server token via interceptors.



### Fixed
* GUI issues:
    * GODT-1894: Fixed typo in alreadyLoggedIn event error message.
    * GODT-1479: Fix hover on “Open Bridge” in status window on macOS.
    * GODT-1899: Status window menu now closes when window is dismissed.
    * GODT-1851: Port field error label now wraps.
    * GODT-1566: GUI shows error notifications for IMAP/SMTP port errors on startup.
    * GODT-1926: Clear port error messages when cancelling the dialog.
    * Other: Fixed cocoa related warnings in bridge-gui on macOS.
* Build issues:
    * GODT-1675: Add resrource file to both launcher and bridge-go.
    * Other: Add WlShellIntegration lib for rpm package.
    * GODT-1935: Fix resource file generation for both Launcher and Bridge.
    * GODT-1942: Use `qmake` to find the `QT6DIR`.
    * Provide launcher for make run-cli target.
* GODT-1931: Fixed bridge crash when checking for update while offline.

## [Bridge 2.4.0] Osney

### Added
* GODT-1551: Upgrade to Qt 6:
    * Change the app architecture.
    * Drop therecipe/qt dependency.
    * Update to go1.18.
    * Update to Qt 6.3.2.
* GODT-1170 GODT-1675: Native Mac M1 release.


## [Bridge 2.3.0] Nihonbashi

### Added
* GODT-1739: Opt-out All Mail visibility in settings file.
    * GODT-1794: CLI wording.
    * GODT-1794: Add confirmation dialog and change wording.
    * GODT-1741: GUI and CLI settings to change visibility of All Mail folder.
    * GODT-1740: Opt-out All Mail visibility in settings file.

### Changed
* GODT-1737: Improve logging during import.
* GODT-1754: Add logs for unilateral updates and SEARCH.

### Fixed
* GODT-1840: Use Safe map for mailboxID cache.
* GODT-1795: Fix automatic installation of profile for AppleMail on macOS Ventura beta (qt 5).
* GODT-1833: Fix gobinsec cache.
* GODT-1799: Fix dependency link.
* Other: Update SSL certificate fingerprint for test.


## [Bridge 2.2.2] Millau

### Added
* Introduced gobinsec cache.

### Fixed
* GODT-1743: Terminate running bridge if has old version.
* GODT-1743: Quit bridge when opening manual install.


## [Bridge 2.2.1] Millau

### Added
* GODT-1550: Add gobinsec check after CI build.
* GODT-1686: Add Label/Folder filtering to pmapi.

### Changed
* Rebranding:
    * GODT-1475: Change systray icons.
    * GODT-1565: Update Bridge application icons.
    * GODT-1564: Update welcome illustration.
* GODT-1626: Update gopenpgp v2.4.7.
* GODT-1627: Update go-srvp to v0.0.5.
* GODT-1523: Reduce unnecessary shell executions. Inspired by @kortschak.
* Other: Add v2 to module name.
* GODT-1562: Update test fingerpring.

### Fixed
* GODT-1659: Convert charset only for `text/*` MIME types.
* GODT-1640: Renew test keys.


## [Bridge 2.2.0] Millau

### Added
* Rebranding:
    * GODT-1508: Splash screen for rebranding.
    * GODT-1542: Update login screen for rebranding.
    * GODT-1260: Renaming.
    * GODT-1502: Rebranding: color and radius.
* GODT-1549: Add notification when address list changes.
* GODT-1560: Dependency licenses update and link.

### Changed
* GODT-1543: Using one buffered event for off and on connection.
* GODT-1550: Update dependencies.
* GODT-1545 GODT-1521: Change wording and enable release notes link.

### Fixed
* GODT-1534: Reset address when leaving split mode.


## [Bridge 2.1.3] London

### Added
GODT-1525: Add keybase/go-keychain/secretservice as new keychain helper.

### Changed
GODT-1527: Change bug report description.

### Fixed
GODT-1537: Manual in-app update mechanism.


## [Bridge 2.1.2] London

### Added
* GODT-1522: Rebuild macOS keychain notification.
* GODT-1437 Add new proxy provider (Quad9 with port).
* GODT-1516: Return notification on missing keychain.

### Changed
* GODT-1451: Do not check for gnome keyring to allow other implementations of secret-service API. Thanks to @remgodow.
* GODT-1516 GODT-1451: KeepassXC is crashing on start. We need to block it until it's fixed.

### Fixed
* GODT-1524: Logout issues with macOS.
* GODT-1503 GODT-1492: Improve email validation and username in bug report.
* GODT-1507: Enable autostart after Qt setup.
* GODT-1515: Do not crash when bridge users got disconnected.


## [Bridge 2.1.1] London

### Added
* GODT-1376: Add first userID to sentry scope.
* GODT-1375: Add host architecture to sentry reports.
* GODT-1364: Add windows CI machine for tests, and build.

### Fixed
* GODT-1499: Remove message from DB once it is not on server any more.


## [Bridge 2.1.0] London

### Fixed
* GODT-1482: Comment or mitigate panics, unlock cache when needed.
* GODT-1481: Always turn off non-encrypted recipient report.

### Changed
* GODT-1494: Update GopenPGP to 2.4.1.
* GODT-1490: Update go sentry and openpgp.
* GODT-1474: Optimising live integration tests.
* GODT-1483: Correct scope in sentry report.
* GODT-1477: Change CoD wording.

### Added
* GODT-1478: Add GUI settings for keychain selection.
* Other: Change copyright year.
* GODT-1329: Dark mode, with autodetect.


## [Bridge 2.0.1] Kwai

### Fixed
* GODT-1468: Fix main windows status and add background context without retry.


## [Bridge 2.0.0] Kwai

### Added
* GODT-22: New GUI style and improved UX:
    * GODT-1168 GODT-1169 Qml artifacts for preview.
    * GODT-1177: Remove Import-Export from repo.
    * GODT-1167 GODT-1179 Make run-qml-preview.
    * GODT-1051: Add factory reset to bridge object.
    * GODT-1179 GODT-658: Components and login flows.
    * GODT-1051: Factory reset button.
    * GODT-1158: Adding cache on disk signals.
    * GODT-1298: Signal GUI is ready and rise window.
    * Other: Reactive show on startup.
    * GODT-1319: Set sourceSize everywhere for images.
    * GODT-1317 Use large png for systray and mark it as mask.
    * GODT-1346: GODT-1340 GODT-1315 QML changes.
    * GODT-1365: Create ComboBox component.
    * GODT-1338: GODT-1343 Help view buttons.
    * GODT-1340: Not crashing, user list updating in main thread.
    * GODT-1345: Adding panic handlers.
    * GODT-1271: Fix Status margins.
    * GODT-1320: Add loading property to each action within a notification.
    * GODT-1210: Add "free user" banner.
    * GODT-1314: Limit description field length within 150/800 bounds.
    * GODT-1250: Fix Port settings wording.
    * GODT-1369: Fix link render and wording in Help view.
    * GODT-1358: Fix wording.
    * GODT-1272: Fix status view layout.
    * GODT-1336: Fix showing window on startup.
    * GODT-175: Add option to attach logs for bug reports.
    * GODT-1272: Ultimate fix for MacOS transparency.
    * GODT-1384: Fix SettingsView scroll.
    * GODT-1385: Fix port setting.
    * GODT-1378: varia GUI fixes.
    * GODT-1390: Fix autostart toggle.
    * GODT-1251: Fix change SMTP settings.
    * GODT-1389: Fix buttons and banner layout.
    * GODT-1316: Set default TextArea and TextField behavior.
    * GODT-1244: Refactor switching stable-early and factory reset.
    * GODT-1351: Cache and update of space bytes in user object.
    * GODT-1351: Fix used size update from mail operations.
    * GODT-1411: refactor SettingView content to fill height.
    * GODT-1327: Reset cache path to default when disabling.
    * GODT-1412: Refactor paths and links.
    * GODT-1226: Fix status window position.
    * GODT-1366: Simple lookup of index and select current user.
    * GODT-1325: Add "already logged in" notification.
    * GODT-1391: Fix link colors across GUI.
    * GODT-1391: Fix color for avatar text.
    * GODT-1442: Fix "Sign In" button.
    * GODT-1428: Fix welcome illustration by using PNG.
    * GODT-1455 Adding links to setup guide.
    * GODT-1456: Make text selectable and clickable.
    * GODT-1459: Wording.
    * GODT-1460 GODT-1462: Adding delete account dialog and fixing status view brief and icon.
    * GODT-1458: Splash screen and wording.
* GODT-1158: Caching encrypted full body messages on disk:
    * GODT-1433: Do not save message to cache if it's a draft.
    * GODT-1431 Do not cache message during new message event when CoD is off.
    * GODT-1381 Treat readonly folder as failure for cache on disk.
    * GODT-1431 Prevent watcher when not using disk on cache.
    * GODT-1381: Use in-memory cache in case local cache is unavailable.
    * GODT-1356 GODT-1302: Cache on disk concurrency and API retries.
    * GODT-1332 Added tests for cache move functions.
    * GODT-1332: moved cache related functions to separate file.
    * GODT-1332 moving cache does not work on Windows.
    * GODT-1367: use waitgroup instead of channel in pool/pchan.
    * GODT-1367: Use sync.Once to only close pool jobs once.
    * GODT-1349: Change cache-related settings when enabling/disabling/moving cache.
    * GODT-1350: stop cacher/worker properly when logging out user.
    * GODT-1158: Store full messages bodies on disk.
* GODT-1433 Adding first integration test for drafts.

### Changed
* GODT-1438: Turn off SW OpenGL on windows and add debug info about graphic renderer.
* GODT-1425: Factory reset enables launch on startup.
* GODT-1433 Message.Type is deprecated, use Flags instead.
* GODT-1388: Refactor Alternative routing.


## [Bridge 1.8.12] James

### Fixed
* GODT-1432: Check if keys are active before unlocking.


## [Bridge 1.8.11] James

### Fixed
* GODT-1415: Only messages which are in Spam should be moved to INBOX once they are marked as not-a-spam.
* GODT-1405: Integration test fix: Prevent unilateral update in FETCH when copying message by append.
* GODT-1392: Fix broken header fields for attachments.
* GODT-1360: Fix live integration test.
* GODT-968: Messages in All Mail should not be able to mark as deleted.
* GODT-967: Append external message to All Mail should be APPEND to Archive instead.
* GODT-966: Append internal message to AllMail should be no action.
* GODT-965: MOVE command should end with error for All Mail.
* GODT-963: STORE removing junk or adding nojunk should move message to inbox.

### Changed
* GODT-1397: Update bbolt to v1.3.6.
* GODT-1410: Remove event ID from sentry report description.
* GODT-1395: CI should fail on go.sum changed.


## [Bridge 1.8.10] James

### Fixed
* GODT-1348: Max 100 conn per host.
* GODT-1204: Handle importing too big messages.
* GODT-1202: Do not update package if it's version older than launcher.
* GODT-1318: Bump gopenpgp to v2.2.2, go-srp to v0.0.1, go-crypto to 52430bf6.
* GODT-219: Update to godog v0.12.1.
* GODT-1205: "RCPT TO" does not contain all addressed from "CC".
* GODT-1103: Cleanup on windows when uninstalling Bridge.


## [Bridge 1.8.9] James

### Fixed
* GODT-1263: Fix crash on invalid or empty header.
* GODT-1235: Fix 401 response error handling.
* GODT-1261: Fix building messages with long key.
* Other: use windows-compatible filename when dumping message in QA builds.


## [Bridge 1.8.8] James

### Changed
* GODT-1234 Set attachment name 'message.eml' for `message/rfc822` attachments.


## [Bridge 1.8.7] James

### Changed
* GODT-1201: Update gopenpgp to 2.1.10.

### Fixed
* GODT-1193: Do not doubly encode parts.


## [Bridge 1.8.6] James

### Removed
* GODT-1187: Remove IMAP/SMTP blocking when no internet.

### Changed
* GODT-1166: Reduce the number of auth for live test.

### Fixed
* GODT-1193: Do not use message.Read permit non-UTF-8 charsets.


## [Bridge 1.8.5] James

### Fixed
* GODT-1189: Draft created on Outlook is synced on web.
* GODT-1190: Fix some random crashes of Bridge on Windows.
* GODT-1191: Fix data loss of some drafts messages when restarting outlook on Windows.

## [Bridge 1.8.4] James

### Added
* GODT-1155: Update gopenpgp v2.1.9 and use go-srp.
* GODT-1044: Lite parser for appended messages.
* GODT-1183: Add test for getting contact emails by email.
* GODT-1184: Preserve signatures in externally signed messages.

### Changed
* GODT-949: Ignore some InvalidMediaParameter errors in lite parser.

### Fixed
* GODT-1161: Guarantee order of responses when creating new message.
* GODT-1162: Fix wrong section 1 error when email has no MIME parts.


## [Bridge 1.8.3] James

### Fixed
* GODT-1182: Use correct contact route.


## [Bridge 1.8.2] James

### Fixed
* GODT-1175: Bug reporting.


## [Bridge 1.8.1] James

### Fixed
* GODT-1165: Handle UID FETCH with sequence range of empty mailbox.


## [Bridge 1.8.0] James

### Added
* GODT-1056 Check encrypted size of the message before upload.
* GODT-1143 Turn off SMTP server while no connection.
* GODT-1089 Explicitly open system preferences window on BigSur.
* GODT-35: Connection manager with resty.

### Fixed
* GODT-1159 SMTP server not restarting after restored internet.
* GODT-1146 Refactor handling of fetching BODY[HEADER] (and similar) regarding trailing newline.
* GODT-1152 Correctly resolve wildcard sequence/UID set.
* GODT-876 Set default from if empty for importing draft.
* Other: Avoid API jail.


## [Bridge 1.7.1] Iron

### Fixed
* GODT-1081 Properly return newlines when returning headers.
* GODT-1150 Externally encrypted messages with missing private key would not be built with custom message.
* GODT-1141 Attachment is named as attachment.bin in some cases.


## [Bridge 1.7.0] Iron

### Added
* GODT-213 New message builder:
* Preserve Content-Type for undecryptable message body.
* Use application/octet-stream for encrypted parts.
* Force no transfer encoding for embedded message/rfc822 parts.
* Remove dead code GetRelatedHeader/GetRelatedBoundary.
* Correctly expect text/plain in custom message text parts.
* Force text/plain for custom message text part.
* Complex external encrypted tests (multipart/alternative, message/rfc822 attachment).

### Fixed
* GODT-1136 DB Cache header from builder and test.
* GODT-1113 Fix tray icon size on macOS Big Sur.
* GODT-947 Force colors in logs.


## [Bridge 1.6.9] HZM

### Fixed
* GODT-1121 'Keep the application up to date' switches off after restarting Bridge.


## [Bridge 1.6.8] HZM

### Fixed
* GODT-1120 Use Info level in internal/app logs.


## [IE 1.3.3] Farg

### Fixed
* GODT-1120 Use Info level in internal/app logs.


## [Bridge 1.6.7] HZM

### Added
* GODT-1111 Add correct metadata to Windows executables.
* GODT-1112 Add application to Windows Firewall exclusion list on install.
* GODT-1077 Track how many times message is built to help understand re-syncs.

### Changed
* GODT-247 Revise all storage locations (cache, config, local etc).

### Fixed
* GODT-948 Parser does not handle embedding of Content-Type: message/rfc822.
* GODT-1079 Correct 9001 error handling on login.

### Security
* GODT-1105 Dylib Hijacking security fix.


## [IE 1.3.2] Farg

### Added
* GODT-1111 Add correct metadata to Windows executables.
* GODT-1112 Add application to Windows Firewall exclusion list on install.

### Changed
* GODT-247 Revise all storage locations (cache, config, local etc).

### Fixed
* GODT-1079 Correct 9001 error handling on login.

### Security
* GODT-1105 Dylib Hijacking security fix.


## [IE 1.3.1] Farg

### Changed
* GODT-1047 No silent updates for Import-Export app.
* GODT-247 Cache and update files moved from user's cache to config.

### Fixed
* Other: include latest go.mod/go.sum changes.
* GODT-803 Fix import to wrong target address.
* GODT-948 Embedded messages.
* GODT-1043 Fix showing long login error in GUI dialog.


## [Bridge 1.6.6] HZM

### Added
* Other: QA build checks for update every 5 minutes.
* Other: QA build adds debug message dump when sending.

### Changed
* GODT-1045 build without Qt by default.

### Fixed
* GODT-1029 Fix tray icon not updating under certain conditions.
* GODT-1062 Fix lost notification bar when window is closed.
* GODT-1058 Install version after changing channel right away only in case of downgrade.
* GODT-1073 Re-write autostart link on every start if turned on in preferences.
* GODT-1055 Fix flaky empty trash test.


## [Bridge 1.6.5] HZM

### Changed
* GODT-1059 Check if keychain is usable on linux before using it by default.


## [Bridge 1.6.4] HZM

### Added
* Other: Autoupdates CLI commands.

### Removed
* Other: Remove credits.

### Changed
* GODT-980 Placeholder for user agent.
* GODT-1036 Event loop Sentry reporting of failures and refresh.
* GODT-957 Increase space to hide difference.
* GODT-937 Add keychain switcher to frontend.
* GODT-1008 Fix transparent dialog under certain conditions.
* GODT-1034 More tolerant connection speed detection.
* GODT-1018 Pre-push git hook to check lints.
* Other: Make all command line flags as const strings.
* GODT-1041 Log IMAP requests to debug Apple Mail re-sync issue.
* Other: Pretty print prefs.json.

### Fixed
* Other: Fix nogui build.
* GODT-317 Fix wrong total mailbox size in Apple Mail.
* Other: Fixing changelog punctuation.
* GODT-797 APPEND waits for EXPUNGE to prevent data loss when Outlook moves from Spam or Trash.


## [Bridge 1.6.3] HZM

### Added
* GODT-337 Desktop files.

### Changed
* GODT-885 Do not explicitly unlabel folders during move to match behaviour of other clients.
* GODT-616 Better user message about wrong mailbox password.
* GODT-1021 Do not allow copy Inbox->Sent or Sent->Inbox.
* GODT-976 Exclude updates from clearing cache and clear cache, including updates, while switching early access off.
* GODT-1033 Retry starting IMAP server after connection was down.

### Fixed
* GODT-1011 Stable integration test deleting many messages using UID EXPUNGE.
* GODT-1015 Use lenient version parser to properly parse version provided by Mac.
* GODT-919 Notify about update right after the start.
* GODT-919 GODT-1022 Logs and signals.


## [IE 1.3.0] Farg

### Changed
* GODT-1019 Remove dependency on go-apple-mobileconfig.
* GODT-928 Reject messages which are too large.
* GODT-999 Sending: do not send empty objects to API.

## [Bridge 1.6.2] HZM

### Fixed
* GODT-1010 Strip angle brackets from external ID.

## [Bridge 1.6.1] HZM

### Added
* GODT-1007 Notify user when version is the latest.

### Fixed
* GODT-787 GODT-978 Fix IE and Bridge importing to Sent not showing up in Inbox (setting up flags properly).
* GODT-1006 Use correct macOS keychain name.
* GODT-1009 Set ContentID if present and not explicitly attachment.
* GODT-1008 Transparent welcome message.


## [Bridge 1.6.0] HZM

### Added
* GODT-705 Allow silent update in Bridge and Import-Export app.
* GODT-958 Release notes per eaach update channel.
* GODT-875 Added GUI dialog on force update.
* GODT-820 Added GUI notification on impossibility of update installation (both silent and manual).
* GODT-870 Added GUI notification on error during silent update.
* GODT-805 Added GUI notification on update available.
* GODT-804 Added GUI notification on silent update installed (prompt to restart).
* GODT-275 Added option to disable autoupdates in settings (default autoupdate is enabled).
* GODT-874 Added manual triggers to Updater module.
* GODT-851 Added support of UID EXPUNGE.

### Removed
* GODT-248 Remove dependency on go-appdir.
* GODT-208 Remove deprecated use of BuildNameToCertificate.

### Fixed
* Check deprecated status code first to better determine API error.
* GODT-831 Fix reporting bug from accounts with empty account name.
* GODT-831 Cancel request of uploading attachment if reading/writing it fails.
* GODT-991 Fix panic when stopping import progress during loading mailboxes info.
* GODT-895 Fix panic when modifying addresses during changing address mode.
* GODT-946 Fix flaky tests notifying changes.
* GODT-979 Fix panic when trying to parse a multipart/alternative section that has no child sections.
* GODT-900 Remove \Deleted flag after re-importing the message (do not delete messages by moving to local folder and back).

### Changed
* Rename channels `beta->early`, `live->stable`.
* Bump gopenpgp dependency to v2.1.3 for improved memory usage.
* GODT-97 Don't log errors caused by SELECT "".
* GODT-806 GUI dialog on manual update. Added autoupdates checkbox. Simplifyed installation process GUI.
* GODT-912 Scroll bar behaviour in settings tab.
* GODT-149 Send heartbeat ASAP on each new calendar day.
* GODT-792 Stop IMAP server while no internet connection.
* GODT-792 Cache message size every time to reduce network traffic.
* GODT-792 Cache body structure in order to reduce network traffic.
* GODT-792 GODT-908 Cache body structure in order to reduce network traffic.
* GODT-908 Do not unpause event loop if other mailbox is still fetching.


## [Bridge 1.5.7] Golden Gate

### Fixed
CSB-331 Fix sending error due to mixed case in sender address.

## [Bridge 1.5.6] Golden Gate

### Added
* GODT-797 EXPUNGE waits for APPEND to prevent data loss when Outlook moves from Spam to Inbox.


## [Bridge 1.5.5] Golden Gate

### Changed
* GODT-922 Fix panic during restarting the bridge.
* GODT-945 Fix panic in integration tests caused by concurrent map writes.
* GODT-732 Fix usage of fontawesome.
* GODT-951 Properly parse message with long lines in header and long header split to multiple lines (upgrading to latest go-message).
* GODT-894 Fix panic when sending while account is logging in.
* GODT-858 Bump go-rfc5322 dependency to v0.5.0 to handle some invalid RFC5322 groups and add support for semicolon delimiter in address-list.
* GODT-923 Fix listener locking.
* GODT-389 Prefer `From` header instead of `MAIL FROM` address.
* GODT-898 Only set ContentID for inline attachments.
* GODT-773 Replace `INTERNALDATE` older than birthday of RFC822 by birthday of RFC822 to not crash Apple Mail.
* GODT-927 Avoid to call API with empty label name.
* GODT-915 Bump go-imap dependency and remove go-imap-specialuse dependency.

## [Bridge 1.5.4] Golden Gate

### Added
* Log warning about permanently deleting messages.

### Fixed
* License path on Arch and Windows.

## [Bridge 1.5.3] Golden Gate [Import-Export 1.2.3] Elbe

### Added
* GODT-906 Handle RFC2047-encoded content transfer encoding values.
* GODT-887 Make supports build with native Qt.

### Changed
* GODT-893 Bump go-rfc5322 dependency to v0.2.1 to properly detect syntax errors during parsing.
* GODT-892 Swap type and value from sentry exception and cut panic handlers from the traceback.
* GODT-854 EXPUNGE and FETCH unilateral responses are returned before OK EXPUNGE or OK STORE, respectively.
* #109 Renamed COPYING.md to not be read by [pkg-go-dev](https://pkg.go.dev/license-policy).

### Removed
* GODT-651 Build creates proper binary names.
* GODT-148 Allow import (using the Import-Export app) of already encrypted messages as is.
* GODT-202 Update to latest go-smtp.

### Fixed
* GODT-135 Support parameters in SMTP `FROM MAIL` command, such as `BODY=7BIT`, or empty value `FROM MAIL:<>` used by some clients.
* GODT-338 GODT-781 GODT-857 GODT-866 Flaky tests.
* GODT-773 Replace old dates with birthday of RFC822 to not crash Apple Mail. Original is available under `X-Original-Date` header.

## [Bridge 1.5.2] Golden Gate

### Changed
* GODT-883 Use `ClearPacket` for `text/plain` with signature.


## [Bridge 1.5.1] Golden Gate

### Added
* GODT-701 Try load messages one-by-one if IMAP server errors with batch load
  and not interrupt the transfer.
* GODT-878 Tests for send packet creation logic.

### Changed
* GODT-651 Build creates proper binary names.
* GODT-878 Fix an issue where the random session key is inadvertently sent to
  the Proton server. The data payload is always encrypted within TLS, but this
  is still a potential privacy problem. Discovered by Proton's internal
  security audit team.
* GODT-878 Refactor and move the send packet creation logic to `pmapi.SendMessageReq`.
* GODT-878 Encryption of session keys moved to pmapi.


## [IE 1.2.1, 1.2.2] Elbe

### Added
* GODT-799 Skipped messages do not change total counts but shows as separate number.

### Fixed
* GODT-799 Fix skipping unwanted folders importing from mbox files.
* GODT-769 Close connection before deleting labels to prevent panics accessing deleted bucket.

### Removed
* GODT-766 Remove GUI popup for IMAP TLS error.


## [Bridge 1.5.0] Golden Gate

### Changed
* Updated go-mbox dependency back to upstream.

### Fixed
* GODT-847 Waiting for unilateral update during deleting the message.
* GODT-849 Show in error counts in the end also lost messages.
* GODT-835 Do not include conversation ID in references to show properly conversation threads in clients.
* GODT-685 Improve deb packaging regarding dejavu font.


## [IE 1.2.0] Elbe

### Added
* GODT-763 Detect Gmail labels from All Mail mbox export (using X-Gmail-Label header).
* GODT-834 Info about tags in BUILDS.md and link to Import-Export page in README.md.
* GODT-777 Support Apple Mail MBOX export format.
* GODT-731 Re-open Import-Export app from the second instance.

### Fixed
* GODT-677 Windows IE: global import settings not fit in window.
* GODT-794 Congo fails to update to Danube.
* GODT-749 Don't force PGP/Inline when sending plaintext messages.
* GODT-764 Fix deadlock in integration tests for Import-Export.
* GODT-662 Do not resume paused transfer progress after dismissing cancel popup.
* GODT-772 Sanitize mailbox names for exporting to follow OS restrictions.
* GODT-771 Show fatal errors after export is terminated.
* GODT-779 Do not propagate updates when progress is stopped.
* GODT-779 Unpause progress during fatal error to properly stop progress.
* GODT-779 Stop ongoing transfer calls sooner (re-check after import request is generated).
* Fix measurement of uploading attachments during transfer.
* GODT-827 Do not spam sentry with bad ID by integration test.
* GODT-700 Fix UTF-7 incompatibility.
* GODT-837 Fix flaky TestFailUnpauseAndStops.
* GODT-782 Don't use TLS pinning when checking connectivity status.

### Changed
* TLS pins conform to official list.


## [Bridge 1.4.5] Forth

### Fixed
* GODT-829 Remove `NoInferior` to display sub-folders in apple mail.

## [Bridge 1.4.4] Forth

### Fixed
* GODT-798 Replace, don't add, transfer encoding when making body 7-bit clean.
* Move/Copy duplicate for emails with References in Outlook.
* CSB-247 Cannot update from 1.4.0.


## [Bridge 1.4.3] Forth

### Changed
* Reverted sending IMAP updates to be not blocking again.

### Fixed
* GODT-783 Settings flags by FLAGS (not using +/-FLAGS) do not change spam state.


## [Bridge 1.4.2] Forth

### Changed
* GODT-761 Use label.Path instead of Name to partially support subfolders for webapp beta release.
* GODT-765 Improve speed of checking whether message is deleted.


## [IE 1.1.2] Danube (beta 2020-09-xx)

### Fixed
* GODT-770 Better handling of extraneous end-of-mail indicator.
* GODT-776 Fix crash when IMAP client connects while account is logging in.
* GODT-744 User agent not being sent to sentry.

### Changed
* Bump crypto version to v0.0.0-20200818122824-ed5d25e28db8.
* GODT-785 Clear separation of different message IDs in integration tests.
### Changed
* GODT-741 Import-Export shows "Unable to parse time" notice instead of zero time in error report window.

* Bump crypto version to v0.0.0-20200818122824-ed5d25e28db8.
* GODT-374 Allow to send calendar update multiple times.

## [IE 1.1.1] Danube (beta 2020-09-xx) [Bridge 1.4.1] Forth (beta 2020-09-xx)

### Fixed
* GODT-752 Parsing message with empty addresses.
* GODT-752 Parsing non-utf8 multipart/alternative message.
* GODT-752 Parsing message with duplicate charset parameter.


## [IE 1.1.0] Danube

### Fixed
* GODT-703 Import-Export showed always at least one total message.
* GODT-738 Fix for mbox files with long lines.
### Fixed
* GODT-732 Do not mix font awesome icon with regular text to avoid issues on Fedora.


## [Bridge 1.4.0] Forth

### Added
* GODT-682 Persistent anonymous API cookies for Import-Export.
* GODT-357 Use go-message to make a better message parser.
* GODT-720 Time measurement of progress for Import-Export.
* GODT-693 Launcher.

### Changed
* GODT-511 User agent format changed.
* Unsilent errors reading mbox files.
* GODT-692 QA build with option to change API URL by ENV variable.
* GODT-704 User agent detected by fake IMAP extension instead of AUTH callback (some clients use LOGIN instead of AUTH).
* GODT-695 Parallel upload for ProtonMail target.

### Removed
* GODT-519 Unused AUTH scope parsing methods.

### Fixed
* GODT-698 Use correct package type for signed PGP/Inline messages.
* Generic bug report window title.
* Fix missing check for unencrypted recipients during sending.
* Version checking for catalina.
* GODT-730 Limit maximal TLS version for Yahoo IMAP server.


## [IE 1.0.x] Congo (v1.0.0 live 2020-09-08)

### Added
* GODT-633 Persistent anonymous API cookies for better load balancing and abuse detection. 
* GODT-461 Add support for `\Deleted` flag.

### Changed
* GODT-462 Pausing event loop while FETCHing to prevent EXPUNGE.
* Wait for unilateral response to be delivered.
* GODT-409 Set flags have to replace all flags.
* GODT-531 Better way to add trusted certificate in macOS.
* Bumped golangci-lint to v1.29.0.
* GODT-549 Check log file size more often to prevent huge log files.
* Bumped various dependencies:
    * Updated andybalholm/cascadia v1.1.0 -> v1.2.0.
    * Updated emersion/go-imap-specialuse 20161227184202-ba031ced6a62 -> 20200722111535-598ff00e4075.
    * Updated emersion/go-sasl 20191210011802-430746ea8b9b -> 20200509203442-7bfe0ed36a21.
    * Updated github.com/go-resty/resty/v2 v2.2.0 -> v2.3.0.
    * Updated github.com/golang/mock v1.4.3 -> v1.4.4.
    * Updated github.com/google/go-cmp v0.4.0 -> v0.5.1.
    * Updated github.com/hashicorp/go-multierror v1.0.0 -> v1.1.0.
    * Updated github.com/jaytaylor/html2text 20200220170450-61d9dc4d7195 -> 20200412013138-3577fbdbcff7.
    * Updated github.com/jhillyerd/enmime v0.8.0 -> v0.8.1.
    * Updated github.com/keybase/go-keychain 20200218013740-86d4642e4ce2 -> 20200502122510-cda31fe0c86d.
    * Updated github.com/logrusorgru/aurora 20200102142835-e9ef32dff381 -> v2.0.3+incompatible.
    * Updated github.com/miekg/dns v1.1.29 -> v1.1.30.
    * Updated github.com/nsf/jsondiff 20190712045011-8443391ee9b6 -> 20200515183724-f29ed568f4ce.
    * Updated github.com/sirupsen/logrus v1.4.2 -> v1.6.0.
    * Updated github.com/stretchr/testify v1.5.1 -> v1.6.1.
    * Updated github.com/therecipe/qt 20200126204426-5074eb6d8c41 -> 20200701200531-7f61353ee73e.
    * Updated github.com/urfave/cli v1.22.3 -> v1.22.4.
    * Updated golang.org/x/net 20200301022130-244492dfa37a -> 20200707034311-ab3426394381.
    * Updated golang.org/x/text v0.3.2 -> v0.3.3.
* Set first-start to false in bridge, not in frontend.
* GODT-400 Refactor sendingInfo.
* GODT-513 Update routes to API v4.
* GODT-551 Do not ignore errors during message flagging.
* GODT-380 Adding IE GUI to Bridge repo and building.
    * BR: extend functionality of PopupDialog.
    * BR: makefile APP_VERSION instead of BRIDGE_VERSION.
    * BR: use common logs function for Qt.
    * BR: change `go.progressDescription` to `string`.
    * IE: Rounded button has fa-icon.
    * IE: `Upgrade` → `Update`.
    * IE: Moving `AccountModel` to `qt-common`.
    * IE: Added `ReportBug` to `internal/importexport`.
    * IE: Added event watch in GUI.
    * IE: Removed `onLoginFinished`.
    * Structure for transfer rules in QML.
* GODT-213 Convert panics from message parser to error.
* GODT-585 Do not allow deleting messages from All Mail.

### Fixed
* GODT-655 Fix date picker with automatic Windows DST.
* GODT-454 Fix send on closed channel when receiving unencrypted send confirmation from GUI.
* GODT-597 Duplicate sending when draft creation takes too long.
* GODT-634 Hover on links in popups.



## [Bridge 1.3.x] Emma (v1.3.2 beta 2020-08-04, v1.3.3 beta 2020-08-06, v1.3.3 live 2020-08-12)

### Added
* GODT-554 Detect and notify about "bad certificate" IMAP TLS error.
* IMAP mailbox info update when new mailbox is created.
* GODT-72 Use ISO-8859-1 encoding if charset is not specified and it isn't UTF-8.

### Changed
* GODT-360 Detect charset embedded in html/xml.
* GODT-354 Do not label/unlabel messages from `All Mail` folder.
* GODT-388 Support for both bridge and import/export credentials by package users.
* GODT-387 Store factory to make store optional.
* GODT-386 Renamed bridge to general users and keep bridge only for bridge stuff.
* GODT-312 Validate recipient emails in send before asking for their public keys.
* GODT-368 Bump docker-credential-helpers version.
* GODT-394 Don't check SMTP message send time in integration tests.
* GODT-280 Migrate to gopenpgp v2.
    * `Unlock()` call on pmapi-client unlocks both User keys and Address keys.
    * Salt is available via `AuthSalt()` method.
* GODT-308 Better user error message when request is canceled.
* GODT-162 User Agent does not contain bridge version, only client in format `client name/client version (os)`.
* GODT-258 Update go-imap to v1.
    * Fix UNSEEN to return sequence number of first unseen message and not count of unseen messages.
    * INBOX name is never quoted.
* GODT-204 `ClientManager`.
    * `Client` is now an interface; `client` is the concrete type.
    * `Client`s are only created by `ClientManager`.
    * Only one `Client` per userID exists at any given time; clients are reused.
    * Tokens are managed by `ClientManager` (`TokenManager` is removed).
    * `expiresAt` is no longer part of `Client`; token expiry and refreshing is handled by `ClientManager`.
    * Auths generated by clients during Auth/AuthRefresh are handled by `ClientManager` (which forwards them to `Bridge`).
    * `ClientManager` is the "one source of truth" for the host URL for all `Client`s.
    * Alternative Routing is enabled/disabled by `ClientManager`.
    * Logging out of `Clients` is handled/retried asynchronously by `ClientManager`.
* GODT-265 Alternative Routing v2 (more resiliant to short term connection drops).
* GODT-310 Alternative parsing of `References` header (old parsing probably malformed message IDs).
* GODT-320 Only report the same TLS issue once every 24 hours.
* GODT-468 Bump go-imap version to get fix for NIL client delimiter.
* GODT-465 Bump go-imap version to get fix for SELECT function.
* GODT-456 Bump bbolt version from 1.3.3 to 1.3.5 to get fixes for unsafe operations.

### Removed
* Dead code from `pkg/message`.

### Fixed
* GODT-356 Fix crash when removing account while mail client is fetching messages (regression from GODT-204).
* GODT-358 Bad timeouts with Alternative Routing.
* GODT-363 Drafts are not deleted when already created on webapp.
* GODT-390 Don't logout user if AuthRefresh fails because internet was off.
* GODT-341 Fixed flaky unittest for Store synchronization cooldown.
* Crash when failing to match necessary html element.
* Crash in message.combineParts when copying nil slice.
* Handle double charset better by using local ParseMediaType instead of mime.ParseMediaType.
* Don't remove log dir.
* GODT-422 Fix element not found (avoid listing credentials, prefer getting).
* GODT-404 Don't keep connections to proxy servers alive if user disables DoH.
* Ensure DoH is used at startup to load users for the initial auth.
* Issue causing deadlock when reloading users keys due to double-locking of a mutex.
* Correctly handle failure to unlock single key.
* GODT-479 Fix flaky integration tests.
* GODT-484 Fix infinite loop when decoding invalid 2231 charset.
* GODT-267 Correctly detect if a message is a draft even if does not have DraftLabel.
* GODT-308 Reduce minimum read speed threshold to avoid issues with flaky internet.
* GODT-321 Changing address ordering would cause all messages to disappear in combined mode.
* GODT-129 Fix custom message PGP by using template.
* GODT-280 Don't assume contact keys are stored armored.
* GODT-427 Fix race condition in auth refresh that could cause user to be logged out.


## [Bridge 1.2.8] Donghai-fix-append  (beta 2020-06-XXX)

### Changed
* GODT-396 reduce number of EXISTS calls.
* GODT-143 Allow appending to Sent folder when sender matches account address.

### Fixed
* GODT-502 Fixed crash when unable to parse a message header.

## [Bridge 1.2.7] Donghai-fix-sync - (beta 2020-05-07 live 2020-04-20)

### Added
* IMAP extension MOVE with UIDPLUS support.
* IMAP extension Unselect.
* More logs about event loop activity.

### Changed
* GODT-313 Reduce number of synchronizations.
    * Do not trigger sync by counts.
    * Cooldown timer for sync retries.
    * Poll interval randomization.
* GODT-225 Do not send an EXISTS reposnse after EXPUNGE or when nothing changed (fixes rebuild of mailboxes in Outlook for Mac).
* GODT-165 Optimization of RebuildMailboxes.
* GODT-282 Completely delete old draft instead moving to trash when user updates draft.
* Adding DSN Sentry as build time parameter.
* GODT-124 Bump go-appdir from v1.0.0 to v1.1.0.
* CSB-72 Skip processing message update event if http statuscode is 422.

### Fixed
* Use correct binary name when finding location of addcert.scpt.


## [Bridge 1.2.6] Donghai - beta (2020-03-31)

### Added
* GODT-145 Support drafts.
    * GODT-211,GODT-231 fix updating subject and other fields.
    * GODT-220 Fix deleting drafts.
    * GODT-224 Fix creating draft from outlook without sender.
    * GODT-230,GODT-232 fix constructing sender address for drafts.
    * Sync already synced draft to newly created drafts mailbox.
    * Add Subject to EventMessageUpdated in pmapi.
* GODT-37 Add body and TLS handshake timeouts.
* GODT-90 Implement DOH (DNS over HTTPS) proxy.
* Noninteractive mode.


### Changed
* Bump version go-1.14.
* Bump dependencies:
| Repo                               | Old Version                        | New Version                        |
| github.com/0xAX/notificator        | v0.0.0-20161214074916-82e921414e03 | v0.0.0-20191016112426-3962a5ea8da1 |
| github.com/ProtonMail/go-autostart | v0.0.0-20171017232241-85d98b097aae | v0.0.0-20181114175602-c5272053443a |
| github.com/abiosoft/ishell         | v0.0.0-20171224170712-50251d04cb42 | v2.0.0+incompatible                |
| github.com/emersion/go-sasl        | v0.0.0-20161116183048-7e096a0a6197 | v0.0.0-20191210011802-430746ea8b9b |
| github.com/fatih/color             | v1.7.0                             | v1.9.0                             |
| github.com/golang/mock             | v1.4.2                             | v1.4.3                             |
| github.com/google/go-cmp           | v0.3.1                             | v0.4.0                             |
| github.com/jaytaylor/html2text     | v0.0.0-20190408195923-01ec452cbe43 | v0.0.0-20200220170450-61d9dc4d7195 |
| github.com/jhillyerd/enmime        | v0.7.0                             | v0.8.0                             |
| github.com/logrusorgru/aurora      | v0.0.0-20190803045625-94edacc10f9b | v0.0.0-20200102142835-e9ef32dff381 |
| github.com/skratchdot/open-golang  | v0.0.0-20160302144031-75fb7ed4208c | v0.0.0-20200116055534-eef842397966 |
| github.com/stretchr/testify        | v1.4.0                             | v1.5.1                             |
| github.com/therecipe/qt            | v0.0.0-20191022233421-590f404884c9 | v0.0.0-20200126204426-5074eb6d8c41 |
| github.com/urfave/cli              | v1.19.1                            | v1.22.3                            |

* Pkg/updates: closing File reader to avoid too many opened files during update.
* Created monorepo with bridge, pmapi, bridge utils, mime and srp.
    * One lint config for all packages and lint fixes in the code.
    * Fix tests for bridge utils to work on MacOS.
    * All tests use testify framework.
    * Processed TODOs or created issues.
    * Cleanup up comments.
* GODT-169 Reduce the number of keyring unlocks.
* CSB-40 return error instead of panic in credential store.
* #577 Avoid multiple send.
* GODT-39 Sync is paging per message ID with ability to continue after interrupted sync.
* Panic handler used in store for event loop and sync.
* GODT-109 Merge only 50 events into one.
* Use v1.0.16 of pmapi.
* GODT-236 Requests to /messages/{read,unread,delete,undelete,label,unlabel} are paged with up to 100 message IDs.

### Fixed
* GODT-227 Mitigate potential crash due to using a logged out pmapi client (proper fix to come in emma release).
* UserIDs were not checked when importing to Sent folder (affects copying from account1/sent to account2/sent).


## [Bridge 1.2.5] Charles - live (2020-03-11) beta (from 2020-02-10)

### Added
* GODT-112 Migration of preferences from c10 to c11.
* GODT-100 Test for external internal ID when appending to Sent folder to return APPEND UID otherwise skip with no error.
* GODT-43 Connection troubleshooting modal.
* GODT-55 Uid support in fake api.
* GODT-88 Increase uid validity on switch mode.
* #951 Implementation of IMAP extension UIDPLUS.
* #964 New store package, see Changed section.

### Removed
* MOVE IMAP extension due to missing interaction with UIDPLUS.

### Changed
* GODT-88 Run mbox sync in parallel when switch password mode (re-init not user).
* GODT-95 Do not throw error when trying to create new mailbox in IMAP root.
* GODT-75 Do not fail on unlabel inside delete.
* #1095 always delete IMAP USER including wrong password.
* Unique pmapi client userID (including #1098).
* Using go.enmime@v0.6.1 snapshot.
* Better detection of non-auth-error.
* Reset `hasAuthChannel` during logout for proper login functionality (set up auth channel and unlock keys).
* Allow `APPEND` messages without parsable email address in sender field.
* #1060 avoid `Append` after internal message ID was found and message was copied to mailbox using `MessageLabel`.
* #1049 Basic usage of store in SMTP package to poll event loop during sending message.
* #1050 pollNow waits for events to be processed.
* #1047 Fix fetch of empty mailbox.
* #1046 Fix removing mailbox counts.
* #1048 For any message build error return custom message.
* When event loop exits with error it logs out user from Bridge.
* #953 #984 First label messages before unlabeling when moving messages.
* Fixes after refactor:
    * Slight memory optimization.
    * #1043 do not stuck in loop for updating message which does not exist anywhere anymore.
    * #1034 fix UID dynamic range for empty list.
    * Fix of sequence number in IMAP IDLE expunge during deleting messages.
    * #1030 #1028 Fix some crashes and bad auths.
    * #953 #984 label messages first during moving them.
* #964 (and #769,#899,#918,#930,#931,#949) refactor of IMAP.
    * Fix of sequence number in IMAP IDLE expunge during deleting messages.
    * Fix not-returning empty result for UID dynamic range as said in RFC.
    * Separated IMAP to store and IMAP.
    * Store is responsible for everything about db and calls to pmapi, including event loop, sync, address mode.
    * IMAP is responsible only for IMAP interfaces.
    * Event loop is only one per ProtonMail account (instead of one per alias).
    * It also means only one database per account (instead of one per address).
    * Changing address mode is not destroying database, only buckets with IDs mapping (keeping metadata for account).
    * Before first sync we set event ID so we will not miss changes happening during sync.
    * Thanks to previous point we are not starting new sync when we finish first one because of unprocessed events.
    * Sync is not blocking event loop (user can get new messages even during sync).
    * Sync is not blocking reading operations (user can list mailboxes even before first sync is done).
    * Sync is not blocking writing operations such as mark messages read/unread and so on.
    * Most operations have to be passed to API and only event loop is writing them to the database.
    * Avoid relying on counts API endpoint; use event counts as much as possible.
    * Separate function for storing message content type and header into database.
    * Sequence number optimised for last item in mailbox.
    * Allow sending IMAP idle update to timeout to avoid blocking bridge.
    * Synchronisation will create a label if not yet present.
    * Labels and Folders (including system folders) are stored in DB together with their counts for offline read-out.
    * AddressIDs for all user addresses are stored in DB.
    * IMAP updates channel is set when an IMAP client connects (and IMAP updates are dropped until then).
    * DB keeps track of address mode (split/combined).
* Event loop starts as soon as user is initialised (i.e. logged in), not just when imap is connected.
* Use pmapi v1.0.13.
* Logout user if initialisation fails.
* Send UserRefreshEvent on user login and logout.
* Use godog v0.8.0 under new name 'cucumber' (instead of DATA-DOG).

### Fixed
* CSB-40 panic in credential store.
* Keyring unlocking locker.
* No panic on failed html parse.
* Too many open files.
* #1057 Logging in to an already logged in user would display unrelated error "invalid mailbox password".
* #1056 Changing mailbox password sometimes didn't log out user.
* #1066 Split address mode can not work when credentials store is cleared.
* #1071 Bridge can think it is in combined mode when actually it's in split mode.
* Missing `enmime` dependency.
* Issue where a failed sync was not attempted again.
* Removing an address would crash bridge.
* #1087 Accounts with capital letters could not be added.
* #1087 Inactive addresses were not filtered out of the store.
* #1087 Unlock with correct key if message is sent to alias and not primary (aka original) address.
* #1109 Receiving an event referencing an address that isn't present could crash bridge.
* Avoid concurrent map writes in imap backend.
* GODT-103 User keys were not unlocked later if they were not unlocked during startup.


## [Bridge 1.2.4] Brooklyn beta (2019-12-16)

### Added
* #976: fix slow authentication.
    * Server security setting in info (GUI, CLI).
    * Default SSL for SMTP based on Mac version.
    * GUI/CLI items to controls SMTP security setup.
    * Set new security and restart.

### Changed
* #961 Timeouts for go-pmapi client with http.Transport.
* Event poll with no change will hang forever. Using separate goroutine and timeout instead of proper fix (will be in refactor).
* Fixed an issue where entering an in-use port multiple times via the CLI would make bridge use it.
* Update therecipe/qt and Qt to 5.13.

## [Bridge 1.2.3] Akashi - live (2019-11-05) beta (2019-10-22)

### Added
* #963 report first-start metric with bridge version.
* #941 report new-login metric, report daily heartbeat.
* #921 remote key lookup via Web Key Directory (WKD).
* #919 TLS issue notification in CLI.

### Changed
* #769 #930 #931 #949 Syncing messages and fetching message and attachments in parallel with five workers.
* #956 #741 update keychain.
* Re-download and re-unlock user keyring when addresses are changed.
* #944 Ugrade go-pmapi dependency to v1.0.4 to support phase one of the key migration.
* #683 Password rehides each time password entry screen is shown.
* Import-Export#219 fix double parameter definition.
* Upgrade go-pm-bridge-utils dependency to v1.0.1.
* #935 Fix wrong download link for linux updates.
* #952 fix error when sending mail with only BCC recipients (use empty slice instead of nil slice).
* Refactor `generateSendingInfo` to simplify logic; add test for this method.
* Generate code-coverage report with `make code-coverage`.
* #942 fix focus window with logout message when trying to connect from the client.
* Do not use panic for second instance.
* #928 do not hide 'no keychain' dialog when upgrade is needed.
* Sending `NO` for errors while `FETCH`.
* #899 Upgrade from Bolt to BBolt.
* Upgrade to gopenpgp.
* Bridge utils in own repository.
* Code made compatible with name changes in go-pmapi.


## [Bridge 1.2.2] - beta and live 2019-09-06

### Changed
* User compare case insensitive.

## [Bridge 1.2.1] - beta and live 2019-09-05

### Changed
* #924 fix start of bridge without internet connection.

## [Bridge 1.2.0] - beta 2019-08-22

### Added
* #903 added http.Client timeout to not hang out forever.
* Closing body after checking internet connection.
* Pedantic lint for bridgeUtils.
* Selected events are buffered and emitted again when frontend loop is ready.
* #890 implemented 2FA endpoint (auth split).
* #888 TLS Cert.
    * Error bar and modal with explanation in GUI.
    * Signal to show error.
    * Add pinning to bridge (only for live API builds).
* #887 #883:
 * Wait before clearing data.
 * Configure which provides pmapi.ClientConfig and app directories.
* #861 restart after clear data.
* Panic handler for all goroutines.
* CD for linux.
* #798.
    * Check counts after sync.
    * Update counts in all mailboxes after sync.
    * `db.Mailbox.RemoveMissing`, `db.Mailbox.PutMany`.
    * `util.NotImplemented`.
    * Tests for sync.
* Bridge core tests:
    * Introduced interfaces: `pmapiClienterFactory`, `pmapiClienter`, `credentialStorer`.
    * Automatic mock generation for  `listener.Listener`, `bridge.pmapiClienter`, `bridge.credentialStorer`.
* #818 REFACTOR: see doc/code-structure.md.
    * Tests for bridge core & utils.
    * Update user before `GetQuota`.
    * Http bridge API.
* Bridge core tests:
    * Introduced interfaces: `pmapiClienterFactory`, `pmapiClienter`, `credentialStorer`.
    * Automatic mock generation for  `listener.Listener`, `bridge.pmapiClienter`, `bridge.credentialStorer`.
* #774 start initialization with sync immediately after login.

### Removed
* Using `PutMeta` for DB to not rewrite header and size.
* `Timeout` for connection (keep only `DialTimeout`).
* #798 `imapMailbox.sync`.
* #818 REFACTOR: see doc/code-structure.md.
    * Bridge global functions `GetAuth`, `GetAuthInfo`, `GetUserSettings` (using member functions of `pmapi.Client` instead).
    * `backend.setCache`: not used.
    * IMAP extension for `XSTOP` and `XFOCUS`.
    * Keychain `Disconnected` is not used,  deleting directly (not using hide).
   * `apiIdFrom(uid bool, id uint32)`, `apiIdRangeFromSeq(uid bool, seq imap.Seq)`: not used.
   * `server/dial.go` not used.
   * Util `CustomMessage`, `StartTicker` not used.

### Changed
* Check before first even sync.
* Do sync in parallel from events.
* Closing event loop by CloseConnectionEvent.
* Allow client to log in with address only.
* Fix IMAP users lock.
* #646 download headers when needed for first time.
* #895 fix of parsing address list.
* #844 do not spam GUI with logout events & sleep after bad login attempt from the client.
* #887 #883 #898 #902 logout account from API and all IMAP connections before clearing cache for account.
* #882 unassign PMAPI client after logout and force to run garbage collector.
* #880, #884, #885, #886 fix of informing user about outgoing non-encrypted e-mail.
* #838 `Sirupsen` -> `sirupsen`.
* #893 save panic report file every time.
* #880 fix of informing user about outgoing non-encrypted e-mail.
* Fix aliases in split mode.
* Fix decrypted data in log notification.
* #471 fix of using font awesome in regular text.
* `SearchMessage` all IDs from DB not depends on `totalOnAPI`.
* #798 populate efficiently.
    * Improved `imap.db.mailbox.Counts()`.
    * `mbox.total,unread` -> `mbox.totalOnAPI,unreadOnAPI`.
    * Always show DB status (even for `IDLE` updates).
    * `imapUser.sync` now takes `labelID` as parameter.
    * Split population by 1000 messages.
    * `db.User.put(msgs,keepCache)` is used in sync to not overwrite `msg.Size` and `msg.Header` in DB.
    * Separate sync function from `backend.labelMailbox` class.
    * `UidNext` uses bolt sequence value instead of cursor position.
* `util.tests.go` moved to `bridgeUtils`.
* #471 fix of using font awesome in regular text.
* #818 REFACTOR: see doc/code-structure.md.
    * No global states/variables anymore.
    * Code separated from one big package into smaller packages (bridge core, utils, IMAP, SMTP, API).
    * Bridge core completely refactored - core should be API over credentials store and PMAPI.
    * Configuration and preferences are on one place; passed as dependency to all packages.
    * Bridge utils separated from the rest of the bridge code to be used in Import/Export.
    * Many channels converted into one listener which can register listeners and emit events to them.
    * Each package is ready to be used with interfaces for possibility of mocking.
    * Removed IMAP extension XFOCUS, used bridge local API instead.
    * Removed IMAP extension XSTOP.
    * Sentry is not used in dev environment.
    * Logs are not cleared after start, clearing is triggered by `watchLogFileSize()` instead.
    * Log path changed one folder level up i.e. from `.../protonmail/bridge/c10` to `.../protonmail/bridge`.
    * Always cleared malformed keychain records.
    * Set credentials version on each `Put`.
    * `util.WriteHeader` -> `imap.writeHeader`.
    * Save `message.ExternalID` for every client not just AppleMail.
    * Server errors reported to frontend by common event listener.
* Handle logout in event loop.


## [Bridge 1.1.6] - 2019-07-09 (beta 2019-07-01)

### Added
* #841 assume text/plain during sending e-mails when missing content type.
* #805 list the new package links in upgrade dialog for linux.
* #802 report the list errors to sentry.
* #508 content related header fields for mail are saved in DB inside `msg.Header`.
* #508 `doNotCacheError` to decide whether to rebuild message.
* CI with lint check.
* Build flag `nogui`.
* Dummy html interface.

### Removed
* #508 content type rewrite on `GetHeader`.
* #508 content type on custom message.

### Changed
* #854 avoid `nil` header and bodystructure on fail (as regression of #508).
* Sanitize version in json file.
* #850 keep correct main and body headers for import (as regression of #508).
* #841 choose parent ID only when there is exactly one message with external ID.
* #811 #proton/backend-communication#11 go-pmapi!57 uid fixed.
* Update Qt 5.11.3 to 5.12.0.
* Using gomodules instead of glide.
* #508 use MIMEType and attachments to choose correct `Content-Type`.
* #508 custom message replaces body before header is created.
* #508 main header has `Content-Type` only after message was fully fetched.
* #770 ignore empty key from data card and support multiple keys for contacts.
* Build tags for simpler build of beta and QA builds.
* Lint corrections.


## [Bridge 1.1.5] - 2019-05-23 (beta 2019-05-23, 2019-05-16)

### Changed
* Fix custom message format.
* #802 accumulated long lines while parsing body structure.
* Process `AddressEvent` before `MessageEvent`.
* #791 updated crypto: fix wrong signature format.
* #793 fix returning size.
* #706 improved internet connection checking.
* #771 updated raven, crypto, pmapi.
* #792 use `INFO` as basic log level.
* Only one crash from second instance.
* During event `MessageID` in log as field.

## [Bridge 1.1.4 live] - 2019-04-10 (beta 2019-04-05, 2019-03-27)

### Added
* Address with port to IMAP debug.
* #750 `backend/events.Manager.LastEvent`.
* #750 `backend.user.areAllEventsProcessed`.
* #750 Wait with message events until all related mailboxes are synchronized.
* Restart limit to 10.
* Release string to raven.

### Changed
* #748 when charset missing assume utf8 and check the validity.
* #750 before sync check that events are up-to-date, if not poll events instead of sync.
* Use pmapi with support of decrypted access token.
* #750 Status is using DB status instead of API.
* Format panic error as string instead of struct dump.
* Validity of local certificate to increased to 20 years.

### Removed
* #750 Synchronization after 450 messages.

## [Bridge 1.1.3] - 2019-03-04

### Added
* Sentry crash reporting in main.
* Program arguments to turn of CPU and memory profiling.
* Full version of program visible on release notes.

### Changed
* #720 only one concurrent DB sync.
* #720 sync every 3 pages.
* #512 extending list of charsets go-pm-mime!4.

## [Bridge 1.1.2] - beta only 2019-02-21

### Changed
* #512 fail on unknown charset.
* #729 #733 visitor for MIME parsing.

## [Bridge 1.1.1] - 2019-02-11
### Added
* #671 include `name` param in attachment `Content-Type` (in addition to `Content-Disposition` param `filename`).
* #671 do not include content headers for section requests e.g. `BODY.PEEK[2]`.
* Version info checks for newer version (do not show dialog when older is online).
* #592 new header `X-Pm-ConversationID-Id` and also added to `References`.
* #666 invoke `panic` while adding account `jakubqa+crash@protonmail.com`.
* #592 new header fields `X-Pm-Date` storing m.Time and `X-Pm-External-Id` storing m.ExternalID.
* #484 search criteria `Unkeyword` support.

### Changed
* Fix srp modulus issue with new `ProtonMail/crypto`.
* Generate version files from main file.
* Be able to set update set on build.
* #597 check on start that certificate will be still valid after one month and generate new cert if not.
* #597 extended certificate validity to 2 years.
* Copyright 2019.
* Exclude `protontech` repos from credits.
* Refactor of `go-pmapi`, `go-pm-crypto`, `go-pm-mime` and `go-srp`.
* Re-signed pubkey key.
* Version, revision and build time is set in main.
* #666 use `bytes.Reader` instead of `bytes.Buffer`.
* #666 clear unused buffers in body structure map.
* No API request for fetch `body[header]`.
* Added crash file counter to pass log tests.
* #484 search fully relies on DB information (no need to reach API).
* #592 `parsingHeader` allows negative time (before 1.1.1970).
* #592 add original header first and then rewrite.
* #592 `Message-Id` rewritten only if not present.
* #592 rename `X-Internal-Id` to `X-Pm-Internal-Id`.
* #592 internal references are added only when not present already.
* #592 field `Date` changed to m.Time only when wrong format or missing `Date`.
* #645 pmapi#26 `Message.Flags` instead of `IsEncrypted`, `Type`, `IsReplied`, `IsRepliedAll`, `IsForwarded`.
* DB: do not allow to put Body or Attachments to db.
* #574 SMTP: can now send more than one email.
* #671 Verbosity levels: `debug` (only bridge), `debug-client` (bridge and client communication), `debug-server` (bridge, whole SMTP/IMAP communication).
* #644 Return rfc.size 0 or correct size of fetched body (stored in DB).
* #671 API requests URI in debug logs.
* #625 Fix search results for Flagged and Unflagged.
* Draft optimization fetch header.
* #656 Fix sending of calendar invite on Outlook on MacOS.
* #46 Allowed to run multiple instances, once per user.

### Removed
* Makefile clean up old deploy code.

### Release notes
* Support multiple Bridge instances running in parallel (one per user).

### Fixed bugs
* SMTP stays authenticated after sent message.
* Reduce memory, processor and number of API calls.

## [Bridge 1.1.0] - 2018-10-22

### Removed
* `go-pmapi.Config.ClientSecret`.
* `go-pmapi.PublicKey.Send`.
* Program argument `main`.
* `backend.getMIMEMessageBodySection` (use `message.BodySection`).
* `backend.getSize` (use `message.BodySection`).

### Added
* IMAP server: more info when write/send/flush error occurs #648.
* Linux package paths inside version-json.
* Draggable popup windows for outgoing non-encrypted message #519.
* Pmapi able to receive plain accessToken go-pmapi#23 #604.
* DB debug switch.
* Clearing message cache when db is cleared.
* Debug string to tests.
* Mime tree section parsing and test.
* Start ticker.
* Dump DB status.
* `backend.ApplicationOutdated()` mechanism: once triggered logout all email clients. On try to login say _application outdated_.
* Force upgrade event (send from event loop).
* New systray with error symbol (used in mac for force update).
* Test for upgrade.
* GUI for upgrade.
* Add native upgrade to updates.
* Dial timeout client.
* Custom `copyRecursively` function.
* When there is fresh version on start show release notes.
* Keychain helper using GNU pass.
* Error message on missing keychain.

### Changed
* Imap `SEARCH` loops until all messages are listed #581.
* Cached message timestamp is renewed on load.
* Message cache ID is userID+messageID.
* Private cache and added bodystructure.
* Remove addresses from `m.ToList` that were not requested in SMTP `TO`.
* IsFirstStart setup before loading Gui. Set it to false right after (don't wait till quit).
* Check `eventMessage` not nil before address check.
* `util.EventChannel` refactor: `SendEvent`->`Send` and new `SendEvent(EventCode)`.
* Information bar keeps on once app is outdated.
* Error dialog for upgrade has option for force upgrade.
* IsFirstStart setup before loading Gui. Set false right after (don't wait till quit).
* Pmapi: access token decrypted only if needed.
* Move `updates` from `frontend` to `util`.
* Move `CheckInternetConnection()` to `util`.
* Makefile clean up and change scripts for building.
* Reorganized updates.
* Start with new versioning.

          1.1.0
          | | `--- bug fix number (internal, irregular, beta releases)
          | `----- minor version (features, release once per month, live release, milestones)
          `------- major version (big changes, once per year, breaking changes, api force upgrade)

* Upgrade restart option in qt-frontend.
* GOOS save functions.
* Windows update procedure.
* Darwin update procedure.
* `zip` replaced by `tgz`.
* Using move instead of write truncated.
* Linux dependencies (pass and gnome-keychain optional).
* `Store.helper` -> `Store.secrets`.

### Release notes
* New self-update procedure.
* Changed restarting mechanism.
* Support for GNU pass for linux.
* Various GUI improvements.

### Fixed bugs
* RFC complaint SEARCH and FETCH responses.
* Additional synchronization of mail database.


## [Bridge 1.0.6 silent] - 2018-08-23
### Added
* New svg icon in linux package.

## [Bridge 1.0.6] - 2018-08-09

### Added
* `backend.GetUserSettings()`.

### Changed
* Related to Desktop-Bridge#561.
* Api flag to build scripts.
* `BodyKey` and `AttachmentKey` contains `Key` and `Algorithm`.
* `event.User.Addresses` -> `event.Addresses`.
* `user.Addresses` -> `client.Addresses()`.
* Typos and fixes.
* Pmapi update.
* `backend.configClient` -> `backend.authClient`.
* `auth.Uid` -> `auth.Uid()`.
* `keyRingForAddress()` -> `Client.KeyRingForAddressID()`.
* `Message.IsRead` -> `Message.Unread`.
* `pmapi.User.Unlock()` -> `pmapi.Client.UnlockAddresses()`.
* `TwoFactor` -> `HasTwoFactor()` and `PasswordMode` -> `HasMailboxPassword()`.
* Icon to match ImportExport.


### Release notes
* Removed deprecated API routes.

### Fixed bugs
* Frequent Thunderbird timeout.
* SMTP requests not case-sensitive.

## [Bridge 1.0.5] - 2018-07-12

### Added
* UpdateCurrentAgent from lastMailClient.
* Current OS.
* Use Qt to set nice OS with version.
* All `client.Do` errors are interpreted as connection issue.
* Moved to internal gitlab.
* Typo `frontend-qml`.
* Better message for case when server is not reachable.
* Setting 1min timeout to IMAP connection.

### Changed
* Password: click2show, click2hide.
* Notification in bug report window.
* Less frequent version check.
* Closing ticker.

### Removed
* Sockets and unused libraries.

### Release notes
* Improved response of IMAP server.
* Sending requests with client information.
* Less frequent notification about new version.

### Fixed bugs
* Support of Outlook calendar event format.
* Too many opened file descriptors issue.
* Fixed 7bit MIME issue while sending.


## [Bridge 1.0.4] - 2018-05-15

### Changed
* Version files available at both download and static.
* MIME `text/calendar` are parsed as attachment.
* UserID as identifier in keychain and pmapi token.
* Keychain format and function refactor.
* Create crash file on panic with full trace.
* Clear old data only in main process (no double keychain typing).
* Create label update API route.
* Selectable text in release notes.

### Added
* Support sending to external PGP recipients.
* Return error codes: `0: Ok`, `2: Frontend crashed`, `3: Bridge already running`, `4: Unknown argument`, `42: Restart application`.

### Release notes
* Support of encryption to external PGP recipients using contacts created on beta.protonmail.com (see https://protonmail.com/blog/pgp-vulnerability-efail/ to understand the vulnerabilities that may be associated with sending to other PGP clients).
* Notification that outgoing email will be delivered as non-encrypted.
* NOTE: Due to a change of the keychain format, you will need to add your account(s) to the Bridge after installing this version.

### Fixed bugs
* Support accounts with same user names.
* Support sending vCalendar event.

## [Bridge 1.0.3] - 2018-03-26
* All from silent updates plus following.

### Changed
* Okay -> "Remind me later".
* Imported message with `text/html` body was imported as `text/plain`.
* Reload cache when labeling Seen/Unseen.
* Merge with Import-Export branch.
    * Inheritable Bug report window.
    * Common functions: WriteHeader (parse PM mail) and CustomMessage (when incorrect keys).
    * Updates refactor.
    * Bug report window.
    * Checkbox and with label (only I/E).
    * Error dialog and Info tooltip (only I/E).
    * Add user modal formatting (colors, text).
    * Account view style.
    * Input box style (used in bug report).
    * Input field style (used in add account and change port).
    * Added style variables for I/E.
    * Tab button style.
    * Window bar style and functionality (closing / minimize window).

### Release notes
* Improved responsiveness in the UI.

### Fixed bugs
* Fixed some formatting issues with imports.
* Fixed port changing via CLI.

## Silent update - 2018-03-13

### Changed
* Remove firewall error message.


## [Bridge 1.0.2] - 2018-03-12
* All from silent updates plus following.

### Added
* UTF-7 support.
* Message when communication between bridge and email client is blocked by firewall (Windows only).

### Changed
* Added gnome-keyring dejavu fonts into linux dependency.
* Corrected parentID when reply/forward: taken from `protonmail.internalid` reference.
* Update user object in backend after unlock to apply address changes.

### Release notes
* UTF7 encoding support for older imported emails.

### Fixed bugs
* Fixed issues with conversation threading.
* Support for multiple "ReplyTo" addresses.
* Fixed issue where some address updates would not register immediately.



## [Bridge 1.0.1-4 (linux only)] Silent deploy - 2018-02-28

### Changed
* More similar look of window title bar to Windows 10 style.
* Qt 5.10 Button Controls 2 conflict (`icon`->`iconText`).

### Added
* Linux default font.
* Multiple reply-to addresses support (also API).
* Command line interface.
* Credits are generated automatically from glide.lock.
* Created script to build linux packages (dep,rpm,PKGBUILD).
* Correct config folders with env variable `$XDG_CONFIG_HOME`.

### Fixed bugs
* Clearing global cache.
* Default linux font problems.
* Support Reply-To multiple addresses.

### Release notes
* Improved visual appearance for win and linux.



## [Bridge 1.0.1] Silent deploy - 2017-12-30

### Changed
* Fixed bug with parsing address list (CC became BCC).



## [Bridge 1.0.1] - 2017-12-20

### Added
* When current log file is more than 10MB open new one, checked every 15min.
* Keep only last three log files including current one, triggered every start and when switching log files.
* Translation context.
* Accessibility objects for button and static text.
* All objects are accessible including focus scope for modals and messages.
* Automatically fill the email client in bug report form.
* Catch corrupted MacOS keychain error and show the link to FAQ.
* Unlabel message.
* Update emptying and filtering routes.
* Parse the address comment as defined in RFC.

### Changed
* Default log level set to Warning.
* Info logs during adding account and connecting client promoted to warning level.
* Log only when email client was changed (previously logged on every assign).
* Force upgrade bubble notification only when requested by API.
* Don't show warning systray icon when "You have then newest version!" bubble message is showed.
* Header date format  RFC822Z -> RFC1123Z.
* IMAP ID and QUOTA responses forced to use quoted strings (fixing SparkMail issue).
* Avoid AddressChanged bubble when no address was changed.

### Release notes
* Reduced log file size and log file history.
* Accessibility support for MacOS VoiceOver and Windows Narrator.
* Improved notification system.
* Supported imports with older address format.



## [Bridge 1.0.0] - 2017-12-06

### Added
* Encoding support of message body, title items, attachment name, for all standard charsets.
* Force update API message handled as new version event.

### Changed
* Refactor `bridge-qtfronted` -> `frontend`.
* Only one main file and basic support of CLI (not finished).
* Common QML package `ProtonUI`, which is used by `BridgeUI` and `ImportExportUI`.
* ChangedUser signal contain address and event type to distinguish between logout, internet off/on, address_change.
* API address changed event handled gracefully (if not possible, logout).
* Update mac keychain (should resolve problem with adding new account to bridge, NOT CONFIRMED).
* Solved hanging GUI on keychain error (should solve all win-7, no-gui errors).
* New systray icons for Mac (black and white no background).
* GUI cosmetics:
    * "Click here to start" triangle position.
    * Wrong cursor type on link.
    * Create main window before notification.

### Release notes
* Better notification when new version is needed or when account address is changed.
* Encoding support for the standard charsets.
* Improved visual appearance.

### Fixed bugs
* Fixed missing GUI for Windows with empty keychain.



## Changelog format
* Changelog [format](http://keepachangelog.com/en/1.0.0/).

### Guiding Principles
* Changelogs are for humans, not machines.
* There should be an entry for every single version.
* The same types of changes should be grouped.
* Versions and sections should be linkable.
* The latest version comes first.
* The release date of each version is displayed.
* Mention whether you follow Semantic Versioning.

### Types of changes
* `Added` for new features.
* `Changed` for changes in existing functionality.
* `Deprecated` for soon-to-be removed features.
* `Removed` for now removed features.
* `Fixed` for any bug fixes.
* `Security` in case of vulnerabilities.
* Additional for in app release notes.
    * `Release notes` in case of vulnerabilities.
    * `Fixed bugs` in case of vulnerabilities.

