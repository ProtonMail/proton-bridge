Feature: Import embedded message
  Background:
    Given there is connected user "user"
    And there is EML file "Inbox/hello.eml"
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

  Scenario: Import it
    When user "user" imports local files
    Then progress result is "OK"
    And transfer exported 1 messages
    And transfer imported 1 messages
    And transfer failed for 0 messages
    And API mailbox "INBOX" for "user" has messages
      | from            | to                 | subject          |
      | foo@example.com | bridgetest@pm.test | Embedded message |
