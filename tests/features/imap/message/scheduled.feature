# need to implement _schedule message_ test step for black
@skip-black
Feature: IMAP interaction with scheduled

  Scenario: Not possible to delete from Scheduled and expunge does nothing
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has 10 messages in "Folders/mbox"
    And the address "[user:user]@[domain]" of account "[user:user]" has 1 messages in "Scheduled"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds
    When IMAP client "1" selects "Scheduled"
    And IMAP client "1" marks message 1 as deleted
    Then it succeeds
    And IMAP client "1" expunges
    Then it fails

  Scenario: Move message from Scheduled is not possible
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name   | type   |
      | mbox   | folder |
      | label  | label  |
      | label2 | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Labels/label2":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | baz     | false  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Sent":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | bax     | false  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Scheduled":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | sch     | false  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds
    Given test skips reporter checks
    When IMAP client "1" moves the message with subject "sch" from "Scheduled" to "Inbox"
    Then it fails
    And IMAP client "1" eventually sees the following messages in "Scheduled":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | sch     | false  |
