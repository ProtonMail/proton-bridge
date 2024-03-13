Feature: SMTP client authentication with address modes
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    And it succeeds

  Scenario: SMTP client can authenticate successfully with secondary address in combine mode
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:alias]@[domain]"
    Then it succeeds

  Scenario: SMTP client can authenticate successfully with secondary address in split mode
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:alias]@[domain]"
    Then it succeeds

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: SMTP client can authenticate successfully with disabled alias in combine mode
    Given the account "[user:user]" has additional disabled address "[alias:disabled]@[domain]"
    And it succeeds
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:disabled]@[domain]"
    Then it fails

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: SMTP client can authenticate successfully with disabled alias in split mode
    Given the account "[user:user]" has additional disabled address "[alias:disabled]@[domain]"
    And it succeeds
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:disabled]@[domain]"
    Then it fails


