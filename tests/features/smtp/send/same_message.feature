Feature: SMTP sending the same message twice
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:to]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: Hello

      World
      """
    And it succeeds

  @long-black
  Scenario: The exact same message is not sent twice
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: Hello

      World
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject | body  |
      | [user:user]@[domain] | [user:to]@[domain] | Hello   | World |
    When user "[user:to]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from                 | to                 | subject | body  |
      | [user:user]@[domain] | [user:to]@[domain] | Hello   | World |


  @long-black
  Scenario: Slight change means different message and is sent twice
    When SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>
      Subject: Hello.

      World
      """
    Then it succeeds
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject | body  |
      | [user:user]@[domain] | [user:to]@[domain] | Hello   | World |
      | [user:user]@[domain] | [user:to]@[domain] | Hello.  | World |
    When user "[user:to]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from                 | to                 | subject | body  |
      | [user:user]@[domain] | [user:to]@[domain] | Hello   | World |
      | [user:user]@[domain] | [user:to]@[domain] | Hello.  | World |