Feature: SMTP with bcc
  Background:
    Given there exists an account with username "user" and password "password"
    Given there exists an account with username "bridgetest" and password "password"
    Given there exists an account with username "bcc" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And the user logs in with username "bcc" and password "password"
    And user "user" connects and authenticates SMTP client "1"

  Scenario: Send message to address in to and bcc
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain], bcc@[domain]":
      """
      Subject: hello
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      hello

      """
    Then it succeeds
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to                  | bcc          | subject | unread |
      | user@[domain] | bridgetest@[domain] | bcc@[domain] | hello   | false  |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [
            {
              "Address": "bridgetest@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bcc@[domain]"
            }
          ]
        }
      }
      """


  Scenario: Send message only to bcc
    When SMTP client "1" sends the following message from "user@[domain]" to "bcc@[domain]":
      """
      Subject: hello
      From: Bridge Test <user@[domain]>

      hello

      """
    Then it succeeds
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to | bcc          | subject |
      | user@[domain] |    | bcc@[domain] | hello   |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bcc@[domain]"
            }
          ]
        }
      }
      """
    When user "bcc" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from          | to | bcc | subject | unread |
      | user@[domain] |    |     | hello   | true   |
