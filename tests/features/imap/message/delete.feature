Feature: IMAP remove messages from mailbox
  Background:
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

  Scenario: Mark message as deleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 2 as deleted
    And it succeeds
    Then IMAP client "1" sees that message 2 has the flag "\Deleted"
    When IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" eventually sees 9 messages in "Folders/mbox"

  Scenario: Mark all messages as deleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks all messages as deleted
    And IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" eventually sees 0 messages in "Folders/mbox"

  Scenario: Mark messages as undeleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks all messages as deleted
    And it succeeds
    But IMAP client "1" marks message 2 as not deleted
    And it succeeds
    And IMAP client "1" marks message 3 as not deleted
    And it succeeds
    When IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" eventually sees 2 messages in "Folders/mbox"

  Scenario: Not possible to delete from All Mail and expunge does nothing
    When IMAP client "1" selects "All Mail"
    And IMAP client "1" marks message 2 as deleted
    And it succeeds
    And IMAP client "1" expunges
    Then it fails

  Scenario: Not possible to delete from Scheduled and expunge does nothing
    When IMAP client "1" selects "Scheduled"
    And IMAP client "1" marks message 1 as deleted
    Then it succeeds
    And IMAP client "1" expunges
    Then it fails
