Feature: SMTP initiation
  Scenario: Empty FROM
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<>"
    Then SMTP response is "OK"

  Scenario: Send without FROM and TO
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "SMTP error: 502 5.5.1 Missing RCPT TO command."

  Scenario: Reset is the same as without FROM and TO
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<[userAddress]>"
    Then SMTP response is "OK"
    When SMTP client sends "RCPT TO:<user@pm.me>"
    Then SMTP response is "OK"
    When SMTP client sends "RSET"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "SMTP error: 502 5.5.1 Missing RCPT TO command"

  Scenario: Send without FROM
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "RCPT TO:<user@pm.me>"
    Then SMTP response is "SMTP error: 502 5.5.1 Missing MAIL FROM command."

  Scenario: Send without TO
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<[userAddress]>"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "SMTP error: 502 5.5.1 Missing RCPT TO command."

  Scenario: Send with empty FROM
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<>"
    Then SMTP response is "OK"
    When SMTP client sends "RCPT TO:<user@pm.me>"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "OK"
    When SMTP client sends "hello\r\n."
    Then SMTP response is "SMTP error: 554 5.0.0 Error: transaction failed, blame it on the weather: missing return path"

  Scenario: Send with empty TO
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<[userAddress]>"
    Then SMTP response is "OK"
    When SMTP client sends "RCPT TO:<>"
    Then SMTP response is "OK"
    When SMTP client sends "DATA"
    Then SMTP response is "OK"
    When SMTP client sends "hello\r\n."
    Then SMTP response is "SMTP error: 554 5.0.0 Error: transaction failed, blame it on the weather: missing recipient"

  Scenario: Allow BODY parameter of MAIL FROM command
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<[userAddress]> BODY=7BIT"
    Then SMTP response is "OK"

  Scenario: Allow AUTH parameter of MAIL FROM command
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<[userAddress]> AUTH=<>"
    Then SMTP response is "OK"

  Scenario: FROM not owned by user
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends "MAIL FROM:<user@pm.test>"
    Then SMTP response is "SMTP error: 451 4.0.0 backend: invalid return path: not owned by user"
