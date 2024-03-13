Feature: Address key usage during SMTP send
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    And it succeeds

  Scenario: Non-active sender in combined mode using non-active key
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And it succeeds
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[user:user]@[domain]"
    And SMTP client "1" sends the following message from "[alias:alias]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[alias:alias]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>

      hello

      """
    Then it succeeds
    And the message used "[alias:alias]@[domain]" key for sending

  Scenario: Non-active sender in split mode using non-active key
    Given bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    And it succeeds
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[user:user]@[domain]"
    And SMTP client "1" sends the following message from "[alias:alias]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[alias:alias]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>

      hello

      """
    Then it succeeds
    And the message used "[alias:alias]@[domain]" key for sending

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: Disabled sender in combined mode fails to send
    Given the account "[user:user]" has additional disabled address "[user:disabled]@[domain]"
    And it succeeds
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And it succeeds
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[user:user]@[domain]"
    And SMTP client "1" sends the following message from "[alias:disabled]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[alias:disabled]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>

      hello

      """
    Then it fails

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: Disabled sender in split mode fails to send
    Given the account "[user:user]" has additional disabled address "[alias:disabled]@[domain]"
    And it succeeds
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    And it succeeds
    When user "[user:user]" connects and authenticates SMTP client "1" with address "[alias:alias]@[domain]"
    And SMTP client "1" sends the following message from "[alias:disabled]@[domain]" to "pm.bridge.qa@gmail.com":
      """
      From: Bridge Test <[alias:disabled]@[domain]>
      To: External Bridge <pm.bridge.qa@gmail.com>

      hello

      """
    Then it fails
