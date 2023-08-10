@regression
Feature: Bridge can fully synchronize an account with high number of messages, and correct number of messages is shown in client
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
      | two   | folder |
      | three | folder |
      | four  | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has 1000 messages in "Folders/one"
    And the address "[user:user]@[domain]" of account "[user:user]" has 5450 messages in "Folders/two"
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Folders/three":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | true   |
    And the address "[user:user]@[domain]" of account "[user:user]" has 9800 messages in "Labels/four"
    Then it succeeds
    When bridge starts
    Then it succeeds
  
  Scenario: The account is synced when the user logs in and the number of messages is correct
    When the user logs in with username "[user:user]" and password "password"
    Then bridge sends sync started and finished events for user "[user:user]"
    Then it succeeds
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following mailbox info:
      | name          | total | unread |
      | INBOX         | 0     | 0      |
      | Drafts        | 0     | 0      |
      | Sent          | 0     | 0      |
      | Starred       | 0     | 0      |
      | Archive       | 0     | 0      |
      | Spam          | 0     | 0      |
      | Trash         | 0     | 0      |
      | All Mail      | 16254 | 3      |
      | Folders       | 0     | 0      |
      | Folders/one   | 1000  | 0      |
      | Folders/two   | 5450  | 0      |
      | Folders/three | 4     | 3      |
      | Labels        | 0     | 0      |
      | Labels/four   | 9800  | 0      |
