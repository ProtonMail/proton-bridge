Feature: IMAP get mailbox info
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | f1   | folder |
      | f1/f2| folder |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Rename folder with subfolders
    When IMAP client "1" renames "Folders/f1" to "Folders/f3"
    And it succeeds
    Then IMAP client "1" sees "Folders/f3"
    Then IMAP client "1" sees "Folders/f3/f2"
    And  IMAP client "1" does not see "Folders/f1"
    And  IMAP client "1" does not see "Folders/f1/f2"
