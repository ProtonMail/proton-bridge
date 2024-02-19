Feature: SMTP sending of plain messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  Scenario: Basic message with attachment to internal account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: Plain with attachment
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject               |
      | [user:user]@[domain] | [user:to]@[domain] | Plain with attachment |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Plain with attachment",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:to]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Plain message with attachment to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain with attachment external
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject                        |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | Plain with attachment external |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Plain with attachment external",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "pm.bridge.qa@gmail.com",
              "Name": "External Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Plain message with attachment to two external accounts
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge 1 <pm.bridge.qa@gmail.com>
      CC: External Bridge 2 <bridgeqa@seznam.cz>
      Subject: Plain with attachment external PGP and external CC
      Content-Type: multipart/mixed; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is body of mail with attachment

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | cc                 | subject                                            |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | bridgeqa@seznam.cz | Plain with attachment external PGP and external CC |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Plain with attachment external PGP and external CC",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "pm.bridge.qa@gmail.com",
              "Name": "External Bridge 1"
            }
          ],
          "CCList": [
            {
              "Address": "bridgeqa@seznam.cz",
              "Name": "External Bridge 2"
            }
          ],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Basic message with multiple different attachments to internal account
    When user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following EML "plain/text_plain_multiple_attachments.template.eml" from "[user:user]@[domain]" to "[user:to]@[domain]"
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                     | to                 | subject                                   |
      | [user:user]@[domain]     | [user:to]@[domain] | Plain with multiple different attachments |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Plain with multiple different attachments",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:to]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
