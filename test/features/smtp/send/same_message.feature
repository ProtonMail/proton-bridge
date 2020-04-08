Feature: SMTP sending the same message twice
  Background:
    Given there is connected user "user"
    And there is SMTP client logged in as "user"
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello

      World

      """
    Then SMTP response is "OK"

  Scenario: The exact same message is not sent twice
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello

      World

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | time | from          | to                        | subject |
      | now  | [userAddress] | bridgetest@protonmail.com | Hello   |

  Scenario: Slight change means different message and is sent twice
    When SMTP client sends message
      """
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello.

      World

      """
    Then SMTP response is "OK"
    And mailbox "Sent" for "user" has messages
      | time | from          | to                        | subject |
      | now  | [userAddress] | bridgetest@protonmail.com | Hello   |
      | now  | [userAddress] | bridgetest@protonmail.com | Hello.  |
