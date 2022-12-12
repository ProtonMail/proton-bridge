Feature: SMTP sending two messages
  Background:
    Given there exists an account with username "user" and password "password"
    And there exists an account with username "user-multi" and password "password"
    And the account "user-multi" has additional address "user-multi-alias@[domain]"
    And there exists an account with username "bridgetest" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And the user logs in with username "user-multi" and password "password"
    And the user sets the address mode of user "user-multi" to "split"

  Scenario: Send two messages in one connection
    When user "user" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      hello

      """
    Then it succeeds
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      world

      """
    Then it succeeds

  Scenario: Send with two addresses of the same user in split mode
    When user "user-multi" connects and authenticates SMTP client "1" with address "user-multi@[domain]"
    And user "user-multi" connects and authenticates SMTP client "2" with address "user-multi-alias@[domain]"
    And SMTP client "1" sends the following message from "user-multi@[domain]" to "bridgetest@[domain]>":
      """
      From: Bridge Test <user-multi@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      hello

      """
    Then it succeeds
    When SMTP client "2" sends the following message from "user-multi@[domain]" to "bridgetest@[domain]>":
      """
      From: Bridge Test <user-multi@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      world

      """
    Then it succeeds

  Scenario: Send with two separate users
    When user "user" connects and authenticates SMTP client "1"
    And user "user-multi" connects and authenticates SMTP client "2"
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]>":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      hello

      """
    Then it succeeds
    When SMTP client "2" sends the following message from "user-multi@[domain]" to "bridgetest@[domain]>":
      """
      From: Bridge Test <user-multi@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      world

      """
    Then it succeeds
