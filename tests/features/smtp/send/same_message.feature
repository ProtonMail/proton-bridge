Feature: SMTP sending the same message twice
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And there exists an account with username "bridgetest@protonmail.com" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And the user logs in with username "bridgetest@protonmail.com" and password "password"
    And user "user@pm.me" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello

      World
      """
    And it succeeds

  Scenario: The exact same message is not sent twice
    When SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello

      World
      """
    Then it succeeds
    When user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from       | to                        | subject | body   |
      | user@pm.me | bridgetest@protonmail.com | Hello   | World  |
    When user "bridgetest@protonmail.com" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from       | to                        | subject | body  |
      | user@pm.me | bridgetest@protonmail.com | Hello   | World |


  Scenario: Slight change means different message and is sent twice
    When SMTP client "1" sends the following message from "user@pm.me" to "bridgetest@protonmail.com":
      """
      From: Bridge Test <user@pm.me>
      To: Internal Bridge <bridgetest@protonmail.com>
      Subject: Hello.

      World
      """
    Then it succeeds
    When user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from       | to                        | subject | body   |
      | user@pm.me | bridgetest@protonmail.com | Hello   | World  |
      | user@pm.me | bridgetest@protonmail.com | Hello.  | World  |
    When user "bridgetest@protonmail.com" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from       | to                        | subject | body  |
      | user@pm.me | bridgetest@protonmail.com | Hello   | World |
      | user@pm.me | bridgetest@protonmail.com | Hello.  | World  |