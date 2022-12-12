Feature: SMTP sending the same message twice
  Background:
    Given there exists an account with username "user" and password "password"
    And there exists an account with username "bridgetest" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    And the user logs in with username "bridgetest" and password "password"
    And user "user" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Hello

      World
      """
    And it succeeds

  Scenario: The exact same message is not sent twice
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Hello

      World
      """
    Then it succeeds
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to                  | subject | body  |
      | user@[domain] | bridgetest@[domain] | Hello   | World |
    When user "bridgetest" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from          | to                  | subject | body  |
      | user@[domain] | bridgetest@[domain] | Hello   | World |


  Scenario: Slight change means different message and is sent twice
    When SMTP client "1" sends the following message from "user@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>
      Subject: Hello.

      World
      """
    Then it succeeds
    When user "user" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from          | to                  | subject | body  |
      | user@[domain] | bridgetest@[domain] | Hello   | World |
      | user@[domain] | bridgetest@[domain] | Hello.  | World |
    When user "bridgetest" connects and authenticates IMAP client "2"
    Then IMAP client "2" eventually sees the following messages in "Inbox":
      | from          | to                  | subject | body  |
      | user@[domain] | bridgetest@[domain] | Hello   | World |
      | user@[domain] | bridgetest@[domain] | Hello.  | World |