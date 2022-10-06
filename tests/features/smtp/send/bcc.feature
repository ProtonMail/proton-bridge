Feature: SMTP with bcc
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    Given there exists an account with username "bridgetest@protonmail.com" and password "password"
    Given there exists an account with username "bcc@protonmail.com" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And the user logs in with username "bcc@protonmail.com" and password "password"
    And user "user@pm.me" connects and authenticates SMTP client "1"

  Scenario: Send message to address in to and bcc
    When SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com, bcc@protonmail.com":
      """
      Subject: hello
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then it succeeds
    When user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from       | to                        | bcc                | subject | unread |
      | user@pm.me | bridgetest@protonmail.com | bcc@protonmail.com | hello   | false  |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [
            {
              "Address": "bridgetest@protonmail.com",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bcc@protonmail.com"
            }
          ]
        }
      }
      """


  Scenario: Send message only to bcc
    When SMTP client "1" sends the following message from "user@pm.me" to "bcc@protonmail.com":
      """
      Subject: hello
      From: Bridge Test <user@pm.me>

      hello

      """
    Then it succeeds
    When user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from       | to  | bcc                | subject |
      | user@pm.me |     | bcc@protonmail.com | hello   |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [],
          "CCList": [],
          "BCCList": [
            {
              "Address": "bcc@protonmail.com"
            }
          ]
        }
      }
      """
    When user "bcc@protonmail.com" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from       | to  | bcc | subject | unread |
      | user@pm.me |     |     | hello   | true   |
