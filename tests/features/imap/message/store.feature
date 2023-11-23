Feature: IMAP marks messages as forwarded
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has 1 messages in "Folders/mbox"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Mark message as forwarded
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 1 as "forwarded"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "forwarded"
    And it succeeds

  @ignore-live
  Scenario: Mark message as forwarded and then revert
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 1 as "forwarded"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "forwarded"
    And it succeeds
    And IMAP client "1" marks message 1 as "unforwarded"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 does not have the flag "forwarded"
    And it succeeds

  Scenario: Mark message as replied
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 1 as "replied"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "\Answered"
    And it succeeds

  @regression
  Scenario: Mark message as replied and then revert
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 1 as "replied"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 has the flag "\Answered"
    And it succeeds
    And IMAP client "1" marks message 1 as "unreplied"
    And it succeeds
    Then IMAP client "1" eventually sees that message at row 1 does not have the flag "\Answered"
    And it succeeds