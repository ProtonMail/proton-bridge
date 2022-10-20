Feature: IMAP Hide All Mail
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Hide All Mail Mailbox
    Given IMAP client "1" sees the following mailbox info:
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
      | Labels        | 0     | 0      |
    When bridge hides all mail
    Then IMAP client "1" sees the following mailbox info:
      | name          | total | unread |
      | INBOX         | 0     | 0      |
      | Drafts        | 0     | 0      |
      | Sent          | 0     | 0      |
      | Starred       | 0     | 0      |
      | Archive       | 0     | 0      |
      | Spam          | 0     | 0      |
      | Trash         | 0     | 0      |
      | Folders       | 0     | 0      |
      | Labels        | 0     | 0      |
    When bridge shows all mail
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
      | Labels        | 0     | 0      |
