Feature: IMAP Hide All Mail
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Hide All Mail Mailbox
    Given IMAP client "1" eventually sees the following mailbox info:
      | name     |
      | INBOX    |
      | Drafts   |
      | Sent     |
      | Starred  |
      | Archive  |
      | Spam     |
      | Trash    |
      | All Mail |
      | Folders  |
      | Labels   |
    When the user hides All Mail
    Then IMAP client "1" eventually sees the following mailbox info:
      | name    |
      | INBOX   |
      | Drafts  |
      | Sent    |
      | Starred |
      | Archive |
      | Spam    |
      | Trash   |
      | Folders |
      | Labels  |
    When the user shows All Mail
    Then IMAP client "1" eventually sees the following mailbox info:
      | name     |
      | INBOX    |
      | Drafts   |
      | Sent     |
      | Starred  |
      | Archive  |
      | Spam     |
      | Trash    |
      | All Mail |
      | Folders  |
      | Labels   |
