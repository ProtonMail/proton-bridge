Feature: SMTP initiation
  Background:
    Given there exists an account with username "user" and password "password"
    And bridge starts
    And the user logs in with username "user" and password "password"
    When user "user" connects and authenticates SMTP client "1"

  Scenario: Send without first announcing FROM and TO
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Reset is the same as without FROM and TO
    When SMTP client "1" sends MAIL FROM "<user@[domain]>"
    Then it succeeds
    When SMTP client "1" sends RCPT TO "<user@[domain]>"
    Then it succeeds
    When SMTP client "1" sends RSET
    Then it succeeds
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Send without FROM
    When SMTP client "1" sends RCPT TO "<user@[domain]>"
    Then it fails with error "Missing MAIL FROM command"

  Scenario: Send without TO
    When SMTP client "1" sends MAIL FROM "<user@[domain]>"
    Then it succeeds
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Send with empty FROM
    When SMTP client "1" sends the following message from "<>" to "bridgetest@[domain]":
      """
      To: Internal Bridge <bridgetest@[domain]>

      this should fail
      """
    Then it fails

  Scenario: Send with empty TO
    When SMTP client "1" sends MAIL FROM "<user@[domain]>"
    Then it succeeds
    When SMTP client "1" sends RCPT TO "<>"
    Then it succeeds
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "invalid recipient"

  Scenario: Allow BODY parameter of MAIL FROM command
    When SMTP client "1" sends MAIL FROM "<user@[domain]> BODY=7BIT"
    Then it succeeds

  Scenario: FROM not owned by user
    When SMTP client "1" sends the following message from "other@[domain]" to "bridgetest@[domain]":
      """
      From: Bridge Test <user@[domain]>
      To: Internal Bridge <bridgetest@[domain]>

      this should fail
      """
    Then it fails