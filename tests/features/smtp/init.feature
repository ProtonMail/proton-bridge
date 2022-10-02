Feature: SMTP initiation
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    When user "user@pm.me" connects and authenticates SMTP client "1"

  Scenario: Send without first announcing FROM and TO
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Reset is the same as without FROM and TO
    When SMTP client "1" sends MAIL FROM "<user@pm.me>"
    Then it succeeds
    When SMTP client "1" sends RCPT TO "<user@pm.me>"
    Then it succeeds
    When SMTP client "1" sends RSET
    Then it succeeds
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Send without FROM
    When SMTP client "1" sends RCPT TO "<user@pm.me>"
    Then it fails with error "Missing MAIL FROM command"

  Scenario: Send without TO
    When SMTP client "1" sends MAIL FROM "<user@pm.me>"
    Then it succeeds
    When SMTP client "1" sends DATA:
      """
      Subject: test
      """
    Then it fails with error "Missing RCPT TO command"

  Scenario: Send with empty FROM
    When SMTP client "1" sends MAIL FROM "<>"
    Then it fails with error "invalid return path"

  Scenario: Send with empty TO
    When SMTP client "1" sends MAIL FROM "<user@pm.me>"
    Then it succeeds
    When SMTP client "1" sends RCPT TO "<>"
    Then it fails with error "invalid recipient"

  Scenario: Allow BODY parameter of MAIL FROM command
    When SMTP client "1" sends MAIL FROM "<user@pm.me> BODY=7BIT"
    Then it succeeds

  Scenario: FROM not owned by user
    When SMTP client "1" sends MAIL FROM "<user@pm.test>"
    Then it fails with error "invalid return path"
