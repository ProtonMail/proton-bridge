Feature: Bridge send usage metrics
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And there exists an account with username "[user:user2]" and password "password"
    And bridge starts


  Scenario: Telemetry availability - No user
    Then bridge telemetry feature is enabled

  Scenario: Telemetry availability - Multi user
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user2]" and password "password"
    And user "[user:user2]" finishes syncing
    When user "[user:user2]" has telemetry set to 0
    Then bridge telemetry feature is disabled
