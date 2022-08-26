Feature: A logged in user is logged out when its auth is revoked.
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"

  Scenario: The auth is revoked while bridge is running
    When the auth of user "user@pm.me" is revoked
    Then bridge sends a deauth event for user "user@pm.me"
    And user "user@pm.me" is listed but not connected

  Scenario: The auth is revoked while bridge is not running
    Given bridge stops
    And the auth of user "user@pm.me" is revoked
    When bridge starts
    Then user "user@pm.me" is listed but not connected