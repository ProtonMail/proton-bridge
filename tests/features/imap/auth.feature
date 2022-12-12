Feature: A user can authenticate an IMAP client
  Background:
    Given there exists an account with username "user" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"

  Scenario: IMAP client can authenticate successfully
    When user "user" connects IMAP client "1"
    Then IMAP client "1" can authenticate

  Scenario: IMAP client can authenticate successfully with different case
    When user "user" connects IMAP client "1"
    Then IMAP client "1" can authenticate with address "USER@[domain]"

  Scenario: IMAP client can authenticate successfully
    When user "user" connects IMAP client "1"
    Then IMAP client "1" can authenticate

  Scenario: IMAP client cannot authenticate with bad username
    When user "user" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate with incorrect username

  Scenario: IMAP client cannot authenticate with bad password
    When user "user" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate with incorrect password

  Scenario: IMAP client cannot authenticate for disconnected user
    When user "user" logs out
    And user "user" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate
