Feature: User deletion with IMAP data removal
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | one  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Folders/one":
      | from          | to            | subject | unread |
      | a@example.com | a@example.com | one     | true   |
      | b@example.com | b@example.com | two     | false  |
      | c@example.com | c@example.com | three   | true   |
      | c@example.com | c@example.com | four    | false   |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    Then it succeeds

  Scenario: User is deleted from Bridge and IMAP data is removed
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info for "Folders/one":
      | name        | total | unread |
      | Folders/one | 4     | 2      |
    And user "[user:user]" is deleted alongside IMAP data for client "1"
    Then it succeeds
