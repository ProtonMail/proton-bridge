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

  # black fails to get parent ID
  @skip-black
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

  # black is changing order of attachments
  @skip-black
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

      R0lGODlhAQABAAAAADs=
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
                "body-is": "<html>\r\n<head>\r\n<meta http-equiv=3D\"content-type\" content=3D\"text/html; charset=3DUTF-8\">\r\n</head>\r\n<body text=3D\"#000000\" bgcolor=3D\"#FFFFFF\">\r\n<p><br>\r\n</p>\r\n<p>Behold! An inline <img moz-do-not-send=3D\"false\"\r\nsrc=3D\"cid:part1.D96BFAE9.E2E1CAE3@protonmail.com\" alt=3D\"\"\r\nwidth=3D\"24\" height=3D\"24\"><br>\r\n</p>\r\n</body>\r\n</html>"
              },
              {
                "content-type": "image/gif",
                "content-type-name": "email-action-left.gif",
                "content-disposition": "inline",
                "content-disposition-filename": "email-action-left.gif",
                "transfer-encoding": "base64",
                "body-is": "R0lGODlhAQABAAAAADs="
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

      R0lGODlhAQABAAAAADs=
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

      R0lGODlhAQABAAAAADs=
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

  # black fails to get parent ID
  @skip-black
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
    When user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following EML "html/foreign_ascii_subject_body.template.eml" from "[user:user]@[domain]" to "pm.bridge.qa@gmail.com"
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                     | to                     | subject        |
      | [user:user]@[domain]     | pm.bridge.qa@gmail.com | Subjεέςτ ¶ Ä È |
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
  # Black changes order of attachments
  @skip-black
  Scenario: HTML message with remote content in Body
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:user2]@[domain]":
    """
    Date: 01 Jan 1980 00:00:00 +0000
    To: Internal Bridge Test <[user:user2]@[domain]>
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
      "to": "Internal Bridge Test <[user:user2]@[domain]>",
      "from": "Bridge Test <[user:user]@[domain]>",
      "subject": "MESSAGE WITH REMOTE CONTENT SENT",
      "content": {
        "content-type": "text/html",
        "content-type-charset": "utf-8",
        "transfer-encoding": "quoted-printable",
        "body-is": "<!DOCTYPE html><html><head>\r\n\r\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\"/>\r\n  </head>\r\n  <body>\r\n    <p><tt>Remote content</tt></p>\r\n    <p><tt><br/>\r\n      </tt></p>\r\n    <p><img src=\"https://bridgeteam.protontech.ch/bridgeteam/tmp/bridge.jpg\" alt=\"Bridge\" width=\"180\" height=\"180\"/></p>\r\n    <p><br/>\r\n    </p>\r\n    <p><tt>Remote content</tt><br/>\r\n    </p>\r\n    <br/>\r\n  \r\n\r\n</body></html>"
      }
    }
    """
