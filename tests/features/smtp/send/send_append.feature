Feature: SMTP sending with APPENDing to Sent
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Send message and append to Sent
    # First do sending.
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: [user:user]@[domain]
      To: Internal Bridge <[user:to]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
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
      | to                 | subject                | body  | message-id        |
      | [user:to]@[domain] | Manual send and append | hello | <bridgemessage42> |

    # Then simulate manual append to Sent mailbox - message should be detected as a duplicate.
    When IMAP client "1" appends the following message to "Sent":
      """
      From: [user:user]@[domain]
      To: Internal Bridge <[user:to]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Manual send and append
      Message-ID: bridgemessage42

      hello

      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | to                 | subject                | body  | message-id        |
      | [user:to]@[domain] | Manual send and append | hello | <bridgemessage42> |