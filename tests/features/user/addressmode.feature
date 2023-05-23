Feature: Address mode
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    And the account "[user:user]" has the following custom mailboxes:
      | name | type   |
      | one  | folder |
      | two  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And the address "[alias:alias]@[domain]" of account "[user:user]" has the following messages in "Folders/two":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    Then it succeeds

  Scenario: The user is in combined mode
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    Then IMAP client "1" eventually sees the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And IMAP client "1" eventually sees the following messages in "Folders/two":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    Then IMAP client "2" eventually sees the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And IMAP client "2" eventually sees the following messages in "Folders/two":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    And IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |

  Scenario: The user is in split mode
    Given the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    Then IMAP client "1" eventually sees the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And IMAP client "1" eventually sees 0 messages in "Folders/two"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    Then IMAP client "2" eventually sees 0 messages in "Folders/one"
    And IMAP client "2" eventually sees the following messages in "Folders/two":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    And IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |

  Scenario: The user switches from combined to split mode and back
    Given the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    And the user sets the address mode of user "[user:user]" to "combined"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    Then IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    Then IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |

  Scenario: The user adds an address while in combined mode
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    Then IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    Then IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    Given the account "[user:user]" has additional address "other@[domain]"
    And bridge sends an address created event for user "[user:user]"
    When user "[user:user]" connects and authenticates IMAP client "3" with address "other@[domain]"
    Then IMAP client "3" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |

  Scenario: The user adds an address while in split mode
    Given the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    And IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    Given the account "[user:user]" has additional address "other@[domain]"
    And bridge sends an address created event for user "[user:user]"
    When user "[user:user]" connects and authenticates IMAP client "3" with address "other@[domain]"
    Then IMAP client "3" eventually sees 0 messages in "All Mail"

  Scenario: The user deletes an address while in combined mode
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    Then IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    Then IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    Given the account "[user:user]" no longer has additional address "[alias:alias]@[domain]"
    And bridge sends an address deleted event for user "[user:user]"
    When user "[user:user]" connects IMAP client "3"
    Then IMAP client "3" cannot authenticate with address "[alias:alias]@[domain]"

  Scenario: The user deletes an address while in split mode
    Given the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    When user "[user:user]" connects and authenticates IMAP client "2" with address "[alias:alias]@[domain]"
    And IMAP client "2" eventually sees the following messages in "All Mail":
      | from       | to         | subject | unread |
      | c@[domain] | c@[domain] | three   | true   |
      | d@[domain] | d@[domain] | four    | false  |
    Given the account "[user:user]" no longer has additional address "[alias:alias]@[domain]"
    And bridge sends an address deleted event for user "[user:user]"
    When user "[user:user]" connects IMAP client "3"
    Then IMAP client "3" cannot authenticate with address "[alias:alias]@[domain]"

  Scenario: The user makes an alias the primary address while in combined mode

  Scenario: The user makes an alias the primary address while in split mode