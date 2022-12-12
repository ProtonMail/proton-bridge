Feature: A user can be deleted
  Background:
    Given there exists an account with username "user" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"

  Scenario: Delete a connected user
    When user "user" is deleted
    Then user "user" is not listed

  Scenario: Delete a disconnected user
    Given user "user" logs out
    When user "user" is deleted
    Then user "user" is not listed