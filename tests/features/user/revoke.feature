Feature: A logged in user is logged out when its auth is revoked.
  Background:
    Given there exists an account with username "user" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"

  Scenario: The auth is revoked while bridge is running
    When the auth of user "user" is revoked
    Then bridge sends a deauth event for user "user"
    And user "user" is listed but not connected

  Scenario: The auth is revoked while bridge is not running
    Given bridge stops
    And the auth of user "user" is revoked
    When bridge starts
    Then user "user" is listed but not connected