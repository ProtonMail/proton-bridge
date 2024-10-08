Feature: Bridge send remote notification observability metrics
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    Then it succeeds
    When bridge starts
    Then it succeeds

  Scenario: Test all possible gluon error observability metrics
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends all possible gluon error observability metrics
    Then it succeeds
