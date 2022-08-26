Feature: IMAP create mailbox
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Create folder
    When IMAP client "1" creates "Folders/mbox"
    Then IMAP client "1" sees "Folders/mbox"

  Scenario: Create label
    When IMAP client "1" creates "Labels/mbox"
    Then IMAP client "1" sees "Labels/mbox"