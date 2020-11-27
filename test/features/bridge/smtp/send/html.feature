Feature: SMTP sending of HTML messages
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"

  Scenario: HTML message to external account
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: HTML text external
      Content-Disposition: inline
      Content-Transfer-Encoding: quoted-printable
      Content-Type: text/html; charset=utf-8
      In-Reply-To: <base64hashOfSomeMessage@protonmail.internalid>

      <html><body>This is body of <b>HTML mail</b> without attachment<body></html>

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject            |
      | [userAddress] | pm.bridge.qa@gmail.com | HTML text external |
    And message is sent with API call
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
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject              |
      | [userAddress] | pm.bridge.qa@gmail.com | Html Inline External |
    And message is sent with API call
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

  Scenario: HTML message with alternative inline to internal account
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject                          |
      | [userAddress] | bridgetest@protonmail.com | Html Inline Alternative Internal |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Html Inline Alternative Internal",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "bridgetest@protonmail.com",
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
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
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
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject                          |
      | [userAddress] | pm.bridge.qa@gmail.com | Html Inline Alternative External |
    And message is sent with API call
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
