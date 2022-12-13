Feature: Bridge can fully sync an account
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has 20 custom folders
    And the account "[user:user]" has 60 custom labels
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" counts 20 mailboxes under "Folders"
    And  IMAP client "1" counts 60 mailboxes under "Labels"

  Scenario: The user changes the gluon path
    When the user changes the gluon path
    And user "[user:user]" connects and authenticates IMAP client "2"
    Then IMAP client "2" counts 20 mailboxes under "Folders"
    And  IMAP client "2" counts 60 mailboxes under "Labels"