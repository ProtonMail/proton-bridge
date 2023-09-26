Feature: IMAP remove messages from Trash
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |
    Then it succeeds

  Scenario Outline: Message in Trash and some other label is not permanently deleted
    Given the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Trash":
      | from              | to                   | subject | body  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | hello |
      | jane.doe@mail.com | name@[domain]        | bar     | world |
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "Trash"
    When IMAP client "1" copies the message with subject "foo" from "Trash" to "Labels/label"
    Then it succeeds
    When IMAP client "1" marks the message with subject "foo" as deleted
    Then it succeeds
    And IMAP client "1" eventually sees 2 messages in "Trash"
    And IMAP client "1" eventually sees 2 messages in "All Mail"
    And IMAP client "1" eventually sees 1 messages in "Labels/label"
    When IMAP client "1" expunges
    Then it succeeds
    And IMAP client "1" eventually sees 1 messages in "Trash"
    And IMAP client "1" eventually sees 2 messages in "All Mail"
    And IMAP client "1" eventually sees 1 messages in "Labels/label"

  Scenario Outline: Message in Trash only is permanently deleted
    Given the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Trash":
      | from              | to                   | subject | body  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | hello |
      | jane.doe@mail.com | name@[domain]        | bar     | world |
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "Trash"
    When IMAP client "1" marks the message with subject "foo" as deleted
    Then it succeeds
    And IMAP client "1" eventually sees 2 messages in "Trash"
    And IMAP client "1" eventually sees 2 messages in "All Mail"
    When IMAP client "1" expunges
    Then it succeeds
    And IMAP client "1" eventually sees 1 messages in "Trash"
    And IMAP client "1" eventually sees 1 messages in "All Mail"