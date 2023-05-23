Feature: A user can authenticate an IMAP client
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:user2]" and password "password2"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    And the account "[user:user2]" has additional disabled address "[alias:alias2]@[domain]"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:user2]" and password "password2"
    Then it succeeds

  Scenario: IMAP client can authenticate successfully
    When user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" can authenticate

  Scenario: IMAP client can authenticate successfully with different case
    When user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" can authenticate with address "{toUpper:[user:user]@[domain]}"

  Scenario: IMAP client can authenticate successfully with secondary address
    Given user "[user:user]" connects and authenticates IMAP client "1" with address "[alias:alias]@[domain]"

  Scenario: IMAP client can not authenticate successfully with disable address
    Given user "[user:user2]" connects and can not authenticate IMAP client "1" with address "[alias:alias2]@[domain]"

  Scenario: IMAP client can authenticate successfully
    When user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" can authenticate

  Scenario: IMAP client cannot authenticate with bad username
    When user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate with incorrect username

  Scenario: IMAP client cannot authenticate with bad password
    When user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate with incorrect password

  Scenario: IMAP client cannot authenticate for disconnected user
    When user "[user:user]" logs out
    And user "[user:user]" connects IMAP client "1"
    Then IMAP client "1" cannot authenticate
