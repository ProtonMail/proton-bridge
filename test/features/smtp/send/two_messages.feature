Feature: SMTP sending two messages
  Scenario: Send two messages in one connection
    Given there is connected user "user"
    And there is SMTP client logged in as "user"
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response is "OK"
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then SMTP response is "OK"

  Scenario: Send to two addresses
    Given there is connected user "userMoreAddresses"
    And there is "userMoreAddresses" in "split" address mode
    And there is SMTP client "smtp1" logged in as "userMoreAddresses" with address "primary"
    And there is SMTP client "smtp2" logged in as "userMoreAddresses" with address "secondary"
    When SMTP client "smtp1" sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response to "smtp1" is "OK"
    When SMTP client "smtp2" sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then SMTP response to "smtp2" is "OK"

  Scenario: Send to two users
    Given there is connected user "user"
    And there is connected user "userMoreAddresses"
    And there is SMTP client "smtp1" logged in as "user"
    And there is SMTP client "smtp2" logged in as "userMoreAddresses"

    When SMTP client "smtp1" sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      hello

      """
    Then SMTP response to "smtp1" is "OK"
    When SMTP client "smtp2" sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>

      world

      """
    Then SMTP response to "smtp2" is "OK"
