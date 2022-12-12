Feature: SMTP sending with APPENDing to Sent
  Background:
    Given there exists an account with username "user" and password "password"
    And there exists an account with username "bridgetest" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And user "user" connects and authenticates SMTP client "1"
    And user "user" connects and authenticates IMAP client "1"

  Scenario: Send message and append to Sent
    # First do sending.
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Manual send and append
      Message-ID: bridgemessage42

      hello

      """
    Then it succeeds
    And the body in the "POST" request to "/mail/v4/messages" is:
      """
      {
        "Message": {
          "Subject": "Manual send and append",
          "ExternalID": "bridgemessage42"
        }
      }
      """
    And IMAP client "1" eventually sees the following messages in "Sent":
      | to                  | subject                | body  | message-id        |
      | bridgetest@[domain] | Manual send and append | hello | <bridgemessage42> |

    # Then simulate manual append to Sent mailbox - message should be detected as a duplicate.
    When IMAP client "1" appends the following message to "Sent":
      """
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Manual send and append
      Message-ID: bridgemessage42

      hello

      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | to                  | subject                | body  | message-id        |
      | bridgetest@[domain] | Manual send and append | hello | <bridgemessage42> |