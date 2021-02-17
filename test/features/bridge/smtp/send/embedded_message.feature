Feature: SMTP sending embedded message
  Scenario: Send it
    Given there is connected user "user"
    And there is SMTP client logged in as "user"
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: Internal Bridge <bridgetest@protonmail.com>
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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject          |
      | [userAddress] | bridgetest@protonmail.com | Embedded message |
