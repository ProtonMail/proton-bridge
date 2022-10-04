Feature: Address mode
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has additional address "alias@pm.me"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
      | two   | folder |
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "one":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And the address "alias@pm.me" of account "user@pm.me" has the following messages in "two":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing

  Scenario: The user is in combined mode
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    Then IMAP client "1" sees the following messages in "Folders/one":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And IMAP client "1" sees the following messages in "Folders/two":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    And IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    Then IMAP client "2" sees the following messages in "Folders/one":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And IMAP client "2" sees the following messages in "Folders/two":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    And IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |

  Scenario: The user is in split mode
    Given the user sets the address mode of "user@pm.me" to "split"
    And user "user@pm.me" finishes syncing
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    Then IMAP client "1" sees the following messages in "Folders/one":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And IMAP client "1" sees 0 messages in "Folders/two"
    And IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    Then IMAP client "2" sees 0 messages in "Folders/one"
    And IMAP client "2" sees the following messages in "Folders/two":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    And IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |

  Scenario: The user switches from combined to split mode and back
    Given the user sets the address mode of "user@pm.me" to "split"
    And user "user@pm.me" finishes syncing
    And the user sets the address mode of "user@pm.me" to "combined"
    And user "user@pm.me" finishes syncing
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    Then IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    Then IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |

  Scenario: The user adds an address while in combined mode
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    Then IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    Then IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    Given the account "user@pm.me" has additional address "other@pm.me"
    And bridge sends an address created event for user "user@pm.me"
    When user "user@pm.me" connects and authenticates IMAP client "3" with address "other@pm.me"
    Then IMAP client "3" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |

  Scenario: The user adds an address while in split mode
    Given the user sets the address mode of "user@pm.me" to "split"
    And user "user@pm.me" finishes syncing
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    And IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    And IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    Given the account "user@pm.me" has additional address "other@pm.me"
    And bridge sends an address created event for user "user@pm.me"
    When user "user@pm.me" connects and authenticates IMAP client "3" with address "other@pm.me"
    Then IMAP client "3" eventually sees 0 messages in "All Mail"

  Scenario: The user deletes an address while in combined mode
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    Then IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    Then IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    Given the account "user@pm.me" no longer has additional address "alias@pm.me"
    And bridge sends an address deleted event for user "user@pm.me"
    When user "user@pm.me" connects IMAP client "3"
    Then IMAP client "3" cannot authenticate with address "alias@pm.me"

  Scenario: The user deletes an address while in split mode
    Given the user sets the address mode of "user@pm.me" to "split"
    And user "user@pm.me" finishes syncing
    When user "user@pm.me" connects and authenticates IMAP client "1" with address "user@pm.me"
    And IMAP client "1" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    When user "user@pm.me" connects and authenticates IMAP client "2" with address "alias@pm.me"
    And IMAP client "2" sees the following messages in "All Mail":
      | from    | to        | subject | unread |
      | c@pm.me | c@pm.me   | three   | true  |
      | d@pm.me | d@pm.me   | four    | false |
    Given the account "user@pm.me" no longer has additional address "alias@pm.me"
    And bridge sends an address deleted event for user "user@pm.me"
    When user "user@pm.me" connects IMAP client "3"
    Then IMAP client "3" cannot authenticate with address "alias@pm.me"

  Scenario: The user makes an alias the primary address while in combined mode

  Scenario: The user makes an alias the primary address while in split mode