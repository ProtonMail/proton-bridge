Feature: SMTP sending two messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:recp]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:recp]" and password "password"
    Then it succeeds


  @long-black
  Scenario: Send from one account to the other
    When user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:recp]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:recp]@[domain]>
      Subject: One account to the other

      hello

      """
    Then it succeeds
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "One account to the other",
          "Sender": {
            "Name": "Bridge Test",
            "Address": "[user:user]@[domain]"
          },
          "ToList": [
            {
              "Name": "Internal Bridge",
              "Address": "[user:recp]@[domain]"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
    And the body in the "POST" request to "/mail/v4/messages/.*" is:
      """
      {
        "Packages": [
          {
            "Addresses": {
              "[user:recp]@[domain]": {
                "Type": 1
              }
            },
            "Type": 1,
            "MIMEType": "text/plain"
          }
        ]
      }
      """
    When user "[user:recp]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                 | to                   | subject                  | body  |
      | [user:user]@[domain] | [user:recp]@[domain] | One account to the other | hello |


  @long-black
  Scenario: Send from one account to the other with attachments
    When user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:recp]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:recp]@[domain]>
      Subject: Plain with attachment internal
      Content-Type: multipart/related; boundary=bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/plain; charset=utf-8

      This is the body

      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606
      Content-Disposition: attachment; filename=outline-light-instagram-48.png
      Content-Id: <9114fe6f0adfaf7fdf7a@protonmail.com>
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAADAAAAAwBAMAAAClLOS0AAAALVBMVEUAAAD/////////////////
      //////////////////////////////////////+hSKubAAAADnRSTlMAgO8QQM+/IJ9gj1AwcIQd
      OXUAAAGdSURBVDjLXJC9SgNBFIVPXDURTYhgIQghINgowyLYCAYtRFAIgtYhpAjYhC0srCRW6YIg
      WNpoHVSsg/gEii+Qnfxq4DyDc3cyMfrBwl2+O+fOHTi8p7LS5RUf/9gpMKL7iT9sK47Q95ggpkzv
      1cvRcsGYNMYsmP+zKN27NR2vcDyTNVdfkOuuniNPMWafvIbljt+YoMEvW8y7lt+ARwhvrgPjhA0I
      BTng7S1GLPlypBvtIBPidY4YBDJFdtnkscQ5JGaGqxC9i7jSDwcwnB8qHWBaQjw1ABI8wYgtVoG6
      9pFkH8iZIiJeulFt4JLvJq8I5N2GMWYbHWDWzM3JZTMdeSWla0kW86FcuI0mfStiNKQ/AhEeh8h0
      YUTffFwrMTT5oSwdojIQ0UKcocgAKRH1HiqhFQmmJa5qRaYHNbRiSsOgslY0NdixItUTUWlZkedP
      HXVyAgAIA1F0wP5btQZPIyTwvAqa/Fl4oacuP+e4XHAjSYpkQkxSiMX+T7FPoZJToSStzED70HCy
      KE3NGCg4jJrC6Ti7AFwZLhnW0gMbzFZc0RmmeAAAAABJRU5ErkJggg==
      --bc5bd30245232f31b6c976adcd59bb0069c9b13f986f9e40c2571bb80aa16606--

      """
    Then it succeeds
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Plain with attachment internal",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:recp]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/plain"
        }
      }
      """
    And the body in the "POST" request to "/mail/v4/messages/.*" is:
      """
      {
        "Packages": [
          {
            "Addresses": {
              "[user:recp]@[domain]": {
                "Type": 1
              }
            },
            "Type": 1,
            "MIMEType": "text/plain"
          }
        ]
      }
      """
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                   | subject                        | body             | attachments                    | unread |
      | [user:user]@[domain] | [user:recp]@[domain] | Plain with attachment internal | This is the body | outline-light-instagram-48.png | false  |
    When user "[user:recp]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from                 | to                   | subject                        | body             | attachments                    | unread |
      | [user:user]@[domain] | [user:recp]@[domain] | Plain with attachment internal | This is the body | outline-light-instagram-48.png | true   |