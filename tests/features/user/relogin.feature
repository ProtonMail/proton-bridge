Feature: A logged out user can login again
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"

  Scenario: Login to disconnected account
    When user "user@pm.me" logs out
    And bridge restarts
    And the user logs in with username "user@pm.me" and password "password"
    Then user "user@pm.me" is listed and connected

  Scenario: Cannot login to removed account
    When user "user@pm.me" is deleted
    Then user "user@pm.me" is not listed