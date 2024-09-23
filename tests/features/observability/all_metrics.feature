Feature: Bridge send remote notification observability metrics
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    Then it succeeds
    When bridge starts
    Then it succeeds

  Scenario: Test all possible heartbeat metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible observability heartbeat metrics
    Then it succeeds

  Scenario: Test all possible user discrimination metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible user distinction metrics
    Then it succeeds

  Scenario: Test all possible sync message event failure observability metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible sync message event failure observability metrics
    Then it succeeds

  Scenario: Test all possible event loop message events observability metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible event loop message events observability metrics
    Then it succeeds

  Scenario: Test all possible sync message building failure observability metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible sync message building failure observability metrics
    Then it succeeds

  Scenario: Test all possible sync message building success observability metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible sync message building success observability metrics
    Then it succeeds


