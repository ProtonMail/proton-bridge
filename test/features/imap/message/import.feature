Feature: IMAP import messages
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  @ignore
  Scenario: Import message with double charset in content type
    When IMAP client imports message to "INBOX"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message with double charset in content type
      Content-Type: text/plain; charset=utf-8; charset=utf-8
      Content-Disposition: inline

      Hello

      """
    Then IMAP response is "OK"

  @ignore
  Scenario: Import message with attachment name encoded by RFC 2047 without quoting
    When IMAP client imports message to "INBOX"
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Message with attachment name encoded by RFC 2047 without quoting
      Content-type: multipart/mixed; boundary="boundary"

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

      Hello íááá

      """
    Then IMAP response is "OK"

