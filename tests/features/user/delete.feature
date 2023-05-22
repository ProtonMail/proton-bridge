Feature: A user can be deleted
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then it succeeds

  Scenario: Delete a connected user
    When user "[user:user]" is deleted
    Then user "[user:user]" is not listed

  Scenario: Delete a disconnected user
    Given user "[user:user]" logs out
    When user "[user:user]" is deleted
    Then user "[user:user]" is not listed