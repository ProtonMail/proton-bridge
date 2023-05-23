Feature: A user can login
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And there exists an account with username "[user:MixedCaps]" and password "password"
    And there exists a disabled account with username "[user:disabled]" and password "password"
    Then it succeeds
    And bridge starts
    Then it succeeds

  Scenario: Login to account
    When the user logs in with username "[user:user]" and password "password"
    Then user "[user:user]" is eventually listed and connected

  Scenario: Login to account with wrong password
    When the user logs in with username "[user:user]" and password "wrong"
    Then user "[user:user]" is not listed

  Scenario: Login to nonexistent account
    When the user logs in with username "nonexistent" and password "unknown"
    Then user "nonexistent" is not listed

  Scenario: Login to account without internet
    Given the internet is turned off
    When the user logs in with username "[user:user]" and password "password"
    Then user "[user:user]" is not listed

  Scenario: Login to account with caps
    When the user logs in with username "[user:MixedCaps]" and password "password"
    Then user "[user:MixedCaps]" is eventually listed and connected

  Scenario: Login to account with disabled primary
    When the user logs in with username "[user:disabled]" and password "password"
    Then user "[user:disabled]" is eventually listed and connected

  Scenario: Login to account without internet but the connection is later restored
    When the user logs in with username "[user:user]" and password "password"
    And bridge stops
    And the internet is turned off
    And bridge starts
    And the internet is turned on
    Then user "[user:user]" is eventually listed and connected

  Scenario: Login to multiple accounts
    Given there exists an account with username "[user:additional]" and password "password"
    When the user logs in with username "[user:user]" and password "password"
    And the user logs in with username "[user:additional]" and password "password"
    Then user "[user:user]" is eventually listed and connected
    And user "[user:additional]" is eventually listed and connected