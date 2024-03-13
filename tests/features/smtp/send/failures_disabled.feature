Feature: SMTP wrong messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional disabled address "[user:disabled]@[domain]"
    And there exists an account with username "[user:to]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" connects and authenticates SMTP client "1"
    Then it succeeds

  # Need to find  way to setup disabled address on black
  @skip-black
  Scenario: Send from a valid address that cannot send
    Given the account "[user:user]" has additional disabled address "[user:disabled]@[domain]"
    When SMTP client "1" sends the following message from "[user:disabled]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test Disabled <[user:disabled]@[domain]>
      To: Internal Bridge <[user:to]@[domain]>

      Hello
      """
    And it fails with error "Error: cannot send from address: [user:disabled]@[domain]"

