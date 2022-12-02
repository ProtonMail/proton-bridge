Feature: IMAP remove messages from mailbox
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |
    And the address "user@pm.me" of account "user@pm.me" has 10 messages in "mbox"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Mark message as deleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks message 2 as deleted
    Then IMAP client "1" sees that message 2 has the flag "\Deleted"
    When IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" sees 9 messages in "Folders/mbox"

  Scenario: Mark all messages as deleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks all messages as deleted
    And IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" sees 0 messages in "Folders/mbox"

  Scenario: Mark messages as undeleted and EXPUNGE
    When IMAP client "1" selects "Folders/mbox"
    And IMAP client "1" marks all messages as deleted
    But IMAP client "1" marks message 2 as not deleted
    And IMAP client "1" marks message 3 as not deleted
    When IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" sees 2 messages in "Folders/mbox"

 Scenario: Not possible to delete from All Mail and expunge does nothing
   When IMAP client "1" selects "All Mail"
   And IMAP client "1" marks message 2 as deleted
   And IMAP client "1" expunges
   Then it fails