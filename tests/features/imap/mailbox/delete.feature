Feature: IMAP delete mailbox
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
      | two   | folder |
      | three | label  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Delete folder
    When IMAP client "1" deletes "Folders/one"
    Then IMAP client "1" does not see "Folders/one"
    But IMAP client "1" sees "Folders/two"
    But IMAP client "1" sees "Labels/three"

  Scenario: Delete label
    When IMAP client "1" deletes "Labels/three"
    Then IMAP client "1" does not see "Labels/three"
    But IMAP client "1" sees "Folders/one"
    But IMAP client "1" sees "Folders/two"

  Scenario: Deleting system mailbox is not possible
    When IMAP client "1" deletes "INBOX"
    Then it fails
    And IMAP client "1" sees "INBOX"
