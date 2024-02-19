Feature: A user can authenticate an SMTP client
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:user2]" and password "password2"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:user2]" and password "password2"
    Then it succeeds

  Scenario: SMTP client can authenticate successfully
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" can authenticate

  Scenario: User agent with only SMTP client connected
    Then the user agent is "NoClient/0.0.1 ([GOOS])"
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" can authenticate
    Then the user agent is "UnknownClient/0.0.1 ([GOOS])"

  Scenario: SMTP client cannot authenticate with wrong username
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" cannot authenticate with incorrect username

  Scenario: SMTP client cannot authenticate with wrong password
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" cannot authenticate with incorrect password

  Scenario: SMTP client cannot authenticate for disconnected user
    When user "[user:user]" logs out
    And user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" cannot authenticate

  Scenario: SMTP client can authenticate successfully with alias
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:alias]@[domain]"
    Then it succeeds

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: SMTP client can not authenticate with disabled address
    Given the account "[user:user2]" has additional disabled address "[alias:disabled]@[domain]"
    And it succeeds
    When user "[user:user2]" connects and authenticates SMTP client "1" with address "[alias:disabled]@[domain]"
    Then it fails

  Scenario: SMTP Logs out user
    Given user "[user:user]" connects SMTP client "1"
    When SMTP client "1" logs out
    Then it succeeds

  Scenario: SMTP client can authenticate two users
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" can authenticate
    When user "[user:user2]" connects SMTP client "2"
    Then SMTP client "2" can authenticate

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: SMTP Authenticates with secondary address of account with disabled primary address
    Given there exists a disabled account with username "[user:user3]" and password "password3"
    And the account "[user:user3]" has additional address "[alias:alias3]@[domain]"
    And it succeeds
    And the user logs in with username "[user:user3]" and password "password3"
    And it succeeds
    When user "[user:user3]" connects and authenticates SMTP client "1" with address "[alias:alias3]@[domain]"
    Then it succeeds
