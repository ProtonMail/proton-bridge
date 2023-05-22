Feature: A user can connect an SMTP client to custom ports
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then it succeeds

  Scenario: Authenticates successfully on custom port
    When the user changes the SMTP port to 1144
    When user "[user:user]" connects SMTP client "1" on port 1144
    Then SMTP client "1" can authenticate