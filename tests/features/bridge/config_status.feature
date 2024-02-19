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
    And config status event "bridge_config_success" is eventually send 1 time


  Scenario: Config Status Success on SMTP
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"
    When user "[user:user]" connects and authenticates SMTP client "1"
    Then config status succeed for user "[user:user]"
    And config status event "bridge_config_success" is eventually send 1 time


  Scenario: Config Status Success send only once
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then config status succeed for user "[user:user]"
    And config status event "bridge_config_success" is eventually send 1 time
    When user "[user:user]" connects and authenticates IMAP client "2"
    Then config status event "bridge_config_success" is not send more than 1 time


  Scenario: Config Status Abort
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    Then config status file exist for user "[user:user]"
    And config status is pending for user "[user:user]"
    When user "[user:user]" is deleted
    Then config status event "bridge_config_abort" is eventually send 1 time


  Scenario: Config Status Recovery from deauth
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then config status succeed for user "[user:user]"
    When the auth of user "[user:user]" is revoked
    Then bridge sends a deauth event for user "[user:user]"
    Then config status is pending with failure for user "[user:user]"
    When the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then config status succeed for user "[user:user]"
    And config status event "bridge_config_recovery" is eventually send 1 time


  Scenario: Config Status Progress
    Then bridge telemetry feature is enabled
    When the user logs in with username "[user:user]" and password "password"
    And config status is pending for user "[user:user]"
    And bridge stops
    And force config status progress to be sent for user"[user:user]"
    And bridge starts
    Then config status event "bridge_config_progress" is eventually send 1 time
