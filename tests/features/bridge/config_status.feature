Feature: Configuration Status Telemetry
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    Then it succeeds

  Scenario: Init config status on user addition
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"

  Scenario: Config Status Success on IMAP
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then config status succeed for user "[user:user]"

  Scenario: Config Status Success on SMTP
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"
    When user "[user:user]" connects and authenticates SMTP client "1"
    Then config status succeed for user "[user:user]"