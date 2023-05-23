Feature: SMTP with bcc
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    And there exists an account with username "[user:bcc]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:bcc]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  Scenario: Send message to address in to and bcc
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain], [user:bcc]@[domain]":
      """
      Subject: hello
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>

      hello

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | bcc                 | subject | unread |
      | [user:user]@[domain] | [user:to]@[domain] | [user:bcc]@[domain] | hello   | false  |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [
            {
              "Address": "[user:to]@[domain]",
              "Name": "Internal Bridge"
            }
          ],
          "CCList": [],
          "BCCList": [
            {
              "Address": "[user:bcc]@[domain]"
            }
          ]
        }
      }
      """


  @long-black
  Scenario: Send message only to bcc
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:bcc]@[domain]":
      """
      Subject: hello
      From: Bridge Test <[user:user]@[domain]>

      hello

      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to | bcc                 | subject |
      | [user:user]@[domain] |    | [user:bcc]@[domain] | hello   |
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "hello",
          "ToList": [],
          "CCList": [],
          "BCCList": [
            {
              "Address": "[user:bcc]@[domain]"
            }
          ]
        }
      }
      """
    When user "[user:bcc]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from                 | to | bcc | subject | unread |
      | [user:user]@[domain] |    |     | hello   | true   |
