Feature: IMAP Hide All Mail
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Hide All Mail Mailbox
    Given IMAP client "1" sees the following mailbox info:
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
      | Labels        |
    When bridge hides all mail
    Then IMAP client "1" sees the following mailbox info:
      | name          |
      | INBOX         |
      | Drafts        |
      | Sent          |
      | Starred       |
      | Archive       |
      | Spam          |
      | Trash         |
      | Folders       |
      | Labels        |
    When bridge shows all mail
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
      | Labels        |
