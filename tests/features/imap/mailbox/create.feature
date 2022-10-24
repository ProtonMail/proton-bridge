Feature: IMAP create mailbox
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name | type   |
      | f1   | folder |
      | f2   | folder |
      | l1   | label  |
      | l2   | label  |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Create folder
    When IMAP client "1" creates "Folders/mbox"
    Then IMAP client "1" sees "Folders/mbox"

  Scenario: Create label
    When IMAP client "1" creates "Labels/mbox"
    Then IMAP client "1" sees "Labels/mbox"

  Scenario: Creating folder or label with existing name is not possible
    When IMAP client "1" creates "Folders/f1"
    Then it fails
    When IMAP client "1" creates "Folders/f2"
    Then it fails
    When IMAP client "1" creates "Labels/l1"
    Then it fails
    When IMAP client "1" creates "Labels/l2"
    Then it fails
    When IMAP client "1" creates "Folders/f3"
    Then it succeeds
    When IMAP client "1" creates "Labels/l3"
    Then it succeeds
    Then IMAP client "1" sees the following mailbox info:
      | name         |
      | INBOX        |
      | Drafts       |
      | Sent         |
      | Starred      |
      | Archive      |
      | Spam         |
      | Trash        |
      | All Mail     |
      | Folders      |
      | Folders/f1   |
      | Folders/f2   |
      | Folders/f3   |
      | Labels       |
      | Labels/l1    |
      | Labels/l2    |
      | Labels/l3    |