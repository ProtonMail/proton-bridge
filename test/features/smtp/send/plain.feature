Feature: SMTP sending of plain messages
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"

  Scenario: Only from and to headers to internal account
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject |
      | [userAddress] | bridgetest@protonmail.com |         |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "",
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
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Only from and to headers to external account
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: External Bridge <pm.bridge.qa@gmail.com>

      hello

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject |
      | [userAddress] | pm.bridge.qa@gmail.com |         |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "",
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

  Scenario: Basic message to internal account
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Plain text internal
      Content-Disposition: inline
      Content-Type: text/plain; charset=utf-8

      This is body of mail 👋

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | subject             |
      | [userAddress] | bridgetest@protonmail.com | Plain text internal |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain text internal",
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
          "MIMEType": "text/plain"
        }
      }
      """

  Scenario: Basic message to external account
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain text external
      Content-Disposition: inline
      Content-Type: text/plain; charset=utf-8

      This is body of mail 👋

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject             |
      | [userAddress] | pm.bridge.qa@gmail.com | Plain text external |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain text external",
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

  Scenario: Message without charset is utf8
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain text no charset external
      Content-Disposition: inline
      Content-Type: text/plain;

      This is body of mail without charset. Please assume utf8

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject                        |
      | [userAddress] | pm.bridge.qa@gmail.com | Plain text no charset external |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain text no charset external",
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

  Scenario: Message without charset is base64-encoded latin1
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain text no charset external
      Content-Disposition: inline
      Content-Type: text/plain;
      Content-Transfer-Encoding: base64

      dGhpcyBpcyBpbiBsYXRpbjEgYW5kIHRoZXJlIGFyZSBsb3RzIG9mIGVzIHdpdGggYWNjZW50czog
      6enp6enp6enp6enp6enp


      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject                        |
      | [userAddress] | pm.bridge.qa@gmail.com | Plain text no charset external |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain text no charset external",
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

  Scenario: Message without charset and content is detected as HTML
    When SMTP client sends message
      """
      From: Bridge Test <[userAddress]>
      To: External Bridge <pm.bridge.qa@gmail.com>
      Subject: Plain, no charset, no content, external
      Content-Disposition: inline
      Content-Type: text/plain;

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                     | subject                                 |
      | [userAddress] | pm.bridge.qa@gmail.com | Plain, no charset, no content, external |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "Plain, no charset, no content, external",
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

  Scenario: RCPT does not contain all CC
    When SMTP client sends "MAIL FROM:<[userAddress]>"
    Then SMTP response is "OK"
    When SMTP client sends "RCPT TO:<bridgetest@protonmail.com>"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "OK"
    When SMTP client sends 
      """
      From: Bridge Test <[userAddress]>
      To: Internal Bridge <bridgetest@protonmail.com>
      CC: Internal Bridge 2 <bridgetest2@protonmail.com>
      Content-Type: text/plain
      Subject: RCPT-CC test

      This is CC missing in RCPT test. Have a nice day!
.
      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | from          | to                        | cc                         | subject      |
      | [userAddress] | bridgetest@protonmail.com | bridgetest2@protonmail.com | RCPT-CC test |
    And message is sent with API call
      """
      {
        "Message": {
          "Subject": "RCPT-CC test",
          "Sender": {
            "Name": "Bridge Test"
          },
          "ToList": [
            {
              "Address": "bridgetest@protonmail.com",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [
            {
              "Address": "bridgetest2@protonmail.com",
              "Name": "Internal Bridge 2"
            }
          ],
          "BCCList": []
        }
      }
      """
    And packages are sent with API call
      """
      {
        "Packages":[
            {
              "Addresses":{
                  "bridgetest@protonmail.com":{
                    "Type":1
                  },
                  "bridgetest2@protonmail.com":{
                    "Type":1
                  }
              },
              "Type":1,
              "MIMEType":"text/plain"
            }
        ]
      }
      """