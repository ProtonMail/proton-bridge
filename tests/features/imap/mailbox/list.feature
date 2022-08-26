Feature: IMAP list mailboxes
  Scenario: List mailboxes
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | mbox1 | folder |
      | mbox2 | label  |
    When bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info:
      | name          | total | unread |
      | INBOX         | 0     | 0      |
      | Drafts        | 0     | 0      |
      | Sent          | 0     | 0      |
      | Starred       | 0     | 0      |
      | Archive       | 0     | 0      |
      | Spam          | 0     | 0      |
      | Trash         | 0     | 0      |
      | All Mail      | 0     | 0      |
      | Folders       | 0     | 0      |
      | Folders/mbox1 | 0     | 0      |
      | Labels        | 0     | 0      |
      | Labels/mbox2  | 0     | 0      |

  Scenario: List multiple times in parallel without crash
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has 20 custom folders
    And the account "user@pm.me" has 60 custom labels
    When bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    When user "user@pm.me" connects and authenticates IMAP client "1"
    And  user "user@pm.me" connects and authenticates IMAP client "2"
    Then IMAP client "1" counts 20 mailboxes under "Folders"
    And  IMAP client "1" counts 60 mailboxes under "Labels"
    Then IMAP client "2" counts 20 mailboxes under "Folders"
    And  IMAP client "2" counts 60 mailboxes under "Labels"