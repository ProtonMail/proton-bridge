Feature: A logged out user can login again
  Background:
    Given there exists an account with username "user" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"

  Scenario: Login to disconnected account
    When user "user" logs out
    And bridge restarts
    And the user logs in with username "user" and password "password"
    Then user "user" is listed and connected

  Scenario: Cannot login to removed account
    When user "user" is deleted
    Then user "user" is not listed