Feature: IMAP import messages
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Import message with double charset in content type
    When IMAP client imports message to "INBOX"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message with double charset in content type
      Content-Type: text/plain; charset=utf-8; charset=utf-8
      Content-Disposition: inline
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello

      """
    Then IMAP response is "OK"

  # I could not find any RFC why this is not valid. But for now our parser is not able to process it.
  @ignore
  Scenario: Import message with attachment name encoded by RFC 2047 without quoting
    When IMAP client imports message to "INBOX"
      """
      From: Bridge Test <bridgetest@pm.test>
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
    Then IMAP response is "OK"

  Scenario: Import message as latin1 without content type
    When IMAP client imports message to "INBOX" with encoding "latin1"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message in latin1 without content type
      Content-Disposition: inline
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello íááá

      """
    Then IMAP response is "OK"

  Scenario: Import message as latin1 with content type
    When IMAP client imports message to "INBOX" with encoding "latin1"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message in latin1 with content type
      Content-Disposition: inline
      Content-Type: text/plain; charset=latin1
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello íááá

      """
    Then IMAP response is "OK"

  Scenario: Import message as latin1 with wrong content type
    When IMAP client imports message to "INBOX" with encoding "latin1"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message in latin1 with wrong content type
      Content-Disposition: inline
      Content-Type: text/plain; charset=KOI8R
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello íááá

      """
    Then IMAP response is "OK"

  Scenario: Import received message to Sent
    When IMAP client imports message to "Sent"
      """
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello

      """
    Then IMAP response is "OK"
    And API mailbox "Sent" for "user" has 0 message
    And API mailbox "INBOX" for "user" has 1 message

  Scenario: Import non-received message to Inbox
    When IMAP client imports message to "INBOX"
      """
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello

      Hello

      """
    Then IMAP response is "OK"
    And API mailbox "INBOX" for "user" has 0 message
    And API mailbox "Sent" for "user" has 1 message

  Scenario Outline: Import message without sender
    When IMAP client imports message to "<mailbox>"
      """
      To: Lionel Richie <lionel@richie.com>
      Subject: RE: Hello, is it me you looking for?

      Nope.

      """
    Then IMAP response is "OK"
    And API mailbox "<mailbox>" for "user" has 1 message

    Examples:
        | mailbox |
        | Drafts  |
        | Archive |
        | Sent    |


  Scenario: Import embedded message
    When IMAP client imports message to "INBOX"
      """
      From: Foo <foo@example.com>
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
    Then IMAP response is "OK"

  # We cannot control internal IDs on live server.
  @ignore-live
  Scenario: Import existing message
    When IMAP client imports message to "INBOX"
      """
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      X-Pm-Internal-Id: 1

      Hello

      """
    Then IMAP response is "OK \[APPENDUID \d 1\] APPEND completed"
    When IMAP client imports message to "INBOX"
      """
      From: Foo <foo@example.com>
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      X-Pm-Internal-Id: 1

      Hello

      """
    Then IMAP response is "OK \[APPENDUID \d 1\] APPEND completed"
