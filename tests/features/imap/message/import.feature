Feature: IMAP import messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Basic message import
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@example.com>",
        "subject": "Basic text/plain message",
        "content": {
          "content-type": "text/plain",
          "body-is": "Hello"
        }
      }
      """

  Scenario: Import message with double charset in content type
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Message with double charset in content type
      Content-Type: text/plain; charset=utf-8; charset=utf-8
      Content-Disposition: inline
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@example.com>",
        "subject": "Message with double charset in content type",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "content-disposition": "",
          "transfer-encoding": "quoted-printable",
          "body-is": "Hello"
        }
      }
      """


  Scenario Outline: Import multipart message with attachment <message>
    When IMAP client "1" appends <message> to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@protonmail.com>",
        "subject": "Message with attachment name",
        "body-contains": "Hello",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/plain",
              "body-is": "Hello"
            },
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "7bit",
              "body-contains": "HELLO"
            },
            {
              "content-type": "application/pdf",
              "content-type-name": "filename",
              "content-disposition": "attachment",
              "content-disposition-filename": "filename",
              "body-is": "somebytes"
            }
          ]
        }
      }
      """
    Examples:
      | message                                                |
      | "multipart/mixed_with_attachment_encoded.eml"          |
      | "multipart/mixed_with_attachment_encoded_no_quote.eml" |
      | "multipart/mixed_with_attachment_no_quote.eml"         |


  # The message is imported as UTF-8 and the content type is determined at build time.
  Scenario: Import message as latin1 without content type
    When IMAP client "1" appends "plain/text_plain_unknown_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Sender <sender@pm.me>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Receiver <receiver@pm.me>",
        "content": {
          "content-type": "text/plain",
          "body-is": "ééééééé"
        }
      }
      """

  # The message is imported and the body is converted to UTF-8.
  Scenario: Import message as latin1 with content type
    When IMAP client "1" appends "plain/text_plain_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Sender <sender@pm.me>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Receiver <receiver@pm.me>",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "body-is": "ééééééé"
        }
      }
      """


  # The message is imported anad the body is wrongly converted (body is corrupted).
  Scenario: Import message as latin1 with wrong content type
    When IMAP client "1" appends "plain/text_plain_wrong_latin1.eml" to "INBOX"
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Sender <sender@pm.me>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Receiver <receiver@pm.me>",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "body-is": ""
        }
      }
      """

  Scenario: Import received message to Sent
    When IMAP client "1" appends the following message to "Sent":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "Sent" with this structure:
      """
      {
        "from": "Foo <foo@example.com>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Bridge Test <bridgetest@pm.test>",
        "subject": "Hello",
        "content": {
          "content-type": "text/plain",
          "body-is": "Hello"
        }
      }
      """
    And IMAP client "1" eventually sees 0 messages in "Inbox"

  Scenario: Import non-received message to Inbox
    When IMAP client "1" appends the following message to "Inbox":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Foo <foo@example.com>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Bridge Test <bridgetest@pm.test>",
        "subject": "Hello",
        "content": {
          "content-type": "text/plain",
          "body-is": "Hello"
        }
      }
      """
    And IMAP client "1" eventually sees 0 messages in "Sent"


  Scenario: Import non-received message to Sent
    When IMAP client "1" appends the following message to "Sent":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Hello

      Hello
      """
    Then it succeeds
    And IMAP client "1" eventually sees 0 messages in "Inbox"
    And IMAP client "1" eventually sees the following message in "Sent" with this structure:
      """
      {
        "from": "Foo <foo@example.com>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Bridge Test <bridgetest@pm.test>",
        "subject": "Hello",
        "content": {
          "content-type": "text/plain",
          "body-is": "Hello"
        }
      }
      """

  Scenario Outline: Import message without sender to <mailbox>
    When IMAP client "1" appends the following message to "<mailbox>":
      """
      From: Somebody@somewhere.org
      Date: 01 Jan 1980 00:00:00 +0000
      To: Lionel Richie <lionel@richie.com>
      Subject: RE: Hello, is it me you looking for?

      Nope.
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "<mailbox>" with this structure:
      """
      {
        "from": "Somebody@somewhere.org",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Lionel Richie <lionel@richie.com>",
        "subject": "RE: Hello, is it me you looking for?",
        "content": {
          "content-type": "text/plain",
          "content-type-charset":"utf-8",
          "transfer-encoding":"quoted-printable",
          "body-is": "Nope."
        }
      }
      """
    Examples:
      | mailbox |
      | Archive |
      | Sent    |

  # The date returned from black is server time.. Black is probably correct we need to fix GPA server
  @skip-black
  Scenario: Import message without sender to Drafts
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: Somebody@somewhere.org
      Date: 01 Jan 1980 00:00:00 +0000
      To: Lionel Richie <lionel@richie.com>
      Subject: RE: Hello, is it me you looking for?

      Nope.
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "Drafts" with this structure:
      """
      {
        "date": "01 Jan 01 00:00 +0000",
        "to": "Lionel Richie <lionel@richie.com>",
        "subject": "RE: Hello, is it me you looking for?",
        "content": {
          "content-type": "text/plain",
          "content-type-charset":"utf-8",
          "transfer-encoding":"quoted-printable",
          "body-is": "Nope."
        }
      }
      """


  Scenario: Import embedded message
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Foo <foo@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Bridge Test <bridgetest@pm.test>
      Subject: Embedded message
      Content-Type: multipart/mixed; boundary="boundary"
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000

      This is a multi-part message in MIME format.

      --boundary
      Content-Type: text/plain; charset=utf-8
      Content-Transfer-Encoding: 7bit

      Hello

      --boundary
      Content-Type: text/html; charset=utf-8
      Content-Transfer-Encoding: 7bit

      <h1> HELLO </h1>

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
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Foo <foo@example.com>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Bridge Test <bridgetest@pm.test>",
        "subject": "Embedded message",
        "body-contains": "Hello",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/plain",
              "content-type-charset": "utf-8",
              "transfer-encoding": "7bit",
              "body-is": "Hello"
            },
            {
              "content-type": "text/html",
              "content-type-charset": "utf-8",
              "transfer-encoding": "7bit",
              "body-contains": "HELLO"
            },
            {
              "content-type": "message/rfc822",
              "content-type-name": "embedded.eml",
              "transfer-encoding": "7bit",
              "content-disposition": "attachment",
              "content-disposition-filename": "embedded.eml",
              "body-is": "From: Bar <bar@example.com>\nTo: Bridge Test <bridgetest@pm.test>\nSubject: (No Subject)\nContent-Type: text/plain; charset=utf-8\nContent-Transfer-Encoding: quoted-printable\n\nhello"
            }
          ]
        }
      }
      """

  @regression
  Scenario: Import message with remote content
    When IMAP client "1" appends the following message to "Inbox":
    """
    Date: 01 Jan 1980 00:00:00 +0000
    To: Bridge Test <bridge@test.com>
    From: Bridge Second Test <bridge_second@test.com>
    Subject: MESSAGE WITH REMOTE CONTENT
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
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
    """
    {
      "date": "01 Jan 80 00:00 +0000",
      "to": "Bridge Test <bridge@test.com>",
      "from": "Bridge Second Test <bridge_second@test.com>",
      "subject": "MESSAGE WITH REMOTE CONTENT",
      "content": {
        "content-type": "multipart/mixed",
        "sections":[
          {
            "content-type": "multipart/alternative",
            "sections":[
              {
                "content-type": "text/plain",
                "content-type-charset": "utf-8",
                "transfer-encoding": "7bit",
                "body-is": "Remote content\n\n\nBridge\n\n\nRemote content"
              },
              {
                "content-type": "text/html",
                "content-type-charset": "utf-8",
                "transfer-encoding": "7bit",
                "body-is": "<!DOCTYPE html>\n<html>\n  <head>\n\n    <meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">\n  </head>\n  <body>\n    <p><tt>Remote content</tt></p>\n    <p><tt><br>\n      </tt></p>\n    <p><img\n        src=\"https://bridgeteam.protontech.ch/bridgeteam/tmp/bridge.jpg\"\n        alt=\"Bridge\" width=\"180\" height=\"180\"></p>\n    <p><br>\n    </p>\n    <p><tt>Remote content</tt><br>\n    </p>\n    <br>\n  </body>\n</html>"
              }
            ]
          }
        ]
      }
    }
    """


  Scenario: Import message with inline image
    When IMAP client "1" appends the following message to "Inbox":
      """
      Date: 01 Jan 1980 00:00:00 +0000
      From: Bridge Second Test <bridge_second@test.com>
      To: Bridge Test <bridge@test.com>
      Subject: Html Inline Importing
      Content-Disposition: inline
      User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.5.0
      MIME-Version: 1.0
      Content-Language: en-US
      Content-Type: multipart/related; boundary="61FA22A41A3F46E8E90EF528"

      This is a multi-part message in MIME format.
      --61FA22A41A3F46E8E90EF528
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

      --61FA22A41A3F46E8E90EF528
      Content-Type: image/gif; name="email-action-left.gif"
      Content-Transfer-Encoding: base64
      Content-ID: <part1.D96BFAE9.E2E1CAE3@protonmail.com>
      Content-Disposition: inline; filename="email-action-left.gif"

      R0lGODlhAQABAAAAADs=

      --61FA22A41A3F46E8E90EF528--

      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
    """
    {
      "date": "01 Jan 80 00:00 +0000",
      "to": "Bridge Test <bridge@test.com>",
      "from": "Bridge Second Test <bridge_second@test.com>",
      "subject": "Html Inline Importing",
      "content": {
        "content-type": "multipart/mixed",
        "sections":[
          {
            "content-type": "multipart/related",
            "sections":[
              {
                "content-type": "text/html",
                "content-type-charset": "utf-8",
                "transfer-encoding": "7bit",
                "body-is": "<html>\n<head>\n<meta http-equiv=\"content-type\" content=\"text/html; charset=UTF-8\">\n</head>\n<body text=\"#000000\" bgcolor=\"#FFFFFF\">\n<p><br>\n</p>\n<p>Behold! An inline <img moz-do-not-send=\"false\"\nsrc=\"cid:part1.D96BFAE9.E2E1CAE3@protonmail.com\" alt=\"\"\nwidth=\"24\" height=\"24\"><br>\n</p>\n</body>\n</html>"
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

  Scenario: Message import with text part and attachment
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Message import with text part
      Content-Type: multipart/mixed; boundary="BOUNDARY"

      This is a multi-part message in MIME format.

      --BOUNDARY
      Content-Type: text/plain; charset=utf-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Hello World

      --BOUNDARY
      Content-Disposition: attachment; filename=image.png
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==

      --BOUNDARY--
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@example.com>",
        "subject": "Message import with text part",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/plain",
              "body-is": "Hello World"
            },
            {
              "content-type": "image/png",
              "content-type-name": "image.png",
              "content-disposition": "attachment",
              "content-disposition-filename": "image.png",
              "transfer-encoding": "base64",
              "body-is": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAA\r\nAABJRU5ErkJggg=="
            }
          ]
        }
      }
      """


  Scenario: Message import without text part
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Message import without text part
      Content-Type: multipart/mixed; boundary="BOUNDARY"

      This is a multi-part message in MIME format.

      --BOUNDARY
      Content-Disposition: attachment; filename=image.png
      Content-Transfer-Encoding: base64
      Content-Type: image/png

      iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQot
      tAAAAABJRU5ErkJggg==

      --BOUNDARY--
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
      """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@example.com>",
        "subject": "Message import without text part",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "text/plain",
              "body-is": ""
            },
            {
              "content-type": "image/png",
              "content-type-name": "image.png",
              "content-disposition": "attachment",
              "content-disposition-filename": "image.png",
              "transfer-encoding": "base64",
              "body-is": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAA\r\nAABJRU5ErkJggg=="
            }
          ]
        }
      }
      """

  Scenario: Message import multipart/related with invalid boundary character
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgetest@pm.test>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Message with invalid boundary
      Content-Type: multipart/related; boundary="------------123456789@tutanota"

      --------------123456789@tutanota
      Content-Type: text/html; charset=UTF-8
      Content-transfer-encoding: base64

      PGRpdiBjbGFzcz0iIj4KPHAgY2xhc3M9IiI+PGEgbmFtZT0iX0hsazE5MDA1NjM2IiByZWw9Im5vb3
      BlbmVyIG5vcmVmZXJyZXIiIHRhcmdldD0iX2JsYW5rIj48c3BhbiBzdHlsZT0ibXNvLWZhcmVhc3Qt

      --------------123456789@tutanota
      Content-Type: image/png;
       name==?UTF-8?B?MC5wbmc=?=
      Content-Transfer-Encoding: base64
      Content-Disposition: attachment;
       filename=image1.png

      iVBORw0KGgoAAAANSUhEUgAAACsAAAArCAYAAADhXXHAAAAPq3pUWHRSYXcgcHJvZmlsZSB0eXBlIG
      V4aWYAAHjarZlrliOpkoT/s4pZAuCAw3J4njM7mOXP54SUlZmV1bd7plNVEVIoAhx/mJsht//nv4/7

      --------------123456789@tutanota
      Content-Type: image/png;
       name==?UTF-8?B?Mi5wbmc=?=
      Content-Transfer-Encoding: base64
      Content-Disposition: attachment;
       filename=img2.png

      iVBORw0KGgoAAAANSUhEUgAAACsAAAArCAYAAADhXXHAAAAR+HpUWHRSYXcgcHJvZmlsZSB0eXBlIG
      V4aWYAAHjarZprdhs5DoX/cxWzBD4Bcjl8njM7mOXPB5bsOI49SU+3nViKLFWxgIv7YMXt//z7uH/x

      --------------123456789@tutanota--

      """
    Then it succeeds
    And IMAP client "1" eventually sees the following message in "INBOX" with this structure:
    """
      {
        "from": "Bridge Test <bridgetest@pm.test>",
        "date": "01 Jan 80 00:00 +0000",
        "to": "Internal Bridge <bridgetest@example.com>",
        "subject": "Message with invalid boundary",
        "content": {
          "content-type": "multipart/mixed",
          "sections":[
            {
              "content-type": "multipart/related",
              "sections": [
                {
                  "content-type": "text/html",
                  "transfer-encoding": "base64",
                  "body-is": "PGRpdiBjbGFzcz0iIj4KPHAgY2xhc3M9IiI+PGEgbmFtZT0iX0hsazE5MDA1NjM2IiByZWw9Im5v\r\nb3BlbmVyIG5vcmVmZXJyZXIiIHRhcmdldD0iX2JsYW5rIj48c3BhbiBzdHlsZT0ibXNvLWZhcmVh\r\nc3Qt"
                },
                {
                  "content-type": "image/png",
                  "transfer-encoding": "base64",
                  "content-disposition": "attachment",
                  "content-disposition-filename": "image1.png",
                  "body-is": "iVBORw0KGgoAAAANSUhEUgAAACsAAAArCAYAAADhXXHAAAAPq3pUWHRSYXcgcHJvZmlsZSB0eXBl\r\nIGV4aWYAAHjarZlrliOpkoT/s4pZAuCAw3J4njM7mOXP54SUlZmV1bd7plNVEVIoAhx/mJsht//n\r\nv4/7"
                },
                {
                  "content-type": "image/png",
                  "transfer-encoding": "base64",
                  "content-disposition": "attachment",
                  "content-disposition-filename": "img2.png",
                  "body-is": "iVBORw0KGgoAAAANSUhEUgAAACsAAAArCAYAAADhXXHAAAAR+HpUWHRSYXcgcHJvZmlsZSB0eXBl\r\nIGV4aWYAAHjarZprdhs5DoX/cxWzBD4Bcjl8njM7mOXPB5bsOI49SU+3nViKLFWxgIv7YMXt//z7\r\nuH/x"
                }
              ]
            }
          ]
        }
      }
      """
