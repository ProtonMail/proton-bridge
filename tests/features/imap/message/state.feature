Feature: IMAP change state of message in mailbox
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | one  | folder |
      | two  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has 5 messages in "Folders/one"
    And the address "[user:user]@[domain]" of account "[user:user]" has 5 messages in "Folders/two"
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from          | to            | subject | unread |
      | a@example.com | b@example.com | one     | true   |
      | c@example.com | d@example.com | two     | false  |
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"

  Scenario: Mark message as read
    When IMAP client "1" selects "Folders/one"
    And IMAP client "1" marks message 1 as "read"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "\Seen"

  Scenario: Mark message as unread
    When IMAP client "1" selects "Folders/one"
    And IMAP client "1" marks message 1 as "unread"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 does not have the flag "\Seen"

  Scenario: Mark message as starred
    When IMAP client "1" selects "Folders/one"
    And IMAP client "1" marks message 1 as "starred"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "\Flagged"
  
  Scenario: Mark message as unstarred
    When IMAP client "1" selects "Folders/one"
    And IMAP client "1" marks message 1 as "unstarred"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 does not have the flag "\Flagged"

  Scenario: Mark message with subject as read/unread
    When IMAP client "1" selects "Inbox"
    And IMAP client "1" marks the message with subject "one" as "read"
    And it succeeds
    And IMAP client "1" marks the message with subject "two" as "unread"
    And it succeeds
    Then IMAP client "1" eventually sees that the message with subject "one" has the flag "\Seen"
    And IMAP client "1" eventually sees that the message with subject "two" does not have the flag "\Seen"

  Scenario: Mark all messages in folder as read/unread
    When IMAP client "1" selects "Folders/two"
    And IMAP client "1" marks all messages as "read"
    And it succeeds
    Then IMAP client "1" eventually sees that all the messages have the flag "\Seen"
    When IMAP client "1" marks all messages as "unread"
    And it succeeds
    Then IMAP client "1" eventually sees that all the messages do not have the flag "\Seen"
