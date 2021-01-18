## v1.1.2
- 2020-09-23

### New

Improving performance

  * Speed up import by implementing parallel processing (parallel fetch, encrypt and upload of messages)
  * Optimising the initial fetch of messages from external accounts

Better message parsing

  * Better handling of attachments and non-standard formatting
  * Improved stability of the message parser

Improved metrics

  * Added persistent anonymous API cookies


### Fixed

Fixed issues causing failing of import

  * Import from mbox files with long lines
  * Improvements to import from Yahoo accounts