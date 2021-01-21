## v1.5.6
- 2021-01-21

### New

Improvements to message parsing

Better error handling

Ensured better message flow by refactoring both address and date parsing

Improved secure connectivity checks

Better deb packaging

More robust error handling

Improved package creation logic

Refactor of sending functions to simplify code maintenance

Added tests for package creation

Support read confirmations

Adding GPLv3 licence button to the GUI

Improved testing

### Fixed

AppleMail crashes (related to timestamps)

Sending messages from aliases in combined inbox mode

Fedora font issues

Ensured that conversations are properly threaded

Fixed Linux font issues (Fedora)

Better handling of Mime encrypted messages

Bridge crashes related to labels handling

GUI popup related to TLS connection error

An issue where a random session key is included in the data payload

Error handling (including improved detection)

Encoding errors

Installation issues on linux


## v1.4.5
- 2020-10-22

### New

Improving Performance

  * Bulletproofing against any potential data loss and/or duplication
  * Performance improvements for handling attachments and non-standard formatting
  * Better stability of the message parser

Outgoing messages support

  * Additional foreign encoding support for outgoing messages
  * Complete refactor of the way messages are parsed to simplify code maintenance
  * Improved User-Agent detection

Added MacOS Big Sur compatibility

Added persistent anonymous API cookies

### Fixed

Fixed rare mail loss when moving from Spam folder

Limited log size

Fixed Linux font issues (mouse hover)



## v1.3.3
- 2020-08-12

### New

Improvements to Alternative Routing

  * Version two of this feature is now more resilient to unstable internet connections, which results in a smoother experience using this feature.
  * Includes fixes to previous implementation of Alternative Routing when first starting the application or when turning it off.
  
Email parsing improvements

  * Improved detection of email encodings embedded in html/xml in addition to message header; add a fallback option if encoding is not specified and decoding as UTF8 fails (ISO-8859-1)
  * tweaked logic of parsing "References" header.

User interaction improvements

  * Some smaller improvements in specific cases to make the interaction with Proton Bridge clearer for the user

Code updates & maintenance

  * Migrated to GopenPGP v2
  * updates to GoIMAPv1
  * increased bbolt version to 1.3.5 and various improvements regarding extensibility and maintainability for upcoming work.
  
General stability improvements

  * Improvements to the behavior of the application under various unstable internet conditions.

### Fixed

Fixed a slew of smaller bugs and some conditions which could cause the application to crash.
