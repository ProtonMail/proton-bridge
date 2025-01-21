Feature: Bridge checks for updates
  Background:
    Given the legacy update kill switch is enabled

  Scenario: Update not available
    Given bridge is version "2.3.0" and the latest available version is "2.3.0" reachable from "2.3.0"
    When bridge starts
    And bridge verifies that the legacy update is enabled
    And bridge checks for updates
    Then bridge sends an update not available event

  Scenario: Update available without automatic updates enabled
    Given bridge is version "2.3.0" and the latest available version is "2.4.0" reachable from "2.3.0"
    And the user has disabled automatic updates
    When bridge starts
    And bridge verifies that the legacy update is enabled
    And bridge checks for updates
    Then bridge sends an update available event for version "2.4.0"

  Scenario: Update available with automatic updates enabled
    Given bridge is version "2.3.0" and the latest available version is "2.4.0" reachable from "2.3.0"
    When bridge starts
    And bridge verifies that the legacy update is enabled
    And bridge checks for updates
    Then bridge sends an update installed event for version "2.4.0"

  Scenario: Manual update available with automatic updates enabled
    Given bridge is version "2.3.0" and the latest available version is "2.4.0" reachable from "2.4.0"
    When bridge starts
    And bridge verifies that the legacy update is enabled
    And bridge checks for updates
    Then bridge sends a manual update event for version "2.4.0"

  Scenario: Update is required to continue using bridge
    Given there exists an account with username "[user:user]" and password "password"
    And bridge is version "2.3.0" and the latest available version is "2.3.0" reachable from "2.3.0"
    And the API requires bridge version at least "2.4.0"
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then bridge sends a forced update event
