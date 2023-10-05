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
              "content-type-name": <filename>,
              "content-disposition": "attachment",
              "content-disposition-filename": <filename>,
              "body-is": "somebytes"
            }
          ]
        }
      }
      """
    Examples:
      | message                                                | filename                  |
      | "multipart/mixed_with_attachment_encoded.eml"          | "=?US-ASCII?Q?filename?=" |
#      | "multipart/mixed_with_attachment_encoded_no_quote.eml" | =?US-ASCII?Q?filename?=   | @todo GODT-2966
#      | "multipart/mixed_with_attachment_no_quote.eml"         | "filename"                | @todo GODT-2966

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
