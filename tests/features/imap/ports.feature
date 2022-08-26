Feature: A user can connect an IMAP client to custom ports
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And the user changes the IMAP port to 1144

  Scenario: Authenticates successfully on custom port
    When user "user@pm.me" connects IMAP client "1" on port 1144
    Then IMAP client "1" can authenticate