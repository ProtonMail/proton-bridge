# ProtonMail Bridge and Import Export
Copyright (c) 2020 Proton Technologies AG

This repository holds the ProtonMail Bridge application.
For a detailed build information see [BUILDS](./BUILDS.md).
For licensing information see [COPYING](./COPYING.md).
For contribution policy see [CONTRIBUTING](./CONTRIBUTING.md).


## Description Bridge
ProtonMail Bridge for e-mail clients.

When launched, Bridge will initialize local IMAP/SMTP servers and render 
its GUI.

To configure an e-mail client, firstly log in using your ProtonMail credentials. 
Open your e-mail client and add a new account using the settings which are 
located in the Bridge GUI. The client will only be able to sync with 
your ProtonMail account when the Bridge is running, thus the option 
to start Bridge on startup is enabled by default.

When the main window is closed, Bridge will continue to run in the
background.

More details [on the public website](https://protonmail.com/bridge).

## Description Import-Export
TODO

## Keychain
You need to have a keychain in order to run the ProtonMail Bridge. On Mac or
Windows, Bridge uses native credential managers. On Linux, use
[Gnome keyring](https://wiki.gnome.org/Projects/GnomeKeyring/)
or
[pass](https://www.passwordstore.org/).


## Environment Variables

### Bridge application
- `BRIDGESTRICTMODE`: tells bridge to turn on `bbolt`'s "strict mode" which checks the database after every `Commit`. Set to `1` to enable.

### Dev build or run
- `APP_VERSION`: set the bridge app version used during testing or building
- `PROTONMAIL_ENV`: when set to `dev` it is not using Sentry to report crashes
- `VERBOSITY`: set log level used during test time and by the makefile

### Integration testing
- `TEST_ENV`: set which env to use (fake or live)
- `TEST_ACCOUNTS`: set JSON file with configured accounts
- `TAGS`: set build tags for tests
- `FEATURES`: set feature dir, file or scenario to test


## Files
### Database
The database stores metadata necessary for presenting messages and mailboxes to an email client:
- Linux: `~/.cache/protonmail/bridge/<cacheVersion>/mailbox-<userID>.db` (unless `XDG_CACHE_HOME` is set, in which case that is used as your `~`)
- macOS: `~/Library/Caches/protonmail/bridge/<cacheVersion>/mailbox-<userID>.db`
- Windows: `%LOCALAPPDATA%\protonmail\bridge\<cacheVersion>\mailbox-<userID>.db`

### Preferences
User preferences are stored in json at the following location:
- Linux: `~/.cache/protonmail/bridge/<cacheVersion>/prefs.json` (unless `XDG_CACHE_HOME` is set, in which case that is used as your `~`)
- macOS: `~/Library/Caches/protonmail/bridge/<cacheVersion>/prefs.json`
- Windows: `%LOCALAPPDATA%\protonmail\bridge\<cacheVersion>\prefs.json`

### IMAP Cache
The currently subscribed mailboxes are held in a json file:
- Linux: `~/.cache/protonmail/bridge/<cacheVersion>/user_info.json` (unless `XDG_CACHE_HOME` is set, in which case that is used as your `~`)
- macOS: `~/Library/Caches/protonmail/bridge/<cacheVersion>/user_info.json`
- Windows: `%LOCALAPPDATA%\protonmail\bridge\<cacheVersion>\user_info.json`

### Lock file
Bridge utilises an on-disk lock to ensure only one instance is run at once. The lock file is here: 
- Linux: `~/.cache/protonmail/bridge/<cacheVersion>/bridge.lock` (unless `XDG_CACHE_HOME` is set, in which case that is used as your `~`)
- macOS: `~/Library/Caches/protonmail/bridge/<cacheVersion>/bridge.lock`
- Windows: `%LOCALAPPDATA%\protonmail\bridge\<cacheVersion>\bridge.lock`

### TLS Certificate and Key
When bridge first starts, it generates a unique TLS certificate and key file at the following locations:
- Linux: `~/.config/protonmail/bridge/{cert,key}.pem` (unless `XDG_CONFIG_HOME` is set, in which case that is used as your `~/.config`)
- macOS: `~/Library/ApplicationSupport/protonmail/bridge/{cert,key}.pem`
- Windows: `%APPDATA%\protonmail\bridge\{cert,key}.pem`

