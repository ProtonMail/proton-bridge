Feature: IMAP list mailboxes
  Scenario: List mailboxes
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox1 | folder |
      | mbox2 | label  |
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info:
      | name          |
      | INBOX         |
      | Drafts        |
      | Sent          |
      | Starred       |
      | Archive       |
      | Spam          |
      | Trash         |
      | All Mail      |
      | Folders       |
      | Folders/mbox1 |
      | Labels        |
      | Labels/mbox2  |

  Scenario: List multiple times in parallel without crash
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has 20 custom folders
    And the account "[user:user]" has 60 custom labels
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1"
    And  user "[user:user]" connects and authenticates IMAP client "2"
    Then IMAP client "1" counts 20 mailboxes under "Folders"
    And  IMAP client "1" counts 60 mailboxes under "Labels"
    Then IMAP client "2" counts 20 mailboxes under "Folders"
    And  IMAP client "2" counts 60 mailboxes under "Labels"