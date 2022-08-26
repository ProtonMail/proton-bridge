Feature: IMAP get mailbox info
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name | type   |
      | f1   | folder |
      | l1   | label  |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

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