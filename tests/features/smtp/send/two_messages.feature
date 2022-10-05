Feature: SMTP sending two messages
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And there exists an account with username "user-multi@pm.me" and password "password"
    And the account "user-multi@pm.me" has additional address "user-multi-alias@pm.me"
    And there exists an account with username "bridgetest@protonmail.com" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And the user logs in with username "user-multi@pm.me" and password "password"
    And the user sets the address mode of "user-multi@pm.me" to "split"

  Scenario: Send two messages in one connection
    When user "user@pm.me" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then it succeeds
    When SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then it succeeds

  Scenario: Send with two addresses of the same user in split mode
    When user "user-multi@pm.me" connects and authenticates SMTP client "1" with address "user-multi@pm.me"
    And user "user-multi@pm.me" connects and authenticates SMTP client "2" with address "user-multi-alias@pm.me"
    And SMTP client "1" sends the following message from "user-multi@pm.me" to "bridgetest@protonmail.com>":
      """
      From: Bridge Test <user-multi@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then it succeeds
    When SMTP client "2" sends the following message from "user-multi@pm.me" to "bridgetest@protonmail.com>":
      """
      From: Bridge Test <user-multi@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then it succeeds

  Scenario: Send with two separate users
    When user "user@pm.me" connects and authenticates SMTP client "1"
    And user "user-multi@pm.me" connects and authenticates SMTP client "2"
    When SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com>":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then it succeeds
    When SMTP client "2" sends the following message from "user-multi@pm.me" to "bridgetest@protonmail.com>":
      """
      From: Bridge Test <user-multi@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then it succeeds
