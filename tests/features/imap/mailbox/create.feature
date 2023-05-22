Feature: IMAP create mailbox
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | f1   | folder |
      | f2   | folder |
      | l1   | label  |
      | l2   | label  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Create folder
    When IMAP client "1" creates "Folders/mbox"
    Then IMAP client "1" sees "Folders/mbox"

  Scenario: Create label
    When IMAP client "1" creates "Labels/mbox"
    Then IMAP client "1" sees "Labels/mbox"

  Scenario: Creating folder or label with existing name is not possible
    When IMAP client "1" creates "Folders/f1"
    Then it fails
    And bridge reports a message with "Failed to create mailbox"
    When IMAP client "1" creates "Folders/f2"
    Then it fails
    And bridge reports a message with "Failed to create mailbox"
    When IMAP client "1" creates "Labels/l1"
    Then it fails
    And bridge reports a message with "Failed to create mailbox"
    When IMAP client "1" creates "Labels/l2"
    Then it fails
    And bridge reports a message with "Failed to create mailbox"
    When IMAP client "1" creates "Folders/f3"
    Then it succeeds
    When IMAP client "1" creates "Labels/l3"
    Then it succeeds
    Then IMAP client "1" eventually sees the following mailbox info:
      | name       |
      | INBOX      |
      | Drafts     |
      | Sent       |
      | Starred    |
      | Archive    |
      | Spam       |
      | Trash      |
      | All Mail   |
      | Folders    |
      | Folders/f1 |
      | Folders/f2 |
      | Folders/f3 |
      | Labels     |
      | Labels/l1  |
      | Labels/l2  |
      | Labels/l3  |

  Scenario: Creating subfolders is possible and they persist after resync
    When IMAP client "1" creates "Folders/f1/f11"
    Then it succeeds
    When IMAP client "1" creates "Folders/f1/f12"
    Then it succeeds
    When IMAP client "1" creates "Folders/f2/f21"
    Then it succeeds
    When IMAP client "1" creates "Folders/f2/f22"
    Then it succeeds
    Then IMAP client "1" eventually sees the following mailbox info:
      | name           |
      | INBOX          |
      | Drafts         |
      | Sent           |
      | Starred        |
      | Archive        |
      | Spam           |
      | Trash          |
      | All Mail       |
      | Folders        |
      | Folders/f1     |
      | Folders/f1/f11 |
      | Folders/f1/f12 |
      | Folders/f2     |
      | Folders/f2/f21 |
      | Folders/f2/f22 |
      | Labels         |
      | Labels/l1      |
      | Labels/l2      |
    When user "[user:user]" is deleted
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following mailbox info:
      | name           |
      | INBOX          |
      | Drafts         |
      | Sent           |
      | Starred        |
      | Archive        |
      | Spam           |
      | Trash          |
      | All Mail       |
      | Folders        |
      | Folders/f1     |
      | Folders/f1/f11 |
      | Folders/f1/f12 |
      | Folders/f2     |
      | Folders/f2/f21 |
      | Folders/f2/f22 |
      | Labels         |
      | Labels/l1      |
      | Labels/l2      |

  Scenario: Changing folder parent is possible and it persists after resync
    When IMAP client "1" creates "Folders/f1/f11"
    Then it succeeds
    When IMAP client "1" creates "Folders/f1/f12"
    Then it succeeds
    When IMAP client "1" creates "Folders/f2/f21"
    Then it succeeds
    When IMAP client "1" creates "Folders/f2/f22"
    Then it succeeds
    Then IMAP client "1" eventually sees the following mailbox info:
      | name           |
      | INBOX          |
      | Drafts         |
      | Sent           |
      | Starred        |
      | Archive        |
      | Spam           |
      | Trash          |
      | All Mail       |
      | Folders        |
      | Folders/f1     |
      | Folders/f1/f11 |
      | Folders/f1/f12 |
      | Folders/f2     |
      | Folders/f2/f21 |
      | Folders/f2/f22 |
      | Labels         |
      | Labels/l1      |
      | Labels/l2      |
    When IMAP client "1" renames "Folders/f1/f11" to "Folders/f2/f11"
    Then it succeeds
    When IMAP client "1" renames "Folders/f1/f12" to "Folders/f2/f12"
    Then it succeeds
    Then IMAP client "1" eventually sees the following mailbox info:
      | name           |
      | INBOX          |
      | Drafts         |
      | Sent           |
      | Starred        |
      | Archive        |
      | Spam           |
      | Trash          |
      | All Mail       |
      | Folders        |
      | Folders/f1     |
      | Folders/f2     |
      | Folders/f2/f11 |
      | Folders/f2/f12 |
      | Folders/f2/f21 |
      | Folders/f2/f22 |
      | Labels         |
      | Labels/l1      |
      | Labels/l2      |
    When user "[user:user]" is deleted
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following mailbox info:
      | name           |
      | INBOX          |
      | Drafts         |
      | Sent           |
      | Starred        |
      | Archive        |
      | Spam           |
      | Trash          |
      | All Mail       |
      | Folders        |
      | Folders/f1     |
      | Folders/f2     |
      | Folders/f2/f11 |
      | Folders/f2/f12 |
      | Folders/f2/f21 |
      | Folders/f2/f22 |
      | Labels         |
      | Labels/l1      |
      | Labels/l2      |

  Scenario: Create 2 levels deep Folder
    When IMAP client "1" creates "Folders/first/second"
    And it succeeds
    Then IMAP client "1" sees "Folders/first/second"

  Scenario: Creating mailbox without prefix is not possible
    Given test skips reporter checks
    When IMAP client "1" creates "mbox"
    Then it fails
    When IMAP client "1" creates "All Mail"
    Then it fails