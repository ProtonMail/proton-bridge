Feature: A user can authenticate an SMTP client
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"

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