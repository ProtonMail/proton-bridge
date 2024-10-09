@gmail-integration
Feature: External sender to Proton recipient sending a plain text message
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then it succeeds
  
  Scenario: Plain text message sent from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      From: <auto.bridge.qa@gmail.com>
      To: <[user:user]@[domain]>
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 8bit
      Subject: Hello World!

      hello

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject      | body  |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Hello World! | hello |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
       """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Hello World!",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "hello"
        }
      }
       """

  Scenario: Plain message with Foreign/Nonascii chars in Subject and Body from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      To: <[user:user]@[domain]>
      From: Bridge Automation <auto.bridge.qa@gmail.com>
      Subject: =?UTF-8?B?U3Vias61zq3Pgs+EIMK2IMOEIMOI?=
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 8bit

      Subjεέςτ ¶ Ä È

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject         | body           |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Subjεέςτ ¶ Ä È  | Subjεέςτ ¶ Ä È |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Subjεέςτ ¶ Ä È",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "Subjεέςτ ¶ Ä È"
        }
      }
      """

  Scenario: Plain message with numbering/ordering in Body from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      To: <[user:user]@[domain]>
      From: Bridge Automation <auto.bridge.qa@gmail.com>
      Subject: Message with Numbering/Ordering in Body
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      **Ordering

      * *Bullet*1
          o Bullet 1.1
      * Bullet 2
          o Bullet 2.1
          o *Bullet 2.2*
              + /Bullet 2.2.1/
          o Bullet 2.3
      * */Bullet 3/*

      Numbering

      1. *Number 1*
          1. */Number/**1.1*
      2. Number 2

          1. */Number 2.1/*
          2. Number 2.2
              1. Number 2.2.1
          3. Number 2.3

      3. /Number 3/

      End
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject                                 |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Message with Numbering/Ordering in Body |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Message with Numbering/Ordering in Body",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "**Ordering\r\n\r\n* *Bullet*1\r\n    o Bullet 1.1\r\n* Bullet 2\r\n    o Bullet 2.1\r\n    o *Bullet 2.2*\r\n        + /Bullet 2.2.1/\r\n    o Bullet 2.3\r\n* */Bullet 3/*\r\n\r\nNumbering\r\n\r\n1. *Number 1*\r\n    1. */Number/**1.1*\r\n2. Number 2\r\n\r\n    1. */Number 2.1/*\r\n    2. Number 2.2\r\n        1. Number 2.2.1\r\n    3. Number 2.3\r\n\r\n3. /Number 3/\r\n\r\nEnd"
        }
      }
      """
      
  Scenario: Plain text message with multiple attachments from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------WI90RPIYF20K6dGXjs7dm2mi"
      Subject: Plain message with different attachments

      This is a multi-part message in MIME format.
      --------------WI90RPIYF20K6dGXjs7dm2mi
      Content-Type: text/plain; charset=UTF-8;
      Content-Transfer-Encoding: 7bit

      Hello, this is a Plain message with different attachments.

      --------------WI90RPIYF20K6dGXjs7dm2mi
      Content-Type: text/html; charset=UTF-8; name="index.html"
      Content-Disposition: attachment; filename="index.html"
      Content-Transfer-Encoding: base64

      PCFET0NUWVBFIGh0bWw+
      --------------WI90RPIYF20K6dGXjs7dm2mi
      Content-Type: text/xml; charset=UTF-8; name="testxml.xml"
      Content-Disposition: attachment; filename="testxml.xml"
      Content-Transfer-Encoding: base64

      PD94bWwgdmVyc2lvbj0iMS4xIj8+PCFET0NUWVBFIF9bPCFFTEVNRU5UIF8gRU1QVFk+XT48
      Xy8+
      --------------WI90RPIYF20K6dGXjs7dm2mi
      Content-Type: text/plain; charset=UTF-8; name="text file.txt"
      Content-Disposition: attachment; filename="text file.txt"
      Content-Transfer-Encoding: base64

      dGV4dCBmaWxl
      --------------WI90RPIYF20K6dGXjs7dm2mi
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject                                  |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Plain message with different attachments |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Plain message with different attachments",
        "content": {
            "content-type": "multipart/mixed",
            "sections": [
                {
                    "content-type": "text/plain",
                    "content-type-charset": "utf-8",
                    "transfer-encoding": "quoted-printable",
                    "body-is": "Hello, this is a Plain message with different attachments."
                },
                {
                    "content-type": "text/plain",
                    "content-type-name": "text file.txt",
                    "content-disposition": "attachment",
                    "content-disposition-filename": "text file.txt",
                    "transfer-encoding": "base64",
                    "body-is": "dGV4dCBmaWxl"
                },
                {
                    "content-type": "text/html",
                    "content-type-name": "index.html",
                    "content-disposition": "attachment",
                    "content-disposition-filename": "index.html",
                    "transfer-encoding": "base64",
                    "body-is": "PCFET0NUWVBFIGh0bWw+"
                },
                {
                    "content-type": "text/xml",
                    "content-type-name": "testxml.xml",
                    "content-disposition": "attachment",
                    "content-disposition-filename": "testxml.xml",
                    "transfer-encoding": "base64",
                    "body-is": "PD94bWwgdmVyc2lvbj0iMS4xIj8+PCFET0NUWVBFIF9bPCFFTEVNRU5UIF8gRU1QVFk+XT48Xy8+"
                }
            ]
        }
      }
      """

  Scenario: Plain message with multiple inline images from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      To: <[user:user]@[domain]>
      From: <auto.bridge.qa@gmail.com>
      Subject: Plain message with multiple inline images to Internal
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Plain message with image 1 multiple image 2 inline image 3 images.

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject                                               |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Plain message with multiple inline images to Internal |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Plain message with multiple inline images to Internal",
        "content": {
          "content-type": "text/plain",
          "content-type-charset": "utf-8",
          "transfer-encoding": "quoted-printable",
          "body-is": "Plain message with image 1 multiple image 2 inline image 3 images.",
          "body-contains": "",
          "sections": []
        }
      }
      """

  Scenario: Plain text message with a large attachment from External to Internal
    Given external client sends the following message from "auto.bridge.qa@gmail.com" to "[user:user]@[domain]":
      """
      Content-Type: multipart/mixed; boundary="------------k0Z3FJiZsGaSFqdJGsr0Oml6"
      To: <[user:user]@[domain]>
      From: Bridge Automation <auto.bridge.qa@gmail.com>
      Subject: Plain message with a large attachment

      This is a multi-part message in MIME format.
      --------------k0Z3FJiZsGaSFqdJGsr0Oml6
      Content-Type: text/plain; charset=UTF-8; format=flowed
      Content-Transfer-Encoding: 7bit

      Hello, this is a plain message with a large attachment.

      --------------k0Z3FJiZsGaSFqdJGsr0Oml6
      Content-Type: application/msword; name="testDoc.doc"
      Content-Disposition: attachment; filename="testDoc.doc"
      Content-Transfer-Encoding: base64
      --------------k0Z3FJiZsGaSFqdJGsr0Oml6--
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Inbox":
      | from                     | to                   | subject                               |
      | auto.bridge.qa@gmail.com | [user:user]@[domain] | Plain message with a large attachment |
    And IMAP client "1" eventually sees the following message in "Inbox" with this structure:
      """
      {
        "from": "auto.bridge.qa@gmail.com",
        "to": "[user:user]@[domain]",
        "subject": "Plain message with a large attachment",
        "content": {
            "content-type": "text/plain",
            "content-type-charset": "utf-8"
        }
      }
      """