Feature: IMAP get mailbox info
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "one":
      | sender  | recipient | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing

  Scenario: Mailbox status reports correct name, total and unread
    When user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info for "Folders/one":
      | name         | total | unread |
      | Folders/one  | 2     | 1      |