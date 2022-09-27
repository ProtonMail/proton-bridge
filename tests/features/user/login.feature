Feature: A user can login
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts

  Scenario: Login to account
    When the user logs in with username "user@pm.me" and password "password"
    Then user "user@pm.me" is listed and connected

  Scenario: Login to account with wrong password
    When the user logs in with username "user@pm.me" and password "wrong"
    Then user "user@pm.me" is not listed

  Scenario: Login to nonexistent account
    When the user logs in with username "other@pm.me" and password "unknown"
    Then user "other@pm.me" is not listed

  Scenario: Login to account without internet
    Given the internet is turned off
    When the user logs in with username "user@pm.me" and password "password"
    Then user "user@pm.me" is not listed

  Scenario: Login to account without internet but the connection is later restored
    When the user logs in with username "user@pm.me" and password "password"
    And bridge stops
    And the internet is turned off
    And bridge starts
    And the internet is turned on
    Then user "user@pm.me" is eventually listed and connected

  Scenario: Login to multiple accounts
    Given there exists an account with username "additional@pm.me" and password "other"
    When the user logs in with username "user@pm.me" and password "password"
    And the user logs in with username "additional@pm.me" and password "other"
    Then user "user@pm.me" is listed and connected
    And user "additional@pm.me" is listed and connected