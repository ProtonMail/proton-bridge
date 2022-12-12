Feature: SMTP sending embedded message
  Background:
    Given there exists an account with username "user" and password "password"
    And there exists an account with username "bridgetest" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And the user logs in with username "bridgetest" and password "password"
    And user "user" connects and authenticates SMTP client "1"

  Scenario: Send it
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Embedded message
      Content-Type: multipart/mixed; boundary="boundary"

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
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to                  | subject          |
      | user@[domain] | bridgetest@[domain] | Embedded message |
    When user "bridgetest" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from          | to                  | subject          | attachments  | unread |
      | user@[domain] | bridgetest@[domain] | Embedded message | embedded.eml | true   |