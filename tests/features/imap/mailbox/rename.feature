Feature: IMAP get mailbox info
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | f1   | folder |
      | l1   | label  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Rename folder
    When IMAP client "1" renames "Folders/f1" to "Folders/f2"
    Then IMAP client "1" sees "Folders/f2"
    And  IMAP client "1" does not see "Folders/f1"

  Scenario: Rename label
    When IMAP client "1" renames "Labels/l1" to "Labels/l2"
    Then IMAP client "1" sees "Labels/l2"
    And  IMAP client "1" does not see "Labels/l1"

  Scenario: Renaming folder to label is not possible
    When IMAP client "1" renames "Folders/f1" to "Labels/f2"
    Then it fails

  Scenario: Renaming system folder is not possible
    When IMAP client "1" renames "Labels/l1" to "Folders/l2"
    Then it fails