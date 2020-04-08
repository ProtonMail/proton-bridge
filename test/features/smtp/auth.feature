Feature: SMTP auth
  Scenario: Ask EHLO
    Given there is connected user "user"
    When SMTP client sends EHLO
    Then SMTP response is "OK"

  Scenario: Authenticates successfully and EHLO successfully
    Given there is connected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "OK"
    When SMTP client sends EHLO
    Then SMTP response is "OK"

  Scenario: Authenticates with bad password
    Given there is connected user "user"
    When SMTP client authenticates "user" with bad password
    Then SMTP response is "SMTP error: 454 backend/credentials: incorrect password"

  Scenario: Authenticates with disconnected user
    Given there is disconnected user "user"
    When SMTP client authenticates "user"
    Then SMTP response is "SMTP error: 454 bridge account is logged out, use bridge to login again"

  Scenario: Authenticates with no user
    When SMTP client authenticates with username "user@pm.me" and password "bridgepassword"
    Then SMTP response is "SMTP error: 454 user user@pm.me not found"

  Scenario: Authenticates with capital letter
    Given there is connected user "userAddressWithCapitalLetter"
    When SMTP client authenticates "userAddressWithCapitalLetter"
    Then SMTP response is "OK"

  Scenario: Authenticates with more addresses - primary one
    Given there is connected user "userMoreAddresses"
    When SMTP client authenticates "userMoreAddresses" with address "primary"
    Then SMTP response is "OK"

  Scenario: Authenticates with more addresses - secondary one
    Given there is connected user "userMoreAddresses"
    When SMTP client authenticates "userMoreAddresses" with address "secondary"
    Then SMTP response is "OK"

  Scenario: Authenticates with more addresses - disabled address
    Given there is connected user "userMoreAddresses"
    When SMTP client authenticates "userMoreAddresses" with address "disabled"
    Then SMTP response is "SMTP error: 454 user .* not found"

  @ignore-live
  Scenario: Authenticates with disabled primary address
    Given there is connected user "userDisabledPrimaryAddress"
    When SMTP client authenticates "userDisabledPrimaryAddress" with address "primary"
    Then SMTP response is "OK"

  Scenario: Authenticates two users
    Given there is connected user "user"
    And there is connected user "userMoreAddresses"
    When SMTP client "smtp1" authenticates "user"
    Then SMTP response to "smtp1" is "OK"
    When SMTP client "smtp2" authenticates "userMoreAddresses" with address "primary"
    Then SMTP response to "smtp2" is "OK"

  Scenario: Logs out user
    Given there is connected user "user"
    When SMTP client logs out
    Then SMTP response is "OK"
