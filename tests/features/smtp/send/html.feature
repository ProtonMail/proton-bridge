Feature: SMTP sending of plain messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  Scenario: HTML message to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: HTML text external
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8
      In-Reply-To: <base64hashOfSomeMessage@protonmail.internalid>

      <html><body>This is body of <b>HTML mail</b> without attachment<body></html>

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject            |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | HTML text external |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "HTML text external",
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
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: HTML message with inline image to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Html Inline External
      Content-Disposition: inline
      User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.5.0
      MIME-Version: 1.0
      Content-Language: en-US
      Content-Type: multipart/related; boundary="------------61FA22A41A3F46E8E90EF528"

      This is a multi-part message in MIME format.
      --------------61FA22A41A3F46E8E90EF528
      Content-Type: text/html; charset=utf-8
      Content-Transfer-Encoding: 7bit

      <html>
      <head>
      <meta http-equiv="content-type" content="text/html; charset=UTF-8">
      </head>
      <body text="#000000" bgcolor="#FFFFFF">
      <p><br>
      </p>
      <p>Behold! An inline <img moz-do-not-send="false"
      src="cid:part1.D96BFAE9.E2E1CAE3@protonmail.com" alt=""
      width="24" height="24"><br>
      </p>
      </body>
      </html>

      --------------61FA22A41A3F46E8E90EF528
      Content-Type: image/gif; name="email-action-left.gif"
      Content-Transfer-Encoding: base64
      Content-ID: <part1.D96BFAE9.E2E1CAE3@protonmail.com>
      Content-Disposition: inline; filename="email-action-left.gif"

      R0lGODlhGAAYANUAACcsKOHs4kppTH6tgYWxiIq0jTVENpG5lDI/M7bRuEaJSkqOTk2RUU+P
      U16lYl+lY2iva262cXS6d3rDfYLNhWeeamKTZGSVZkNbRGqhbOPt4////+7u7qioqFZWVlNT
      UyIiIgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACwAAAAAGAAYAAAG
      /8CNcLjRJAqVRqNSSGiI0GFgoKhar4NAdHioMhyRCYUyiTgY1cOWUH1ILgIDAGAQXCSPKgHa
      XUAyGCCCg4IYGRALCmpCAVUQFgiEkiAIFhBVWhtUDxmRk5IIGXkDRQoMEoGfHpIYEmhGCg4X
      nyAdHB+SFw4KRwoRArQdG7eEAhEKSAoTBoIdzs/Cw7iCBhMKSQoUAIJbQ8QgABQKStnbIN1C
      3+HjFcrMtdDO6dMg1dcFvsCfwt+CxsgJYs3a10+QLl4aTKGitYpQq1eaFHDyREtQqFGMHEGq
      SMkSJi4K/ACiZQiRIihsJL6JM6fOnTwK9kTpYgqMGDJm0JzsNuWKTw0FWdANMYJECRMnW4IA
      ADs=
      --------------61FA22A41A3F46E8E90EF528--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject              |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | Html Inline External |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Html Inline External",
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
          "MIMEType": "text/html"
        }
      }
      """
    And IMAP client "1" eventually sees the following message in "Sent" with this structure:
    """
    {
      "date": "01 Jan 01 00:00 +0000",
      "to": "External Bridge <pm.bridge.qa@gmail.com>",
      "from": "Bridge Test <[user:user]@[domain]>",
      "subject": "Html Inline External",
      "content": {
        "content-type": "multipart/mixed",
        "sections":[
          {
            "content-type": "multipart/related",
            "sections":[
              {
                "content-type": "text/html",
                "content-type-charset": "utf-8",
                "transfer-encoding": "quoted-printable",
                "body-is": "<html><head>\r\n<meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8\"/>\r\n</head>\r\n<body text=3D\"#000000\" bgcolor=3D\"#FFFFFF\">\r\n<p><br/>\r\n</p>\r\n<p>Behold! An inline <img moz-do-not-send=3D\"false\" src=3D\"cid:part1.D96BFA=\r\nE9.E2E1CAE3@protonmail.com\" alt=3D\"\" width=3D\"24\" height=3D\"24\"/><br/>\r\n</p>\r\n\r\n\r\n</body></html>"
              },
              {
                "content-type": "image/gif",
                "content-type-name": "email-action-left.gif",
                "content-disposition": "inline",
                "content-disposition-filename": "email-action-left.gif",
                "transfer-encoding": "base64",
                "body-is": "R0lGODlhGAAYANUAACcsKOHs4kppTH6tgYWxiIq0jTVENpG5lDI/M7bRuEaJSkqOTk2RUU+PU16l\r\nYl+lY2iva262cXS6d3rDfYLNhWeeamKTZGSVZkNbRGqhbOPt4////+7u7qioqFZWVlNTUyIiIgAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\r\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACwAAAAAGAAYAAAG/8CNcLjRJAqVRqNS\r\nSGiI0GFgoKhar4NAdHioMhyRCYUyiTgY1cOWUH1ILgIDAGAQXCSPKgHaXUAyGCCCg4IYGRALCmpC\r\nAVUQFgiEkiAIFhBVWhtUDxmRk5IIGXkDRQoMEoGfHpIYEmhGCg4XnyAdHB+SFw4KRwoRArQdG7eE\r\nAhEKSAoTBoIdzs/Cw7iCBhMKSQoUAIJbQ8QgABQKStnbIN1C3+HjFcrMtdDO6dMg1dcFvsCfwt+C\r\nxsgJYs3a10+QLl4aTKGitYpQq1eaFHDyREtQqFGMHEGqSMkSJi4K/ACiZQiRIihsJL6JM6fOnTwK\r\n9kTpYgqMGDJm0JzsNuWKTw0FWdANMYJECRMnW4IAADs="
              }
            ]
          }
        ]
      }
    }
    """

  Scenario: HTML message with alternative inline to internal account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:user2]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:user2]@[domain]>
      Subject: Html Inline Alternative Internal
      Content-Disposition: inline
      User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.5.0
      MIME-Version: 1.0
      Content-Type: multipart/alternative; boundary="------------5A259F4DE164B5ADA313F644"
      Content-Language: en-US

      This is a multi-part message in MIME format.
      --------------5A259F4DE164B5ADA313F644
      Content-Type: text/plain; charset=utf-8
      Content-Transfer-Encoding: 7bit


      Behold! An inline


      --------------5A259F4DE164B5ADA313F644
      Content-Type: multipart/related; boundary="------------61FA22A41A3F46E8E90EF528"


      --------------61FA22A41A3F46E8E90EF528
      Content-Type: text/html; charset=utf-8
      Content-Transfer-Encoding: 7bit

      <html>
      <head>
      <meta http-equiv="content-type" content="text/html; charset=UTF-8">
      </head>
      <body text="#000000" bgcolor="#FFFFFF">
      <p><br>
      </p>
      <p>Behold! An inline <img moz-do-not-send="false"
      src="cid:part1.D96BFAE9.E2E1CAE3@protonmail.com" alt=""
      width="24" height="24"><br>
      </p>
      </body>
      </html>

      --------------61FA22A41A3F46E8E90EF528
      Content-Type: image/gif; name="email-action-left.gif"
      Content-Transfer-Encoding: base64
      Content-ID: <part1.D96BFAE9.E2E1CAE3@protonmail.com>
      Content-Disposition: inline; filename="email-action-left.gif"

      R0lGODlhGAAYANUAACcsKOHs4kppTH6tgYWxiIq0jTVENpG5lDI/M7bRuEaJSkqOTk2RUU+P
      U16lYl+lY2iva262cXS6d3rDfYLNhWeeamKTZGSVZkNbRGqhbOPt4////+7u7qioqFZWVlNT
      UyIiIgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACwAAAAAGAAYAAAG
      /8CNcLjRJAqVRqNSSGiI0GFgoKhar4NAdHioMhyRCYUyiTgY1cOWUH1ILgIDAGAQXCSPKgHa
      XUAyGCCCg4IYGRALCmpCAVUQFgiEkiAIFhBVWhtUDxmRk5IIGXkDRQoMEoGfHpIYEmhGCg4X
      nyAdHB+SFw4KRwoRArQdG7eEAhEKSAoTBoIdzs/Cw7iCBhMKSQoUAIJbQ8QgABQKStnbIN1C
      3+HjFcrMtdDO6dMg1dcFvsCfwt+CxsgJYs3a10+QLl4aTKGitYpQq1eaFHDyREtQqFGMHEGq
      SMkSJi4K/ACiZQiRIihsJL6JM6fOnTwK9kTpYgqMGDJm0JzsNuWKTw0FWdANMYJECRMnW4IA
      ADs=
      --------------61FA22A41A3F46E8E90EF528--

      --------------5A259F4DE164B5ADA313F644--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject                          |
      | [user:user]@[domain] | [user:user2]@[domain] | Html Inline Alternative Internal |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Html Inline Alternative Internal",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "[user:user2]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [],
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: HTML message with alternative inline to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Html Inline Alternative External
      Content-Disposition: inline
      User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.5.0
      MIME-Version: 1.0
      Content-Type: multipart/alternative; boundary="------------5A259F4DE164B5ADA313F644"
      Content-Language: en-US

      This is a multi-part message in MIME format.
      --------------5A259F4DE164B5ADA313F644
      Content-Type: text/plain; charset=utf-8
      Content-Transfer-Encoding: 7bit


      Behold! An inline


      --------------5A259F4DE164B5ADA313F644
      Content-Type: multipart/related; boundary="------------61FA22A41A3F46E8E90EF528"


      --------------61FA22A41A3F46E8E90EF528
      Content-Type: text/html; charset=utf-8
      Content-Transfer-Encoding: 7bit

      <html>
      <head>
      <meta http-equiv="content-type" content="text/html; charset=UTF-8">
      </head>
      <body text="#000000" bgcolor="#FFFFFF">
      <p><br>
      </p>
      <p>Behold! An inline <img moz-do-not-send="false"
      src="cid:part1.D96BFAE9.E2E1CAE3@protonmail.com" alt=""
      width="24" height="24"><br>
      </p>
      </body>
      </html>

      --------------61FA22A41A3F46E8E90EF528
      Content-Type: image/gif; name="email-action-left.gif"
      Content-Transfer-Encoding: base64
      Content-ID: <part1.D96BFAE9.E2E1CAE3@protonmail.com>
      Content-Disposition: inline; filename="email-action-left.gif"

      R0lGODlhGAAYANUAACcsKOHs4kppTH6tgYWxiIq0jTVENpG5lDI/M7bRuEaJSkqOTk2RUU+P
      U16lYl+lY2iva262cXS6d3rDfYLNhWeeamKTZGSVZkNbRGqhbOPt4////+7u7qioqFZWVlNT
      UyIiIgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACwAAAAAGAAYAAAG
      /8CNcLjRJAqVRqNSSGiI0GFgoKhar4NAdHioMhyRCYUyiTgY1cOWUH1ILgIDAGAQXCSPKgHa
      XUAyGCCCg4IYGRALCmpCAVUQFgiEkiAIFhBVWhtUDxmRk5IIGXkDRQoMEoGfHpIYEmhGCg4X
      nyAdHB+SFw4KRwoRArQdG7eEAhEKSAoTBoIdzs/Cw7iCBhMKSQoUAIJbQ8QgABQKStnbIN1C
      3+HjFcrMtdDO6dMg1dcFvsCfwt+CxsgJYs3a10+QLl4aTKGitYpQq1eaFHDyREtQqFGMHEGq
      SMkSJi4K/ACiZQiRIihsJL6JM6fOnTwK9kTpYgqMGDJm0JzsNuWKTw0FWdANMYJECRMnW4IA
      ADs=
      --------------61FA22A41A3F46E8E90EF528--

      --------------5A259F4DE164B5ADA313F644--

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject                          |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | Html Inline Alternative External |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Html Inline Alternative External",
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
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: HTML message with extremely long line (greater than default 2000 line limit) to external account
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: HTML text external
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8
      In-Reply-To: <base64hashOfSomeMessage@protonmail.internalid>

      <html><body>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa<body></html>

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                     | subject            |
      | [user:user]@[domain] | pm.bridge.qa@gmail.com | HTML text external |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "HTML text external",
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
          "MIMEType": "text/html"
        }
      }
      """

  Scenario: HTML message with Foreign/Nonascii chars in Subject and Body to external
    When there exists an account with username "bridgetest" and password "password"
    And the user logs in with username "bridgetest" and password "password"
    And user "bridgetest" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following EML "html/foreign_ascii_subject_body.eml" from "bridgetest@proton.local" to "pm.bridge.qa@gmail.com"
    Then it succeeds
    When user "bridgetest" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                    | to                     | subject        |
      | bridgetest@proton.local | pm.bridge.qa@gmail.com | Subjεέςτ ¶ Ä È |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Subjεέςτ ¶ Ä È",
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
          "MIMEType": "text/html"
        }
      }
      """

  # It is expected for the structure check to look a bit different. More info on GODT-3011
  @regression
  Scenario: HTML message with remote content in Body
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
    """
    Date: 01 Jan 1980 00:00:00 +0000
    To: Internal Bridge Test <[user:to]@[domain]>
    From: Bridge Test <[user:user]@[domain]>
    Subject: MESSAGE WITH REMOTE CONTENT SENT
    Content-Type: multipart/alternative;
      boundary="------------vUMV7TiM65KWBg30p6OgD3Vp"

    This is a multi-part message in MIME format.
    --------------vUMV7TiM65KWBg30p6OgD3Vp
    Content-Type: text/plain; charset=utf-8; format=flowed
    Content-Transfer-Encoding: 7bit

    Remote content


    Bridge


    Remote content


    --------------vUMV7TiM65KWBg30p6OgD3Vp
    Content-Type: text/html; charset=utf-8
    Content-Transfer-Encoding: 7bit

    <!DOCTYPE html>
    <html>
      <head>

        <meta http-equiv="content-type" content="text/html; charset=utf-8">
      </head>
      <body>
        <p><tt>Remote content</tt></p>
        <p><tt><br>
          </tt></p>
        <p><img
            src="https://bridgeteam.protontech.ch/bridgeteam/tmp/bridge.jpg"
            alt="Bridge" width="180" height="180"></p>
        <p><br>
        </p>
        <p><tt>Remote content</tt><br>
        </p>
        <br>
      </body>
    </html>

    --------------vUMV7TiM65KWBg30p6OgD3Vp--

    """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    And IMAP client "1" eventually sees the following message in "Sent" with this structure:
    """
    {
      "date": "01 Jan 01 00:00 +0000",      
      "to": "Internal Bridge Test <[user:to]@[domain]>",
      "from": "Bridge Test <[user:user]@[domain]>",
      "subject": "MESSAGE WITH REMOTE CONTENT SENT",
      "content": {
        "content-type": "text/html",
        "content-type-charset": "utf-8",
        "transfer-encoding": "quoted-printable",
        "body-is": "<!DOCTYPE html><html><head>\n\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\"/>\n  </head>\n  <body>\n    <p><tt>Remote content</tt></p>\n    <p><tt><br/>\n      </tt></p>\n    <p><img src=\"https://bridgeteam.protontech.ch/bridgeteam/tmp/bridge.jpg\" alt=\"Bridge\" width=\"180\" height=\"180\"/></p>\n    <p><br/>\n    </p>\n    <p><tt>Remote content</tt><br/>\n    </p>\n    <br/>\n  \n\n</body></html>"
      }
    }
    """

Scenario: Forward a message containing various attachments
  When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:user2]@[domain]":
    """
    Content-Type: multipart/mixed; boundary="------------MQ01Z9UM8OaR9z39TvzDfdIq"
    Subject: Fwd: Reply to this message, it has various attachments.
    References: <something@protonmail.ch>
    To: <[user:user2]@[domain]>
    From: <[user:user]@[domain]>
    In-Reply-To: <something@protonmail.ch>
    X-Forwarded-Message-Id: <something@protonmail.ch>

    This is a multi-part message in MIME format.
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: text/plain; charset=UTF-8; format=flowed
    Content-Transfer-Encoding: 7bit

    Forwarding a message with various attachments in it!



    -------- Forwarded Message --------
    Subject: 	Reply to this message, it has various attachments.
    Date: 	Thu, 26 Oct 2023 10:41:55 +0000
    From: 	Gjorgji Testing <gorgitesting@protonmail.com>
    Reply-To: 	Gjorgji Testing <gorgitesting@protonmail.com>
    To: 	Gjorgji Test v3 <gorgitesting3@protonmail.com>




    For real!

    *Gjorgji Testing
    TesASID <https://www.youtube.com/watch?v=MifXUbrjYr8>
    *
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: text/html; charset=UTF-8; name="index.html"
    Content-Disposition: attachment; filename="index.html"
    Content-Transfer-Encoding: base64

    IDwhRE9DVFlQRSBodG1sPg0KPGh0bWw+DQo8aGVhZD4NCjx0aXRsZT5QYWdlIFRpdGxlPC90
    aXRsZT4NCjwvaGVhZD4NCjxib2R5Pg0KDQo8aDE+TXkgRmlyc3QgSGVhZGluZzwvaDE+DQo8
    cD5NeSBmaXJzdCBwYXJhZ3JhcGguPC9wPg0KDQo8L2JvZHk+DQo8L2h0bWw+IA==
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: text/xml; charset=UTF-8; name="testxml.xml"
    Content-Disposition: attachment; filename="testxml.xml"
    Content-Transfer-Encoding: base64

    PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPCFET0NUWVBFIHN1aXRl
    IFNZU1RFTSAiaHR0cDovL3Rlc3RuZy5vcmcvdGVzdG5nLTEuMC5kdGQiID4KCjxzdWl0ZSBu
    YW1lPSJBZmZpbGlhdGUgTmV0d29ya3MiPgoKICAgIDx0ZXN0IG5hbWU9IkFmZmlsaWF0ZSBO
    ZXR3b3JrcyIgZW5hYmxlZD0idHJ1ZSI+CiAgICAgICAgPGNsYXNzZXM+CiAgICAgICAgICAg
    IDxjbGFzcyBuYW1lPSJjb20uY2xpY2tvdXQuYXBpdGVzdGluZy5hZmZOZXR3b3Jrcy5Bd2lu
    VUtUZXN0Ii8+CiAgICAgICAgPC9jbGFzc2VzPgogICAgPC90ZXN0PgoKPC9zdWl0ZT4=
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: application/pdf; name="test.pdf"
    Content-Disposition: attachment; filename="test.pdf"
    Content-Transfer-Encoding: base64

    JVBERi0xLjUKJeLjz9MKNyAwIG9iago8PAovVHlwZSAvRm9udERlc2NyaXB0b3IKL0ZvbnRO
    MjM0NAolJUVPRgo=
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;
      name="test.xlsx"
    Content-Disposition: attachment; filename="test.xlsx"
    Content-Transfer-Encoding: base64

    UEsDBBQABgAIAAAAIQBi7p1oXgEAAJAEAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIo
    UQIAABEAAAAAAAAAAAAAAAAARBcAAGRvY1Byb3BzL2NvcmUueG1sUEsBAi0AFAAGAAgAAAAh
    AGFJCRCJAQAAEQMAABAAAAAAAAAAAAAAAAAAvBkAAGRvY1Byb3BzL2FwcC54bWxQSwUGAAAA
    AAoACgCAAgAAexwAAAAA
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document;
      name="test.docx"
    Content-Disposition: attachment; filename="test.docx"
    Content-Transfer-Encoding: base64

    UEsDBBQABgAIAAAAIQDfpNJsWgEAACAFAAATAAgCW0NvbnRlbnRfVHlwZXNdLnhtbCCiBAIo
    cHAueG1sUEsBAi0AFAAGAAgAAAAhABA0tG9uAQAA4QIAABEAAAAAAAAAAAAAAAAA2xsAAGRv
    Y1Byb3BzL2NvcmUueG1sUEsBAi0AFAAGAAgAAAAhAJ/mlBIqCwAAU3AAAA8AAAAAAAAAAAAA
    AAAAgB4AAHdvcmQvc3R5bGVzLnhtbFBLBQYAAAAACwALAMECAADXKQAAAAA=
    --------------MQ01Z9UM8OaR9z39TvzDfdIq
    Content-Type: text/plain; charset=UTF-8; name="text file.txt"
    Content-Disposition: attachment; filename="text file.txt"
    Content-Transfer-Encoding: base64

    dGV4dCBmaWxl

    --------------MQ01Z9UM8OaR9z39TvzDfdIq--

    """
  Then it succeeds
  When user "[user:user]" connects and authenticates IMAP client "1"
  Then IMAP client "1" eventually sees the following messages in "Sent":
    | from                 | to                    | subject                                                 | X-Forwarded-Message-Id  |
    | [user:user]@[domain] | [user:user2]@[domain] | Fwd: Reply to this message, it has various attachments. | something@protonmail.ch |
  And IMAP client "1" eventually sees 1 messages in "Sent"
  When the user logs in with username "[user:user2]" and password "password"
  And user "[user:user2]" connects and authenticates IMAP client "2"
  And user "[user:user2]" finishes syncing
  And it succeeds
  Then IMAP client "2" eventually sees the following messages in "Inbox":
    | from                 | to                    | subject                                                 | X-Forwarded-Message-Id  |
    | [user:user]@[domain] | [user:user2]@[domain] | Fwd: Reply to this message, it has various attachments. | something@protonmail.ch |
  Then IMAP client "2" eventually sees the following message in "Inbox" with this structure:
    """
    {
      "from": "[user:user]@[domain]",
      "to": "[user:user2]@[domain]",
      "subject": "Fwd: Reply to this message, it has various attachments.",
      "content": {
        "content-type": "multipart/mixed",
        "sections":[
          {
            "content-type": "text/plain",
            "content-type-charset": "utf-8",
            "transfer-encoding": "quoted-printable",
            "body-is": "Forwarding a message with various attachments in it!\r\n\r\n\r\n\r\n-------- Forwarded Message --------\r\nSubject: \tReply to this message, it has various attachments.\r\nDate: \tThu, 26 Oct 2023 10:41:55 +0000\r\nFrom: \tGjorgji Testing <gorgitesting@protonmail.com>\r\nReply-To: \tGjorgji Testing <gorgitesting@protonmail.com>\r\nTo: \tGjorgji Test v3 <gorgitesting3@protonmail.com>\r\n\r\n\r\n\r\n\r\nFor real!\r\n\r\n*Gjorgji Testing\r\nTesASID <https://www.youtube.com/watch?v=3DMifXUbrjYr8>\r\n*"
          },
          {
            "content-type": "text/html",
            "content-type-name": "index.html",
            "content-disposition": "attachment",
            "content-disposition-filename": "index.html",
            "transfer-encoding": "base64"
          },
          {
            "content-type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "content-type-name": "test.docx",
            "content-disposition": "attachment",
            "content-disposition-filename": "test.docx",
            "transfer-encoding": "base64"
          },
          {
            "content-type": "application/pdf",
            "content-type-name": "test.pdf",
            "content-disposition": "attachment",
            "content-disposition-filename": "test.pdf",
            "transfer-encoding": "base64"
          },
          {
            "content-type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            "content-type-name": "test.xlsx",
            "content-disposition": "attachment",
            "content-disposition-filename": "test.xlsx",
            "transfer-encoding": "base64"
          },
          {
            "content-type": "text/xml",
            "content-type-name": "testxml.xml",
            "content-disposition": "attachment",
            "content-disposition-filename": "testxml.xml",
            "transfer-encoding": "base64"
          },
          {
            "content-type": "text/plain",
            "content-type-name": "text file.txt",
            "content-disposition": "attachment",
            "content-disposition-filename": "text file.txt",
            "transfer-encoding": "base64",
            "body-is": "dGV4dCBmaWxl"
          }
        ]
      }
    }
    """