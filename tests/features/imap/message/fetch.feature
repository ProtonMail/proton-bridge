Feature: IMAP Fetch
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from              | to                   | subject | date                          |
      | john.doe@mail.com | [user:user]@[domain] | foo     |  13 Jul 69 00:00 +0000        |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Fetch very old message
    Given IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | date                  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | 13 Aug 82 00:00 +0000 |
    Then IMAP client "1" sees header "X-Original-Date: Sun, 13 Jul 1969 00:00:00 +0000" in message with subject "foo" in "INBOX"


  Scenario: Fetch from deleted cache
    When the user deletes the gluon cache
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | date                  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | 13 Aug 82 00:00 +0000 |
