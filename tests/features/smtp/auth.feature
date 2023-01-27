Feature: A user can authenticate an SMTP client
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user2:user2]" and password "password2"
    And there exists a disabled account with username "[user3:user3]" and password "password3"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    And the account "[user2:user2]" has additional disabled address "[alias2:alias2]@[domain]"
    And the account "[user3:user3]" has additional address "[alias3:alias3]@[domain]"
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user2:user2]" and password "password2"
    And the user logs in with username "[user3:user3]" and password "password3"

  Scenario: SMTP client can authenticate successfully
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" can authenticate

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

  Scenario: SMTP client can not authenticate with disabled address
    When user "[user2:user2]" connects and authenticates SMTP client "1" with address "[alias2:alias2]@[domain]"
    Then it fails

  Scenario: SMTP Logs out user
    Given user "[user:user]" connects SMTP client "1"
    When SMTP client "1" logs out
    Then it succeeds

  Scenario: SMTP client can authenticate two users
    When user "[user:user]" connects SMTP client "1"
    Then SMTP client "1" can authenticate
    When user "[user2:user2]" connects SMTP client "2"
    Then SMTP client "2" can authenticate

  @ignore-live
  Scenario: SMTP Authenticates with secondary address of account with disabled primary address
    When user "[user3:user3]" connects and authenticates SMTP client "1" with address "[alias3:alias3]@[domain]"
    Then it succeeds
