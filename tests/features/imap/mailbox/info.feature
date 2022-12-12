Feature: IMAP get mailbox info
  Background:
    Given there exists an account with username "user" and password "password"
    And the account "user" has the following custom mailboxes:
      | name | type   |
      | one  | folder |
    And the address "user@[domain]" of account "user" has the following messages in "one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And bridge starts
    And the user logs in with username "user" and password "password"
    And user "user" finishes syncing

  Scenario: Mailbox status reports correct name, total and unread
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info for "Folders/one":
      | name        | total | unread |
      | Folders/one | 2     | 1      |