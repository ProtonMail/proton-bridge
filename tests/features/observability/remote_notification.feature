Feature: Bridge send remote notification observability metrics
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    Then it succeeds
    When bridge starts
    Then it succeeds


  Scenario: Send notification 'received' and 'processed' observability metric
    When the user logs in with username "[user:user1]" and password "password"
    And the user with username "[user:user1]" sends the following remote notification observability metric "received"
    Then it succeeds
    And the user with username "[user:user1]" sends the following remote notification observability metric "processed"
    Then it succeeds


