Feature: IMAP import messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Basic message import
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from               | to                     | subject                  | body  |
      | bridgetest@pm.test | bridgetest@example.com | Basic text/plain message | Hello |

  Scenario: Import message with double charset in content type
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Message with double charset in content type
      Content-Type: text/plain; charset=utf-8; charset=utf-8
      Content-Disposition: inline
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from               | to                     | subject                                     | body  |
      | bridgetest@pm.test | bridgetest@example.com | Message with double charset in content type | Hello |


  Scenario: Import message with attachment name encoded by RFC 2047 without quoting
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message with attachment name encoded by RFC 2047 without quoting
      Content-type: multipart/mixed; boundary="boundary"
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      --boundary
      Content-Type: text/plain

      Hello

      --boundary
      Content-Type: application/pdf; name==?US-ASCII?Q?filename?=
      Content-Disposition: attachment; filename==?US-ASCII?Q?filename?=

      somebytes

      --boundary--

      """
    Then it succeeds


  # The message is imported as UTF-8 and the content type is determined at build time.
  Scenario: Import message as latin1 without content type
    When IMAP client "1" appends "text_plain_unknown_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from         | to             | body    |
      | sender@pm.me | receiver@pm.me | ééééééé |

  # The message is imported and the body is converted to UTF-8.
  Scenario: Import message as latin1 with content type
    When IMAP client "1" appends "text_plain_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from         | to             | body    |
      | sender@pm.me | receiver@pm.me | ééééééé |

  # The message is imported anad the body is wrongly converted (body is corrupted).
  Scenario: Import message as latin1 with wrong content type
    When IMAP client "1" appends "text_plain_wrong_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from         | to             |
      | sender@pm.me | receiver@pm.me |

  Scenario: Import received message to Sent
    When IMAP client "1" appends the following message to "Sent":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from            | to                 | subject | body  |
      | foo@example.com | bridgetest@pm.test | Hello   | Hello |
    And IMAP client "1" eventually sees 0 messages in "Inbox"

  Scenario: Import non-received message to Inbox
    When IMAP client "1" appends the following message to "Inbox":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from            | to                 | subject | body  |
      | foo@example.com | bridgetest@pm.test | Hello   | Hello |
    And IMAP client "1" eventually sees 0 messages in "Sent"

  Scenario: Import non-received message to Sent
    When IMAP client "1" appends the following message to "Sent":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from            | to                 | subject | body  |
      | foo@example.com | bridgetest@pm.test | Hello   | Hello |
    And IMAP client "1" eventually sees 0 messages in "Inbox"

  Scenario Outline: Import message without sender to <mailbox>
    When IMAP client "1" appends the following message to "<mailbox>":
      """
      From: Somebody@somewhere.org
      Date: 01 Jan 1980 00:00:00 +0000
      To: Lionel Richie <lionel@richie.com>
      Subject: RE: Hello, is it me you looking for?

      Nope.
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "<mailbox>":
      | to                | subject                              | body  |
      | lionel@richie.com | RE: Hello, is it me you looking for? | Nope. |

    Examples:
      | mailbox |
      | Drafts  |
      | Archive |
      | Sent    |

  Scenario: Import embedded message
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Embedded message
      Content-Type: multipart/mixed; boundary="boundary"
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      This is a multi-part message in MIME format.
      --boundary
      Content-Type: text/plain; charset=utf-8
      Content-Transfer-Encoding: 7bit


      --boundary
      Content-Type: message/rfc822; name="embedded.eml"
      Content-Transfer-Encoding: 7bit
      Content-Disposition: attachment; filename="embedded.eml"

      From: Bar <bar@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Subject: (No Subject)
      Content-Type: text/plain; charset=utf-8
      Content-Transfer-Encoding: quoted-printable

      hello

      --boundary--

      """
    Then it succeeds
