Feature: IMAP list mailboxes
  Scenario: List mailboxes
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox1 | folder |
      | mbox2 | label  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds
    And IMAP client "1" eventually sees the following mailbox info:
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

  Scenario: List with scheduled mail
    Given there exists an account with username "[user:user]" and password "password"
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Scheduled":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | sch     | false  |
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following mailbox info:
      | name          | total |
      | INBOX         | 0     |
      | Drafts        | 0     |
      | Sent          | 0     |
      | Starred       | 0     |
      | Archive       | 0     |
      | Spam          | 0     |
      | Trash         | 0     |
      | All Mail      | 1     |
      | Folders       | 0     |
      | Labels        | 0     |
      | Scheduled     | 1     |
